package openrouter

import (
	"context"
	"errors"
	"math"
)

func Embed(ctx context.Context, field, value string) ([]float64, error) {
	model, size := "qwen/qwen3-embedding-8b", 1024
	var input any = value
	if field == "image" {
		model, size = "google/gemini-embedding-2", 768
		input = []any{map[string]any{"content": []any{map[string]any{"type": "image_url", "image_url": map[string]string{"url": value}}}}}
	}
	var res struct {
		Data []struct{ Embedding []float64 }
	}
	if err := Post(ctx, "embeddings", map[string]any{"model": model, "dimensions": size, "input": input, "encoding_format": "float"}, &res); err != nil {
		return nil, err
	}
	if len(res.Data) != 1 {
		return nil, errors.New("missing embedding")
	}
	v := res.Data[0].Embedding
	if field == "text" && len(v) == 4096 {
		v = v[:1024]
	}
	if len(v) != size {
		return nil, errors.New("unexpected embedding size")
	}
	var sum float64
	for _, f := range v {
		sum += f * f
	}
	if sum == 0 || math.IsInf(sum, 0) || math.IsNaN(sum) {
		return nil, errors.New("invalid embedding")
	}
	norm := math.Sqrt(sum)
	for i := range v {
		v[i] /= norm
	}
	return v, nil
}
