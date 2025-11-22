package backup

import (
	"os"
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

	// Default format is "2006-01-02-15-04"; Try to parse DateTag
	if cfg.DateTagFormat != "2006-01-02-15-04" {
		t.Fatalf("expected default DateTagFormat to be 2006-01-02-15-04, got %s", cfg.DateTagFormat)
	}
	if _, err := time.ParseInLocation(cfg.DateTagFormat, cfg.DateTag, time.Local); err != nil {
		t.Fatalf("expected DateTag to parse with default format, got error: %v", err)
	}
}

func TestLoadConfigFromEnv_CustomDateFormat(t *testing.T) {
	// Set a custom date format
	os.Setenv("DATE_TAG_FORMAT", "2006-01-02")
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

	if cfg.DateTagFormat != "2006-01-02" {
		t.Fatalf("expected DateTagFormat to equal custom format, got %s", cfg.DateTagFormat)
	}
	if _, err := time.ParseInLocation(cfg.DateTagFormat, cfg.DateTag, time.Local); err != nil {
		t.Fatalf("expected DateTag to parse with custom format, got error: %v", err)
	}
}
