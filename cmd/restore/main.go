package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"nchc-vmbr/internal/rclone"
	restore "nchc-vmbr/internal/restore"
	"nchc-vmbr/internal/util"
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

	// Transfer the exported image from the CS bucket to the destination S3
	if err := util.Transfer(cfg); err != nil {
		log.Fatalf("failed to transfer exported image: %v", err)
	}

	const waitTimeout = 5 * time.Minute
	const pollInterval = 5 * time.Second
	deadline := time.Now().Add(waitTimeout)
	fileName := util.ApplyStrftime(cfg.BackupRestoreImage, cfg.Now)
	for {
		exists, err := rclone.ObjectExists(*cfg.DstS3Cfg, fileName)
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

	log.Println("Transferred exported snapshot to destination S3 successfully")

	if err := restore.Run(context.Background(), cfg); err != nil {
		log.Fatalf("restore failed: %v", err)
	}
}
