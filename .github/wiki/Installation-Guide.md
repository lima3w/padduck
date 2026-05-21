# Installation Guide

---

## Supported Platforms

Padduck runs anywhere Docker and Docker Compose are available:

- Linux (x86_64, ARM64) — **recommended for production**
- macOS (development only)
- Windows with WSL2 (development only)

---

## System Requirements

| Requirement | Minimum | Recommended |
|------------|---------|-------------|
| CPU | 1 core | 2+ cores |
| RAM | 512 MB | 2 GB |
| Disk | 5 GB | 20 GB |
| Docker Engine | 24.x | latest |
| Docker Compose | v2.20 | latest |

For local development outside Docker: **Go 1.26+** and **Node.js 20+**.

---

## Quick Install (Docker Compose)

```bash
# 1. Get the code
git clone https://github.com/lima3w/padduck.git
cd padduck

# 2. Create environment file
cp .env.example .env

# 3. Generate MFA encryption key (required for production)
openssl rand -hex 32
# Paste the output as MFA_ENCRYPTION_KEY in .env

# 4. Set a strong database password in .env

# 5. Start the stack
docker compose up --build
```

The first build takes a few minutes. When all three services are healthy:

```
padduck-frontend-1  | /docker-entrypoint.sh: Configuration complete; ready for start up
padduck-backend-1   | ========================================
padduck-backend-1   |   Admin password (first boot):  <generated-password>
padduck-backend-1   |   Set ADMIN_PASSWORD env var to override.
padduck-backend-1   | ========================================
```

Open **http://localhost:3000** and log in as `admin`.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | `padduck` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `padduck` | **Change before any shared deployment** |
| `POSTGRES_DB` | `padduck` | Database name |
| `DATABASE_URL` | *(derived)* | Full connection string; overrides the three variables above |
| `SERVER_PORT` | `8080` | Backend port inside container |
| `ENVIRONMENT` | `production` | `production` (JSON logs) or `development` (text logs) |
| `MFA_ENCRYPTION_KEY` | *(empty)* | **Required for production** — 64 hex characters |
| `ADMIN_PASSWORD` | *(auto-generated)* | Leave empty to auto-generate on first boot |
| `RESET_ADMIN_PASSWORD` | `false` | Set `true` to force-reset admin password on next boot |
| `TRUSTED_PROXIES` | *(none)* | Comma-separated IPs/CIDRs for X-Real-IP forwarding |
| `SCAN_MAX_CONCURRENT_JOBS` | `5` | Parallel scan job limit |
| `SESSION_COOKIE_SECURE` | `auto` | `true`, `false`, or `auto` (sets Secure flag only over HTTPS) |
| `APP_VERSION` | *(empty)* | Optional build metadata for admin update-check panel |
| `GIT_COMMIT` | *(empty)* | Optional build metadata |
| `BUILD_DATE` | *(empty)* | Optional build metadata |

---

## Docker Installation (Custom)

To use custom images or override the Compose file:

```yaml
# docker-compose.override.yml
services:
  backend:
    environment:
      - MFA_ENCRYPTION_KEY=your64hexcharshere
      - TRUSTED_PROXIES=10.0.0.0/8
  frontend:
    ports:
      - "80:3000"
```

---

## Reverse Proxy Setup

Place nginx or Caddy in front of Padduck for TLS termination.

### nginx example

```nginx
server {
    listen 443 ssl;
    server_name ipam.example.com;

    ssl_certificate     /etc/ssl/certs/ipam.crt;
    ssl_certificate_key /etc/ssl/private/ipam.key;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto https;
    }
}
```

Set `TRUSTED_PROXIES=127.0.0.1` and `SESSION_COOKIE_SECURE=auto` (or `true`).

---

## SSL/TLS Setup

- **Recommended**: Terminate TLS at nginx/Caddy (see above)
- Set `SESSION_COOKIE_SECURE=auto` — automatically marks session cookies Secure when HTTPS is detected
- Ensure `X-Forwarded-Proto: https` is passed from the proxy

---

## Initial Admin Setup

1. Log in as `admin` with the generated password
2. Navigate to **Admin → Settings** to configure SMTP, registration policy, and session timeouts
3. Create additional user accounts or configure LDAP/OAuth2/SAML
4. Change the admin password under **My Settings → Security**

---

## Verify the Install

```bash
# Backend health
curl -s http://localhost:8080/health
# Expected: {"status":"ok"}

# Frontend health (proxied through nginx)
curl -s http://localhost:3000/health
# Expected: {"status":"ok"}
```

---

## Deploying the Scan Agent

The scan agent is an optional Go binary for discovering live hosts in subnets the main server can't reach.

```bash
docker run -d \
  -e IPAM_SERVER_URL=https://padduck.example.com \
  -e IPAM_AGENT_TOKEN=<token-from-admin-panel> \
  -e POLL_INTERVAL=30 \
  ghcr.io/lima3w/padduck-agent:latest
```

Create the agent token under **Admin → Scan Agents → Create Agent**.

---

## Upgrade Process

```bash
git pull
docker compose pull   # if using pre-built images
docker compose pull
docker compose up -d
```

Database migrations run automatically on startup. Always read [Changelog](Changelog) before upgrading.

---

## Backup & Restore

```bash
# Backup (creates gzip SQL dump; prunes backups older than 30 days)
DATABASE_URL=postgres://padduck:pass@localhost/padduck ./tools/backup.sh ./backups

# Restore
DATABASE_URL=postgres://padduck:pass@localhost/padduck ./tools/restore.sh ./backups/padduck_backup_20260507_120000.sql.gz
```

---

## Kubernetes Installation

Kubernetes manifests are not yet included. For Kubernetes deployment:

1. Create a `Deployment` for the backend (stateless, scalable)
2. Create a `StatefulSet` or use an external PostgreSQL (e.g., CloudNative-PG)
3. Create a `Deployment` for the frontend (nginx)
4. Use `ConfigMap` and `Secret` for environment variables
5. Expose via `Ingress` with TLS

See the repository's README for scaling considerations.
