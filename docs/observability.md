# Observability Guide

TagSha exposes structured logs, Prometheus metrics, and health endpoints. This document explains how each component works, what is measured, and how to use the pre-built Grafana dashboard.

---

## Health Endpoint

### `GET /health`

Returns the current health status of the API and its dependencies.

**Response (healthy)**
```json
{
  "status": "healthy",
  "version": "v1.2.3",
  "checks": {
    "redis": "ok"
  },
  "uptime_seconds": 3600.5
}
```

**Response (degraded — Redis unreachable)**
```json
{
  "status": "degraded",
  "version": "v1.2.3",
  "checks": {
    "redis": "error"
  },
  "uptime_seconds": 120.1
}
```

| Status | HTTP code | Meaning |
|---|---|---|
| `healthy` | 200 | All dependencies reachable |
| `degraded` | 200 | Redis unreachable; API still serving (stale cache possible) |
| `unhealthy` | 503 | Reserved for catastrophic failure |

> Docker health checks call `/health` every 15 seconds. The container is marked unhealthy after 3 consecutive failures.

---

## Metrics Endpoint

### `GET /metrics`

Exposes Prometheus-compatible metrics in the standard text exposition format.

```bash
curl http://localhost:8080/metrics
```

### Available Metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `tagsha_requests_total` | Counter | `method`, `path`, `status` | HTTP requests served |
| `tagsha_request_duration_seconds` | Histogram | `path` | Request latency |
| `tagsha_cache_operations_total` | Counter | `operation`, `result` | Cache HIT/MISS/ERROR |
| `tagsha_github_api_calls_total` | Counter | `path`, `status` | GitHub API calls |
| `tagsha_github_api_latency_seconds` | Histogram | `path` | GitHub API round-trip |
| `tagsha_github_rate_limit_remaining` | Gauge | — | GitHub rate limit credits remaining |
| `tagsha_errors_total` | Counter | `type` | Errors by category |
| `tagsha_rate_limit_hits_total` | Counter | `path` | Client-side rate limit rejections |

---

## Structured Logging

All logs are emitted as **JSON to stdout**, compatible with any log aggregator (Loki, Fluentd, Datadog, Splunk).

**Example log line (request)**
```json
{
  "level": "info",
  "request_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "method": "GET",
  "path": "/api/v1/tags/golang/go",
  "status": 200,
  "duration_ms": 42,
  "ip": "172.17.0.1",
  "user_agent": "Mozilla/5.0",
  "time": "2026-03-15T12:00:00Z",
  "message": "request"
}
```

**Log levels**

| Level | Meaning |
|---|---|
| `debug` | Detailed cache/API internals (development only) |
| `info` | Every HTTP request (default) |
| `warn` | Cache errors, rate limit warnings |
| `error` | GitHub API failures, Redis connectivity |

Set log level via:
```bash
TAGSHA_LOG_LEVEL=debug
```

> **Privacy note:** IP addresses are logged by default. Under UK DPA 2018 / GDPR, IP addresses are personal data. If logging IPs is not required, configure your log pipeline to drop the `ip` field, or set `TAGSHA_TRUSTED_PROXY_CIDRS=""` to prevent X-Real-IP from being trusted.

---

## Prometheus Configuration

Prometheus is pre-configured in `infrastructure/prometheus/prometheus.yml` to scrape the API every 15 seconds.

```yaml
scrape_configs:
  - job_name: tagsha-api
    static_configs:
      - targets: ['api:8080']
    metrics_path: /metrics
    scrape_interval: 15s
```

### Alert Rules

Five alert rules are provisioned in `infrastructure/prometheus/alerts.yml`:

| Alert | Condition | Severity |
|---|---|---|
| `HighErrorRate` | >10% 5xx rate for 5 min | warning |
| `GitHubRateLimitLow` | <100 remaining credits | warning |
| `GitHubRateLimitExhausted` | 0 credits remaining | critical |
| `HighP99Latency` | P99 >3s for 5 min | warning |
| `RateLimitHitsHigh` | >10 client rejections/s | warning |

---

## Grafana Dashboard

Grafana is provisioned automatically at startup. No manual import required.

**Access:** `https://yourdomain.com/grafana`

**Dashboard panels:**

1. **Request Rate (req/s)** — breakdown by method, path, status
2. **Latency Percentiles** — P50 / P95 / P99
3. **Cache Hit Ratio** — proportion of requests served from cache
4. **GitHub Rate Limit Gauge** — remaining API credits (colour-coded)
5. **Error Rate by Type** — validation / GitHub / cache errors
6. **Rate Limit Hits** — client rejections over time
7. **GitHub API Latency P99** — upstream response time

Default time range: last 1 hour, auto-refreshed every 30 seconds.

---

## Accessing Prometheus Directly

Prometheus is on the internal Docker network only. Access it via SSH tunnel:

```bash
ssh -L 9090:localhost:9090 user@your-server
# Then open: http://localhost:9090
```

### Useful PromQL Queries

```promql
# Current request rate
rate(tagsha_requests_total[5m])

# Cache hit ratio over 5 minutes
rate(tagsha_cache_operations_total{result="hit"}[5m])
/ rate(tagsha_cache_operations_total{operation="get"}[5m])

# Error rate
rate(tagsha_errors_total[5m])

# GitHub rate limit remaining
tagsha_github_rate_limit_remaining

# P99 request latency
histogram_quantile(0.99, rate(tagsha_request_duration_seconds_bucket[5m]))
```
