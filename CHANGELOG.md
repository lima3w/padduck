# Changelog

## v1.31.1

- Fixed admin password file writing to `data/admin-password` in the working directory instead of `/run/ipam`, which required root filesystem access.
- Added `entrypoint.sh` to fix bind-mount ownership at startup; backend container now drops to an unprivileged user via `su-exec`.
- Fixed gosec `#nosec` annotation referencing wrong rule (`G306` → `G703`).
- Fixed `feature_firewall_enabled` missing from the allowed config keys, causing a "unknown config key" error on save.

## v1.31.0

- Added firewall zones and firewall zone mappings.
- Updated frontend assets: new logo, favicon set, and web app manifest icons.

## v1.30.0 Padduck Rebrand and GitHub Migration

- Renamed project to Padduck across all surfaces: Go module, frontend metadata, storage keys, API artifacts, docs, and deployment config.
- Migrated CI/CD from Gitea Actions to GitHub Actions.
- Docker images now built and published to GitHub Container Registry (`ghcr.io/lima3w/padduck-backend`, `ghcr.io/lima3w/padduck-frontend`).
- `docker-compose.yml` updated to pull images from ghcr.io instead of building locally.
- Added docs: `index.md`, `getting-started.md`, `troubleshooting.md`.

## v1.26.0 API And SDK Stabilization

- Froze the stable v1 OpenAPI contract at `1.26.0`.
- Added retry-safe idempotency keys for automation write endpoints.
- Standardized validation error responses with field-level details.
- Versioned outbound webhook event payloads and added a sample payload endpoint.
- Added generated API client examples for JavaScript and Python.
- Added OpenAPI contract tests and changelog automation.

<!-- api-contract:1.26.0 -->

API contract snapshot:

- OpenAPI version: `1.26.0`
- Public API path count: `194`
- OpenAPI SHA-256: `12b186d567d9`
