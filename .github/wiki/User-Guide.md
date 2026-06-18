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

## Networks

Networks are top-level organizational containers for subnets.

| Action | How |
|--------|-----|
| List networks | Sidebar → **Networks** |
| Create network | Networks page → **+ New Network** |
| Edit / delete | Row → kebab (⋮) menu |

---

## Subnets

Subnets define CIDR blocks within a network.

| Action | How |
|--------|-----|
| List all subnets | Sidebar → **Subnets** |
| Create subnet | Network detail → **+ New Subnet** |
| View utilization | Subnet row shows `assigned / total` |
| Edit / delete | Subnet detail → kebab menu |

### Subnet Detail Page

Shows all IP addresses in the subnet with status, `assigned_to`, description, and tags. Utilization bar updates as addresses are allocated.

### Resize a Subnet

1. Open a subnet → kebab (⋮) menu → **Resize**
2. Enter the new CIDR prefix (e.g. `10.0.0.0/22`)
3. Click **Resize**

If existing IP records or child subnets would fall outside the new range, the modal lists the conflicts and asks you to type `CONFIRM` before proceeding. Confirming deletes those records.

### Merge Subnets

Merge combines two or more sibling subnets (same prefix length, same parent network) into their common supernet (prefix − 1).

1. Open a subnet → kebab (⋮) menu → **Merge**
2. Select one or more sibling subnets from the list
3. The resulting prefix is shown (e.g. merging two `/24`s produces a `/23`)
4. Click **Merge**

If no siblings with the same prefix length exist in the network, the option is shown but the merge button is unavailable.

### Split a Subnet

Split divides a subnet into equal child subnets at a longer prefix.

1. Open a subnet → kebab (⋮) menu → **Split**
2. Enter the new prefix length (must be longer than the current prefix; produces at most 256 children)
3. Click **Split**

If any existing IP falls on a network or broadcast address of a child subnet, the split is blocked and the conflicting addresses are listed.

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

### Scan Profiles

Scan profiles let you override per-scan settings on a per-subnet basis. Create profiles under **Admin → Discovery → Scan Profiles → + New Profile**.

| Field | Description |
|-------|-------------|
| Name | Required |
| Scan Type | `Ping`, `SNMP`, or `Ping + SNMP` |
| Ping Concurrency | Concurrent probes (default 20) |
| TCP Ports | Comma-separated ports to probe (optional) |
| DNS Lookup | Resolve PTR records during scan |
| SNMP Community | Overrides the global SNMP community string for this profile |
| SNMP Version | `v2c` (default) |

Assign a profile to a subnet in the subnet's edit form. When a scan job runs on that subnet, the profile's settings take effect instead of the job's defaults.

### Scan Retention

**Admin → Discovery → Scan Retention** controls how long raw scan history is kept.

| Setting | Description |
|---------|-------------|
| Raw History Days | Delete raw scan results older than this many days (default 90) |
| Enable Rollup | Compact results older than **Rollup After Days** into daily summaries instead of deleting them |
| Rollup After Days | Age at which results are rolled up (default 30) |

Click **Save** then **Run Prune Now** to apply immediately. Pruning runs automatically on schedule otherwise.

### Topology View

Open a network's detail page and click **Topology** to see a Cytoscape graph of its subnets. Each node shows the subnet CIDR and is color-coded by utilization: green (< 50%), amber (50–80%), red (≥ 80%). Click a node to see its utilization percentage, VLAN association, and a link to its detail page.

### Topology Hints

**Admin → Discovery → Topology Hints** shows inferred relationships between subnets that Padduck has detected but not yet confirmed. Each hint has a confidence score.

| Status | Meaning |
|--------|---------|
| Suggested | Detected but not reviewed |
| Confirmed | Accepted — used to refine the topology view |
| Dismissed | Rejected false positive |

Filter by status and click **Confirm** or **Dismiss** on each row.

---

## Devices

Sidebar → **Devices**. Track physical and virtual devices linked to IP addresses.

| Field | Description |
|-------|-------------|
| Hostname | Required |
| Type | Device type (drives vendor/model autocomplete suggestions) |
| Vendor / Model | Free text with type-filtered suggestions |
| OS Version | Free text |
| Location / Rack | Optional; available when those features are enabled |
| SNMP | Version, community string, or v3 credentials |
| Custom Fields | Admin-defined fields (if configured) |

