# Padduck

**Padduck** is a modern, production-grade IP Address Management (IPAM) platform built with Go, React, and PostgreSQL. It replaces spreadsheet-based IP tracking with a structured, auditable, API-first system.

[![CI](https://github.com/lima3w/padduck/actions/workflows/ci.yml/badge.svg)](https://github.com/lima3w/padduck/actions)

---

## What is Padduck?

Padduck gives infrastructure teams a single source of truth for IP address space. You organize your network into **Sections → Subnets → IP Addresses**, manage VRFs and VLANs, run discovery scans, track devices, and integrate with DNS — all through a clean web UI and a stable REST API.

---

## Core Features

| Feature | Description |
|---------|-------------|
| **Hierarchical IPAM** | Sections → Subnets → IP Addresses with utilization tracking |
| **VRF & VLAN Management** | Full virtual routing and switching domain support |
| **Network Discovery** | ICMP ping scans with remote scan agents |
| **DNS Integration** | Zone management and record tracking |
| **RBAC** | Fine-grained role-based access control |
| **MFA** | TOTP two-factor authentication |
| **External Auth** | LDAP, OAuth2, SAML2 |
| **Stable REST API** | OpenAPI 1.26.0, contract-frozen v1 |
| **Webhooks** | Outbound event subscriptions with replay |
| **Audit Logging** | Immutable record of all changes |
| **Automation** | Idempotent allocation, reservation, and policy evaluation |
| **Reports** | Utilization trends, inactive IPs, duplicate detection |

---

## Design Philosophy

- **API-first** — every UI action is backed by a documented REST endpoint
- **Automation-friendly** — idempotency keys, dry-run mode, policy evaluation
- **Audit by default** — all writes are logged with user, IP, and timestamp
- **Secure by default** — MFA, RBAC, session management, sensitive-value redaction
- **Self-hostable** — single `docker compose up` deployment, no cloud dependencies

---

## Quick Links

| Resource | Link |
|----------|------|
| Repository | [github.com/lima3w/padduck](https://github.com/lima3w/padduck) |
| OpenAPI Spec | `GET /api/openapi.yaml` on your instance |
| Installation | [Installation Guide](Installation-Guide) |
| User Guide | [User Guide](User-Guide) |
| API Docs | [API Documentation](API-Documentation) |
| FAQ | [FAQ](FAQ) |
| Troubleshooting | [Troubleshooting](Troubleshooting) |

---

## Current Status

- **API Contract**: Stable v1 (frozen at OpenAPI 1.26.0)
- **Frontend**: v0.3.0
- **Deployment**: Docker Compose (production-ready)
- **License**: GPL-3.0

---

## Getting Started

```bash
git clone https://github.com/lima3w/padduck.git
cd padduck
cp .env.example .env
# Edit .env — set POSTGRES_PASSWORD and MFA_ENCRYPTION_KEY
openssl rand -hex 32   # paste output as MFA_ENCRYPTION_KEY
docker compose pull
docker compose up -d
```

Open `http://localhost:3000` and log in as `admin`. The generated password is printed to the backend log on first boot.

See the [Installation Guide](Installation-Guide) for the full walkthrough.

---

## Contributing

See [CONTRIBUTING.md](https://github.com/lima3w/padduck/src/branch/main/README.md) in the repository for local setup, coding standards, and PR workflow.

---

## Community & Support

- **Issues**: [GitHub Issues](https://github.com/lima3w/padduck/issues)
- **License**: [GPL-3.0](https://github.com/lima3w/padduck/src/branch/main/LICENSE)
