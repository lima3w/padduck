
# Padduck

A modern IP Address Management (IPAM) platform built with Go, React, and PostgreSQL.

![CI Test Suite](https://gitea.lima3.dev/Lima3-Automations/padduck/actions/workflows/ci.yml/badge.svg)

## Purpose
Replace spreadsheet-based IP tracking with a structured, API-first system.

## Stack
- Backend: Go (Fiber)
- Frontend: React (Vite)
- Database: PostgreSQL
- Deployment: Docker Compose

## Run
docker compose up --build

Configuration is read from environment variables first. Docker Compose will
also read a local `.env` file for variable interpolation. Common deployment
variables:

- `POSTGRES_USER` (default `ipam`)
- `POSTGRES_PASSWORD` (default `ipam`)
- `POSTGRES_DB` (default `ipam`)
- `DATABASE_URL` (default derived from the PostgreSQL variables)
- `ADMIN_PASSWORD` (empty means generate on first boot)
- `RESET_ADMIN_PASSWORD` (default `false`)
- `MFA_ENCRYPTION_KEY` (required in production; 64 hex characters)
- `SESSION_COOKIE_SECURE` (`auto`, `true`, or `false`; unset/`auto` marks
  session cookies secure only when the request is HTTPS or forwarded as HTTPS)
- `APP_VERSION`, `GIT_COMMIT`, `BUILD_DATE` (optional build metadata used by
  admin update checks)

Private repository update checks can be configured under **Admin Settings →
Updates**. Use a read-only repository API token and the latest-release API URL
for your Gitea/GitHub repository; the token is stored server-side and is never
sent to the browser after save.

## Deployment
Automated deployment to `gitea-runner.lab` is configured via `.gitea/workflows/deploy.yml`.
- Runs tests on all pushes to `main`
- Builds and deploys both backend and frontend services
- Verifies health endpoint before marking deployment successful

## Docs
- docs/openapi.yaml
- docs/roadmap.md
- docs/user-guide.md
- docs/v1.29-network-modules.md
- docs/v1.30-optional-tools.md
- docs/v1-maintenance-policy.md
