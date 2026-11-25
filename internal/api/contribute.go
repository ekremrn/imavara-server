package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"imavara.local/server/internal/catalog"

	"imavara.local/server/internal/platform/vectors"
	"io"
	"mime"
	"net/http"
	"strings"
)

func (s Server) contribute(w http.ResponseWriter, r *http.Request) {
	owner := browser(r)
	ctx := r.Context()
	var data []byte
	var id, license, url, point, hash, originalStatus string
	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	r.Body = http.MaxBytesReader(w, r.Body, 24<<20)
	if media == "multipart/form-data" {
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			reply(w, 400, map[string]string{"error": "Dosya okunamadı."})
			return
		}
		defer r.MultipartForm.RemoveAll()
		license = r.FormValue("license_scope")
		f, _, err := r.FormFile("image")
		if err != nil {
			reply(w, 400, map[string]string{"error": "Görsel eksik."})
			return
		}
		defer f.Close()
		data, err = io.ReadAll(f)
		if err != nil {
			fail(w, err)
			return
		}
		sum := sha256.Sum256(data)
		hash = hex.EncodeToString(sum[:])
		point = uuid.NewSHA1(uuid.NameSpaceOID, sum[:]).String()
	} else {
		var in struct {
			ID      string `json:"asset_id"`
			License string `json:"license_scope"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		id, license = in.ID, in.License
		err := s.DB.QueryRowContext(ctx, "SELECT sha256,point_id,status FROM assets WHERE id::text=$1 AND browser_id=$2", id, owner).Scan(&hash, &point, &originalStatus)
		if err != nil {
			assetError(w, r, err)
			return
		}
		if originalStatus == "accepted" || originalStatus == "published" {
			reply(w, 200, map[string]string{"asset_id": id, "status": originalStatus})
			return
		}
	}
	if license != "shared" {
		reply(w, 400, map[string]string{"error": "Ortak kullanım izni gerekli."})
		return
	}
	p, err := vectors.Read(ctx, point)
	if err != nil {
		fail(w, err)
		return
	}
	if p != nil && p.Payload["sha256"] == hash {
		reply(w, 409, map[string]string{"status": "duplicate"})
		return
	}
	if data != nil {
		a, err := catalog.Save(ctx, data)
		if err != nil {
			reply(w, 422, map[string]string{"error": "Görsel kaydedilemedi."})
			return
		}
		tx, err := s.DB.BeginTx(ctx, nil)
		if err != nil {
			fail(w, err)
			return
		}
		defer tx.Rollback()
		if err = insertAsset(ctx, tx, a, owner, "contribution", "", nil); err != nil {
			fail(w, err)
			return
		}
		if err = tx.Commit(); err != nil {
			fail(w, err)
			return
		}
		id = a.ID
		url = catalog.URL(data, a.Mime)
		originalStatus = "private"
	} else {
		url, _, err = s.reference(ctx, id, owner)
		if err != nil {
			assetError(w, r, err)
			return
		}
	}
	state, raw, err := catalog.Judge(ctx, url)
	if err != nil {
		reply(w, 502, map[string]string{"error": "Katalog kararı alınamadı.", "asset_id": id})
		return
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback()
	var version int
	err = tx.QueryRowContext(ctx, `UPDATE assets SET status=$1,license_scope='shared',decision=$2,version=version+1 WHERE id=$3 AND status=$4 RETURNING version`, state, string(raw), id, originalStatus).Scan(&version)
	var dbErr *pq.Error
	if errors.As(err, &dbErr) && dbErr.Code == "23505" {
		reply(w, 409, map[string]string{"status": "duplicate"})
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		reply(w, 409, map[string]string{"error": "Görsel durumu değişti."})
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	if state == "accepted" {
		if _, err = tx.ExecContext(ctx, "INSERT INTO catalog_jobs(asset_id,version) VALUES ($1,$2)", id, version); err != nil {
			fail(w, err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	reply(w, 200, map[string]any{"asset_id": id, "status": state, "decision": raw})
}
func (s Server) withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PATCH" {
		var in struct {
			Share bool `json:"share"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		if in.Share {
			body, _ := json.Marshal(map[string]string{"asset_id": r.PathValue("id"), "license_scope": "shared"})
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			r.Header.Set("Content-Type", "application/json")
			s.contribute(w, r)
			return
		}
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback()
	var point, previous string
	err = tx.QueryRowContext(r.Context(), "SELECT point_id,status FROM assets WHERE id::text=$1 AND browser_id=$2 FOR UPDATE", r.PathValue("id"), browser(r)).Scan(&point, &previous)
	if err != nil {
		assetError(w, r, err)
		return
	}
	if previous == "published" || previous == "accepted" {
		if err = vectors.Remove(r.Context(), point); err != nil {
			fail(w, err)
			return
		}
	}
	status := "withdrawn"
	if r.Method == "DELETE" {
		status = "removed"
	}
	if _, err = tx.ExecContext(r.Context(), "UPDATE assets SET status=$1,version=version+1 WHERE id=$2", status, r.PathValue("id")); err != nil {
		fail(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(204)
}
