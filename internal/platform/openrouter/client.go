package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"imavara.local/server/internal/platform/observe"
	"io"
	"net/http"
	"os"
	"time"
)

var ErrDisabled = errors.New("OpenRouter credentials missing")
var client = &http.Client{Timeout: 120 * time.Second}

func Post(ctx context.Context, path string, payload any, dest any) error {
	defer observe.Duration("provider", time.Now())
	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		return ErrDisabled
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	endpoint := "https://openrouter.ai/api/v1/" + path
	if path == "decisions" {
		endpoint = "https://openrouter.ai/api/alpha/decisions"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		observe.Add("provider_errors_total", 1)
		return &StatusError{Code: res.StatusCode, Path: path}
	}
	var raw json.RawMessage
	if err = json.NewDecoder(io.LimitReader(res.Body, 64<<20)).Decode(&raw); err != nil {
		return err
	}
	var meta struct{ Usage struct{ Cost float64 } }
	if json.Unmarshal(raw, &meta) == nil {
		observe.Add("provider_cost_usd_total", meta.Usage.Cost)
	}
	return json.Unmarshal(raw, dest)
}

type StatusError struct {
	Code int
	Path string
}

func (e *StatusError) Error() string { return fmt.Sprintf("OpenRouter %s: HTTP %d", e.Path, e.Code) }
func DefiniteFailure(err error) bool {
	var s *StatusError
	return errors.Is(err, ErrDisabled) || (errors.As(err, &s) && s.Code >= 400 && s.Code < 500 && s.Code != 408)
}
