package workers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"image"
	"imavara.local/server/internal/catalog"
	"imavara.local/server/internal/platform/observe"
	"imavara.local/server/internal/platform/openrouter"
	"imavara.local/server/internal/platform/vectors"
	"strings"
	"time"
)

type analysis struct {
	Caption, Description string
	Main                 string   `json:"main_subject"`
	Relations            []string `json:"actions_and_relations"`
	Keywords, Theme      []string
	People               string `json:"people_presence"`
	Composition          struct {
		Subject            string   `json:"subject_position"`
		Copy               []string `json:"copy_space"`
		Framing, Viewpoint string
	}
	Appearance struct {
		Style  string
		Colors []string `json:"color_family"`
	}
	Text struct{ Status string }
}

func (r runtime) index(ctx context.Context, j Job) error {
	var mime, hash, point, owner, origin, prompt string
	var w, h int
	err := r.db.QueryRowContext(ctx, `SELECT mime_type,sha256,point_id,browser_id,origin,width,height,generation_prompt FROM assets WHERE id=$1 AND version=$2 AND status='accepted'`, j.Asset, j.Version).Scan(&mime, &hash, &point, &owner, &origin, &w, &h, &prompt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	raw, err := readObject(ctx, "analysis", analysisKey(j))
	if err != nil {
		return err
	}
	var saved struct{ Analysis analysis }
	if err = json.Unmarshal(raw, &saved); err != nil {
		return err
	}
	a := saved.Analysis
	description := strings.Join([]string{a.Description, a.Main, strings.Join(a.Relations, ". "), a.Composition.Framing, a.Composition.Viewpoint, "subject " + a.Composition.Subject, "copy space " + strings.Join(a.Composition.Copy, ", "), strings.Join(a.Keywords, ", ")}, ". ")
	if prompt != "" {
		description += ". Generation brief: " + prompt
	}
	text, err := openrouter.Embed(ctx, "text", description)
	if err != nil {
		return err
	}
	img, err := readObject(ctx, "originals", j.Object)
	if err != nil {
		return err
	}
	imageVector, err := openrouter.Embed(ctx, "image", catalog.URL(img, mime))
	if err != nil {
		return err
	}
	orientation := "square"
	if w > h {
		orientation = "landscape"
	} else if h > w {
		orientation = "portrait"
	}
	payload := map[string]any{"asset_id": j.Asset, "sha256": hash, "visibility": "public", "catalog_status": "published", "license_scope": "shared", "aspect_ratio": catalog.Ratio(w, h), "width": w, "height": h, "mime_type": mime, "orientation": orientation, "style": a.Appearance.Style, "theme": a.Theme, "color_family": a.Appearance.Colors, "negative_space": a.Composition.Copy, "published_at": time.Now().UTC().Format(time.RFC3339)}
	decoded, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return err
	}
	payload["has_alpha"] = false
	if opaque, ok := decoded.(interface{ Opaque() bool }); ok {
		payload["has_alpha"] = !opaque.Opaque()
	}
	if a.People != "unknown" {
		payload["has_people"] = a.People == "present"
	}
	if a.Text.Status != "unknown" {
		payload["has_text"] = a.Text.Status != "absent"
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	err = tx.QueryRowContext(ctx, "SELECT version FROM assets WHERE id=$1 AND version=$2 AND status='accepted' FOR UPDATE", j.Asset, j.Version).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = vectors.Write(ctx, vectors.Point{ID: point, Vector: map[string][]float64{"text": text, "image": imageVector}, Payload: payload}); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE assets SET status='published',published_at=now(),caption=$2 WHERE id=$1", j.Asset, a.Caption); err != nil {
		return err
	}
	if origin == "contribution" {
		if _, err = tx.ExecContext(ctx, `INSERT INTO credit_movements(browser_id,operation_key,reason,amount) VALUES ($1,$2,'contribution',1) ON CONFLICT DO NOTHING`, owner, "contribution:"+hash); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	var seconds float64
	if r.db.QueryRowContext(ctx, "SELECT EXTRACT(EPOCH FROM published_at-created_at) FROM assets WHERE id=$1", j.Asset).Scan(&seconds) == nil {
		observe.Add("publication_seconds_sum", seconds)
		observe.Add("publication_seconds_count", 1)
	}
	return nil
}
