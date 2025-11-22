package backup

import (
	"os"
	"testing"
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
	os.Setenv("SRC_VM", "test-vm")
	defer os.Unsetenv("SRC_VM")
	os.Setenv("SNAPSHOT_NAME", "snapshot-repo")
	defer os.Unsetenv("SNAPSHOT_NAME")
	os.Setenv("CS_BUCKET", "my-bucket")
	defer os.Unsetenv("CS_BUCKET")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.BaseURL == "" || cfg.Token == "" || cfg.ProjectSysCode == "" || cfg.VMName == "" || cfg.RepoName == "" || cfg.CSBucket == "" {
		t.Fatalf("expected all config fields to be set, got %+v", cfg)
	}
}
