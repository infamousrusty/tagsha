# Changelog

All notable changes to TagSha are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Web GUI frontend (React + TypeScript)
- Public SaaS deployment mode
- Self-hosted deployment mode with settings page
- GHCR and Docker Hub registry support in UI

### Changed
- Go builder upgraded from 1.22 to 1.24 (resolves 6 stdlib CVEs)
- All CI action SHAs pinned for supply-chain security
- Buildx switched to `docker` driver for self-hosted runners

### Fixed
- `docker/metadata-action` SHA corrected (resolves workflow init failure)
- `go.sum` populated (resolves Docker build failure)
- Runner added to `docker` group (resolves socket permission denied)

---

## [0.1.0] - 2026-03-15

### Added
- Initial release
- Go backend with chi router, zerolog, Prometheus metrics
- Redis tag cache with configurable TTL
- GitHub API tag resolution with pagination
- Rate limiting per IP via Redis sliding window
- Docker Compose stack (API, Redis, Caddy, Grafana, Prometheus)
- Trivy container and filesystem scanning in CI
- Gosec SAST scanning in CI
- Dependabot for Go modules, npm, and GitHub Actions
- `SECURITY.md` with responsible disclosure policy
- MIT License
