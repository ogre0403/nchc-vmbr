package restore

import (
	"os"
	"testing"
	"time"

	util "nchc-vmbr/internal/util"
)

func TestLoadConfigFromEnv_Basic(t *testing.T) {
	os.Setenv("API_PROTOCOL", "https")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("DATE_TAG", "2025-01-02-15-04")
	defer os.Unsetenv("DATE_TAG")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.BaseURL == "" || cfg.Token == "" || cfg.ProjectSysCode == "" || cfg.RepoName == "" || cfg.BackupRestoreImage == "" || cfg.CSBucket == "" {
		t.Fatalf("expected config values to be populated, got %+v", cfg)
	}
}

func TestLoadConfigFromEnv_DefaultsAndPassword(t *testing.T) {
	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	// The current implementation requires RESTORE_CS_BUCKET and RESTORE_IMAGE
	// so provide both.
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "test.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")
	os.Unsetenv("RESTORE_PASSWORD_BASE64")
	os.Setenv("RESTORE_PASSWORD", "rawpwd")
	defer os.Unsetenv("RESTORE_PASSWORD")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.OsType != "linux" {
		t.Fatalf("expected default OsType linux, got %s", cfg.OsType)
	}
	if cfg.DateTag == "" {
		t.Fatalf("expected DateTag to be set, got empty")
	}

}

func TestLoadConfigFromEnv_DefaultDateFormat(t *testing.T) {
	origNow := nowFunc
	nowFunc = func() time.Time { return time.Date(2025, 11, 22, 10, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)) }
	defer func() { nowFunc = origNow }()

	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup-%Y-%m-%d.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// DateTagFormat field no longer exists for restore flow; assert DateTag is formatted with the default
	want := util.ApplyStrftime("%Y-%m-%d-%H-%M", nowFunc())
	if cfg.DateTag != want {
		t.Fatalf("expected DateTag to equal %s, got %s", want, cfg.DateTag)
	}
}

func TestLoadConfigFromEnv_MissingKeypairOrSG(t *testing.T) {
	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	// Intentionally omit RESTORE_KEYPAIR_ID and RESTORE_SECURITYGROUP_ID

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected error due to missing keypair and security group environment variables, got nil")
	}
}

func TestLoadConfigFromEnv_MissingFlavor(t *testing.T) {
	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	// Intentional: omit flavor
	os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected error due to missing flavor, got nil")
	}
}

func TestLoadConfigFromEnv_TagNumParsing(t *testing.T) {
	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")
	os.Setenv("RESTORE_TAG_NUM", "7")
	defer os.Unsetenv("RESTORE_TAG_NUM")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.TagNum != 7 {
		t.Fatalf("expected TagNum to be 7, got %d", cfg.TagNum)
	}
}

func TestLoadConfigFromEnv_TransferDefaultFalse(t *testing.T) {
	// Default should be false and S3 config pointers should be nil when RESTORE_TRANSFR_FROM_S3 is not set
	os.Unsetenv("RESTORE_TRANSFR_FROM_S3")

	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")

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
	// When RESTORE_TRANSFR_FROM_S3=true and S3 env vars are present, pointers should be initialized
	os.Setenv("RESTORE_TRANSFR_FROM_S3", "true")
	defer os.Unsetenv("RESTORE_TRANSFR_FROM_S3")

	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")

	// S3 envs required when enabling transfer
	os.Setenv("RESTORE_SRC_S3_ENDPOINT", "https://src.example.com")
	defer os.Unsetenv("RESTORE_SRC_S3_ENDPOINT")
	os.Setenv("RESTORE_SRC_S3_ACCESS_KEY", "src-access")
	defer os.Unsetenv("RESTORE_SRC_S3_ACCESS_KEY")
	os.Setenv("RESTORE_SRC_S3_SECRET_KEY", "src-secret")
	defer os.Unsetenv("RESTORE_SRC_S3_SECRET_KEY")
	os.Setenv("RESTORE_SRC_S3_BUCKET", "src-bucket")
	defer os.Unsetenv("RESTORE_SRC_S3_BUCKET")

	os.Setenv("RESTORE_DST_S3_ENDPOINT", "https://dst.example.com")
	defer os.Unsetenv("RESTORE_DST_S3_ENDPOINT")
	os.Setenv("RESTORE_DST_S3_ACCESS_KEY", "dst-access")
	defer os.Unsetenv("RESTORE_DST_S3_ACCESS_KEY")
	os.Setenv("RESTORE_DST_S3_SECRET_KEY", "dst-secret")
	defer os.Unsetenv("RESTORE_DST_S3_SECRET_KEY")
	os.Setenv("RESTORE_DST_S3_BUCKET", "dst-bucket")
	defer os.Unsetenv("RESTORE_DST_S3_BUCKET")

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
	os.Unsetenv("RESTORE_TRANSFR_FROM_S3")

	os.Setenv("API_PROTOCOL", "http")
	defer os.Unsetenv("API_PROTOCOL")
	os.Setenv("API_HOST", "api.example.com")
	defer os.Unsetenv("API_HOST")
	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("PROJECT_SYS_CODE", "proj-123")
	defer os.Unsetenv("PROJECT_SYS_CODE")
	os.Setenv("RESTORE_REPO", "rocky")
	defer os.Unsetenv("RESTORE_REPO")
	os.Setenv("RESTORE_CS_BUCKET", "my-bucket")
	defer os.Unsetenv("RESTORE_CS_BUCKET")
	os.Setenv("RESTORE_IMAGE", "backup.img")
	defer os.Unsetenv("RESTORE_IMAGE")
	os.Setenv("RESTORE_FLAVOR_ID", "flavor-1")
	defer os.Unsetenv("RESTORE_FLAVOR_ID")
	os.Setenv("RESTORE_NETWORK_ID", "net-1")
	defer os.Unsetenv("RESTORE_NETWORK_ID")
	os.Setenv("RESTORE_KEYPAIR_ID", "kp-1")
	defer os.Unsetenv("RESTORE_KEYPAIR_ID")
	os.Setenv("RESTORE_SECURITYGROUP_ID", "sg-1")
	defer os.Unsetenv("RESTORE_SECURITYGROUP_ID")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error loading config, got %v", err)
	}

	if err := util.Transfer(cfg); err == nil {
		t.Fatalf("expected error when calling Transfer while not configured, got nil")
	}
}
