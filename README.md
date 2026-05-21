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

```bash
docker compose pull
docker compose up -d
```

Open `http://localhost:3000` and log in as `admin`. The generated password is printed to the backend log on first boot.

## Configuration

Configuration is read from environment variables. Docker Compose will also read a local `.env` file for variable interpolation if one is present — it is not created automatically.

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | `padduck` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `padduck` | PostgreSQL password |
| `POSTGRES_DB` | `padduck` | PostgreSQL database name |
| `DATABASE_URL` | derived | Overrides the individual PostgreSQL variables |
| `ADMIN_PASSWORD` | _(generated)_ | Initial admin password; printed to logs on first boot if unset |
| `RESET_ADMIN_PASSWORD` | `false` | Force-reset the admin password on next boot |
| `MFA_ENCRYPTION_KEY` | _(required in production)_ | 64 hex characters; generate with `openssl rand -hex 32` |
| `SESSION_COOKIE_SECURE` | `auto` | `auto` marks session cookies secure when behind HTTPS; set `true` or `false` to override |
| `IMAGE_TAG` | `latest` | Pin to a specific release tag (e.g. `v1.30.0`) |

Update checks can be enabled under **Admin Settings → Updates**. The backend checks the GitHub releases API automatically — no configuration required.

## Documentation

Full documentation is on the [GitHub Wiki](https://github.com/lima3w/padduck/wiki), including:

- [Installation Guide](https://github.com/lima3w/padduck/wiki/Installation-Guide)
- [Configuration](https://github.com/lima3w/padduck/wiki/Configuration)
- [User Guide](https://github.com/lima3w/padduck/wiki/User-Guide)
- [API Documentation](https://github.com/lima3w/padduck/wiki/API-Documentation)
- [Troubleshooting](https://github.com/lima3w/padduck/wiki/Troubleshooting)
