# Frequently Asked Questions

---

## General Questions

### What is Padduck?

Padduck is a self-hosted, open-source IP Address Management (IPAM) platform. It helps infrastructure teams organize, allocate, and audit IP address space through a web UI and REST API.

### Is Padduck free?

Yes. Padduck is open-source software licensed under **GPL-3.0**. You can use, modify, and distribute it freely under the terms of that license.

### What makes Padduck different from NetBox?

Padduck is lighter, more automation-focused, and easier to deploy. Key differences:
- Single `docker compose up` deployment vs. NetBox's more complex setup
- Stable, frozen v1 API contract — automation scripts don't break on upgrades
- Built-in idempotency keys for safe automation retries
- Stronger focus on the IPAM workflow rather than full DCIM

See [Comparison With NetBox](Comparison-With-NetBox) for a detailed comparison.

### Does Padduck work without internet access?

Yes. All Go dependencies are vendored. Docker images can be pre-pulled and stored in a private registry. No external services are required at runtime.

### Is Padduck production-ready?

Yes. The v1 API is frozen at OpenAPI 1.26.0, the frontend is at v0.3.0, and Padduck has been running in production environments. It includes MFA, RBAC, audit logging, health checks, and backup utilities.

---

## Installation Questions

### What are the minimum requirements?

- Docker 24.x + Docker Compose v2.20
- 512 MB RAM (2 GB recommended)
- 5 GB disk space

### Do I need to set MFA_ENCRYPTION_KEY?

**Yes, for production.** Without it, MFA secrets are encrypted with a random per-process key and lost on restart. Generate with `openssl rand -hex 32`.

### Can I change the default ports?

Yes. The frontend defaults to port 3000. Override with Docker Compose:
```yaml
services:
  frontend:
    ports:
      - "80:3000"
```

### Can I run behind a reverse proxy?

Yes. See [Installation Guide](Installation-Guide). Set `TRUSTED_PROXIES` to your proxy's IP and configure `X-Forwarded-Proto: https`.

### How do I reset the admin password?

```bash
# In .env:
RESET_ADMIN_PASSWORD=true
ADMIN_PASSWORD=new-password

# Restart backend, then remove RESET_ADMIN_PASSWORD from .env
docker compose restart backend
```

---

## Migration Questions

### Can I import data from NetBox?

Yes, via CSV import or the automation API. Export from NetBox, transform to Padduck's CSV format, and import via **Admin → Data Tools → Import**.

### Can I import from a spreadsheet?

Yes. Export your spreadsheet as CSV, format it to match the Padduck import schema, and use the bulk import endpoint.

### Does Padduck support IPv6?

Yes. IP addresses are stored as PostgreSQL INET type, which supports both IPv4 and IPv6. Some UI views are optimized for IPv4.

### What is the v2 migration bundle?

An export format for migrating from Padduck v1 to v2 (when v2 is released). Check readiness at **Admin → V2 Compatibility**.

---

## Performance Questions

### How many subnets/IPs can Padduck handle?

At typical production scale:
- 10,000+ subnets with sub-100ms list responses
- 1,000,000+ IP records (PostgreSQL handles this at scale)
- Pagination ensures list endpoints remain fast at any size

### How fast is IP allocation?

< 20ms typical for a single allocation (API-to-database round trip).

### Can I run multiple backend instances?

Yes — the backend is stateless. Put a load balancer in front and ensure all instances share the same `MFA_ENCRYPTION_KEY` and `TRUSTED_PROXIES` settings.

---

## Security Questions

### Are passwords stored in plaintext?

No. Passwords are hashed using bcrypt with a random salt. The hash is never returned via the API.

### Are API tokens stored in plaintext?

No. Tokens are stored as hashes. The plaintext token is shown only once at creation.

### Can I enforce MFA for all users?

Currently MFA is opt-in per user. Enforcement policies are on the roadmap.

### How are webhook signatures verified?

Padduck signs webhook payloads with HMAC-SHA256 using the endpoint secret. The signature is in the `X-IPAM-Signature-256` header.

```python
import hmac, hashlib
expected = 'sha256=' + hmac.new(secret.encode(), body, hashlib.sha256).hexdigest()
assert hmac.compare_digest(expected, header_value)
```

---

## Licensing Questions

### What license is Padduck under?

**GPL-3.0**. This means:
- You can use, modify, and distribute Padduck freely
- Modified versions must also be open-source under GPL-3.0
- You cannot use Padduck code in proprietary software without complying with GPL-3.0

### Can I use Padduck commercially?

Yes — GPL-3.0 allows commercial use. You can use Padduck internally at your company, or offer Padduck as a hosted service, as long as you comply with GPL-3.0 terms (source availability for modifications).

---

## Troubleshooting Questions

### The admin password wasn't printed to the log

The password banner only appears on first boot. Use `RESET_ADMIN_PASSWORD=true` to force it again.

### MFA codes are rejected after restart

`MFA_ENCRYPTION_KEY` is not set or changed. Set it to a stable 64-char hex value and restart.

### I see `SESSION_COOKIE_SECURE` issues

If accessing over plain HTTP: set `SESSION_COOKIE_SECURE=false` in `.env`. If behind HTTPS reverse proxy: ensure `X-Forwarded-Proto: https` is forwarded and your proxy IP is in `TRUSTED_PROXIES`.

See [Troubleshooting](Troubleshooting) for more.
