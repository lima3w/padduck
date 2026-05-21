# Troubleshooting

---

## Startup Errors

### `Failed to connect to database`

**Symptom**: Backend exits immediately with a database connection error.

**Causes & fixes**:
1. `db` container not yet healthy: `docker compose ps` — wait for healthy status
2. Wrong credentials: verify `POSTGRES_PASSWORD` in `.env` matches what `db` was initialized with
3. Changed password after first boot: `docker compose down -v && docker compose up --build` (destroys data — backup first!)
4. Custom `DATABASE_URL` pointing to wrong host/port

```bash
# Verify DB is reachable
docker compose exec db pg_isready -U padduck
```

---

### `Failed to run migrations`

**Symptom**: Backend starts but fails during migration phase.

**Causes & fixes**:
1. Check the error for which migration file failed
2. Verify the DB user has `CREATE TABLE`, `ALTER TABLE`, `CREATE INDEX` privileges
3. If previous run crashed mid-migration: inspect `schema_migrations` table for a dirty entry and clean it manually

```sql
-- Find dirty migrations
SELECT * FROM schema_migrations WHERE dirty = true;

-- Clear dirty flag (use exact version from error)
DELETE FROM schema_migrations WHERE version = 'NNNN_failed_migration';
```

---

### `Failed to initialize admin password`

**Symptom**: Startup fails after migration step.

**Fix**: Resolve database errors first. If DB is clean, check if the `admin` user row exists in the `users` table.

---

## Admin Password

### Password banner not shown on startup

**Cause**: The banner only prints on first boot (when password has never been set).

**Fix**: Force reset:
```bash
# .env:
RESET_ADMIN_PASSWORD=true
ADMIN_PASSWORD=my-new-password   # or leave blank to auto-generate

# Restart:
docker compose restart backend

# After: remove RESET_ADMIN_PASSWORD from .env
```

---

### `could not write admin password file`

**Cause**: Backend can't write to `/run/ipam/admin-password` inside the container.

**Fix**: Password is still printed to the log — read it there. Set `ADMIN_PASSWORD` in `.env` to avoid auto-generation.

---

## Login Errors

### `invalid username or password`

Double-check the username (`admin` by default) and password. Use `RESET_ADMIN_PASSWORD=true` if the password is lost.

---

### `too many failed login attempts from this IP`

Rate-limit triggered. Wait for the lockout window to expire. If behind NAT/shared proxy, other users on the same IP may have triggered it.

---

### `account temporarily locked due to too many failed login attempts`

Account-level lockout. Wait for the window to expire, or ask an admin to unlock the account via **Admin → Users**.

---

### `Account is suspended / rejected / disabled`

Contact your Padduck administrator. Admins can change account state via **Admin → Users**.

---

### Login redirects loop or MFA prompt never appears

**Cause**: Session cookies have `Secure` flag but you're accessing over plain HTTP.

**Fix**:
- For plain HTTP development: `SESSION_COOKIE_SECURE=false` in `.env`
- Behind HTTPS reverse proxy: ensure proxy sends `X-Forwarded-Proto: https` and proxy IP is in `TRUSTED_PROXIES`

---

## MFA Issues

### MFA codes rejected after backend restart

**Cause**: `MFA_ENCRYPTION_KEY` not set — random per-process key used; secrets lost on restart.

**Fix**:
1. `openssl rand -hex 32` → set as `MFA_ENCRYPTION_KEY` in `.env`
2. Restart backend
3. Users must re-enroll MFA (old secrets are unreadable with new key)

> **Important**: Never change `MFA_ENCRYPTION_KEY` in production without a re-enrollment plan.

---

## Scan Agent Issues

### Agent exits: `IPAM_SERVER_URL and IPAM_AGENT_TOKEN must be set`

Both required env vars must be provided:
```bash
docker run -d \
  -e IPAM_SERVER_URL=https://padduck.example.com \
  -e IPAM_AGENT_TOKEN=<token> \
  padduck/agent:latest
```

Create the token under **Admin → Scan Agents → Create Agent**.

---

## Health Check Failures

### Container stuck in `starting` or `unhealthy`

```bash
# Check status
docker compose ps

# Check backend logs
docker compose logs --tail=100 backend

# Check for port conflicts
ss -tlnp | grep -E '3000|8080'
```

Common causes:
- Backend waiting for a slow migration
- Port conflict (something else on 3000 or 8080)
- Backend exited with fatal error before healthcheck passed

---

## Database Issues

### Database disk full

```bash
# Check disk usage
docker system df

# Check PostgreSQL data volume
docker compose exec db df -h /var/lib/postgresql/data
```

Fix: expand disk, then `docker compose restart db`.

### Migration table corrupted

```bash
docker compose exec db psql -U padduck padduck -c "SELECT * FROM schema_migrations;"
```

If a migration is stuck dirty, manually clean it and restart the backend.

---

## Performance Issues

### API responses slow (> 500ms)

1. Check database connection: `docker compose exec db pg_isready`
2. Check for missing indexes: look at slow query log
3. Check `SCAN_MAX_CONCURRENT_JOBS` — reduce if scans are competing with API requests
4. Check disk I/O: `iostat -x 1 5` on the host

### Frontend slow to load

1. Verify nginx is serving static assets (not proxying to backend)
2. Check `docker compose logs frontend` for errors
3. Clear browser cache — stale JS can cause issues after upgrades

---

## Upgrade Failures

### Backend fails to start after upgrade

1. Check [[Changelog]] for breaking changes and migration notes
2. Check `docker compose logs backend` for migration errors
3. Restore from backup if needed

```bash
# Restore database from backup
docker compose stop backend
DATABASE_URL=... ./tools/restore.sh ./backups/padduck_backup_YYYYMMDD_HHMMSS.sql.gz
docker compose up backend
```

---

## Debugging Techniques

### Enable debug logging

```bash
# .env:
ENVIRONMENT=development

docker compose restart backend
```

Development mode uses text logs (more readable for debugging).

### Access database directly

```bash
docker compose exec db psql -U padduck padduck

# Useful queries:
SELECT username, status, failed_logins, locked_until FROM users;
SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 5;
SELECT COUNT(*) FROM audit_log;
```

### Check audit log

```bash
# Most recent 20 audit entries
docker compose exec db psql -U padduck padduck -c \
  "SELECT created_at, username, action, object_type, status FROM audit_log ORDER BY created_at DESC LIMIT 20;"
```
