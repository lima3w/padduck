# Glossary

---

## Networking Terms

**CIDR (Classless Inter-Domain Routing)** — A method for allocating IP addresses using prefix notation (e.g., `10.0.0.0/24`). The number after `/` is the prefix length.

**Subnet** — A logical subdivision of an IP network. Defined by a network address and prefix length (e.g., `192.168.1.0/24`).

**Prefix** — Synonym for subnet in IPAM contexts. A CIDR block.

**Gateway** — The router IP address for a subnet (typically the first or last usable host address).

**Broadcast address** — The last IP in an IPv4 subnet, used for broadcast traffic. Not allocatable as a host address.

**Network address** — The first IP in a subnet (e.g., `10.0.0.0` in a /24). Not allocatable as a host address.

**Host range** — The allocatable IP addresses in a subnet (network address + 1 to broadcast - 1).

**VLAN (Virtual LAN)** — A logical Layer-2 network segment identified by a numeric ID (1–4094). Used to segregate network traffic without physical separation.

**VRF (Virtual Routing and Forwarding)** — A technology that allows multiple routing table instances to coexist on the same router. Enables overlapping IP ranges in different routing domains.

**BGP (Border Gateway Protocol)** — The routing protocol used between autonomous systems on the internet and in large networks.

**OSPF (Open Shortest Path First)** — An interior gateway routing protocol using link-state routing.

**NAT (Network Address Translation)** — Mapping of IP addresses, typically between private and public addresses.

**DHCP (Dynamic Host Configuration Protocol)** — Protocol for automatically assigning IP addresses to hosts on a network.

**DNS (Domain Name System)** — System for translating hostnames to IP addresses and vice versa.

**ICMP (Internet Control Message Protocol)** — Network protocol used for diagnostic messages; `ping` uses ICMP echo requests.

---

## IPAM Terms

**IPAM (IP Address Management)** — The discipline of planning, tracking, and managing IP address space.

**Allocation** — The process of assigning an IP address to a host or service.

**Reservation** — Holding an IP address without actively assigning it.

**Utilization** — The percentage of a subnet's host addresses that are assigned or reserved.

**Lease** — An IP assignment with an expiry date. When expired, the address can be released.

**Section** — In Padduck: a top-level organizational grouping for subnets (e.g., "Data Center", "Cloud VPC").

**Supernet** — A CIDR block that contains smaller subnets (e.g., `/16` is a supernet of multiple `/24`s).

**Overlap** — When two subnets share IP address space — usually a configuration error.

---

## Padduck-Specific Terms

**Section** — Top-level container for subnets in Padduck. Purely organizational (no network address).

**Scan Job** — A configured network scan that runs ICMP ping against target subnets to detect live hosts.

**Scan Agent** — A lightweight Go binary that polls the Padduck backend for scan assignments and reports results. Enables scanning subnets the main server can't reach.

**Scan Profile** — A reusable scan configuration (timing, method, retention) that can be applied to multiple scan jobs.

**Discovery Conflict** — A discrepancy between what's in IPAM and what's live on the network (e.g., a live host with no IPAM record).

**Break-Glass** — An emergency admin access mechanism that bypasses primary authentication when it's unavailable.

**Idempotency Key** — A client-supplied UUID that prevents duplicate writes on retry. If the same key is seen again, the original response is returned.

**ADR (Architecture Decision Record)** — A document capturing a significant architectural decision, its context, and consequences.

---

## Acronyms

| Acronym | Full Form |
|---------|-----------|
| IPAM | IP Address Management |
| CIDR | Classless Inter-Domain Routing |
| VLAN | Virtual Local Area Network |
| VRF | Virtual Routing and Forwarding |
| DHCP | Dynamic Host Configuration Protocol |
| DNS | Domain Name System |
| ICMP | Internet Control Message Protocol |
| NAT | Network Address Translation |
| BGP | Border Gateway Protocol |
| OSPF | Open Shortest Path First |
| RBAC | Role-Based Access Control |
| MFA | Multi-Factor Authentication |
| TOTP | Time-based One-Time Password |
| LDAP | Lightweight Directory Access Protocol |
| OIDC | OpenID Connect |
| SAML | Security Assertion Markup Language |
| SSRF | Server-Side Request Forgery |
| SBOM | Software Bill of Materials |
| ADR | Architecture Decision Record |
| API | Application Programming Interface |
| REST | Representational State Transfer |
| SPA | Single Page Application |
| HMAC | Hash-based Message Authentication Code |
| TLS | Transport Layer Security |
| GDPR | General Data Protection Regulation |

---

## Architecture Terminology

**Handler** — In Padduck's Go backend: an HTTP handler function that parses the request, calls a service, and returns a response.

**Service** — Business logic layer in the backend. Handlers call services; services call repositories.

**Repository** — Data access layer in the backend. Contains SQLC-generated database query functions.

**SQLC** — A tool that generates type-safe Go code from SQL queries. Used throughout the Padduck backend.

**Fiber** — The Go web framework Padduck uses (similar to Express.js for Node.js).

**pgx** — A high-performance PostgreSQL driver for Go.

**Vite** — The build tool and dev server for the Padduck frontend.

**Lazy loading** — Loading JavaScript code only when it's needed (used for all Padduck page components to minimize initial bundle size).
