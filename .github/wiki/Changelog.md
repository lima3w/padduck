# Changelog

This page summarizes major releases. The full `CHANGELOG.md` is in the [repository](https://github.com/lima3w/padduck).

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
