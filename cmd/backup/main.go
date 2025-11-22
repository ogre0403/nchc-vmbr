package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"cloud-sdk-sample/internal/backup"
)

func main() {
	// Load .env (if present) and environment variables
	if err := godotenv.Load(); err != nil {
		// If .env not found that's okay: we'll still use environment variables
		if !os.IsNotExist(err) {
			log.Printf("warning: failed to load .env: %v", err)
		}
	}

	// Load configuration from environment variables
	cfg, err := backup.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	ctx := context.Background()

	if err := backup.Run(ctx, cfg); err != nil {
		log.Fatalf("backup failed: %v", err)
	}
}
