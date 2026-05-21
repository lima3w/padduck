# Automation-First Design

---

## Philosophy

Padduck is built for infrastructure teams that automate everything. The web UI is a consumer of the same REST API available to your scripts, Terraform, Ansible, and CI pipelines. There is no functionality in the UI that isn't available via API.

This is not an accident — it's a design constraint.

---

## What "Automation-First" Means

### 1. Every UI Action Has an API Equivalent

When you click **Allocate** in the Padduck UI, it calls `POST /api/v1/automation/ip-addresses/allocate`. The same endpoint your Terraform provisioner calls. The same endpoint your deployment pipeline calls.

There are no hidden admin-only mutations, no UI-only workflows that automation can't replicate.

### 2. Idempotency Is Built In

Automation scripts fail. Networks drop. Retries happen. Without idempotency, a retry can allocate the same IP twice.

Padduck's automation endpoints accept an `Idempotency-Key` header (a UUID you generate). If the same key is seen again within the deduplication window, the original response is returned — no duplicate allocation, no error to handle.

```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $(uuidgen)" \
  -H "Content-Type: application/json" \
  -d '{"subnet_id": 42, "assigned_to": "web-01"}' \
  https://padduck.example.com/api/v1/automation/ip-addresses/allocate
```

### 3. Dry-Run Mode

Validate before you commit. Add `"dry_run": true` to any automation write request:

```json
{
  "subnet_id": 42,
  "assigned_to": "web-01",
  "dry_run": true
}
```

The response shows what would happen (policy evaluation, address selection) without making any changes. Use this in CI to verify allocations are valid before deploying.

### 4. Stable API Contract

Automation scripts break when APIs change. Padduck's v1 API is **frozen at OpenAPI 1.26.0**. No breaking changes to existing endpoints, request fields, or response fields — ever. New endpoints and optional fields can be added, but nothing you depend on will disappear.

### 5. Automation Policies

Define rules that govern how automation interacts with IPAM:

```json
{
  "workflow": "ip_address",
  "action": "allocate",
  "effect": "allow",
  "conditions": {"subnet_id": "42"}
}
```

Effects: `allow`, `deny`, or `manual_review` (queues for human approval).

Evaluate a policy without committing:
```bash
POST /api/v1/automation/policies/evaluate
```

---

## Automation Endpoint Reference

| Workflow | Endpoint | Idempotency |
|----------|----------|-------------|
| Allocate next IP | `POST /api/v1/automation/ip-addresses/allocate` | ✅ |
| Reserve specific IP | `POST /api/v1/automation/ip-addresses/reserve` | ✅ |
| Release IP | `POST /api/v1/automation/ip-addresses/:id/release` | ✅ |
| Validate DNS update | `POST /api/v1/automation/dns/update` | ✅ |
| Register device | `POST /api/v1/automation/devices/register` | ✅ |
| Evaluate policy | `POST /api/v1/automation/policies/evaluate` | N/A (read-only) |

---

## Integration Patterns

### Terraform

```hcl
resource "null_resource" "ip_allocation" {
  provisioner "local-exec" {
    command = <<-EOT
      IP=$(curl -s -X POST \
        -H "Authorization: Bearer ${var.padduck_token}" \
        -H "Idempotency-Key: tf-${var.hostname}" \
        -H "Content-Type: application/json" \
        -d '{"subnet_id":${var.subnet_id},"assigned_to":"${var.hostname}"}' \
        ${var.padduck_url}/api/v1/automation/ip-addresses/allocate | jq -r .address)
      echo "Allocated: $IP"
    EOT
  }
}
```

### Ansible

```yaml
- name: Allocate IP from Padduck
  uri:
    url: "{{ padduck_url }}/api/v1/automation/ip-addresses/allocate"
    method: POST
    headers:
      Authorization: "Bearer {{ padduck_token }}"
      Idempotency-Key: "{{ inventory_hostname }}"
    body_format: json
    body:
      subnet_id: "{{ subnet_id }}"
      assigned_to: "{{ inventory_hostname }}"
  register: padduck_result

- set_fact:
    allocated_ip: "{{ padduck_result.json.address }}"
```

### CI/CD Pipeline

```yaml
- name: Allocate staging IP
  run: |
    RESPONSE=$(curl -sf -X POST \
      -H "Authorization: Bearer $PADDUCK_TOKEN" \
      -H "Idempotency-Key: "ci-${{ github.run_id }}" \
      -H "Content-Type: application/json" \
      -d "{"subnet_id": $STAGING_SUBNET_ID, "assigned_to": "$SERVICE_NAME"}" \
      "$PADDUCK_URL/api/v1/automation/ip-addresses/allocate")
    echo "ALLOCATED_IP=$(echo $RESPONSE | jq -r .address)" >> $GITHUB_ENV
```

---

## Webhooks for Event-Driven Automation

Trigger automation when IPAM events occur:

```
ip_address.assigned → notify CMDB, update DNS, trigger provisioning
ip_address.released → trigger deprovisioning workflow
subnet.created → notify network team
```

Configure at **Admin → Webhooks**. Each delivery is signed with HMAC-SHA256 for verification.

---

## API Client Libraries

See [API Documentation](API-Documentation) for JavaScript and Python examples.

A native Terraform provider and Ansible collection are planned for a future release.
