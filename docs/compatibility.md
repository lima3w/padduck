# API Compatibility Policy

This document defines what padduck guarantees to operators and API consumers, what counts as a breaking change, and how deprecations are managed.

---

## What is v1?

The v1 API is the `/api/v1` namespace. It includes all endpoints, request/response field names, pagination conventions, auth token behavior, and webhook payload shapes documented under that prefix. Behavior not described in the documentation is not guaranteed.

---

## Support Window

| Phase | Guarantee |
|-------|-----------|
| **Active** (current major version) | Bug fixes and security patches. No intentional breaking changes. New optional fields may be added. |
| **Maintenance** (one major version after retirement) | Security patches only. No new features or bug fixes. |
| **End of life** | Routes removed. Returns `410 Gone`. |

v1 is currently **active**. It will enter maintenance when v2.0.0 is declared stable and will reach end of life no earlier than v3.0.0.

Changes to the support window will be announced in the CHANGELOG with at least two release cycles of notice.

---

## Breaking Change Definition

The following are **breaking changes** and will only occur at major version boundaries (v1.x.y → v2.0.0):

- Removing an endpoint.
- Removing or renaming a JSON field in a response.
- Changing a field's type (e.g., string → integer).
- Changing a field from optional to required in a request body.
- Changing authentication or authorization requirements for an endpoint.
- Changing the meaning of an existing error code.
- Removing a query parameter that was previously accepted.

---

## Non-Breaking Additions

The following changes may occur in any patch or minor release and are **not** considered breaking:

- Adding a new endpoint.
- Adding a new optional field to a response.
- Adding a new optional query parameter.
- Adding a new optional field to a request body.
- Adding new values to an enumerated field that was previously open-ended.
- Adding new response headers (including `Deprecation` and `Link`).
- Changing prose in error messages (only the `error` key is stable, not the `message`).
- Performance improvements that do not change observable behavior.

---

## Deprecation Process

1. The endpoint or field is marked deprecated in the CHANGELOG under `### Deprecated`.
2. Deprecated endpoints begin returning `Deprecation: true` and `Link: <successor>; rel="successor-version"` response headers on every response.
3. The deprecated item remains fully functional for a minimum of **two release cycles** (patch versions do not count; minor or major releases count).
4. Removal occurs only at a major version boundary and is listed in the Breaking Change Registry below.

---

## Breaking Change Registry

All known or planned breaking changes between v1 and v2 are listed here. This table is the authoritative record; the [migration guide](migration-v1-to-v2.md) provides implementation details for each entry.

| Change | Affects | Target version | Migration |
|--------|---------|----------------|-----------|
| `colour` → `color` in IP tag responses | `GET /api/v1/tags`, webhook payloads | v2.0.0 | Rename field in consumer |
| Bare array responses → `{ "data": [...], "meta": {...} }` envelope | All paginated list endpoints | v2.0.0 | Unwrap `.data`; read pagination from `.meta` |
| v1 endpoints removed after sunset | All `/api/v1/*` routes | v3.0.0 (earliest) | Migrate to `/api/v2/*` equivalents |

Fields marked *(planned)* in the migration guide will be added to this registry as they are confirmed.

---

## Endpoint Stability Status

| Endpoint | Status | v2 equivalent | Notes |
|----------|--------|---------------|-------|
| `GET /api/v1/networks` | Deprecated | `GET /api/v2/networks` | Emits `Deprecation: true` header |
| `POST /api/v1/networks` | Stable | `POST /api/v2/networks` *(planned)* | — |
| `GET /api/v1/networks/:id` | Stable | `GET /api/v2/networks/:id` *(planned)* | — |
| `PUT /api/v1/networks/:id` | Stable | `PUT /api/v2/networks/:id` *(planned)* | — |
| `DELETE /api/v1/networks/:id` | Stable | `DELETE /api/v2/networks/:id` *(planned)* | — |
| All other `/api/v1/*` endpoints | Stable | *(v2 equivalents planned)* | No deprecation headers yet |

---

## Operator Upgrade Checklist

Use this checklist to assess your exposure before upgrading padduck to a new major version.

### API consumers (scripts, automation, CI)

- [ ] **JSON field names**: scan your code for the field names listed in the Breaking Change Registry. If you parse `colour`, rename to `color` before upgrading to v2.
- [ ] **Response shape**: if you access response arrays directly (e.g., `response.data[0]`), check whether the endpoint now returns `{ "data": [...] }`. Wrap access with `.data[0]`.
- [ ] **Pagination**: verify your pagination loop reads `meta.total` and `meta.page` from the v2 envelope rather than the v1 top-level keys.
- [ ] **Error fields**: if your code reads `error.message`, note that message text is not stable. Use `error.error` (the machine-readable key) for branching logic.
- [ ] **Deprecation headers**: add logging or alerting for responses that include `Deprecation: true` so you know which integrations still call v1 endpoints.

### Webhook consumers

- [ ] **Payload field names**: check webhook handler code for any field listed in the Breaking Change Registry.
- [ ] **Event type strings**: verify that the event type values your consumers filter on are still emitted. Consult the migration guide for renames.
- [ ] **Signature verification**: if you verify webhook HMAC signatures, confirm the signing algorithm has not changed (it has not as of v1.33.x).

### Infrastructure / operators

- [ ] **`--migrate-dry-run`**: run `./padduck --migrate-dry-run` before each major-version upgrade to preview schema changes without applying them.
- [ ] **`V1_COMPAT_SUNSET`**: set this env var to your planned migration deadline to receive startup warnings as a reminder.
- [ ] **Health checks**: confirm your load balancer or orchestrator health check URL (`/api/v1/health`) is still valid after upgrade.
- [ ] **Reverse proxies**: if you proxy `/api/v1` to a different backend than `/api/v2`, update your routing rules before upgrading.
