# Configuration

---

## Configuration Sources

Padduck reads configuration from **environment variables** at startup. When using Docker Compose, values are interpolated from your `.env` file.

`GET /api/v1/admin/config` returns current runtime configuration (sensitive values redacted). `PATCH /api/v1/admin/config` updates settings at runtime (no restart required for most settings).

---

## Core Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | `padduck` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `padduck` | **Change before any shared deployment** |
| `POSTGRES_DB` | `padduck` | Database name |
| `DATABASE_URL` | *(derived)* | Full PostgreSQL connection string; overrides the three above |
| `SERVER_PORT` | `8080` | Port the backend listens on |
| `ENVIRONMENT` | `production` | `production` (JSON logs) or `development` (text logs) |
| `MFA_ENCRYPTION_KEY` | *(empty)* | **Required for production.** 64 hex chars. `openssl rand -hex 32` |
| `ADMIN_PASSWORD` | *(auto-generated)* | Initial admin password (auto-generated if empty) |
| `RESET_ADMIN_PASSWORD` | `false` | Set `true` to force reset on next boot; remove after use |
| `TRUSTED_PROXIES` | *(none)* | Comma-separated IPs/CIDRs trusted for `X-Real-IP` forwarding |
| `SCAN_MAX_CONCURRENT_JOBS` | `5` | Max parallel scan jobs |
| `SESSION_COOKIE_SECURE` | `auto` | `true` forces Secure flag; `auto` sets it only over HTTPS |
| `APP_VERSION` | *(empty)* | Build metadata for admin update-check panel |
| `GIT_COMMIT` | *(empty)* | Build metadata |
| `BUILD_DATE` | *(empty)* | Build metadata |

---

## Runtime Configuration (Admin Settings)

These settings are persisted in the database and configurable via **Admin → Settings** or the API.

### Authentication & Registration

| Key | Default | Description |
|-----|---------|-------------|
| `registration_enabled` | `true` | Allow self-registration |
| `require_email_verification` | `false` | Require email verification before login |
| `require_admin_approval` | `false` | New accounts need admin approval |
| `privacy_policy_version` | `1.0` | Policy version shown to users |

### Session Timeouts

| Key | Default | Description |
|-----|---------|-------------|
| `session_idle_timeout_minutes` | `60` | Minutes of inactivity before session expires |
| `session_absolute_timeout_hours` | `168` | Maximum session lifetime in hours (7 days) |

### SMTP / Email

| Key | Default | Description |
|-----|---------|-------------|
| `smtp_host` | — | SMTP server hostname |
| `smtp_port` | `587` | SMTP port |
| `smtp_username` | — | SMTP authentication username |
| `smtp_password` | — | SMTP password (redacted in GET responses) |
| `smtp_from` | — | From address for outbound email |
| `smtp_use_tls` | `true` | Use STARTTLS |

---

## Authentication Providers

### LDAP Configuration

Configure under **Admin → Identity → LDAP** or via API.

| Setting | Description |
|---------|-------------|
| Server URL | `ldap://ldap.example.com:389` or `ldaps://` |
| Bind DN | Service account DN for searches |
| Bind Password | Service account password (redacted in responses) |
| Base DN | User search base |
| User filter | LDAP filter for user lookups (e.g., `(uid=%s)`) |
| Group sync | Optional group-to-role mapping |

### OAuth2 / OIDC Configuration

Configure under **Admin → Identity → OAuth2**.

| Setting | Description |
|---------|-------------|
| Client ID | OAuth2 application client ID |
| Client Secret | OAuth2 secret (redacted in responses) |
| Authorization URL | Provider authorization endpoint |
| Token URL | Provider token endpoint |
| Userinfo URL | Provider userinfo endpoint |
| Scopes | Requested scopes (e.g., `openid email profile`) |

### SAML Configuration

Configure under **Admin → Identity → SAML**.

| Setting | Description |
|---------|-------------|
| IdP Metadata URL | Identity provider SAML metadata URL |
| SP Entity ID | Service provider entity ID |
| SP ACS URL | Assertion Consumer Service URL |
| Certificate | SP signing certificate |

---

## DNS Integration

Configure DNS provider integration under **Admin → DNS**.

| Setting | Description |
|---------|-------------|
| Provider | DNS provider type |
| API key / Token | Provider authentication (redacted in responses) |
| Zone sync enabled | Automatically sync zone data |

---

## DHCP Integration

Configure DHCP server integration under **Admin → DHCP**.

| Setting | Description |
|---------|-------------|
| Server URL | DHCP server API endpoint |
| API key | Authentication token (redacted in responses) |

---

## Notification Settings

| Setting | Description |
|---------|-------------|
| Webhook endpoints | Configured under **Admin → Webhooks** |
| SMTP | Used for email notifications and password resets |

---

## Feature Flags

Admins can enable/disable features via **Admin → Features** or `GET/PATCH /api/v1/admin/features`.

Feature flags allow:
- Disabling experimental features
- Enabling beta features for testing
- Controlling access to specific functionality per instance

---

## Backup Configuration

```bash
# Backup script reads DATABASE_URL from environment
DATABASE_URL=postgres://padduck:pass@localhost/padduck ./tools/backup.sh ./backups
```

Backup retention: files older than 30 days are pruned automatically by the backup script.

---

## Update Check Configuration

Configure private repository update checks under **Admin → Settings → Updates**:

| Setting | Description |
|---------|-------------|
| Repository URL | Your Gitea/GitHub latest-release API URL |
| API token | Read-only repository token (never sent to browser) |
| Check interval | How often to check for updates |
