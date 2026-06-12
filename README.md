# Padduck

A modern IP Address Management (IPAM) platform built with Go, React, and PostgreSQL.

[![CI Test Suite](https://github.com/lima3w/padduck/actions/workflows/ci.yml/badge.svg)](https://github.com/lima3w/padduck/actions/workflows/ci.yml)

## Purpose
Replace spreadsheet-based IP tracking with a structured, API-first system.

## Stack
- Backend: Go (Fiber)
- Frontend: React (Vite)
- Database: PostgreSQL
- Deployment: Docker Compose + GitHub Container Registry

## Quick Start

You do not need to clone this repository or build images locally. The backend
and frontend images are published to GitHub Container Registry; only the Compose
file is required on the host.

```bash
mkdir padduck
cd padduck
curl -fsSLO https://raw.githubusercontent.com/lima3w/padduck/main/docker-compose.yml
curl -fsSL https://raw.githubusercontent.com/lima3w/padduck/main/.env.example -o .env
# Edit .env and set POSTGRES_PASSWORD to a strong value before continuing
docker compose pull
docker compose up -d
```

`POSTGRES_PASSWORD` is **required** — the stack will not start without it. On
first startup, the backend creates a persistent MFA encryption key in
`./data/backend/mfa-encryption-key` if `MFA_ENCRYPTION_KEY` is not set.

Open `http://localhost:3000` and log in as `admin`. The generated password is
printed to the backend log on first boot.

## Configuration

Configuration is read from environment variables. Docker Compose will also read a local `.env` file for variable interpolation if one is present — it is not created automatically.

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | `padduck` | PostgreSQL username |
| `POSTGRES_PASSWORD` | **required** | PostgreSQL password — must be set in `.env` before first run |
| `POSTGRES_DB` | `padduck` | PostgreSQL database name |
| `DATABASE_URL` | derived | Overrides the individual PostgreSQL variables |
| `ADMIN_PASSWORD` | _(generated)_ | Initial admin password; printed to logs on first boot if unset |
| `RESET_ADMIN_PASSWORD` | `false` | Force-reset the admin password on next boot |
| `MFA_ENCRYPTION_KEY` | generated if unset | Optional override; 64 hex characters; generate with `openssl rand -hex 32` |
| `SESSION_COOKIE_SECURE` | `auto` | `auto` marks session cookies secure when behind HTTPS; set `true` or `false` to override |
| `FRONTEND_PORT` | `3000` | Host port the UI is exposed on |
| `FRONTEND_BIND` | `127.0.0.1` | Network interface the frontend port binds to. Defaults to loopback; place a TLS-terminating reverse proxy in front for production. Set to `0.0.0.0` only with additional network-level access control |
| `IMAGE_TAG` | `1.31.25` | Pinned release version (GHCR tags have no `v` prefix). To upgrade, set `IMAGE_TAG=<new>` in `.env`, then run `docker compose pull && docker compose up -d` |

Update checks can be enabled under **Admin Settings → Updates**. The backend checks the GitHub releases API automatically — no configuration required.

## Documentation

Full documentation is on the [GitHub Wiki](https://github.com/lima3w/padduck/wiki), including:

- [Installation Guide](https://github.com/lima3w/padduck/wiki/Installation-Guide)
- [Configuration](https://github.com/lima3w/padduck/wiki/Configuration)
- [User Guide](https://github.com/lima3w/padduck/wiki/User-Guide)
- [API Documentation](https://github.com/lima3w/padduck/wiki/API-Documentation)
- [Troubleshooting](https://github.com/lima3w/padduck/wiki/Troubleshooting)

## Development Conventions

- **Timestamps must be written in UTC.** The schema uses `TIMESTAMP` (without
  time zone) columns: pgx stores a `time.Time`'s wall-clock digits as-is and
  reads them back as UTC, so any local-time value written to the database — or
  passed as a SQL query parameter — is wrong by the host's UTC offset. Always
  use `time.Now().UTC()` for values that reach the database. The repository
  package enforces this with a test (`repository/utc_guard_test.go`);
  service-layer code must apply the same rule when constructing times for
  repository calls. Read-side comparisons against scanned values
  (`time.Now().After(row.ExpiresAt)`) compare instants and are safe either way.

## Releasing

Releases are deliberate — pushing to `main` does **not** publish anything.
To cut a release:

1. In the release PR, add a `## vX.Y.Z` section at the top of `CHANGELOG.md`
   (the changelog gate enforces this on every PR).
2. Merge to `main`.
3. When ready to release, tag that commit and push the tag:
   ```bash
   git tag vX.Y.Z && git push origin vX.Y.Z
   ```

The tag push triggers `release.yml`, which runs the full test suite (backend,
frontend, e2e) and only then publishes the GHCR images and the GitHub release.
Multiple PRs can land on `main` between releases; they all ship in the next
tag. Image tags on GHCR have no `v` prefix (`1.31.27`).
