# Deployment Guide

tagsha is distributed as two Docker images published to the GitHub Container Registry (GHCR):

| Image | Description |
|---|---|
| `ghcr.io/infamousrusty/tagsha-api` | Go REST API (scratch image, UID 1000) |
| `ghcr.io/infamousrusty/tagsha-frontend` | React SPA served by nginx (includes `/api/` reverse proxy) |

A Redis instance is required for API response caching.

---

## Prerequisites

- Docker Engine 24+ or Docker Desktop
- Portainer CE/BE (optional but recommended for homelab)
- A GitHub Personal Access Token (optional — see [GitHub Token](#github-token))

---

## GitHub Token

The API works **without** a token but is subject to GitHub’s unauthenticated rate limit of **60 requests per hour**. Providing a fine-grained PAT raises this to **5,000 requests per hour**.

To create a token:

1. Go to **GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens**
2. Click **Generate new token**
3. Set a descriptive name (e.g. `tagsha-readonly`)
4. Under **Repository permissions**, set **Contents** to **Read-only**
5. No other scopes are required
6. Copy the token and set it as `TAGSHA_GITHUB_TOKEN` in your env file

> **Security note:** never commit a real token to version control.
> `stack.env` is gitignored — only `stack.env.example` is committed.

---

## Deploy with Portainer

### 1. Prepare your environment file

```bash
cp deploy/stack.env.example deploy/stack.env
# Edit stack.env with your values
```

At minimum, review:
- `TAGSHA_VERSION` — pin to a release tag for production (e.g. `v0.0.5`)
- `TAGSHA_GITHUB_TOKEN` — optional, see above
- `TAGSHA_FRONTEND_PORT` — host port to expose the UI on (default `3000`)

### 2. Create the stack in Portainer

1. Open Portainer → **Stacks** → **Add stack**
2. Give the stack a name (e.g. `tagsha`)
3. Select **Upload** and upload `deploy/docker-compose.yml`,
   or select **Web editor** and paste the file contents
4. Scroll to **Environment variables** → click **Load variables from .env file**
   and upload your `deploy/stack.env`
5. Click **Deploy the stack**

Portainer will pull the images from GHCR and start all three services.
Once all healthchecks pass, the UI is available at:

```
http://<your-host>:3000
```

---

## Deploy with Docker Compose (CLI)

```bash
# Clone the repo
git clone https://github.com/infamousrusty/tagsha.git
cd tagsha

# Prepare env file
cp deploy/stack.env.example deploy/stack.env
$EDITOR deploy/stack.env

# Pull and start
docker compose -f deploy/docker-compose.yml --env-file deploy/stack.env up -d

# Check status
docker compose -f deploy/docker-compose.yml ps

# View logs
docker compose -f deploy/docker-compose.yml logs -f
```

---

## Updating

```bash
# Pull latest images and recreate containers
docker compose -f deploy/docker-compose.yml --env-file deploy/stack.env pull
docker compose -f deploy/docker-compose.yml --env-file deploy/stack.env up -d
```

In Portainer: **Stacks → tagsha → Pull and redeploy**.

---

## Stopping / Removing

```bash
# Stop without removing volumes
docker compose -f deploy/docker-compose.yml down

# Stop and remove all data
docker compose -f deploy/docker-compose.yml down -v
```

In Portainer: **Stacks → tagsha → Delete stack**.

---

## Configuration Reference

| Variable | Default | Description |
|---|---|---|
| `TAGSHA_VERSION` | `latest` | Image tag to deploy |
| `TAGSHA_GITHUB_TOKEN` | *(empty)* | GitHub PAT (optional) |
| `TAGSHA_ENV` | `production` | Runtime environment |
| `TAGSHA_LOG_LEVEL` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |
| `TAGSHA_CACHE_TTL_SECONDS` | `300` | Redis cache TTL in seconds |
| `TAGSHA_RATE_LIMIT_RPM` | `60` | API rate limit (requests per minute per IP) |
| `TAGSHA_MAX_PAGES` | `10` | Max GitHub API pages to fetch per request |
| `TAGSHA_DOMAIN` | `localhost` | Hostname (used for trusted proxy / CORS) |
| `TAGSHA_FRONTEND_PORT` | `3000` | Host port for the frontend |

---

## Architecture

```
Host
 │
 :3000
 │
 ┌──────────────┐ tagsha-external network
 │ frontend │
 │ nginx:1.27 │
 └─────┬─────┘
 │ /api/* tagsha-internal network
 ▼
 ┌───────────┐ ┌───────────┐
 │ api │◄───────│ redis │
 │ :8080 │ │ :6379 │
 └───────────┘ └───────────┘
```

The `api` and `redis` containers are on the **internal** network only and are not reachable from the host. Only `frontend` bridges both networks.
