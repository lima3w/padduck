# Padduck Refactor Notes

**Audit date:** 2026-06-13  
**Implementation date:** 2026-06-13  
**Auditor:** Claude Code (automated)  
**Branch:** chore/refactor-cleanup (off main @ 32f19b8)

---

## Implementation Summary

All actionable issues from the audit have been implemented on branch `chore/refactor-cleanup` targeting v1.31.32. Final state: `go build`, `go vet`, `go test ./...` all clean; `npm run build` clean.

| Issue | Title | Status |
|---|---|---|
| #126 | PHPIpam status values violate DB constraint | ✅ Fixed |
| #127 | BulkDeleteIPs N+1 DELETE | ✅ Fixed — single `DELETE … WHERE id = ANY($1)` |
| #128 | Dead helper functions + `var _` sentinels | ✅ Removed |
| #129 | Inline admin checks bypassing token scope | ✅ Fixed — consolidated to `requireAdmin` |
| #130 | Handler errors inconsistent (714 inline fiber.Map) | ✅ Fixed — all use `RespondError` |
| #131 | matchesCron/validateCron in wrong file | ✅ Moved to `services/cron.go` |
| #132 | 52 `log.Printf` in services | ✅ Migrated to `slog` |
| #133 | Redundant `loggingMiddleware` | ✅ Removed |
| #134 | scanDevice/scanInterface accept concrete pgx.Row | ✅ Fixed — use scanner interface |
| #135 | fmt.Errorf %w in 5 packages | ✅ Closed — already compliant |
| #136 | Dead `ListDevices` wrappers | ✅ Removed |
| #137 | Magic numbers in discovery.go | ✅ Extracted to named constants |
| #138 | repository/devices.go (845 lines) | ✅ Split into 6 files |
| #139 | repository/reports.go (780 lines) | ✅ Split into 4 files |
| #140 | handlers/external_auth.go (617 lines) | ✅ Split into 4 files |
| #141 | IPAddressesPage.jsx (1,486 lines) | ✅ Split — 4 sub-components extracted |
| #142 | UserSettingsPage.jsx (1,187 lines) | ✅ Split — 7 tab components extracted |
| #143 | DeviceDetailPage.jsx (1,084 lines) | ✅ Split — 7 panel components extracted |
| #144 | SubnetsPage.jsx (1,032 lines) | ✅ Split — 4 modal components extracted |
| #145 | console.error in ErrorBoundary.jsx | ✅ Closed — already DEV-guarded |
| #146 | American English in API/DB schema | ⏳ Deferred to v2.0.0 milestone |
| #147 | Full fmt.Errorf %w sweep | ✅ Closed — codebase already compliant |
| #148 | Token scope security decision | ✅ Decided (Option A) — implemented via #129 |

## Deferred / Out of Scope

- **#122** (N+1 in export functions), **#123** (N+1 in utilisation snapshots) — filed, not implemented in this patch; require new repository methods
- **#146** (American English in public API/DB) — v2.0.0 breaking change
- **#129 partial** — "self OR admin" patterns in `requests.go` and `users.go GetUser` intentionally left unchanged; they are ownership checks, not admin gates

---

## Baseline Validation Results

| Check | Result |
|-------|--------|
| `go build ./...` | PASS (no output) |
| `go vet ./...` | PASS (no output) |
| `go test ./...` | PASS (all 14 packages, no failures) |
| `npm run build` | PASS (built in 1.55 s) |

---

## Already-Filed Issues (not duplicated here)

- **#122** — N+1 query in export functions (`import.go`)
- **#123** — `TakeUtilisationSnapshots` N+1 + fetches full rows to count
- **#124** — LDAP `SyncGroups` silently drops error
- **#125** — LDAP `AssignRoleToUser` ignores non-duplicate errors

---

## Findings

### Go — Bugs / Correctness

