# Getting Started

This guide walks you from zero to your first allocated IP address. It covers the Docker Compose deployment path, which is the fastest way to get a working instance.

---

## Prerequisites

| Requirement | Minimum version | Notes |
|---|---|---|
| [Docker](https://docs.docker.com/get-docker/) | 24.x | Engine + CLI |
| [Docker Compose](https://docs.docker.com/compose/install/) | v2.20 | Ships with Docker Desktop; standalone install on Linux |
| Outbound internet access | — | Required to download the Compose file and pull images |

> **Local development only** — if you want to run the backend or frontend outside Docker, you also need Go 1.22+ and Node.js 20+. Those paths are not covered here.

---

## Install

### 1. Create a deployment directory

```bash
mkdir padduck
cd padduck
```

You do not need to clone the repository or build images locally. The backend
and frontend images are published to GitHub Container Registry.

### 2. Download the Compose file

```bash
curl -fsSLO https://raw.githubusercontent.com/lima3w/padduck/main/docker-compose.yml
```

**`POSTGRES_PASSWORD` is required** — the stack will not start without it. Copy
the example file and set a strong value before running `docker compose up`:

```bash
curl -fsSL https://raw.githubusercontent.com/lima3w/padduck/main/.env.example -o .env
# Edit .env and set POSTGRES_PASSWORD to a strong value, e.g.:
#   openssl rand -base64 32
```

See the [Configuration](#configuration) section below for all available variables.

### 3. Start the stack

```bash
docker compose pull
docker compose up -d
```

SCREENSHOT: Terminal output of `docker compose up -d` completing successfully, showing all three services (db, backend, frontend) marked as healthy.

The first image pull can take a few minutes. On first startup, the backend
creates a persistent MFA encryption key in `./data/backend/mfa-encryption-key`
if `MFA_ENCRYPTION_KEY` is not set. When all three services are healthy you will
see the backend print the admin password (if you left `ADMIN_PASSWORD` blank):

```
========================================
  Admin password (first boot):  <generated-password>
  Set ADMIN_PASSWORD env var to override.
========================================
```

> **Tip:** The generated password is also written to `/run/ipam/admin-password` inside the backend container (mode `0600`). You can read it with:
> ```bash
> docker compose exec backend cat /run/ipam/admin-password
> ```

---

## Configuration

All settings are read from environment variables. Docker Compose interpolates them from your `.env` file automatically.

| Variable | Default | Description |
|---|---|---|
| `POSTGRES_USER` | `padduck` | PostgreSQL user created at first boot |
| `POSTGRES_PASSWORD` | **required** | Must be set in `.env` before first run; the stack will not start without it |
| `POSTGRES_DB` | `padduck` | Database name |
| `DATABASE_URL` | *(derived)* | Full connection string; overrides the three variables above when set |
| `SERVER_PORT` | `8080` | Port the backend listens on inside the container |
| `ENVIRONMENT` | `production` | Controls the log format: `production` emits JSON logs, anything else emits text logs |
| `LOG_LEVEL` | `warn` | Log verbosity: `warn` (warnings and errors only), `info` (adds operational logging such as scan job lifecycle), `debug` (full verbosity), or `error`. Unknown values fall back to `warn` with a startup notice |
| `MFA_ENCRYPTION_KEY` | generated if unset | Optional override for the backend-managed persistent key. Must be 64 hex characters. Generate with `openssl rand -hex 32` |
| `ADMIN_PASSWORD` | *(auto-generated)* | Leave empty to auto-generate on first boot; set to use a specific password |
| `RESET_ADMIN_PASSWORD` | `false` | Set to `true` to force-reset the admin password on next boot, then remove the variable |
| `TRUSTED_PROXIES` | *(none)* | Comma-separated IPs/CIDRs to trust for `X-Real-IP` forwarding (e.g. your load balancer) |
| `SCAN_MAX_CONCURRENT_JOBS` | `5` | Maximum number of scan jobs that run at the same time |
| `SESSION_COOKIE_SECURE` | `auto` | `true` forces the Secure flag on session cookies; `auto` sets it only when the request arrives over HTTPS |
| `APP_VERSION` / `GIT_COMMIT` / `BUILD_DATE` | *(empty)* | Optional build metadata shown in the admin update-check panel |

---

## Verify the install

Once the stack is up, confirm the backend and frontend are healthy:

```bash
# Backend health from inside the Compose network
docker compose exec backend wget -qO- http://127.0.0.1:8080/health
# Expected: {"status":"ok"}

# Frontend health
curl -s http://localhost:3000/health
# Expected: {"status":"ok"}
```

Then open `http://localhost:3000` in your browser and log in with username `admin` and the password from the startup log.

SCREENSHOT: The Padduck login screen showing the username and password fields.

---

## Production: TLS and HSTS

By default the frontend binds only to `127.0.0.1` (loopback), so plain HTTP
from the local machine works for development but the port is not reachable from
the network. This is intentional: in production, Padduck **must** be placed
behind a TLS-terminating reverse proxy (nginx, Caddy, Traefik, etc.). Sessions
are cookie-based; without TLS, credentials and session cookies cross the
network in cleartext.

### Architecture

```
Internet → reverse proxy (TLS) → 127.0.0.1:3000 (frontend container)
                                       ↓ (internal Docker network)
                                  backend:8080
```

The reverse proxy terminates HTTPS and forwards to `127.0.0.1:3000` on the
same host. Session cookies are automatically marked Secure when the request
arrives over HTTPS (`SESSION_COOKIE_SECURE=auto`).

If you need the frontend reachable on all interfaces (not recommended for
production), set `FRONTEND_BIND=0.0.0.0` in `.env`.

### HSTS

Set the `Strict-Transport-Security` (HSTS) header **at the layer that terminates TLS** — it must not be set on a plain-HTTP backend response. Examples:

**nginx**

```nginx
server {
    listen 443 ssl;
    server_name padduck.example.com;
    # ... ssl_certificate / ssl_certificate_key ...

    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains" always;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Caddy** (sets HSTS automatically when it serves HTTPS, but to be explicit):

```caddyfile
padduck.example.com {
    header Strict-Transport-Security "max-age=63072000; includeSubDomains"
    reverse_proxy 127.0.0.1:3000
}
```

**Traefik** (dynamic configuration):

```yaml
http:
  middlewares:
    hsts:
      headers:
        stsSeconds: 63072000
        stsIncludeSubdomains: true
```

Also set `TRUSTED_PROXIES` to your proxy's address so the backend sees real client IPs, and consider `SESSION_COOKIE_SECURE=true` once TLS is in place (the default `auto` detects HTTPS via `X-Forwarded-Proto`).

---

## First real task

Once you're logged in, here is the happy path for recording your first subnet and allocating an IP.

### 1. Create a Section

Sections are top-level groupings (e.g. "Data Center", "Cloud VPC", "Office").

1. In the sidebar, click **Sections**
2. Click **+ New Section**
3. Enter a name (e.g. `Lab`) and click **Save**

### 2. Add a Subnet

1. Open your new Section
2. Click **+ New Subnet**
3. Fill in the network address (e.g. `10.10.0.0`) and prefix length (e.g. `24`)
4. Click **Save**

Padduck immediately calculates how many host addresses are in the block and shows utilization as `0 / 254`.

### 3. Allocate the first IP

1. Open the subnet you just created
2. Click **Allocate**
3. Enter an **Assigned to** value — a hostname, service name, or any label that identifies what will use this address (e.g. `web-01`)
4. Click **Allocate** to confirm

Padduck assigns the lowest available host address and marks it `assigned`.

SCREENSHOT: The subnet detail page showing one assigned IP with its "Assigned to" label and utilization updated to `1 / 254`.

---

## Next steps

- **Add more subnets and structure** — see [User Guide → Sections and Subnets](user-guide.md#sections)
- **Set up network discovery** — see [User Guide → Discovery](user-guide.md#discovery-network-scanning)
- **Automate IP allocation** — see [API Client Examples](api-client-examples.md)
- **Deploy the scan agent** — see below

### Deploying the scan agent

The scan agent is an optional Go binary that polls the backend for assigned scan jobs and posts results back. Deploy it on any host that can reach subnets the main backend cannot.

Configure it with three environment variables:

| Variable | Description |
|---|---|
| `IPAM_SERVER_URL` | Base URL of your Padduck instance (e.g. `https://padduck.example.com`) |
| `IPAM_AGENT_TOKEN` | Bearer token created under **Admin → Scan Agents** |
| `POLL_INTERVAL` | Polling interval in seconds (default: `30`) |

Download the agent binary for your platform from the GitHub Releases page, then
run it with those environment variables set.
