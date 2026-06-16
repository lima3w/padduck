# API Documentation

---

## API Overview

Padduck provides a **stable REST API** under `/api/v1/`. The API is the foundation for all UI interactions and external automation.

- **Base URL**: `https://your-padduck-instance/api/v1`
- **OpenAPI Spec**: `GET /api/openapi.yaml` (OpenAPI 3.0.3, version 1.31.32)
- **Contract stability**: v1 is frozen — no breaking changes without a new version
- **Total endpoints**: 199 documented paths (275 operations) in the OpenAPI spec, plus internal/admin routes not yet covered there

---

## Authentication

### Session Cookie (Browser)

Used by the web UI. Log in via `POST /auth/login`, then session cookie is set automatically.

### Bearer Token (API clients)

```bash
curl -H "Authorization: Bearer <token>" https://padduck.example.com/api/v1/subnets
```

Generate tokens under **My Settings → API Tokens** or via admin.

### Token Scopes

| Scope | Allowed Operations |
|-------|--------------------|
| `read` | GET requests only |
| `write` | GET + POST + PUT + DELETE for IPAM resources |
| `admin` | Full access including admin endpoints |

---

## API Versioning

- Current stable version: **v1** (frozen at OpenAPI 1.26.0)
- Additive changes (new optional fields, new endpoints) are allowed within v1
- Breaking changes require a new API version
- Version compatibility: `GET /api/v1/admin/compatibility/v2-warnings`

---

## Error Handling

All errors return JSON:

```json
{
  "error": "validation_error",
  "code": "VALIDATION_ERROR",
  "fields": [
    {"field": "prefix_length", "message": "must be between 1 and 32"}
  ]
}
```

| Status Code | Meaning |
|------------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad request / validation error |
| `401` | Not authenticated |
| `403` | Not authorized |
| `404` | Resource not found |
| `409` | Conflict (duplicate, idempotency) |
| `422` | Unprocessable entity |
| `500` | Internal server error |

---

## Pagination

List endpoints support cursor or offset pagination:

```
GET /api/v1/subnets?page=2&limit=50
```

Response includes `total`, `page`, `limit`, and `items`.

---

## Idempotency Keys

Automation write endpoints accept `Idempotency-Key` for safe retry:

```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $(uuidgen)" \
  -H "Content-Type: application/json" \
  -d '{"subnet_id": 42, "assigned_to": "web-01"}' \
  https://padduck.example.com/api/v1/automation/ip-addresses/allocate
```

If the same idempotency key is sent again, the original response is returned without re-executing.

---

## Rate Limiting

The API includes a token-bucket rate limiter. When throttled, you receive `429 Too Many Requests`. Token usage and rate limit status are visible in **Admin → Integrations → API Token Analytics**.

---

## Key Endpoint Groups

### Authentication
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/login` | Log in (username + password) |
| `POST` | `/auth/logout` | Log out |
| `POST` | `/auth/mfa/verify` | Verify TOTP code |
| `POST` | `/auth/password-reset` | Request password reset |
| `GET` | `/auth/me` | Current user info |

### Networks
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/networks` | List networks |
| `POST` | `/api/v1/networks` | Create network |
| `GET` | `/api/v1/networks/:id` | Get network |
| `PUT` | `/api/v1/networks/:id` | Update network |
| `DELETE` | `/api/v1/networks/:id` | Delete network |

### Subnets
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/subnets` | List subnets |
| `POST` | `/api/v1/subnets` | Create subnet |
| `GET` | `/api/v1/subnets/:id` | Get subnet |
| `PUT` | `/api/v1/subnets/:id` | Update subnet |
| `DELETE` | `/api/v1/subnets/:id` | Delete subnet |
| `GET` | `/api/v1/networks/:id/subnets` | List subnets in a network |

### IP Addresses
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/subnets/:subnetID/ip-addresses` | List IPs in subnet |
| `POST` | `/api/v1/subnets/:subnetID/ip-addresses` | Create IP record in subnet |
| `POST` | `/api/v1/ip-addresses/quick-create` | Create IP record by address (subnet inferred) |
| `GET` | `/api/v1/ip-addresses/:id` | Get IP |
| `PUT` | `/api/v1/ip-addresses/:id` | Update IP |
| `DELETE` | `/api/v1/ip-addresses/:id` | Delete IP |
| `GET` | `/api/v1/ip-addresses/search` | Search IPs globally |

