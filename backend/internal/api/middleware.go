// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/infamousrusty/tagsha/internal/metrics"
)

type contextKey string

const (
	contextKeyRequestID contextKey = "request_id"
	contextKeyLogger    contextKey = "logger"
)

// RequestIDMiddleware injects a unique request ID into every request context and response header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), contextKeyRequestID, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestIDFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value(contextKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// LoggingMiddleware emits structured JSON log entries for every request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		log.Info().
			Str("request_id", requestIDFromCtx(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Int64("duration_ms", time.Since(start).Milliseconds()).
			Str("ip", extractIP(r)).
			Str("user_agent", r.UserAgent()).
			Msg("request")

		metrics.RequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(rw.status),
		).Inc()

		metrics.RequestDuration.
			WithLabelValues(r.URL.Path).
			Observe(time.Since(start).Seconds())
	})
}

// SecurityHeadersMiddleware adds production-grade HTTP security headers.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; frame-ancestors 'none'")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements a sliding-window rate limiter backed by Redis.
type RateLimiter struct {
	client *redis.Client
	rpm    int64
	log    zerolog.Logger
}

// NewRateLimiter creates a RateLimiter.
func NewRateLimiter(client *redis.Client, rpm int) *RateLimiter {
	return &RateLimiter{client: client, rpm: int64(rpm), log: log.With().Str("component", "rate_limiter").Logger()}
}

// Middleware enforces per-IP rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		key := "rl:" + ip
		ctx := r.Context()

		// count is int64 (Redis INCR return type) — kept as int64 throughout
		// to avoid any integer conversion.
		count, err := rl.client.Incr(ctx, key).Result()
		if err == nil && count == 1 {
			_ = rl.client.Expire(ctx, key, time.Minute).Err()
		}

		remaining := rl.rpm - count
		if remaining < 0 {
			remaining = 0
		}

		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rl.rpm, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		if count > rl.rpm {
			w.Header().Set("Retry-After", "60")
			metrics.RateLimitHitsTotal.WithLabelValues(r.URL.Path).Inc()
			rl.log.Warn().
				Str("ip", ip).
				Str("path", r.URL.Path).
				Msg("rate limit exceeded")
			writeError(w, http.StatusTooManyRequests, ErrRateLimited,
				"Rate limit exceeded. Maximum "+strconv.FormatInt(rl.rpm, 10)+" requests per minute.",
				requestIDFromCtx(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MaxBytesMiddleware limits incoming request body size.
func MaxBytesMiddleware(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" && isTrustedProxy(r.RemoteAddr) {
		return ip
	}
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
