const BASE = '/api/v1'

function getHeaders() {
  const token = localStorage.getItem('auth_token')
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }
}

export async function getLocations(params = {}) {
  const qs = new URLSearchParams(params).toString()
  const res = await fetch(`${BASE}/locations${qs ? '?' + qs : ''}`, { headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function getLocationTree() {
  const res = await fetch(`${BASE}/locations/tree`, { headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function getLocation(id) {
  const res = await fetch(`${BASE}/locations/${id}`, { headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function createLocation(data) {
  const res = await fetch(`${BASE}/locations`, { method: 'POST', headers: getHeaders(), body: JSON.stringify(data) })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function updateLocation(id, data) {
  const res = await fetch(`${BASE}/locations/${id}`, { method: 'PUT', headers: getHeaders(), body: JSON.stringify(data) })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
  return res.json()
}

export async function deleteLocation(id) {
  const res = await fetch(`${BASE}/locations/${id}`, { method: 'DELETE', headers: getHeaders() })
  if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
}
