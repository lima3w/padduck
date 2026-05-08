const BASE = '/api/v1'

function getHeaders() {
  const token = localStorage.getItem('auth_token')
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }
}

export async function getRacks(locationId) {
  const qs = locationId ? `?location_id=${locationId}` : ''
  const res = await fetch(`${BASE}/racks${qs}`, { headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function getRack(id) {
  const res = await fetch(`${BASE}/racks/${id}`, { headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function createRack(data) {
  const res = await fetch(`${BASE}/racks`, { method: 'POST', headers: getHeaders(), body: JSON.stringify(data) })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function updateRack(id, data) {
  const res = await fetch(`${BASE}/racks/${id}`, { method: 'PUT', headers: getHeaders(), body: JSON.stringify(data) })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function deleteRack(id) {
  const res = await fetch(`${BASE}/racks/${id}`, { method: 'DELETE', headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
}

export async function getRackDevices(id) {
  const res = await fetch(`${BASE}/racks/${id}/devices`, { headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}
