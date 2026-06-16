# Why Padduck Exists

---

## The Problem

Every infrastructure team eventually reaches the same inflection point: the IP address spreadsheet stops working.

It starts innocuously. Someone creates a Google Sheet or Excel file. Columns for IP, hostname, purpose. Works great for 50 addresses. Then 500. Then:

- Two people edit simultaneously and overwrite each other
- Someone allocates an IP that's already in use (it was in column C, row 847)
- A VM gets decommissioned but the IP is never released — "reserved" forever
- No one knows why `10.0.1.50` is reserved — the person who added it left two years ago
- Compliance asks for an audit of all IP allocations last quarter — impossible

**The spreadsheet becomes a liability instead of an asset.**

---

## Why Not Existing Tools?

The obvious alternatives have their own problems:

### Commercial IPAM (Infoblox, SolarWinds)

- Enterprise pricing ($10K–$100K+/year)
- Complex deployment requirements
- Features you'll never use, vendor lock-in you definitely will

### NetBox

A great open-source option, but:
- Heavier deployment (Python/Django stack, more dependencies)
- Full DCIM focus means significant complexity for teams that just need IPAM
- API stability has historically been a concern for automation scripts
- Self-hosting complexity is higher than it needs to be

### phpIPAM and similar

- Older codebases
- Less API-first design
- Less suited for automation-heavy workflows

---

## What Padduck Does Differently

### Deployment is a single command

```bash
docker compose up --build
```

That's it. One PostgreSQL container, one Go backend, one nginx + React frontend. No Python virtual environments, no Celery workers, no Redis instances, no complex initial setup.

### The API contract is a promise

Infrastructure teams write automation against APIs. When the API changes without warning, automation breaks. Padduck's v1 API is **frozen** — no breaking changes, period. If a breaking change is needed, it waits for v2.

### Automation is a first-class concern

Every write operation that automation might retry has **idempotency key support**. Allocate the same IP twice with the same key — you get the same response, no duplicate. Add `dry_run: true` to validate a policy decision without committing.

### The audit trail is non-negotiable

Every allocation, reservation, release, user change, and configuration update is logged with who, when, from where, and what changed. Not as an afterthought — as a core data model requirement.

### Data stays yours

No cloud dependencies. No license server. No mandatory "phone home." Your IP data is yours, on your infrastructure, under your control. The one exception is a fully optional, off-by-default telemetry snapshot an admin can turn on — see [Data Ownership Philosophy](Data-Ownership-Philosophy) for exactly what it sends.

---

## The Name

A padduck is an old Scottish word for a frog. Frogs are survivors — they've been around for 360 million years and adapt to almost any environment. They're also a bit unusual, which felt right for an IPAM tool that takes a different approach.

See also the repository's README

---

## Who Padduck Is For

Infrastructure teams of 5–500 people who:
- Have outgrown spreadsheets
- Can't justify or don't want enterprise IPAM pricing
- Want to automate IP allocation from Terraform, Ansible, or scripts
- Need an audit trail for compliance or incident response
- Prefer self-hosted, open-source software they control
