package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"imavara.local/server/internal/platform/openrouter"
)

func Judge(ctx context.Context, url string) (string, json.RawMessage, error) {
	checks := map[string]string{"watermark": "Does the image contain an intrusive watermark?", "privacy": "Does it expose legible private personal information?", "damage": "Is the subject severely corrupted or unusable?", "scope": "Is the content unsuitable as a reusable stock visual?"}
	questions := map[string]any{}
	for key, prompt := range checks {
		questions[key] = map[string]any{"type": "noul", "instructions": prompt, "criteria": map[string]string{"true": "Visible evidence supports this condition.", "false": "The condition is absent."}}
	}
	var res struct {
		Answers map[string]struct {
			Score *float64 `json:"noul"`
		}
	}
	err := openrouter.Post(ctx, "decisions", map[string]any{"model": "openai/gpt-6-luna-decisions", "state": []any{map[string]any{"type": "image_url", "image_url": map[string]string{"url": url}}}, "questions": questions}, &res)
	if err != nil {
		return "", nil, err
	}
	state := "accepted"
	for k := range checks {
		a := res.Answers[k]
		if a.Score == nil || *a.Score < 0 || *a.Score > 1 {
			return "", nil, errors.New("incomplete decision")
		}
		if *a.Score >= 0.8 {
			state = "rejected"
		} else if *a.Score >= 0.4 && state != "rejected" {
			state = "review"
		}
	}
	raw, err := json.Marshal(res.Answers)
	return state, raw, err
}
