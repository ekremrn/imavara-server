package main

import (
	"imavara.local/server/internal/platform/observe"
	"imavara.local/server/internal/workers"
	"log"
)

func main() {
	observe.Start("image-extract-worker")
	log.Fatal(workers.Run(true))
}
