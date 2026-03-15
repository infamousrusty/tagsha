# TagSha Infrastructure

Docker Compose stack definition and configuration for all TagSha services.

---

## Services

| Service | Image | Port (internal) | Description |
|---|---|---|---|
| `caddy` | `caddy:2.8.4-alpine` | 80/443 (public) | TLS termination, static SPA, reverse proxy |
| `api` | `ghcr.io/infamousrusty/tagsha-api` | 8080 | Go backend API |
| `redis` | `redis:7.2.4-alpine` | 6379 | Cache and rate limit counters |
| `prometheus` | `prom/prometheus:v2.51.2` | 9090 | Metrics scraper |
| `grafana` | `grafana/grafana:10.4.2` | 3000 | Dashboards |

---

## Networks

| Network | Type | Services |
|---|---|---|
| `public` | bridge (internet-accessible) | caddy |
| `internal` | bridge (`internal: true`) | api, redis, prometheus, grafana |

No internal service can initiate outbound internet connections.

---

## Deployment

See [../docs/deployment.md](../docs/deployment.md) for full step-by-step instructions.

**Quick start:**
```bash
make secrets-init          # Create placeholder secret files
cp ../.env.example ../.env # Copy and edit environment config
make docker-up             # Build frontend + start stack
```

---

## Overlays

| File | Purpose |
|---|---|
| `docker-compose.yml` | Base production stack |
| `docker-compose.dev.yml` | Development overrides (expose ports, debug logging) |
| `docker-compose.secure.yml` | Security hardening (`no-new-privileges:true`) |

---

## Secrets

Secrets are read from files in the `../secrets/` directory.

| File | Purpose |
|---|---|
| `secrets/github_token` | GitHub Personal Access Token |
| `secrets/grafana_admin_password` | Grafana admin password |

The `secrets/` directory is excluded from git via `.gitignore`. **Never commit secret values.**
