package main

import (
	"context"
	"imavara.local/server/internal/platform/vectors"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for {
		err := vectors.Bootstrap(ctx)
		if err == nil {
			return
		}
		log.Print(err)
		select {
		case <-ctx.Done():
			log.Fatal(ctx.Err())
		case <-time.After(4 * time.Second):
		}
	}
}
