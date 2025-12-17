package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"nchc-vmbr/internal/backup"
	"nchc-vmbr/internal/rclone"
	"nchc-vmbr/internal/util"
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

	// If transfer is not configured or no source S3 is provided, skip transfer.
	if cfg.SrcS3Cfg == nil || !cfg.TransferS3 {
		log.Println("Transfer disabled (no source S3 config or transfer flag off); skipping transfer")
		return
	}

	// Wait for the source object to exist before starting the transfer.
	const waitTimeout = 5 * time.Minute
	const pollInterval = 5 * time.Second
	deadline := time.Now().Add(waitTimeout)
	fileName := util.ApplyStrftime(cfg.BackupRestoreImage, cfg.Now)
	for {
		exists, err := rclone.ObjectExists(*cfg.SrcS3Cfg, fileName)
		if err != nil {
			log.Fatalf("failed to check source object existence: %v", err)
		}
		if exists {
			break
		}
		if time.Now().After(deadline) {
			log.Fatalf("timed out waiting for source object to appear: %s", fileName)
		}
		time.Sleep(pollInterval)
	}

	// Transfer the exported image from the CS bucket to the destination S3
	if err := util.Transfer(cfg); err != nil {
		log.Fatalf("failed to transfer exported image: %v", err)
	}

	log.Println("Transferred exported snapshot to destination S3 successfully")
}
