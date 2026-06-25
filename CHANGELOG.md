# Changelog

## v1.33.20

### Fixed
- **Compose port env vars ignored in healthchecks**: `SERVER_PORT` and `FRONTEND_INTERNAL_PORT` now flow through to both the port mapping and the Docker healthcheck test commands. Previously both were hardcoded (`8080`/`3000`), causing containers to be marked unhealthy when either port was changed via env var.

## v1.33.19

### Fixed
- **Backend startup on fresh install**: `db.Connect` now retries the initial database ping up to 10 times with exponential backoff (1 s → 16 s cap, ~2.5 min total) instead of failing immediately. Eliminates the race condition where the backend container exits before PostgreSQL finishes initializing on a cold Docker install.

## v1.33.18

### Added
- **Drift review workflow** (issue #15): after every scan job, `DiscoveryService` compares each observed state to the authoritative `ip_addresses` record and generates a `drift_item` for any diverged fields (`hostname` via PTR record, `mac_address` via SNMP).
- **`drift_items` table** (migration `20260623_001_drift_items.up.sql`): one open item per resource at a time, enforced by a partial unique index on `(resource_type, resource_id) WHERE status = 'open'`. Re-scans update the existing open item rather than creating duplicates.
- **Three resolution actions** — each closes the drift item and audit-logs the decision:
  - `POST /api/v1/admin/drift/:id/accept` — writes all observed field values back to the authoritative record, then marks accepted
  - `POST /api/v1/admin/drift/:id/dismiss` — marks expected; no authoritative change
  - `POST /api/v1/admin/drift/:id/escalate` — creates a `discovery_conflicts` record per diverged field for investigation, then marks escalated
- **List and detail endpoints**: `GET /api/v1/admin/drift` (filterable by `?status=`; defaults to `open`) and `GET /api/v1/admin/drift/:id`; both are org-scoped when the caller has an org context.

## v1.33.17

### Added
- **Observed state tracking** (issue #14): new `observed_states` table stores a rolling per-resource snapshot of what the scanner last saw, completely separate from the authoritative `ip_addresses` and `devices` tables.
- **Authoritative data is never modified by scan results** without explicit user action — observed data lives solely in `observed_states`.
- **Scan integration**: `DiscoveryService.ScanSubnet` upserts an `observed_states` row for every scanned IP after each ping/port-scan/SNMP pass. Captured fields include `is_alive`, `ptr_record`, `response_time_ms`, `fwd_rev_mismatch`, `open_ports` (when port scan is enabled), and `snmp_hostname`/`snmp_mac_address` (when SNMP is enabled).
- **Unregistered host tracking**: IPs seen by the scanner but not matched to any `ip_addresses` record are stored with `resource_id = NULL`; a dedicated partial unique index keeps them deduplicated per IP address.
- **API endpoints**:
  - `GET /api/v1/admin/discovery/observed?resource_type=ip_address&resource_id=<id>` — fetch the latest observed snapshot for a specific registered resource
  - `GET /api/v1/admin/discovery/unregistered` — list IPs seen by scanner but absent from `ip_addresses`; org-scoped when caller has an org context

## v1.33.16

### Added
- **Tenant quotas, retention, and integration settings** (issue #12): new `organization_settings` table (migration `20260621_001_org_settings.up.sql`) stores per-org resource limits and SMTP overrides.
- **`OrgSettingsService`**: `CheckQuota` enforces `max_users`, `max_webhooks`, and `max_api_tokens`; returns `QuotaExceededError` (HTTP 422) when a limit would be breached. Subnet/IP quota enforcement is deferred until those tables have `organization_id` columns.
- **Quota enforcement** in handlers: `POST /users`, `POST /admin/webhooks`, and `POST /auth/me/tokens` all call `CheckQuota` before creating the resource.
- **Per-org SMTP overrides**: `organization_settings.smtp_host`, `smtp_port`, and `smtp_from` stored per org; accessible via `OrgSettingsService.GetOrgSMTPOverride`.
- **Per-org audit retention**: `organization_settings.audit_retention_days` overrides the global retention; `POST /admin/audit/prune` now applies per-org retention after the global prune pass.
- **Org settings API**:
  - `GET /api/v1/admin/organization/settings` — org admins view their own quota and SMTP settings
  - `PUT /api/v1/admin/organization/settings` — org admins may update SMTP override fields (quota fields are preserved from platform-admin values)
  - `GET /api/v1/platform/organizations/:id/settings` — platform admins view any org's settings
  - `PUT /api/v1/platform/organizations/:id/settings` — platform admins set all quota and integration fields for any org

## v1.33.15

### Added
- **Platform admin role** (issue #11): `is_platform_admin BOOLEAN` column on `users`; users with this flag bypass org-scoping across all platform endpoints.
- **Platform API namespace** at `/api/v1/platform/` — all routes require `is_platform_admin = true`:
  - `GET /platform/organizations` — list all orgs
  - `GET /platform/organizations/:id` — org detail with user count stat
  - `GET /platform/audit-log` — cross-org audit log; optionally filter by `?org_id=`
  - `POST /platform/impersonate` — returns a 1-hour Bearer token scoped to a target org; all actions taken with it are audit-logged under the requesting platform admin's identity
  - `PUT /platform/users/:id/platform-admin` — grant or revoke `is_platform_admin` on any user (platform admin only)
- **Impersonation token mechanics**: `api_tokens.impersonated_org_id` column added; when a token carries this column, `AuthMiddleware` uses it as the effective `orgID` instead of the user's native org, so the platform admin seamlessly acts within the target org's data scope.
- **`PLATFORM_ADMIN_EMAIL` env var**: on startup, if set, the matching user is idempotently promoted to platform admin (safe to leave in place after first boot).
- **`PlatformAdminMiddleware`**: Fiber middleware that returns 403 to anyone without `is_platform_admin = true`.

## v1.33.14

### Added
- **Org-scoped API tokens, webhooks, audit logs, reports, and dashboard** (issue #10): all major data surfaces now carry an `organization_id` column (migration `20260619_003_org_scope.up.sql`) and are automatically filtered to the caller's org via the `orgID` already stored in Fiber context by `AuthMiddleware`.
- **`PermV2PlatformAdmin` (`auth:platform:admin`)**: new admin-only permission that allows cross-org audit log queries via `GET /api/v1/admin/audit?all_orgs=true`, useful for platform-level compliance audits.
- **`orgIDFromCtx` helper** in `handlers/audit_helper.go`: extracts `*int64` org from Fiber locals; `auditLog` auto-injects it into every `AuditEntry` when not explicitly set.
- **Repository changes**: `ListAPITokenAnalytics`, `ListWebhookEndpoints`, `CreateWebhookEndpoint`, `ListScheduledReports`, `CreateScheduledReport`, `GetDashboardSummary` all accept an `orgID *int64` parameter; `nil` is used by background workers to query across all orgs without filtering.
- **Cache bypass for org-scoped queries**: `IPAMService.GetDashboardSummary` and `ReportsService.ListScheduledReports` skip the global cache when `orgID` is non-nil to prevent cross-org cache poisoning.

## v1.33.13

### Added
- **Delegated administration / scoped role grants** (issue #9): `role_grants` table allows org admins to grant individual permissions to users, either globally or scoped to a specific resource type and ID (`scope_type`, `scope_id`).
- **`CheckPermission` extended**: after legacy-role and custom-role checks fail, the identity service now queries `role_grants` for a matching global or scoped direct grant, so delegated users get the right access without elevated legacy roles.
- **`IdentityService.CreateGrant`**: validates the grantor holds the permission being granted (preventing privilege escalation) before inserting the row.
- **Grant CRUD API**: `GET /api/v1/admin/users/:id/grants` (PermV2OrgRead), `POST /api/v1/admin/role-grants` (PermV2OrgWrite), `DELETE /api/v1/admin/role-grants/:id` (PermV2OrgWrite).
- **Admin UI**: expanded user row in Admin → Users now shows a "Direct Permission Grants" section with per-grant revoke confirmations and an "Add Grant" modal (permission input, optional scope type + scope ID).

## v1.33.12

### Added
- **Organizations scaffold** (issue #8): multi-tenancy foundation with an `organizations` table, nullable `organization_id` FK on `users`, and a seed migration that assigns all existing users to a default "Default" org on upgrade.
- **`OrganizationService`** (`services/organization_service.go`): `Create` (validates lowercase-alphanumeric-hyphen slug), `Get`, `List`, `Delete`, `EnsureDefault` (idempotent startup seeder).
- **`repository/organizations.go`**: `CreateOrganization`, `GetOrganization`, `GetOrganizationBySlug`, `ListOrganizations`, `DeleteOrganization`, `OrganizationExists`, `EnsureDefaultOrganization`.
- **`models.Organization`** struct; `User.OrganizationID *int64` field added to the `User` model and reflected in all repository SELECT/Scan calls.
- **Admin REST API** under `/api/v1/admin/organizations`: `GET` (list), `POST` (create), `DELETE /:id`. Protected by new `auth:org:read` / `auth:org:write` permissions (admin-only via legacy role map).
- **`orgID` in Fiber context**: `AuthMiddleware` and `OptionalAuthMiddleware` now store `user.OrganizationID` in `c.Locals("orgID")` for both session-cookie and Bearer-token auth paths.
- **Startup seeding**: `main.go` calls `svc.Ops.Organizations.EnsureDefault(ctx)` after admin password init, creating the default org and assigning orphaned users on first boot after upgrade.

## v1.33.11

### Added
- **`docs/compatibility.md`**: API compatibility policy defining the support window (v1 active → maintenance at v2.0.0 → end of life at v3.0.0 earliest), breaking change definition, non-breaking addition categories, deprecation process (minimum two release cycles + `Deprecation` header), breaking change registry, endpoint stability table, and an operator upgrade checklist for API consumers, webhook consumers, and infrastructure operators.
- Linked from `docs/migration-v1-to-v2.md` and `README.md`.

## v1.33.10

### Fixed
- **Data race in `JobService.Get()`**: the read lock was released before `publicJob()` copied the job struct, allowing a concurrent `run()` goroutine to write `FinishedAt` while the copy was in progress. Lock is now held for the duration of the copy. Caught by `-race` in CI.
- **Missing `-- +migrate Up/Down` annotations** in `20260618_001_background_jobs` migration files, causing the CI annotation check and the backend health check to fail.

## v1.33.9

### Added
- **`--migrate-dry-run` flag**: pass `--migrate-dry-run` to the server binary to print all pending migration IDs and their SQL without applying them, then exit cleanly. Useful for pre-deploy validation.
- **`V1_COMPAT_SUNSET` env var**: set to an ISO 8601 date (`YYYY-MM-DD`) to log a startup warning reminding operators to migrate API consumers before v1 routes are retired.
- **`docs/migration-v1-to-v2.md`**: comprehensive migration guide covering response shape changes (new `{ data, meta }` envelope), field renames (`colour` → `color`), endpoint mapping table, webhook payload changes with before/after examples, automation script updates, and operator configuration reference.

## v1.33.8

### Added
- **Typed event bus**: `EventBus` in `services/event_bus.go` provides synchronous in-process pub-sub with typed `Subscribe` / `Publish`. Handlers run in the publisher's goroutine; panics per-handler are caught and logged so a bad subscriber cannot crash the caller. Core domain event types defined in `services/events.go`: subnet CRUD, IP CRUD, scan completed, user login/logout, and workflow request events.
- **AuditService subscribes to bus**: `AuditService.SubscribeTo(bus)` registers a wildcard handler that calls `Log()` for any `AuditableEvent`. `OpsManager` exposes the bus as `EventBus`.
- **WorkflowService decoupled from AuditService**: `WorkflowService` now receives `*EventBus` instead of `*AuditService`. The one service-layer audit call (`request_comment_added`) is replaced with `bus.Publish(RequestCommentAddedEvent{...})`.

## v1.33.7

### Added
- **Persistent background jobs**: job state is now written to a new `background_jobs` PostgreSQL table (migration `20260618_001_background_jobs`). `JobService` inserts a row on `Enqueue`, marks it `running` when the goroutine starts, and writes the final status, progress percentage, diagnostics, and result on completion. `List()` merges live in-memory jobs with DB history so completed jobs survive process restarts. `Get()` falls back to the DB when the job is no longer in memory. `Cancel()` updates the DB row. On startup, any rows still marked `running` are automatically reset to `failed` (crash recovery). `Retry()` still requires the runner closure to be in memory (closures can't be serialized). `NewJobService(nil)` keeps the original in-memory-only behavior for tests.

## v1.33.6

### Added
- **v2 API scaffold**: introduced `/api/v2` route group with standard response envelope `{ "data": ..., "meta": { "page", "limit", "total" } }`. Helper functions `V2List`, `V2Item`, and `V2Meta` in `handlers/v2_response.go`. Reference endpoint `GET /api/v2/networks` returns paginated networks in the v2 envelope. `API_V2_BASE` constant exported from `frontend/src/api/client.js`.
- **Deprecation headers on v1 networks**: `GET /api/v1/networks` now emits `Deprecation: true` and `Link: </api/v2/networks>; rel="successor-version"` on every response (including 401/403) per RFC 8594.

## v1.33.5

### Internal
- **Workflow domain extraction**: extracted `WorkflowService` from the root `Service` struct — subnet requests, IP request approval (including auto-allocation and DNS sync), request comments, and all custom field definition/value methods (~33 methods) now live in `services.WorkflowService`, exposed via `OpsManager.Workflow`. `WorkflowService` receives `*IPAMService`, `*DNSService`, `*AuditService`, and `*NotificationService` at construction time. `IPAMService` extracted as a local var in `NewService` so it can be shared with `WorkflowService`. Handler files updated to `h.ops.Workflow.*`. Service test file (`requests_test.go`) updated to use `&WorkflowService{}`. `docs/domain-boundaries.md` updated; residual table replaced with a note that all domains are extracted.

## v1.33.4

### Internal
- **Customers domain extraction**: extracted `CustomerService` from the root `Service` struct — customer CRUD and customer associations now live in `services.CustomerService`, exposed via `OpsManager.Customers`. `CustomerService` receives `*repository.Repository` at construction time. `ListCustomersPaginated` moved from `dashboard.go` to `CustomerService`. Handler files updated to `h.ops.Customers.*`. `docs/domain-boundaries.md` updated; Customers removed from the residual table.

## v1.33.3

### Internal
- **Infrastructure domain extraction**: extracted `InfrastructureService` from the root `Service` struct — devices (with SNMP credential encryption/decryption, interfaces, IP associations), racks, locations (including tree builder and pagination), and nameservers now live in `services.InfrastructureService`, exposed via `OpsManager.Infrastructure`. `InfrastructureService` receives `*repository.Repository` and `encryptionKey` at construction time. `SetCustomFieldValues` logic inlined as a private method using the repo directly (no cross-domain dep). `ListLocationsPaginated` moved from `dashboard.go` to `InfrastructureService`. Handler files updated to `h.ops.Infrastructure.*`. `*Service` forwarding stub added for `CreateDevice` (required by the `automationIPAM` interface). Service test files updated to use `&InfrastructureService{}` directly. `docs/domain-boundaries.md` updated; Infrastructure removed from the residual table.

## v1.33.2

### Internal
- **Identity domain extraction**: extracted `IdentityService` from the root `Service` struct — users, RBAC roles/permissions, API tokens, web sessions, password management, account security (lockout/unlock), and the Grafana datasource proxy (~76 methods) now live in `services.IdentityService`, exposed via `OpsManager.Identity`. `IdentityService` receives `*ConfigService`, `*EmailService`, `*MFAService`, and `*NotificationService` at construction time (no monolithic back-reference). All handler files updated to `h.ops.Identity.*`; `*Service` forwarding stubs added for `InitAdminPassword`/`ForceResetAdminPassword` (called from `main.go`). Service source files gutted to `package services`. Unit and integration tests updated to call methods via `svc.Ops.Identity.*`. Handler tests updated to use `minHandler()` helper so `requirePerm` can resolve `IdentityService.CheckPermission` without a nil dereference. `docs/domain-boundaries.md` updated; Identity removed from the residual table.

## v1.33.1

### Internal
- **IPAM domain extraction**: extracted `IPAMService` from the root `Service` struct — networks, subnets, IP addresses, VRFs, VLANs, VLAN domains/groups, tags, search, dashboard summary/activity, IPv6 delegations, and subnet split/merge/resize now live in `services.IPAMService`, exposed via `OpsManager.IPAM`. `IPAMService` receives `*ConfigService` and `*DNSService` at construction time (no monolithic back-reference). All handler files updated to `h.ops.IPAM.*`; `*Service` forwarding stubs added for `automationIPAM` interface compatibility. Service source files gutted to `package services`. Unit tests updated to call methods via `svc.Ops.IPAM.*`.

## v1.33.0

### Internal
- **Domain module boundaries** (#2): extracted `NetworkModulesService` from the root `Service` struct — NAT rules, firewall zones/mappings, DHCP servers/leases, circuit providers/circuits, and BGP autonomous systems now live in `services.NetworkModulesService`, exposed via `OpsManager.NetworkModules`; handlers updated to `h.ops.NetworkModules.*`. Moved misplaced `CustomerAssociation` methods from `network_modules.go` to `customers.go`. Moved `ListAutonomousSystemsPaginated` from `dashboard.go` to the new service. Integration tests updated to construct `NetworkModulesService` directly. Added `docs/domain-boundaries.md` defining the full planned domain extraction map.

## v1.32.17

### Internal
- **Dependency updates** (#216): frontend ESLint 9→10, migrated from `eslint-plugin-react` to `@eslint-react/eslint-plugin` (ESLint 10-native); added `@eslint/js` and `typescript` as explicit deps; fixed missing `key` prop on fragment in `AdminUsersPage`. Backend: `go-oidc/v3` v3.18→v3.19.

## v1.32.16

### Internal
- **OpsManager Step 3** (#192): broke `AuditService`'s `*Service` back-reference (now takes `repo auditRepo`, `config *ConfigService`, `webhooks *WebhookService`); created `AuthManager` bundling Email, Registration, MFA, Notification, LDAP, OAuth2, SAML. `Service.Auth *AuthManager` replaces those 7 fields; `Handler.auth *AuthManager` is injected alongside `Handler.ops`. Handler files updated to `h.auth.*`; service-layer callers updated to `s.Auth.*`. `Service` struct is down to 6 fields (Config, Audit, Auth, Ops + 2 private).

## v1.32.15

### Internal
- **OpsManager Step 2** (#192): broke `*Service` back-references in `DNSService`, `AutomationService`, and `TelemetryService`. Each now receives narrow interfaces/concrete deps (config, repo, LDAP/OAuth2/SAML) instead of the whole `*Service`. All three are now housed in `OpsManager`; `Service` no longer holds DNS, Automation, or Telemetry fields. Handler files updated to `h.ops.DNS.*`, `h.ops.Automation.*`, `h.ops.Telemetry.*`. `Service` struct is down to 12 fields (Config, Email, Registration, MFA, Audit, Notification, LDAP, OAuth2, SAML, repository, encryptionKey, Ops).

## v1.32.14

### Internal
- **OpsManager extraction** (#192 Step 1): extracted Discovery, Reports, Import, Jobs, Webhooks, and Topology out of the monolithic `Service` struct into a new `OpsManager` struct (`services/ops_manager.go`). `Service.Ops` holds the manager; `Handler.ops` receives it directly. All 17 ops handler files updated to use `h.ops.*`. Sub-services with `*Service` back-references (DNS, Automation, Telemetry) remain on `Service` for now.

## v1.32.13

### Internal
- **Backend refactor** (#187, #188, #194, #191, #203): rename `permCheck` → `requirePerm` (returns `bool`), stop leaking internal error details in 500-level backup responses, remove redundant pagination pre-reads in 10 list handlers, add production `sslmode=disable` startup check, and add `go mod verify` to CI and `make ci-local`.
- **Frontend refactor** (#195–#202): extract shared `SortTh` component; introduce `ToastContext` and global Axios error interceptor; decompose `SubnetsPage` into `useSubnetModals`, `useSubnetSearch`, and `SubnetTable`; fix suppressed `useEffect` dependency arrays in `DevicesPage`, `IPAddressesPage`, `LocationDetailPage`, and `SubnetsPage`; lift feature-flag fetching into a single `FeaturesProvider`; add `VITE_API_URL` env var support; add 15 s Axios request timeout and Fiber `ReadTimeout`/`WriteTimeout`.
- **Auth provider validation** (#191): `UpdateLDAPConfig`, `UpdateOAuth2Config`, and `UpdateSAMLConfig` now reject `enabled: true` when required credentials (host/baseDN, clientID/URLs, IDP metadata/entityID/ACS) are missing, returning a 400 with a clear message instead of failing at login time.
- **Auth session consolidation** (#190): extract `issueSessionCookie` helper to `auth_shared.go`; OAuth2 and SAML redirect flows now call it instead of duplicating session creation, cookie writing, and audit logging inline.
- **N+1 threshold alert fix** (#189): `CheckThresholdAlerts` now uses `BulkGetLatestUtilization` and `BulkGetAlertCooldowns` to reduce 3N queries to 3 queries regardless of subnet count; cooldown upserts and clears are batched similarly.
- **Audit old/new values** (#204): `UpdateSubnet`, `UpdateNetwork`, and `UpdateIPMeta` now fetch the record before mutating and populate `OldValues` in the audit log entry alongside the existing `NewValues`. `UpdateIPMeta` was also missing its audit log entirely — that is now added.
- **OpenAPI sync enforcement** (#193): `make check-openapi-sync` diffs `docs/openapi.yaml` against `backend/docs/openapi.yaml` and fails loudly if they diverge; CI runs this step before every backend build.
- **Env var documentation** (#205): `.env.example` documents the missing `SESSION_COOKIE_SECURE` variable with accepted values and behavior. New `make validate-env` target checks required variables are set before `docker compose up`.

## v1.32.12

### Improvements
- **Sortable Networks and Subnets lists**: click any column header to sort ascending or descending. Networks sort by Name or Description. Subnets sort by Network address (IP-numeric order), Prefix length, or Description. Sort preference is persisted per-browser across page reloads.

## v1.32.11

### Bug Fixes
- **Subnet resize conflict modal**: conflicting IPs and subnets now display as readable strings (e.g. `192.168.1.5 (hostname)`, `10.0.0.0/24`) instead of `[object Object]`.

## v1.32.10

### Features
- **Telemetry opt-in setup page**: admins who have never configured telemetry are redirected to `/setup/telemetry` on first login. The page explains in plain English exactly what is collected, gives explicit assurances that data is never used for marketing or sales, is never sold or shared, and is completely anonymous. An expandable section lists every category of collected data. Two buttons — "Enable Telemetry" and "No Thanks" — record the choice and return to the dashboard. The decision can be changed at any time in Admin Settings → Telemetry.

## v1.32.9

### Improvements
- **Telemetry destination is now hardcoded**: snapshots always go to `base.lima3.dev` — the PocketBase URL and token fields have been removed from the admin UI and config. Authentication uses a public write-only key sent as `X-Padduck-Analytics-Key`; access is further gated by the PocketBase collection create rule. The body now includes `analytics_key_version: "v1"` for future rule versioning. `subnet_utilization_avg_pct` is always sent as a number (0 when no IPv4 subnets exist) to satisfy the collection's `>= 0` rule check. The Telemetry settings tab now shows the destination hostname as a read-only note.

## v1.32.8

### Improvements
- **Telemetry: deployment type and mode selectors**: the Telemetry settings tab now includes "Deployment Type" (Docker, Docker Compose, Kubernetes, Bare Metal) and "Deployment Mode" (Self-Hosted, On-Premises, Development, Test/Staging) dropdowns. Selections are stored in config and included in every snapshot as `deployment_type` and `deployment_mode`. The `edition` field is now sent as `"community"` rather than `"unknown"`.

## v1.32.7

### Features
- **Telemetry opt-in UI and sender**: adds the Admin Settings > Telemetry tab, the scheduled background sender, and an admin "Send test snapshot now" action — completing the opt-in analytics feature. The Telemetry tab lets admins enable or disable telemetry, configure the destination PocketBase URL and service token, choose a daily or weekly snapshot period, and set optional locale fields (UI locale, timezone region, country/region codes). A collapsible "What is collected?" section explains exactly what the snapshot contains with plain-English disclosure. Clicking "Send Test Snapshot Now" posts a snapshot immediately and reports success or failure inline. The background job starts with the server and fires on the configured period (24 h or 168 h); it is a no-op when telemetry is disabled. The PocketBase token is masked on read and stored encrypted via the existing sensitive-key pattern.

## v1.32.6

### Improvements
- **Telemetry snapshot now collects all schema fields**: the `CollectSnapshot` method previously omitted active-user counts, subnet utilization statistics, locale metadata, and JSON extension fields. All fields defined in the `padduck_analytics` schema are now populated. Active users (7-day and 30-day) are derived from `audit_logs` so both UI sessions and API token activity are counted. IPv4 subnet utilization percentiles (mean, median, p75, p90, p95) and threshold bucket counts are computed live from the IP address table using the same theoretical-capacity formula as the dashboard. Feature flag states are collected into `feature_flags_json`; `extra_metrics_json` includes `devices_total`. Locale fields (`ui_locale`, `timezone_region`, `country_code`, `region_code`) are read from admin config keys that will be exposed in the opt-in settings UI in the next increment. Nothing is transmitted — the opt-in toggle and sender follow in subsequent increments.

## v1.32.5

### Security
- **Patch libssl3 and libcrypto3 to 3.5.7-r0 in Alpine runtime images**: explicitly pins `libssl3=3.5.7-r0` and `libcrypto3=3.5.7-r0` in the `backend` and `agent` Dockerfiles (both use `alpine:3.22` as the runtime base). The frontend image is unaffected — it uses a Debian base and already runs `apt-get upgrade` at build time.

## v1.32.4

### Bug Fixes
- **VLAN detail page "Network" column always showed "—"**: the subnets table on the VLAN detail page referenced `subnet.sectionId` (always undefined after the camelCase interceptor) instead of `subnet.networkId`. The Network column link and the "Networks" relationship count in the summary panel both now use the correct field.
- **Subnet VLAN assignment not visible on return navigation**: after assigning a subnet to a VLAN from the VLAN detail page and navigating back to Networks > Subnets, the subnet's VLAN column appeared blank. The subnets page only re-fetched data when the network ID changed, so returning to the same URL served stale data. The page now also re-fetches whenever the navigation key changes, so any return trip triggers a fresh load.

## v1.32.3

### Bug Fixes
- **Dashboard utilization calculation was wrong**: the overview "assigned/total IPs" stat and utilization percentage were calculated using the count of IP records in the database, not the theoretical address capacity of subnets. Total IPs now equals the sum of `GREATEST((2^(32-prefix_length)-2), 1)` across all IPv4 subnets, so utilization reflects real subnet capacity.
- **"Add IP" error shown below modal instead of inside it**: validation errors (invalid address, API rejection) from the "New IP Address" modal were surfacing on the subnet page underneath the open modal where users could not see them. Errors now display inline inside the modal above the submit button.
- **MAC address input allowed consecutive separators**: the MAC address field now prevents consecutive separator characters (`:`, `-`, `.`, space) and caps total input length at 17 characters.

### Improvements
- **American spelling throughout**: all user-visible text, JSON field names, Go symbol names, API routes, and in-code identifiers now use American spelling (`utilization`, not `utilisation`). The database table `subnet_utilisation_history`, column `utilisation_pct`, config key `utilisation_snapshot_interval_hours`, and stored report type value `utilisation_summary` are preserved unchanged. The API route `/api/v1/subnets/:id/utilisation/history` is renamed to `/api/v1/subnets/:id/utilization/history`.

## v1.32.1

### Bug Fixes
- **Check-changelog CI failed on minor/major version bumps**: the `check-changelog` workflow previously required the CHANGELOG top version to be exactly `next_patch(latest_tag)`, which rejected valid minor or major bumps (e.g. v1.31.42 → v1.32.0). The check now accepts any semver version strictly greater than the latest tag.

## v1.32.0

### Features
- **Telemetry foundation (silent, nothing sent yet)**: adds the data collection layer for the opt-in analytics feature planned for v1.50.0. A stable `install_id` UUID is generated on first boot and persisted in the config table — never derived from hostname, IP, or any identifying property. A new `TelemetryService.CollectSnapshot()` method assembles a `TelemetrySnapshot` struct from a single-round-trip batched COUNT query (users, customers, locations, VLANs, subnets with IPv4/IPv6 split and five prefix-length bucket counts) plus feature flag reads (LDAP, OIDC, SAML, SNMP community, anonymous API). Field names match the `padduck_analytics` PocketBase schema exactly. No data is transmitted and no UI changes ship in this increment; the opt-in toggle and sender follow in a future patch.

## v1.31.42

### Tests
- **Expanded handler test coverage for issues #153, #154, #155**: added 19 new unit tests across three handler test files. `audit_test.go` gains 10 `buildAuditFilter` tests covering default limit, custom limit, invalid/negative limit (ignored), offset, all string filter params, resource_id, and since/until date parsing. `custom_fields_test.go` gains 6 `validateCustomFieldParams` tests covering valid params, all entity_type/field_type combinations, invalid entity_type, invalid field_type, both invalid, and empty strings. `search_test.go` gains 3 auth tests (no-user → 401) for `GlobalSearch`, `SearchNetworks`, and `SearchIPAddressesGlobal`.

## v1.31.41

### Bug Fixes
- **Frontend CI failed on ESLint warnings after v1.31.40 merge**: patched `form-data` past 4.0.5 to resolve GHSA-hmw2-7cc7-3qxx (CRLF injection), and cleared all 36 pre-existing ESLint warnings across 31 frontend files — `react-hooks/exhaustive-deps` violations (functions missing from `useEffect` dependency arrays), `no-unused-vars` (dead state variables and unused imports), and one unused `eslint-disable` directive. No functional changes.

## v1.31.40

### Security
- **Webhook URLs are now validated against SSRF attacks**: the create/update webhook endpoint previously accepted any string as a URL. It now rejects non-`http`/`https` schemes and destinations that resolve to loopback or RFC-1918 addresses (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, link-local). Invalid URLs return a 400 with a clear field-level error.

### Features
- **Per-page size selector on the IP Addresses page**: a dropdown in the pagination bar lets users choose between 10, 25, 50, and 100 rows per page.
- **Customers page now shows Associations tab**: the Customers detail view gains an Associations tab listing all objects linked to a customer. The `network` object type is also now accepted when adding or editing an association.

### Bug Fixes
- **Unbounded `?limit` parameter could cause excessive memory use**: `parseListOptions` now silently caps the `limit` query parameter at 1000 rows.
- **Oversized string inputs reached the database without a clear error**: length limits (255 characters, matching the DB column size) are now enforced at the handler level for username, hostname, location name, and webhook name. Violations return a 400 instead of a cryptic Postgres error.
- **Invalid MAC addresses were not rejected at the API boundary**: `CreateIPAddress` and `UpdateIPMeta` now validate and normalize MAC addresses using the existing `NormalizeMAC` helper. Invalid formats return a 400; valid addresses are stored in lowercase colon-separated form.
- **Custom field `entity_type` and `field_type` produced opaque DB errors when invalid**: the create and update custom field handlers now validate both fields against their allowed value sets before hitting the database.

## v1.31.34

### Features
- **Technitium DHCP integration — lease sync, scope import, and reservation push**: when a Technitium DNS/DHCP server is configured in Admin Settings, a new DHCP section on the DNS settings tab exposes three capabilities. "Sync Leases Now" pulls all leases from every configured DHCP scope and upserts them into the DHCP Leases table, creating a sentinel DHCP Server record (vendor: technitium) on first sync and linking each lease to the matching subnet and IP address where possible. "Load Scopes" lists all Technitium DHCP scopes; individual scopes can be imported as new subnets in a chosen network. Subnets can be linked to a Technitium scope name (via the subnet edit modal) to enable per-IP reservation push: assigned IP addresses on linked subnets show "Reserve" and "Unreserve" buttons that create or delete static DHCP reservations in Technitium, provided the IP has a MAC address configured.
- **Customer association picker rework**: the "Add Association" form (Customers page) is now a modal dialog instead of a cramped inline row. Object Type labels are human-readable ("IP Address", "DHCP Server", etc.). The raw Object ID number input is replaced by a context-sensitive combobox: selecting a type loads the appropriate list (devices, VLANs, racks, locations, circuits, etc.) and filters as you type; subnets and IP addresses use a search-as-you-type approach via the global search endpoint. The associations table now shows the resolved object name (e.g. "10.10.0.0/24 — Office LAN") instead of the raw database ID.
- **Interface name combobox when associating an IP with a device**: the interface name field on the "Associate IP" modal now searches the device's existing interfaces as you type. Matching interfaces appear in a dropdown for quick selection. If no match is found, a "Create interface" option is shown that creates the interface immediately and selects it, eliminating the need to navigate away to the device interfaces tab first.

### Bug Fixes
- **SNMP community string could not be revealed after saving**: the global SNMP community string field on the Admin Scanner settings tab always showed a masked placeholder after page load. The reveal button (eye icon) now fetches the real value from the server on first click and caches it locally, so the field can be revealed at any time, including immediately after saving. The existing `GET /admin/config/reveal` endpoint used for other sensitive fields (SMTP password, PowerDNS key) has been extended to cover the SNMP community string.
- **IP addresses displayed with a /32 CIDR suffix on the device detail page**: querying IP addresses associated with a device cast the `address` column to text using `::text`, which includes the host prefix (e.g. `192.168.1.5/32`). Changed to `host()` which returns only the bare IP address. The same fix was applied to the inactive IPs report, conflict export, DHCP lease list, login history, and the device interface list — all of which could show spurious `/32` or `/128` suffixes.
- **New IP Address dialog showed "[object Object]" as the pre-filled network prefix**: the "New IP" button in a subnet's IP address list called `onClick={openCreate}`, passing the browser click event as the `prefillAddress` argument. Changed to `onClick={() => openCreate()}` so no argument is passed and the function correctly derives the network prefix from the subnet.

## v1.31.33

### Bug Fixes
- **Show all IPs displayed a /32 suffix after every address in full range view**: the generate_series SELECT expression cast the computed inet address to text (`::text`), which includes the host prefix length. Changed to `host()` to return the bare IP address, matching the existing behaviour of the regular IP list query.
- **IP address global search never returned results**: `GET /ip-addresses/search` was registered after `GET /ip-addresses/:id` in the router, so Fiber matched the literal string "search" as an id parameter and called `GetIPAddress` instead of `SearchIPAddressesGlobal`. Static paths are now registered before the parameterised routes. This also resolves the "IP already exists" error on the device detail page — that error appeared because users hit quick-create on an address that already existed, since search had returned nothing.
- **Edit IP dialog had no way to set or change the associated device**: the Edit IP modal now includes a device search field. Typing filters the loaded device list; selecting a device associates it with the IP; a clear button removes the association. The current device is pre-populated when the modal opens.
- **Notification emails showed `<no value>` for the timestamp field**: the `Queue` function populated `Username` and `AppURL` template variables automatically but never populated `Timestamp`, so all 10 notification templates rendered the timestamp placeholder as `<no value>`. The timestamp is now injected by the queue function at enqueue time.
- **Notification preference checkboxes always appeared checked and toggling was inverted**: the preferences `useEffect` in the notifications settings tab assigned the raw snake_case API response (`login_success`, `login_failed`, …) directly to state that was read with camelCase keys (`loginSuccess`, `loginFailed`, …). Every preference read as `undefined`, which the `?? true` fallback treated as enabled, so all boxes appeared checked regardless of actual saved values. The first toggle set the value to `true` (not `false`), meaning a user who toggled a preference once and saved accidentally kept it enabled. The response is now explicitly mapped to camelCase on load.

## v1.31.32

### Security
- **API tokens with non-admin scope could reach admin-only endpoints**: 37 handlers used an inline `Role != "admin"` check that ignored the token scope entirely, allowing a scoped API token (e.g. read-only) issued to an admin-role account to reach admin-only operations. All inline checks have been consolidated to the `requireAdmin` helper, which additionally enforces that the token's scope must be `"admin"` (or absent, for browser sessions). The two "self OR admin" ownership patterns in user and request handlers are unaffected.

### Bug Fixes
- **PHPIpam import silently discarded all IPs with non-reserved statuses**: `phpIpamStateToStatus` returned `"active"`, `"dhcp"`, and `"inactive"`, all of which violate the `ip_addresses.status` DB constraint (`available`, `assigned`, `reserved`). All INSERT rows were rejected. PHPIpam states 1/used/active and 3/dhcp now map to `"assigned"`; the default fallback maps to `"available"`.
- **LDAP group sync errors were silently dropped**: the comment at the call site said "log but don't block login" but the error was discarded with `_ = err`. The error is now logged via `slog.Warn`.
- **LDAP role assignment suppressed all errors, not just duplicate-key violations**: `AssignRoleToUser` errors were fully ignored. Non-duplicate errors (DB connectivity failures, permission issues) are now logged; only `23505 unique_violation` is still silently skipped.

### Changes
- **Bulk IP delete now issues a single query**: `BulkDeleteIPs` previously called `DeleteIPAddress` in a loop — one round-trip per ID. It now issues a single `DELETE … WHERE id = ANY($1) RETURNING id` and kicks off DNS cleanup goroutines for the deleted records.
- **Handler error responses are now consistent**: 716 inline `c.Status(…).JSON(fiber.Map{"error": …})` calls across 47 handler files have been replaced with `RespondError`, which adds a structured `code` field alongside `error` and routes through the request logger.
- **Service-layer logging migrated to structured slog**: 52 `log.Printf` calls across `dns.go`, `audit.go`, `notification.go`, `requests.go`, `reports.go`, and `registration.go` converted to `slog.Error`/`slog.Warn`/`slog.Info` with key-value pairs. `[tag]` prefixes dropped.
- **Discovery service magic numbers extracted to named constants**: `defaultPingConcurrency` (20), `maxPingConcurrency` (100), `defaultScanResultLimit` (100), `maxScanResultLimit` (1000), `defaultPortList`, and `agentOfflineThreshold` (15 min) replace inline literals across `discovery.go`.
- **Cron helpers moved to a dedicated file**: `matchesCron` and `validateCron` were defined in `discovery.go` but also consumed by `reports.go`. Extracted to `services/cron.go`.
- **Repository and handler files split for maintainability**: `repository/devices.go` (845 lines) → 6 files; `repository/reports.go` (780 lines) → 4 files; `handlers/external_auth.go` (617 lines) → 4 files by auth provider.
- **Large frontend page components split into focused sub-components**: `IPAddressesPage.jsx` (1,486 → ~580 lines), `UserSettingsPage.jsx` (1,187 → 60 lines), `DeviceDetailPage.jsx` (1,084 → 500 lines), `SubnetsPage.jsx` (1,032 → 771 lines). Extracted components live under `src/pages/ip/`, `src/pages/user/`, `src/pages/device/`, `src/pages/subnet/`.

### Cleanup
- Removed dead helper functions in `services/reports.go` (`cidrFromSubnet`, `parseIPNet`, `subnetNetworkName`) and a dead model sentinel in `services/mfa.go` that were kept alive only by `var _ =` guards.
- Removed redundant `loggingMiddleware` in `handlers/handlers.go` — fully superseded by the Fiber logger registered in `main.go`.
- Removed dead `Service.ListDevices` and `Repository.ListDevices` wrappers never called by any handler.
- `scanDevice` and `scanInterface` in `repository/devices.go` now accept `interface{ Scan(...any) error }` instead of the concrete `pgx.Row`, consistent with all other scan helpers in the repository.

## v1.31.30

### Bug Fixes
- **Dockhand/Grype could not scan the frontend image on Docker 29 with the containerd image store**: the frontend image was published as a single Docker schema-v2 manifest, which Docker 29's containerd-backed `overlayfs` store exports as an incomplete archive (manifest descriptor only, no blobs). Both Trivy and Grype failed with "file blobs/sha256/… not found in tar". The release and deploy workflows now build the frontend image with `platforms: linux/amd64`, `provenance: mode=min`, and `sbom: true`, which causes `build-push-action` to publish an OCI image index. OCI indexes export correctly on Docker 29 and pass the scanner's local archive path.

## v1.31.29

### Bug Fixes
- **Adding a device interface with a VLAN ID always failed with a foreign key violation**: the VLAN ID field accepted a raw number that was sent as the VLAN's database primary key, but users naturally typed the 802.1Q tag number (e.g. 101) which rarely matches the DB ID. The field is now a dropdown populated from the configured VLANs list and submits the correct database ID. The interface table now shows the VLAN tag and name instead of the opaque DB ID.
- **Scan profile SNMP settings were stored but never applied during discovery**: discovery jobs read SNMP community string and version from global config only, ignoring per-profile overrides. Profile SNMP fields are now copied onto the job at run time and applied after global config is read, so per-profile overrides take effect correctly.
- **DNS zone serials not populated from Technitium or PowerDNS**: the `Zone` struct in both provider clients was missing the `serial` field, so the DNS Zones page always showed no serial. Both client structs and the provider-agnostic `ZoneInfo` type now include the serial.
- **IP update errors returned a generic "Failed to update IP" message**: the `UpdateIPMeta` handler was returning HTTP 500 for all service errors (including user-input validation like an invalid MAC address), discarding the error message. It now returns HTTP 400 with the descriptive error string, consistent with the create-IP endpoint.

### Changes
- **IP address list toggle: replaced two checkboxes with a slide toggle**: the "Hide unassigned" and "Show all IPs" checkboxes have been replaced with a single slide toggle. Default (toggle off) hides unassigned addresses; toggling on shows the full CIDR range including unrecorded IPs. Full-range mode is paginated server-side so large subnets (e.g. /16) remain usable.
- **DNS nameserver auto-populate is now opt-in**: when a Technitium DNS server is configured in Admin Settings, saving the DNS config now shows a "Also add as a nameserver" checkbox instead of silently creating a nameserver record. An optional name field lets the record be labelled before saving.
- **Technitium nameserver hint in scan job settings**: when a scan job's type is set to SNMP or Ping+SNMP, a hint is shown below the selector explaining that the SNMP community string is set in Admin Settings → Scanner and that per-subnet overrides can be configured in Scan Profiles.
- **MAC address input filters non-hex characters**: the MAC address fields on the IP address list page now silently drop any character that is not a hex digit, colon, dash, dot, or space as the user types, preventing accidental invalid input before submission.
- **Device IP association: create IP if not found**: when searching for an IP to associate with a device and no match is found, a "Create [address] and select" button appears (shown when the search input looks like a valid IP address). Clicking it calls the quick-create endpoint which places the IP in the most-specific matching subnet and immediately selects it.

## v1.31.28

### Bug Fixes
- **IP addresses could be saved outside their subnet's CIDR**: the create-IP form pre-fills the network prefix but allowed users to backspace it and enter any address. The backend now validates that the submitted IP falls within the subnet's network address and prefix length before inserting; the frontend validates the same for IPv4 addresses and surfaces an inline error immediately.
- **DHCP and Circuits pages showed "feature disabled" when the Locations (or Customers) feature was disabled**: both pages called `getLocations()` in the same `Promise.all` as their own feature-gated API calls. A 404 from a disabled secondary feature poisoned the whole load and surfaced the backend's "feature disabled" error message. Locations and Customers are now fetched independently with graceful fallbacks so the primary page data always loads.

### Changes
- **Full range view for IPv4 subnets**: a "Show all IPs" checkbox on the IP address list reveals every address in the subnet's CIDR — including ones not yet in the database. Unrecorded addresses appear dimmed with an "available" badge and a "Create" action that opens the new-IP form pre-filled with that address. The view uses PostgreSQL `generate_series` with page-sized offset arithmetic so only one page of rows is generated at a time — efficient even for /8 subnets. IPv6 subnets are excluded (range too large to enumerate). Sort headers and the "Hide unassigned" filter are suppressed while the full range view is active.
- **Sortable columns and hide-unassigned filter on IP address list**: clicking Address, Hostname, Status, MAC Address, or Last Seen headings sorts the full paginated list ascending/descending via the backend (arrows indicate active sort direction; inactive columns show ↕). A "Hide unassigned" checkbox above the table filters out available IPs server-side so only assigned and reserved addresses are shown.
- **MAC address format validation and normalization**: MAC addresses are now validated and normalized to lowercase colon-separated form (`aa:bb:cc:dd:ee:ff`) wherever they are written. The backend accepts colon, dash, dot (Cisco), and unseparated hex formats — all stored as `aa:bb:cc:dd:ee:ff`. Invalid values are rejected with a descriptive error. The create-IP and edit-IP forms normalize the value on blur so users see the canonical form before saving.
- **IP addresses now link to a device instead of storing a free-text "Assigned To" field**: the `assigned_to` column has been removed from `ip_addresses`. The assign-IP modal now shows a device picker drawn from existing device records. The linked device's hostname appears in the IP list under a "Device" column (with a link to the device detail page), and any lease-expiry badge moves there too. The `assigned_to` field has been removed from all reports, imports, exports, and automation responses.
- **Top Utilised Subnets now uses CIDR-derived capacity**: utilisation percentage and ranking are calculated against the total addressable IPs in the subnet (`2^(32-prefix) - 2`, minimum 1), rather than the count of IP records entered. A /24 with 40 assigned IPs now shows ~15.7% instead of an inflated figure based on however many records exist.

## v1.31.27

### Changes
- **GitHub Actions major version bumps** (Dependabot): actions/checkout 4.3.1 → 6.0.3, actions/setup-node 4.4.0 → 6.4.0, actions/setup-go 5.6.0 → 6.4.0, docker/login-action 3.7.0 → 4.2.0, softprops/action-gh-release 2.6.2 → 3.0.0. All are the Node 20 → 24 runtime migration; reviewed against our usage with no breaking changes (caching is configured explicitly, the Go toolchain is pinned, and only the wiki workflow pushes — with its own token URL). All pins kept as commit SHAs.

## v1.31.26

### Bug Fixes
- **docker-compose default `IMAGE_TAG` pointed at a nonexistent tag**: GHCR image tags have no `v` prefix (`1.31.25`, not `v1.31.25`), so the pinned default added in v1.31.25 could not be pulled by fresh deployments. The default is now `1.31.25` and the docs note the tag format.
- **Changelog gate no longer blocks Dependabot**: the check-changelog workflow now skips Dependabot PRs, which only bump pinned action SHAs and cannot edit the changelog.

## v1.31.25

### Security
- **Username enumeration via login timing**: Login attempts for nonexistent usernames (or accounts with no password set) returned immediately, skipping the ~100ms bcrypt comparison that runs for real users. A dummy bcrypt comparison now runs on those paths so all pre-MFA login failures take the same time. (#75)
- **Account lockout response confirmed account existence**: The lockout check ran before password verification, so the distinct 429 "account locked" response let anyone confirm an account exists by triggering lockout with bogus attempts. The check now runs after password verification: locked accounts still reject the correct password, but callers without valid credentials get the same generic failure as for nonexistent accounts. (#76)
- **Avatar uploads not validated as images**: Custom avatars were stored as client-supplied base64 data URLs with the declared media type trusted and later served back as the Content-Type. Uploads must now decode as a real PNG/JPEG/GIF/WebP within 4096x4096, the stored media type is rewritten to the decoded format, and avatar responses send `X-Content-Type-Options: nosniff`. (#77)
- **SSRF regression tests for webhook deliveries**: The dial-time SSRF guard (`internal/netguard`) was already in place; added regression tests pinning cloud metadata endpoints (IPv4 and IPv6), RFC 1918 ranges, and IPv4-mapped-IPv6 bypass attempts. (#74)
- **Stored XSS via URL custom fields**: URL-type custom field values were rendered directly into anchor hrefs, so a `javascript:` value became a clickable script link. Values are now linked only when they use `http://`/`https://`; anything else renders as plain text. (#69)
- **Admin-configured appUrl rendered unchecked**: The header logo link used the admin-configured `appUrl` without scheme validation; it is now accepted only when it is an http(s) URL. (#73)
- **Content-Security-Policy header**: The frontend nginx config now sends a same-origin CSP (scripts/fonts/styles/connections restricted to `'self'`, with allowances only for inline style attributes, Gravatar images, and data: image previews). (#70)
- **Self-hosted Inter font**: The Inter typeface is now bundled via `@fontsource/inter` instead of loaded from fonts.googleapis.com — no more third-party requests on page load, and `font-src` is locked to `'self'`. (#71)
- **HSTS documentation**: New "Production: TLS and HSTS" section in getting-started.md with nginx/Caddy/Traefik examples, since TLS (and therefore HSTS) terminates upstream of the bundled containers. (#72)
- **Agent: server-supplied CIDR size capped**: A hostile or compromised server could send the agent an arbitrarily broad CIDR (e.g. `0.0.0.0/0`) and OOM it or use it as a mass scanner. Prefixes broader than `/16` are now rejected, and IPs are iterated lazily instead of materialized up front. (#78)
- **Agent: plain-HTTP server URL requires opt-in**: The agent silently sent its bearer token in cleartext when `PADDUCK_SERVER_URL` was `http://`. Startup now fails unless `PADDUCK_ALLOW_INSECURE=true` is set, which logs a prominent warning instead. (#79)
- **Agent: container runs as non-root**: The agent image now runs as a dedicated `padduck` user, with `cap_net_raw` granted to the ping binary at build time so ICMP scanning still works without extra runtime flags. Also fixes the agent image build (builder bumped to go 1.26.4 to match go.mod). (#80)
- **Backup/restore: password no longer visible in `ps`**: The scripts passed the full `DATABASE_URL` (password included) as a `pg_dump`/`psql` argument. The password is now split out and passed via `PGPASSWORD`. (#81)
- **Backup dumps owner-only**: `backup.sh` now sets `umask 077` — dumps (which contain password hashes and encrypted credentials) are created `600` in a `700` directory. (#82)
- **Backup/restore: `DATABASE_URL` required**: The `postgres://padduck:padduck@localhost` fallback is gone; both scripts fail with a clear error when `DATABASE_URL` is unset. (#83)
- **docker-compose: `POSTGRES_PASSWORD` required**: The compose file no longer defaults the database password to `padduck`; startup fails with a clear message until it is set in `.env`. `MFA_ENCRYPTION_KEY` keeps its auto-generation behavior (the backend creates a persistent key on first boot). (#84)
- **docker-compose: hardened service security options**: All services now run with `no-new-privileges:true` and `cap_drop: ALL` (with only the capabilities each service actually needs re-added); backend and frontend containers run with read-only root filesystems and tmpfs mounts. Verified against a full stack boot with healthy services. (#87)
- **GitHub Actions pinned to commit SHAs**: All third-party actions across the five workflow files are pinned to full commit SHAs (each verified against its release tag) with Dependabot configured to keep them updated weekly. (#100)
- **Account lockout was inoperative on non-UTC deployments**: The brute-force window, IP throttle window, and lockout expiry compared local time against UTC-stored rows, so on hosts not running UTC, failed login attempts were never counted and lockout simply never triggered. All time handling in the security service now uses UTC. (found by #91)
- **Sessions now end when the password does**: Neither changing a password nor resetting it by token revoked existing sessions, so an intruder's session survived the password change meant to lock them out. Reset-by-token now revokes every session; changing your own password revokes all sessions except the one making the change. (found by #102)
- **MFA login was broken on non-UTC deployments**: MFA challenges were written with local timestamps into timezone-less columns and read back as UTC, so on any host behind UTC every challenge was instantly expired. Challenge and confirmation writes now use UTC. (found by #101)

### Bug Fixes
- **User queries crashed on passwordless users**: Every user SELECT/RETURNING scanned `password_hash` into a non-nullable string, so any user with a NULL hash (first-boot admin, the plain CreateUser path) broke the query. Now COALESCEd to the empty string the service layer already expects. (found by #90)
- **Device lists crashed for devices without a type**: All seven device list/get queries scanned NULL device-type columns from the LEFT JOIN into non-nullable fields. The shared column list is now NULL-safe. (found by #90)
- **Password-reset rollback was broken**: The `20260513_001_create_sessions` down migration deleted session-timeout keys from `config` while the up migration inserts into `configs` — any rollback past that point failed mid-chain. (found by #99)
- **All local-time database writes audited and fixed**: The four UTC-skew bugs found by the test campaign were instances of one pattern (`time.Now()` written to TIMESTAMP columns). All 47 remaining `time.Now()` sites were audited and the fifteen DB-bound ones fixed (notification retries, OAuth2 state, email verification, impersonation sessions, webhook retries, IP leases, retention cutoffs, and more); a repository-package test now fails on any bare `time.Now()` so the pattern cannot return. (#111)
- **API token expiry skewed on non-UTC deployments**: Token expiry, rotation grace, and session absolute-expiry were written with local time and read back as UTC — extended tokens expired five hours early on a UTC-5 host. All expiry writes now use UTC. (found by #102)

### Testing
- **DB-backed integration tests**: New `internal/testdb` harness gives each test package its own scratch Postgres database (CI runs a postgres:18 service; `make test-integration` boots a throwaway local container). The repository layer's hand-built dynamic SQL — device search, custom-field filters, pagination, injection resistance — is now exercised against real Postgres. (#90)
- **All 99 migration pairs executed in CI**: every migration runs up, down to an empty schema, and up again on every CI run, replacing two permanently-skipped stubs. (#99)
- **MFA service covered end to end** (0% → 88%): TOTP enrollment/confirm/disable, backup-code single-use and regeneration, AES-GCM helpers, and the full challenge flow including replay and expiry. (#101)
- **Password and token lifecycle covered**: change/reset flows with token replay and expiry, admin password init/force-reset, API token rotation and extension. (#102)
- **Auth middlewares at 97-100% coverage** (from 0-18%): session and Bearer auth, token scopes, rate limiting, optional and anonymous-API modes. (#103)
- **Discovery/scan engine and network modules covered**: scan job lifecycle, a real ping scan of TEST-NET asserting the run state machine, the duplicate-run guard, agent token lifecycle, agent result ingestion with auto-add, and CRUD+validation for NAT/firewall/DHCP/circuits. (#104)
- **Login, lockout, CSRF, and session paths covered end to end** through the real routing stack: wrong-password and unknown-user responses pinned byte-identical, lockout to 429, the MFA login flow, CSRF enforcement on cookie-authenticated mutations, and the avatar upload/serve round-trip. Auth service and handler files now average 65-71% (targets were 60%). (#91)
- **Frontend test suite tripled** (9 → 18 files, 65 tests): ProtectedRoute gating, the MFA login step, a full CRUD workflow, and the Dashboard/Devices/DeviceDetail/Subnets pages — including a page-level pin that `javascript:` custom-field URLs render inert. Vitest coverage thresholds are now enforced in CI and ci-local. (#105, #94)

### Changes
- **Configurable log level**: New `LOG_LEVEL` variable (default `warn`; options `info`, `debug`, `error`) controls verbosity independently of the JSON/text format, with a startup line stating the active configuration. Scan engine and webhook worker logs are now leveled: failures at warn (visible by default), lifecycle events behind `LOG_LEVEL=info`. Upgraders who want the previous scan lifecycle lines should set `LOG_LEVEL=info`. (#110)
- **API client split into domain modules**: The 529-line `api/client.js` monolith is now eight domain modules (ipam, auth, devices, dns, vlans, modules, app, admin); `client.js` keeps only the axios core and interceptors. All import sites updated. (#96)
- **Admin Settings split into per-tab components**: The 1,277-line page is now a 187-line shell with eleven tab components under `src/pages/admin/`, each owning its tab-specific state; a new test pins tab navigation and per-tab save behavior. (#97)
- **TanStack Query adopted for data fetching**: `QueryClientProvider` at the root with devtools in development; ten pages migrated (Dashboard, Networks with mutations and cache invalidation, and eight report/list pages), replacing hand-rolled useEffect+axios state. Remaining pages migrate with the same patterns. (#93)
- **Playwright E2E suite**: Six tests run against the full stack built from the working tree (docker-compose.e2e.yml): login success and failure, unauthenticated redirect, session persistence across reload, CSRF enforcement on cookie-authenticated mutations, and network CRUD through the real UI. Runs in CI with trace upload on failure, and locally via `make e2e`. (#106)
- **docker-compose: `IMAGE_TAG` defaults to a pinned release** (`v1.31.24`) instead of `latest`, making deployments reproducible; upgrade instructions documented in `.env.example` and the README. (#85)
- **docker-compose: frontend binds to loopback by default**: The frontend port now binds to `127.0.0.1` (override with `FRONTEND_BIND=0.0.0.0`); the documented production architecture places a TLS-terminating reverse proxy in front. (#86)
- **Releases are now gated on tests**: `release.yml` calls the CI workflow (`workflow_call`) and the release job requires it to pass — a failing backend or frontend test blocks the image push, the GitHub release, and the agent binary uploads. (#98)
- **Frontend error boundaries**: A top-level error boundary replaces the blank-screen failure mode with a friendly error page, and a per-route boundary keeps the header/sidebar usable when a single page crashes (auto-resets on navigation). (#92)
- **`make ci-local` now matches GitHub CI**: the local gate runs `npm ci`, `npm audit`, `npm run lint`, and `npm test` in the same order as the remote pipeline, and `frontend-install` uses `npm ci` for reproducible installs. `npm audit --omit=dev --audit-level=high` also runs in the remote frontend CI job (currently 0 findings). (#88, #89, #95)
- **Agent HTTP layer tested**: httptest-based coverage for heartbeat, job fetch, result posting, endpoint construction, and a full poll cycle — agent module coverage rose from 23% to 72%. (#107)
- **OpenAPI spec version unstuck**: `info.version` was pinned at 1.26.0 since the v1.26 SDK stabilization even though the API surface changed in v1.28; it now reads 1.31.25 and the contract test documents the policy (the spec version tracks the release that last changed the API contract). (#108)

## v1.31.24

### Bug Fixes
- **DHCP sidebar item visible when feature disabled**: Sidebar initialized feature flags from `DEFAULT_FEATURES` (all enabled) before the API responded, causing disabled features to flash or persist in the nav. Features state now starts as `null`; gated items are hidden until the API confirms they are enabled.
- **DNS settings: `dns_auto_add_ips_enabled` rejected on save**: Both `dns_auto_add_ips_enabled` and `dns_auto_remove_ips_enabled` were missing from the config handler allowlist, causing a "unknown config key" error on every save. Added both keys.
- **Config allowlist gaps**: Six additional config keys read by the backend had no write path — `session_idle_timeout_minutes`, `session_absolute_timeout_hours`, `api_token_default_expiration_days`, `api_token_rate_limit_per_minute`, `api_token_rotation_grace_period_hours`, and `privacy_policy_version` — all added to the allowlist.
- **IP hostname not editable after creation**: The edit-IP form did not include the hostname field; `UpdateIPAddressFull` also did not update it. Both fixed.
- **IP `dns_name` not saving**: `dns_name` was sent by the frontend on create and edit but neither the handler request structs nor the repository queries included it. Fixed across the full stack: handler → service → repository → SQL.
- **Utilization history failed to load**: PostgreSQL rejected `($2 || ' days')::interval` when pgx sent a typed `int8` — the `||` operator cannot concatenate integer and text. Changed to `$2 * INTERVAL '1 day'` in all four affected queries.
- **Locations not hidden when feature disabled**: `DevicesPage` did not check the locations feature flag; location columns, filters, and form fields now gate on `features.locations`.
- **IP association search showed no suggestions**: The device-detail IP search dropdown was `absolute`-positioned inside a modal with `overflow-hidden`, clipping it. Switched to inline flow rendering.
- **Interface description text appeared disabled**: Description column color was `text-gray-500 dark:text-gray-400` (too faded); changed to `text-gray-700 dark:text-gray-200`.
- **Bulk delete IPs**: Added a two-step bulk delete action to the IP list (confirmation required). New `POST /api/v1/admin/ip-addresses/bulk-delete` endpoint performs sequential deletion and returns a count.
- **Scan job hover blown out in dark mode**: List rows and result table rows lacked `dark:hover:bg-gray-700/50`; added.

### Changes
- **Scheduled Reports moved to Reports page**: Removed from Admin Settings; now appears as a tab in the Reports page (admin only).
- **Mobile responsiveness**: Added hamburger menu and slide-in sidebar drawer for small screens. Tables wrapped in `overflow-x-auto` across all list pages to prevent horizontal overflow.

## v1.31.23

### Features
- **Vendor/model suggestions**: Device create/edit form now shows vendor and model suggestions filtered by device type, drawn from a bundled `vendors.json` catalog that can be manually updated.
- **Scan Agents tab in Discovery**: Discovery menu now has a dedicated "Scan Agents" tab, replacing the link in Settings → Tools.

### Bug Fixes
- **IP creation "no rows in result set"**: Creating an IP address returned an error even though the record was saved. Root cause: PostgreSQL CTE snapshot visibility — the outer SELECT in `WITH ins AS (INSERT…RETURNING id) SELECT…` could not see the newly inserted row. Fixed by splitting into two queries: INSERT RETURNING id, then SELECT by that id.
- **Utilization history failed to load**: The `GetSubnetUtilisationHistory` handler used the wrong permission string (`"subnets:read"` instead of `"ipam:subnet:read"`), causing all non-admin users to get a 403. Fixed.
- **Subnets breadcrumb navigated to `networks/undefined/subnets`**: Frontend used `subnet.sectionId` but the API returns `networkId` after camelCase normalization. Fixed to use `subnet.networkId`.
- **Devices and Requests failed to load / internal server error on device create**: Migration `20260609_003` renamed `sections` → `networks` but missed `devices.section_id` and `subnet_requests.section_id`. Added migration `20260609_004` to rename those columns and update related indexes.
- **Discovery auto-adds IPs as "Available"**: Discovered IPs are active on the network; status corrected to `"assigned"` in both local-scan and remote-agent paths.
- **Device type selector showed slug and name**: Type dropdown was rendering `{icon} {name}` (which included the slug icon); now shows name only.

### Changes
- Settings → Tools: Removed "Scan Jobs" and "Scan Profiles" entries (duplicated the Discovery menu). "Scan Agents" moved to the Discovery tab.
- Settings → Tools: Split "Reports & Authentication" into separate "Reports" and "Authentication" sections.
- Backups: Export filenames renamed to `padduck_export.csv`, `padduck_export.json`, and `padduck_v2_migration_bundle.zip`.

### Database Migrations
- `20260609_004_rename_remaining_section_columns`: Renames `devices.section_id` → `devices.network_id` and `subnet_requests.section_id` → `subnet_requests.network_id`; updates related indexes.

## v1.31.22

### Bug Fixes
- **Scan auto-add: IP address displayed with /32**: IP addresses were selected from the database using `address::text`, which includes the CIDR prefix for host addresses (e.g. `192.168.1.5/32`). Changed to `host(ip.address)` — matching the existing pattern in the subnets repository — to return the bare address.
- **Scan auto-add: dns_name not populated**: After auto-adding a discovered IP, `SetIPAddressPTRFromScan` was not called on the new record, so `dns_name` was never set even when a PTR record was resolved. Now called immediately after a successful auto-add (requires "Discover reverse DNS" enabled on the job and PTR records present in DNS).

## v1.31.21

### Bug Fixes
- **Dark mode user menu hover**: User menu dropdown items used `dark:hover:bg-gray-700` (bright gray) against a navy background forced by the global dark mode override, producing a blown-out highlight. Changed hover to `#0d2848` (pd-700) to match the sidebar.

## v1.31.20

### Bug Fixes
- **Scan auto-add IPs**: Fixed auto-discovered IPs never being saved — the insert used status `"active"`, which violates the DB `CHECK` constraint; corrected to `"available"`. Affects both local-scan and remote-agent paths.
- **Dark mode list selection**: Selected and hovered list rows in Scan Jobs and Admin Roles pages were unreadable in dark mode (white text on pale-blue background); added `dark:bg-blue-900/20` to fix contrast.

## v1.31.19

### Bug Fixes
- **Frontend healthcheck**: Fixed the compose healthcheck using `wget`, which is not available in the `nginx:1.31.1-trixie` (Debian-based) image; switched to `curl`.

## v1.31.18

### Features
- **Scan job improvements**: Run Now button stays disabled until the scan finishes; scan type, concurrency, schedule, auto-add IPs, discover DNS, and overwrite DNS options are now configured directly on the job in a unified Settings tab — separate Scan Profiles removed.
- **Auto-add IPs from scans**: Active IPs discovered during a scan are automatically added to the appropriate subnet when the job's auto-add option is enabled.
- **Per-job DNS options**: New `discover_dns` and `dns_overwrite` flags on scan jobs control whether PTR records are looked up and whether existing `dns_name` values are overwritten.
- **Scan results enhancements**: Added a "Hide down" toggle to filter non-alive IPs; IPs that were alive in the previous scan but are now gone display an amber "Gone" badge.
- **Scan Retention fix**: The retention settings tab no longer errors on a fresh installation — defaults are inserted automatically if no settings row exists.

### Changes
- Renamed "Sections" to "Networks" throughout the application (UI labels, API routes, database table, and all code references). Migration: `20260609_003_rename_sections_to_networks`.
- Fixed topology view showing a double CIDR (e.g. `10.0.0.0/32/24`) by using PostgreSQL `host()` instead of `::text` cast on INET columns.

### Bug Fixes
- **Subnet split**: The original subnet is now deleted after a split instead of being kept as a container, preventing overlapping address space. IPs are moved to the correct child subnet during the transaction.
- **Subnet split blocking**: If any existing IP falls on a network or broadcast address of a child subnet, the split is blocked and the conflicting IPs are shown to the user.

### Database Migrations
- `20260609_002_scan_job_dns_options`: Adds `discover_dns` (default `true`) and `dns_overwrite` (default `false`) columns to `scan_jobs`.
- `20260609_003_rename_sections_to_networks`: Renames the `sections` table to `networks` and the `section_id` column to `network_id`.

## v1.31.17

### Build
- Updated frontend runtime image from `nginx:1.28.3-alpine` to `nginx:1.31.1-trixie` (Debian-based).
- Replaced `apk upgrade` with `apt-get upgrade` to pull latest Debian security patches.
- Switched healthcheck from `wget` to `curl` to match the Debian nginx image toolset.

## v1.31.16

### Build
- Updated frontend builder image from `node:22.16-alpine` to `node:22.22.3-alpine`.
- Updated frontend runtime image from `nginx:1.28-alpine` to `nginx:1.28.3-alpine`.
- Added `apk upgrade` to the nginx stage to pull in latest Alpine package security patches (nginx, libcrypto3, libssl3, curl, musl, nghttp2-libs, zlib, xz-libs, tiff, libpng, libxml2, and others).

## v1.31.15

### Build
- Updated backend Dockerfile base image from `golang:1.26.3-alpine` to `golang:1.26.4-alpine` to match the go directive.

## v1.31.14

### Dependencies
- Updated frontend npm packages: axios 1.16.1→1.17.0, cytoscape 3.33.3→3.34.0, react/react-dom 19.2.6→19.2.7, react-router-dom 7.15.1→7.17.0, vite 8.0.13→8.0.16, vitest 4.1.6→4.1.8.
- Added `eslint:recommended` to the flat ESLint config; added vitest globals to test file config; fixed undeclared `historyError` state in the utilization history section.
- Updated Go backend dependencies: jackc/pgx/v5, x/crypto, x/net, x/sys, x/text, x/sync, go-sqlite3, go-runewidth, go-colorable, go-internal.
- Bumped the `go` directive to 1.26.4 in both the backend and agent modules.

## v1.31.13

### Authentication
- Added a **Change Password** option under User → Security settings, allowing users to update their password without admin intervention.

### Network / IP Addresses
- The New IP Address form now pre-fills the network portion of the IP field based on the selected subnet, reducing manual entry errors.
- Fixed IP addresses not appearing in subnet list views due to a field key mismatch and an incomplete paginated query.
- Fixed an encode error in the split-subnet IP move query.

### Configuration
- Added `anonymous_api_enabled` to the allowed configuration keys and included it in the seed migration so it is recognized on fresh installs.

## v1.31.12

### Authentication
- The login page now hides the self-registration link when self-registration is disabled.
- Public instance metadata now exposes the registration-enabled state so unauthenticated pages can reflect the current setting without exposing admin configuration.

### Release / CI
- Main-branch release automation now requires the top changelog entry to increment the latest release by exactly one patch version, creates the matching tag, and publishes versioned release images.
- Frontend CI and Docker builds now use npm 11.16.0.

## v1.31.11

### Installation
- Updated the Docker Compose PostgreSQL bind mount for `postgres:18` from `/var/lib/postgresql/data` to `/var/lib/postgresql`, matching the official PostgreSQL 18 image layout.
- Existing deployments that already started with the previous mount should verify where their database files were initialized before changing mounts on a live system.

## v1.31.10

### Installation
- Docker Compose installs now require only `docker-compose.yml` by default. The backend and frontend images are pulled from GitHub Container Registry, and `.env` is only needed for overrides.
- Updated the README and getting-started guide to document the compose-only install path.
- `.env.example` now keeps built-in Compose defaults commented out so local config files can stay focused on explicit overrides.

### MFA
- Production deployments without `MFA_ENCRYPTION_KEY` now create and reuse a persistent backend-managed key at `./data/backend/mfa-encryption-key`.
- Migration readiness checks now accept either an explicit `MFA_ENCRYPTION_KEY` or the backend-managed persistent key file.
- Troubleshooting and user documentation now explain how to preserve or restore the MFA key.

### Security / CI
- Scoped MFA key file access and backup data-file reads with Go's root-scoped filesystem API.
- Updated CI to Go 1.26.4 so `govulncheck` runs against the fixed standard library.

## v1.31.9

### Backups
- Added a unified **Backups** page (`/admin/backups`) that consolidates Data Export, Data Import, and the new complete system backup.
- New **Download Complete Backup** produces a ZIP archive containing the full PostgreSQL database dump, all admin configuration settings (as JSON), and any files stored in `./data/` (avatars, agent binary, etc.).
- New **Restore from Backup** — upload a backup ZIP to restore the database, configuration, and files. Includes a two-step confirmation to prevent accidental data loss.
- The existing CSV/JSON data export and CSV import (subnets, IP addresses, phpIPAM) are now embedded in the Backups page as sub-sections.
- "Backups" sidebar link added under Admin; the Admin Overview now shows a single Backups card instead of separate Export and Import cards.

### DNS Integration
- **Auto-add IPs**: when enabled (`dns_auto_add_ips_enabled`), the DNS sync picks up A/AAAA records from the configured DNS provider and inserts any IP not already in IPAM into the matching subnet.
- **Auto-remove IPs**: when enabled (`dns_auto_remove_ips_enabled`), IPs previously added by the DNS sync that are no longer present in DNS records are removed from IPAM.
- Both options are toggleable in Admin → Settings → DNS. They default to disabled.

### Discovery / Scan Jobs
- **Auto-add discovered IPs**: new per-job toggle (enabled by default) — when a scan finds a live IP address that is not yet in IPAM, it is automatically added to the matching subnet.
- The toggle is exposed in the scan job creation and edit form.

### SNMP
- Added a **show/hide toggle** (eye icon) next to the global SNMP community string field in Admin → Settings so the value can be revealed without having to clear and retype it.

### Audit Log
- Fixed the **Prune** button on the Audit Retention page — the `POST /api/v1/admin/audit/prune` route was not registered; it now correctly calls the prune handler.
- Fixed the **export cap**: the repository previously reset any limit above 1,000 to 100, meaning CSV exports silently returned only 100 rows. The cap is now 100,000.
- Fixed the **Save** button on the Audit Retention page — if the settings row did not exist yet, the UPDATE matched nothing and returned an error; the handler now upserts the row before updating.
- Changed the minimum retention period from 1 to 30 days in the UI to match the database constraint.

### Nameservers
- Removed the hard-coded admin-only block from the Nameservers page. Non-admin users who have been granted nameserver permissions via RBAC can now view the list. Write controls (Add, Edit, Delete) remain admin-only.
- A clear "access denied" message is shown if the API returns 403 instead of a generic error.

### Scan Agents
- New **Download latest scan agent binary** section at the top of the Scan Agents admin page with one-click links for Linux x64/ARM64, macOS x64/ARM64, and Windows x64.
- Agent privilege model documented in `agent/PRIVILEGES.md` — the agent uses the system `ping` binary (no raw sockets), so elevated privileges are not required for the agent process itself.

### CI / Release
- GitHub Actions release workflow now builds the scan agent for all five platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) and attaches the binaries as assets to the GitHub Release.

## v1.31.3

### Navigation
- Removed Admin Overview sidebar item.
- Consolidated Users, Roles, and Permission Presets into a single "Users & Roles" page with tabs. Old routes redirect automatically.

### Admin Settings
- "Test PowerDNS Connection" button moved inside the PowerDNS section card.
- Feature toggle rows now render correctly in dark mode (hover no longer obscures text).
- Feature toggles that are absent from the database now default to enabled, preventing a save from accidentally disabling them.

### System Health
- Removed Quick Links section.
- Removed "Requires pg_dump to be installed in the backend container" note — pg_dump is now included in the backend Docker image.

### Backup
- `postgresql17-client` added to the backend Docker image so pg_dump works out of the box.

### Discovery
- Scan job creation now accepts a network address / CIDR (e.g. `192.168.1.0/24`) and resolves it to the matching subnet automatically. No more "at least one subnet id is required" error.

### Bug Fixes
- Audit retention settings no longer return a 500 error when the default row doesn't exist yet — the row is now created reliably without relying on `RETURNING` from a no-op insert.
- Subnet IP addresses page: removed Data Quality section. Utilization History no longer shows an error on empty data — it shows a friendly empty state instead.

## v1.31.2

### UI / UX
- Added version number to the bottom of the sidebar navigation (admin users only).
- Features settings page now reloads back to the Features tab after saving.
- FeatureGate "feature disabled" message now includes a direct link to Admin → Settings → Features.
- Role Management NavLink no longer incorrectly highlights when viewing Permission Presets.
- Break Glass moved from a standalone sidebar link to a tab on the Admin Users page.
- Audit retention settings are now auto-initialized on fresh installs (no more 500 error on first visit).
- Discovery sidebar entry consolidated into a single tabbed page (Scan Jobs, Scan Profiles, Scan Retention, Topology Hints, Conflicts).
- Integration Templates removed from sidebar navigation.
- Privacy Consent removed from sidebar; a Privacy Policy link added to the user menu above Logout.
- System Health: Migrations section removed; Scan Agents card shows a friendly message when no agents are registered.
- Backup & Restore: renamed from "Rehearsal"; added a Download Backup (.sql) button that streams a pg_dump.
- Modal dialogs no longer steal focus back to the modal container on every re-render; text inputs retain focus correctly.

## v1.31.1

- Fixed admin password file writing to `data/admin-password` in the working directory instead of `/run/ipam`, which required root filesystem access.
- Added `entrypoint.sh` to fix bind-mount ownership at startup; backend container now drops to an unprivileged user via `su-exec`.
- Fixed gosec `#nosec` annotation referencing wrong rule (`G306` → `G703`).
- Fixed `feature_firewall_enabled` missing from the allowed config keys, causing a "unknown config key" error on save.

## v1.31.0

- Added firewall zones and firewall zone mappings.
- Updated frontend assets: new logo, favicon set, and web app manifest icons.

## v1.30.0 Padduck Rebrand and GitHub Migration

- Renamed project to Padduck across all surfaces: Go module, frontend metadata, storage keys, API artifacts, docs, and deployment config.
- Migrated CI/CD from Gitea Actions to GitHub Actions.
- Docker images now built and published to GitHub Container Registry (`ghcr.io/lima3w/padduck-backend`, `ghcr.io/lima3w/padduck-frontend`).
- `docker-compose.yml` updated to pull images from ghcr.io instead of building locally.
- Added docs: `index.md`, `getting-started.md`, `troubleshooting.md`.

## v1.26.0 API And SDK Stabilization

- Froze the stable v1 OpenAPI contract at `1.26.0`.
- Added retry-safe idempotency keys for automation write endpoints.
- Standardized validation error responses with field-level details.
- Versioned outbound webhook event payloads and added a sample payload endpoint.
- Added generated API client examples for JavaScript and Python.
- Added OpenAPI contract tests and changelog automation.

<!-- api-contract:1.26.0 -->

API contract snapshot:

- OpenAPI version: `1.26.0`
- Public API path count: `194`
- OpenAPI SHA-256: `12b186d567d9`
