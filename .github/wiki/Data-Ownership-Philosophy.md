# Data Ownership Philosophy

---

## Core Principle

**Your IP data is yours. Always.**

Padduck is self-hosted, open-source software. There is no cloud service, no telemetry, no license server, no mandatory external connectivity. Your network topology, IP allocations, device inventory, and audit logs live entirely on your infrastructure, under your control.

---

## What This Means in Practice

### No Phone Home

Padduck does not send any data to external services by default. The only optional external connectivity is:

- **Update checks** — polls a GitHub/Gitea API endpoint you configure to check for new releases. You choose the URL and provide your own read-only token. This can be left unconfigured.
- **SMTP** — for sending email notifications. You configure your own SMTP server.
- **Webhooks** — you configure the endpoints; Padduck delivers events to them.
- **External auth** — LDAP/OAuth2/SAML: connections to your identity providers.

All of these are **operator-configured** and **optional**.

### Audit Logs Are Yours

The audit log is stored in your PostgreSQL database. You export it. You retain it. You delete it (via retention policy). There is no third-party audit service.

### GDPR Is an Architecture Feature

GDPR compliance requirements shaped the data model:

- Users can download all their data: `GET /api/v1/auth/me/export`
- Users can request deletion: `POST /api/v1/auth/me/deletion-request`
- Privacy policy consent is tracked per user with version history
- Admins can perform GDPR-compliant account deletion: `POST /api/v1/admin/users/:id/gdpr-delete`

These aren't afterthoughts — they're built-in API endpoints.

### Open Source = No Lock-In

Because Padduck is GPL-3.0 open source:

- You can read every line of code that touches your data
- You can fork it if the project direction changes
- You can export all your data via the API or directly from PostgreSQL
- There's no proprietary data format — everything is PostgreSQL, which you can query directly

### Migration Path Always Exists

The stable v1 API contract means your automation scripts don't break. When a v2 is needed, migration tooling is built **before** the v2 release:

- `GET /api/v1/admin/compatibility/v2-readiness` — migration readiness check
- `GET /api/v1/admin/export/v2-migration-bundle` — export your data in v2 format

You control when to migrate. The old version keeps working in the meantime.

---

## What We Will Never Do

- Send your IP data or network topology to external services
- Require a cloud account or license verification to function
- Include analytics, crash reporting, or usage tracking without explicit opt-in
- Lock your data in a proprietary format that can't be exported

---

## For Operators

As a Padduck operator, you are the data controller for your users' data. Your responsibilities:

1. Keep Padduck updated (security patches)
2. Secure the PostgreSQL database (encryption at rest, access control)
3. Manage backup and retention of the database
4. Configure `MFA_ENCRYPTION_KEY` to protect MFA secrets
5. Review and act on GDPR deletion requests from your users

Padduck provides the tools; you provide the operational discipline.

---

## Related

- [Security](Security) — How Padduck protects data in transit and at rest
- [Licensing and Legal](Licensing-and-Legal) — Privacy policy framework
- [Why Padduck Exists](Why-Padduck-Exists) — Why self-hosted, open-source IPAM matters
