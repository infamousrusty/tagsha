package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tagsha_requests_total",
			Help: "Total number of HTTP requests handled.",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tagsha_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"path"},
	)

	GitHubAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tagsha_github_api_calls_total",
			Help: "Total GitHub API calls made.",
		},
		[]string{"endpoint", "status"},
	)

	GitHubAPILatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tagsha_github_api_latency_seconds",
			Help:    "GitHub API call latency in seconds.",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"endpoint"},
	)

	CacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tagsha_cache_operations_total",
			Help: "Cache operations by type and result.",
		},
		[]string{"operation", "result"},
	)

	GitHubRateLimitRemaining = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tagsha_github_rate_limit_remaining",
			Help: "GitHub API rate limit requests remaining.",
		},
	)

	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tagsha_rate_limit_hits_total",
			Help: "Number of requests rejected by rate limiter.",
		},
		[]string{"path"},
	)

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tagsha_errors_total",
			Help: "Total application errors by type.",
		},
		[]string{"type"},
	)

	CacheHitRatio = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tagsha_cache_hit_ratio",
			Help: "Rolling cache hit ratio (0–1).",
		},
	)
)
