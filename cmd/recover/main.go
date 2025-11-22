package main

import (
	"context"
	"log"

	recoverflow "cloud-sdk-sample/internal/recover"
)

func main() {
	cfg, err := recoverflow.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	if err := recoverflow.Run(context.Background(), cfg); err != nil {
		log.Fatalf("recover failed: %v", err)
	}
}
