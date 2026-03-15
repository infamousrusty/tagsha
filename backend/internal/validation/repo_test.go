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
		{"torvalds/linux", "torvalds", "linux"},
		{"rust-lang/rust", "rust-lang", "rust"},
		{"https://github.com/golang/go", "golang", "go"},
		{"https://github.com/rust-lang/rust", "rust-lang", "rust"},
		{"github.com/infamousrusty/tagsha", "infamousrusty", "tagsha"},
		{"owner/repo.name", "owner", "repo.name"},
		{"owner/repo-name", "owner", "repo-name"},
		{"owner/repo_name", "owner", "repo_name"},
		{"https://github.com/owner/repo.git", "owner", "repo"},
		{"A/B", "A", "B"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := validation.ParseInput(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Owner != tc.owner || got.Repo != tc.repo {
				t.Errorf("got {%s, %s}, want {%s, %s}", got.Owner, got.Repo, tc.owner, tc.repo)
			}
		})
	}
}

func TestParseInput_Invalid(t *testing.T) {
	cases := []string{
		// SSRF attempts
		"https://evil.com/owner/repo",
		"https://localhost/owner/repo",
		"https://192.168.1.1/owner/repo",
		"https://10.0.0.1/owner/repo",
		"https://[::1]/owner/repo",
		"http://169.254.169.254/latest/meta-data",
		// Path traversal
		"../../../etc/passwd",
		"../../owner/repo",
		// Injection attempts
		"owner; ls -la/repo",
		"owner/repo; rm -rf /",
		"owner/<script>/repo",
		"owner/repo$(whoami)",
		// Format errors
		"owner",
		"",
		"/",
		"owner/",
		"/repo",
		// Invalid owner
		"-owner/repo",
		"owner-/repo",
		// Too long
		strings.Repeat("a", 300),
	}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := validation.ParseInput(input)
			if err == nil {
				t.Errorf("expected error for %q, got nil", input)
			}
		})
	}
}
