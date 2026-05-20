# IPAM Next — User Guide

## Overview

IPAM Next is a web-based IP Address Management system. It lets you organize your network space into **Sections** → **Subnets** → **IP Addresses**, manage VRFs and VLANs, track discovery scans, and administer users.

---

## Getting Started

### Login

Navigate to the application URL and enter your username and password. If Multi-Factor Authentication (MFA) is enabled on your account you will be prompted for your TOTP code after the password step.

### First Boot

On first boot the admin password is auto-generated and printed to the server log:

```
========================================
  Admin password (first boot):  <password>
  Set ADMIN_PASSWORD env var to override.
========================================
```

Set `ADMIN_PASSWORD` in your environment to use a specific password, or `RESET_ADMIN_PASSWORD=true` to force-reset it.

When deployed with Docker Compose, host or runner environment variables are
used first. Compose can also interpolate values from a local `.env` file. The
most common deployment variables are `POSTGRES_USER`, `POSTGRES_PASSWORD`,
`POSTGRES_DB`, `DATABASE_URL`, `ADMIN_PASSWORD`, `RESET_ADMIN_PASSWORD`, and
`MFA_ENCRYPTION_KEY`.

---

## Sections

A **Section** is a top-level grouping for subnets (e.g. "Data Center", "Cloud", "Office").

| Action | How |
|--------|-----|
| List sections | Sidebar → Sections |
| Create section | Sections page → **+ New Section** |
| Edit / delete | Section row → kebab menu |

---

## Subnets

Subnets belong to a Section and define a CIDR block.

| Field | Description |
|-------|-------------|
| Network address | e.g. `10.0.0.0` |
| Prefix length | e.g. `24` for /24 |
| Description | Optional free text |

### Subnet Utilization

Each subnet shows utilization as assigned / total IPs. Navigate to a subnet to see the breakdown.

---

## IP Addresses

Each IP address lives inside a subnet and has one of three statuses:

| Status | Meaning |
|--------|---------|
| `available` | Not yet used |
| `assigned` | In use by a host or service |
| `reserved` | Held but not actively assigned |

### Allocate Next Available IP

Use **Allocate** on a subnet to automatically assign the next free IP. Provide the `assigned_to` value (hostname, service name, etc.).

### Leased IPs

An IP can be assigned with an expiry date. When the lease expires the IP can be released with **Release Expired**.

---

## VRFs and VLANs

**VRFs** (Virtual Routing and Forwarding instances) allow you to group subnets and VLANs by routing domain.

**VLANs** can be associated with a VRF. Each VLAN has a numeric ID (1–4094) and a name.

---

## Discovery (Network Scanning)

IPAM Next can scan your subnets to detect live hosts using ICMP ping.

### Creating a Scan Job

1. Admin panel → **Scan Jobs** → **+ New Scan Job**
2. Name the job and select one or more subnets
3. Optionally set a **cron schedule** (e.g. `0 2 * * *` for 2 AM daily)
4. Click **Create**

### Running a Scan

- Click **Run Now** to start immediately
- Scheduled jobs run automatically based on their cron expression

### Scan Results

Results are stored per-IP with liveness status and response time. View them from the scan job detail page or the subnet page.

---

## API Tokens

Generate API tokens for programmatic access.

1. **My Settings** → **API Tokens** → **Create Token**
2. Choose a descriptive token name
3. Copy the token — it is only shown once

Authenticate API requests with:
```
Authorization: Bearer <token>
```

### Token Scopes

| Scope | Allowed operations |
|-------|--------------------|
| `read` | GET requests only |
| `write` | GET + POST + PUT + DELETE for IPAM resources |
| `admin` | Full access including admin endpoints |

Admins can review token usage from **Admin** → **Integrations**. The view shows
usage counts, last-used details, and the active per-token rate limit so
automation owners can see whether workflows are close to throttling.

---

## Automation and Integrations

Use **Admin** → **Integrations** to create API tokens, review common platform
templates, inspect token usage, and see active automation policies.

### Webhook Delivery Controls

Use **Admin** → **Webhooks** to configure outbound subscriptions. Each endpoint
can subscribe by event wildcard, object type, tag filter, and simple
`key=value` conditions. Recent deliveries include payload inspection data, and
failed or retrying deliveries are grouped by endpoint, event, status, and error.
Use **Replay** to reset a failed delivery to `pending` after the downstream
system is healthy.

