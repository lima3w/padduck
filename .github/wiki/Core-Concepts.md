# Core Concepts

This page defines the fundamental building blocks of the Padduck data model. Understanding these concepts makes the rest of the system intuitive.

---

## What is IPAM?

**IP Address Management (IPAM)** is the discipline of planning, tracking, and managing the IP address space used by a network. At its core, IPAM answers: *what IP addresses exist, who owns them, and what are they for?*

---

## Hierarchy: Sections → Subnets → IP Addresses

Padduck organizes IP space in three levels:

```
Section
└── Subnet (CIDR block)
    └── IP Address
```

### Sections

A **Section** is a top-level organizational grouping. Think of it as a logical container for related subnets. Examples:

- `Data Center`
- `Cloud VPC - Production`
- `Office Networks`
- `Lab`

Sections have no network address — they are purely organizational.

### Subnets

A **Subnet** defines a CIDR block (e.g. `10.10.0.0/24`). It belongs to a Section and optionally to a VRF. Key attributes:

| Attribute | Description |
|-----------|-------------|
| Network address | e.g. `10.10.0.0` |
| Prefix length | e.g. `24` for /24 |
| VRF | Optional routing domain |
| VLAN | Optional Layer-2 association |
| Description | Free text |
| Tags | Organizational metadata |

Padduck calculates utilization automatically (assigned / total host addresses).

### IP Addresses

An **IP Address** lives inside a subnet and has one of three statuses:

| Status | Meaning |
|--------|---------|
| `available` | Not yet allocated |
| `assigned` | Actively in use by a host or service |
| `reserved` | Held but not actively in use |

Key attributes: `assigned_to` (hostname or label), `description`, `lease_expires_at`, `mac_address`, `tags`.

---

## VRFs (Virtual Routing & Forwarding)

A **VRF** is a virtual routing domain. Organizations use VRFs to isolate routing tables — for example, separating production, staging, and management traffic that might use overlapping address space.

In Padduck, subnets and VLANs can be associated with a VRF. The same CIDR block (`10.0.0.0/24`) can exist in multiple VRFs without conflict.

---

## VLANs

A **VLAN** (Virtual LAN) is a Layer-2 broadcast domain identified by a numeric ID (1–4094). VLANs can be:

- Grouped into **VLAN Domains** (e.g., a physical switch fabric)
- Associated with a VRF
- Linked to subnets for full L2/L3 documentation

---

## Prefixes

In IPAM terminology, a **prefix** is any CIDR block (/8, /16, /24, etc.). Padduck uses "subnet" for CIDR blocks, but the terms are interchangeable in most contexts.

---

## Reservations

A **reservation** is an IP address in `reserved` status — claimed to prevent accidental allocation, but not yet actively assigned. Common uses: reserved for infrastructure devices, future use, or administrative holds.

---

## DHCP Concepts

Padduck tracks **DHCP leases** as a record of what address was dynamically assigned to what MAC address and when. This is read-only metadata — Padduck does not act as a DHCP server.

---

## DNS Concepts

Padduck tracks **DNS zones** and **nameservers** as documentation. It can validate that DNS records align with IPAM data. Padduck does not act as an authoritative DNS server.

---

## Tenancy Model

Padduck supports optional **tenant** tagging for subnets and IP addresses, useful for organizations managing IP space on behalf of multiple customers or business units. Tenancy is metadata — access control is managed through RBAC roles, not per-tenant isolation.

---

## Permissions Model

See [[Security]] for the full RBAC model. In brief:

- Every user has a **role** (admin, operator, viewer, or custom)
- Roles define what actions are permitted (read, write, admin)
- Permissions are checked on every API request

---

## Address Lifecycle

```
available → assigned (allocate)
available → reserved (reserve)
reserved  → assigned (allocate)
assigned  → available (release)
assigned  → reserved (reserve)
```

Leased IPs have an `expires_at` date. When expired, they can be released with "Release Expired".

---

## Tags & Metadata

**Tags** are free-form key/value labels that can be applied to subnets, IP addresses, VRFs, VLANs, devices, and more. Use tags for organizational metadata like environment (`env=prod`), owner (`team=platform`), or cost center.

**Custom Fields** extend the data model with admin-defined attributes specific to your organization.

---

## Audit Logging

Every significant create, update, and delete action generates an **audit log entry** recording:

- The user who performed the action
- The source IP address
- The timestamp
- The object type and ID
- The action performed
- A diff of changed values (sensitive values are redacted)

Audit logs are immutable and exportable.

---

## Discovery

**Discovery** (network scanning) uses ICMP ping to detect live hosts on subnets. Results are compared against IPAM records to surface conflicts (e.g., an address responding that has no IPAM record). Remote **scan agents** extend discovery to network segments the main server can't reach directly.
