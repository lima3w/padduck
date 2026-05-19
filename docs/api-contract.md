# Stable v1 API Contract

The v1 API contract is frozen at OpenAPI version `1.26.0` with `x-api-contract: stable-v1`.

Compatibility rules:

- Existing v1 paths, methods, required request fields, response status codes, and response field names are treated as stable.
- Additive optional fields and new endpoints are allowed in v1.
- Breaking changes require a new API version or an explicit compatibility shim.
- Write-heavy automation endpoints support `Idempotency-Key` for retry-safe clients.
- Validation errors use `code: VALIDATION_ERROR` with a `fields` array.
- Outbound webhooks include `schema_version` in the payload and `X-IPAM-Event-Schema-Version` in delivery headers.
