# v1 → v2 Migration Guide

> For the stability guarantee and deprecation policy that governs this migration, see [compatibility.md](compatibility.md).

This document covers every breaking change between the padduck v1 and v2 APIs, and explains how to update automation, webhooks, and direct API consumers before the v1 sunset date.

---

## Timeline

| Date | Event |
|------|-------|
| v1.33.6 | `/api/v2` introduced; `GET /api/v2/networks` is the first v2 endpoint |
| v2.0.0 (planned) | All core endpoints available on `/api/v2` |
| `V1_COMPAT_SUNSET` (operator-configured) | v1 routes retired |

v1 routes remain registered and functional throughout the v2 build-out. Once you set the `V1_COMPAT_SUNSET` environment variable, the server logs a warning on every startup. After the sunset date, v1 routes will return `410 Gone`.

---

## Detecting Deprecated Endpoints

Every v1 endpoint that has a v2 equivalent returns two headers:

```
Deprecation: true
Link: </api/v2/networks>; rel="successor-version"
```

These headers appear on **all** responses from the deprecated endpoint, including 401 and 403, so consumers can detect them regardless of auth status. Per [RFC 8594](https://www.rfc-editor.org/rfc/rfc8594).

---

## Response Shape Changes

### Paginated lists

v1 endpoints return bare arrays by default, with an optional paginated envelope when `?page=` or `?limit=` query parameters are supplied:

```json
[ { "id": 1, "name": "example" }, ... ]
```

v2 endpoints **always** return the standard envelope:

```json
{
  "data": [ { "id": 1, "name": "example" }, ... ],
  "meta": {
    "page": 1,
    "limit": 25,
    "total": 142
  }
}
```

Update pagination loop code from:

```js
// v1
const networks = await api.get('/api/v1/networks?page=1&limit=25')
const items = networks.data.data   // envelope only when params present
const total = networks.data.total
```

To:

```js
// v2 — envelope is always present
const networks = await apiV2.get('/api/v2/networks?page=1&limit=25')
const items = networks.data.data
const total = networks.data.meta.total
```

### Single-item responses

v2 single-item responses wrap the object in a `data` key:

```json
{ "data": { "id": 1, "name": "example" } }
```

v1 returns the object directly:

```json
{ "id": 1, "name": "example" }
```

---

## Field Renames (British → American spelling)

The following JSON field names change from British to American English spelling in v2. v1 continues to return the British spellings until sunset.

| Resource | v1 field | v2 field |
|----------|----------|----------|
| `tags` | `colour` | `color` |

---

## Endpoint Changes

### Networks (Sections)

| v1 | v2 | Notes |
|----|----|----|
| `GET /api/v1/networks` | `GET /api/v2/networks` | v2 always paginates; bare-array response removed |
| `POST /api/v1/networks` | `POST /api/v2/networks` | *(planned)* |
| `GET /api/v1/networks/:id` | `GET /api/v2/networks/:id` | *(planned)* |
| `PUT /api/v1/networks/:id` | `PUT /api/v2/networks/:id` | *(planned)* |
| `DELETE /api/v1/networks/:id` | `DELETE /api/v2/networks/:id` | *(planned)* |

### Pagination Parameters

v1 uses `?page=` and `?limit=`. v2 uses the same parameter names — no change required.

---

## Authentication

API token authentication and session cookie authentication are unchanged between v1 and v2. The same token used with `/api/v1` works with `/api/v2`.

---

## Webhook Payload Changes

Webhook payloads are generated from audit log entries. The event type format changes:

| v1 event type | v2 event type |
|---------------|---------------|
| `section.section_created` | `network.created` *(planned)* |

Field renames within payloads mirror the API field renames listed above.

### Before (v1 webhook payload for a tag event):

```json
{
  "event_type": "tag.created",
  "resource_type": "tag",
  "new_values": {
    "colour": "#3B82F6"
  }
}
```

### After (v2 webhook payload):

```json
{
  "event_type": "tag.created",
  "resource_type": "tag",
  "new_values": {
    "color": "#3B82F6"
  }
}
```

Update your webhook consumers to accept either spelling during the transition period, then switch to the v2 field names before the sunset date.

---

## Automation Script Updates

If you use the padduck API in scripts or automation, update base URLs:

```bash
# v1 (deprecated)
curl -H "Authorization: Bearer $TOKEN" https://ipam.example.com/api/v1/networks

# v2
curl -H "Authorization: Bearer $TOKEN" https://ipam.example.com/api/v2/networks
```

---

## Migration Dry-Run

Before applying pending database migrations, you can preview what will run without making any changes:

```bash
./padduck --migrate-dry-run
```

This connects to the database, prints each pending migration ID and its SQL, then exits without applying anything. Exit code is always 0.

---

## Operator Configuration

| Env var | Type | Description |
|---------|------|-------------|
| `V1_COMPAT_SUNSET` | ISO 8601 date (`YYYY-MM-DD`) | When set, the server logs a startup warning reminding operators to migrate consumers before this date. v1 routes remain active until explicitly disabled in a future release. |

Example:

```bash
V1_COMPAT_SUNSET=2027-01-01 ./padduck
```

Startup log output:

```
level=WARN msg="v1 API compat mode: v1 routes will be retired after the configured sunset date" v1_compat_sunset=2027-01-01 action="migrate consumers to /api/v2 before this date"
```
