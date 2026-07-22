# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.2.0] - 2026-07-22

### Added

- Restaurant update/delete now also honor the RBAC permissions
  `restaurants.update` / `restaurants.delete`: roles carrying them can manage
  any restaurant, in addition to the creator and users-table admins

### Fixed

- Added the missing `sessions.last_used_at` column (migration `000010`);
  every token refresh warned with `SQLSTATE 42703` since the v2.0.0 auth
  overhaul shipped code for a column no migration created

## [2.1.0] - 2026-07-22

### Added

- Custom S3-compatible endpoint support via `S3_ENDPOINT` (MinIO, Cloudflare
  R2, Hetzner, ...): the client uses path-style addressing and object URLs
  are built as `<endpoint>/<bucket>/<key>`

### Fixed

- `/api/health` requests no longer flood the request log and metrics: the
  skip now applies to every client, not only Docker's wget healthcheck

## [2.0.0] - 2026-07-21

### ⚠ Breaking / upgrade notes

- **PostgreSQL 16 → 18**: the 16 data directory is incompatible with 18, and the
  compose files now mount the volume at `/var/lib/postgresql` (version-aware
  layout of the 18 images). Existing deployments must migrate manually:
  1. `docker compose exec -T db pg_dumpall -U <user> > backup.sql`
  2. `docker compose down && docker volume rm <project>_postgres_data`
  3. pull/update, `docker compose up -d db`, then
     `docker compose exec -T db psql -U <user> -d postgres < backup.sql`
     ("role/database already exists" errors are expected and harmless)
- **Auth overhaul**: refresh tokens moved to rotating httpOnly cookies and the
  access token now lives in memory only (never in localStorage). Existing
  sessions are logged out once; leftover localStorage tokens are purged
  automatically on first load.
- **Dev tooling**: the `typescript` devDependency aliases the 6.x API bridge
  (`@typescript/typescript6`) while builds run on native TypeScript 7 via
  `@typescript/native`. Do not run `npm install -D typescript@latest` — it
  would break the eslint integration. Collapse to a single package once
  typescript-eslint supports the TS7 API.

### Added

- Price range feature end to end: $–$$$$ selector in the restaurant form,
  muted indicator on cards and detail view, working max-price filter
  (migration `000009_restaurant_price_range`)
- Accessible dialog system: shared Modal with focus trap, ARIA dialog
  semantics, Escape handling and focus restore; ConfirmDialog replaces every
  native `confirm()`, toasts replace every `alert()` (38 call sites)
- Backend test foundation: `internal/auth` at 95.3% coverage (Argon2id
  roundtrips, JWT expiry/tampering/`alg=none` rejection) plus 170+ handler
  guard tests
- SBOM and provenance attestations attached to the published Docker images

### Changed

- "White Table" design system (light "Mittagsmenü" / dark "Abendkarte")
  replaces the brutalist neon theme across the entire UI
- Router migrated from gorilla/mux to chi v5 — behavior-neutral (12 API
  baselines byte-identical); wrong-method requests now return 405 instead
  of 404
- Toolchain refresh: native TypeScript 7 (typechecking ~2.9× faster),
  React 19, Vite 8, Tailwind 4, Go 1.26.5
- React Query used consistently outside the admin area: mutations invalidate
  caches instead of manual reloads; cached data renders without spinner
  flicker
- `GetRestaurants` decomposed from a 422-line function into focused
  query-builder helpers; API error responses no longer leak internal details

### Fixed

- `?sort=name` and `?price_range=` returned 500 (ambiguous ORDER BY; missing
  column)
- Validation hardening: proper email parsing, non-positive path IDs rejected
  at 52 call sites, duplicate-key detection via pg error codes
- Suggestion-to-restaurant conversion is atomic; search results no longer
  clobbered by stale responses (AbortController); object-URL leak in photo
  previews plugged
- CI pipeline repaired: invalid trivy action tag and outdated Go patch level
  with known stdlib CVEs

## [1.3.0] - 2026-02-01

### Changed

#### UI & styling
- Added base `.card`, `.label`, and text/border utility classes using CSS variables
- Aligned AdminUsers and RestaurantDetail/RestaurantCard to use CSS variables instead of Tailwind `dark:` modifiers
- Removed deprecated `confirmClassName` from ConfirmDialog; use `isDangerous` for danger actions

#### Security & robustness
- Image uploads (menu photos, avatar) now validated by magic bytes in addition to Content-Type
- Rate limiter: optional trusted proxy IP via `TRUST_PROXY`, graceful cleanup on shutdown
- OIDC state store protected with mutex; cleanup goroutine shuts down gracefully
- Database: `DB_PASSWORD` required in production; default admin gets secure random password when not set

#### Bug fixes & performance
- Added `rows.Err()` checks after row iteration in restaurants and lists handlers
- Batch food-type inserts when creating restaurants; combined COUNT queries in admin analytics
- Replaced bubble sort with `sort.Slice` in metrics percentile calculation
- New `PUT /api/ratings/{id}` endpoint for editing reviews (ownership enforced)

#### Restaurant list
- Paginated list now ordered by newest first (new restaurants appear on first page)
- Normal “All Restaurants” view uses same unpaginated API as admin (restaurants + suggestions in sync)
- On 409 “already exists” from add suggestion, restaurant list is invalidated so it refreshes

