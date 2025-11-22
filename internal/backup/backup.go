package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cloudsdk "github.com/Zillaforge/cloud-sdk"
	vpsservers "github.com/Zillaforge/cloud-sdk/models/vps/servers"
	vrmrepos "github.com/Zillaforge/cloud-sdk/models/vrm/repositories"
	vrmtags "github.com/Zillaforge/cloud-sdk/models/vrm/tags"
	vrm "github.com/Zillaforge/cloud-sdk/modules/vrm/core"
)

// Config holds configuration options for a backup run.
type Config struct {
	BaseURL        string
	Token          string
	ProjectSysCode string
	VMName         string
	RepoName       string
	CSBucket       string
	OsType         string
	DateTag        string
}

// LoadConfigFromEnv loads configuration from environment variables. It returns an error if required
// variables are missing.
func LoadConfigFromEnv() (*Config, error) {
	baseURL := fmt.Sprintf("%s://%s", os.Getenv("API_PROTOCOL"), os.Getenv("API_HOST"))
	token := os.Getenv("API_TOKEN")
	projectSysCode := os.Getenv("PROJECT_SYS_CODE")
	vmName := os.Getenv("SRC_VM")
	repoName := os.Getenv("SNAPSHOT_NAME")
	csBucket := os.Getenv("CS_BUCKET")

	if token == "" || projectSysCode == "" || vmName == "" || repoName == "" || csBucket == "" {
		return nil, fmt.Errorf("missing required environment variables (API_TOKEN, PROJECT_SYS_CODE, SRC_VM, SNAPSHOT_NAME, CS_BUCKET)")
	}

	// Use Taiwan timezone for tagging; fallback to fixed offset.
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Printf("warning: failed to load Asia/Taipei timezone: %v; falling back to fixed UTC+8", err)
		loc = time.FixedZone("UTC+8", 8*3600)
	}

	dateTag := time.Now().In(loc).Format("2006-01-02-15-04")

	cfg := &Config{
		BaseURL:        baseURL,
		Token:          token,
		ProjectSysCode: projectSysCode,
		VMName:         vmName,
		RepoName:       repoName,
		CSBucket:       csBucket,
		OsType:         "linux",
		DateTag:        dateTag,
	}

	return cfg, nil
}

// Run performs the complete backup flow using the provided configuration.
func Run(ctx context.Context, cfg *Config) error {
	client, err := cloudsdk.New(cfg.BaseURL, cfg.Token)
	if err != nil {
		return fmt.Errorf("failed to create SDK client: %w", err)
	}

	projClient, err := client.Project(ctx, cfg.ProjectSysCode)
	if err != nil {
		return fmt.Errorf("failed to create project client: %w", err)
	}

	vpsClient := projClient.VPS()
	servers, err := vpsClient.Servers().List(ctx, &vpsservers.ServersListRequest{Name: cfg.VMName})
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}
	if len(servers) == 0 {
		return fmt.Errorf("no server found with name %s", cfg.VMName)
	}
	vmID := servers[0].ID
	log.Printf("Found VM ID: %s", vmID)

	vrmClient := projClient.VRM()

	// Check repository
	repos, err := vrmClient.Repositories().List(ctx, &vrmrepos.ListRepositoriesOptions{})
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	var repoID string
	for _, r := range repos {
		if r == nil {
			continue
		}
		if r.Name == cfg.RepoName {
			repoID = r.ID
			break
		}
	}

	var snapshotResp *vrmrepos.CreateSnapshotResponse
	if repoID == "" {
		log.Println("Repository not found, creating snapshot (new repo)")
		req := &vrmrepos.CreateSnapshotFromNewRepositoryRequest{Name: cfg.RepoName, OperatingSystem: cfg.OsType, Version: cfg.DateTag}
		snapshotResp, err = vrmClient.Repositories().Snapshot(ctx, vmID, req)
		if err != nil {
			return fmt.Errorf("failed to create snapshot into new repository: %w", err)
		}
	} else {
		log.Println("Repository found, creating snapshot into existing repository")
		req := &vrmrepos.CreateSnapshotFromExistingRepositoryRequest{RepositoryID: repoID, Version: cfg.DateTag}
		snapshotResp, err = vrmClient.Repositories().Snapshot(ctx, vmID, req)
		if err != nil {
			return fmt.Errorf("failed to create snapshot into existing repository: %w", err)
		}
	}

	if snapshotResp == nil || snapshotResp.Tag == nil {
		return fmt.Errorf("snapshot response missing tag info")
	}

	tagID := snapshotResp.Tag.ID
	log.Printf("Snapshot created, repository ID: %s, tag ID: %s", snapshotResp.Repository.ID, tagID)

	// Wait for tag to become available.
	log.Printf("Waiting for tag %s to become available (SDK default wait options)...", tagID)
	if err := vrm.WaitForTagAvailable(ctx, vrmClient.Tags(), tagID); err != nil {
		return fmt.Errorf("tag %s did not become available: %w", tagID, err)
	}
	log.Printf("Tag %s is now available", tagID)

	// Export snapshot to CS
	downloadReq := &vrmtags.DownloadTagRequest{
		Filepath: fmt.Sprintf("dss-public://%s/backup-%s.img", cfg.CSBucket, cfg.DateTag),
	}

	if err := vrmClient.Tags().Download(ctx, tagID, downloadReq); err != nil {
		return fmt.Errorf("failed to export tag to S3: %w", err)
	}

	log.Println("Exported snapshot to S3 successfully")
	return nil
}
