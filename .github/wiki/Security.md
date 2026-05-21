# Security

---

## Security Overview

Padduck is designed with security as a core concern, not an afterthought. Key security properties:

- All writes are authenticated and authorized
- All significant actions are audited
- Sensitive values are redacted in logs and API responses
- MFA is available for all accounts
- RBAC controls access at the endpoint level
- SSRF protections are enforced on outbound HTTP calls

---

## Threat Model

| Threat | Mitigation |
|--------|-----------|
| Unauthorized access | Authentication required for all non-public endpoints |
| Privilege escalation | RBAC checked server-side on every request |
| Credential theft | MFA, session timeout, account lockout |
| Brute force | Rate limiting per IP and per account |
| Session hijacking | HttpOnly, Secure, SameSite=Lax cookies |
| SSRF via webhooks | URL validation against private/loopback ranges |
| SQL injection | Parameterized queries via SQLC (no raw SQL concatenation) |
| XSS | React's built-in output encoding; strict CSP headers |
| Audit tampering | Audit log is append-only; no delete endpoint |
| Sensitive data exposure | Passwords, API keys, tokens redacted in logs and responses |

---

## Authentication Security

### Password Security

- Passwords are hashed with a modern, salted algorithm (never stored in plaintext)
- Minimum complexity requirements enforced
- Account lockout after repeated failures (configurable threshold)
- Rate limiting on login attempts per source IP

### Session Management

- Session cookies: `HttpOnly`, `Secure` (in production), `SameSite=Lax`
- `SESSION_COOKIE_SECURE=auto` — sets Secure flag only when request arrives over HTTPS
- Configurable idle timeout (`session_idle_timeout_minutes`, default 60)
- Configurable absolute timeout (`session_absolute_timeout_hours`, default 168)
- Sessions invalidated on logout

### MFA

- TOTP (Time-based One-Time Password) per RFC 6238
- QR code setup with authenticator app compatibility (Google Authenticator, Authy, 1Password, etc.)
- Backup codes for account recovery
- MFA secrets encrypted at rest using `MFA_ENCRYPTION_KEY`
- If `MFA_ENCRYPTION_KEY` is not set in production, a per-process random key is used (secrets lost on restart)

---

## Authorization Model

Padduck uses **Role-Based Access Control (RBAC)**:

### Built-in Roles

| Role | Access Level |
|------|-------------|
| `admin` | Full access including admin endpoints |
| `operator` | Read/write IPAM resources, no admin functions |
| `viewer` | Read-only access |

### Custom Roles

Admins can create custom roles with specific permission combinations. Permission flags include:

- `subnets:read`, `subnets:write`
- `ip-addresses:read`, `ip-addresses:write`
- `vrfs:read`, `vrfs:write`
- `devices:read`, `devices:write`
- `dns:read`, `dns:write`
- `audit:read`
- `admin:users`, `admin:config`, `admin:roles`
- And more...

### Break-Glass Access

Emergency admin access is available when primary authentication fails (e.g., LDAP outage). Break-glass access is logged to the audit trail.

---

## Encryption Standards

| Data | Encryption |
|------|-----------|
| MFA secrets | AES-GCM with `MFA_ENCRYPTION_KEY` (64 hex chars) |
| Passwords in transit | TLS (enforce at reverse proxy) |
| Passwords at rest | Bcrypt hash |
| API tokens | Stored as hashes; shown once at creation |
| Database | Encrypted at rest via disk/volume encryption (infrastructure responsibility) |

---

## Secret Management

| Secret | How to Manage |
|--------|--------------|
| `MFA_ENCRYPTION_KEY` | Generate with `openssl rand -hex 32`; store in secrets manager or `.env` |
| `POSTGRES_PASSWORD` | Change from default before any shared deployment |
| API tokens | Created per-user; rotated from My Settings |
| Webhook secrets | Admin-configured; used for HMAC-SHA256 signature |
| LDAP bind password | Stored in admin config; redacted in API responses |
| SMTP password | Stored in admin config; redacted in API responses |

**Key rotation**: Rotating `MFA_ENCRYPTION_KEY` invalidates all existing MFA enrollments. Plan rotation carefully with a re-enrollment window.

---

## Audit Logging

All significant actions generate an audit log entry:

- User identity (username + user ID)
- Source IP address
- Timestamp (UTC)
- Object type and ID
- Action (create, update, delete, login, logout, etc.)
- Changed field values

**Sensitive values are redacted** in audit payloads:
- Passwords
- API tokens and keys
- SNMP community strings
- SMTP passwords
- MFA secrets

Audit logs are append-only. There is no API endpoint to delete audit entries. Configure retention via **Admin → Audit Retention**.

---

## Session Management

```
POST /auth/login
→ Session cookie issued (HttpOnly, Secure, SameSite=Lax)

Idle timeout: 60 minutes (configurable)
Absolute timeout: 168 hours / 7 days (configurable)

POST /auth/logout
→ Session invalidated server-side
```

All active sessions are tracked and can be viewed/revoked under **My Settings → Sessions** (if enabled).

---

## Vulnerability Disclosure Policy

To report a security vulnerability in Padduck:

1. **Do not** open a public issue
2. Email the maintainers directly (see repository contact info)
3. Include: description, reproduction steps, affected versions, potential impact
4. Allow reasonable time for a fix before public disclosure

---

## Secure Development Practices

- All inputs validated server-side
- Parameterized queries via SQLC (no string-concatenated SQL)
- No sensitive values in logs
- Outbound HTTP calls validated against private/loopback ranges (SSRF prevention)
- Security linting via `gosec` in CI
- Vulnerability scanning via `govulncheck` in CI
- All dependencies vendored for supply chain integrity

---

## Compliance Goals

- **GDPR**: User data export (`GET /auth/me/export`), deletion request (`POST /auth/me/deletion-request`), privacy policy consent tracking
- **Audit trails**: Immutable audit log supports compliance requirements
- **Access control**: RBAC supports least-privilege access models
