import { useEffect, useState } from 'react'
import { createDHCPLease, createDHCPServer, deleteDHCPLease, deleteDHCPServer, getDHCPLeases, getDHCPServers, updateDHCPLease, updateDHCPServer } from '../api/modules'
import { getLocations } from '../api/locations'
import Modal from '../components/Modal'

const SERVER_EMPTY = { name: '', address: '', vendor: '', version: '', location_id: '', description: '', status: 'active' }
const LEASE_EMPTY = { server_id: '', ip_address: '', mac_address: '', hostname: '', state: 'active' }

export default function DHCPPage() {
  const [servers, setServers] = useState([])
  const [leases, setLeases] = useState([])
  const [locations, setLocations] = useState([])
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState({})
  const [error, setError] = useState('')

  useEffect(() => { load() }, [])

  async function load() {
    try {
      const [serversRes, leasesRes, locationsRes] = await Promise.all([getDHCPServers(), getDHCPLeases(), getLocations()])
      setServers(serversRes.data || [])
      setLeases(leasesRes.data || [])
      setLocations(Array.isArray(locationsRes) ? locationsRes : [])
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load DHCP data')
    }
  }

  async function saveServer(e) {
    e.preventDefault()
    const body = { ...form, location_id: form.location_id ? Number(form.location_id) : null }
    try {
      if (modal?.item?.id) await updateDHCPServer(modal.item.id, body)
      else await createDHCPServer(body)
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Save failed')
    }
  }

  async function saveLease(e) {
    e.preventDefault()
    const body = { ...form, server_id: Number(form.server_id) }
    try {
      if (modal?.item?.id) await updateDHCPLease(modal.item.id, body)
      else await createDHCPLease(body)
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Save failed')
    }
  }

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">DHCP</h1>
        <div className="flex gap-2">
          <button onClick={() => { setForm(SERVER_EMPTY); setModal({ type: 'server' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">+ Server</button>
          <button onClick={() => { setForm(LEASE_EMPTY); setModal({ type: 'lease' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">+ Lease</button>
        </div>
      </div>
      {error && <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>}

      <network>
        <h2 className="text-lg font-semibold mb-3">Servers</h2>
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50"><tr><th className="px-4 py-3 text-left">Name</th><th className="px-4 py-3 text-left">Address</th><th className="px-4 py-3 text-left">Vendor</th><th className="px-4 py-3 text-left">Location</th><th /></tr></thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {servers.length === 0 && <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-500">No DHCP servers yet.</td></tr>}
              {servers.map(s => <tr key={s.id}><td className="px-4 py-3 font-medium">{s.name}</td><td className="px-4 py-3 font-mono">{s.address}</td><td className="px-4 py-3">{s.vendor || '-'}</td><td className="px-4 py-3">{s.locationName || '-'}</td><td className="px-4 py-3 text-right space-x-2"><button className="text-blue-600 text-xs" onClick={() => { setForm({ ...SERVER_EMPTY, ...s, location_id: s.locationId || '' }); setModal({ type: 'server', item: s }) }}>Edit</button><button className="text-red-600 text-xs" onClick={async () => { await deleteDHCPServer(s.id); load() }}>Delete</button></td></tr>)}
            </tbody>
          </table>
        </div>
      </network>

      <network>
        <h2 className="text-lg font-semibold mb-3">Leases</h2>
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50"><tr><th className="px-4 py-3 text-left">IP Address</th><th className="px-4 py-3 text-left">MAC</th><th className="px-4 py-3 text-left">Hostname</th><th className="px-4 py-3 text-left">Server</th><th className="px-4 py-3 text-left">State</th><th /></tr></thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {leases.length === 0 && <tr><td colSpan={6} className="px-4 py-6 text-center text-gray-500">No DHCP leases yet.</td></tr>}
              {leases.map(l => <tr key={l.id}><td className="px-4 py-3 font-mono">{l.ipAddress}</td><td className="px-4 py-3 font-mono">{l.macAddress}</td><td className="px-4 py-3">{l.hostname || '-'}</td><td className="px-4 py-3">{l.serverName || l.serverId}</td><td className="px-4 py-3">{l.state}</td><td className="px-4 py-3 text-right space-x-2"><button className="text-blue-600 text-xs" onClick={() => { setForm({ ...LEASE_EMPTY, server_id: l.serverId, ip_address: l.ipAddress, mac_address: l.macAddress, hostname: l.hostname, state: l.state }); setModal({ type: 'lease', item: l }) }}>Edit</button><button className="text-red-600 text-xs" onClick={async () => { await deleteDHCPLease(l.id); load() }}>Delete</button></td></tr>)}
            </tbody>
          </table>
        </div>
      </network>

      {modal?.type === 'server' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">DHCP Server</h2><form onSubmit={saveServer} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <input required placeholder="Name" value={form.name || ''} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input required placeholder="IP address" value={form.address || ''} onChange={e => setForm(f => ({ ...f, address: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder="Vendor" value={form.vendor || ''} onChange={e => setForm(f => ({ ...f, vendor: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder="Version" value={form.version || ''} onChange={e => setForm(f => ({ ...f, version: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <select value={form.location_id || ''} onChange={e => setForm(f => ({ ...f, location_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">No location</option>{locations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}</select>
        <select value={form.status || 'active'} onChange={e => setForm(f => ({ ...f, status: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">Active</option><option value="disabled">Disabled</option><option value="planned">Planned</option><option value="retired">Retired</option></select>
        <input placeholder="Description" value={form.description || ''} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className="md:col-span-2 border rounded px-3 py-2 text-sm" />
        <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setModal(null)} className="px-4 py-2 border rounded text-sm">Cancel</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">Save</button></div>
      </form></Modal>}

      {modal?.type === 'lease' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">DHCP Lease</h2><form onSubmit={saveLease} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <select required value={form.server_id || ''} onChange={e => setForm(f => ({ ...f, server_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">Select server</option>{servers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}</select>
        <select value={form.state || 'active'} onChange={e => setForm(f => ({ ...f, state: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">Active</option><option value="expired">Expired</option><option value="reserved">Reserved</option><option value="declined">Declined</option><option value="released">Released</option></select>
        <input required placeholder="IP address" value={form.ip_address || ''} onChange={e => setForm(f => ({ ...f, ip_address: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input required placeholder="MAC address" value={form.mac_address || ''} onChange={e => setForm(f => ({ ...f, mac_address: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder="Hostname" value={form.hostname || ''} onChange={e => setForm(f => ({ ...f, hostname: e.target.value }))} className="md:col-span-2 border rounded px-3 py-2 text-sm" />
        <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setModal(null)} className="px-4 py-2 border rounded text-sm">Cancel</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">Save</button></div>
      </form></Modal>}
    </div>
  )
}
