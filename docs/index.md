# Padduck

Padduck is a web-based **IP Address Management (IPAM)** platform. It gives network and infrastructure teams a single, structured place to track every IP address, subnet, VLAN, and VRF in their environment — replacing spreadsheets and tribal knowledge with a searchable, audited database backed by a stable REST API.

---

## Who it's for

| Role | How Padduck helps |
|---|---|
| **Network engineers** | Manage subnets, VLANs, VRFs, NAT rules, and DHCP leases in one UI |
| **Infrastructure / DevOps** | Automate IP allocation and DNS updates via the REST API or webhook events |
| **Security teams** | Review a full audit log of every change, enforce MFA and RBAC, and document firewall zones |
| **Admins** | Manage users, roles, scan agents, custom fields, and site configuration |

---

## Key capabilities

- **Hierarchical addressing** — organize your space as Sections → Subnets → IP Addresses
- **VRFs and VLANs** — model routing domains and layer-2 segments, link them to subnets
- **Network discovery** — schedule or trigger ICMP ping scans; a lightweight remote agent handles isolated segments
- **DNS integration** — sync and validate DNS records directly from Padduck
- **NAT, DHCP, and circuits** — document NAT rules, DHCP servers/leases, and physical/logical circuits (v1.29+)
- **Firewall zone mapping** — associate security zones with IPAM objects or CIDR ranges (v1.30+)
- **Stable REST API** — frozen v1 contract (OpenAPI `1.26.0`) with idempotency keys, token scopes, and generated client examples
- **Automation endpoints** — allocate, reserve, and release IPs; validate DNS updates; register devices; evaluate policies — all from CI/CD or scripts
- **Webhooks** — subscribe to any event with wildcard, object-type, tag, and field-condition filters; replay failed deliveries
- **RBAC and MFA** — role-based permissions, TOTP-based multi-factor authentication, and session controls
- **Full audit log** — every significant action is recorded; admins can filter and export

---

## Architecture

```
Browser
  └─▶  Frontend  (React / Vite, served by nginx on :3000)
          └─▶  Backend API  (Go / Fiber, internal :8080)
                    ├─▶  PostgreSQL  (data store)
                    └─▶  Scan Agent  (optional, Go binary — polls backend for jobs)
```

All three services ship as Docker images and are wired together by the included `docker-compose.yml`. The scan agent is a separate, stateless binary you deploy on any host that can reach the subnets you want to discover.

---

## Documentation

| Document | Contents |
|---|---|
| [Getting Started](getting-started.md) | Prerequisites, install, configuration, first login, first task |
| [User Guide](user-guide.md) | Full feature reference — sections, subnets, IPs, VRFs, VLANs, scanning, API tokens, webhooks, MFA, admin |
| [Troubleshooting](troubleshooting.md) | Startup errors, login problems, agent issues, common misconfigurations |
| [API Contract](api-contract.md) | Stability guarantees, idempotency, validation errors, v2 compatibility |
| [API Client Examples](api-client-examples.md) | JavaScript and Python snippets for common automation workflows |
| [OpenAPI spec](openapi.yaml) | Machine-readable API definition (also served live at `GET /api/openapi.yaml`) |
| [Internationalization (i18n)](i18n.md) | Locale file structure, how to add a new language, translation coverage |
