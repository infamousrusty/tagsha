package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	ghclient "github.com/infamousrusty/tagsha/internal/github"
	"github.com/infamousrusty/tagsha/internal/cache"
	"github.com/infamousrusty/tagsha/internal/metrics"
	"github.com/infamousrusty/tagsha/internal/validation"
	"github.com/rs/zerolog/log"
)

// TagsResponse is the JSON response body for the tag listing endpoint.
type TagsResponse struct {
	Owner                   string         `json:"owner"`
	Repo                    string         `json:"repo"`
	TotalCount              int            `json:"total_count"`
	Truncated               bool           `json:"truncated"`
	Tags                    []ghclient.Tag `json:"tags"`
	CachedAt                string         `json:"cached_at"`
	GitHubRateLimitRemaining int           `json:"github_rate_limit_remaining"`
}

// ResolveRequest is the JSON body for the /resolve endpoint.
type ResolveRequest struct {
	Query string `json:"query"`
}

// ResolveResponse is the JSON response body for the /resolve endpoint.
type ResolveResponse struct {
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	RedirectURL string `json:"redirect_url"`
}

// Handler groups all API handler dependencies.
type Handler struct {
	github    *ghclient.Service
	cache     *cache.Cache
	startTime time.Time
	version   string
}

// New creates a new Handler.
func New(gh *ghclient.Service, c *cache.Cache, version string) *Handler {
	return &Handler{
		github:    gh,
		cache:     c,
		startTime: time.Now(),
		version:   version,
	}
}

// GetTags handles GET /api/v1/tags/{owner}/{repo}.
func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := requestIDFromCtx(ctx)

	// Extract and validate path parameters.
	owner := pathParam(r, "owner")
	repo := pathParam(r, "repo")

	if owner == "" || repo == "" {
		writeError(w, http.StatusBadRequest, ErrInvalidInput,
			"Missing owner or repository in path.", reqID)
		return
	}

	// Re-validate even though router already matched — defence in depth.
	ident, err := validation.ParseInput(owner + "/" + repo)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("validation").Inc()
		writeError(w, http.StatusBadRequest, ErrInvalidInput, err.Error(), reqID)
		return
	}

	cacheKey := cache.TagsKey(ident.Owner, ident.Repo)

	// Try primary cache.
	var cached TagsResponse
	hit, err := h.cache.Get(ctx, cacheKey, &cached)
	if err != nil {
		log.Warn().Err(err).Str("request_id", reqID).Msg("cache read error")
	}
	if hit {
		w.Header().Set("X-Cache", "HIT")
		writeJSON(w, http.StatusOK, cached)
		return
	}

	// Fetch from GitHub.
	tags, rl, err := h.github.FetchAllTags(ctx, ident.Owner, ident.Repo)
	if err != nil {
		if errors.Is(err, ghclient.ErrNotFound) {
			metrics.ErrorsTotal.WithLabelValues("not_found").Inc()
			// Try stale cache before returning 404.
			var stale TagsResponse
			if staleHit, _ := h.cache.GetStale(ctx, cacheKey, &stale); staleHit {
				w.Header().Set("X-Cache", "STALE")
				writeJSON(w, http.StatusOK, stale)
				return
			}
			writeError(w, http.StatusNotFound, ErrRepoNotFound,
				"Repository not found or is private.", reqID)
			return
		}
		if errors.Is(err, ghclient.ErrRateLimited) {
			// Serve stale data if available.
			var stale TagsResponse
			if staleHit, _ := h.cache.GetStale(ctx, cacheKey, &stale); staleHit {
				w.Header().Set("X-Cache", "STALE")
				writeJSON(w, http.StatusOK, stale)
				return
			}
			metrics.ErrorsTotal.WithLabelValues("github_rate_limited").Inc()
			writeError(w, http.StatusBadGateway, ErrRateLimited,
				"GitHub API rate limit exceeded. Please try again later.", reqID)
			return
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("GitHub API error")
		metrics.ErrorsTotal.WithLabelValues("github_error").Inc()
		writeError(w, http.StatusBadGateway, ErrGitHubAPIError,
			"GitHub API is temporarily unavailable.", reqID)
		return
	}

	rateLimitRemaining := 0
	if rl != nil {
		rateLimitRemaining = rl.Remaining
	}

	response := TagsResponse{
		Owner:                    ident.Owner,
		Repo:                     ident.Repo,
		TotalCount:               len(tags),
		Tags:                     tags,
		CachedAt:                 time.Now().UTC().Format(time.RFC3339),
		GitHubRateLimitRemaining: rateLimitRemaining,
	}

	// Write to primary and stale cache.
	if err := h.cache.Set(ctx, cacheKey, response, 0); err != nil {
		log.Warn().Err(err).Str("request_id", reqID).Msg("cache write error")
	}
	_ = h.cache.SetStale(ctx, cacheKey, response)

	w.Header().Set("X-Cache", "MISS")
	writeJSON(w, http.StatusOK, response)
}

// Resolve handles POST /api/v1/resolve.
// Parses a free-form GitHub repository identifier and redirects to the tag endpoint.
func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := requestIDFromCtx(ctx)

	var req ResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidInput,
			"Invalid request body: expected {\"query\": \"owner/repo\"}.", reqID)
		return
	}

	ident, err := validation.ParseInput(req.Query)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("validation").Inc()
		writeError(w, http.StatusBadRequest, ErrInvalidInput, err.Error(), reqID)
		return
	}

	writeJSON(w, http.StatusOK, ResolveResponse{
		Owner:       ident.Owner,
		Repo:        ident.Repo,
		RedirectURL: "/api/v1/tags/" + ident.Owner + "/" + ident.Repo,
	})
}

// Health handles GET /health.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	overall := "healthy"

	if err := h.cache.Ping(ctx); err != nil {
		checks["redis"] = "error"
		overall = "degraded"
		log.Error().Err(err).Msg("Redis health check failed")
	} else {
		checks["redis"] = "ok"
	}

	statusCode := http.StatusOK
	if overall == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, map[string]interface{}{
		"status":         overall,
		"version":        h.version,
		"checks":         checks,
		"uptime_seconds": time.Since(h.startTime).Seconds(),
	})
}

// pathParam extracts a URL path parameter by name from the Chi context.
func pathParam(r *http.Request, key string) string {
	// chi.URLParam inlined to avoid a direct chi import in handler.
	return r.PathValue(key)
}