## [1.0.0] - 2025-01-03

### Added

#### Authentication & Security
- Multi-mode authentication system (none/local/oauth/both)
- JWT token-based authentication with Argon2id password hashing
- Generic OIDC integration supporting Authentik, Keycloak, Auth0, Okta, and other providers
- Configurable authentication modes via AUTH_MODE environment variable
- Session management with refresh tokens
- Rate limiting (100 req/min per IP)
- Security headers (XSS, clickjacking, MIME sniffing protection)
- Input sanitization and validation
- Request size limits (10MB max)

#### Core Features
- Restaurant management (CRUD operations)
- Multi-dimensional rating system (food, service, ambiance)
- Google Maps integration for restaurant search
- Embedded maps showing restaurant locations
- Get directions to restaurants
- Restaurant suggestion workflow (pending, approved, tested, rejected)
- Menu photo uploads (AWS S3 or local storage)
- Cultural category management
- Food type management
- Dark/Light theme toggle

#### Infrastructure & DevOps
- Docker Compose setup for local development
- Production Docker Compose configuration
- Nginx reverse proxy with SSL/TLS support
- Automated Docker image publishing to GitHub Container Registry
- Multi-platform support (linux/amd64, linux/arm64)
- GitHub Actions CI/CD pipeline
- Automated release workflow with binary builds
- GitHub Pages documentation site
- Comprehensive deployment scripts (deploy.sh, backup.sh, restore.sh)

#### Monitoring & Logging
- Structured logging with zerolog
- Request tracing with UUID correlation
- Real-time metrics tracking
- Performance monitoring (p50, p95, p99)
- Health check endpoints
- Swagger/OpenAPI documentation

#### Documentation
- Complete deployment guide (DEPLOYMENT.md)
- Quick start guide (QUICKSTART_PRODUCTION.md)
- Production checklist (PRODUCTION_CHECKLIST.md)
- Deployment summary (DEPLOYMENT_SUMMARY.md)
- Authentication setup guide (AUTHENTICATION.md)
- Git workflow guide (GIT_WORKFLOW.md)
- Contributing guidelines (CONTRIBUTING.md)
- GitHub setup guide (GITHUB_SETUP.md)
- Automated documentation deployment via GitHub Pages

#### Database
- PostgreSQL 16 with health checks
- Database migrations support
- Users and sessions tables
- Optimized indexes
- Connection pooling

#### Testing
- Comprehensive test suite
- GitHub Actions CI with automated testing
- Security scanning with Trivy
- Code coverage reporting
- Linting for Go and TypeScript

### Technical Stack

#### Backend
- Go 1.24
- Gorilla Mux router
- pgx/v5 PostgreSQL driver
- coreos/go-oidc/v3 for OIDC
- AWS SDK v2 for S3
- rs/cors for CORS handling
- zerolog for structured logging

#### Frontend
- React 18
- TypeScript
- Vite build tool
- Tailwind CSS
- Google Maps JavaScript API integration

#### Infrastructure
- Docker and Docker Compose
- Nginx for reverse proxy
- PostgreSQL 16
- GitHub Actions for CI/CD
- GitHub Container Registry for Docker images
- GitHub Pages for documentation

### Configuration

#### Environment Variables
- `AUTH_MODE` - Authentication mode (none/local/oauth/both)
- `JWT_SECRET_KEY` - Secret key for JWT signing
- `OIDC_ISSUER_URL` - OIDC provider issuer URL
- `OIDC_CLIENT_ID` - OIDC client identifier
- `OIDC_CLIENT_SECRET` - OIDC client secret
- `OIDC_REDIRECT_URL` - OIDC redirect URL
- `DATABASE_URL` - PostgreSQL connection string
- `GOOGLE_MAPS_API_KEY` - Google Maps API key
- `AWS_ACCESS_KEY_ID` - AWS access key (optional)
- `AWS_SECRET_ACCESS_KEY` - AWS secret key (optional)
- `S3_BUCKET_NAME` - S3 bucket name (optional)
- `ALLOWED_ORIGINS` - CORS allowed origins
- `DEBUG` - Enable debug logging

### Breaking Changes
- Replaced Google OAuth with generic OIDC (requires OIDC_* environment variables instead of GOOGLE_OAUTH_*)
- AUTH_MODE now defaults to "none" for backward compatibility

### Security
- All passwords hashed with Argon2id (64MB memory, 3 iterations, parallelism 2)
- JWT tokens signed with HS256
- Access tokens expire in 15 minutes
- Refresh tokens expire in 7 days
- Rate limiting on all endpoints
- Security headers enabled
- HTTPS enforced in production

### Docker Images
- Pre-built images available at `ghcr.io/your-username/the-nom-database/backend`
- Pre-built images available at `ghcr.io/your-username/the-nom-database/frontend`
- Multi-platform support (amd64, arm64)
- Automatic tagging with version, latest, develop, and SHA

## [0.1.0] - Initial Development

### Added
- Basic restaurant management
- Simple rating system
- Google Maps integration
- Docker development environment

---

[Unreleased]: https://github.com/your-username/the-nom-database/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/your-username/the-nom-database/releases/tag/v1.0.0
[0.1.0]: https://github.com/your-username/the-nom-database/releases/tag/v0.1.0
