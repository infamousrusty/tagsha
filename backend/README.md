# TagSha Backend

The backend is a Go 1.22 HTTP API service that resolves GitHub repository tags to exact commit SHAs.

---

## Package Structure

```
backend/
├── cmd/server/          Entry point — wires config, cache, GitHub client, router, and starts HTTP server
├── internal/
│   ├── api/             HTTP handlers, middleware (rate limit, logging, security headers), router
│   ├── cache/           Redis-backed cache with primary TTL and stale-fallback
│   ├── config/          Environment variable configuration with defaults and validation
│   ├── github/          GitHub REST API client, tag fetcher, commit enrichment service
│   ├── metrics/         Prometheus metric definitions (counters, histograms, gauges)
│   └── validation/      Input validation and SSRF mitigation for repository identifiers
├── Dockerfile          Two-stage build: golang:1.22-alpine → scratch
├── go.mod / go.sum     Pinned Go module dependencies
```

---

## Configuration

All configuration is via environment variables. See `/.env.example` for the full list.

| Variable | Default | Description |
|---|---|---|
| `TAGSHA_PORT` | `8080` | Listen port |
| `TAGSHA_ENV` | `development` | `production` or `development` |
| `TAGSHA_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `TAGSHA_REDIS_URL` | `redis://redis:6379/0` | Redis DSN |
| `TAGSHA_CACHE_TTL_SECONDS` | `300` | Primary cache TTL |
| `TAGSHA_RATE_LIMIT_RPM` | `50` | Per-IP requests per minute |
| `TAGSHA_MAX_PAGES` | `10` | Maximum tag pages to fetch (100/page) |
| `TAGSHA_TRUSTED_PROXY_CIDRS` | `172.16.0.0/12,10.0.0.0/8,...` | Trusted proxy subnets |
| `TAGSHA_GITHUB_TOKEN` | *(empty)* | GitHub PAT (prefer Docker secret) |

In production, provide the GitHub token via Docker secret:
```bash
echo 'ghp_...' > secrets/github_token
```

---

## Running Locally

```bash
# Run tests
go test -race ./...

# Build binary
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o tagsha-api ./cmd/server

# Run (requires Redis)
TAGSHA_REDIS_URL=redis://localhost:6379/0 \
TAGSHA_GITHUB_TOKEN=ghp_yourtoken \
./tagsha-api
```

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check with dependency status |
| `GET` | `/metrics` | Prometheus metrics |
| `POST` | `/api/v1/resolve` | Parse a repository identifier |
| `GET` | `/api/v1/tags/{owner}/{repo}` | Fetch all tags with resolved SHAs |

See the root `README.md` for full request/response examples.

---

## Security Model

- Input validated by regex before any downstream call
- SSRF blocked at URL parser (github.com only) and HTTP client (api.github.com redirects only)
- Secrets loaded from Docker secret files or environment variables, never logged
- All containers run as `USER 1000:1000` with `cap_drop: ALL` and `read_only: true`
