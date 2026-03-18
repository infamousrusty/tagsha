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
)

// knownRegistryPrefixes are bare hostname prefixes (no scheme) that identify
// registry image refs. Checked via strings.HasPrefix on the lowercased input.
var knownRegistryPrefixes = []string{
	"ghcr.io/",
	"docker.io/",
	"registry-1.docker.io/",
	"index.docker.io/",
	"registry.hub.docker.com/",
}

// knownWebPrefixes are URL-like prefixes that must go through url.Parse.
// Order matters: more-specific entries (hub.docker.com) before generic http.
var knownWebPrefixes = []string{
	"https://github.com/",
	"http://github.com/",
	"github.com/",
	"https://ghcr.io/",
	"http://ghcr.io/",
	"https://docker.io/",
	"http://docker.io/",
	"https://registry-1.docker.io/",
	"http://registry-1.docker.io/",
	"https://index.docker.io/",
	"http://index.docker.io/",
	"https://registry.hub.docker.com/",
	"http://registry.hub.docker.com/",
	"https://hub.docker.com/",
	"http://hub.docker.com/",
	"hub.docker.com/",
}

// RepoIdentifier holds a validated owner and repository name.
type RepoIdentifier struct {
	Owner string
	Repo  string
}

// ParseInput accepts any of the following forms and extracts a GitHub owner/repo:
//
// Plain:
//
//	owner/repo
//	owner/repo:tag
//
// GitHub:
//
//	github.com/owner/repo
//	https://github.com/owner/repo
//	https://github.com/owner/repo.git
//
// GHCR:
//
//	ghcr.io/owner/repo
//	ghcr.io/owner/repo:tag
//	ghcr.io/owner/repo:tag@sha256:<digest>
//	ghcr.io/owner/repo@sha256:<digest>
//	https://ghcr.io/owner/repo:latest
//
// Docker Hub registry endpoints:
//
//	docker.io/owner/repo[:tag]
//	registry-1.docker.io/owner/repo[:tag]
//	index.docker.io/owner/repo[:tag]
//	registry.hub.docker.com/owner/repo[:tag]
//
// Docker Hub web UI:
//
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

	lower := strings.ToLower(input)

	// 1. Recognised URL/web prefix - parse via url.Parse with strict host check.
	for _, pfx := range knownWebPrefixes {
		if strings.HasPrefix(lower, pfx) {
			return parseKnownURL(input)
		}
	}

	// 2. Recognised registry bare prefix (e.g. ghcr.io/owner/repo:tag).
	for _, pfx := range knownRegistryPrefixes {
		if strings.HasPrefix(lower, pfx) {
			// Slice off the registry hostname + slash, keep owner/repo[:tag][@digest]
			return parseRegistryRef(input[len(pfx):])
		}
	}

	// 3. Reject anything that still looks like a URL with an unknown host.
	//    This catches http://, https://, or any other scheme.
	if strings.Contains(input, "://") {
		return nil, errors.New(
			"unsupported registry or host: only github.com, ghcr.io, docker.io, " +
				"registry-1.docker.io, index.docker.io, registry.hub.docker.com, " +
				"and hub.docker.com are supported",
		)
	}

	// 4. Reject if the first path segment contains a dot - it looks like an
	//    unrecognised hostname (e.g. quay.io/owner/repo, gcr.io/project/image).
	if idx := strings.Index(input, "/"); idx != -1 {
		firstSeg := strings.ToLower(input[:idx])
		if strings.Contains(firstSeg, ".") {
			return nil, errors.New(
				"unsupported registry or host: only github.com, ghcr.io, docker.io, " +
					"registry-1.docker.io, index.docker.io, registry.hub.docker.com, " +
					"and hub.docker.com are supported",
			)
		}
	}

	// 5. Bare owner/repo[:tag] with no host.
	return parseBareRef(input)
}

// parseKnownURL parses an input that matched a known URL/web prefix.
// url.Parse is safe here because we only call it after a whitelist prefix check.
func parseKnownURL(raw string) (*RepoIdentifier, error) {
	normalised := raw
	if !strings.HasPrefix(strings.ToLower(normalised), "http") {
		normalised = "https://" + normalised
	}

	u, err := url.Parse(normalised)
	if err != nil {
		return nil, errors.New("invalid URL format")
	}

	host := strings.ToLower(u.Hostname())

	switch host {
	case "github.com":
		return parseGitHubURL(u)

	case "hub.docker.com", "www.hub.docker.com":
		return parseDockerHubWebURL(u)

	case "ghcr.io",
		"docker.io",
		"registry-1.docker.io",
		"index.docker.io",
		"registry.hub.docker.com":
		return parseRegistryRef(u.Path)

	default:
		// Should not be reached given prefix whitelisting above, but kept as
		// a defence-in-depth guard.
		return nil, errors.New(
			"unsupported registry or host: only github.com, ghcr.io, docker.io, " +
				"registry-1.docker.io, index.docker.io, registry.hub.docker.com, " +
				"and hub.docker.com are supported",
		)
	}
}

// parseGitHubURL handles https://github.com/owner/repo[.git][/anything].
func parseGitHubURL(u *url.URL) (*RepoIdentifier, error) {
	parts := splitPath(u.Path)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("GitHub URL must contain owner and repository: github.com/owner/repo")
	}
	repo := strings.TrimSuffix(parts[1], ".git")
	return validateParts(parts[0], repo)
}

// parseDockerHubWebURL handles Docker Hub web UI URLs:
//
//	hub.docker.com/r/owner/repo[/...]
//	hub.docker.com/repository/docker/owner/repo[/...]
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

// parseRegistryRef handles image refs of the form owner/repo[:tag][@sha256:<digest>].
// The registry hostname has already been stripped before this is called.
//
// Per the Docker distribution spec:
//
//	reference = name [ ":" tag ] [ "@" digest ]
func parseRegistryRef(path string) (*RepoIdentifier, error) {
	path = strings.TrimPrefix(path, "/")

	// Strip digest (@sha256:...) - appears after tag or directly after name.
	if idx := strings.Index(path, "@"); idx != -1 {
		path = path[:idx]
	}

	// Strip tag (:tag) - only the last colon to avoid matching port numbers
	// in a future multi-level image name.
	if idx := strings.LastIndex(path, ":"); idx != -1 {
		path = path[:idx]
	}

	parts := splitPath(path)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("registry image ref must be owner/repo (optionally with :tag or @digest)")
	}
	return validateParts(parts[0], parts[1])
}

// parseBareRef handles raw owner/repo[:tag] input with no host component.
func parseBareRef(input string) (*RepoIdentifier, error) {
	// Strip digest.
	if idx := strings.Index(input, "@"); idx != -1 {
		input = input[:idx]
	}
	// Strip tag - only from the repo segment (after the first slash).
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

func splitPath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func validateParts(owner, repo string) (*RepoIdentifier, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	repo = strings.TrimSuffix(repo, ".git")

	if !ownerRegex.MatchString(owner) {
		return nil, errors.New(
			"invalid owner name: must be 1-39 alphanumeric characters or hyphens, cannot start or end with a hyphen",
		)
	}
	if !repoRegex.MatchString(repo) {
		return nil, errors.New(
			"invalid repository name: must be 1-100 alphanumeric characters, dots, hyphens, or underscores",
		)
	}

	return &RepoIdentifier{Owner: owner, Repo: repo}, nil
}
