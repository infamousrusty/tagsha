# TagSha Architecture

## Overview

TagSha is a stateless API service backed by Redis for caching. All persistent state lives in Redis. The application is designed to be horizontally scalable — multiple API instances can run behind a load balancer sharing the same Redis instance.

## Component Diagram

```
                    INTERNET
                       │
              [Caddy 2 — Port 443]
              TLS termination
              Security headers
              Static file serving (SPA)
              Rate limiting (coarse)
                       │
          ╔══════════╗ ╔════════════╗
          ║ API (Go) ║ ║ React SPA  ║
          ║ port 8080║ ║ served from ║
          ╚══════════╝ ║ /srv/front ║
               │        ╚════════════╝
        ┌─────┴─────┐
   [Redis 7]  [GitHub API]
   Cache       REST API
   Rate limit  api.github.com

   [Prometheus] ← scrapes /metrics
   [Grafana]    ← queries Prometheus
```

## Request Flow

1. Browser sends `POST /api/v1/resolve` with a repository identifier (e.g., `golang/go`)
2. API validates and normalises the input, rejecting anything that is not a valid GitHub owner/repo
3. Browser calls `GET /api/v1/tags/{owner}/{repo}`
4. API checks Redis cache; on HIT returns immediately
5. On MISS: API calls GitHub REST API (`/repos/{owner}/{repo}/tags`) to list tags, then enriches each with a commit detail call
6. Result is serialised to JSON, written to Redis (TTL 5 min default), and returned to the caller
7. A long-lived "stale" copy is also stored in Redis for graceful degradation during GitHub outages or rate limit exhaustion

## Internal Network Isolation

All internal services (API, Redis, Prometheus, Grafana) communicate over an internal Docker bridge network with no external internet access. Only Caddy is connected to the public network.

## Caching Strategy

| Cache key | TTL | Purpose |
|---|---|---|
| `tagsha:tags:{owner}:{repo}` | 300s (configurable) | Primary tag cache |
| `stale:tagsha:tags:{owner}:{repo}` | 3600s | Stale fallback |

If the primary cache is expired AND the GitHub API returns an error (rate limited, network failure), the stale copy is served with an `X-Cache: STALE` header.