**Associate an IP with a device:** open the device → **IP Addresses** tab → associate action:
1. Search by address or hostname (or **+ Create `<address>` and select** to create a new IP record)
2. Type an interface name — existing interfaces appear as suggestions; if none match, **+ Create interface "name"** creates it immediately
3. Optionally mark as primary address, then submit

---

## Locations and Racks

Sidebar → **Locations** / **Racks** (feature-gated).

**Locations** support a parent/child hierarchy (site → building → floor → room → cage).

| Field | Description |
|-------|-------------|
| Name | Required |
| Type | `site`, `building`, `floor`, `room`, `cage`, or `other` |
| Parent Location | Builds the tree |
| Address / City / Region / Country | Optional |
| Status | `active`, `planned`, or `retired` |

**Racks** belong to a location and have a name, size (U), and description. Assign devices to a rack by selecting it in the device form, then setting **Rack Unit Start** and **Rack Unit Size**.

---

## Circuits

Sidebar → **Circuits** (feature-gated). Three record types:

**Circuit Provider** — carrier name, account number, support contact, portal URL.

**Physical Circuit** — carrier-assigned Circuit ID, type (e.g. ethernet), status (`active` / `planned` / `down` / `retired`), bandwidth (Mbps), two endpoint locations, customer, install date.

**Logical Circuit** — overlay service on a physical circuit; service ID, type (e.g. l2vpn), status, bandwidth, customer.

---

## DNS Zones

**DNS → Zones** tracks DNS zones for documentation and monitoring. Open a zone to see nameservers, serial number, and zone health. Add nameservers via **+ Nameserver**. If Technitium or PowerDNS is configured, zone serial and record counts are pulled live.

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

---

## DHCP

Sidebar → **DHCP** (feature-gated). Manage DHCP servers and leases independently of the Technitium integration (see the Technitium section below for active sync).

**DHCP Servers** — Add with **+ Server**:

| Field | Description |
|-------|-------------|
| Name | Required |
| Address | Server IP or hostname |
| Vendor | Free text (e.g. `technitium`) |
| Location | Optional |
| Status | `active` or other |

**DHCP Leases** — Add with **+ Lease**: server, IP address, MAC address, hostname, and state (`active`, `expired`, `reserved`, `declined`, `released`).

### Technitium DHCP Integration

Configure the Technitium connection under **Admin → Settings → DNS** (scroll to the DHCP section):

| Field | Description |
|-------|-------------|
| Technitium URL | Base URL of the Technitium server |
| API Token | Technitium admin API token |
| Default Zone | Default DNS zone for auto-created records |
| Skip TLS Verify | Disable certificate validation (development only) |

Click **Test Connection** to verify. Then:

- **Sync Leases Now** — pulls all leases from every configured Technitium DHCP scope and upserts them into the DHCP Leases table, linking each lease to the matching subnet and IP where possible.
- **Load Scopes** — lists all Technitium DHCP scopes. Select a network and click **Import as subnet** on any scope to create a subnet from it.

**Per-IP DHCP reservations** — link a subnet to a Technitium scope by entering the scope name in the subnet edit form (**Technitium Scope Name** field). Assigned IPs on that subnet show **Reserve** and **Unreserve** action buttons in the IP list, which create or delete static reservations in Technitium. The IP must have a MAC address set for this to work.

---

## Single Sign-On (SSO)

SSO providers are configured under **Admin → Settings → Authentication**.

### LDAP / Active Directory

**Admin → Settings → Authentication → LDAP**

| Field | Default | Description |
|-------|---------|-------------|
| Enabled | off | Toggle to activate |
| Host | — | LDAP server hostname or IP |
| Port | 389 | LDAP port |
| TLS Mode | `none` | `none`, `starttls`, or `ldaps` |
| Skip Cert Verify | off | Disable TLS verification (development only) |
| Bind DN | — | Service account DN for directory queries |
| Bind Password | — | Service account password |
| Base DN | — | Search base (e.g. `dc=example,dc=com`) |
| User Filter | `(sAMAccountName=%s)` | Filter to locate a user by login name (`%s` is replaced with the entered username) |
| Username Attribute | `sAMAccountName` | Attribute mapped to the Padduck username |
| Email Attribute | `mail` | Attribute mapped to the user's email |

