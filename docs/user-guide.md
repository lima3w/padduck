# Padduck — User Guide

The full user guide lives in the [project wiki](https://github.com/lima3w/padduck/wiki/User-Guide).

Topics covered there include: login and MFA, dashboard, networks, subnets (including resize / merge / split), IP addresses, VRFs and VLANs, discovery (scan agents, scan profiles, retention), topology view, devices, locations and racks, circuits, DNS zones, DHCP (standalone + Technitium integration), SSO (LDAP / OAuth2 / SAML), reports and scheduled reports, firewall zones, NAT rules, custom fields, audit logs, audit retention, requests and approvals, API tokens, webhooks, automation policies, Grafana data source, GDPR / data privacy, and the full API reference.

---

## Quick Reference

### Admin password (first boot)

The auto-generated password is printed to the server log and written to `./data/backend/admin-password` on the host (mounted at `/app/data/admin-password` inside the container). Set `ADMIN_PASSWORD` to override, or `RESET_ADMIN_PASSWORD=true` to force-reset.

### API authentication

```
Authorization: Bearer <token>
```

Generate tokens under **My Settings → API Tokens**.

### OpenAPI spec

```
GET /api/openapi.yaml
```