### Automation (Idempotent)
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/automation/ip-addresses/allocate` | Allocate next available IP |
| `POST` | `/api/v1/automation/ip-addresses/reserve` | Reserve specific IP |
| `POST` | `/api/v1/automation/ip-addresses/:id/release` | Release IP |
| `POST` | `/api/v1/automation/dns/update` | Validate DNS update |
| `POST` | `/api/v1/automation/devices/register` | Register device |
| `POST` | `/api/v1/automation/policies/evaluate` | Evaluate automation policy |

All automation write endpoints accept `dry_run: true`.

### VRFs & VLANs
| Method | Path | Description |
|--------|------|-------------|
| `GET/POST` | `/api/v1/vrfs` | List/create VRFs |
| `GET/PUT/DELETE` | `/api/v1/vrfs/:id` | Get/update/delete VRF |
| `GET/POST` | `/api/v1/vlans` | List/create VLANs |
| `GET/POST` | `/api/v1/vlan-domains` | List/create VLAN domains |

### Discovery
| Method | Path | Description |
|--------|------|-------------|
| `GET/POST` | `/api/v1/scan-jobs` | List/create scan jobs |
| `POST` | `/api/v1/scan-jobs/:id/run` | Run scan job immediately |
| `GET/POST` | `/api/v1/scan-profiles` | List/create scan profiles |
| `GET/POST` | `/api/v1/scan-agents` | List/create scan agents |

### Admin
| Method | Path | Description |
|--------|------|-------------|
| `GET/POST` | `/api/v1/admin/users` | List/create users |
| `GET/POST` | `/api/v1/admin/roles` | List/create roles |
| `GET` | `/api/v1/admin/audit-log` | View audit log |
| `GET/PATCH` | `/api/v1/admin/config` | System configuration |
| `GET/POST` | `/api/v1/admin/webhooks` | List/create webhooks |
| `GET/POST` | `/api/v1/admin/features` | Feature flags |

---

## Webhooks

Configure outbound webhooks under **Admin → Webhooks**. Each delivery includes:

- `X-IPAM-Event-Schema-Version` header
- `X-IPAM-Signature-256` header (HMAC-SHA256 of body using endpoint secret)
- `schema_version` in payload body

Webhook URLs pointing to private/loopback/link-local addresses are rejected (SSRF protection).

---

## Example Requests

### Allocate next IP (JavaScript)

```javascript
const res = await fetch('/api/v1/automation/ip-addresses/allocate', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
    'Idempotency-Key': crypto.randomUUID(),
  },
  body: JSON.stringify({ subnet_id: 42, assigned_to: 'web-01' }),
});
const ip = await res.json();
console.log(ip.address); // e.g. "10.10.0.5"
```

### Reserve specific IP (Python)

```python
import requests, uuid

response = requests.post(
    'https://padduck.example.com/api/v1/automation/ip-addresses/reserve',
    headers={
        'Authorization': f'Bearer {token}',
        'Idempotency-Key': str(uuid.uuid4()),
    },
    json={'subnet_id': 10, 'address': '10.0.0.25', 'hostname': 'printer-25'},
    timeout=10,
)
response.raise_for_status()
print(response.json())
```

---

## V2 Compatibility

Before planning a v2 migration:

```bash
# Check migration readiness
GET /api/v1/admin/compatibility/v2-readiness

# See deprecation warnings
GET /api/v1/admin/compatibility/deprecations

# Export migration bundle
GET /api/v1/admin/export/v2-migration-bundle
```

See also the repository's `/docs/` directory for API design decisions.
