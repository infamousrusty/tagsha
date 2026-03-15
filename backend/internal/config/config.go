package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration.
type Config struct {
	Port             int
	RedisURL         string
	CacheTTL         time.Duration
	RateLimitRPM     int
	LogLevel         string
	Env              string
	GitHubToken      string
	GitHubAPIBaseURL string
	MaxPages         int
	RequestSizeLimit int64
	Domain           string
}

// Load reads configuration from environment variables and Docker secrets.
func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("TAGSHA_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid TAGSHA_PORT: %w", err)
	}

	ttlSec, err := strconv.Atoi(getEnv("TAGSHA_CACHE_TTL_SECONDS", "300"))
	if err != nil {
		return nil, fmt.Errorf("invalid TAGSHA_CACHE_TTL_SECONDS: %w", err)
	}

	rpm, err := strconv.Atoi(getEnv("TAGSHA_RATE_LIMIT_RPM", "50"))
	if err != nil {
		return nil, fmt.Errorf("invalid TAGSHA_RATE_LIMIT_RPM: %w", err)
	}

	maxPages, err := strconv.Atoi(getEnv("TAGSHA_MAX_PAGES", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid TAGSHA_MAX_PAGES: %w", err)
	}

	token := readSecret("TAGSHA_GITHUB_TOKEN", "/run/secrets/github_token")

	return &Config{
		Port:             port,
		RedisURL:         getEnv("TAGSHA_REDIS_URL", "redis://localhost:6379/0"),
		CacheTTL:         time.Duration(ttlSec) * time.Second,
		RateLimitRPM:     rpm,
		LogLevel:         getEnv("TAGSHA_LOG_LEVEL", "info"),
		Env:              getEnv("TAGSHA_ENV", "development"),
		GitHubToken:      token,
		GitHubAPIBaseURL: "https://api.github.com",
		MaxPages:         maxPages,
		RequestSizeLimit: 65536,
		Domain:           getEnv("TAGSHA_DOMAIN", "localhost"),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// readSecret reads a secret from an env var first, then a Docker secret file.
func readSecret(envKey, secretPath string) string {
	if v := os.Getenv(envKey); v != "" {
		return strings.TrimSpace(v)
	}
	if data, err := os.ReadFile(secretPath); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}
