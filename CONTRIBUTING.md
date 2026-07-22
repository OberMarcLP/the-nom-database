# Contributing to The Nom Database

Thanks for considering a contribution! This document describes how the project is developed and released.

## Development setup

Prerequisites: Go 1.26+, Node 26+, Docker (for the database), `make`.

```bash
# database only (PostgreSQL 18 in Docker)
make db

# backend with live env from .env
make backend        # = cd backend && go run ./cmd/server

# frontend dev server (Vite, http://localhost:5173)
make frontend       # = cd frontend && npm run dev
```

Or run the full stack the way production runs it, built from source:

```bash
docker compose -f docker-compose.dev.yml up --build
```

Copy `.env.example` to `.env` first and fill in at least `DATABASE_URL` and `JWT_SECRET_KEY`.

## Tests & checks

Please make sure these pass before opening a PR:

```bash
make test           # backend go test + frontend tests
cd backend && go vet ./... && gofmt -l .
cd frontend && npm run lint && npm run build
```

- Backend: table-driven tests, aim for meaningful coverage on critical paths (auth, handlers, middleware)
- Frontend: test user-facing behavior (React Testing Library)

## Conventions

- **Commits** follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`, `ci:` — optionally scoped, e.g. `feat(maps): …`
- **Go**: `gofmt`, `go vet`, error wrapping, no ignored errors, prepared statements only
- **TypeScript/React**: functional components + hooks, typed props, React Query for server state
- **UI**: follow the brutalist design system in `frontend/src/styles/index.css` — reuse the existing CSS classes, no blur effects, no hardcoded colors
- **Migrations**: `make migrate-create NAME=add_thing` creates a numbered pair in `backend/db/migrations_new/`; always provide a working `down` migration

## Pull requests

1. Fork / branch from `main`
2. Keep PRs focused — one feature or fix per PR
3. Update docs (`README.md`, `.env.example`, `CHANGELOG.md` under **Unreleased**) when behavior or configuration changes
4. CI must be green (backend tests, frontend tests, security scan, CodeQL)

## Release process

Releases are tag-driven and follow [Semantic Versioning](https://semver.org):

1. Update `CHANGELOG.md` (move Unreleased → new version section) and bump `frontend/package.json`
2. Commit as `chore: release vX.Y.Z`, then tag: `git tag -a vX.Y.Z -m "Release X.Y.Z"` and push the tag
3. GitHub Actions builds and publishes multi-arch Docker images to GHCR (`latest`, `X`, `X.Y`, `vX.Y.Z`) with SBOM and provenance, and creates the GitHub release

## Reporting issues

Please include: what you did, what you expected, what happened, relevant backend logs (`docker compose logs backend`), and your deployment flavor (compose prod/dev, reverse proxy, auth mode).
