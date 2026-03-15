# Security Model

## Threat Model

TagSha is a read-only proxy to the public GitHub API. It does not store user data, process payments, or have privileged system access. The primary threats are:

| Threat | Mitigation |
|---|---|
| SSRF via malicious repository URL | URL validated against allowlist (github.com only); HTTP client redirect restricted to api.github.com |
| Injection (command, SQL, XSS) | All inputs validated via strict regex before use; no SQL; frontend renders all data as text nodes |
| GitHub token exfiltration | Token never logged; loaded from Docker secret or env var; not embedded in binary or image |
| Abuse / DoS | Redis-backed per-IP sliding window rate limiter (default 50 req/min); Caddy connection limits |
| Dependency vulnerabilities | Trivy container scan + Nancy Go audit on every CI run; dependency-review on PRs |
| Privileged container escape | All containers run as non-root; all Linux capabilities dropped; read-only filesystem; no host mounts |
| Internal service exposure | Internal Docker network is isolated from internet; only Caddy on public network |

## Input Validation

All repository identifiers are validated before any downstream call:

1. Input length capped at 256 characters
2. URL-form inputs: only `github.com` hostname accepted
3. Owner name: regex `^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,37}[a-zA-Z0-9])?$`
4. Repo name: regex `^[a-zA-Z0-9_\.\-]{1,100}$`
5. Path traversal sequences rejected by the regex

Validation logic is in `backend/internal/validation/repo.go` with full test coverage including adversarial inputs.

## Secrets Management

- GitHub token: Docker secret file at `/run/secrets/github_token` (preferred) or env var
- Grafana password: Docker secret file at `/run/secrets/grafana_admin_password`
- No secrets are hardcoded in source code or Docker images
- Secret files are excluded from git via `.gitignore`

## Container Hardening

```yaml
cap_drop: [ALL]        # Drop all Linux capabilities
read_only: true        # Read-only root filesystem
user: "1000:1000"      # Non-root user
tmpfs: [/tmp]          # In-memory tmp only
no_new_privileges: true  # Applied by Docker default
```

## Network Isolation

Internal services communicate only on the `internal` Docker bridge network, which has `internal: true` set, meaning containers on this network cannot initiate connections to the internet. Only Caddy is on the `public` network.

## Security Response

To report a security vulnerability, please open a GitHub Security Advisory at:
https://github.com/infamousrusty/tagsha/security/advisories/new

Do not report security vulnerabilities via public GitHub issues.
