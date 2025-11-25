package config_test

import (
	"testing"
	"time"

	cfg "nchc-vmbr/internal/config"
)

func TestConfigInstantiation(t *testing.T) {
	c := &cfg.Config{
		BaseURL: "https://api.example",
		Token:   "abc123",
		Now:     time.Now(),
	}
	if c.BaseURL == "" || c.Token == "" {
		t.Fatalf("expected config to be initialized; got %+v", c)
	}
}
