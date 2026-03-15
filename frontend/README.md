# TagSha Frontend

React 18 + TypeScript 5 single-page application bundled with Vite.

---

## Tech Stack

| Library | Purpose |
|---|---|
| React 18 | UI framework |
| TypeScript 5 | Type safety |
| Vite | Build tool and dev server |
| Vitest | Unit test runner |
| ESLint | Linting |

---

## Development

```bash
npm ci           # Install pinned dependencies
npm run dev      # Start dev server at http://localhost:5173
npm run build    # Production build to dist/
npm run preview  # Preview production build locally
npm run test     # Run Vitest unit tests
npm run lint     # ESLint
npm run type-check  # TypeScript type checking without emit
```

---

## Build Output

`npm run build` produces a static bundle in `frontend/dist/`. In the Docker Compose stack, this directory is mounted read-only into Caddy at `/srv/frontend` and served directly.

---

## Configuration

The frontend communicates with the backend API at `/api/v1/*`. In development, Vite proxies these requests to `http://localhost:8080`.

No environment variables are embedded in the frontend build. The API base URL is always a relative path, so the same build works across all domains.

---

## Component Overview

| Component | Description |
|---|---|
| `App` | Root component; state management, API calls |
| `SearchBar` | Repository identifier input and submit |
| `TagList` | Table of resolved tags with filter |
| `TagRow` | Single tag row with copy-SHA button |
| `CopyButton` | Clipboard copy with visual feedback |

---

## Security

- All data rendered as React text nodes (no `dangerouslySetInnerHTML`)
- No third-party analytics, tracking scripts, or cookies
- CSP set by Caddy (not inline) prevents XSS
- No secrets or tokens handled in the frontend
