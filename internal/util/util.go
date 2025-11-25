package util

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	vrmtags "github.com/Zillaforge/cloud-sdk/models/vrm/tags"
	vrmcore "github.com/Zillaforge/cloud-sdk/modules/vrm/core"

	config "nchc-vmbr/internal/config"
	rclone "nchc-vmbr/internal/rclone"
)

// strftime-to-Go mappings
var strftimeMap = map[string]string{
	"%Y": "2006",
	"%y": "06",
	"%m": "01",
	"%d": "02",
	"%H": "15",
	"%M": "04",
	"%S": "05",
}

// ApplyStrftime converts a strftime-like format into Go's time layout, and then formats the time.
func ApplyStrftime(format string, t time.Time) string {
	// Build the Go time layout from the strftime-like format by walking
	// the string and substituting tokens while converting literal digit runs
	// into placeholders so they won't be misinterpreted as Go layout tokens.
	var b strings.Builder
	i := 0
	placeholderMap := make(map[string]string)
	// generate alphabetic placeholders (Z, ZZ, ZZZ, ...)
	var phIndex int
	nextPlaceholder := func() string {
		phIndex++
		return fmt.Sprintf("__LIT_%s__", strings.Repeat("Z", phIndex))
	}
	digitRe := regexp.MustCompile(`\d+`)
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			tok := format[i : i+2]
			if v, ok := strftimeMap[tok]; ok {
				b.WriteString(v)
				i += 2
				continue
			}
			// unknown %token: write it literally (quote it)
			b.WriteString("'" + tok + "'")
			i += 2
			continue
		}
		// accumulate a run of literal chars until next '%'
		start := i
		for i < len(format) && format[i] != '%' {
			i++
		}
		lit := format[start:i]
		if lit != "" {
			// Replace digit sequences in the literal with placeholders so
			// they are not accidentally parsed as Go layout tokens like "1".
			out := digitRe.ReplaceAllStringFunc(lit, func(s string) string {
				ph := nextPlaceholder()
				placeholderMap[ph] = s
				return ph
			})
			b.WriteString(out)
		}
	}
	layout := b.String()
	formatted := t.Format(layout)
	// Replace placeholders back to the original digit strings.
	for ph, val := range placeholderMap {
		formatted = strings.ReplaceAll(formatted, ph, val)
	}
	return formatted
}

// BuildCSFilepath returns the path dss-public://{bucket}/{filename}.
// filename may contain strftime tokens (e.g. %Y) which will be applied with time.Time t.
func BuildCSFilepath(bucket string, filename string, t time.Time) string {
	fname := filename
	if strings.Contains(fname, "%") {
		fname = ApplyStrftime(fname, t)
	}
	return fmt.Sprintf("dss-public://%s/%s", bucket, fname)
}

// RequireEnv verifies that the named environment variables are set and
// returns an error listing any missing variables. It is useful for
// command-specific validation where different commands require different
// environment variables.
func RequireEnv(names ...string) error {
	var missing []string
	for _, n := range names {
		if os.Getenv(n) == "" {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// PruneRepositoryTags ensures that the number of tags in the repository
// identified by repoID does not exceed maxTags. If there are more tags than
// maxTags, the oldest tags are deleted until the number of tags is <= maxTags.
func PruneRepositoryTags(ctx context.Context, vrmClient *vrmcore.Client, repoID string, maxTags int) error {
	if maxTags <= 0 {
		return nil
	}
	repoRes, err := vrmClient.Repositories().Get(ctx, repoID)
	if err != nil {
		return fmt.Errorf("failed to get repository %s: %w", repoID, err)
	}

	lister := repoRes.Tags()
	deleter := vrmClient.Tags()

	// Fetch repository subresource to list tags scoped to this repo.
	opts := &vrmtags.ListTagsOptions{Limit: -1}
	tags, err := lister.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) <= maxTags {
		return nil
	}
	// Sort by CreatedAt ascending (oldest first)
	sort.Slice(tags, func(i, j int) bool { return tags[i].CreatedAt.Before(tags[j].CreatedAt) })

	toDelete := len(tags) - maxTags
	for i := 0; i < toDelete; i++ {
		t := tags[i]
		if t == nil || t.ID == "" {
			continue
		}
		// Delete the tag; bubble up any error.

		if err := deleter.Delete(ctx, t.ID); err != nil {
			return fmt.Errorf("failed to delete tag %s: %w", t.ID, err)
		}
	}
	return nil
}

// Transfer performs S3 transfer using rclone for backup or restore operations.
func Transfer(cfg *config.Config) error {
	// Ensure transfer was enabled and S3 configs were initialized.
	if cfg == nil {
		return fmt.Errorf("nil config")
	}
	if !cfg.TransferS3 || cfg.SrcS3Cfg == nil || cfg.DstS3Cfg == nil {
		return fmt.Errorf("S3 transfer not configured; set RESTORE_TRANSFR_FROM_S3=true or BACKUP_TRANSFR_TO_S3=true and provide S3 configuration env vars to enable transfer")
	}

	fileName := cfg.BackupRestoreImage
	if strings.Contains(fileName, "%") {
		fileName = ApplyStrftime(fileName, cfg.Now)
	}

	dstRemote := fileName

	// Run transfer using rclone helper. Initialize librclone for this operation.
	rclone.Init()
	defer rclone.Close()

	jobID, totalSize, err := rclone.CopyFileAsync(*cfg.SrcS3Cfg, fileName, *cfg.DstS3Cfg, dstRemote)
	if err != nil {
		return fmt.Errorf("failed to start transfer job: %w", err)
	}

	ok, dur, err := rclone.WaitJob(jobID, totalSize, 5*time.Second, true)
	if err != nil {
		return fmt.Errorf("transfer job error: %w", err)
	}
	if !ok {
		return fmt.Errorf("transfer job failed after %.2fs", dur)
	}
	return nil
}
