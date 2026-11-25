package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/lib/pq"
	"imavara.local/server/internal/api"
	"imavara.local/server/internal/credits"
	"imavara.local/server/migrations"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

// The integration database is disposable and supplied explicitly by the test runner.
func TestSessionAndConcurrentReservations(t *testing.T) {
	dsn := os.Getenv("IMAVARA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL database")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err = migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(api.Server{DB: db}.Routes())
	defer server.Close()
	account := func(client *http.Client) (string, int) {
		t.Helper()
		res, err := client.Get(server.URL + "/v1/me")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var data struct {
			ID      string
			Balance int
		}
		if err = json.NewDecoder(res.Body).Decode(&data); err != nil || data.ID == "" {
			t.Fatalf("invalid account: %v", err)
		}
		return data.ID, data.Balance
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	id, balance := account(client)
	other, _ := account(&http.Client{})
	again, _ := account(client)
	if balance != 100 || again != id || other == id {
		t.Fatal("browser initialization or isolation failed")
	}
	_, err = db.Exec(`INSERT INTO credit_movements(browser_id,operation_key,reason,amount) VALUES ($1,'test-balance','fixture',-97)`, id)
	if err != nil {
		t.Fatal(err)
	}
	ledger := credits.Ledger{DB: db}
	type result struct {
		key   string
		start bool
		err   error
	}
	ch := make(chan result, 2)
	var wg sync.WaitGroup
	for _, key := range []string{"one", "two"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			_, ok, err := ledger.ReserveOnlyOnce(ctx, id, key, []byte(key))
			ch <- result{key, ok, err}
		}(key)
	}
	wg.Wait()
	close(ch)
	winner := ""
	rejected := 0
	for result := range ch {
		if result.start && result.err == nil {
			winner = result.key
		} else if errors.Is(result.err, credits.ErrInsufficient) {
			rejected++
		} else {
			t.Fatalf("unexpected result: %+v", result)
		}
	}
	if rejected != 1 || winner == "" {
		t.Fatal("overspend")
	}
	_, available := account(client)
	if available != 0 {
		t.Fatal("reservation not reflected")
	}
	if _, start, err := ledger.ReserveOnlyOnce(ctx, id, winner, []byte(winner)); err != nil || start {
		t.Fatal("duplicate reservation")
	}
	if err = ledger.Submit(ctx, id, winner); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err = credits.Finish(ctx, tx, id, winner, nil); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	_, available = account(client)
	if available != 3 {
		t.Fatal("failed generation lost credit")
	}
}
