package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	restore "nchc-vmbr/internal/restore"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("warning: failed to load .env: %v", err)
		}
	}

	cfg, err := restore.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	if err := restore.Run(context.Background(), cfg); err != nil {
		log.Fatalf("restore failed: %v", err)
	}
}
