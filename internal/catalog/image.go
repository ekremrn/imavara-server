package catalog

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
	"image"
	"image/jpeg"
	_ "image/png"
	"imavara.local/server/internal/platform/storage"
)

type Image struct {
	ID, Hash, PointID, Mime string
	Width, Height           int
	Original, Preview       string
}

func URL(data []byte, mime string) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}
func Ratio(w, h int) string {
	a, b := w, h
	for b > 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d", w/a, h/a)
}
func Save(ctx context.Context, data []byte) (Image, error) {
	var out Image
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 48_000_000 {
		return out, errors.New("unsupported image or pixel count")
	}
	if format != "png" && format != "jpeg" && format != "webp" {
		return out, errors.New("expected JPEG, WebP or PNG")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return out, err
	}
	scale := (max(cfg.Width, cfg.Height) + 639) / 640
	if scale < 1 {
		scale = 1
	}
	preview := image.NewRGBA(image.Rect(0, 0, max(1, cfg.Width/scale), max(1, cfg.Height/scale)))
	for y := 0; y < preview.Rect.Dy(); y++ {
		for x := 0; x < preview.Rect.Dx(); x++ {
			preview.Set(x, y, img.At(x*scale, y*scale))
		}
	}
	var thumb bytes.Buffer
	if err = jpeg.Encode(&thumb, preview, &jpeg.Options{Quality: 82}); err != nil {
		return out, err
	}
	id := uuid.NewString()
	hash := sha256.Sum256(data)
	out = Image{id, hex.EncodeToString(hash[:]), uuid.NewSHA1(uuid.NameSpaceOID, hash[:]).String(), "image/" + format, cfg.Width, cfg.Height, id + "/original." + format, id + "/preview.jpg"}
	if err = storage.Put(ctx, "originals", out.Original, out.Mime, data); err != nil {
		return out, err
	}
	return out, storage.Put(ctx, "previews", out.Preview, "image/jpeg", thumb.Bytes())
}
