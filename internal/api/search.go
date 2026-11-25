package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"imavara.local/server/internal/platform/observe"
	"imavara.local/server/internal/platform/openrouter"
	"imavara.local/server/internal/platform/vectors"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func mustMatch(name string, value any) any {
	return map[string]any{"key": name, "match": map[string]any{"value": value}}
}
func conditions(filters map[string]any) ([]any, error) {
	out := []any{mustMatch("visibility", "public"), mustMatch("catalog_status", "published"), mustMatch("license_scope", "shared")}
	for name, value := range filters {
		switch name {
		case "aspect_ratio", "orientation", "mime_type", "style", "theme", "color_family", "negative_space":
			text, ok := value.(string)
			if !ok || text == "" {
				return nil, errors.New("filtre değeri geçersiz")
			}
			out = append(out, mustMatch(name, text))
		case "has_people", "has_text", "has_alpha":
			v, ok := value.(bool)
			if !ok {
				return nil, errors.New("filtre doğru/yanlış olmalı")
			}
			out = append(out, mustMatch(name, v))
		case "min_width", "min_height":
			n, ok := value.(float64)
			if !ok || n <= 0 || n != float64(int(n)) {
				return nil, errors.New("boyut geçersiz")
			}
			out = append(out, map[string]any{"key": strings.TrimPrefix(name, "min_"), "range": map[string]any{"gte": int(n)}})
		default:
			return nil, errors.New("bilinmeyen filtre")
		}
	}
	return out, nil
}
func (s Server) search(w http.ResponseWriter, r *http.Request) {
	defer observe.Duration("search", time.Now())
	var in struct {
		TopK           *int           `json:"top_k"`
		IncludeNearest bool           `json:"include_nearest"`
		Query          string         `json:"query"`
		Reference      string         `json:"reference_asset_id"`
		Image          string         `json:"image_data_url"`
		Filters        map[string]any `json:"filters"`
	}
	defaultTopK := 5
	in.TopK = &defaultTopK
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 24<<20)).Decode(&in); err != nil {
		reply(w, 400, map[string]string{"error": "Geçersiz arama."})
		return
	}
	if in.TopK == nil || *in.TopK < 1 || *in.TopK > 10 {
		reply(w, http.StatusBadRequest, map[string]string{"error": "top_k 1 ile 10 arasında bir tam sayı olmalı."})
		return
	}
	filter, err := conditions(in.Filters)
	if err != nil {
		reply(w, 400, map[string]string{"error": err.Error()})
		return
	}
	ctx := r.Context()
	owner := browser(r)
	field := "text"
	value := "Instruct: Retrieve stock photographs matching the user's description.\nQuery: " + in.Query
	var embedding []float64
	if in.Reference != "" {
		c, err := s.getCard(ctx, in.Reference, owner)
		if err != nil {
			assetError(w, r, err)
			return
		}
		field = "image"
		if c.Status == "published" {
			var pointID string
			err = s.DB.QueryRowContext(ctx, "SELECT point_id FROM assets WHERE id=$1", c.ID).Scan(&pointID)
			if err != nil {
				fail(w, err)
				return
			}
			p, err := vectors.Read(ctx, pointID)
			if err != nil {
				fail(w, err)
				return
			}
			if p != nil {
				embedding = p.Vector["image"]
			}
		}
		if len(embedding) == 0 {
			value, _, err = s.reference(ctx, c.ID, owner)
			if err != nil {
				fail(w, err)
				return
			}
		}
	} else if in.Image != "" {
		if !strings.HasPrefix(in.Image, "data:image/jpeg;base64,") && !strings.HasPrefix(in.Image, "data:image/png;base64,") {
			reply(w, 400, map[string]string{"error": "JPEG veya PNG gerekli."})
			return
		}
		field, value = "image", in.Image
	} else if strings.TrimSpace(in.Query) == "" {
		reply(w, 400, map[string]string{"error": "Arama metni veya referans gerekli."})
		return
	}
	if len(embedding) == 0 {
		embedding, err = openrouter.Embed(ctx, field, value)
		if err != nil {
			reply(w, 502, map[string]string{"error": "Arama modeli kullanılamıyor."})
			return
		}
	}
	threshold := 0.70
	if val := os.Getenv("SEARCH_" + strings.ToUpper(field) + "_THRESHOLD"); val != "" {
		threshold, err = strconv.ParseFloat(val, 64)
		if err != nil {
			fail(w, err)
			return
		}
	}
	floor := threshold
	if in.IncludeNearest {
		floor = min(threshold, 0.45)
	}
	points, err := vectors.Query(ctx, field, embedding, filter, floor, *in.TopK)
	if err != nil {
		fail(w, err)
		return
	}
	searchID := uuid.NewString()
	items := []card{}
	var nearest *card
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback()
	for _, p := range points {
		id, ok := p.Payload["asset_id"].(string)
		if !ok {
			continue
		}
		c, err := scanCard(tx.QueryRowContext(ctx, "SELECT "+columns+",0 FROM assets a WHERE id::text=$2 AND status='published' AND license_scope='shared' FOR SHARE", owner, id))
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			fail(w, err)
			return
		}
		level := "strong"
		rank := len(items) + 1
		if p.Score < threshold {
			if len(items) > 0 || nearest != nil {
				break
			}
			level = "nearest"
		}
		c.Match = &searchMatch{Score: p.Score, Percent: int(math.Round(100 * min(1, max(0, p.Score)))), Level: level}
		if level == "nearest" {
			nearest = &c
		} else {
			items = append(items, c)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO search_returns(search_id,asset_id,rank) VALUES ($1,$2,$3)", searchID, id, rank); err != nil {
			fail(w, err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	if len(items) == 0 {
		observe.Add("search_empty_total", 1)
	}
	reply(w, 200, map[string]any{"top_k": *in.TopK, "items": items, "search_id": searchID, "nearest": nearest, "score_threshold": threshold, "candidate_threshold": floor, "metric": "cosine", "vector": field})
}
