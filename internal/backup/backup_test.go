package backup

import (
	util "nchc-vmbr/internal/util"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.BaseURL == "" || cfg.Token == "" || cfg.ProjectSysCode == "" || cfg.VMName == "" || cfg.RepoName == "" || cfg.CSBucket == "" {
		t.Fatalf("expected all config fields to be set, got %+v", cfg)
	}
}

func TestLoadConfigFromEnv_DefaultDateFormat(t *testing.T) {
	// Ensure DATE_TAG_FORMAT is not set
	os.Unsetenv("DATE_TAG_FORMAT")
	// Set required env vars
	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Default format is strftime %Y-%m-%d-%H-%M which corresponds to the previous layout
	// Ensure the DateTag is formed using util.ApplyStrftime with the default format
	wantDateTag := util.ApplyStrftime("%Y-%m-%d-%H-%M", cfg.Now)
	if cfg.DateTag != wantDateTag {
		t.Fatalf("expected DateTag to equal %s, got %s", wantDateTag, cfg.DateTag)
	}
}

func TestLoadConfigFromEnv_CustomDateFormat(t *testing.T) {
	// Set a custom date format (strftime-style)
	os.Setenv("DATE_TAG_FORMAT", "%Y-%m-%d")
	defer os.Unsetenv("DATE_TAG_FORMAT")

	// Set required env vars
	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// DATE_TAG_FORMAT should affect the resulting DateTag even though it is not stored in cfg
	wantDateTag := util.ApplyStrftime("%Y-%m-%d", cfg.Now)
	if cfg.DateTag != wantDateTag {
		t.Fatalf("expected DateTag to equal %s, got %s", wantDateTag, cfg.DateTag)
	}
}

func TestBuildCSFilepath_Strftime(t *testing.T) {
	// Fix the time used by nowFunc to be deterministic
	origNow := nowFunc
	nowFunc = func() time.Time { return time.Date(2025, 11, 22, 10, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)) }
	defer func() { nowFunc = origNow }()

	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")
	os.Setenv("BACKUP_IMAGE", "backup-%Y-%m-%d.img")
	defer os.Unsetenv("BACKUP_IMAGE")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := util.BuildCSFilepath(cfg.CSBucket, cfg.BackupRestoreImage, cfg.Now)
	want := "dss-public://my-bucket/backup-2025-11-22.img"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestBuildCSFilepath_BraceFormat(t *testing.T) {
	origNow := nowFunc
	nowFunc = func() time.Time { return time.Date(2025, 11, 22, 10, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)) }
	defer func() { nowFunc = origNow }()

	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")
	os.Setenv("BACKUP_IMAGE", "backup-{YYYY}-{MM}-{DD}.img")
	defer os.Unsetenv("BACKUP_IMAGE")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := util.BuildCSFilepath(cfg.CSBucket, cfg.BackupRestoreImage, cfg.Now)
	// Brace tokens shouldn't be converted, so they remain in the filename as literal characters.
	if !strings.Contains(got, "{YYYY}") {
		t.Fatalf("expected braces to remain literal in filename, got %s", got)
	}
	if !strings.HasSuffix(got, ".img") {
		t.Fatalf("expected filepath to end with .img, got %s", got)
	}
}

func TestBuildCSFilepath_DefaultFallback(t *testing.T) {
	// Brace-style token support has been removed; braces are left literal in filenames.
	origNow := nowFunc
	nowFunc = func() time.Time { return time.Date(2025, 11, 22, 10, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)) }
	defer func() { nowFunc = origNow }()

	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")
	os.Setenv("BACKUP_IMAGE", "backup-{YYYY}-{MM}-{DD}.img")
	defer os.Unsetenv("BACKUP_IMAGE")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := util.BuildCSFilepath(cfg.CSBucket, cfg.BackupRestoreImage, cfg.Now)
	if !strings.Contains(got, "{YYYY}") {
		t.Fatalf("expected braces to remain literal in filename, got %s", got)
	}
	if !strings.HasSuffix(got, ".img") {
		t.Fatalf("expected filepath to end with .img, got %s", got)
	}
}

func TestLoadConfigFromEnv_MissingEnv(t *testing.T) {
	// Do not set any environment variables to simulate missing required variables.
	os.Unsetenv("API_PROTOCOL")
	os.Unsetenv("API_HOST")
	os.Unsetenv("API_TOKEN")
	os.Unsetenv("PROJECT_SYS_CODE")
	os.Unsetenv("BACKUP_SRC_VM")
	os.Unsetenv("BACKUP_REPO")
	os.Unsetenv("BACKUP_CS_BUCKET")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected error due to missing envs, got nil")
	}
	msg := err.Error()
	// Check that the error lists some of the required env names
	if !strings.Contains(msg, "API_TOKEN") || !strings.Contains(msg, "BACKUP_REPO") {
		t.Fatalf("error did not list missing env names; got %s", msg)
	}
}

