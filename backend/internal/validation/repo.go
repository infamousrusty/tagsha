package validation

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	// ownerRegex matches valid GitHub usernames:
	// 1-39 chars, alphanumeric + hyphens, no leading/trailing hyphen.
	ownerRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,37}[a-zA-Z0-9])?$`)

	// repoRegex matches valid GitHub repository names.
	repoRegex = regexp.MustCompile(`^[a-zA-Z0-9_\.\-]{1,100}$`)

	// knownRegistries lists all registry hostnames we strip before parsing.
	// These are never passed to the GitHub API — only owner/repo is extracted.
	knownRegistries = []string{
		"ghcr.io",
		"docker.io",
		"registry-1.docker.io",
		"index.docker.io",
		"registry.hub.docker.com",
	}

	// dockerHubWebHosts are Docker Hub web UI hostnames whose URL paths we
	// parse differently (hub.docker.com/r/owner/repo or
	// hub.docker.com/repository/docker/owner/repo).
	dockerHubWebHosts = []string{
		"hub.docker.com",
		"www.hub.docker.com",
	}
)

// RepoIdentifier holds a validated owner and repository name.
type RepoIdentifier struct {
	Owner string
	Repo  string
}

// ParseInput accepts any of the following forms and extracts a GitHub owner/repo:
//
//	// Plain
//	owner/repo
//	owner/repo:tag
//
//	// GitHub
//	github.com/owner/repo
//	https://github.com/owner/repo
//	https://github.com/owner/repo.git
//
//	// GHCR
//	ghcr.io/owner/repo
//	ghcr.io/owner/repo:tag
//	ghcr.io/owner/repo:tag@sha256:<digest>
//	ghcr.io/owner/repo@sha256:<digest>
//	https://ghcr.io/owner/repo:latest
//
//	// Docker Hub registry endpoints
//	docker.io/owner/repo[:tag]
//	registry-1.docker.io/owner/repo[:tag]
//	index.docker.io/owner/repo[:tag]
//	registry.hub.docker.com/owner/repo[:tag]
//
//	// Docker Hub web UI
//	https://hub.docker.com/r/owner/repo
//	https://hub.docker.com/r/owner/repo/tags
//	https://hub.docker.com/repository/docker/owner/repo
//	https://hub.docker.com/repository/docker/owner/repo/general
//
// It is the primary SSRF and injection defence at the input layer.
func ParseInput(input string) (*RepoIdentifier, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("input must not be empty")
	}
	if len(input) > 512 {
		return nil, errors.New("input exceeds maximum length of 512 characters")
	}

	// Normalise: add scheme so url.Parse works uniformly.
	normalised := input
	if !strings.HasPrefix(normalised, "http://") && !strings.HasPrefix(normalised, "https://") {
		normalised = "https://" + normalised
	}

	u, err := url.Parse(normalised)
	if err != nil {
		return nil, errors.New("invalid URL format")
	}

	host := strings.ToLower(u.Hostname())

	// -----------------------------------------------------------------
	// GitHub web URLs  (github.com)
	// -----------------------------------------------------------------
	if host == "github.com" {
		return parseGitHubURL(u)
	}

	// -----------------------------------------------------------------
	// Docker Hub web UI  (hub.docker.com)
	// -----------------------------------------------------------------
	for _, h := range dockerHubWebHosts {
		if host == h {
			return parseDockerHubWebURL(u)
		}
	}

	// -----------------------------------------------------------------
	// Registry endpoints: ghcr.io, docker.io, registry-1.docker.io …
	// These all follow  registry/owner/repo[:tag][@digest]
	// -----------------------------------------------------------------
	for _, reg := range knownRegistries {
		if host == reg {
			return parseRegistryRef(u.Path)
		}
	}

	// -----------------------------------------------------------------
	// Bare  owner/repo[:tag]  (no recognised host prefix)
	// -----------------------------------------------------------------
	if host != "" && u.Path != "" {
		// Input looked like  somehost/path — reject unknown hosts to
		// prevent SSRF via a registry we don't explicitly allow.
		return nil, errors.New(
			"unsupported registry or host: only github.com, ghcr.io, docker.io, " +
				"registry-1.docker.io, index.docker.io, registry.hub.docker.com, " +
				"and hub.docker.com are supported",
		)
	}

	// No host detected — treat raw input as  owner/repo[:tag]
	return parseBareRef(input)
}

// ---------------------------------------------------------------------------
// GitHub  https://github.com/owner/repo[.git][/anything]
// ---------------------------------------------------------------------------
func parseGitHubURL(u *url.URL) (*RepoIdentifier, error) {
	parts := splitPath(u.Path)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("GitHub URL must contain owner and repository: github.com/owner/repo")
	}
	repo := strings.TrimSuffix(parts[1], ".git")
	return validateParts(parts[0], repo)
}

// ---------------------------------------------------------------------------
// Docker Hub web UI
//
//	hub.docker.com/r/owner/repo[/...]
//	hub.docker.com/repository/docker/owner/repo[/...]
// ---------------------------------------------------------------------------
func parseDockerHubWebURL(u *url.URL) (*RepoIdentifier, error) {
	parts := splitPath(u.Path)

	switch {
	case len(parts) >= 3 && parts[0] == "r":
		// /r/owner/repo[/tags]
		return validateParts(parts[1], parts[2])

	case len(parts) >= 4 && parts[0] == "repository" && parts[1] == "docker":
		// /repository/docker/owner/repo[/general]
		return validateParts(parts[2], parts[3])

	default:
		return nil, errors.New(
			"unsupported Docker Hub URL format; expected " +
				"hub.docker.com/r/owner/repo or " +
				"hub.docker.com/repository/docker/owner/repo",
		)
	}
}

// ---------------------------------------------------------------------------
// Registry image ref  /owner/repo[:tag][@sha256:<digest>]
//
// Docker distribution spec:
//   name      = owner "/" repo
//   reference = name [ ":" tag ] [ "@" digest ]
//   digest    = "sha256:" hex
// ---------------------------------------------------------------------------
func parseRegistryRef(path string) (*RepoIdentifier, error) {
	path = strings.TrimPrefix(path, "/")

	// Strip digest  (@sha256:…) — appears after tag or directly after name.
	if idx := strings.Index(path, "@"); idx != -1 {
		path = path[:idx]
	}

	// Strip tag  (:tag) — only from the repo segment.
	if idx := strings.LastIndex(path, ":"); idx != -1 {
		path = path[:idx]
	}

	parts := splitPath(path)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("registry image ref must be owner/repo (optionally with :tag or @digest)")
	}
	return validateParts(parts[0], parts[1])
}

// ---------------------------------------------------------------------------
// Bare  owner/repo[:tag]  — input had no recognised host.
// ---------------------------------------------------------------------------
func parseBareRef(input string) (*RepoIdentifier, error) {
	// Strip digest.
	if idx := strings.Index(input, "@"); idx != -1 {
		input = input[:idx]
	}
	// Strip tag — only from the repo segment (after the first slash).
	if slashIdx := strings.Index(input, "/"); slashIdx != -1 {
		owner := input[:slashIdx]
		repo := input[slashIdx+1:]
		if colonIdx := strings.Index(repo, ":"); colonIdx != -1 {
			repo = repo[:colonIdx]
		}
		return validateParts(owner, repo)
	}
	return nil, errors.New("format must be owner/repo, a registry image ref, or a supported URL")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func splitPath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func validateParts(owner, repo string) (*RepoIdentifier, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	repo = strings.TrimSuffix(repo, ".git")

	if !ownerRegex.MatchString(owner) {
		return nil, errors.New(
			"invalid owner name: must be 1\u201339 alphanumeric characters or hyphens, cannot start or end with a hyphen",
		)
	}
	if !repoRegex.MatchString(repo) {
		return nil, errors.New(
			"invalid repository name: must be 1\u2013100 alphanumeric characters, dots, hyphens, or underscores",
		)
	}

	return &RepoIdentifier{Owner: owner, Repo: repo}, nil
}
