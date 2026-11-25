package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"imavara.local/server/internal/catalog"
	"imavara.local/server/internal/platform/storage"
	"io"
	"net/http"
	"strconv"
)

type searchMatch struct {
	Score   float64 `json:"score"`
	Percent int     `json:"percent"`
	Level   string  `json:"level"`
}

type card struct {
	Match          *searchMatch `json:"match,omitempty"`
	Owned          bool         `json:"owned"`
	Acquired       bool         `json:"acquired"`
	Ratio          string       `json:"aspect_ratio"`
	RequestedRatio string       `json:"requested_ratio,omitempty"`
	ID             string       `json:"id"`
	Caption        string       `json:"caption"`
	Width          int          `json:"width"`
	Height         int          `json:"height"`
	Preview        string       `json:"preview_url"`
	Status         string       `json:"status"`
	Origin         string       `json:"origin"`
	Mime           string       `json:"mime_type"`
	Count          int          `json:"count"`
}

const columns = "a.id,a.caption,a.width,a.height,a.status,a.origin,a.mime_type,a.requested_ratio,(a.browser_id=$1),EXISTS(SELECT 1 FROM acquisitions aq WHERE aq.asset_id=a.id AND aq.browser_id=$1)"
const visible = "(a.status='published' OR a.browser_id=$1 OR EXISTS(SELECT 1 FROM acquisitions e WHERE e.asset_id=a.id AND e.browser_id=$1))"

func scanCard(row interface{ Scan(...any) error }) (card, error) {
	var c card
	err := row.Scan(&c.ID, &c.Caption, &c.Width, &c.Height, &c.Status, &c.Origin, &c.Mime, &c.RequestedRatio, &c.Owned, &c.Acquired, &c.Count)
	c.Ratio = catalog.Ratio(c.Width, c.Height)
	c.Preview = "/v1/assets/" + c.ID + "/preview"
	return c, err
}
func (s Server) getCard(ctx context.Context, id, owner string) (card, error) {
	return scanCard(s.DB.QueryRowContext(ctx, "SELECT "+columns+",0 FROM assets a WHERE "+visible+" AND a.id::text=$2", owner, id))
}
func assetError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	fail(w, err)
}
func (s Server) detail(w http.ResponseWriter, r *http.Request) {
	c, err := s.getCard(r.Context(), r.PathValue("id"), browser(r))
	if err != nil {
		assetError(w, r, err)
		return
	}
	reply(w, 200, c)
}
func (s Server) listAssets(w http.ResponseWriter, r *http.Request) {
	args := []any{browser(r)}
	where := "a.status='published'"
	if r.URL.Path == "/v1/me/assets" {
		where = "((a.browser_id=$1 AND a.status<>'removed') OR EXISTS(SELECT 1 FROM acquisitions e WHERE e.asset_id=a.id AND e.browser_id=$1))"
	}
	order := "a.published_at DESC NULLS LAST,a.created_at DESC"
	count := "0"
	if r.URL.Path == "/v1/stats/assets" {
		days, _ := strconv.Atoi(r.URL.Query().Get("days"))
		if days <= 0 {
			days = 30
		}
		table := "acquisitions"
		if r.URL.Query().Get("metric") == "search_returns" {
			table = "search_returns"
		}
		count = "(SELECT count(*) FROM " + table + " x WHERE x.asset_id=a.id AND x.created_at>=now()-$2::int*interval '1 day')"
		args = append(args, days)
		order = "uses DESC," + order
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	args = append(args, offset)
	rows, err := s.DB.QueryContext(r.Context(), "SELECT "+columns+","+count+" AS uses FROM assets a WHERE "+where+" ORDER BY "+order+",a.id LIMIT 30 OFFSET "+fmt.Sprintf("$%d", len(args)), args...)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	items := []card{}
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			fail(w, err)
			return
		}
		items = append(items, c)
	}
	if err = rows.Err(); err != nil {
		fail(w, err)
		return
	}
	reply(w, 200, map[string]any{"items": items})
}
func (s Server) file(w http.ResponseWriter, r *http.Request) {
	preview := r.PathValue("kind") == "preview"
	if !preview && r.PathValue("kind") != "original" {
		http.NotFound(w, r)
		return
	}
	clause := "(a.browser_id=$1 OR EXISTS(SELECT 1 FROM acquisitions e WHERE e.asset_id=a.id AND e.browser_id=$1))"
	bucket, col := "originals", "object_key"
	if preview {
		clause = visible
		bucket, col = "previews", "preview_key"
	}
	var key, mime string
	err := s.DB.QueryRowContext(r.Context(), "SELECT "+col+",mime_type FROM assets a WHERE "+clause+" AND a.id::text=$2", browser(r), r.PathValue("id")).Scan(&key, &mime)
	if err != nil {
		assetError(w, r, err)
		return
	}
	obj, err := storage.Get(r.Context(), bucket, key)
	if err != nil {
		fail(w, err)
		return
	}
	defer obj.Close()
	if _, err = obj.Stat(); err != nil {
		fail(w, err)
		return
	}
	if preview {
		mime = "image/jpeg"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, no-store")
	if !preview {
		w.Header().Set("Content-Disposition", "attachment")
	}
	io.Copy(w, obj)
}
func (s Server) acquire(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	owner, id := browser(r), r.PathValue("id")
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback()
	var locked string
	if err = tx.QueryRowContext(ctx, "SELECT id FROM browsers WHERE id=$1 FOR UPDATE", owner).Scan(&locked); err != nil {
		fail(w, err)
		return
	}
	var owned, acquired bool
	var status string
	err = tx.QueryRowContext(ctx, "SELECT browser_id=$1,status,EXISTS(SELECT 1 FROM acquisitions WHERE asset_id=a.id AND browser_id=$1) FROM assets a WHERE id::text=$2 FOR SHARE", owner, id).Scan(&owned, &status, &acquired)
	if err != nil {
		assetError(w, r, err)
		return
	}
	charged := 0
	if !owned && !acquired {
		if status != "published" {
			http.NotFound(w, r)
			return
		}
		var available int
		err = tx.QueryRowContext(ctx, "SELECT (SELECT sum(amount) FROM credit_movements WHERE browser_id=$1)-(SELECT coalesce(sum(amount),0) FROM credit_holds WHERE browser_id=$1)", owner).Scan(&available)
		if err != nil {
			fail(w, err)
			return
		}
		if available < 1 {
			reply(w, 402, map[string]string{"error": "Yetersiz kredi."})
			return
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO acquisitions(browser_id,asset_id) VALUES ($1,$2)", owner, id); err != nil {
			fail(w, err)
			return
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO credit_movements(browser_id,operation_key,reason,amount) VALUES ($1,$2,'acquisition',-1)", owner, "acquire:"+id); err != nil {
			fail(w, err)
			return
		}
		charged = 1
	}
	if err = tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	reply(w, 200, map[string]any{"charged": charged, "original_url": "/v1/assets/" + id + "/original"})
}