func TestLoadConfigFromEnv_TagNumParsing(t *testing.T) {
	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")
	os.Setenv("BACKUP_TAG_NUM", "5")
	defer os.Unsetenv("BACKUP_TAG_NUM")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.TagNum != 5 {
		t.Fatalf("expected TagNum to be 5, got %d", cfg.TagNum)
	}
}

func TestLoadConfigFromEnv_TransferDefaultFalse(t *testing.T) {
	// Default should be false and S3 config pointers should be nil when BACKUP_TRANSFR_TO_S3 is not set
	os.Unsetenv("BACKUP_TRANSFR_TO_S3")

	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.TransferS3 != false {
		t.Fatalf("expected TransferS3 to be false by default, got %v", cfg.TransferS3)
	}
	if cfg.SrcS3Cfg != nil || cfg.DstS3Cfg != nil {
		t.Fatalf("expected S3 configs to be nil when transfer disabled, got %+v %+v", cfg.SrcS3Cfg, cfg.DstS3Cfg)
	}
}

func TestLoadConfigFromEnv_TransferEnabled(t *testing.T) {
	// When BACKUP_TRANSFR_TO_S3=true and S3 env vars are present, pointers should be initialized
	os.Setenv("BACKUP_TRANSFR_TO_S3", "true")
	defer os.Unsetenv("BACKUP_TRANSFR_TO_S3")

	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")

	// S3 envs required when enabling transfer
	os.Setenv("BACKUP_SRC_S3_ENDPOINT", "https://src.example.com")
	defer os.Unsetenv("BACKUP_SRC_S3_ENDPOINT")
	os.Setenv("BACKUP_SRC_S3_ACCESS_KEY", "src-access")
	defer os.Unsetenv("BACKUP_SRC_S3_ACCESS_KEY")
	os.Setenv("BACKUP_SRC_S3_SECRET_KEY", "src-secret")
	defer os.Unsetenv("BACKUP_SRC_S3_SECRET_KEY")
	os.Setenv("BACKUP_SRC_S3_BUCKET", "src-bucket")
	defer os.Unsetenv("BACKUP_SRC_S3_BUCKET")

	os.Setenv("BACKUP_DST_S3_ENDPOINT", "https://dst.example.com")
	defer os.Unsetenv("BACKUP_DST_S3_ENDPOINT")
	os.Setenv("BACKUP_DST_S3_ACCESS_KEY", "dst-access")
	defer os.Unsetenv("BACKUP_DST_S3_ACCESS_KEY")
	os.Setenv("BACKUP_DST_S3_SECRET_KEY", "dst-secret")
	defer os.Unsetenv("BACKUP_DST_S3_SECRET_KEY")
	os.Setenv("BACKUP_DST_S3_BUCKET", "dst-bucket")
	defer os.Unsetenv("BACKUP_DST_S3_BUCKET")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.TransferS3 != true {
		t.Fatalf("expected TransferS3 to be true, got %v", cfg.TransferS3)
	}
	if cfg.SrcS3Cfg == nil || cfg.DstS3Cfg == nil {
		t.Fatalf("expected S3 configs to be initialized when transfer enabled, got %+v %+v", cfg.SrcS3Cfg, cfg.DstS3Cfg)
	}
	if cfg.SrcS3Cfg.Bucket != "src-bucket" || cfg.DstS3Cfg.Bucket != "dst-bucket" {
		t.Fatalf("unexpected bucket values: %+v %+v", cfg.SrcS3Cfg, cfg.DstS3Cfg)
	}
}

func TestTransfer_ReturnsErrorWhenNotConfigured(t *testing.T) {
	// Default config from env should not enable transfer so Transfer should return an informative error
	os.Unsetenv("BACKUP_TRANSFR_TO_S3")

	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("BACKUP_SRC_VM", "test-vm")
	defer os.Unsetenv("BACKUP_SRC_VM")
	os.Setenv("BACKUP_REPO", "snapshot-repo")
	defer os.Unsetenv("BACKUP_REPO")
	os.Setenv("BACKUP_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("BACKUP_CS_BUCKET")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error loading config, got %v", err)
	}

	if err := util.Transfer(cfg); err == nil {
		t.Fatalf("expected error when calling Transfer while not configured, got nil")
	}
}