**Group → role mappings**: scroll down to the **Group Mappings** section. Add a mapping by entering an LDAP group DN and selecting a Padduck role. Users whose LDAP group membership matches are assigned that role on login.

Click **Test Connection** to verify the bind credentials before saving.

### OAuth2 / OIDC

**Admin → Settings → Authentication → OAuth2**

| Field | Description |
|-------|-------------|
| Enabled | Toggle to activate |
| Provider Name | Label shown on the login button |
| Discovery URL | OIDC discovery endpoint (fills the URLs below automatically if provided) |
| Authorization URL | OAuth2 authorization endpoint |
| Token URL | OAuth2 token endpoint |
| Userinfo URL | Endpoint to fetch the user profile |
| Client ID | Application client ID |
| Client Secret | Application client secret (stored encrypted) |
| Scopes | Space-separated scopes (default `openid email profile`) |

If your provider supports OIDC discovery, enter only the **Discovery URL** and save — the other URLs are populated automatically.

### SAML 2.0

**Admin → Settings → Authentication → SAML**

| Field | Description |
|-------|-------------|
| Enabled | Toggle to activate |
| IdP Metadata URL | URL to your identity provider's metadata XML (fetched at startup) |
| IdP Metadata XML | Paste metadata XML directly (alternative to URL) |
| Entity ID | Service provider entity ID (leave blank to use the default) |

The page shows your **ACS URL** (`/auth/saml/acs`) and **SP certificate** — copy these into your IdP's application configuration.

---

## Reports

Sidebar → **Reports**.

### Ad-hoc Reports

| Tab | Shows |
|-----|-------|
| **Utilization Trends** | Per-subnet utilization vs. 7 days ago, sortable by current %, prior %, or delta |
| **Inactive IPs** | Assigned/reserved IPs with no activity; filter by threshold (30/60/90/180 days) and network; supports CSV export and bulk **Release** |
| **Duplicate Detection** | Duplicate device hostnames and conflicting IP assignments |
| **Reconciliation Center** | Stale IP assignments, DNS drift (IP/name/PTR mismatches), and overlapping subnet CIDRs |

### Scheduled Reports

**Reports → Scheduled Reports** (admin only) → **+ New Report**:

| Field | Description |
|-------|-------------|
| Name | Display name |
| Report Type | Template (see below) |
| Schedule | Standard 5-field cron expression, e.g. `0 8 * * 1` for Monday 8 AM |
| Recipients | Comma-separated email addresses |
| Format | `CSV` or `PDF` |

Available templates: Utilization Summary, Inactive IPs, Subnet Gaps, VLAN Assignment, IP Age, DNS Audit, Stale Leases, Inactive Devices, Failed Scans. Each row has **Run Now**, **Edit**, and **Delete** actions.

---

## Firewall Zones

Sidebar → **Firewall Zones** (feature-gated). Zones are named segments (e.g. DMZ, LAN, WAN) used to document security boundaries.

**Create a zone** with **+ Zone**: name, description, color, status (`active` or other).

**Assign objects to a zone** with **+ Mapping**:

| Field | Description |
|-------|-------------|
| Zone | Required |
| Object Type | `cidr`, `network`, `subnet`, `ip_address`, `device`, `vlan`, `vrf`, `nat_rule`, and others |
| Object | Database record of the selected type (or a free-form CIDR for the `cidr` type) |
| Direction | `both`, `in`, or `out` |
| Status | `active` or other |

---

## NAT Rules

Sidebar → **NAT Rules** (feature-gated). Document static and dynamic NAT mappings.

**Create a rule** with **+ NAT Rule**:

| Field | Description |
|-------|-------------|
| Name | Required |
| Type | `static` or `dynamic` |
| Internal CIDR | Inside address or range |
| External CIDR | Outside address or range |
| Protocol | `any`, `tcp`, `udp`, or `icmp` |
| Internal / External Port | Optional port mapping |
| Customer | Optional customer link |
| Status | `active`, `planned`, `down`, or `retired` |

---

## Custom Fields

**Admin → Custom Fields** lets you add fields to subnets, IP addresses, or devices.