#### F1: `phpIpamStateToStatus` returns status values that violate the DB constraint
**File:** `backend/services/import.go:635-647`  
Returns `"active"`, `"dhcp"`, and `"inactive"` but the `ip_addresses.status` column has a CHECK constraint `IN ('available', 'assigned', 'reserved')`. Any PHPIpam import will silently fail at the DB layer for all rows whose state is 1/used/active/3/dhcp or the default. Should map to `"assigned"` / `"reserved"` / `"available"`.

#### F2: `BulkDeleteIPs` issues one DELETE per IP — N+1 in handler layer
**File:** `backend/handlers/reports.go:467-493`  
Iterates `req.IPIDs` and calls `h.service.DeleteIPAddress(ctx, id)` in a loop. For large batches this is O(n) round-trips. A single `DELETE FROM ip_addresses WHERE id = ANY($1)` would suffice. Contrast with `BulkReleaseIPs` which already delegates to a batch-capable repository method.

#### F3: Dead helper functions kept alive only by `var _ =` sentinel
**File:** `backend/services/reports.go:871-892`  
Three unexported functions — `cidrFromSubnet`, `parseIPNet`, `subnetNetworkName` — are unreachable by any production path. They are only kept from the compiler by:
```go
var _ = cidrFromSubnet
var _ = parseIPNet
var _ = subnetNetworkName
```
`subnetNetworkName` is a no-op placeholder ("returns the section name by extracting it from a join. Placeholder."). Also `models.UserMFASettings{}` is held alive in `backend/services/mfa.go:349` by the same pattern, though `UserMFASettings` is used by the repository; the sentinel is likely stale. All four sentinels should be removed and the placeholder function bodies filled in or deleted.

---

### Go — Inconsistent Patterns

#### F4: 41 handlers use inline `admin.Role != "admin"` checks instead of `requireAdmin`
**Files:** `backend/handlers/admin_features.go`, `audit.go`, `approvals.go`, `config.go`, `notifications.go`, `requests.go`, `roles.go`, `subnets.go`, `tags.go`, `updates.go`, `users.go`  
`requireAdmin` is defined in `tags.go` and used only in that file. 41 other handlers in 10 files replicate `c.Locals("user").(*models.User)` + `admin.Role != "admin"` inline and produce raw `fiber.Map{"error": "admin access required"}` responses instead of using `RespondError`. This bypasses the structured logging in `RespondError` and creates a maintenance hazard if the admin-check logic ever changes.

#### F5: ~714 handlers return raw `fiber.Map{"error": ...}` instead of `RespondError`
**Files:** All handler files  
`RespondError` / `h.StatusBadRequest` / `h.StatusNotFound` etc. exist and add structured logging + consistent JSON shape (`{"error":..., "code":..., "details":...}`). But 714 call sites across 47 handler files still use `c.Status(...).JSON(fiber.Map{"error": "..."})` directly, producing responses without an error `code` field. This makes client-side error handling harder and loses structured log context.

#### F6: `matchesCron` / `validateCron` are defined in `discovery.go` but used by `reports.go`
**File:** `backend/services/discovery.go:587-621`, used in `backend/services/reports.go:717`  
Both functions are package-private to `services` so this compiles fine, but semantically they belong to a shared cron utility. If `discovery.go` is ever split out, the dependency is invisible. Consider moving them to a dedicated `cron.go` file in the `services` package.

#### F7: Logging is split between `log.Printf` and `slog`
**Scope:** `backend/services/` (58 uses of `log.Printf`), `backend/handlers/handlers.go:654`  
`main.go` sets `slog` as the default logger and bridges `log.Printf` through it. However services (`reports.go`, `notification.go`, `registration.go`, `requests.go`, `dns.go`, `audit.go`) still emit `log.Printf` directly, missing structured key-value fields. `handlers.go:654` also uses `log.Printf("%s %s", method, path)` inside `loggingMiddleware`, a pattern that duplicates what the Fiber logger middleware in `main.go` already logs.

