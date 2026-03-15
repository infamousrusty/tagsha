package validation

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	// ownerRegex matches valid GitHub usernames:
	// 1–39 chars, alphanumeric + hyphens, no leading/trailing hyphen.
	ownerRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,37}[a-zA-Z0-9])?$`)

	// repoRegex matches valid GitHub repository names.
	repoRegex = regexp.MustCompile(`^[a-zA-Z0-9_\.\-]{1,100}$`)
)

// RepoIdentifier holds a validated owner and repository name.
type RepoIdentifier struct {
	Owner string
	Repo  string
}

// ParseInput parses "owner/repo" or "https://github.com/owner/repo" (and variants).
// It is the primary SSRF and injection defence at the input layer.
func ParseInput(input string) (*RepoIdentifier, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("input must not be empty")
	}
	if len(input) > 256 {
		return nil, errors.New("input exceeds maximum length of 256 characters")
	}

	// Detect URL-like inputs.
	if strings.HasPrefix(input, "http://") ||
		strings.HasPrefix(input, "https://") ||
		strings.HasPrefix(input, "github.com/") {
		return parseURL(input)
	}

	return parseOwnerRepo(input)
}

func parseURL(raw string) (*RepoIdentifier, error) {
	if !strings.HasPrefix(raw, "http") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("invalid URL format")
	}

	// SSRF: Only allow github.com — reject all other hosts.
	host := strings.ToLower(u.Hostname())
	if host != "github.com" {
		return nil, errors.New("only github.com repositories are supported")
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] == "" || pathParts[1] == "" {
		return nil, errors.New("URL must contain both owner and repository: github.com/owner/repo")
	}

	return validateParts(pathParts[0], pathParts[1])
}

func parseOwnerRepo(input string) (*RepoIdentifier, error) {
	parts := strings.SplitN(input, "/", 2)
	if len(parts) != 2 {
		return nil, errors.New("format must be owner/repo or a full github.com URL")
	}
	return validateParts(parts[0], parts[1])
}

func validateParts(owner, repo string) (*RepoIdentifier, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)

	// Strip trailing .git suffix (common in clone URLs).
	repo = strings.TrimSuffix(repo, ".git")

	if !ownerRegex.MatchString(owner) {
		return nil, errors.New(
			"invalid owner name: must be 1–39 alphanumeric characters or hyphens, cannot start or end with a hyphen",
		)
	}
	if !repoRegex.MatchString(repo) {
		return nil, errors.New(
			"invalid repository name: must be 1–100 alphanumeric characters, dots, hyphens, or underscores",
		)
	}

	return &RepoIdentifier{Owner: owner, Repo: repo}, nil
}
