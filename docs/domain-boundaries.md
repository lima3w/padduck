# Domain Module Boundaries

This document defines how backend business logic is organized into domain modules and how those modules are injected into the HTTP handler layer.

## Injection pattern

The `Handler` struct receives named domain managers at construction time. Each manager groups sub-services for one functional domain:

```go
type Handler struct {
    service *services.Service   // residual — shrinks each patch
    ops     *services.OpsManager
    auth    *services.AuthManager
}
```

As domains are extracted the residual `service` field loses methods. When it is empty it will be removed. Handlers call `h.ops.X.Method()` or `h.auth.X.Method()` — never directly into the repository.

## Domains

### Operations (`h.ops` — `OpsManager`)

| Sub-service field | Responsibility |
|---|---|
| `Discovery` | Scan jobs, scan profiles, scan agents, scan history, scan retention, discovery conflicts, fingerprints |
| `Reports` | Utilization trends, inactive IP reports, scheduled reports, reconciliation, duplicates, exports |
| `Import` | CSV/JSON/phpIPAM imports |
| `Jobs` | Background job tracking and cancellation |
| `Webhooks` | Webhook endpoint CRUD, delivery history, replay |
| `Topology` | Topology hints, network topology views |
| `DNS` | PowerDNS/Technitium integration, DNS zone browser |
| `Automation` | Automation policies, idempotent write endpoints (allocate/reserve/release/register/dns-update) |
| `Telemetry` | Opt-in usage telemetry |
| `NetworkModules` | NAT rules, firewall zones/mappings, DHCP servers/leases, circuit providers/circuits (physical and logical), BGP autonomous systems |
| `IPAM` | Networks, subnets, IP addresses, VRFs, VLANs, VLAN domains/groups, tags, global and scoped search, dashboard summary/activity, IPv6 delegations, subnet split/merge/resize |
| `Identity` | Users, RBAC roles/permissions, API tokens, web sessions, password management, account security (lockout/unlock), Grafana datasource proxy |
| `Infrastructure` | Devices (SNMP creds, interfaces, IP associations), racks, locations (tree), nameservers |
| `Customers` | Customer CRUD, customer associations |

### Identity & Auth (`h.auth` — `AuthManager`)

| Sub-service field | Responsibility |
|---|---|
| `Email` | Transactional email delivery |
| `Registration` | New account registration and email verification |
| `MFA` | TOTP setup, confirmation, backup codes |
| `Notification` | In-app and email notification preferences |
| `LDAP` | LDAP directory integration |
| `OAuth2` | OAuth2/OIDC provider integration |
| `SAML` | SAML2 identity provider integration |

### Residual on `*Service` (to be extracted in subsequent patches)

These method groups remain on the root `Service` struct and will be extracted domain by domain. Each extraction becomes its own patch release.

| Planned domain | Service files | Approx. methods |
|---|---|---|
| **Workflow** | requests, custom\_fields | ~33 |

## Extraction rules

1. A sub-service has a `repo *repository.Repository` field. Where needed, narrow cross-domain deps (e.g. `*ConfigService`, `*DNSService`) are injected at constructor time rather than taking a full `*Service` reference. The goal is no circular or monolithic back-references.
2. Methods that were on `*Service` are moved verbatim; only the receiver type changes.
3. Package-level helpers (e.g. `defaultString`, `validateCIDRLike`) live in the file that first defines them and are accessible to all files in the `services` package.
4. Integration tests for a sub-service construct it directly (`NewXxxService(repo)`) rather than going through `NewService`.
5. Handler files are updated atomically with the service extraction so the build never breaks mid-PR.
