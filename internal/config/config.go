package config

import (
	"time"

	"nchc-vmbr/internal/rclone"
)

type VPSSetting struct {
	FlavorID        string
	NetworkID       string
	KeypairID       string
	SecurityGroupID string
}

// Config is a shared configuration struct used by different commands.
// It intentionally contains the superset of fields used by both the
// backup and restore workflows so callers can migrate gradually.
type Config struct {
	BaseURL        string
	Token          string
	ProjectSysCode string

	RepoName string
	CSBucket string

	// Backup/restore image names
	BackupRestoreImage string

	// VM-related fields (restore-specific)
	VPSSetting *VPSSetting
	VMName     string

	DateTag string
	OsType  string

	TagNum int
	Now    time.Time

	SrcS3Cfg *rclone.S3Config
	DstS3Cfg *rclone.S3Config

	// Transfer flags (kept separate for a staged migration)
	TransferS3 bool
}