### Inbound Automation Endpoints

Bearer-token clients can use controlled automation endpoints for common
workflows:

| Workflow | Endpoint |
|----------|----------|
| Allocate next IP | `POST /api/v1/automation/ip-addresses/allocate` |
| Reserve specific IP | `POST /api/v1/automation/ip-addresses/reserve` |
| Release IP | `POST /api/v1/automation/ip-addresses/:id/release` |
| Validate DNS update | `POST /api/v1/automation/dns/update` |
| Register device | `POST /api/v1/automation/devices/register` |
| Evaluate policy | `POST /api/v1/automation/policies/evaluate` |

All automation write endpoints accept `dry_run: true` to evaluate policy
decisions without committing the change.

### Automation Policies

Admins can create simple approval and validation policies with:

| Field | Values |
|-------|--------|
| `workflow` | `ip_address`, `dns`, `device`, or `*` |
| `action` | `allocate`, `reserve`, `release`, `update`, `register`, or `*` |
| `effect` | `allow`, `deny`, or `manual_review` |
| `conditions` | Exact or prefix fields such as `subnet_id=42` or `hostname=prod-*` |

---

## Multi-Factor Authentication (MFA)

Enable TOTP-based MFA under **My Settings** → **Security**:

1. Click **Enable MFA**
2. Scan the QR code with your authenticator app
3. Enter a TOTP code to confirm
4. Save your **backup codes** in a safe place — they can be used if you lose access to your authenticator

---

## User Management (Admin)

Accessible via **Admin** → **Users**.

### Actions

| Action | Endpoint |
|--------|----------|
| List users | `GET /api/v1/users` |
| Create user | `POST /api/v1/admin/users` |
| Update role | `PUT /api/v1/users/:id/role` |
| Update email | `PUT /api/v1/admin/users/:id/email` |
| Suspend user | `POST /api/v1/admin/users/:id/suspend` |
| Unsuspend | `POST /api/v1/admin/users/:id/unsuspend` |
| Impersonate | `POST /api/v1/admin/users/:id/impersonate` |
| Bulk import (CSV) | `POST /api/v1/admin/users/bulk-import` |
| GDPR delete | `POST /api/v1/admin/users/:id/gdpr-delete` |

### CSV Import Format

```csv
username,email,role
alice,alice@example.com,user
bob,bob@example.com,viewer
```

The `role` column is optional and defaults to `user`.

---

## GDPR / Data Privacy

### Export your data

`GET /api/v1/auth/me/export` — downloads a JSON file with all your data.

### Request account deletion

`POST /api/v1/auth/me/deletion-request` — flags your account for deletion review by an admin.

### Privacy policy

Use **My Settings** → **Privacy** to review the current privacy policy version,
the version recorded on your account, and accept the current policy if your
recorded consent is out of date.

API clients can accept the current privacy policy via
`POST /api/v1/auth/me/accept-privacy`. The current version is available at
`GET /api/v1/privacy-policy/version`.

---

## Audit Logs

All significant actions are recorded. Admins can view and export logs from **Admin** → **Audit Logs**.

Filters: user, action, date range, IP address, status.

---

## Database Backup & Restore

### Backup

```bash
DATABASE_URL=postgres://user:pass@host/db ./scripts/backup.sh ./backups
```

Creates a gzip-compressed SQL dump in `./backups/`. Backups older than 30 days are pruned automatically.

### Restore

```bash
DATABASE_URL=postgres://user:pass@host/db ./scripts/restore.sh ./backups/ipam_backup_20260507_120000.sql.gz
```

Prompts for confirmation before overwriting the database.

---

## Configuration Reference

Set via Admin → Config or environment variables.

| Key | Description | Default |
|-----|-------------|---------|
| `registration_enabled` | Allow self-registration | `true` |
| `require_email_verification` | Require email verification | `false` |
| `require_admin_approval` | Require admin approval for new users | `false` |
| `privacy_policy_version` | Current policy version shown to users | `1.0` |
| `session_idle_timeout_minutes` | Idle session timeout | `60` |
| `session_absolute_timeout_hours` | Absolute session timeout | `168` |
| `smtp_host` | SMTP server hostname | — |
| `smtp_port` | SMTP port | `587` |
| `smtp_username` | SMTP credentials | — |
| `smtp_from` | From address for emails | — |

---

## API Reference

The full OpenAPI specification is available at `GET /api/openapi.yaml`.
