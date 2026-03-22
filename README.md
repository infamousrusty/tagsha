# TagSha

> Resolve any GitHub repository's version tags to exact, reproducible commit SHAs.

[![Status: Work in Progress](https://img.shields.io/badge/Status-Work_in_Progress-orange.svg)](https://github.com/infamousrusty/tagsha)
[![GHCR Version](https://ghcr-badge.egpl.dev/infamousrusty/tagsha/latest_tag?color=2496ED&label=version&trim=major)](https://github.com/infamousrusty/tagsha/pkgs/container/tagsha)

[![CI](https://github.com/infamousrusty/tagsha/actions/workflows/ci.yml/badge.svg)](https://github.com/infamousrusty/tagsha/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/infamousrusty/tagsha)](https://goreportcard.com/report/github.com/infamousrusty/tagsha)
[![License: LGPL-3.0-or-later](https://img.shields.io/badge/License-LGPL--3.0--or--later-blue.svg)](LICENSE)

[![Release](https://img.shields.io/github/v/release/infamousrusty/tagsha)](https://github.com/infamousrusty/tagsha/releases)
[![Last Commit](https://img.shields.io/github/last-commit/infamousrusty/tagsha)](https://github.com/infamousrusty/tagsha/commits)
[![Repo Size](https://img.shields.io/github/repo-size/infamousrusty/tagsha)](https://github.com/infamousrusty/tagsha)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Finfamousrusty%2Ftagsha.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Finfamousrusty%2Ftagsha?ref=badge_shield)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Finfamousrusty%2Ftagsha.svg?type=shield&issueType=security)](https://app.fossa.com/projects/git%2Bgithub.com%2Finfamousrusty%2Ftagsha?ref=badge_shield&issueType=security)

---

## What is TagSha?

TagSha is a self-hosted web service that looks up a GitHub repository’s tags and resolves each to its exact commit SHA. This makes it trivially safe to pin any dependency, infrastructure component, or CI action to a specific, immutable commit rather than a floating tag.

**Use cases:**
- Pinning GitHub Actions to SHA instead of tag
- Auditing infrastructure-as-code dependency versions
- Generating reproducible dependency manifests
- Investigating what commit a release tag actually points to

---

## Architecture

```
Browser
  │
  └─► Caddy (TLS termination, static files, reverse proxy)
          │
          ├─► API (Go — chi router, zerolog, prometheus)
          │     └─► Redis (tag cache, rate limit counters)
          │     └─► GitHub API (github.com/api.github.com)
          │
          ├─► Frontend (React + TypeScript, served as static files)
          │
          └─► Grafana (dashboards, served at /grafana)
                └─► Prometheus (metrics scraper)
```

See [docs/architecture.md](docs/architecture.md) for full detail.

---

## Quick Start

### Prerequisites

- Docker 24+ with Compose v2
- Node.js 20+ (for frontend build)
- A GitHub Personal Access Token (optional but recommended)

### 1. Clone

```bash
git clone https://github.com/infamousrusty/tagsha.git
cd tagsha
```

### 2. Initialise secrets

```bash
make secrets-init
# Then edit secrets/github_token with a real token
```

### 3. Configure

```bash
cp .env.example .env
# Edit .env — set TAGSHA_DOMAIN at minimum
```

### 4. Build frontend and deploy

```bash
make docker-up
```

The stack will be available at your configured domain (or `http://localhost` in dev mode).

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `TAGSHA_PORT` | `8080` | API listen port |
| `TAGSHA_ENV` | `development` | `production` or `development` |
| `TAGSHA_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `TAGSHA_REDIS_URL` | `redis://redis:6379/0` | Redis connection URL |
| `TAGSHA_CACHE_TTL_SECONDS` | `300` | Tag cache TTL (seconds) |
| `TAGSHA_RATE_LIMIT_RPM` | `50` | Requests per minute per IP |
| `TAGSHA_MAX_PAGES` | `10` | Maximum tag pages to fetch (100 tags/page) |
| `TAGSHA_DOMAIN` | `localhost` | Domain for Caddy TLS and Grafana URLs |
| `TAGSHA_GITHUB_TOKEN` | *(empty)* | GitHub PAT (or use Docker secret `github_token`) |

> **Security note:** Never set `TAGSHA_GITHUB_TOKEN` in `.env` for production. Use the Docker secret file at `secrets/github_token` instead.

---

## API Reference

### `GET /health`

Returns application and dependency health status.

```json
{
  "status": "healthy",
  "version": "v1.2.3",
  "checks": { "redis": "ok" },
  "uptime_seconds": 3600.5
}
```

### `GET /metrics`

Prometheus-compatible metrics endpoint.

### `POST /api/v1/resolve`

Parse any GitHub repository identifier into a canonical owner/repo pair.

**Request:**
```json
{ "query": "https://github.com/golang/go" }
```

**Response:**
```json
{ "owner": "golang", "repo": "go", "redirect_url": "/api/v1/tags/golang/go" }
```

### `GET /api/v1/tags/{owner}/{repo}`

Returns all tags for a repository with resolved commit SHAs.

**Response:**
```json
{
  "owner": "golang",
  "repo": "go",
  "total_count": 42,
  "truncated": false,
  "tags": [
    {
      "name": "go1.22.2",
      "sha": "a9a4c73c3e5a87e1f6e3e9f89c4b2d8d6a9f1234",
      "message": "go1.22.2",
      "author_name": "Gopher Bot",
      "date": "2024-03-05T17:00:00Z",
      "commit_url": "https://github.com/golang/go/commit/a9a4c73"
    }
  ],
  "cached_at": "2026-03-15T12:00:00Z",
  "github_rate_limit_remaining": 4987
}
```

**Response headers:**
- `X-Cache: HIT | MISS | STALE` — cache status
- `X-Request-ID` — unique request identifier for tracing
- `X-RateLimit-Limit` / `X-RateLimit-Remaining` — rate limit status

---

## Development

```bash
# Run backend tests
make test-backend

# Run frontend tests
make test-frontend

# Start dev stack (hot-reloadable frontend + dockerised backend)
make docker-dev
# Then: cd frontend && npm run dev

# Run integration tests (requires running stack)
TAGSHA_API_URL=http://localhost:8080 go test -tags integration ./tests/integration/...
```

---

## Security

- All user input is validated against strict regex patterns before use
- SSRF is mitigated at the URL parsing layer and at the HTTP client layer (redirect restriction)
- Rate limiting is enforced per IP via Redis sliding window counters
- Secrets are never logged or embedded in binaries
- Docker containers run as non-root with no Linux capabilities
- Container images are scanned by Trivy on every CI run
- Dependency audits run on every pull request

See [docs/security.md](docs/security.md) for the full security model.

---

## Observability

- `/health` — dependency health checks
- `/metrics` — Prometheus metrics (requests, latency, cache, GitHub rate limit)
- Structured JSON logs with request IDs
- Grafana dashboard provisioned automatically
- Prometheus alert rules included

See [docs/deployment.md](docs/deployment.md) for Grafana access details.

---

## License

[LGPL-3.0-or-later](LICENSE)
