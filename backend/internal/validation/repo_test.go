package validation_test

import (
	"strings"
	"testing"

	"github.com/infamousrusty/tagsha/internal/validation"
)

func TestParseInput_Valid(t *testing.T) {
	cases := []struct {
		input string
		owner string
		repo  string
	}{
		// --- Plain owner/repo ---------------------------------------------------
		{"torvalds/linux", "torvalds", "linux"},
		{"rust-lang/rust", "rust-lang", "rust"},
		{"owner/repo.name", "owner", "repo.name"},
		{"owner/repo-name", "owner", "repo-name"},
		{"owner/repo_name", "owner", "repo_name"},
		{"A/B", "A", "B"},

		// --- Plain with tag (tag stripped) --------------------------------------
		{"owner/repo:latest", "owner", "repo"},
		{"owner/repo:v1.2.3", "owner", "repo"},
		{"owner/repo:sha-abc1234", "owner", "repo"},

		// --- GitHub web URLs ----------------------------------------------------
		{"https://github.com/golang/go", "golang", "go"},
		{"https://github.com/rust-lang/rust", "rust-lang", "rust"},
		{"github.com/infamousrusty/tagsha", "infamousrusty", "tagsha"},
		{"https://github.com/owner/repo.git", "owner", "repo"},
		{"https://github.com/owner/repo/tree/main", "owner", "repo"},
		{"https://github.com/owner/repo/releases/tag/v1.0.0", "owner", "repo"},

		// --- GHCR bare ----------------------------------------------------------
		{"ghcr.io/owner/repo", "owner", "repo"},
		{"ghcr.io/infamousrusty/tagsha", "infamousrusty", "tagsha"},

		// --- GHCR with tag ------------------------------------------------------
		{"ghcr.io/owner/repo:latest", "owner", "repo"},
		{"ghcr.io/owner/repo:v1.2.3", "owner", "repo"},
		{"ghcr.io/owner/repo:sha-abc1234", "owner", "repo"},
		{"ghcr.io/owner/repo:main", "owner", "repo"},

		// --- GHCR with digest ---------------------------------------------------
		{"ghcr.io/owner/repo@sha256:abc123def456", "owner", "repo"},

		// --- GHCR with tag + digest ---------------------------------------------
		{"ghcr.io/owner/repo:v1.0.0@sha256:deadbeef", "owner", "repo"},
		{"ghcr.io/infamousrusty/tagsha-api:latest@sha256:cafebabe", "infamousrusty", "tagsha-api"},

		// --- GHCR full URL ------------------------------------------------------
		{"https://ghcr.io/owner/repo", "owner", "repo"},
		{"https://ghcr.io/owner/repo:latest", "owner", "repo"},
		{"https://ghcr.io/owner/repo:v2.0@sha256:aabbcc", "owner", "repo"},
		{"http://ghcr.io/owner/repo:latest", "owner", "repo"},

		// --- Docker Hub registry endpoints --------------------------------------
		{"docker.io/owner/repo", "owner", "repo"},
		{"docker.io/owner/repo:latest", "owner", "repo"},
		{"docker.io/owner/repo:v1.2.3", "owner", "repo"},
		{"registry-1.docker.io/owner/repo", "owner", "repo"},
		{"registry-1.docker.io/owner/repo:stable", "owner", "repo"},
		{"index.docker.io/owner/repo:latest", "owner", "repo"},
		{"registry.hub.docker.com/owner/repo:latest", "owner", "repo"},

		// --- Docker Hub full URLs -----------------------------------------------
		{"https://docker.io/owner/repo:latest", "owner", "repo"},
		{"https://registry-1.docker.io/owner/repo:latest", "owner", "repo"},

		// --- Docker Hub web UI --------------------------------------------------
		{"https://hub.docker.com/r/owner/repo", "owner", "repo"},
		{"https://hub.docker.com/r/owner/repo/tags", "owner", "repo"},
		{"https://hub.docker.com/r/owner/repo/", "owner", "repo"},
		{"hub.docker.com/r/owner/repo", "owner", "repo"},
		{"https://hub.docker.com/repository/docker/owner/repo", "owner", "repo"},
		{"https://hub.docker.com/repository/docker/owner/repo/general", "owner", "repo"},
		{"https://hub.docker.com/repository/docker/owner/repo/tags", "owner", "repo"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := validation.ParseInput(tc.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if got.Owner != tc.owner || got.Repo != tc.repo {
				t.Errorf("ParseInput(%q) = {%s, %s}, want {%s, %s}",
					tc.input, got.Owner, got.Repo, tc.owner, tc.repo)
			}
		})
	}
}

func TestParseInput_Invalid(t *testing.T) {
	cases := []struct {
		input string
		desc  string
	}{
		// --- SSRF / unknown hosts -----------------------------------------------
		{"https://evil.com/owner/repo", "unknown host"},
		{"https://localhost/owner/repo", "localhost"},
		{"https://192.168.1.1/owner/repo", "private IP"},
		{"https://10.0.0.1/owner/repo", "private IP"},
		{"https://[::1]/owner/repo", "loopback IPv6"},
		{"http://169.254.169.254/latest/meta-data", "IMDS endpoint"},
		{"gcr.io/owner/repo", "unsupported registry"},
		{"quay.io/owner/repo", "unsupported registry"},
		{"mcr.microsoft.com/owner/repo", "unsupported registry"},
		{"someregistry.example.com/owner/repo", "unknown registry host"},

		// --- Path traversal -----------------------------------------------------
		{"../../../etc/passwd", "path traversal"},
		{"../../owner/repo", "path traversal"},

		// --- Injection attempts -------------------------------------------------
		{"owner; ls -la/repo", "shell injection"},
		{"owner/repo; rm -rf /", "shell injection"},
		{"owner/<script>/repo", "XSS attempt"},
		{"owner/repo$(whoami)", "command substitution"},

		// --- Format errors ------------------------------------------------------
		{"owner", "no slash"},
		{"", "empty"},
		{"/", "bare slash"},
		{"owner/", "empty repo"},
		{"/repo", "empty owner"},
		{"ghcr.io/owner", "missing repo in ghcr ref"},
		{"docker.io/owner", "missing repo in docker.io ref"},
		{"https://hub.docker.com/r/owner", "hub URL missing repo"},
		{"https://hub.docker.com/explore", "hub URL no owner/repo"},

		// --- Invalid owner names ------------------------------------------------
		{"-owner/repo", "leading hyphen"},
		{"owner-/repo", "trailing hyphen"},
		{"ghcr.io/-owner/repo", "leading hyphen in ghcr"},
		{"docker.io/-owner/repo:latest", "leading hyphen in docker.io"},

		// --- Too long -----------------------------------------------------------
		{strings.Repeat("a", 600), "exceeds max length"},
	}

	for _, tc := range cases {
		t.Run(tc.desc+": "+tc.input, func(t *testing.T) {
			_, err := validation.ParseInput(tc.input)
			if err == nil {
				t.Errorf("expected error for %q (%s), got nil", tc.input, tc.desc)
			}
		})
	}
}