#### F8: `loggingMiddleware` in `handlers.go` duplicates the Fiber logger registered in `main.go`
**Files:** `backend/handlers/handlers.go:649-656`, `backend/main.go:235-242`  
Both middlewares skip `/health` and log the request method + path. The Fiber logger already captures method, path, status, latency, and client IP. `loggingMiddleware` only adds method + path via `log.Printf`. The handler-level middleware is redundant noise and should be removed.

#### F9: Scanner `devices.go` uses `pgx.Row` while all other scan helpers use `interface{ Scan(...) error }`
**File:** `backend/repository/devices.go:85,411`  
`scanDevice` and `scanInterface` accept `pgx.Row` (a concrete type) while every other repository scan helper uses `interface{ Scan(dest ...any) error }`. The concrete type prevents them from being called on `pgx.Rows.Scan(...)` without an adapter, making them less flexible. This is inconsistent with the pattern established in `racks.go`, `locations.go`, `scan.go`, etc.

#### F10: Inconsistent error wrapping — 1,742 `fmt.Errorf` without `%w` vs 487 with `%w`
**Scope:** entire backend  
Most service-layer and repository errors wrap with `fmt.Errorf("...: %w", err)`, allowing `errors.Is`/`errors.As` to traverse the chain. But 1,742 call sites (78%) omit `%w`, producing opaque strings. Systematic callers (e.g. handlers checking `errors.Is(err, pgx.ErrNoRows)`) would silently not match. Priority targets: `backend/internal/netguard/` (all 9 Errorf calls), `backend/repository/nameservers.go`, `backend/repository/racks.go`, `backend/repository/tags.go`, `backend/services/autonomous_systems.go`.

---

### Go — Dead / Unused Code

#### F11: `Service.ListDevices` is a trivial wrapper that is never called
**File:** `backend/services/devices.go:23-28`, `backend/repository/devices.go:100-102`  
`ListDevices(ctx, limit, offset)` delegates to `ListDevicesWithOptions`. The handler (`handlers/devices.go`) calls `ListDevicesWithOptions` or `ListAllDevices` directly; `ListDevices` is never called except by tests that also test the wrapper. The repository-level wrapper (`r.ListDevices → r.ListDevicesWithOptions`) is similarly dead. They add a thin layer of confusion.

#### F12: Magic numbers for default scan concurrency, port limits, and agent timeout
**File:** `backend/services/discovery.go:159,362,514,522,530,706`  
The value `20` appears three times as the default ping concurrency (lines 159, 362, 493), `1000` and `100` as result-limit bounds (514-523), `50` as scan-run history limit (530), and `15 * time.Minute` as the agent stale threshold (706). None are named constants. The same value `20` also appears in `services/devices.go:26,33`. Defining `defaultPingConcurrency`, `defaultResultLimit`, `maxResultLimit`, `scanRunHistoryLimit`, and `agentStaleThreshold` would make the intent clear.

#### F13: `backend/services/reports.go` unused import sentinel `var _ = strings.TrimSpace`
**File:** `backend/services/reports.go:889`  
`strings` is imported but only used via this blank-identifier sentinel. Removing the dead helper functions (F3) and the sentinel would eliminate the import entirely.

---

### Go — Large Files

#### F14: `backend/repository/devices.go` is 845 lines combining five distinct concerns
**File:** `backend/repository/devices.go`  
The file contains: device types, device CRUD, device SNMP credentials, device–IP associations, and device interfaces. Each is a coherent sub-domain that could live in `repository/device_types.go`, `repository/device_interfaces.go`, etc.

#### F15: `backend/repository/reports.go` is 780 lines combining six distinct concerns
**File:** `backend/repository/reports.go`  
Contains: utilization snapshots, alert cooldowns, scheduled reports, inactive IPs, duplicates, and subnet gap/VLAN/DNS audit queries. Could be split into `repository/utilization.go`, `repository/scheduled_reports.go`, `repository/report_queries.go`.

