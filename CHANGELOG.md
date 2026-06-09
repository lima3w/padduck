# Changelog

## v1.31.22

### Bug Fixes
- **Scan auto-add: IP address displayed with /32**: IP addresses were selected from the database using `address::text`, which includes the CIDR prefix for host addresses (e.g. `192.168.1.5/32`). Changed to `host(ip.address)` — matching the existing pattern in the subnets repository — to return the bare address.
- **Scan auto-add: dns_name not populated**: After auto-adding a discovered IP, `SetIPAddressPTRFromScan` was not called on the new record, so `dns_name` was never set even when a PTR record was resolved. Now called immediately after a successful auto-add (requires "Discover reverse DNS" enabled on the job and PTR records present in DNS).

## v1.31.21

### Bug Fixes
- **Dark mode user menu hover**: User menu dropdown items used `dark:hover:bg-gray-700` (bright gray) against a navy background forced by the global dark mode override, producing a blown-out highlight. Changed hover to `#0d2848` (pd-700) to match the sidebar.

## v1.31.20

### Bug Fixes
- **Scan auto-add IPs**: Fixed auto-discovered IPs never being saved — the insert used status `"active"`, which violates the DB `CHECK` constraint; corrected to `"available"`. Affects both local-scan and remote-agent paths.
- **Dark mode list selection**: Selected and hovered list rows in Scan Jobs and Admin Roles pages were unreadable in dark mode (white text on pale-blue background); added `dark:bg-blue-900/20` to fix contrast.

## v1.31.19

### Bug Fixes
- **Frontend healthcheck**: Fixed the compose healthcheck using `wget`, which is not available in the `nginx:1.31.1-trixie` (Debian-based) image; switched to `curl`.

## v1.31.18

### Features
- **Scan job improvements**: Run Now button stays disabled until the scan finishes; scan type, concurrency, schedule, auto-add IPs, discover DNS, and overwrite DNS options are now configured directly on the job in a unified Settings tab — separate Scan Profiles removed.
- **Auto-add IPs from scans**: Active IPs discovered during a scan are automatically added to the appropriate subnet when the job's auto-add option is enabled.
- **Per-job DNS options**: New `discover_dns` and `dns_overwrite` flags on scan jobs control whether PTR records are looked up and whether existing `dns_name` values are overwritten.
- **Scan results enhancements**: Added a "Hide down" toggle to filter non-alive IPs; IPs that were alive in the previous scan but are now gone display an amber "Gone" badge.
- **Scan Retention fix**: The retention settings tab no longer errors on a fresh installation — defaults are inserted automatically if no settings row exists.

### Changes
- Renamed "Sections" to "Networks" throughout the application (UI labels, API routes, database table, and all code references). Migration: `20260609_003_rename_sections_to_networks`.
- Fixed topology view showing a double CIDR (e.g. `10.0.0.0/32/24`) by using PostgreSQL `host()` instead of `::text` cast on INET columns.

### Bug Fixes
- **Subnet split**: The original subnet is now deleted after a split instead of being kept as a container, preventing overlapping address space. IPs are moved to the correct child subnet during the transaction.
- **Subnet split blocking**: If any existing IP falls on a network or broadcast address of a child subnet, the split is blocked and the conflicting IPs are shown to the user.

### Database Migrations
- `20260609_002_scan_job_dns_options`: Adds `discover_dns` (default `true`) and `dns_overwrite` (default `false`) columns to `scan_jobs`.
- `20260609_003_rename_sections_to_networks`: Renames the `sections` table to `networks` and the `section_id` column to `network_id`.

## v1.31.17

### Build
- Updated frontend runtime image from `nginx:1.28.3-alpine` to `nginx:1.31.1-trixie` (Debian-based).
- Replaced `apk upgrade` with `apt-get upgrade` to pull latest Debian security patches.
- Switched healthcheck from `wget` to `curl` to match the Debian nginx image toolset.

## v1.31.16

### Build
- Updated frontend builder image from `node:22.16-alpine` to `node:22.22.3-alpine`.
- Updated frontend runtime image from `nginx:1.28-alpine` to `nginx:1.28.3-alpine`.
- Added `apk upgrade` to the nginx stage to pull in latest Alpine package security patches (nginx, libcrypto3, libssl3, curl, musl, nghttp2-libs, zlib, xz-libs, tiff, libpng, libxml2, and others).

## v1.31.15

### Build
- Updated backend Dockerfile base image from `golang:1.26.3-alpine` to `golang:1.26.4-alpine` to match the go directive.

## v1.31.14

### Dependencies
- Updated frontend npm packages: axios 1.16.1→1.17.0, cytoscape 3.33.3→3.34.0, react/react-dom 19.2.6→19.2.7, react-router-dom 7.15.1→7.17.0, vite 8.0.13→8.0.16, vitest 4.1.6→4.1.8.
- Added `eslint:recommended` to the flat ESLint config; added vitest globals to test file config; fixed undeclared `historyError` state in the utilization history section.
- Updated Go backend dependencies: jackc/pgx/v5, x/crypto, x/net, x/sys, x/text, x/sync, go-sqlite3, go-runewidth, go-colorable, go-internal.
- Bumped the `go` directive to 1.26.4 in both the backend and agent modules.

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
