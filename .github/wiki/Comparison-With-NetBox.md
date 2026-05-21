# Comparison With NetBox

Both Padduck and NetBox are open-source, self-hosted tools that include IPAM functionality. This page explains how they differ to help you choose the right tool.

---

## Summary

| Dimension | Padduck | NetBox |
|-----------|---------|--------|
| **Primary focus** | IPAM + automation | Full DCIM + IPAM |
| **Deployment** | `docker compose up` (3 services) | More complex (Python/Django + Celery + Redis) |
| **Backend** | Go (fast, single binary) | Python/Django |
| **API stability** | Frozen v1 contract | Evolving |
| **Automation focus** | Idempotency keys, dry-run, policies | Strong but less specialized |
| **License** | GPL-3.0 | Apache-2.0 |
| **DCIM features** | Minimal (devices, locations, racks as context) | Full DCIM (circuits, power, console, etc.) |
| **Age/maturity** | Newer, actively developed | Mature, large community |
| **Community size** | Smaller, focused | Large, established |

---

## When to Choose Padduck

**Choose Padduck if you:**

- Primarily need IPAM and network discovery, not full DCIM
- Value simple deployment (`docker compose up`) over feature breadth
- Write automation against the API and need guaranteed stability
- Want idempotency keys and dry-run support built in
- Prefer Go's operational simplicity over Python's ecosystem
- Are starting fresh and want to grow into complexity

---

## When to Choose NetBox

**Choose NetBox if you:**

- Need full DCIM (power distribution, console servers, physical cabling, circuit management)
- Already have a NetBox investment and team expertise
- Need the large NetBox plugin ecosystem
- Have a large community supporting your deployment

---

## Feature Comparison

### Core IPAM

| Feature | Padduck | NetBox |
|---------|---------|--------|
| Subnets (prefixes) | ✅ | ✅ |
| IP addresses | ✅ | ✅ |
| VRFs | ✅ | ✅ |
| VLANs | ✅ | ✅ |
| IP utilization | ✅ | ✅ |
| IPv6 | ✅ | ✅ |
| Hierarchical prefix nesting | Sections → Subnets | Regions → Site → Prefixes |
| DHCP tracking | ✅ | ✅ |
| DNS zone tracking | ✅ | ✅ |

### DCIM

| Feature | Padduck | NetBox |
|---------|---------|--------|
| Devices | Lightweight (context) | Full (with device types, interfaces) |
| Locations/Racks | Basic | Comprehensive |
| Power distribution | ❌ | ✅ |
| Console servers | ❌ | ✅ |
| Physical cabling | ❌ | ✅ |
| Circuits | Basic tracking | Full circuit management |

### Automation & API

| Feature | Padduck | NetBox |
|---------|---------|--------|
| REST API | ✅ Frozen v1 | ✅ Evolving |
| Idempotency keys | ✅ Built-in | Partial |
| Dry-run mode | ✅ Built-in | Via plugins |
| Webhooks | ✅ With replay | ✅ |
| Automation policies | ✅ Built-in | Via scripts |

### Operations

| Feature | Padduck | NetBox |
|---------|---------|--------|
| Network discovery | ✅ ICMP + remote agents | Via plugins |
| Audit logging | ✅ Built-in | ✅ Built-in |
| MFA | ✅ TOTP | ✅ TOTP |
| LDAP | ✅ | ✅ |
| OAuth2/SAML | ✅ | ✅ |
| RBAC | ✅ Custom roles | ✅ |
| Reports | ✅ Built-in | Via plugins |

---

## Migrating from NetBox to Padduck

1. Export prefixes/IPs/VRFs/VLANs from NetBox
2. Transform to Padduck CSV import format
3. Import via **Admin → Data Tools → Import**
4. Reconfigure automation scripts to use Padduck's API
5. Set up webhooks to replace any NetBox event-driven automation

See [Integrations](Integrations) for details.
