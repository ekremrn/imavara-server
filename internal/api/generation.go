package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"imavara.local/server/internal/catalog"
	"imavara.local/server/internal/credits"
	"imavara.local/server/internal/platform/openrouter"
	"imavara.local/server/internal/platform/storage"
	"io"
	"net/http"
	"strings"
	"time"
)

func (s Server) reference(ctx context.Context, id, owner string) (string, string, error) {
	var key, mime string
	var w, h int
	err := s.DB.QueryRowContext(ctx, "SELECT a.object_key,a.mime_type,a.width,a.height FROM assets a WHERE "+visible+" AND a.id::text=$2", owner, id).Scan(&key, &mime, &w, &h)
	if err != nil {
		return "", "", err
	}
	obj, err := storage.Get(ctx, "originals", key)
	if err != nil {
		return "", "", err
	}
	defer obj.Close()
	data, err := io.ReadAll(io.LimitReader(obj, 40<<20))
	if err != nil {
		return "", "", err
	}
	return catalog.URL(data, mime), catalog.Ratio(w, h), nil
}
func insertAsset(ctx context.Context, tx *sql.Tx, a catalog.Image, owner, origin, caption string, parent *string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO assets(id,browser_id,point_id,sha256,object_key,preview_key,mime_type,width,height,origin,caption,parent_id)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, a.ID, owner, a.PointID, a.Hash, a.Original, a.Preview, a.Mime, a.Width, a.Height, origin, caption, parent)
	return err
}
func (s Server) generate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Prompt    string `json:"prompt"`
		Reference string `json:"reference_asset_id"`
		Ratio     string `json:"aspect_ratio"`
	}
	key := r.Header.Get("Idempotency-Key")
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&in); err != nil || strings.TrimSpace(in.Prompt) == "" || key == "" || len(key) > 256 {
		reply(w, 400, map[string]string{"error": "Tarif ve işlem kimliği gerekli."})
		return
	}
	ctx := r.Context()
	owner := browser(r)
	body, _ := json.Marshal(in)
	digest := sha256.Sum256(body)
	req := map[string]any{"model": "meta/muse-image", "prompt": in.Prompt, "n": 1, "response_format": "b64_json"}
	var parent *string
	if in.Reference != "" {
		url, ratio, err := s.reference(ctx, in.Reference, owner)
		if err != nil {
			assetError(w, r, err)
			return
		}
		req["input_references"] = []any{map[string]any{"type": "image_url", "image_url": map[string]string{"url": url}}}
		parent = &in.Reference
		if in.Ratio == "" {
			in.Ratio = ratio
		}
	}
	if in.Ratio == "" {
		in.Ratio = "1:1"
	}
	req["aspect_ratio"] = in.Ratio
	l := credits.Ledger{DB: s.DB}
	g, start, err := l.ReserveOnlyOnce(ctx, owner, key, digest[:])
	if errors.Is(err, credits.ErrInsufficient) {
		reply(w, 402, map[string]string{"error": "Yetersiz kredi."})
		return
	}
	if errors.Is(err, credits.ErrKeyConflict) {
		reply(w, 409, map[string]string{"error": "İşlem kimliği başka isteğe ait."})
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	if !start {
		if g.AssetID.Valid {
			c, err := s.getCard(ctx, g.AssetID.String, owner)
			if err != nil {
				fail(w, err)
				return
			}
			reply(w, 200, c)
			return
		}
		reply(w, 409, map[string]string{"status": g.Status, "error": "İşlem daha önce alındı."})
		return
	}
	if err = l.Submit(ctx, owner, key); err != nil {
		fail(w, err)
		return
	}
	var result struct {
		Data []struct {
			B64 string `json:"b64_json"`
		}
	}
	// Complete bookkeeping even when the browser disconnects after the provider accepts the work.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
	defer cancel()
	for attempt := 0; attempt < 3; attempt++ {
		err = openrouter.Post(ctx, "images", req, &result)
		var rejected *openrouter.StatusError
		if !errors.As(err, &rejected) || rejected.Code != http.StatusTooManyRequests || attempt == 2 {
			break
		}
		select {
		case <-time.After(time.Duration(attempt+1) * time.Second):
		case <-ctx.Done():
		}
	}
	if err != nil || len(result.Data) != 1 {
		// No deliverable means no user charge. Keep an ambiguous provider outcome terminal to prevent a duplicate order.
		outcome := "rejected"
		if err != nil && !openrouter.DefiniteFailure(err) {
			outcome = "unknown"
		}
		if finishErr := s.closeUndelivered(context.WithoutCancel(ctx), owner, key, outcome); finishErr != nil {
			fail(w, finishErr)
			return
		}
		reply(w, 502, map[string]string{"error": "Görsel teslim edilemedi."})
		return
	}
	data, err := base64.StdEncoding.DecodeString(result.Data[0].B64)
	if err != nil {
		if releaseErr := s.releaseGeneration(ctx, owner, key); releaseErr != nil {
			fail(w, releaseErr)
			return
		}
		fail(w, err)
		return
	}
	a, err := catalog.Save(ctx, data)
	if err != nil {
		if releaseErr := s.releaseGeneration(ctx, owner, key); releaseErr != nil {
			fail(w, releaseErr)
			return
		}
		fail(w, err)
		return
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback()
	if err = insertAsset(ctx, tx, a, owner, "generated", in.Prompt, parent); err != nil {
		fail(w, err)
		return
	}
	if _, err = tx.ExecContext(ctx, "UPDATE assets SET requested_ratio=$2, status='pending_decision', license_scope='shared', generation_prompt=caption WHERE id=$1", a.ID, in.Ratio); err != nil {
		fail(w, err)
		return
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO catalog_jobs(asset_id,version) SELECT id,version FROM assets WHERE id=$1", a.ID); err != nil {
		fail(w, err)
		return
	}
	if err = credits.Finish(ctx, tx, owner, key, &a.ID); err != nil {
		fail(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	c, err := s.getCard(ctx, a.ID, owner)
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, 201, c)
}
func (s Server) releaseGeneration(ctx context.Context, owner, key string) error {
	return s.closeUndelivered(ctx, owner, key, "undelivered")
}
func (s Server) closeUndelivered(ctx context.Context, owner, key, outcome string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = credits.Finish(ctx, tx, owner, key, nil); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE generations SET provider_outcome=$3 WHERE browser_id=$1 AND operation_key=$2", owner, key, outcome); err != nil {
		return err
	}
	return tx.Commit()
}
