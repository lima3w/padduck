# Changelog

This page summarizes major releases. The full `CHANGELOG.md` is in the [repository](https://github.com/lima3w/padduck).

---

## v1.32.x — Opt-In Telemetry

- Anonymous, opt-in usage telemetry: off by default, admin-enabled under Admin Settings → Telemetry, with a one-time setup page nudging new admins to make an explicit choice on first login
- Snapshot contains only aggregate counts and percentages (e.g. total subnets, utilization percentiles, which optional features are enabled) — never IPs, hostnames, MAC addresses, usernames, or other identifying data
- Telemetry destination and auth are hardcoded to a Padduck-operated endpoint with a public write-only key, so there is no URL/token to configure or leak
- American spelling consistency pass across the API, database, and code (`/utilisation/history` renamed to `/utilization/history`)
- Dashboard utilization now calculated from real subnet address capacity instead of raw IP record counts

---

## v1.31.x — Security Hardening, Testing Investment, and DHCP Integration

- Major security pass: timing-safe login, fixed an account-lockout enumeration bug, validated avatar uploads, fixed a stored-XSS hole in custom fields, added a CSP header, SSRF regression tests, and a hardened scan agent (runs as non-root, caps server-supplied CIDR size, requires opt-in for plain-HTTP servers)
- Fixed a privilege-escalation bug where scoped API tokens (e.g. read-only) could reach admin-only endpoints despite their non-admin scope
- "Sections" renamed to "Networks" throughout the application — UI labels, API routes, the database table, and all code references
- Technitium DHCP integration: lease sync, scope import, and per-IP reservation push from the subnet and IP address pages
- IP address management UX overhaul: full CIDR range view (including unrecorded addresses), sortable columns, MAC address validation/normalization, and subnet-boundary validation on IP creation
- Hardened Docker Compose deployment: dropped capabilities, read-only root filesystems, required secrets (no more default database password), pinned image tags
- Large investment in automated testing: DB-backed integration tests, full migration up/down/up verification in CI, expanded MFA/auth/discovery service coverage, and a new Playwright end-to-end suite
- Numerous UTC time-handling bugs fixed across sessions, MFA challenges, and token expiry that broke functionality on non-UTC hosts

---

## v1.30.x — Optional Tools

- Added optional tools framework controlled by admin feature flags
- Enhanced admin panel with unified overview grouping all admin functions
- Improved integration templates for common automation workflows
- API token analytics with rate-limit visibility
- V2 migration bundle export (zip archive with manifest, JSON, and CSV exports)

---

## v1.29.x — Network Modules

- Network modules: Firewall zones, NAT rules, circuit tracking
- BGP autonomous system tracking
- Enhanced device management with relationship panels
- Locations and racks with device assignment tracking
- Topology visualization improvements (Cytoscape-powered)

---

## v1.26.x — API Contract Freeze

- **API contract frozen at OpenAPI 1.26.0** (`x-api-contract: stable-v1`)
- Idempotency key support for all automation write endpoints
- Webhook signature verification (`X-IPAM-Signature-256`)
- CSV export with formula-prefix escaping (OWASP CSV injection protection)
- Sensitive value redaction in audit payloads and admin config responses
- SSRF protection for webhook and update-check URLs
- V2 compatibility warning endpoints

---

## Earlier Releases

Key milestones in earlier development:

- **Branding refresh** — project renamed to Padduck, full UI rebrand
- **SAML2 authentication** — enterprise SSO support
- **Scan agents** — remote network discovery agents
- **Webhook system** — outbound event subscriptions with replay
- **Custom fields** — admin-defined extensible attributes
- **GDPR tools** — user data export and deletion workflows
- **Audit retention** — configurable log pruning policies
- **Break-glass access** — emergency admin bypass
- **Automation policies** — approval and validation workflows
- **Reports** — utilization trends, inactive IPs, duplicate detection
- **Grafana integration** — metrics dashboard support

---

## Breaking Changes Policy

Breaking changes to the v1 API are **not permitted**. If a breaking change is required, it will be introduced in v2 with a migration path.

Backward-compatible additions (new optional fields, new endpoints) are allowed in v1 releases.

---

## Migration Notes

Before any major upgrade:
1. Read the Changelog (this page) for breaking changes
2. Back up the database
3. Check v2 compatibility at **Admin → V2 Compatibility** (if planning a major version jump)

See [Troubleshooting](Troubleshooting) for recovery procedures.
