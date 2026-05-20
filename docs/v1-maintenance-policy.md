# V1 Long-Term Maintenance Policy

IPAM Next v1 remains the stable production line while v2 is developed. The v1
line is maintained for operators who need predictable APIs, migrations, and
deployment behavior during the v2 transition.

## Support Scope

- Security fixes for supported v1 releases.
- Data-loss and correctness fixes for IPAM, DNS, authentication, RBAC, audit,
  reporting, import, export, and migration tooling.
- Backward-compatible API additions needed for operations or migration.
- Compatibility warnings, readiness checks, and migration bundle export updates
  required for a supported v2 upgrade path.
- Documentation corrections for supported v1 behavior.

## Compatibility Rules

- Existing v1 API paths, methods, required request fields, response field names,
  and documented status codes remain stable unless a security issue requires a
  change.
- Additive optional fields, new endpoints, and new warnings are allowed.
- Database migrations must be forward-safe and preserve existing data.
- Paired SQL migration files must keep up/down sections separate; `.up.sql`
  files may not contain `-- +migrate Down` blocks and `.down.sql` files may not
  contain `-- +migrate Up` blocks.
- Configuration defaults must remain safe for existing deployments.
- Deprecated v1 surfaces remain available until v2 provides a documented
  migration path.

## Deprecation Process

Deprecations are reported through:

- Admin UI: **Admin** -> **V2 Compatibility**
- API: `GET /api/v1/admin/compatibility/deprecations`
- API: `GET /api/v1/admin/compatibility/v2-warnings`

Each deprecation entry includes the v1 surface, expected v2 change, impacted
APIs, fields, or workflows, and recommended remediation. New deprecations should
be documented before the related v2 migration requirement is enforced.

## Migration Readiness

Administrators should run readiness checks before creating a v2 migration
bundle:

- Admin UI: **Admin** -> **V2 Compatibility**
- API: `GET /api/v1/admin/compatibility/v2-readiness`

Readiness checks cover schema, runtime configuration, integrations, custom
fields, roles, API tokens, and webhook subscriptions. `fail` blocks migration
readiness. `warn` indicates work that should be reviewed and either corrected or
accepted before export.

## Release And Patch Expectations

- Patch releases use backward-compatible migrations and do not require manual
  data cleanup unless the release notes say otherwise.
- Security fixes may disable unsafe behavior when no compatible mitigation is
  available.
- Operators should keep v1 within the latest supported minor release before
  attempting a v2 migration bundle export.
- Release notes must call out new compatibility warnings, readiness checks, and
  migration bundle format changes.

## End Of Maintenance

V1 maintenance continues until a v2 migration path is generally available and
the project publishes an explicit end-of-maintenance date. That date should be
announced with enough lead time for operators to export, test, and complete a v2
migration.
