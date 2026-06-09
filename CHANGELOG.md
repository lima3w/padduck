# Changelog

## v1.31.13

### Authentication
- Added a **Change Password** option under User → Security settings, allowing users to update their password without admin intervention.

### Network / IP Addresses
- The New IP Address form now pre-fills the network portion of the IP field based on the selected subnet, reducing manual entry errors.
- Fixed IP addresses not appearing in subnet list views due to a field key mismatch and an incomplete paginated query.
- Fixed an encode error in the split-subnet IP move query.

### Configuration
- Added `anonymous_api_enabled` to the allowed configuration keys and included it in the seed migration so it is recognized on fresh installs.

## v1.31.12

### Authentication
- The login page now hides the self-registration link when self-registration is disabled.
- Public instance metadata now exposes the registration-enabled state so unauthenticated pages can reflect the current setting without exposing admin configuration.

### Release / CI
- Main-branch release automation now requires the top changelog entry to increment the latest release by exactly one patch version, creates the matching tag, and publishes versioned release images.
- Frontend CI and Docker builds now use npm 11.16.0.

## v1.31.11

### Installation
- Updated the Docker Compose PostgreSQL bind mount for `postgres:18` from `/var/lib/postgresql/data` to `/var/lib/postgresql`, matching the official PostgreSQL 18 image layout.
- Existing deployments that already started with the previous mount should verify where their database files were initialized before changing mounts on a live system.

## v1.31.10

### Installation
- Docker Compose installs now require only `docker-compose.yml` by default. The backend and frontend images are pulled from GitHub Container Registry, and `.env` is only needed for overrides.
- Updated the README and getting-started guide to document the compose-only install path.
- `.env.example` now keeps built-in Compose defaults commented out so local config files can stay focused on explicit overrides.

### MFA
- Production deployments without `MFA_ENCRYPTION_KEY` now create and reuse a persistent backend-managed key at `./data/backend/mfa-encryption-key`.
- Migration readiness checks now accept either an explicit `MFA_ENCRYPTION_KEY` or the backend-managed persistent key file.
- Troubleshooting and user documentation now explain how to preserve or restore the MFA key.

### Security / CI
- Scoped MFA key file access and backup data-file reads with Go's root-scoped filesystem API.
- Updated CI to Go 1.26.4 so `govulncheck` runs against the fixed standard library.

## v1.31.9

### Backups
- Added a unified **Backups** page (`/admin/backups`) that consolidates Data Export, Data Import, and the new complete system backup.
- New **Download Complete Backup** produces a ZIP archive containing the full PostgreSQL database dump, all admin configuration settings (as JSON), and any files stored in `./data/` (avatars, agent binary, etc.).
- New **Restore from Backup** — upload a backup ZIP to restore the database, configuration, and files. Includes a two-step confirmation to prevent accidental data loss.
- The existing CSV/JSON data export and CSV import (subnets, IP addresses, phpIPAM) are now embedded in the Backups page as sub-sections.
- "Backups" sidebar link added under Admin; the Admin Overview now shows a single Backups card instead of separate Export and Import cards.

### DNS Integration
- **Auto-add IPs**: when enabled (`dns_auto_add_ips_enabled`), the DNS sync picks up A/AAAA records from the configured DNS provider and inserts any IP not already in IPAM into the matching subnet.
- **Auto-remove IPs**: when enabled (`dns_auto_remove_ips_enabled`), IPs previously added by the DNS sync that are no longer present in DNS records are removed from IPAM.
- Both options are toggleable in Admin → Settings → DNS. They default to disabled.

### Discovery / Scan Jobs
- **Auto-add discovered IPs**: new per-job toggle (enabled by default) — when a scan finds a live IP address that is not yet in IPAM, it is automatically added to the matching subnet.
- The toggle is exposed in the scan job creation and edit form.

### SNMP
- Added a **show/hide toggle** (eye icon) next to the global SNMP community string field in Admin → Settings so the value can be revealed without having to clear and retype it.

### Audit Log
- Fixed the **Prune** button on the Audit Retention page — the `POST /api/v1/admin/audit/prune` route was not registered; it now correctly calls the prune handler.
- Fixed the **export cap**: the repository previously reset any limit above 1,000 to 100, meaning CSV exports silently returned only 100 rows. The cap is now 100,000.
- Fixed the **Save** button on the Audit Retention page — if the settings row did not exist yet, the UPDATE matched nothing and returned an error; the handler now upserts the row before updating.
- Changed the minimum retention period from 1 to 30 days in the UI to match the database constraint.

### Nameservers
- Removed the hard-coded admin-only block from the Nameservers page. Non-admin users who have been granted nameserver permissions via RBAC can now view the list. Write controls (Add, Edit, Delete) remain admin-only.
- A clear "access denied" message is shown if the API returns 403 instead of a generic error.

### Scan Agents
- New **Download latest scan agent binary** section at the top of the Scan Agents admin page with one-click links for Linux x64/ARM64, macOS x64/ARM64, and Windows x64.
- Agent privilege model documented in `agent/PRIVILEGES.md` — the agent uses the system `ping` binary (no raw sockets), so elevated privileges are not required for the agent process itself.

### CI / Release
- GitHub Actions release workflow now builds the scan agent for all five platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) and attaches the binaries as assets to the GitHub Release.

## v1.31.3

### Navigation
- Removed Admin Overview sidebar item.
- Consolidated Users, Roles, and Permission Presets into a single "Users & Roles" page with tabs. Old routes redirect automatically.

### Admin Settings
- "Test PowerDNS Connection" button moved inside the PowerDNS section card.
- Feature toggle rows now render correctly in dark mode (hover no longer obscures text).
- Feature toggles that are absent from the database now default to enabled, preventing a save from accidentally disabling them.

### System Health
- Removed Quick Links section.
- Removed "Requires pg_dump to be installed in the backend container" note — pg_dump is now included in the backend Docker image.

### Backup
- `postgresql17-client` added to the backend Docker image so pg_dump works out of the box.

### Discovery
- Scan job creation now accepts a network address / CIDR (e.g. `192.168.1.0/24`) and resolves it to the matching subnet automatically. No more "at least one subnet id is required" error.

### Bug Fixes
- Audit retention settings no longer return a 500 error when the default row doesn't exist yet — the row is now created reliably without relying on `RETURNING` from a no-op insert.
- Subnet IP addresses page: removed Data Quality section. Utilization History no longer shows an error on empty data — it shows a friendly empty state instead.

## v1.31.2

### UI / UX
- Added version number to the bottom of the sidebar navigation (admin users only).
- Features settings page now reloads back to the Features tab after saving.
- FeatureGate "feature disabled" message now includes a direct link to Admin → Settings → Features.
- Role Management NavLink no longer incorrectly highlights when viewing Permission Presets.
- Break Glass moved from a standalone sidebar link to a tab on the Admin Users page.
- Audit retention settings are now auto-initialized on fresh installs (no more 500 error on first visit).
- Discovery sidebar entry consolidated into a single tabbed page (Scan Jobs, Scan Profiles, Scan Retention, Topology Hints, Conflicts).
- Integration Templates removed from sidebar navigation.
- Privacy Consent removed from sidebar; a Privacy Policy link added to the user menu above Logout.
- System Health: Migrations section removed; Scan Agents card shows a friendly message when no agents are registered.
- Backup & Restore: renamed from "Rehearsal"; added a Download Backup (.sql) button that streams a pg_dump.
- Modal dialogs no longer steal focus back to the modal container on every re-render; text inputs retain focus correctly.

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
