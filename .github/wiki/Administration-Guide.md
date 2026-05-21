# Administration Guide

---

## Admin Overview

Access the admin panel via **Admin** in the left sidebar (visible to users with admin role). The overview page groups all admin functions:

- **Identity & Access**: Users, Roles, Permissions
- **Configuration**: System settings, SMTP, session timeouts
- **Integrations & Automation**: API tokens, webhooks, automation policies
- **Discovery Operations**: Scan jobs, scan agents, discovery profiles
- **Audit & Reports**: Audit logs, scheduled reports, utilization
- **Data Tools**: Import, export, bulk operations
- **System Health**: Health check, compatibility, update check

---

## User Management

### List Users

**Admin → Users** — view all users, their roles, status, and last activity.

### Create a User

**Admin → Users → Create User** or via CSV import:

```csv
username,email,role
alice,alice@example.com,operator
bob,bob@example.com,viewer
```

`POST /api/v1/admin/users/bulk-import`

### User States

| State | Meaning | Action |
|-------|---------|--------|
| `active` | Normal access | — |
| `pending` | Awaiting admin approval | Approve or reject |
| `suspended` | Temporarily blocked | Unsuspend |
| `rejected` | Registration denied | — |
| `disabled` | Deactivated | Re-enable |

### Update User Role

User detail → **Edit** → change role dropdown.

### Suspend / Unsuspend

User detail → **Suspend** (or **Unsuspend**). Add a reason (recorded in audit log).

### GDPR Delete

User detail → **GDPR Delete** — permanently removes user data per GDPR requirements.

### Impersonation

Admins can impersonate a user to troubleshoot (`POST /api/v1/admin/users/:id/impersonate`). Impersonation is logged to the audit trail.

---

## Role Management

**Admin → Roles**

### Built-in Roles

| Role | Scope |
|------|-------|
| `admin` | Full system access |
| `operator` | Read/write IPAM, no admin |
| `viewer` | Read-only |

### Custom Roles

1. **Admin → Roles → + New Role**
2. Set name and description
3. Select permission flags
4. Save and assign to users

---

## Permissions

Permissions are assigned at the role level. Each permission covers a resource and action (e.g., `subnets:write`). See [[Security]] for the full permission list.

---

## System Configuration

**Admin → Settings** (or `GET/PATCH /api/v1/admin/config`)

| Setting | Description |
|---------|-------------|
| `registration_enabled` | Allow self-registration |
| `require_email_verification` | Email verification required before login |
| `require_admin_approval` | New accounts need admin approval |
| `privacy_policy_version` | Current policy version shown to users |
| `session_idle_timeout_minutes` | Idle session expiry (default: 60) |
| `session_absolute_timeout_hours` | Max session length (default: 168) |
| `smtp_host` | SMTP server for email |
| `smtp_port` | SMTP port (default: 587) |
| `smtp_username` | SMTP credentials |
| `smtp_from` | From address for outbound email |

Sensitive values (SMTP password, etc.) are redacted in GET responses. Submitting the redaction placeholder leaves the stored value unchanged.

---

## Tenant Administration

Tags and custom fields serve as the primary tenancy mechanism. For strict multi-tenant isolation, deploy separate Padduck instances per tenant.

---

## System Health

**Admin → System Health**

The health endpoint (`GET /health`) returns:
```json
{"status": "ok"}
```

The admin health panel shows:
- Database connectivity
- Migration status
- Background job queue depth
- Last scan job completion

---

## Maintenance Tasks

### Clear Expired Leases

**Admin → Data Tools → Release Expired IPs**

Or via API:
```bash
POST /api/v1/admin/maintenance/release-expired-leases
```

### Audit Log Retention

**Admin → Audit Retention** — configure how long audit entries are kept. Older entries are pruned automatically.

### Purge Old Scan Results

Configure retention under **Admin → Discovery → Retention Policies**.

---

## Performance Tuning

| Setting | Recommendation |
|---------|----------------|
| `SCAN_MAX_CONCURRENT_JOBS` | Increase for faster scanning on capable hosts; reduce to protect network |
| Database connection pool | Tune via `DATABASE_URL` pool parameters |
| `ENVIRONMENT=production` | Ensures JSON logging (less I/O than text logs) |
| Audit retention | Set a retention policy to limit audit table growth |

---

## Troubleshooting

See [[Troubleshooting]] for common errors and fixes.

Common admin issues:
- Users can't log in → Check account state, LDAP/SSO config
- MFA codes rejected → Verify `MFA_ENCRYPTION_KEY` is consistent across restarts
- Emails not sending → Check SMTP config in **Admin → Settings**
- Scan jobs not running → Check scan agent connectivity and token

---

## Security Hardening

1. Set a strong `POSTGRES_PASSWORD` (not the default)
2. Set `MFA_ENCRYPTION_KEY` to a random 64-char hex string
3. Enable MFA for all admin accounts
4. Set `SESSION_COOKIE_SECURE=true` when behind HTTPS
5. Configure `TRUSTED_PROXIES` to your load balancer IPs only
6. Enable `require_email_verification` and `require_admin_approval` for self-registration
7. Review audit logs regularly
8. Rotate API tokens periodically
9. Apply OS and Docker security updates to the host

---

## Upgrade Management

Before upgrading:
1. Check [[Changelog]] for breaking changes
2. Back up the database
3. Review v2 compatibility if planning a major version upgrade (`GET /api/v1/admin/compatibility/v2-readiness`)

Upgrade procedure:
```bash
git pull
docker compose up --build -d
docker compose logs backend | grep -E "migration|error"
```

Migrations run automatically on startup. Monitor logs for migration errors.

---

## Disaster Recovery

### Backup

```bash
DATABASE_URL=postgres://padduck:pass@localhost/padduck ./tools/backup.sh ./backups
```

Backups are gzip-compressed SQL dumps. Files older than 30 days are pruned automatically.

### Restore

```bash
docker compose stop backend
DATABASE_URL=postgres://padduck:pass@localhost/padduck ./tools/restore.sh ./backups/padduck_backup_YYYYMMDD_HHMMSS.sql.gz
docker compose up backend
```

### Recovery Time Objective

With database backups:
- RTO: ~10-30 minutes (restore + restart)
- RPO: Time since last backup

For near-zero RPO: use PostgreSQL streaming replication with a standby.
