package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"imavara.local/server/internal/platform/observe"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type route struct {
	name, description, method, path string
	fields, required                []string
}

var routes = []route{
	{"search_images", "Katalogda metinle veya referansla benzer görselleri ara.", "POST", "/v1/search", []string{"query", "reference_asset_id", "filters", "top_k"}, nil},
	{"get_image", "Görsel detayını oku.", "GET", "/v1/assets/{asset}", []string{"asset_id"}, []string{"asset_id"}},
	{"get_balance", "Kullanılabilir bakiye ve tarifeyi oku.", "GET", "/v1/me", nil, nil},
	{"acquire_image", "Seçilen orijinali edin.", "POST", "/v1/assets/{asset}/acquire", []string{"asset_id"}, []string{"asset_id"}},
	{"generate_image", "Açık taleple yeni görsel veya referanslı varyasyon üret.", "POST", "/v1/generations", []string{"prompt", "reference_asset_id", "aspect_ratio"}, []string{"prompt"}},
	{"contribute_image", "Sahibinin açık izniyle görseli katalog kabulüne gönder.", "POST", "/v1/contributions", []string{"asset_id", "license_scope"}, []string{"asset_id", "license_scope"}},
}
var apiClient = &http.Client{Timeout: 150 * time.Second}

func forward(ctx context.Context, r route, call *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	observe.Add("tool_calls_total", 1)
	var args map[string]any
	if len(call.Params.Arguments) > 0 {
		if err := json.Unmarshal(call.Params.Arguments, &args); err != nil {
			return nil, err
		}
	}
	path := r.path
	if strings.Contains(path, "{asset}") {
		value, _ := args["asset_id"].(string)
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, errors.New("geçersiz görsel kimliği")
		}
		path = strings.Replace(path, "{asset}", id.String(), 1)
	}
	base := os.Getenv("API_URL")
	if base == "" {
		base = "http://api:8080"
	}
	if r.name == "search_images" {
		if args == nil {
			args = map[string]any{}
		}
		args["include_nearest"] = true
	}
	payload, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, r.method, base+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	if call.Extra == nil || call.Extra.Header.Get("Cookie") == "" {
		return nil, errors.New("oturum gerekli")
	}
	req.Header.Set("Cookie", call.Extra.Header.Get("Cookie"))
	req.Header.Set("Idempotency-Key", call.Extra.Header.Get("X-Turn-ID")+":"+r.name)
	req.Header.Set("Content-Type", "application/json")
	res, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}, IsError: res.StatusCode/100 != 2}, nil
}
func main() {
	observe.Start("mcp")

	s := mcp.NewServer(&mcp.Implementation{Name: "imavara", Version: "0.1.0"}, nil)
	for _, r := range routes {
		props := map[string]any{}
		for _, name := range r.fields {
			if name == "top_k" {
				props[name] = map[string]any{"type": "integer", "minimum": 1, "maximum": 10, "default": 5, "description": "En fazla kaç uygun görsel döndürüleceği."}
				continue
			}
			kind := "string"
			if name == "filters" {
				kind = "object"
			}
			props[name] = map[string]any{"type": kind}
		}
		schema := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
		if len(r.required) > 0 {
			schema["required"] = r.required
		}
		s.AddTool(&mcp.Tool{Name: r.name, Description: r.description, InputSchema: schema}, func(ctx context.Context, call *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return forward(ctx, r, call)
		})
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)
	log.Fatal(http.ListenAndServe(":8081", mux))
}
