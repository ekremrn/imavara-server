package main

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"imavara.local/server/internal/api"
	"imavara.local/server/internal/credits"
	"imavara.local/server/internal/platform/observe"
	"imavara.local/server/migrations"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	observe.Start("api")

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = migrations.Apply(ctx, db); err != nil {
		log.Fatal(err)
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			deadline, end := context.WithTimeout(context.Background(), 10*time.Second)
			if err := (credits.Ledger{DB: db}).ReleaseExpired(deadline); err != nil {
				log.Print(err)
			}
			end()
			<-ticker.C
		}
	}()
	srv := http.Server{Addr: ":8080", Handler: api.Server{DB: db}.Routes(), ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(srv.ListenAndServe())
}
