package vectors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const collection = "/collections/assets_v1"

var client = &http.Client{Timeout: 60 * time.Second}

func call(ctx context.Context, method, path string, payload any, dest any) (int, error) {
	base := os.Getenv("QDRANT_URL")
	if base == "" {
		base = "http://qdrant:6333"
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return res.StatusCode, fmt.Errorf("qdrant %s %s: HTTP %d", method, path, res.StatusCode)
	}
	if dest != nil {
		err = json.NewDecoder(io.LimitReader(res.Body, 16<<20)).Decode(dest)
	}
	return res.StatusCode, err
}

func Bootstrap(ctx context.Context) error {
	code, err := call(ctx, "GET", collection, nil, nil)
	if code == 404 {
		config := map[string]any{"vectors": map[string]any{"text": map[string]any{"size": 1024, "distance": "Cosine", "on_disk": true}, "image": map[string]any{"size": 768, "distance": "Cosine", "on_disk": true}}, "shard_number": 16, "replication_factor": 1, "quantization_config": map[string]any{"binary": map[string]any{"always_ram": true}}, "hnsw_config": map[string]int{"m": 16, "ef_construct": 100}}
		_, err = call(ctx, "PUT", collection, config, nil)
	}
	if err != nil {
		return err
	}
	for kind, fields := range map[string][]string{"keyword": {"visibility", "catalog_status", "license_scope", "aspect_ratio", "orientation", "mime_type", "style", "theme", "color_family", "negative_space"}, "integer": {"width", "height"}, "uuid": {"owner_browser_id"}, "bool": {"has_people", "has_text", "has_alpha"}, "datetime": {"published_at"}} {
		for _, field := range fields {
			if _, err := call(ctx, "PUT", collection+"/index?wait=true", map[string]string{"field_name": field, "field_schema": kind}, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

type Point struct {
	ID      string               `json:"id"`
	Vector  map[string][]float64 `json:"vector,omitempty"`
	Payload map[string]any       `json:"payload"`
	Score   float64              `json:"score,omitempty"`
}

func Write(ctx context.Context, p Point) error {
	_, err := call(ctx, "PUT", collection+"/points?wait=true", map[string]any{"points": []Point{p}}, nil)
	return err
}
func Read(ctx context.Context, id string) (*Point, error) {
	var res struct{ Result Point }
	code, err := call(ctx, "GET", collection+"/points/"+id, nil, &res)
	if code == 404 {
		return nil, nil
	}
	return &res.Result, err
}
func Remove(ctx context.Context, id string) error {
	_, err := call(ctx, "POST", collection+"/points/delete?wait=true", map[string]any{"points": []string{id}}, nil)
	return err
}
func Query(ctx context.Context, field string, v []float64, filters []any, threshold float64, limit int) ([]Point, error) {
	if field != "text" && field != "image" {
		return nil, errors.New("unknown vector")
	}
	var res struct{ Result struct{ Points []Point } }
	_, err := call(ctx, "POST", collection+"/points/query", map[string]any{"using": field, "query": v, "filter": map[string]any{"must": filters}, "score_threshold": threshold, "limit": limit, "with_payload": true, "params": map[string]any{"hnsw_ef": 128, "quantization": map[string]any{"rescore": true, "oversampling": 3}}}, &res)
	return res.Result.Points, err
}
