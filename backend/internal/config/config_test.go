package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("default port = %d, want 8080", cfg.Port)
	}
	if cfg.RateLimitRPM != 50 {
		t.Errorf("default rpm = %d, want 50", cfg.RateLimitRPM)
	}
	if cfg.MaxPages != 10 {
		t.Errorf("default max pages = %d, want 10", cfg.MaxPages)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("TAGSHA_PORT", "9090")
	t.Setenv("TAGSHA_RATE_LIMIT_RPM", "100")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("port = %d, want 9090", cfg.Port)
	}
	if cfg.RateLimitRPM != 100 {
		t.Errorf("rpm = %d, want 100", cfg.RateLimitRPM)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	os.Setenv("TAGSHA_PORT", "notanumber")
	defer os.Unsetenv("TAGSHA_PORT")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid port, got nil")
	}
}
