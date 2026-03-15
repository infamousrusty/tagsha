package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ghclient "github.com/infamousrusty/tagsha/internal/github"
)

func newMockGitHub(t *testing.T, tags []map[string]interface{}, commits map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "4999")
		w.Header().Set("X-RateLimit-Reset", "9999999999")
		switch r.URL.Path {
		case "/repos/owner/repo/tags":
			json.NewEncoder(w).Encode(tags)
		default:
			for sha, data := range commits {
				if r.URL.Path == "/repos/owner/repo/commits/"+sha {
					json.NewEncoder(w).Encode(data)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestFetchAllTags_EnrichedData(t *testing.T) {
	tags := []map[string]interface{}{
		{"name": "v1.0.0", "commit": map[string]string{"sha": "deadbeef", "url": ""}},
	}
	commits := map[string]map[string]interface{}{
		"deadbeef": {
			"sha":      "deadbeef",
			"html_url": "https://github.com/owner/repo/commit/deadbeef",
			"commit": map[string]interface{}{
				"message": "release v1.0.0\n\nDetailed description.",
				"author": map[string]string{
					"name":  "Alice",
					"email": "alice@example.com",
					"date":  "2026-01-01T00:00:00Z",
				},
			},
		},
	}

	ts := newMockGitHub(t, tags, commits)
	defer ts.Close()

	client := ghclient.New("", ts.URL)
	service := ghclient.NewService(client, 5)

	result, rl, err := service.FetchAllTags(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("FetchAllTags: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len(tags) = %d, want 1", len(result))
	}

	tag := result[0]
	if tag.Name != "v1.0.0" {
		t.Errorf("Name = %q, want v1.0.0", tag.Name)
	}
	if tag.SHA != "deadbeef" {
		t.Errorf("SHA = %q, want deadbeef", tag.SHA)
	}
	if tag.AuthorName != "Alice" {
		t.Errorf("AuthorName = %q, want Alice", tag.AuthorName)
	}
	if tag.Message != "release v1.0.0" {
		t.Errorf("Message = %q, want 'release v1.0.0'", tag.Message)
	}
	if tag.CommitURL != "https://github.com/owner/repo/commit/deadbeef" {
		t.Errorf("CommitURL = %q unexpected", tag.CommitURL)
	}
	if rl == nil || rl.Remaining != 4999 {
		t.Errorf("RateLimitInfo unexpected: %+v", rl)
	}
}

func TestFetchAllTags_Empty(t *testing.T) {
	ts := newMockGitHub(t, []map[string]interface{}{}, nil)
	defer ts.Close()

	client := ghclient.New("", ts.URL)
	service := ghclient.NewService(client, 5)

	result, _, err := service.FetchAllTags(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("FetchAllTags: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d tags", len(result))
	}
}

func TestFetchAllTags_CommitEnrichmentPartialFailure(t *testing.T) {
	tags := []map[string]interface{}{
		{"name": "v2.0.0", "commit": map[string]string{"sha": "missing", "url": ""}},
	}
	ts := newMockGitHub(t, tags, nil)
	defer ts.Close()

	client := ghclient.New("", ts.URL)
	service := ghclient.NewService(client, 5)

	result, _, err := service.FetchAllTags(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("FetchAllTags: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 partial tag, got %d", len(result))
	}
	if result[0].Name != "v2.0.0" {
		t.Errorf("Name = %q, want v2.0.0", result[0].Name)
	}
	if result[0].SHA != "missing" {
		t.Errorf("SHA = %q, want missing", result[0].SHA)
	}
	if result[0].AuthorName != "" {
		t.Errorf("AuthorName should be empty on partial failure, got %q", result[0].AuthorName)
	}
}

func TestFetchAllTags_Pagination(t *testing.T) {
	page1 := make([]map[string]interface{}, 100)
	for i := range page1 {
		sha := "sha" + string(rune('a'+i%26))
		page1[i] = map[string]interface{}{
			"name":   "v0.0." + string(rune('a'+i%26)),
			"commit": map[string]string{"sha": sha, "url": ""},
		}
	}

	var page int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "4999")
		if r.URL.Path == "/repos/owner/repo/tags" {
			page++
			if page == 1 {
				json.NewEncoder(w).Encode(page1)
			} else {
				json.NewEncoder(w).Encode([]interface{}{})
			}
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"sha": "x", "html_url": "",
				"commit": map[string]interface{}{
					"message": "x",
					"author":  map[string]string{"name": "", "email": "", "date": ""},
				},
			})
		}
	}))
	defer ts.Close()

	client := ghclient.New("", ts.URL)
	service := ghclient.NewService(client, 5)

	result, _, err := service.FetchAllTags(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("FetchAllTags pagination: %v", err)
	}
	if len(result) != 100 {
		t.Errorf("expected 100 tags from page 1, got %d", len(result))
	}
	if page != 2 {
		t.Errorf("expected 2 page fetches, got %d", page)
	}
}

func TestFetchAllTags_RateLimited(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	client := ghclient.New("", ts.URL)
	service := ghclient.NewService(client, 5)

	_, _, err := service.FetchAllTags(context.Background(), "owner", "repo")
	if err == nil {
		t.Fatal("expected error on rate limit, got nil")
	}
}
