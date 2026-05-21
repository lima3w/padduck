# Product Vision

---

## Mission Statement

Padduck exists to make IP address management simple, auditable, and automation-friendly for infrastructure teams of any size — without requiring expensive commercial IPAM tools or maintaining fragile spreadsheets.

---

## Problems Padduck Solves

| Problem | How Padduck Fixes It |
|---------|----------------------|
| Spreadsheet drift | Single structured database with change history |
| No audit trail | Immutable audit log of every allocation and change |
| Manual allocation | API-driven allocation with idempotency and dry-run |
| Siloed DNS/DHCP data | Integrated zone and lease tracking |
| No access control | Fine-grained RBAC with role presets |
| Discovery blind spots | ICMP scan agents deployable anywhere on the network |
| Tool lock-in | Stable REST API, open-source, self-hosted |

---

## Target Users

**Network Engineers** — allocate and track IP space, manage VRFs/VLANs, run discovery.

**Infrastructure/DevOps Teams** — automate IP allocation in Terraform/Ansible pipelines, use the API to register devices and subnets.

**Security Teams** — audit all IP allocations, review who owns which address, enforce reservation policies.

**System Administrators** — manage users, roles, LDAP/SSO integration, system health.

---

## Why Another IPAM?

Most IPAM options fall into one of two traps:

1. **Over-engineered commercial tools** — expensive licensing, complex installation, features you'll never use.
2. **Spreadsheets and wikis** — no validation, no audit trail, break under team scale.

Padduck targets the middle: a self-hostable, open-source platform with enterprise-grade features (MFA, RBAC, audit logs, webhooks, stable API) but a deployment complexity of `docker compose up`.

See also: [Comparison With NetBox](Comparison-With-NetBox), [Why Padduck Exists](Why-Padduck-Exists)

---

## Competitive Landscape

| Tool | Type | Notes |
|------|------|-------|
| NetBox | Open-source IPAM/DCIM | Feature-rich but complex; Padduck is lighter and automation-focused |
| phpIPAM | Open-source IPAM | PHP-based; less API-first |
| Infoblox | Commercial | Enterprise pricing |
| SolarWinds IPAM | Commercial | Enterprise pricing |
| Spreadsheets | DIY | No validation, no audit |

---

## Long-Term Vision

- Full DNS/DHCP lifecycle management
- Multi-tenancy with strict data isolation
- Terraform provider and Ansible collection
- Cloud IP inventory sync (AWS, GCP, Azure VPCs)
- v2 API with GraphQL support

---

## Guiding Principles

1. **API-first** — the UI is a consumer of the same API available to automation
2. **Stable contracts** — breaking changes require a new API version
3. **Audit everything** — no silent mutations
4. **Secure defaults** — MFA, secure cookies, sensitive-value redaction out of the box
5. **Operator-friendly** — health checks, structured logs, graceful shutdown

---

## Non-Goals

- Padduck is **not** a network monitoring tool (no SNMP polling, no alerting on packet loss)
- Padduck is **not** a firewall manager (firewall zone tracking is metadata only)
- Padduck is **not** a full DCIM (device/rack/location is supporting context, not the primary focus)
- Padduck does **not** require cloud connectivity — it is fully self-hostable
