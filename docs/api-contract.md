# Stable v1 API Contract

The v1 API contract is frozen at OpenAPI version `1.26.0` with `x-api-contract: stable-v1`.

Compatibility rules:

- Existing v1 paths, methods, required request fields, response status codes, and response field names are treated as stable.
- Additive optional fields and new endpoints are allowed in v1.
- Breaking changes require a new API version or an explicit compatibility shim.
- Write-heavy automation endpoints support `Idempotency-Key` for retry-safe clients.
- Validation errors use `code: VALIDATION_ERROR` with a `fields` array.
- Outbound webhooks include `schema_version` in the payload and `X-IPAM-Event-Schema-Version` in delivery headers.

## V2 Compatibility Warnings

Administrators can review known v2 compatibility warnings at:

- `GET /api/v1/admin/compatibility/v2-warnings`
- `GET /api/v1/admin/export/v2-migration-bundle`

The response groups warnings by API, field, and workflow areas and includes
recommended remediation work for v1 clients before a v2 upgrade. Clients should
prefer top-level IP address endpoints, send idempotency keys for automation
writes, and avoid depending solely on legacy role fields.

The v2 migration bundle endpoint returns an `application/zip` archive containing
`manifest.json`, `data/ipam-v1-export.json`, `data/ipam-v1-export.csv`, and a
short README. The JSON export is the canonical migration input; the CSV export is
included for inspection and fallback workflows.
