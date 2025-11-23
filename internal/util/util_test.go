package util

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestApplyStrftime_LiteralDigits(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*3600)
	now := time.Date(2025, 11, 23, 10, 18, 0, 0, loc)
	in := "backup-%Y-%m-%d-%H-18.img"
	got := ApplyStrftime(in, now)
	want := "backup-2025-11-23-10-18.img"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestApplyStrftime_Default(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*3600)
	now := time.Date(2025, 11, 23, 10, 18, 30, 0, loc)
	in := "%Y-%m-%d-%H-%M"
	got := ApplyStrftime(in, now)
	want := "2025-11-23-10-18"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestRequireEnv_AllPresent(t *testing.T) {
	os.Setenv("FOO", "1")
	defer os.Unsetenv("FOO")
	os.Setenv("BAR", "2")
	defer os.Unsetenv("BAR")

	if err := RequireEnv("FOO", "BAR"); err != nil {
		t.Fatalf("expected no error when envs present, got %v", err)
	}
}

func TestRequireEnv_Missing(t *testing.T) {
	os.Unsetenv("MISSING_1")
	os.Unsetenv("MISSING_2")
	os.Setenv("PRESENT", "1")
	defer os.Unsetenv("PRESENT")

	err := RequireEnv("PRESENT", "MISSING_1", "MISSING_2")
	if err == nil {
		t.Fatalf("expected error for missing envs, got nil")
	}
	// ensure both missing entries are present in the error message
	msg := err.Error()
	if !strings.Contains(msg, "MISSING_1") || !strings.Contains(msg, "MISSING_2") {
		t.Fatalf("error message did not include missing vars: %s", msg)
	}
}
