package github

import (
	"context"
	"fmt"
)

// Service orchestrates tag fetching and enrichment with commit metadata.
type Service struct {
	client   *Client
	maxPages int
}

// NewService creates a new GitHub tag service.
func NewService(client *Client, maxPages int) *Service {
	return &Service{client: client, maxPages: maxPages}
}

// FetchAllTags retrieves all tags (up to maxPages pages) for a repository and
// enriches each with commit metadata. Tags already contain the commit SHA from
// the listing endpoint, so enrichment only requires one extra call per tag for
// author, message, and date — which is necessary but batched efficiently.
func (s *Service) FetchAllTags(ctx context.Context, owner, repo string) ([]Tag, *RateLimitInfo, error) {
	const perPage = 100
	var all []rawTag
	var lastRL *RateLimitInfo

	for page := 1; page <= s.maxPages; page++ {
		rawTags, rl, err := s.client.GetTags(ctx, owner, repo, page, perPage)
		if err != nil {
			return nil, rl, fmt.Errorf("fetching tags page %d: %w", page, err)
		}
		lastRL = rl
		all = append(all, rawTags...)
		// GitHub returns fewer items than perPage on the last page.
		if len(rawTags) < perPage {
			break
		}
	}

	enriched := make([]Tag, 0, len(all))
	for _, rt := range all {
		commit, rl, err := s.client.GetCommit(ctx, owner, repo, rt.Commit.SHA)
		if rl != nil {
			lastRL = rl
		}
		if err != nil {
			// Do not fail the entire response for one commit resolution failure.
			// Return the tag with partial data.
			enriched = append(enriched, Tag{
				Name: rt.Name,
				SHA:  rt.Commit.SHA,
			})
			continue
		}
		enriched = append(enriched, Tag{
			Name:        rt.Name,
			SHA:         rt.Commit.SHA,
			Message:     firstLine(commit.Commit.Message),
			AuthorName:  commit.Commit.Author.Name,
			AuthorEmail: commit.Commit.Author.Email,
			Date:        commit.Commit.Author.Date,
			CommitURL:   commit.HTMLURL,
		})
	}

	return enriched, lastRL, nil
}

// firstLine returns only the first line of a commit message.
func firstLine(msg string) string {
	for i, c := range msg {
		if c == '\n' {
			return msg[:i]
		}
	}
	return msg
}
