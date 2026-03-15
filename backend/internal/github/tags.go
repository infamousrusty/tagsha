package github

import (
	"context"
	"fmt"
)

// rawTag is the structure returned by the GitHub list-tags endpoint.
type rawTag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

// CommitAuthor holds author information for a commit.
type CommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// CommitDetail holds full commit metadata.
type CommitDetail struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string       `json:"message"`
		Author  CommitAuthor `json:"author"`
	} `json:"commit"`
	HTMLURL string `json:"html_url"`
}

// Tag is the enriched, resolved tag returned to callers.
type Tag struct {
	Name      string       `json:"name"`
	SHA       string       `json:"sha"`
	Message   string       `json:"message"`
	AuthorName  string     `json:"author_name"`
	AuthorEmail string     `json:"author_email"`
	Date      string       `json:"date"`
	CommitURL string       `json:"commit_url"`
}

// GetTags fetches one page of tags for a repository.
// It uses the lightweight tags endpoint which already contains the commit SHA,
// avoiding an extra API call per tag for the vast majority of cases.
func (c *Client) GetTags(ctx context.Context, owner, repo string, page, perPage int) ([]rawTag, *RateLimitInfo, error) {
	path := fmt.Sprintf("/repos/%s/%s/tags?per_page=%d&page=%d", owner, repo, perPage, page)
	var tags []rawTag
	rl, err := c.doRequest(ctx, path, &tags)
	return tags, rl, err
}

// GetCommit fetches full commit detail for a given SHA.
func (c *Client) GetCommit(ctx context.Context, owner, repo, sha string) (*CommitDetail, *RateLimitInfo, error) {
	path := fmt.Sprintf("/repos/%s/%s/commits/%s", owner, repo, sha)
	var commit CommitDetail
	rl, err := c.doRequest(ctx, path, &commit)
	return &commit, rl, err
}

// CheckRateLimit calls the GitHub rate limit endpoint to verify connectivity.
func (c *Client) CheckRateLimit(ctx context.Context) (*RateLimitInfo, error) {
	var result interface{}
	return c.doRequest(ctx, "/rate_limit", &result)
}
