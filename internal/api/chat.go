package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type chatMessage struct {
	Role    string          `json:"role"`
	Content string          `json:"content"`
	Cards   json.RawMessage `json:"cards"`
	Turn    string          `json:"turn_id"`
}

func (s Server) chats(w http.ResponseWriter, r *http.Request) {
	owner := browser(r)
	if r.Method == "POST" {
		var id string
		if err := s.DB.QueryRowContext(r.Context(), "INSERT INTO conversations(browser_id) VALUES ($1) RETURNING id", owner).Scan(&id); err != nil {
			fail(w, err)
			return
		}
		reply(w, 201, map[string]string{"id": id})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), "SELECT id,title FROM conversations WHERE browser_id=$1 ORDER BY created_at DESC", owner)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	type chat struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	list := []chat{}
	for rows.Next() {
		var c chat
		if err = rows.Scan(&c.ID, &c.Title); err != nil {
			fail(w, err)
			return
		}
		list = append(list, c)
	}
	if err = rows.Err(); err != nil {
		fail(w, err)
		return
	}
	reply(w, 200, map[string]any{"items": list})
}
func (s Server) messages(ctx context.Context, id string) ([]chatMessage, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT role,content,cards,turn_id FROM messages WHERE conversation_id=$1 ORDER BY created_at,id", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []chatMessage{}
	for rows.Next() {
		var item chatMessage
		if err = rows.Scan(&item.Role, &item.Content, &item.Cards, &item.Turn); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s Server) conversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var exists string
	err := s.DB.QueryRowContext(r.Context(), "SELECT id FROM conversations WHERE id::text=$1 AND browser_id=$2", id, browser(r)).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	if r.Method == "DELETE" {
		_, err = s.DB.ExecContext(r.Context(), "DELETE FROM conversations WHERE id=$1 AND browser_id=$2", id, browser(r))
		if err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(204)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/ws") {
		s.turn(w, r, id)
		return
	}
	items, err := s.messages(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, 200, map[string]any{"items": items})
}

type chatEvent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	Cards    json.RawMessage `json:"cards,omitempty"`
	Sequence int             `json:"sequence"`
	Turn     string          `json:"turn_id"`
}

func (s Server) turn(w http.ResponseWriter, r *http.Request, id string) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetReadLimit(64 << 10)
	conn.SetReadDeadline(time.Now().Add(time.Minute))
	var in struct {
		Turn      string `json:"turn_id"`
		Text      string `json:"text"`
		Reference string `json:"reference_asset_id"`
	}
	if err = conn.ReadJSON(&in); err != nil || strings.TrimSpace(in.Text) == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	seq := 0
	emit := func(e chatEvent) {
		seq++
		e.Sequence = seq
		e.Turn = in.Turn
		conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
		conn.WriteJSON(e)
	}
	res, err := s.DB.ExecContext(ctx, `INSERT INTO messages(conversation_id,turn_id,role,content) VALUES($1,$2,'user',$3) ON CONFLICT DO NOTHING`, id, in.Turn, in.Text)
	if err != nil {
		emit(chatEvent{Type: "error", Text: "Mesaj kaydedilemedi."})
		return
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		emit(chatEvent{Type: "reload"})
		return
	}
	_, err = s.DB.ExecContext(ctx, "UPDATE conversations SET title=left($2,72) WHERE id=$1 AND title='Yeni sohbet'", id, in.Text)
	history, err := s.messages(ctx, id)
	if err != nil {
		emit(chatEvent{Type: "error", Text: "Sohbet okunamadı."})
		return
	}
	cookie, _ := r.Cookie("imavara_session")
	body, err := json.Marshal(map[string]any{"messages": history, "turn_id": in.Turn, "reference_asset_id": in.Reference})
	if err != nil {
		return
	}
	url := os.Getenv("AGENT_URL")
	if url == "" {
		url = "http://agent:8082"
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url+"/turn", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	upstream, err := http.DefaultClient.Do(req)
	var content strings.Builder
	var cards = json.RawMessage("[]")
	completed := false
	if err == nil {
		defer upstream.Body.Close()
		if upstream.StatusCode == 200 {
			dec := json.NewDecoder(io.LimitReader(upstream.Body, 4<<20))
			for {
				var e chatEvent
				if dec.Decode(&e) != nil {
					break
				}
				switch e.Type {
				case "delta":
					content.WriteString(e.Text)
				case "cards":
					cards = e.Cards
				case "tool":
				case "done":
					completed = true
				default:
					continue
				}
				if completed {
					break
				}
				emit(e)
			}
		}
	}
	if !completed {
		msg := "Agent yanıtı tamamlanamadı."
		content.WriteString(" " + msg)
	}
	_, err = s.DB.ExecContext(context.WithoutCancel(ctx), "INSERT INTO messages(conversation_id,turn_id,role,content,cards) VALUES($1,$2,'assistant',$3,$4)", id, in.Turn, strings.TrimSpace(content.String()), string(cards))
	if err != nil || !completed {
		emit(chatEvent{Type: "error", Text: "Agent yanıtı tamamlanamadı."})
		return
	}
	emit(chatEvent{Type: "done"})
}
