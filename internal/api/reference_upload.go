package api

import (
	"io"
	"net/http"

	"imavara.local/server/internal/catalog"
)

func (s Server) uploadReference(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 24<<20)
	reader, err := r.MultipartReader()
	if err != nil {
		reply(w, 400, map[string]string{"error": "Referans görseli seç."})
		return
	}
	part, err := reader.NextPart()
	if err != nil || part.FormName() != "image" {
		reply(w, 400, map[string]string{"error": "Görsel eksik."})
		return
	}
	data, err := io.ReadAll(part)
	if err != nil {
		reply(w, 413, map[string]string{"error": "Görsel yüklenemedi."})
		return
	}
	ctx := r.Context()
	a, err := catalog.Save(ctx, data)
	if err != nil {
		reply(w, 422, map[string]string{"error": "Görsel okunamadı veya kaydedilemedi."})
		return
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback()
	if err = insertAsset(ctx, tx, a, browser(r), "contribution", "", nil); err != nil {
		fail(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	c, err := s.getCard(ctx, a.ID, browser(r))
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, 201, c)
}
