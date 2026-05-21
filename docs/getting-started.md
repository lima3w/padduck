# Getting Started

This guide walks you from zero to your first allocated IP address. It covers the Docker Compose deployment path, which is the fastest way to get a working instance.

---

## Prerequisites

| Requirement | Minimum version | Notes |
|---|---|---|
| [Docker](https://docs.docker.com/get-docker/) | 24.x | Engine + CLI |
| [Docker Compose](https://docs.docker.com/compose/install/) | v2.20 | Ships with Docker Desktop; standalone install on Linux |
| Outbound internet access | — | Required to pull images on first build |

> **Local development only** — if you want to run the backend or frontend outside Docker, you also need Go 1.22+ and Node.js 20+. Those paths are not covered here.

---

## Install

### 1. Get the code

```bash
git clone https://gitea.lima3.dev/Lima3-Automations/padduck.git
cd padduck
```

### 2. Create your environment file

Copy the example file and open it in your editor:

```bash
cp .env.example .env
```

At minimum, set a strong database password and an MFA encryption key before continuing:

```bash
# Generate the MFA key (requires openssl):
openssl rand -hex 32
```

Paste the output as the value of `MFA_ENCRYPTION_KEY` in your `.env` file. See the [Configuration](#configuration) section below for all available variables.

### 3. Start the stack

```bash
docker compose up --build
```

SCREENSHOT: Terminal output of `docker compose up --build` completing successfully, showing all three services (db, backend, frontend) marked as healthy.

The first build takes a few minutes. When all three services are healthy you will see the backend print the admin password (if you left `ADMIN_PASSWORD` blank):

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
| `POSTGRES_USER` | `ipam` | PostgreSQL user created at first boot |
| `POSTGRES_PASSWORD` | `ipam` | **Change before any shared deployment** |
| `POSTGRES_DB` | `ipam` | Database name |
| `DATABASE_URL` | *(derived)* | Full connection string; overrides the three variables above when set |
| `SERVER_PORT` | `8080` | Port the backend listens on inside the container |
| `ENVIRONMENT` | `production` | Set to `development` for debug-level text logs; `production` emits JSON logs |
| `MFA_ENCRYPTION_KEY` | *(empty)* | **Required for persistent MFA.** 64 hex characters. Generate with `openssl rand -hex 32` |
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
# Backend health
curl -s http://localhost:8080/health
# Expected: {"status":"ok"}

# Frontend health
curl -s http://localhost:3000/health
# Expected: {"status":"ok"}
```

Then open `http://localhost:3000` in your browser and log in with username `admin` and the password from the startup log.

SCREENSHOT: The Padduck login screen showing the username and password fields.

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

Run the agent container:

```bash
docker run -d \
  -e IPAM_SERVER_URL=https://padduck.example.com \
  -e IPAM_AGENT_TOKEN=<your-token> \
  -e POLL_INTERVAL=30 \
  gitea.lima3.dev/Lima3-Automations/padduck/agent:latest
```

<!-- TODO: verify the agent image registry path -->
