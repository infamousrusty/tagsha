// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// NewRouter constructs the Chi router with all middleware and routes mounted.
func NewRouter(h *Handler, redisClient *redis.Client, rpm int, bodyLimit int64) http.Handler {
	r := chi.NewRouter()

	rl := NewRateLimiter(redisClient, rpm)

	// Global middleware stack (applied in order).
	r.Use(RequestIDMiddleware)
	r.Use(LoggingMiddleware)
	r.Use(SecurityHeadersMiddleware)
	r.Use(MaxBytesMiddleware(bodyLimit))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Public health and metrics endpoints — no rate limiting.
	r.Get("/health", h.Health)
	r.Handle("/metrics", promhttp.Handler())

	// API v1 routes — rate limited.
	r.Group(func(r chi.Router) {
		r.Use(rl.Middleware)
		r.Use(middleware.Compress(5, "application/json"))

		// CORS headers for browser access.
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
				if req.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, req)
			})
		})

		r.Post("/api/v1/resolve", h.Resolve)
		r.Get("/api/v1/tags/{owner}/{repo}", h.GetTags)
	})

	return r
}
