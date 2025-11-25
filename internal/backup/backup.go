package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	config "nchc-vmbr/internal/config"
	rclone "nchc-vmbr/internal/rclone"
	util "nchc-vmbr/internal/util"

	cloudsdk "github.com/Zillaforge/cloud-sdk"
	vpsservers "github.com/Zillaforge/cloud-sdk/models/vps/servers"
	vrmrepos "github.com/Zillaforge/cloud-sdk/models/vrm/repositories"
	vrmtags "github.com/Zillaforge/cloud-sdk/models/vrm/tags"
	vrm "github.com/Zillaforge/cloud-sdk/modules/vrm/core"
)

// nowFunc can be overridden by tests for deterministic timestamp generation.
var nowFunc = time.Now

// backup.Config is now provided by internal/config.Config (shared struct)

// LoadConfigFromEnv loads configuration from environment variables. It returns an error if required
// variables are missing.
func LoadConfigFromEnv() (*config.Config, error) {
	// Validate required environment variables using the util helper. This
	// produces an error that explicitly names which variables are missing.
	if err := util.RequireEnv(
		"API_TOKEN", "API_PROTOCOL", "API_HOST", "PROJECT_SYS_CODE",
		"BACKUP_SRC_VM", "BACKUP_REPO", "BACKUP_CS_BUCKET"); err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("%s://%s", os.Getenv("API_PROTOCOL"), os.Getenv("API_HOST"))
	token := os.Getenv("API_TOKEN")
	projectSysCode := os.Getenv("PROJECT_SYS_CODE")
	vmName := os.Getenv("BACKUP_SRC_VM")
	repoName := os.Getenv("BACKUP_REPO")
	csBucket := os.Getenv("BACKUP_CS_BUCKET")

	// Use Taiwan timezone for tagging; fallback to fixed offset.
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Printf("warning: failed to load Asia/Taipei timezone: %v; falling back to fixed UTC+8", err)
		loc = time.FixedZone("UTC+8", 8*3600)
	}

	// Allow customizing the date tag format via environment variable DATE_TAG_FORMAT.
	// Accept a strftime-style format (e.g. %Y-%m-%d-%H-%M), convert and apply it using util.ApplyStrftime.
	// If not set, default to %Y-%m-%d-%H-%M to mimic the original layout 2006-01-02-15-04.
	dateTagFormat := os.Getenv("DATE_TAG_FORMAT")
	if dateTagFormat == "" {
		dateTagFormat = "%Y-%m-%d-%H-%M"
	}

	now := nowFunc().In(loc)
	dateTag := util.ApplyStrftime(dateTagFormat, now)

	// Default BACKUP_IMAGE is a strftime pattern; use ISO date-like pattern by default.
	backupImage := os.Getenv("BACKUP_IMAGE")
	if backupImage == "" {
		backupImage = "backup-%Y-%m-%d.img"
	}

	// Parse BACKUP_TAG_NUM
	tagNum := 2
	if v := os.Getenv("BACKUP_TAG_NUM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			tagNum = n
		}
	}

	// Read BACKUP_TRANSFR_TO_S3 env var — default to "false" if not set.
	transferFlag := false
	if v := os.Getenv("BACKUP_TRANSFR_TO_S3"); v != "" {
		// Accept common true-ish values (true, 1, yes, y)
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "y":
			transferFlag = true
		default:
			transferFlag = false
		}
	}

	var srcCfg rclone.S3Config
	var dstCfg rclone.S3Config
	var srcPtr *rclone.S3Config
	var dstPtr *rclone.S3Config

	// Only require and populate S3 configuration when transfer is enabled.
	if transferFlag {
		if err := util.RequireEnv(
			"BACKUP_SRC_S3_ENDPOINT", "BACKUP_SRC_S3_ACCESS_KEY", "BACKUP_SRC_S3_SECRET_KEY", "BACKUP_SRC_S3_BUCKET",
			"BACKUP_DST_S3_ENDPOINT", "BACKUP_DST_S3_ACCESS_KEY", "BACKUP_DST_S3_SECRET_KEY", "BACKUP_DST_S3_BUCKET",
		); err != nil {
			return nil, err
		}

		srcCfg = rclone.S3Config{
			Endpoint:  os.Getenv("BACKUP_SRC_S3_ENDPOINT"),
			AccessKey: os.Getenv("BACKUP_SRC_S3_ACCESS_KEY"),
			SecretKey: os.Getenv("BACKUP_SRC_S3_SECRET_KEY"),
			Bucket:    os.Getenv("BACKUP_SRC_S3_BUCKET"),
		}
		dstCfg = rclone.S3Config{
			Endpoint:  os.Getenv("BACKUP_DST_S3_ENDPOINT"),
			AccessKey: os.Getenv("BACKUP_DST_S3_ACCESS_KEY"),
			SecretKey: os.Getenv("BACKUP_DST_S3_SECRET_KEY"),
			Bucket:    os.Getenv("BACKUP_DST_S3_BUCKET"),
		}
		srcPtr = &srcCfg
		dstPtr = &dstCfg
	}

	cfg := &config.Config{
		BaseURL:            baseURL,
		Token:              token,
		ProjectSysCode:     projectSysCode,
		VMName:             vmName,
		RepoName:           repoName,
		CSBucket:           csBucket,
		OsType:             "linux",
		DateTag:            dateTag,
		BackupRestoreImage: backupImage,
		TagNum:             tagNum,
		Now:                now,
		SrcS3Cfg:           srcPtr,
		DstS3Cfg:           dstPtr,
		TransferS3:         transferFlag,
	}

	return cfg, nil
}

// Run performs the complete backup flow using the provided configuration.
func Run(ctx context.Context, cfg *config.Config) error {
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
		// If there is a tag retention policy defined, prune older tags first.
		if cfg.TagNum > 0 {
			// Prune the repository tags using the VRM client wrapper. The function
			// will query the repository's tag subresource and delete the oldest tags
			// if the configured limit is exceeded.
			if err := util.PruneRepositoryTags(ctx, vrmClient, repoID, cfg.TagNum-1); err != nil {
				return fmt.Errorf("failed to prune repository tags: %w", err)
			}
		}
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
		Filepath: util.BuildCSFilepath(cfg.CSBucket, cfg.BackupRestoreImage, cfg.Now),
	}

	if err := vrmClient.Tags().Download(ctx, tagID, downloadReq); err != nil {
		return fmt.Errorf("failed to export tag to S3: %w", err)
	}

	log.Println("Exported snapshot to S3 successfully")
	return nil
}
