# IPAM Next Roadmap

This roadmap assumes the v1.19.0 cleanup milestone has shipped. At that point
the product has closed the known frontend/backend gaps around password reset,
sessions, notification preferences, VRF/rack management, lease workflows, DNS
IPv6/Technitium parity, SNMP configuration, admin user lifecycle actions,
custom roles, scheduled reports, privacy consent, and integration health.

The goal for the remaining v1 line is to make IPAM Next dependable for daily
operations without forcing a breaking architecture change. The goal for v2 is
to turn that stable product into a larger network source-of-truth platform.

## v1 Line

### v1.20.0 Operations Polish

Focus: make the completed v1.19 feature set easier to operate at scale.

- Add saved filters and column preferences for IPs, devices, racks, VLANs,
  VRFs, users, audit logs, scan jobs, and reports.
- Add bulk actions for common admin workflows: assign tags, change status,
  release leases, update DNS names, move devices between racks, and deactivate
  stale users.
- Add dashboard widgets for lease expiry, DNS drift, scan failures, webhook
  failures, expiring API tokens, and near-capacity subnets.
- Add CSV export parity for every major table view.
- Add consistent empty, loading, error, and permission-denied states across
  all admin and operator pages.
- Add keyboard-friendly command/search entry points for common navigation and
  object lookup.

### v1.21.0 Data Quality And Reconciliation

Focus: help operators trust the database.

- Add a reconciliation center for DNS, discovery, leases, device inventory,
  and stale assignments.
- Add duplicate detection for DNS names, MAC addresses, device hostnames, and
  conflicting IP ownership.
- Add import preview and dry-run validation for sections, subnets, IPs, VLANs,
  devices, racks, users, and custom fields.
- Add per-object change history views that summarize audit log entries in the
  context of the object being viewed.
- Add data quality scores for subnets and device inventory completeness.
- Add scheduled remediation reports for stale leases, unresolved DNS drift,
  unused reservations, failed scans, and inactive devices.

### v1.22.0 Discovery And Inventory Depth

Focus: improve the accuracy and usefulness of network discovery.

- Add device fingerprinting from scan results, SNMP system data, open ports,
  DNS records, and historical observations.
- Add discovery confidence levels and conflict review before overwriting
  operator-entered data.
- Add scan profiles with per-subnet overrides for ICMP, TCP port scan, DNS
  lookup, SNMPv2c, and SNMPv3 behavior.
- Add scan agent health, version, and last-contact reporting.
- Add topology hints by linking devices, interfaces, IPs, VLANs, racks, and
  locations.
- Add discovery retention policies and rollups so scan history stays useful
  without unbounded growth.

### v1.23.0 Automation And Integration Workflows

Focus: make integrations actionable rather than just connected.

- Add webhook retry controls, manual replay, failure grouping, and payload
  inspection.
- Add outbound event subscriptions by object type, event type, and tag/filter
  conditions.
- Add inbound automation endpoints for controlled allocation, reservation,
  release, DNS update, and device registration workflows.
- Add API token usage analytics and per-token rate-limit visibility.
- Add integration templates for common automation platforms and network tooling.
- Add a policy engine for simple approval and validation rules before changes
  are committed.

### v1.24.0 Security, Compliance, And Enterprise Readiness

Focus: harden the product for stricter environments.

- Add structured permission presets and permission-diff review for custom role
  changes.
- Add break-glass admin controls with stronger audit requirements.
- Add session risk signals, admin-enforced MFA policy, token expiry policy, and
  inactive-user policy controls.
- Add audit log integrity checks, retention policies, and archive exports.
- Add privacy-policy version history and consent reporting.
- Add backup status visibility, restore rehearsal documentation, and deployment
  health checks in the admin UI.

### v1.25.0 Scale And Performance

Focus: keep the v1 architecture fast and predictable as datasets grow.

- Add pagination, server-side sorting, and indexed filtering to all large list
  endpoints.
- Add background jobs for long-running imports, exports, scans, DNS checks,
  reports, and webhook replays.
- Add job progress, cancellation, retry, and failure diagnostics.
- Add database index review for subnet/IP search, audit filters, discovery
  results, leases, webhook deliveries, and report history.
- Add cache boundaries for expensive dashboard and reporting queries.
- Add performance budgets and regression tests for common operator workflows.

### v1.26.0 API And SDK Stabilization

Focus: make the v1 API predictable for external automation.

- Freeze public API contracts for stable resources and document compatibility
  guarantees.
- Add generated API client examples for shell, Python, and TypeScript.
- Add idempotency keys for write-heavy automation endpoints.
- Add consistent error codes and validation response shapes across the API.
- Add webhook event schema versioning and sample payloads.
- Add OpenAPI contract tests and changelog automation.

### v1.27.0 UX Consolidation

Focus: reduce page-by-page inconsistency before v2.

- Standardize table actions, detail drawers, destructive confirmations, filters,
  object links, timestamps, status badges, and audit references.
