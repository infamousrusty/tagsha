package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/infamousrusty/tagsha/internal/metrics"
)

// Sentinel errors for callers to inspect.
var (
	ErrNotFound    = errors.New("repository not found")
	ErrRateLimited = errors.New("GitHub API rate limit exceeded")
)

// RateLimitInfo carries GitHub rate limit metadata from response headers.
type RateLimitInfo struct {
	Remaining int
	Reset     int64
}

// Client wraps the GitHub REST API with rate-limit awareness and metrics.
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// New creates a new GitHub Client.
func New(token, baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
			// Restrict redirects to api.github.com only — SSRF mitigation.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if req.URL.Host != "api.github.com" {
					return fmt.Errorf("redirect to non-GitHub host denied: %s", req.URL.Host)
				}
				if len(via) >= 3 {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
		token:   token,
		baseURL: baseURL,
	}
}

// doRequest executes a GET request to the GitHub API and decodes the JSON response.
func (c *Client) doRequest(ctx context.Context, path string, result interface{}) (*RateLimitInfo, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "tagsha/1.0 (https://github.com/infamousrusty/tagsha)")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	metrics.GitHubAPILatency.WithLabelValues(path).Observe(time.Since(start).Seconds())

	if err != nil {
		metrics.GitHubAPICallsTotal.WithLabelValues(path, "network_error").Inc()
		metrics.ErrorsTotal.WithLabelValues("github_network").Inc()
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	rl := &RateLimitInfo{
		Remaining: parseHeader(resp.Header.Get("X-RateLimit-Remaining")),
		Reset:     int64(parseHeader(resp.Header.Get("X-RateLimit-Reset"))),
	}
	metrics.GitHubRateLimitRemaining.Set(float64(rl.Remaining))
	metrics.GitHubAPICallsTotal.WithLabelValues(path, strconv.Itoa(resp.StatusCode)).Inc()

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusNotFound:
		return rl, ErrNotFound
	case http.StatusForbidden, http.StatusTooManyRequests:
		return rl, ErrRateLimited
	default:
		metrics.ErrorsTotal.WithLabelValues("github_api").Inc()
		return rl, fmt.Errorf("GitHub API returned unexpected status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return rl, fmt.Errorf("decoding GitHub response: %w", err)
	}
	return rl, nil
}

func parseHeader(val string) int {
	n, _ := strconv.Atoi(val)
	return n
}
