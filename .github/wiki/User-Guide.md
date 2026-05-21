# User Guide

---

## Overview

Padduck is accessed through your browser at the URL your administrator provides (e.g., `https://ipam.example.com`). All features are available via the left sidebar navigation.

---

## Login

1. Navigate to your Padduck URL
2. Enter your username and password
3. If MFA is enabled: enter your TOTP code from your authenticator app
4. External auth (LDAP/OAuth2/SAML): click the provider button on the login page

---

## Dashboard

The dashboard provides a quick view of:
- Total subnets and IP addresses
- Utilization trends
- Recent activity
- System health status

Use **Ctrl+K** (or **Cmd+K** on Mac) to open global search from anywhere.

---

## Sections

Sections are top-level organizational containers for subnets.

| Action | How |
|--------|-----|
| List sections | Sidebar → **Sections** |
| Create section | Sections page → **+ New Section** |
| Edit / delete | Row → kebab (⋮) menu |

---

## Subnets

Subnets define CIDR blocks within a section.

| Action | How |
|--------|-----|
| List all subnets | Sidebar → **Subnets** |
| Create subnet | Section detail → **+ New Subnet** |
| View utilization | Subnet row shows `assigned / total` |
| Edit / delete | Subnet detail → kebab menu |

### Subnet Detail Page

Shows all IP addresses in the subnet with status, `assigned_to`, description, and tags. Utilization bar updates as addresses are allocated.

---

## IP Addresses

### Statuses

| Status | Meaning |
|--------|---------|
| `available` | Free to allocate |
| `assigned` | In active use |
| `reserved` | Held, not actively used |

### Allocate Next Available IP

1. Open a subnet
2. Click **Allocate**
3. Enter `Assigned to` (hostname, service name, or label)
4. Optionally set description, lease expiry, MAC address, tags
5. Click **Allocate**

Padduck assigns the lowest available host address.

### Reserve a Specific IP

1. Open a subnet
2. Click the IP address row
3. Click **Reserve** and fill in details

### Release an IP

From the IP detail page, click **Release** to return it to `available`.

### Bulk Operations

Select multiple IP addresses using checkboxes, then choose a bulk action (release, tag, export).

---

## VRFs

**Admin → VRFs** (or sidebar if enabled)

| Action | How |
|--------|-----|
| Create VRF | VRF list → **+ New VRF** |
| Assign subnet to VRF | Subnet create/edit → VRF field |

---

## VLANs

**Admin → VLANs**

VLANs have a numeric ID (1–4094), a name, and an optional VLAN domain. They can be associated with subnets.

---

## Discovery (Network Scanning)

### Create a Scan Job

1. **Admin → Scan Jobs → + New Scan Job**
2. Name the job, select target subnets
3. Optionally set a cron schedule (e.g., `0 2 * * *` for 2 AM daily)
4. Click **Create**

### Run a Scan

- **Run Now** to start immediately
- Scheduled jobs run automatically

### View Results

Results appear in the scan job detail page and on subnet pages. Live hosts are highlighted. Conflicts (addresses responding with no IPAM record) are flagged.

### Scan Conflicts

**Admin → Discovery → Conflicts** shows discovered addresses that conflict with IPAM records. Use **Resolve** to update records or dismiss false positives.

---

## Devices

Track physical and virtual devices associated with IP addresses.

| Field | Description |
|-------|-------------|
| Name | Device hostname or label |
| IP address | Primary IP (links to IPAM record) |
| Location | Rack/location context |
| Tags | Organizational metadata |

Access via **Devices** in the sidebar.

---

## Locations & Racks

Track physical locations (data centers, offices) and rack assignments for devices.

---

## DNS Zones

**DNS → Zones** tracks DNS zones as documentation. Nameservers can be added and zone health can be monitored.

---

## Searching

### Global Search

Press **Ctrl+K** / **Cmd+K** to open the command palette. Search across:
- Subnets (by CIDR or description)
- IP addresses (by address or `assigned_to`)
- Devices
- DNS records

Use **arrow keys** to navigate results, **Enter** to open, **Escape** to close.

### Filtered Search

List pages (subnets, IPs, etc.) include filter bars for status, tags, VRF, and text search.

---

## Importing Data

**Admin → Data Tools → Import**

- Bulk import IPs from CSV
- Bulk import users from CSV
- NetBox migration export ingestion

### CSV Import Format for IPs

```csv
subnet_cidr,address,status,assigned_to,description
10.0.0.0/24,10.0.0.10,assigned,web-01,Primary web server
10.0.0.0/24,10.0.0.11,reserved,,Reserved for DR
```

---

## Exporting Data

From any list page, click **Export** to download CSV or JSON. Exports:
- Escape formula-prefix characters (OWASP CSV injection protection)
- Redact sensitive values from audit payloads

---

## Tags

Tags are key/value labels (or plain labels) applicable to most objects.

Examples: `env=prod`, `team=platform`, `region=us-east`

Add tags from any object's detail page. Filter by tag in list views.

---

## Audit Logs

**Admin → Audit Logs**

Every significant action is logged with:
- User and source IP
- Timestamp
- Object type and ID
- Action performed
- Changed field values (sensitive values redacted)

Filter by user, action type, date range, or object. Export to CSV or JSON.

---

## Notifications

Webhook-based outbound notifications are configured under **Admin → Webhooks**. In-app notification preferences are in **My Settings → Notifications**.

---

## User Preferences

**My Settings** (top-right user menu):

| Setting | Location |
|---------|----------|
| Change password | My Settings → Security |
| Enable/disable MFA | My Settings → Security |
| API tokens | My Settings → API Tokens |
| Theme (dark/light) | My Settings → Appearance |
| Privacy settings | My Settings → Privacy |

---

## Multi-Factor Authentication

### Setup MFA

1. **My Settings → Security → Enable MFA**
2. Scan the QR code with your authenticator app (Google Authenticator, Authy, 1Password, etc.)
3. Enter a TOTP code to confirm
4. **Save your backup codes** — store them securely offline

### Disable MFA

**My Settings → Security → Disable MFA** (requires current TOTP code)

---

## API Tokens

### Create a Token

1. **My Settings → API Tokens → Create Token**
2. Enter a descriptive name
3. Select scope: `read`, `write`, or `admin`
4. Copy the token — it is shown only once

Use in requests:
```
Authorization: Bearer <token>
```

---

## Automation & Integrations

See [API Documentation](API-Documentation) for automation endpoints.

Access **Admin → Integrations** to:
- Manage API tokens (admin view of all tokens)
- View token usage analytics and rate-limit status
- Browse integration templates
- Manage automation policies
- Configure webhooks
