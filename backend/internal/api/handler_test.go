// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/infamousrusty/tagsha/internal/api"
	"github.com/infamousrusty/tagsha/internal/cache"
	ghclient "github.com/infamousrusty/tagsha/internal/github"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestHandler(t *testing.T) (*api.Handler, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := cache.New("redis://"+mr.Addr(), 5*time.Minute)
	if err != nil {
		t.Fatalf("cache.New: %v", err)
	}
	// Point GitHub client at a server that returns empty tag arrays by default.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "4999")
		switch r.URL.Path {
		case "/repos/golang/go/tags":
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"name": "go1.22.0", "commit": map[string]string{"sha": "abc123", "url": "/repos/golang/go/commits/abc123"}},
			})
		case "/repos/golang/go/commits/abc123":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"sha":      "abc123",
				"html_url": "https://github.com/golang/go/commit/abc123",
				"commit": map[string]interface{}{
					"message": "go1.22.0",
					"author": map[string]string{
						"name":  "Gopher Bot",
						"email": "bot@golang.org",
						"date":  "2024-02-06T17:00:00Z",
					},
				},
			})
		case "/repos/notfound/repo/tags":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ts.Close)

	ghClient := ghclient.New("", ts.URL)
	service := ghclient.NewService(ghClient, 5)
	h := api.New(service, c, "test")
	return h, mr
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestHealth_OK(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.Health(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", body["status"])
	}
	if body["version"] != "test" {
		t.Errorf("version = %v, want test", body["version"])
	}
}

func TestHealth_RedisDegraded(t *testing.T) {
	h, mr := newTestHandler(t)
	mr.Close() // simulate Redis down

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.Health(rr, r)

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "degraded" {
		t.Errorf("health status = %v, want degraded when Redis is down", body["status"])
	}
}

// ---------------------------------------------------------------------------
// Resolve
// ---------------------------------------------------------------------------

func TestResolve_ValidOwnerRepo(t *testing.T) {
	h, _ := newTestHandler(t)

	body := bytes.NewBufferString(`{"query":"golang/go"}`)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", body)
	r.Header.Set("Content-Type", "application/json")
	h.Resolve(rr, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var resp struct {
		Owner       string `json:"owner"`
		Repo        string `json:"repo"`
		RedirectURL string `json:"redirect_url"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Owner != "golang" || resp.Repo != "go" {
		t.Errorf("resolve = {%s, %s}, want {golang, go}", resp.Owner, resp.Repo)
	}
	if resp.RedirectURL != "/api/v1/tags/golang/go" {
		t.Errorf("redirect_url = %s, want /api/v1/tags/golang/go", resp.RedirectURL)
	}
}

func TestResolve_ValidGitHubURL(t *testing.T) {
	h, _ := newTestHandler(t)
	body := bytes.NewBufferString(`{"query":"https://github.com/golang/go"}`)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", body)
	r.Header.Set("Content-Type", "application/json")
	h.Resolve(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestResolve_EmptyQuery(t *testing.T) {
	h, _ := newTestHandler(t)
	body := bytes.NewBufferString(`{"query":""}`)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", body)
	h.Resolve(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("empty query: status = %d, want 400", rr.Code)
	}
}

func TestResolve_SSRFAttempt(t *testing.T) {
	h, _ := newTestHandler(t)
	for _, q := range []string{
		`{"query":"https://evil.com/owner/repo"}`,
		`{"query":"https://169.254.169.254/owner/repo"}`,
		`{"query":"../../../etc/passwd"}`,
	} {
		body := bytes.NewBufferString(q)
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", body)
		h.Resolve(rr, r)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("SSRF query %s: status = %d, want 400", q, rr.Code)
		}
	}
}

func TestResolve_InvalidJSON(t *testing.T) {
	h, _ := newTestHandler(t)
	body := bytes.NewBufferString(`not-json`)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", body)
	h.Resolve(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("invalid JSON: status = %d, want 400", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// GetTags
// ---------------------------------------------------------------------------

func TestGetTags_CacheMissThenHit(t *testing.T) {
	h, _ := newTestHandler(t)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/tags/golang/go", nil)
	r.SetPathValue("owner", "golang")
	r.SetPathValue("repo", "go")

	// First request — should be a MISS (fetches from stub GitHub API).
	rr := httptest.NewRecorder()
	h.GetTags(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("MISS status = %d, want 200", rr.Code)
	}
	if rr.Header().Get("X-Cache") != "MISS" {
		t.Errorf("first request X-Cache = %q, want MISS", rr.Header().Get("X-Cache"))
	}

	// Second identical request — should be a HIT.
	rr2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/api/v1/tags/golang/go", nil)
	r2.SetPathValue("owner", "golang")
	r2.SetPathValue("repo", "go")
	h.GetTags(rr2, r2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("HIT status = %d, want 200", rr2.Code)
	}
	if rr2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("second request X-Cache = %q, want HIT", rr2.Header().Get("X-Cache"))
	}
}

func TestGetTags_TagDataIntegrity(t *testing.T) {
	h, _ := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/tags/golang/go", nil)
	r.SetPathValue("owner", "golang")
	r.SetPathValue("repo", "go")
	rr := httptest.NewRecorder()
	h.GetTags(rr, r)

	var resp struct {
		Owner      string `json:"owner"`
		Repo       string `json:"repo"`
		TotalCount int    `json:"total_count"`
		Tags       []struct {
			Name       string `json:"name"`
			SHA        string `json:"sha"`
			AuthorName string `json:"author_name"`
		} `json:"tags"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalCount != 1 {
		t.Errorf("total_count = %d, want 1", resp.TotalCount)
	}
	if len(resp.Tags) == 0 {
		t.Fatal("expected at least 1 tag")
	}
	if resp.Tags[0].Name != "go1.22.0" {
		t.Errorf("tag name = %q, want go1.22.0", resp.Tags[0].Name)
	}
	if resp.Tags[0].SHA != "abc123" {
		t.Errorf("tag SHA = %q, want abc123", resp.Tags[0].SHA)
	}
	if resp.Tags[0].AuthorName != "Gopher Bot" {
		t.Errorf("author = %q, want Gopher Bot", resp.Tags[0].AuthorName)
	}
}

func TestGetTags_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/tags/notfound/repo", nil)
	r.SetPathValue("owner", "notfound")
	r.SetPathValue("repo", "repo")
	rr := httptest.NewRecorder()
	h.GetTags(rr, r)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestGetTags_MissingOwner(t *testing.T) {
	h, _ := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/tags//go", nil)
	r.SetPathValue("owner", "")
	r.SetPathValue("repo", "go")
	rr := httptest.NewRecorder()
	h.GetTags(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Stale cache key format
// ---------------------------------------------------------------------------

func TestGetTags_StaleKeyFormat(t *testing.T) {
	// Confirms TagsKey produces a non-empty, consistent key used by the
	// stale fallback path in the handler.
	key := cache.TagsKey("golang", "go")
	if key == "" {
		t.Error("TagsKey must not be empty")
	}
	key2 := cache.TagsKey("golang", "go")
	if key != key2 {
		t.Errorf("TagsKey not deterministic: %q != %q", key, key2)
	}
}