#### F16: `backend/handlers/external_auth.go` is 617 lines handling LDAP, OAuth2, and SAML
**File:** `backend/handlers/external_auth.go`  
Three independent authentication providers with separate config, test, and flow handlers crammed into one file. Should be split into `handlers/external_auth_ldap.go`, `handlers/external_auth_oauth2.go`, `handlers/external_auth_saml.go`.

---

### Frontend — Bugs / Correctness

#### F17: British spellings baked into public API JSON field names and database column names
**Files:** `backend/models/models.go` (`colour`, `utilisation_pct`), `backend/repository/tags.go`, `backend/services/tags.go`, `backend/services/vlans.go`, `frontend/src/pages/AdminTagsPage.jsx`, `frontend/src/components/TagBadge.jsx`  
The JSON field `colour` (used in the Tags API response and in the `ip_tags` database column) and `utilisation_pct` (on subnet models) are British spellings. These are now part of the public API contract and in the database schema, so renaming them requires a migration and a versioned API change. This is a **deferred/high-risk** item — filed for awareness rather than immediate action.

---

### Frontend — Large Components

#### F18: `IPAddressesPage.jsx` is 1,486 lines with six embedded sub-components
**File:** `frontend/src/pages/IPAddressesPage.jsx`  
Contains `DelegationsTab`, `PortBadges`, `SortTh`, `UtilisationHistorySection`, `loadColumnVisibility`, and the 1,159-line `IPAddressesPage` itself. Each sub-component can be extracted to its own file or to `components/`.

#### F19: `UserSettingsPage.jsx` is 1,187 lines with seven tab components embedded
**File:** `frontend/src/pages/UserSettingsPage.jsx`  
Contains `ProfileTab`, `SecurityTab`, `TokensTab`, `LoginHistoryTab`, `NotificationsTab`, `SessionsTab`, `PrivacyTab` — all defined in the same file as the parent page. Following the pattern already established for `admin/AdminSettingsPage.jsx` (which delegates to separate tab files in `pages/admin/`), each tab could live in `pages/user/`.

#### F20: `DeviceDetailPage.jsx` is 1,084 lines as a single monolithic component
**File:** `frontend/src/pages/DeviceDetailPage.jsx`  
No internal sub-components are extracted. The page renders IP associations, interfaces, SNMP credentials, scan results, fingerprints, and change history in a single render function.

#### F21: `SubnetsPage.jsx` is 1,032 lines as a near-monolithic component
**File:** `frontend/src/pages/SubnetsPage.jsx`  
A single `SubnetsPage` default export handles the full subnet list, split/merge/resize modals, scan profile assignment, and VLAN assignment. Only one helper function (`splitCidrPreview`) is extracted.

#### F22: `ScanJobsPage.jsx` and `AdminUsersPage.jsx` each exceed 800 lines
**Files:** `frontend/src/pages/ScanJobsPage.jsx` (816 lines), `frontend/src/pages/AdminUsersPage.jsx` (986 lines)  
`AdminUsersPage.jsx` contains a full break-glass management panel (`BreakGlassTab`) and user management in one file. `ScanJobsPage.jsx` embeds the full scan-job creation/edit form in a single component.

---

### Frontend — Code Quality

#### F23: `console.error` left in production code in `ErrorBoundary.jsx`
**File:** `frontend/src/components/ErrorBoundary.jsx:142,183`  
Two `console.error` calls in the error boundary catch handlers will appear in production browser consoles. These should use a proper error-reporting hook (or at minimum be guarded by `process.env.NODE_ENV !== 'production'`).

---

## Issues Created

See GitHub for full issue bodies. Issues filed below (numbers assigned after creation):

