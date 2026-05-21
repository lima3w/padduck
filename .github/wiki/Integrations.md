# Integrations

---

## Overview

Padduck integrates with external systems through:

1. **Outbound webhooks** — Padduck pushes events to your systems
2. **Automation API** — your systems pull from or push to Padduck
3. **External auth** — LDAP, OAuth2, SAML for identity federation
4. **DNS and DHCP** — tracking and sync with network services
5. **Grafana** — metrics and dashboards

---

## DNS Providers

Padduck tracks DNS zones and records as documentation. DNS integration allows syncing zone data with your authoritative DNS servers.

Configure under **Admin → DNS**.

Supported integration patterns:
- **Read-only tracking**: import zone data for documentation
- **Bidirectional sync**: keep IPAM records aligned with DNS records
- **Validation**: alert when DNS doesn't match IPAM allocations

---

## DHCP Providers

Padduck tracks DHCP leases for visibility into dynamic address assignments.

Configure under **Admin → DHCP**.

---

## Cloud Integrations

Cloud IP inventory can be imported into Padduck to track VPC subnets and assigned instances. Currently supported via manual CSV import or API push from automation scripts.

Planned: native AWS/GCP/Azure VPC sync.

---

## VMware / Hyper-V / Proxmox Integration

Track VMs in Padduck as **Devices** with associated IP addresses. Import VM inventory via:
- CSV import (**Admin → Data Tools → Import**)
- Automation API (`POST /api/v1/automation/devices/register`)
- Direct API calls with Terraform or Ansible

---

## NetBox Migration

Migrating from NetBox to Padduck:

1. Export data from NetBox (subnets, IPs, VRFs, VLANs, devices)
2. Transform to Padduck CSV import format
3. Use **Admin → Data Tools → Import** or the bulk import API
4. Validate utilization matches between systems
5. Configure webhooks to replace any NetBox automation hooks

The v2 migration bundle export (`GET /api/v1/admin/export/v2-migration-bundle`) assists with major version migration workflows.

See also: [Comparison With NetBox](Comparison-With-NetBox)

---

## Terraform Integration

Use Padduck's REST API from Terraform:

```hcl
resource "null_resource" "ipam_allocation" {
  provisioner "local-exec" {
    command = <<EOT
      curl -X POST \
        -H "Authorization: Bearer ${var.padduck_token}" \
        -H "Idempotency-Key: ${random_uuid.key.result}" \
        -H "Content-Type: application/json" \
        -d '{"subnet_id": ${var.subnet_id}, "assigned_to": "${var.hostname}"}' \
        ${var.padduck_url}/api/v1/automation/ip-addresses/allocate
    EOT
  }
}
```

A native Terraform provider is planned for a future release.

---

## Ansible Integration

Use Padduck's API from Ansible:

```yaml
- name: Allocate IP from Padduck
  uri:
    url: "{{ padduck_url }}/api/v1/automation/ip-addresses/allocate"
    method: POST
    headers:
      Authorization: "Bearer {{ padduck_token }}"
      Idempotency-Key: "{{ ansible_host }}-{{ lookup('pipe', 'date +%Y%m%d') }}"
    body_format: json
    body:
      subnet_id: "{{ subnet_id }}"
      assigned_to: "{{ inventory_hostname }}"
  register: ip_result

- debug:
    msg: "Allocated IP: {{ ip_result.json.address }}"
```

---

## Webhooks

### Configuration

**Admin → Webhooks → + New Webhook**

| Setting | Description |
|---------|-------------|
| URL | Target endpoint (private/loopback addresses rejected) |
| Secret | Used for HMAC-SHA256 signature verification |
| Events | Filter by event type (e.g., `ip_address.assigned`) |
| Object type filter | Filter by object type |
| Tag filter | Only trigger for tagged objects |
| Conditions | Simple `key=value` field conditions |

### Event Delivery

Each delivery includes:
- `X-IPAM-Event-Schema-Version` header
- `X-IPAM-Signature-256` header (HMAC-SHA256 of body)
- `schema_version` in JSON body

### Webhook Signature Verification

```python
import hmac, hashlib

def verify_webhook(payload: bytes, secret: str, signature: str) -> bool:
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)
```

### Replay Failed Deliveries

From **Admin → Webhooks → [endpoint] → Deliveries**, click **Replay** on failed or retrying deliveries to re-queue them after the downstream system recovers.

---

## SIEM Integrations

Export audit logs to your SIEM:
- `GET /api/v1/admin/audit-log` with date range filters
- Export format: JSON or CSV
- Configure a scheduled export job or poll on a cron schedule

For real-time integration, configure a webhook with a broad event filter and forward deliveries to your SIEM ingestion endpoint.

---

## Grafana Integration

Grafana integration is built into Padduck:

1. **Admin → Integrations → Grafana**
2. Provide your Grafana URL and API key
3. Padduck will push utilization dashboards

Alternatively, connect Grafana directly to your PostgreSQL database for custom dashboards using the official Grafana PostgreSQL data source.

---

## Integration Templates

Pre-built integration templates are available under **Admin → Integrations → Templates**. Templates provide example API calls, webhook configurations, and automation patterns for common workflows.

---

## API Client Examples

See [API Documentation](API-Documentation) for full API reference.

### JavaScript

```javascript
const padduck = {
  baseURL: 'https://padduck.example.com',
  token: process.env.PADDUCK_TOKEN,

  async allocateIP(subnetId, assignedTo) {
    const res = await fetch(`${this.baseURL}/api/v1/automation/ip-addresses/allocate`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
        'Idempotency-Key': crypto.randomUUID(),
      },
      body: JSON.stringify({ subnet_id: subnetId, assigned_to: assignedTo }),
    });
    if (!res.ok) throw new Error(`IPAM error: ${res.status}`);
    return res.json();
  }
};
```

### Python

```python
import os, uuid, requests

class PadduckClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers['Authorization'] = f'Bearer {token}'

    def allocate_ip(self, subnet_id, assigned_to):
        r = self.session.post(
            f'{self.base_url}/api/v1/automation/ip-addresses/allocate',
            headers={'Idempotency-Key': str(uuid.uuid4())},
            json={'subnet_id': subnet_id, 'assigned_to': assigned_to},
            timeout=10,
        )
        r.raise_for_status()
        return r.json()
```
