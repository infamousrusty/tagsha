# Contributing to TagSha

Thank you for your interest in contributing to TagSha. This document explains the workflow, standards, and review process for external and internal contributors.

---

## Code of Conduct

TagSha follows the [Contributor Covenant](https://www.contributor-covenant.org/) Code of Conduct. By participating, you agree to uphold a respectful and inclusive environment.

---

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+ and npm
- Docker 24+ with Compose v2
- Git

### Fork and Clone

```bash
git clone https://github.com/YOUR_USERNAME/tagsha.git
cd tagsha
```

### Set Up Development Environment

```bash
# Initialise secret placeholder files
make secrets-init

# Install frontend dependencies
cd frontend && npm ci && cd ..

# Download Go dependencies
cd backend && go mod download && cd ..

# Start the full development stack
make docker-dev
```

---

## Branch Strategy

| Branch | Purpose |
|---|---|
| `main` | Production-ready code. Protected. Requires PR + review. |
| `develop` | Integration branch for in-progress work |
| `feature/your-feature` | Feature branches, branched from `develop` |
| `fix/your-fix` | Bug fix branches |
| `docs/your-docs` | Documentation-only changes |

---

## Development Workflow

1. Create a branch: `git checkout -b feature/my-feature`
2. Make your changes
3. Run tests: `make test`
4. Run linting: `make lint`
5. Commit using [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: add tag filtering by date
   fix: correct stale cache key prefix
   docs: update deployment guide
   test: add handler unit tests
   refactor: extract validation helpers
   ```
6. Push and open a pull request against `develop`

---

## Code Standards

### Backend (Go)

- All code must pass `go vet ./...` and `staticcheck ./...`
- All exported functions must have a doc comment
- Error values must be wrapped with `fmt.Errorf("context: %w", err)`
- No `panic()` in production paths
- Tests must use `testing.T` and the standard library only (no third-party assertion libs except `miniredis` for cache tests)
- New packages must include at least one `_test.go` file
- Secrets must never be hardcoded or logged

### Frontend (TypeScript)

- Strict TypeScript — no `any` types without documented justification
- ESLint must report no errors
- All API responses must be validated before use
- No `console.log` in production code
- All components must be tested with Vitest

### Docker / Infrastructure

- All containers must run as non-root
- New services must be added to the health check matrix
- Secrets must use Docker secrets, not environment variables, in production

---

## Running Tests

```bash
# All tests
make test

# Backend only (with race detector)
make test-backend

# Frontend only
make test-frontend

# Integration tests (requires running stack)
TAGSHA_API_URL=http://localhost:8080 \
  go test -tags integration ./tests/integration/...
```

---

## Pull Request Process

1. All CI checks must pass before review
2. One approving review required from a maintainer
3. Squash merge preferred for feature branches
4. Update `docs/CHANGELOG.md` with a summary of changes in the `[Unreleased]` section
5. Add or update tests for all changed functionality
6. If adding a new endpoint, update `README.md` API reference

---

## Security Issues

Do **not** open a public GitHub issue for security vulnerabilities.

Report via GitHub Security Advisories:
https://github.com/infamousrusty/tagsha/security/advisories/new

---

## Questions

For questions about the codebase, open a GitHub Discussion:
https://github.com/infamousrusty/tagsha/discussions
