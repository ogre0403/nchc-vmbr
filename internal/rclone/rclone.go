package rclone

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/rclone/rclone/backend/s3"    // import s3 backend
	_ "github.com/rclone/rclone/fs/operations" // import operations
	"github.com/rclone/rclone/librclone/librclone"
)

// rpc is a package-level RPC function wrapper; tests may override this
// to provide deterministic responses.
var rpc = librclone.RPC

// S3Config holds the minimal credential config for an S3 backend.
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

// Init initializes the librclone runtime.
func Init() {
	librclone.Initialize()
}

// Close finalizes the librclone runtime.
func Close() {
	librclone.Finalize()
}

// BuildS3Fs builds an rclone "fs" configuration string for the s3 backend.
// Example:
//
//	:s3,provider=Other,endpoint='https://host',access_key_id=abc,secret_access_key=xyz,env_auth=false
func BuildS3Fs(cfg S3Config) string {
	// We quote the endpoint to preserve characters like ':' in the RPC encoding
	return fmt.Sprintf(":s3,provider=Other,endpoint='%s',access_key_id=%s,secret_access_key=%s,env_auth=false", cfg.Endpoint, cfg.AccessKey, cfg.SecretKey)
}

// CopyFileAsync starts a copy job via rclone's operations/copyfile RPC in async mode,
// returns the job ID and the source size (if available, -1 when unknown).
func CopyFileAsync(src S3Config, srcRemote string, dst S3Config, dstRemote string) (int64, int64, error) {
	// Attempt to query remote size up-front so callers can show progress as a percentage.
	totalSize, _ := GetRemoteSize(src, srcRemote)
	srcFs := BuildS3Fs(src) + ":" + src.Bucket
	dstFs := BuildS3Fs(dst) + ":" + dst.Bucket

	req := struct {
		SrcFs     string `json:"srcFs"`
		SrcRemote string `json:"srcRemote"`
		DstFs     string `json:"dstFs"`
		DstRemote string `json:"dstRemote"`
		Async     bool   `json:"_async"`
	}{
		SrcFs:     srcFs,
		SrcRemote: srcRemote,
		DstFs:     dstFs,
		DstRemote: dstRemote,
		Async:     true,
	}

	b, _ := json.Marshal(req)
	out, status := rpc("operations/copyfile", string(b))
	if status != 200 {
		return 0, -1, fmt.Errorf("RPC call failed (status %d): %s", status, out)
	}

	var resp struct {
		JobId int64 `json:"jobid"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return 0, -1, fmt.Errorf("failed to parse operations/copyfile response: %w", err)
	}
	return resp.JobId, totalSize, nil
}

// GetRemoteSize returns the size of a remote object (if available) using operations/stat.
// Returns -1 when size can't be determined or an error occurs.
func GetRemoteSize(src S3Config, remote string) (int64, error) {
	req := struct {
		Fs     string `json:"fs"`
		Remote string `json:"remote"`
	}{Fs: BuildS3Fs(src) + ":" + src.Bucket, Remote: remote}
	b, _ := json.Marshal(req)
	out, status := rpc("operations/stat", string(b))
	if status != 200 {
		return -1, fmt.Errorf("operations/stat failed (status %d): %s", status, out)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return -1, fmt.Errorf("failed to parse operations/stat response: %w", err)
	}
	if item, ok := parsed["item"].(map[string]interface{}); ok {
		if v, found := item["Size"]; found {
			switch t := v.(type) {
			case float64:
				return int64(t), nil
			case string:
				if n, err := strconv.ParseInt(t, 10, 64); err == nil {
					return n, nil
				}
			}
		}
	}
	return -1, nil
}

// WaitJob polls rclone job status until it finishes. Poll interval is configurable.
// If totalSize > 0, progress will be printed as percentage complete instead of raw bytes.
// It returns (success, durationSeconds, error)
func WaitJob(jobID int64, totalSize int64, pollInterval time.Duration, showProgress bool) (bool, float64, error) {
	statusReq := struct {
		JobId int64 `json:"jobid"`
	}{JobId: jobID}
	statusReqBytes, _ := json.Marshal(statusReq)

	for {
		time.Sleep(pollInterval)

		out, status := rpc("job/status", string(statusReqBytes))
		if status != 200 {
			// Keep polling on transient errors
			log.Printf("warning: job/status returned status %d: %s", status, out)
			continue
		}

		var jobStatus struct {
			Finished bool    `json:"finished"`
			Success  bool    `json:"success"`
			Error    string  `json:"error"`
			Duration float64 `json:"duration"`
		}
		if err := json.Unmarshal([]byte(out), &jobStatus); err != nil {
			log.Printf("warning: failed to parse job/status: %v", err)
			continue
		}

		if jobStatus.Finished {
			if jobStatus.Success {
				return true, jobStatus.Duration, nil
			}
			return false, jobStatus.Duration, fmt.Errorf("job failed: %s", jobStatus.Error)
		}

		if showProgress {
			statsOut, statsStatus := rpc("core/stats", "{}")
			if statsStatus == 200 {
				var stats struct {
					Bytes int64   `json:"bytes"`
					Speed float64 `json:"speed"`
				}
				if err := json.Unmarshal([]byte(statsOut), &stats); err == nil {
					if totalSize > 0 {
						pct := (float64(stats.Bytes) / float64(totalSize)) * 100.0
						if pct > 100.0 {
							pct = 100.0
						}
						log.Printf("copy progress: %.1f%% complete, Speed: %.2f MB/s", pct, stats.Speed/1024/1024)
					} else {
						log.Printf("copy progress: speed=%.2f MB/s", stats.Speed/1024/1024)
					}
				}
			}
		}
	}
}

// ObjectExists checks whether an object at the given remote path exists
// in the specified S3 configuration. It returns true when the object
// is present, false when it is not present, or an error if an RPC
// failure occurs that doesn't clearly indicate absence.
func ObjectExists(cfg S3Config, remote string) (bool, error) {
	req := struct {
		Fs     string `json:"fs"`
		Remote string `json:"remote"`
	}{Fs: BuildS3Fs(cfg) + ":" + cfg.Bucket, Remote: remote}
	b, _ := json.Marshal(req)
	out, status := rpc("operations/stat", string(b))
	if status == 200 {
		// Parse and ensure item exists
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			return false, fmt.Errorf("failed to parse operations/stat response: %w", err)
		}
		if item, ok := parsed["item"].(map[string]interface{}); ok && item != nil {
			return true, nil
		}
		return false, nil
	}

	// If the RPC indicates not found, treat as non-existent rather than fatal.
	low := strings.ToLower(out)
	if strings.Contains(low, "not found") || strings.Contains(low, "no such file") || strings.Contains(low, "does not exist") {
		return false, nil
	}

	return false, fmt.Errorf("operations/stat failed (status %d): %s", status, out)
}
