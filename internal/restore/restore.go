package restore

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cloudsdk "github.com/Zillaforge/cloud-sdk"
	vrmrepos "github.com/Zillaforge/cloud-sdk/models/vrm/repositories"

	util "cloud-sdk-sample/internal/util"

	vpsservers "github.com/Zillaforge/cloud-sdk/models/vps/servers"
	vps "github.com/Zillaforge/cloud-sdk/modules/vps/core"
	vrm "github.com/Zillaforge/cloud-sdk/modules/vrm/core"
)

var nowFunc = time.Now

// Config for restore operation
type Config struct {
	BaseURL        string
	Token          string
	ProjectSysCode string
	RepoName       string // repository/image name
	ImageFilePath  string // dss-public://bucket/path/file.img
	FlavorID       string
	NetworkID      string
	KeypairID      string
	SecurityGroup  string
	VMNamePrefix   string
	DateTag        string
	OsType         string
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	if err := util.RequireEnv(
		"API_TOKEN", "API_PROTOCOL", "API_HOST", "PROJECT_SYS_CODE",
		"RESTORE_REPO", "RESTORE_CS_BUCKET", "RESTORE_IMAGE",
		"RESTORE_FLAVOR_ID", "RESTORE_NETWORK_ID",
		"RESTORE_KEYPAIR_ID", "RESTORE_SECURITYGROUP_ID",
	); err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("%s://%s", os.Getenv("API_PROTOCOL"), os.Getenv("API_HOST"))
	token := os.Getenv("API_TOKEN")
	projectSysCode := os.Getenv("PROJECT_SYS_CODE")

	repoName := os.Getenv("RESTORE_REPO")

	// Required runtime configuration: all must be provided via environment vars
	flavorID := os.Getenv("RESTORE_FLAVOR_ID")
	networkID := os.Getenv("RESTORE_NETWORK_ID")
	sgID := os.Getenv("RESTORE_SECURITYGROUP_ID")
	keypairID := os.Getenv("RESTORE_KEYPAIR_ID")

	vmNamePrefix := os.Getenv("RESTORE_DST_VM")
	if vmNamePrefix == "" {
		vmNamePrefix = "restore-dst-vm"
	}

	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Printf("warning: failed to load Asia/Taipei timezone: %v; falling back to fixed UTC+8", err)
		loc = time.FixedZone("UTC+8", 8*3600)
	}
	// restore flow does not need a DateTagFormat; DateTag is computed directly
	dateTagFormat := os.Getenv("DATE_TAG_FORMAT")
	if dateTagFormat == "" {
		dateTagFormat = "%Y-%m-%d-%H-%M"
	}

	// compute current time and date tag after timezone loc is available
	now := nowFunc().In(loc)
	dateTag := util.ApplyStrftime(dateTagFormat, now)
	imagePath := util.BuildCSFilepath(os.Getenv("RESTORE_CS_BUCKET"), os.Getenv("RESTORE_IMAGE"), now)

	cfg := &Config{
		BaseURL:        baseURL,
		Token:          token,
		ProjectSysCode: projectSysCode,
		RepoName:       repoName,
		ImageFilePath:  imagePath,
		FlavorID:       flavorID,
		KeypairID:      keypairID,
		NetworkID:      networkID,
		SecurityGroup:  sgID,
		VMNamePrefix:   vmNamePrefix,
		DateTag:        dateTag,
		OsType:         "linux",
	}
	return cfg, nil
}

// Run executes the restore workflow.
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
	vrmClient := projClient.VRM()

	// Check repository presence
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

	// Upload image to repository (create repo or add tag)
	var uploadResp *vrmrepos.UploadImageResponse
	if repoID == "" {
		log.Printf("Repository %s not found; creating and uploading image", cfg.RepoName)
		req := &vrmrepos.UploadToNewRepositoryRequest{
			Name:            cfg.RepoName,
			OperatingSystem: cfg.OsType,
			Description:     "restore upload",
			Version:         cfg.DateTag,
			Type:            "common",
			DiskFormat:      "raw",
			ContainerFormat: "bare",
			Filepath:        cfg.ImageFilePath,
		}

		uploadResp, err = vrmClient.Repositories().Upload(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to upload image to create repository: %w", err)
		}
		repoID = uploadResp.Repository.ID
	} else {
		log.Printf("Repository %s found; creating new tag version %s", cfg.RepoName, cfg.DateTag)
		req := &vrmrepos.UploadToExistingRepositoryRequest{
			RepositoryID:    repoID,
			Version:         cfg.DateTag,
			Type:            "common",
			DiskFormat:      "raw",
			ContainerFormat: "bare",
			Filepath:        cfg.ImageFilePath,
		}
		uploadResp, err = vrmClient.Repositories().Upload(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to upload image into existing repository: %w", err)
		}
	}

	if uploadResp == nil || uploadResp.Tag == nil {
		return fmt.Errorf("upload returned missing tag info")
	}
	tagID := uploadResp.Tag.ID
	log.Printf("Uploaded image, repoID=%s tagID=%s", repoID, tagID)

	// Wait for tag to become active
	log.Printf("Waiting for tag %s to become active...", tagID)
	if err := vrm.WaitForTagActive(ctx, vrmClient.Tags(), tagID); err != nil {
		return fmt.Errorf("tag %s did not become available: %w", tagID, err)
	}
	log.Printf("Tag %s is active", tagID)

	vmName := fmt.Sprintf("%s-%s", cfg.VMNamePrefix, cfg.DateTag)

	// Create VM from tag
	createReq := &vpsservers.ServerCreateRequest{
		Name:      vmName,
		ImageID:   tagID,
		FlavorID:  cfg.FlavorID,
		KeypairID: cfg.KeypairID,
		NICs: []vpsservers.ServerNICCreateRequest{
			{
				NetworkID: cfg.NetworkID,
				SGIDs:     []string{cfg.SecurityGroup},
			},
		},
	}

	created, err := vpsClient.Servers().Create(ctx, createReq)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Wait for the server to become active using SDK waiter.
	serverID := created.ID
	log.Printf("Waiting for server %s to become active...", serverID)
	if err := vps.WaitForServerActive(ctx, vpsClient.Servers(), serverID); err != nil {
		return fmt.Errorf("server %s did not become active: %w", serverID, err)
	}

	log.Printf("VM %s created successfully", vmName)
	return nil
}
