# Deployment Guide

## Production Deployment

### Requirements

- Docker Engine 24+ with Compose v2
- A domain with DNS A record pointing to your server
- Ports 80 and 443 open

### Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/infamousrusty/tagsha.git
   cd tagsha
   ```

2. **Configure secrets**
   ```bash
   mkdir -p secrets
   echo 'ghp_your_real_token_here' > secrets/github_token
   echo 'strong_grafana_password'  > secrets/grafana_admin_password
   chmod 600 secrets/*
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env: set TAGSHA_DOMAIN to your actual domain
   ```

4. **Update Caddyfile domain**
   Edit `infrastructure/caddy/Caddyfile` and replace `tagsha.yourdomain.com`
   with your actual domain. Also update the admin email.

5. **Build the frontend**
   ```bash
   make frontend-build
   ```

6. **Start the stack**
   ```bash
   VERSION=v1.0.0 docker compose -f infrastructure/docker-compose.yml up -d
   ```

7. **Verify**
   ```bash
   curl https://yourdomain.com/health
   ```

### Upgrading

```bash
git pull
make frontend-build
VERSION=v1.x.x docker compose -f infrastructure/docker-compose.yml up -d --build api
```

---

## Observability Access

### Grafana
Available at `https://yourdomain.com/grafana`
- Default user: `admin`
- Password: content of `secrets/grafana_admin_password`
- Dashboard: **TagSha → TagSha Overview**

### Prometheus
Not exposed publicly. Access via SSH tunnel:
```bash
ssh -L 9090:localhost:9090 user@yourserver
# Then visit http://localhost:9090
```

---

## Runbook: GitHub Rate Limit Exhausted

1. Check Grafana — `tagsha_github_rate_limit_remaining` metric
2. Cache continues serving stale data — no user impact for cached repos
3. Add/rotate `secrets/github_token` (5000 req/hour vs 60 unauthenticated)
4. Restart API to pick up new token: `docker compose restart api`

## Runbook: Redis Down

1. Health endpoint will return `degraded`
2. API continues to serve GitHub data but without caching or rate limiting
3. Restart Redis: `docker compose restart redis`
4. Cache will auto-repopulate on next requests
