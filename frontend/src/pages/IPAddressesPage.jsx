import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getSubnet, getIPAddresses, createIPAddress, assignIPAddress, releaseIPAddress, deleteIPAddress } from '../api/client'
import Modal from '../components/Modal'

const STATUS_COLORS = {
  available: 'bg-green-100 text-green-700',
  assigned: 'bg-blue-100 text-blue-700',
  reserved: 'bg-yellow-100 text-yellow-700',
}

export default function IPAddressesPage() {
  const { subnetID } = useParams()
  const [subnet, setSubnet] = useState(null)
  const [ips, setIPs] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { assign: ip }
  const [form, setForm] = useState({ address: '', hostname: '', status: 'available', assigned_to: '' })
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => { load() }, [subnetID])

  async function load() {
    try {
      setLoading(true)
      const [subRes, ipRes] = await Promise.all([getSubnet(subnetID), getIPAddresses(subnetID)])
      setSubnet(subRes.data)
      setIPs(ipRes.data)
    } catch {
      setError('Failed to load IP addresses')
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm({ address: '', hostname: '', status: 'available', assigned_to: '' })
    setModal('create')
  }

  function openAssign(ip) {
    setForm({ assigned_to: '' })
    setModal({ assign: ip })
  }

  async function handleCreate(e) {
    e.preventDefault()
    setSaving(true)
    try {
      await createIPAddress(subnetID, { address: form.address, hostname: form.hostname, status: form.status })
      setModal(null)
      load()
    } catch {
      setError('Failed to create IP address')
    } finally {
      setSaving(false)
    }
  }

  async function handleAssign(e) {
    e.preventDefault()
    setSaving(true)
    try {
      await assignIPAddress(modal.assign.ID, { assigned_to: form.assigned_to })
      setModal(null)
      load()
    } catch {
      setError('Failed to assign IP address')
    } finally {
      setSaving(false)
    }
  }

  async function handleRelease(id) {
    try {
      await releaseIPAddress(id)
      load()
    } catch {
      setError('Failed to release IP address')
    }
  }

  async function handleDelete(id) {
    try {
      await deleteIPAddress(id)
      setDeleteConfirm(null)
      load()
    } catch {
      setError('Failed to delete IP address')
    }
  }

  if (loading) return <p className="text-gray-500">Loading IP addresses...</p>

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/sections" className="hover:text-blue-600">Sections</Link>
        <span>/</span>
        {subnet && (
          <Link to={`/sections/${subnet.SectionID}/subnets`} className="hover:text-blue-600">Subnets</Link>
        )}
        <span>/</span>
        <span className="text-gray-800 font-medium font-mono">{subnet?.NetworkAddress}/{subnet?.PrefixLength}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800">IP Addresses</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + New IP
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Address</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Hostname</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Status</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Assigned To</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {ips.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-400">No IP addresses yet</td></tr>
            )}
            {ips.map(ip => (
              <tr key={ip.ID} className="border-b last:border-0 hover:bg-gray-50">
                <td className="px-4 py-3 font-mono font-medium text-gray-800">{ip.Address}</td>
                <td className="px-4 py-3 text-gray-500">{ip.Hostname || '—'}</td>
                <td className="px-4 py-3">
                  <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[ip.Status] || 'bg-gray-100 text-gray-600'}`}>
                    {ip.Status}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-500">{ip.AssignedTo || '—'}</td>
                <td className="px-4 py-3 text-right space-x-2">
                  {ip.Status !== 'assigned' && (
                    <button onClick={() => openAssign(ip)} className="text-gray-400 hover:text-blue-600 text-xs">Assign</button>
                  )}
                  {ip.Status === 'assigned' && (
                    <button onClick={() => handleRelease(ip.ID)} className="text-gray-400 hover:text-yellow-600 text-xs">Release</button>
                  )}
                  {deleteConfirm === ip.ID ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => handleDelete(ip.ID)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteConfirm(ip.ID)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal === 'create' && (
        <Modal title="New IP Address" onClose={() => setModal(null)}>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">IP Address</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="192.168.0.10"
                value={form.address}
                onChange={e => setForm(f => ({ ...f, address: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Hostname</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="server01.example.com"
                value={form.hostname}
                onChange={e => setForm(f => ({ ...f, hostname: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.status}
                onChange={e => setForm(f => ({ ...f, status: e.target.value }))}
              >
                <option value="available">Available</option>
                <option value="assigned">Assigned</option>
                <option value="reserved">Reserved</option>
              </select>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Add IP'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal?.assign && (
        <Modal title={`Assign ${modal.assign.Address}`} onClose={() => setModal(null)}>
          <form onSubmit={handleAssign} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Assign To</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="server name or user"
                value={form.assigned_to}
                onChange={e => setForm(f => ({ ...f, assigned_to: e.target.value }))}
                required
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Assign'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