- Unify settings pages around account, security, notifications, tokens,
  sessions, login history, and privacy.
- Unify admin pages around users, roles, approvals, settings, integrations,
  audit, scans, reports, and system health.
- Add object relationship panels so sections, subnets, IPs, devices, racks,
  VLANs, VRFs, DNS records, scans, and reports cross-link consistently.
- Complete accessibility pass for keyboard navigation, focus states, contrast,
  labels, and modal behavior.

### v1.28.0 Pre-v2 Compatibility Release

Focus: prepare users and integrations for v2.

- Add compatibility warnings for APIs, fields, permissions, and workflows that
  will change in v2.
- Add export tooling that can produce a v2 migration bundle.
- Add migration readiness checks for schema, config, integrations, custom
  fields, roles, tokens, and webhook subscriptions.
- Add admin-visible deprecation reporting.
- Add a documented v1 long-term maintenance policy.

## v2 Line

### v2.0.0 Platform Architecture

Focus: introduce the breaking changes needed for a durable network
source-of-truth platform.

- Move from page-driven feature modules to domain modules with explicit
  contracts: identity, inventory, addressing, topology, discovery, DNS,
  automation, reporting, and compliance.
- Version the API under a new v2 namespace with stable resource schemas,
  cursor pagination, consistent errors, idempotency, and explicit partial
  update semantics.
- Introduce a background job system as a first-class backend primitive.
- Introduce a typed event bus for audit, webhooks, automation rules, scan
  results, DNS sync, reporting, and notifications.
- Add a formal migration path from v1 data and configuration.
- Keep a documented v1 compatibility mode where practical, but do not preserve
  weak v1 internals that block the v2 model.

### v2.1.0 Multi-Tenancy And Delegation

Focus: support managed-service and multi-organization use cases.

- Add tenants or organizations as a first-class isolation boundary.
- Add scoped roles and delegated administration by tenant, section, location,
  VRF, tag, or custom grouping.
- Add tenant-aware API tokens, webhook subscriptions, audit logs, reports, and
  dashboard views.
- Add cross-tenant operator views for platform admins.
- Add tenant-level quotas, retention policies, and integration settings.

### v2.2.0 Source Of Truth And Intent

Focus: distinguish desired state from observed state.

- Add an intent model for allocations, reservations, DNS records, device
  metadata, VLAN/VRF assignments, and rack placement.
- Track observed state from scans, SNMP, DNS, imports, and integrations
  separately from operator-approved state.
- Add drift review workflows with approve, ignore, apply, and rollback actions.
- Add policy validation before intent changes are accepted.
- Add change windows and scheduled activation for high-risk updates.

### v2.3.0 Topology And Relationships

Focus: model the network as a graph, not just lists.

- Add interfaces, links, circuits, device roles, platforms, sites, locations,
  racks, VLANs, VRFs, prefixes, IPs, DNS records, and services as connected
  objects.
- Add topology views for rack, site, VRF, subnet, VLAN, and device contexts.
- Add dependency analysis before deleting or modifying shared resources.
- Add relationship-aware search and impact analysis.
- Add import/export formats for graph-style inventory data.

### v2.4.0 Automation Engine

Focus: let teams build controlled workflows inside IPAM Next.

- Add a rule builder for conditions, approvals, actions, and notifications.
- Add built-in actions for allocation, reservation, release, DNS sync, webhook
  dispatch, report generation, scan execution, and user notification.
- Add dry-run and simulation support for automation rules.
- Add per-rule audit trails, execution history, and rollback hooks.
- Add secret handling for integration credentials used by automation actions.

### v2.5.0 High Availability And Large Deployments

Focus: support production environments with high availability expectations.

- Add horizontal worker support for scans, jobs, webhooks, reports, and imports.
- Add leader election or distributed locking for scheduled work.
- Add object storage support for exports, reports, backups, and large imports.
- Add database migration safety checks and online migration guidance.
- Add observability endpoints for metrics, traces, queue depth, job failures,
  webhook latency, scan throughput, and API latency.

### v2.6.0 Ecosystem And Extensibility

Focus: make IPAM Next easier to extend without patching core code.

- Add plugin-style integration points for discovery sources, DNS providers,
  notification channels, importers, exporters, and automation actions.
- Add a stable extension manifest and sandboxed execution model for supported
  extension types.
- Add integration certification tests.
- Add marketplace-style documentation for first-party and community extensions.
- Add compatibility checks so extensions declare supported v2 API and event
  schema versions.

## Release Principles

- Keep v1 releases backward-compatible unless a security issue requires a
  behavior change.
- Prefer completing operator workflows over adding new isolated endpoints.
- Treat scans, DNS sync, webhooks, reports, and imports as asynchronous work
  once they can run long enough to block user requests.
- Keep observed network data separate from operator-approved source-of-truth
  data as v2 approaches.
- Add migration checks before making breaking v2 schema or API changes.
