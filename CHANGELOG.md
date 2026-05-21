# Changelog

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
