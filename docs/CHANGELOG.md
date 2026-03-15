# Changelog

All notable changes to TagSha are documented here.

This project follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) and [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- `TAGSHA_TRUSTED_PROXY_CIDRS` environment variable for configurable trusted proxy CIDR ranges (SEC-01 remediation)
- Trusted proxy guard on `extractIP()` to prevent rate limit bypass via spoofed `X-Real-IP`
- `infrastructure/docker-compose.secure.yml` overlay adding `no-new-privileges:true` to all containers (SEC-03 remediation)
- Grafana dashboard JSON provisioned automatically at `infrastructure/grafana/dashboards/tagsha.json` (GAP-03 remediation)
- `.github/dependabot.yml` for automated dependency updates across Go, npm, Docker, and GitHub Actions (GAP-04 remediation)
- `.github/golangci.yml` linting configuration and `golangci-lint` step in CI (GAP-05 remediation)
- `backend/internal/cache/cache_test.go` — full unit test coverage for TTL, stale, miss, ping, and TagsKey (GAP-01 remediation)
- `backend/internal/api/handler_test.go` — handler unit tests with mock GitHub server and miniredis (GAP-02 remediation)
- `backend/internal/github/service_test.go` — service mock tests covering enrichment, pagination, partial failure, rate limit (GAP-02 remediation)
- `docs/observability.md` — logging, metrics, health, Grafana, and PromQL reference
- `docs/contributing.md` — contributor workflow, code standards, PR process
- `docs/CHANGELOG.md` — this file

---

## [0.1.0] - 2026-03-15

### Added
- Initial production release of TagSha
- Go backend API with chi router, zerolog structured logging, Prometheus metrics
- Redis-backed caching with primary TTL and stale fallback
- GitHub REST API integration with tag and commit enrichment
- Two-layer SSRF protection (URL parser + HTTP client redirect restriction)
- Per-IP rate limiting via Redis sliding window
- React + TypeScript frontend SPA with tag search, SHA copy, and filtering
- Caddy reverse proxy with automatic TLS and security headers
- Docker Compose production stack (Caddy, API, Redis, Prometheus, Grafana)
- Multi-stage Docker build to `FROM scratch` runtime image
- GitHub Actions CI/CD: lint, test, Gosec SAST, Trivy scan, GHCR publish
- Grafana dashboard with 7 panels auto-provisioned
- Prometheus alert rules for error rate, rate limit, and latency
- Integration test suite covering health, resolve, SSRF, and security headers
- `docs/architecture.md`, `docs/deployment.md`, `docs/security.md`
- `SECURITY.md`, `LICENSE`, `Makefile`, `.env.example`
