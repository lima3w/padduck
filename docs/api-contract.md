# Stable v1 API Contract

> **Documentation map**
> - **This file** — formal v1 compatibility rules and stability guarantees
> - **`docs/openapi.yaml`** — machine-readable OpenAPI 3.0.3 spec (current version `1.32.17`); also served live at `GET /api/openapi.yaml`
> - **[API Documentation wiki](https://github.com/lima3w/padduck/wiki/API-Documentation)** — human-readable quick reference, authentication guide, endpoint listing, and example requests

The v1 API contract was established at OpenAPI spec version `1.26.0` (`x-api-contract: stable-v1`). All changes since then have been additive only — no breaking changes without a new API version.

Compatibility rules:

- Existing v1 paths, methods, required request fields, response status codes, and response field names are treated as stable.
- Additive optional fields and new endpoints are allowed in v1.
- Breaking changes require a new API version or an explicit compatibility shim.
- Write-heavy automation endpoints support `Idempotency-Key` for retry-safe clients.
- Validation errors use `code: VALIDATION_ERROR` with a `fields` array.
- Outbound webhooks include `schema_version` in the payload and `X-IPAM-Event-Schema-Version` in delivery headers.
- Outbound webhooks and update checks reject URLs that resolve to private,
  loopback, link-local, multicast, or unspecified addresses, including
  redirects to those addresses.
- CSV exports escape spreadsheet formula prefixes in cell values.
- Sensitive values in audit payloads and admin configuration responses are
  redacted, including SNMP communities, passwords, API keys, tokens, and
  secrets.