| Issue # | Title | Label |
|---------|-------|-------|
| [#126](https://github.com/lima3w/padduck/issues/126) | bug: phpIpamStateToStatus returns status values that violate the DB constraint | bug |
| [#127](https://github.com/lima3w/padduck/issues/127) | refactor: BulkDeleteIPs issues one DELETE per IP (N+1 in handler layer) | enhancement |
| [#128](https://github.com/lima3w/padduck/issues/128) | cleanup: remove dead helper functions kept alive by var _ = sentinels in reports.go and mfa.go | enhancement |
| [#129](https://github.com/lima3w/padduck/issues/129) | refactor: consolidate inline admin role checks to use requireAdmin helper | enhancement |
| [#130](https://github.com/lima3w/padduck/issues/130) | refactor: migrate handler error responses from raw fiber.Map to RespondError | enhancement |
| [#131](https://github.com/lima3w/padduck/issues/131) | refactor: move matchesCron/validateCron to a shared cron.go in services package | enhancement |
| [#132](https://github.com/lima3w/padduck/issues/132) | refactor: migrate service-layer log.Printf calls to slog | enhancement |
| [#133](https://github.com/lima3w/padduck/issues/133) | cleanup: remove redundant loggingMiddleware (duplicates Fiber logger) | enhancement |
| [#134](https://github.com/lima3w/padduck/issues/134) | refactor: fix scanDevice/scanInterface to accept scanner interface instead of pgx.Row | enhancement |
| [#135](https://github.com/lima3w/padduck/issues/135) | refactor: add %w to fmt.Errorf calls in netguard, nameservers, racks, tags, autonomous_systems | enhancement |
| [#136](https://github.com/lima3w/padduck/issues/136) | cleanup: remove dead Service.ListDevices wrapper | enhancement |
| [#137](https://github.com/lima3w/padduck/issues/137) | refactor: extract named constants for magic numbers in discovery.go | enhancement |
| [#138](https://github.com/lima3w/padduck/issues/138) | refactor: split repository/devices.go (845 lines) into sub-files by concern | enhancement |
| [#139](https://github.com/lima3w/padduck/issues/139) | refactor: split repository/reports.go (780 lines) into sub-files by concern | enhancement |
| [#140](https://github.com/lima3w/padduck/issues/140) | refactor: split handlers/external_auth.go (617 lines) by provider | enhancement |
| [#141](https://github.com/lima3w/padduck/issues/141) | refactor: split IPAddressesPage.jsx (1,486 lines) — extract sub-components | enhancement |
| [#142](https://github.com/lima3w/padduck/issues/142) | refactor: split UserSettingsPage.jsx (1,187 lines) — extract tab components | enhancement |
| [#143](https://github.com/lima3w/padduck/issues/143) | refactor: split DeviceDetailPage.jsx (1,084 lines) into panel components | enhancement |
| [#144](https://github.com/lima3w/padduck/issues/144) | refactor: split SubnetsPage.jsx (1,032 lines) — extract modal and form components | enhancement |
| [#145](https://github.com/lima3w/padduck/issues/145) | cleanup: remove console.error from ErrorBoundary.jsx production code | enhancement |

---

## Deferred / High-Risk Items (human review required before filing)

1. **British spellings in public API (`colour`, `utilisation_pct`)** — These are now wire-format field names in the JSON API and PostgreSQL column names. Renaming them requires a breaking API change + a migration. The team should decide whether to rename them in a major version bump (v2) or keep them as-is for backward compatibility. *Not filed as an actionable issue.*

2. **`fmt.Errorf` without `%w` — 1,742 call sites** — This is too large a surface to fix in one issue. The filed issue (#135) scopes it to the highest-priority files where callers already do `errors.Is` checks. The rest should be handled incrementally by subsystem.

3. **`requireAdmin` scope** — The helper in `tags.go` also checks `tokenScope`. The 41 inline checks in other files do NOT check `tokenScope`. This may be intentional (the other handlers are already behind RBAC), but it warrants a deliberate audit before consolidation.