**Create a field** via **+ Add Field**:

| Field | Description |
|-------|-------------|
| Applies To | `Subnets`, `IP Addresses`, or `Devices` |
| Internal Name | API key (lowercase, no spaces) |
| Label | Display label in the UI |
| Field Type | `text`, `number`, `textarea`, `dropdown`, `checkbox`, `date`, `url`, or `email` |
| Options | Value/label pairs for `dropdown` type |
| Required | Enforce a value on create/edit |
| Default Value | Pre-filled value |
| Searchable | Include in list-page filter bars |

Drag rows to reorder — order controls the display sequence on object forms. Custom fields appear on the create/edit form for the selected object type.

---

## Audit Retention

**Admin → Audit → Retention** sets how long audit log entries are kept.

| Setting | Minimum | Default |
|---------|---------|---------|
| Retention Period (days) | 30 | 365 |
| Archive Mode | — | Off (entries are deleted, not archived) |

Save your settings, then click **Run Prune Now** to delete entries older than the configured period immediately. Automatic pruning also runs on schedule.

---

## Requests and Approvals

Non-admin users can submit requests for new subnets or IP allocations that require admin review.

### Submitting a Request (any user)

- **My Requests** in the sidebar shows your submitted requests (pending, approved, rejected, cancelled)
- From a network or subnet list, use the **Request** action to fill in the details and submit
- Pending requests can be **Cancelled**; rejected requests can be **Re-requested** with adjustments

### Reviewing Requests (admin)

**Admin → Requests** shows all subnet and IP requests.

1. Filter by status (`pending`, `approved`, `rejected`, `cancelled`) and requester
2. Click a row to view the full request details and comments thread
3. Click **Approve** or **Reject** and optionally add a reviewer note
4. The requester is notified by email (if SMTP is configured)

---

## Telemetry

**Admin → Settings → Telemetry**

Telemetry is **opt-in** and disabled by default. When enabled, Padduck sends a daily or weekly anonymous snapshot to the Padduck project. The snapshot never contains hostnames, IP addresses, usernames, descriptions, or any data you have entered. It contains only aggregate counts (e.g. total subnets, total users) and configuration choices (e.g. deployment type, enabled features).

### Enable or Disable

1. **Admin → Settings → Telemetry**
2. Toggle **Enable Telemetry** on or off
3. Click **Save**

On first admin login, if telemetry has never been configured, you are redirected to a setup page that explains what is collected before asking you to choose. You can change your decision at any time from the Telemetry settings tab.

### What Is Collected

| Field | Example |
|-------|---------|
| Install ID (random UUID, not linked to hardware) | `a1b2c3d4-…` |
| Padduck version | `1.32.11` |
| Deployment type | `Docker Compose` |
| Deployment mode | `Self-Hosted` |
| Total subnets, IPs, users, VLANs, customers | counts only |
| IPv4 utilization percentiles | 50th, 75th, 90th, 95th |
| Feature flags (LDAP, OIDC, SAML, SNMP enabled) | true/false |
| Snapshot period | `daily` or `weekly` |

Telemetry data is never used for marketing or sales, never sold, never shared with third parties, and never linked to any identifying property of your deployment.

### Send a Test Snapshot

Click **Send Test Snapshot Now** to post a snapshot immediately and confirm it succeeds before enabling the schedule.

---


## Grafana Data Source

**Admin → Grafana Data Source** — connect Grafana to IPAM data using the SimpleJSON protocol.

1. Copy the **Datasource URL** shown on the page (`https://your-padduck/api/grafana`)
2. Enter a token name and click **Generate** to create a `read`-scoped API token
3. In Grafana, add a **JSON API** or **SimpleJSON** datasource using the URL above
4. Under **Custom HTTP Headers**, add `Authorization: Bearer <token>`
5. Save & Test — the health check should return `ok`

Available metrics:

| Metric | Description |
|--------|-------------|
| `subnet_utilization` | All subnets with CIDR, network, used/total IPs, and utilization % |
| `ip_by_status` | IP counts grouped by status (assigned, available, reserved, …) |
| `section_summary` | Per-network subnet count, total IPs, and used IPs |

In a panel, select a metric and render as **Table** or **Stat**.
