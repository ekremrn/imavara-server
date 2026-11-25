package main

import (
	"flag"
	"fmt"
	"imavara.local/server/internal/workers"
	"log"
)

func main() {
	queue := flag.String("queue", "assets.extract.v1", "assets.extract.v1 or assets.index.v1")
	limit := flag.Int("limit", 1, "maximum failed jobs to replay")
	flag.Parse()
	count, err := workers.Replay(*queue, *limit)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Requeued %d failed jobs\n", count)
}
