# API Client Examples

These examples target the stable v1 API contract described by `docs/openapi.yaml`.

## JavaScript

```js
const baseURL = 'https://ipam.example.com'
const token = process.env.IPAM_TOKEN

async function allocateIP(subnetId, assignedTo) {
  const res = await fetch(`${baseURL}/api/v1/automation/ip-addresses/allocate`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
      'Idempotency-Key': crypto.randomUUID(),
    },
    body: JSON.stringify({ subnet_id: subnetId, assigned_to: assignedTo }),
  })
  if (!res.ok) throw new Error(`IPAM request failed: ${res.status}`)
  return res.json()
}
```

## Python

```python
import os
import uuid
import requests

base_url = "https://ipam.example.com"
token = os.environ["IPAM_TOKEN"]

response = requests.post(
    f"{base_url}/api/v1/automation/ip-addresses/reserve",
    headers={
        "Authorization": f"Bearer {token}",
        "Idempotency-Key": str(uuid.uuid4()),
    },
    json={"subnet_id": 10, "address": "10.0.0.25", "hostname": "printer-25"},
    timeout=10,
)
response.raise_for_status()
print(response.json())
```

## Webhook Verification

Outbound webhooks include `X-IPAM-Event-Schema-Version` and `X-IPAM-Signature-256`.
Compute an HMAC-SHA256 over the raw request body using the endpoint secret and compare
it to the header value.
