package restore

import (
	"os"
	"testing"
	"time"

	util "cloud-sdk-sample/internal/util"
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

	if cfg.BaseURL == "" || cfg.Token == "" || cfg.ProjectSysCode == "" || cfg.RepoName == "" || cfg.ImageFilePath == "" {
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
	// (RESTORE_IMAGE_PATH alone is not accepted), so provide both.
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
	defer os.Unsetenv("RESTORE_IMAGE")
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
	defer os.Unsetenv("RESTORE_IMAGE")
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
