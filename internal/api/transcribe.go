package api

import (
	"bytes"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"imavara.local/server/internal/platform/openrouter"
)

func audioFormat(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0x1a, 0x45, 0xdf, 0xa3}):
		return "webm"
	case len(data) >= 12 && string(data[4:8]) == "ftyp":
		return "m4a"
	case bytes.HasPrefix(data, []byte("OggS")):
		return "ogg"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WAVE":
		return "wav"
	default:
		return ""
	}
}
func (s Server) transcribe(w http.ResponseWriter, r *http.Request) {
	// Read the recording in memory so multipart uploads never spill raw voice to disk.
	const maxRecording = 16 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxRecording+4096)
	reader, err := r.MultipartReader()
	if err != nil {
		reply(w, 400, map[string]string{"error": "Ses kaydı multipart biçiminde gönderilmeli."})
		return
	}
	part, err := reader.NextPart()
	if err != nil || part.FormName() != "audio" {
		reply(w, 400, map[string]string{"error": "Ses kaydı eksik."})
		return
	}
	data, err := io.ReadAll(io.LimitReader(part, maxRecording+1))
	if err != nil || len(data) > maxRecording {
		reply(w, 413, map[string]string{"error": "Kayıt okunamadı veya boyut sınırını aştı."})
		return
	}
	format := audioFormat(data)
	if format == "" {
		reply(w, 415, map[string]string{"error": "Bu ses biçimi desteklenmiyor."})
		return
	}
	var result struct {
		Text string `json:"text"`
	}
	input := map[string]any{"model": "openai/gpt-transcribe", "input_audio": map[string]string{"format": format, "data": base64.StdEncoding.EncodeToString(data)}}
	if language := r.URL.Query().Get("language"); language == "tr" || language == "en" {
		input["language"] = language
	}
	err = openrouter.Post(r.Context(), "audio/transcriptions", input, &result)
	if err != nil {
		slog.Warn("transcription unavailable", "error", err)
		reply(w, 502, map[string]string{"error": "Ses şu anda yazıya çevrilemedi."})
		return
	}
	result.Text = strings.TrimSpace(result.Text)
	if result.Text == "" {
		reply(w, 422, map[string]string{"error": "Kayıtta konuşma algılanmadı."})
		return
	}
	reply(w, 200, result)
}
