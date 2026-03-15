# Contributing to TagSha

Thank you for taking the time to contribute! This document explains how to get involved.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Commit Convention](#commit-convention)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold it.

---

## Getting Started

1. **Fork** the repository
2. **Clone** your fork: `git clone https://github.com/YOUR_USERNAME/tagsha.git`
3. **Set upstream**: `git remote add upstream https://github.com/infamousrusty/tagsha.git`
4. **Create a branch**: `git checkout -b feat/your-feature-name`

---

## How to Contribute

- **Bug fixes** — always welcome, please open an issue first for non-trivial fixes
- **Features** — open a feature request issue before starting work so we can discuss scope
- **Documentation** — typos, improvements, and new guides are always appreciated
- **Security issues** — please follow [SECURITY.md](SECURITY.md) and do **not** open a public issue

---

## Development Setup

### Prerequisites

- Docker 24+ with Compose v2
- Go 1.24+
- Node.js 22+
- Make

### Running locally

```bash
# Copy and edit environment
cp .env.example .env

# Initialise secrets
make secrets-init

# Start full stack
make docker-up

# Backend tests
make test-backend

# Frontend tests
make test-frontend

# Lint backend
make lint-backend
```

---

## Pull Request Process

1. Ensure all tests pass: `make test-backend test-frontend`
2. Ensure linting passes: `make lint-backend`
3. Update relevant documentation if your change affects behaviour
4. Fill in the pull request template completely
5. Link any related issues using `Closes #123`
6. Request a review from `@infamousrusty`

PRs that do not pass CI will not be merged.

---

## Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

Examples:
feat(backend): add Docker Hub registry support
fix(frontend): correct tag sort order
docs(readme): update quick start instructions
chore(ci): pin action SHAs
security(backend): patch SSRF validation
```

**Types**: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `security`, `perf`

---

## Reporting Bugs

Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.yml) issue template. Include:
- TagSha version
- Steps to reproduce
- Expected vs actual behaviour
- Relevant logs

---

## Suggesting Features

Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.yml) issue template. Describe the problem you are trying to solve, not just the solution.
