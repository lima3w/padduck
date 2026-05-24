# Padduck Scan Agent — Privilege Model

## What the Agent Does

The Padduck scan agent is a lightweight binary that:

1. Polls the Padduck server for assigned scan jobs (via HTTP).
2. Enumerates all host IPs in one or more CIDR subnets.
3. Pings each IP to test liveness.
4. Posts the results (alive/dead, response time) back to the Padduck server.

## How Ping Is Implemented

The agent uses the system's external `ping` binary via `exec.Command("ping", ...)`.
It does **not** open raw sockets directly. Specifically:

```go
cmd := exec.Command("ping", "-c", "1", "-W", <timeout_seconds>, host)
err := cmd.Run()
```

It passes a single-count ping (`-c 1`) with a configurable timeout (`-W`).

## Required OS-Level Permissions

Because the agent delegates ICMP to the system `ping` binary rather than
opening raw sockets itself, **the agent process does not need elevated privileges**.

| Component | Privilege needed |
|---|---|
| Agent process itself | Normal user — no root, no `setcap` required |
| System `ping` binary | Already privileged on all major platforms (see below) |
| HTTP polling to server | Normal outbound TCP — no special permissions |

### Linux

On modern Linux distributions the system `ping` binary (usually
`/bin/ping` or `/usr/bin/ping`) already has the `cap_net_raw` capability
set or is installed setuid-root by the package manager. No action is needed
for the agent.

You can verify:

```bash
# Look for setuid bit
ls -l $(which ping)

# Or look for cap_net_raw capability
getcap $(which ping)
# Expected: /usr/bin/ping cap_net_raw+ep
```

### macOS

The macOS `ping` binary (`/sbin/ping`) is installed with the setuid-root
bit set by the OS. No additional configuration is required.

### Windows

On Windows, ICMP echo requests are allowed for all users by default via
the Windows Sockets API that the system `ping.exe` uses. No elevated
privileges are needed.

## Recommended Setup — Running as a Dedicated Non-Root User

Even though root is not required, it is good practice to run the agent as a
dedicated, low-privilege service account:

```bash
# Create a dedicated user (no home dir, no login shell)
sudo useradd --system --no-create-home --shell /usr/sbin/nologin padduck-agent

# Place the binary somewhere the user can execute it
sudo install -m 755 -o root -g root padduck-agent /usr/local/bin/padduck-agent
```

The agent binary itself needs no special capabilities or setuid bit.

## Running as a systemd Service

Create `/etc/systemd/system/padduck-agent.service`:

```ini
[Unit]
Description=Padduck Scan Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=padduck-agent
Group=padduck-agent

# Required environment variables
Environment=PADDUCK_SERVER_URL=https://padduck.example.com
Environment=PADDUCK_AGENT_TOKEN=<your-agent-token>
# Optional: override poll interval (default 30 seconds)
# Environment=POLL_INTERVAL=60

ExecStart=/usr/local/bin/padduck-agent

Restart=on-failure
RestartSec=15s

# Harden the service
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now padduck-agent
sudo systemctl status padduck-agent
```

View logs:

```bash
journalctl -u padduck-agent -f
```

## Summary

| Question | Answer |
|---|---|
| Needs root? | No |
| Needs `setcap`? | No (agent delegates to system `ping`) |
| Needs raw socket access? | No |
| Safe to run as a non-root service account? | Yes — recommended |
| Uses external `ping` binary? | Yes (`exec.Command("ping", ...)`) |
| Systemd service supported? | Yes |
