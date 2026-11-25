// Command snapshot backs up each local Qdrant peer to Garage and restores into a new collection.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"imavara.local/server/internal/platform/storage"
)

const source = "assets_v1"

var nodes = []string{"http://qdrant:6333"}
var httpClient = &http.Client{Timeout: 10 * time.Minute}
var identifier = regexp.MustCompile("^[a-zA-Z0-9_-]{1,64}$")

func send(ctx context.Context, method, url, kind string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if kind != "" {
		req.Header.Set("Content-Type", kind)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 != 2 {
		defer res.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return nil, fmt.Errorf("%s: HTTP %d %s", url, res.StatusCode, b)
	}
	return res, nil
}

type manifest struct {
	Objects []string `json:"objects"`
	Points  int      `json:"points"`
}

func count(ctx context.Context, collection string) (int, error) {
	res, err := send(ctx, "POST", nodes[0]+"/collections/"+collection+"/points/count", "application/json", strings.NewReader(`{"exact":true}`))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	var out struct{ Result struct{ Count int } }
	err = json.NewDecoder(res.Body).Decode(&out)
	return out.Result.Count, err
}
func backup(ctx context.Context, prefix string) error {
	var m manifest
	n, err := count(ctx, source)
	if err != nil {
		return err
	}
	m.Points = n
	for i, node := range nodes {
		path := node + "/collections/" + source + "/snapshots"
		res, err := send(ctx, "POST", path, "", nil)
		if err != nil {
			return err
		}
		var out struct{ Result struct{ Name string } }
		err = json.NewDecoder(res.Body).Decode(&out)
		res.Body.Close()
		if err != nil {
			return err
		}
		res, err = send(ctx, "GET", path+"/"+out.Result.Name, "", nil)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s/node-%d.snapshot", prefix, i)
		err = storage.Stream(ctx, "snapshots", key, "application/octet-stream", res.Body, res.ContentLength)
		res.Body.Close()
		if err != nil {
			return err
		}
		m.Objects = append(m.Objects, key)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return storage.Put(ctx, "snapshots", prefix+"/manifest.json", "application/json", raw)
}
func restore(ctx context.Context, prefix, target string) error {
	if !identifier.MatchString(target) || target == source {
		return errors.New("choose a new collection name")
	}
	res, err := httpClient.Get(nodes[0] + "/collections/" + target)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != 404 {
		return errors.New("restore requires an unused collection name")
	}
	f, err := storage.Get(ctx, "snapshots", prefix+"/manifest.json")
	if err != nil {
		return err
	}
	defer f.Close()
	var m manifest
	if err = json.NewDecoder(f).Decode(&m); err != nil {
		return err
	}
	if len(m.Objects) != len(nodes) {
		return errors.New("incomplete snapshot manifest")
	}
	for i, key := range m.Objects {
		obj, err := storage.Get(ctx, "snapshots", key)
		if err != nil {
			return err
		}
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		go func() {
			part, err := writer.CreateFormFile("snapshot", "node.snapshot")
			if err == nil {
				_, err = io.Copy(part, obj)
			}
			if err == nil {
				err = writer.Close()
			}
			pw.CloseWithError(err)
			obj.Close()
		}()
		res, err := send(ctx, "POST", nodes[i]+"/collections/"+target+"/snapshots/upload?priority=snapshot", writer.FormDataContentType(), pr)
		pr.Close()
		if err != nil {
			return err
		}
		res.Body.Close()
	}
	actual, err := count(ctx, target)
	if err != nil {
		return err
	}
	if actual != m.Points {
		return fmt.Errorf("restored %d of %d points", actual, m.Points)
	}
	// The restored vector index must return a known point for its original image vector.
	if actual > 0 {
		res, err := send(ctx, "POST", nodes[0]+"/collections/"+source+"/points/scroll", "application/json", strings.NewReader(`{"limit":1,"with_vector":true}`))
		if err != nil {
			return err
		}
		var page struct {
			Result struct {
				Points []struct {
					ID     string
					Vector map[string][]float64
				}
			}
		}
		err = json.NewDecoder(res.Body).Decode(&page)
		res.Body.Close()
		if err != nil {
			return err
		}
		if len(page.Result.Points) != 1 {
			return errors.New("missing reference point")
		}
		probe := page.Result.Points[0]
		raw, _ := json.Marshal(map[string]any{"using": "image", "query": probe.Vector["image"], "limit": 1, "filter": map[string]any{"must": []any{map[string]any{"has_id": []string{probe.ID}}}}})
		res, err = send(ctx, "POST", nodes[0]+"/collections/"+target+"/points/query", "application/json", bytes.NewReader(raw))
		if err != nil {
			return err
		}
		var hits struct {
			Result struct{ Points []struct{ ID string } }
		}
		err = json.NewDecoder(res.Body).Decode(&hits)
		res.Body.Close()
		if err != nil {
			return err
		}
		if len(hits.Result.Points) != 1 || hits.Result.Points[0].ID != probe.ID {
			return errors.New("restored reference search failed")
		}
	}
	return nil
}
func main() {
	target := flag.String("restore-to", "", "new collection for restore")
	prefix := flag.String("id", "", "backup ID")
	flag.Parse()
	if *prefix == "" && *target == "" {
		*prefix = time.Now().UTC().Format("20060102T150405.000000000Z")
		*prefix = strings.ReplaceAll(*prefix, ".", "-")
	}
	if !identifier.MatchString(*prefix) {
		log.Fatal("invalid snapshot ID")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
	defer cancel()
	var err error
	if *target == "" {
		err = backup(ctx, *prefix)
	} else {
		err = restore(ctx, *prefix, *target)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout, *prefix)
}
