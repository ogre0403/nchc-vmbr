package util

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
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
	// generate alphabetic placeholders (A, B, C, ...)
	var phIndex int
	nextPlaceholder := func() string {
		p := fmt.Sprintf("__LIT_%d__", phIndex)
		phIndex++
		return p
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
