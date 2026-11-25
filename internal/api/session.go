package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"imavara.local/server/internal/platform/observe"
	"log/slog"
	"net/http"
)

type Server struct{ DB *sql.DB }
type browserKey struct{}

func browser(r *http.Request) string { return r.Context().Value(browserKey{}).(string) }
func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Warn("response interrupted")
	}
}
func fail(w http.ResponseWriter, err error) {
	observe.Error(err)
	slog.Error("request failed", "error", err)
	reply(w, 500, map[string]string{"error": "İşlem tamamlanamadı."})
}

func (s Server) session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observe.Add("api_requests_total", 1)
		var id string
		cookie, err := r.Cookie("imavara_session")
		if err == nil {
			hash := sha256.Sum256([]byte(cookie.Value))
			err = s.DB.QueryRowContext(r.Context(), "SELECT id FROM browsers WHERE session_hash=$1", hash[:]).Scan(&id)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				fail(w, err)
				return
			}
		}
		if id == "" {
			var secret [32]byte
			if _, err = rand.Read(secret[:]); err != nil {
				fail(w, err)
				return
			}
			token := base64.RawURLEncoding.EncodeToString(secret[:])
			hash := sha256.Sum256([]byte(token))
			err = s.DB.QueryRowContext(r.Context(), `WITH created AS (
      INSERT INTO browsers(session_hash) VALUES ($1) RETURNING id
   ) INSERT INTO credit_movements(browser_id,operation_key,reason,amount)
   SELECT id,'welcome','welcome',100 FROM created RETURNING browser_id`, hash[:]).Scan(&id)
			if err != nil {
				fail(w, err)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "imavara_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil, MaxAge: 31536000})
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), browserKey{}, id)))
	})
}

func (s Server) Routes() http.Handler {
	private := http.NewServeMux()
	private.HandleFunc("POST /v1/references", s.uploadReference)
	private.HandleFunc("POST /v1/contributions", s.contribute)
	private.HandleFunc("PATCH /v1/assets/{id}", s.withdraw)
	private.HandleFunc("DELETE /v1/assets/{id}", s.withdraw)
	private.HandleFunc("POST /v1/search", s.search)
	private.HandleFunc("POST /v1/generations", s.generate)
	private.HandleFunc("GET /v1/assets", s.listAssets)
	private.HandleFunc("GET /v1/me/assets", s.listAssets)
	private.HandleFunc("GET /v1/stats/assets", s.listAssets)
	private.HandleFunc("GET /v1/assets/{id}", s.detail)
	private.HandleFunc("GET /v1/assets/{id}/{kind}", s.file)
	private.HandleFunc("POST /v1/assets/{id}/acquire", s.acquire)
	private.HandleFunc("GET /v1/conversations", s.chats)
	private.HandleFunc("POST /v1/conversations", s.chats)
	private.HandleFunc("GET /v1/conversations/{id}", s.conversation)
	private.HandleFunc("DELETE /v1/conversations/{id}", s.conversation)
	private.HandleFunc("GET /v1/conversations/{id}/ws", s.conversation)
	private.HandleFunc("POST /v1/audio/transcriptions", s.transcribe)
	private.HandleFunc("GET /v1/me", s.me)
	private.HandleFunc("GET /v1/me/credits/transactions", s.history)
	private.HandleFunc("GET /v1/pricing", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, map[string]int{"acquire": 1, "generate": 3, "variation": 3, "contribution_reward": 1, "search": 0, "chat": 0, "transcribe": 0})
	})
	root := http.NewServeMux()
	root.Handle("/v1/", s.session(private))
	root.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.DB.PingContext(r.Context()); err != nil {
			reply(w, 503, map[string]bool{"ready": false})
			return
		}
		reply(w, 200, map[string]bool{"ready": true})
	})
	return root
}

func (s Server) me(w http.ResponseWriter, r *http.Request) {
	var balance int
	err := s.DB.QueryRowContext(r.Context(), `SELECT COALESCE(sum(amount),0) -
  (SELECT COALESCE(sum(amount),0) FROM credit_holds WHERE browser_id=$1)
  FROM credit_movements WHERE browser_id=$1`, browser(r)).Scan(&balance)
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, 200, map[string]any{"id": browser(r), "balance": balance})
}
func (s Server) history(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,reason,amount,created_at FROM credit_movements WHERE browser_id=$1 ORDER BY id DESC`, browser(r))
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	type movement struct {
		ID      int64  `json:"id"`
		Reason  string `json:"reason"`
		Amount  int    `json:"amount"`
		Created string `json:"created_at"`
	}
	items := []movement{}
	for rows.Next() {
		var m movement
		if err = rows.Scan(&m.ID, &m.Reason, &m.Amount, &m.Created); err != nil {
			fail(w, err)
			return
		}
		items = append(items, m)
	}
	if err = rows.Err(); err != nil {
		fail(w, err)
		return
	}
	reply(w, 200, map[string]any{"items": items})
}
