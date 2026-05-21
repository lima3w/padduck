# Troubleshooting

This page covers errors with direct evidence in the codebase — error strings, log output, or error-handling blocks. Each entry includes what you will see, why it happens, and how to fix it.

---

## Startup errors

### `Failed to connect to database`

**Where you see it:** Backend container logs at startup.

```
Failed to connect to database: <detail>
```

**Why it happens:** The backend could not open a connection to PostgreSQL. Common causes:

- The `db` service is not yet healthy when the backend starts (should not happen with the provided `docker-compose.yml`, which waits for the `db` healthcheck)
- `DATABASE_URL` is set but points to the wrong host, port, or credentials
- `POSTGRES_PASSWORD` in your `.env` does not match the password the `db` container was initialized with

**Fix:**

1. Check that the `db` container is running and healthy: `docker compose ps`
2. Verify `DATABASE_URL` (or `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB`) in your `.env`
3. If you changed the password after first boot, destroy the volume and re-create: `docker compose down -v && docker compose up --build`

---

### `Failed to run migrations`

**Where you see it:** Backend container logs at startup, immediately after the database connection succeeds.

```
Failed to run migrations: <detail>
```

**Why it happens:** The backend applies database migrations on every boot. This error means a migration could not run — usually a permissions problem or a partially-applied migration from a crashed previous run.

**Fix:**

1. Check the full error detail in the log for which migration file failed
2. Confirm the database user has `CREATE TABLE`, `ALTER TABLE`, and `CREATE INDEX` privileges
3. If a previous run crashed mid-migration, you may need to manually inspect the `schema_migrations` table and remove the stuck entry, then restart

---

### `Failed to initialize admin password`

**Where you see it:** Backend container logs at startup.

```
Failed to initialize admin password: <detail>
```

**Why it happens:** The backend could not set or generate the admin user's password during boot. This typically follows a database connectivity or migration error.

**Fix:** Resolve any earlier database errors first. If those are clear, check whether the `admin` user row exists in the database.

---

## Admin password

### Generated password not printed to the log

**Where you see it:** Startup log; you expect the banner but don't see it.

**Why it happens:** The banner is only printed **on first boot** — when the admin user's password has never been set. On subsequent restarts the banner is suppressed because the password is already initialized.

**Fix:** Use `RESET_ADMIN_PASSWORD=true` to force-reset:

```bash
# In your .env:
RESET_ADMIN_PASSWORD=true
ADMIN_PASSWORD=my-new-password   # or leave blank to auto-generate
```

Restart the backend, read the new password from the log, then **remove `RESET_ADMIN_PASSWORD` from your `.env`** so it does not reset again on the next boot.

---

### `could not write admin password file; set ADMIN_PASSWORD explicitly`

**Where you see it:** Backend container logs at startup (as a warning, not a fatal error).

```
level=WARN msg="could not write admin password file; set ADMIN_PASSWORD explicitly" error="..."
```

**Why it happens:** The backend tries to write the generated password to `/run/ipam/admin-password` (mode `0600`) inside the container. If the container's filesystem does not allow that write, the warning is emitted. The password is still printed to the log — the file is a convenience, not a requirement.

**Fix:** Read the password from the startup log. If you want the file, ensure the container has a writable `/run/ipam` directory or set `ADMIN_PASSWORD` in your `.env` to avoid auto-generation.

---

## Login errors

### `invalid username or password`

**Where you see it:** Login form.

**Why it happens:** The credentials do not match any active account.

**Fix:** Double-check the username (`admin` by default) and password. Use `RESET_ADMIN_PASSWORD=true` if you have lost the admin password (see above).

---

### `too many failed login attempts from this IP, please try again later`

**Where you see it:** Login form after repeated failures from the same IP address.

**Why it happens:** The backend rate-limits login attempts per source IP to slow brute-force attacks.

**Fix:** Wait for the lockout window to expire, then try again. If you are behind a NAT or shared proxy, other users on the same IP may trigger this limit.

---

### `account temporarily locked due to too many failed login attempts`

**Where you see it:** Login form after repeated failures for a specific account.

**Why it happens:** The backend tracks per-account failed login attempts. After the threshold is crossed, the account is locked until the lockout window expires.

**Fix:** Wait for the lock to expire (the lock duration is shown in the admin audit log). Admins can review the lockout from **Admin → Users** and, if necessary, unlock the account manually.

---

### `Account is suspended` / `Account rejected` / `Account disabled`

**Where you see it:** Login form (HTTP 403 response).

**Why it happens:** The account has been placed in a non-active state by an admin:

| State | Meaning |
|---|---|
| `suspended` | Temporarily blocked by an admin |
| `rejected` | A self-registration request was denied |
| `disabled` | Account has been administratively deactivated |

**Fix:** Contact your Padduck administrator to restore the account. Admins use **Admin → Users** to change account status.

---

### Login redirects or MFA prompt never appears

**Why it happens:** Session cookies are marked `Secure` and the browser discards them over plain HTTP. By default (`SESSION_COOKIE_SECURE=auto`), the backend sets `Secure` only when the request arrives over HTTPS or with an `X-Forwarded-Proto: https` header.

**Fix:**

- If you are accessing Padduck over plain HTTP in development, set `SESSION_COOKIE_SECURE=false` in your `.env`
- If you are behind a load balancer or reverse proxy that terminates TLS, ensure it forwards `X-Forwarded-Proto: https` and that the proxy's IP is listed in `TRUSTED_PROXIES`

---

## MFA

### MFA secrets lost after restart

**Where you see it:** Users with MFA enrolled receive an "invalid code" error after the backend restarts.

**Why it happens:** `MFA_ENCRYPTION_KEY` was not set (or was set to an empty string). In that case the backend generates a random per-process key at startup, so MFA secrets encrypted in the previous process cannot be decrypted after a restart.

**Fix:**

1. Generate a key: `openssl rand -hex 32`
2. Set `MFA_ENCRYPTION_KEY=<output>` in your `.env`
3. Restart the backend
4. Users whose MFA secrets were encrypted with the old key will need to re-enroll

> **Important:** Once you set `MFA_ENCRYPTION_KEY` in production, do not change it without a migration plan. Rotating the key invalidates all existing MFA enrollments.

---

## Scan agent

### Agent exits immediately: `IPAM_SERVER_URL and IPAM_AGENT_TOKEN must be set`

**Where you see it:** Agent container logs.

**Why it happens:** One or both required environment variables are missing.

**Fix:** Provide both variables when starting the agent:

```bash
docker run -d \
  -e IPAM_SERVER_URL=https://padduck.example.com \
  -e IPAM_AGENT_TOKEN=<token> \
  ...
```

---

## Health check failures

### Frontend or backend container stuck in `starting` / `unhealthy`

**Where you see it:** `docker compose ps` shows a service as unhealthy or the stack never finishes starting.

**Why it happens:** The Docker Compose healthchecks poll the `/health` endpoint. Common causes:

- The backend is waiting for a database migration that is taking longer than expected
- A port conflict prevents the service from binding (check if something else is on port `3000` or `8080`)
- The backend exited with a fatal error before the healthcheck passed (check `docker compose logs backend`)

**Fix:**

```bash
# Check all container logs at once:
docker compose logs --tail=50

# Check a specific service:
docker compose logs backend
```

Resolve the underlying error (usually database or migration related — see above), then restart with `docker compose up`.
