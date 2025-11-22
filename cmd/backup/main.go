package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vrm "github.com/Zillaforge/cloud-sdk/modules/vrm/core"

	"github.com/joho/godotenv"

	cloudsdk "github.com/Zillaforge/cloud-sdk"
	vpsservers "github.com/Zillaforge/cloud-sdk/models/vps/servers"
	vrmrepos "github.com/Zillaforge/cloud-sdk/models/vrm/repositories"
	vrmtags "github.com/Zillaforge/cloud-sdk/models/vrm/tags"
)

func main() {
	// Load .env (if present) and environment variables
	if err := godotenv.Load(); err != nil {
		// If .env not found that's okay: we'll still use environment variables
		if !os.IsNotExist(err) {
			log.Printf("warning: failed to load .env: %v", err)
		}
	}

	// Required inputs from environment variables
	baseURL := fmt.Sprintf("%s://%s", os.Getenv("API_PROTOCOL"), os.Getenv("API_HOST"))
	token := os.Getenv("API_TOKEN")
	project_sys_code := os.Getenv("PROJECT_SYS_CODE")
	vmName := os.Getenv("SRC_VM")
	repoName := os.Getenv("SNAPSHOT_NAME")
	csBucket := os.Getenv("CS_BUCKET")

	if token == "" {
		log.Fatalf("API_TOKEN is required in .env or environment variables")
	}

	if project_sys_code == "" {
		log.Fatalf("PROJECT_SYS_CODE is required in .env or environment variables")
	}

	if vmName == "" {
		log.Fatalf("SRC_VM is required in .env or environment variables")
	}

	if repoName == "" {
		log.Fatalf("SNAPSHOT_NAME is required in .env or environment variables")
	}

	if csBucket == "" {
		log.Fatalf("CS_BUCKET is required in .env or environment variables")
	}

	// Include hours and minutes in the date tag so that versions are unique per snapshot
	// Use UTC+8 timezone (Asia/Taipei) for tag timestamps; fallback to fixed UTC+8 if tz DB not available.
	// Format: YYYY-MM-DD-HH-MM (e.g., 2025-11-22-14-30)
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Printf("warning: failed to load Asia/Taipei timezone: %v; falling back to fixed UTC+8", err)
		loc = time.FixedZone("UTC+8", 8*3600)
	}
	dateTag := time.Now().In(loc).Format("2006-01-02-15-04")

	osType := "linux"

	ctx := context.Background()

	// Create SDK client
	client, err := cloudsdk.New(baseURL, token)
	if err != nil {
		log.Fatalf("failed to create SDK client: %v", err)
	}

	// Create project-scoped client
	projClient, err := client.Project(ctx, project_sys_code)
	if err != nil {
		log.Fatalf("failed to create project client: %v", err)
	}

	// Find VM
	vpsClient := projClient.VPS()
	servers, err := vpsClient.Servers().List(ctx, &vpsservers.ServersListRequest{Name: vmName})
	if err != nil {
		log.Fatalf("failed to list servers: %v", err)
	}
	if len(servers) == 0 {
		log.Fatalf("no server found with name %s", vmName)
	}
	vmID := servers[0].ID
	fmt.Printf("Found VM ID: %s\n", vmID)

	// VRM client
	vrmClient := projClient.VRM()

	// Check repository existence
	repos, err := vrmClient.Repositories().List(ctx, &vrmrepos.ListRepositoriesOptions{})
	if err != nil {
		log.Fatalf("failed to list repositories: %v", err)
	}

	var repoID string
	for _, r := range repos {
		if r == nil {
			continue
		}
		if r.Name == repoName {
			repoID = r.ID
			break
		}
	}

	// Create or update snapshot
	var snapshotResp *vrmrepos.CreateSnapshotResponse
	if repoID == "" {
		fmt.Println("Repository not found, creating snapshot (new repo)")
		req := &vrmrepos.CreateSnapshotFromNewRepositoryRequest{Name: repoName, OperatingSystem: osType, Version: dateTag}
		snapshotResp, err = vrmClient.Repositories().Snapshot(ctx, vmID, req)
		if err != nil {
			log.Fatalf("failed to create snapshot into new repository: %v", err)
		}
	} else {
		fmt.Println("Repository found, creating snapshot into existing repository")
		req := &vrmrepos.CreateSnapshotFromExistingRepositoryRequest{RepositoryID: repoID, Version: dateTag}
		snapshotResp, err = vrmClient.Repositories().Snapshot(ctx, vmID, req)
		if err != nil {
			log.Fatalf("failed to create snapshot into existing repository: %v", err)
		}
	}

	if snapshotResp == nil || snapshotResp.Tag == nil {
		log.Fatalf("snapshot response missing tag info")
	}

	tagID := snapshotResp.Tag.ID
	fmt.Printf("Snapshot created, repository ID: %s, tag ID: %s\n", snapshotResp.Repository.ID, tagID)

	// Wait for tag to become available before attempting to download/export
	fmt.Printf("Waiting for tag %s to become available (SDK default wait options)...\n", tagID)
	// Use SDK-provided waiter to wait for the tag status to become AVAILABLE.
	if err := vrm.WaitForTagAvailable(ctx, vrmClient.Tags(), tagID); err != nil {
		log.Fatalf("tag %s did not become available: %v", tagID, err)
	}
	fmt.Printf("Tag %s is now available\n", tagID)

	// Export snapshot to CS
	downloadReq := &vrmtags.DownloadTagRequest{
		Filepath: fmt.Sprintf("dss-public://%s/backup-%s.img", csBucket, dateTag),
	}

	if err := vrmClient.Tags().Download(ctx, tagID, downloadReq); err != nil {
		log.Fatalf("failed to export tag to S3: %v", err)
	}

	fmt.Println("Exported snapshot to S3 successfully")
}
