import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import Modal from '../components/Modal'
import Pagination from '../components/Pagination'

const DEFAULT_LIMIT = 50

const EMPTY_FORM = { hostname: '', type_id: '', description: '', vendor: '', model: '', os_version: '' }

export default function DevicesPage() {
  const [devices, setDevices] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [deviceTypes, setDeviceTypes] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [filterHostname, setFilterHostname] = useState('')
  const [filterTypeId, setFilterTypeId] = useState('')
  const [filterOnline, setFilterOnline] = useState('')
  const [isFiltered, setIsFiltered] = useState(false)
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)

  const token = localStorage.getItem('token')
  const headers = { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }

  useEffect(() => {
    loadDeviceTypes()
    load(1)
  }, [])

  async function loadDeviceTypes() {
    try {
      const res = await fetch('/api/v1/device-types', { headers })
      if (res.ok) setDeviceTypes(await res.json() || [])
    } catch {}
  }

  async function load(p = page) {
    try {
      setLoading(true)
      const res = await fetch(`/api/v1/devices?page=${p}&limit=${DEFAULT_LIMIT}`, { headers })
      if (!res.ok) throw new Error()
      const data = await res.json()
      setDevices(data.devices ?? [])
      setTotal(data.total ?? 0)
      setPage(p)
    } catch {
      setError('Failed to load devices')
    } finally {
      setLoading(false)
    }
  }

  async function handleSearch(e) {
    e.preventDefault()
    const body = {}
    if (filterHostname.trim()) body.hostname = filterHostname.trim()
    if (filterTypeId) body.type_id = parseInt(filterTypeId)
    if (filterOnline !== '') body.is_online = filterOnline === 'true'

    if (!Object.keys(body).length) {
      setIsFiltered(false)
      load(1)
      return
    }

    try {
      setLoading(true)
      setIsFiltered(true)
      const res = await fetch('/api/v1/devices/search', {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
      })
      if (!res.ok) throw new Error()
      const data = await res.json()
      setDevices(data.devices ?? [])
      setTotal(data.total ?? 0)
      setPage(1)
    } catch {
      setError('Failed to search devices')
    } finally {
      setLoading(false)
    }
  }

  function handleClearSearch() {
    setFilterHostname('')
    setFilterTypeId('')
    setFilterOnline('')
    setIsFiltered(false)
    load(1)
  }

  function handlePageChange(newPage) {
    setPage(newPage)
    load(newPage)
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(device) {
    setForm({
      hostname: device.hostname || '',
      type_id: device.type_id ? String(device.type_id) : '',
      description: device.description || '',
      vendor: device.vendor || '',
      model: device.model || '',
      os_version: device.os_version || '',
    })
    setModal({ edit: device })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        hostname: form.hostname,
        type_id: form.type_id ? parseInt(form.type_id) : null,
        description: form.description || null,
        vendor: form.vendor || null,
        model: form.model || null,
        os_version: form.os_version || null,
      }
      if (modal === 'create') {
        const res = await fetch('/api/v1/devices', { method: 'POST', headers, body: JSON.stringify(body) })
        if (!res.ok) { const d = await res.json(); throw new Error(d.error || 'Failed') }
      } else {
        const id = modal.edit.id
        const res = await fetch(`/api/v1/devices/${id}`, { method: 'PUT', headers, body: JSON.stringify(body) })
        if (!res.ok) { const d = await res.json(); throw new Error(d.error || 'Failed') }
      }
      setModal(null)
      load(page)
    } catch (err) {
      setError(err.message || 'Failed to save device')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      const res = await fetch(`/api/v1/devices/${id}`, { method: 'DELETE', headers })
      if (!res.ok) throw new Error()
      setDeleteConfirm(null)
      load(page)
    } catch {
      setError('Failed to delete device')
    }
  }

  if (loading && devices.length === 0) return <p className="text-gray-500">Loading devices...</p>

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Devices</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + Add Device
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="mb-4">
        <form onSubmit={handleSearch} className="flex gap-2 flex-wrap">
          <input
            type="text"
            placeholder="Search hostname..."
            value={filterHostname}
            onChange={e => setFilterHostname(e.target.value)}
            className="flex-1 min-w-40 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
          />
          <select
            value={filterTypeId}
            onChange={e => setFilterTypeId(e.target.value)}
            className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
          >
            <option value="">All Types</option>
            {deviceTypes.map(t => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
          <select
            value={filterOnline}
            onChange={e => setFilterOnline(e.target.value)}
            className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
          >
            <option value="">Online/Offline</option>
            <option value="true">Online only</option>
            <option value="false">Offline only</option>
          </select>
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm font-medium disabled:opacity-50"
          >
            Search
          </button>
          {(isFiltered || filterHostname || filterTypeId || filterOnline) && (
            <button
              type="button"
              onClick={handleClearSearch}
              className="px-4 py-2 bg-gray-400 text-white rounded hover:bg-gray-500 text-sm font-medium"
            >
              Clear
            </button>
          )}
        </form>
      </div>

      {!isFiltered && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">
          {total} device{total !== 1 ? 's' : ''}
        </p>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Hostname</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Vendor / Model</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Section</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">IPs</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {devices.length === 0 && (
              <tr><td colSpan={7} className="px-4 py-6 text-center text-gray-400">No devices yet</td></tr>
            )}
            {devices.map(d => (
              <tr key={d.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  <span title={d.type?.name || ''}>
                    {d.type?.icon || '🖥️'}
                  </span>
                </td>
                <td className="px-4 py-3 font-medium">
                  <Link to={`/devices/${d.id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                    {d.hostname}
                  </Link>
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {[d.vendor, d.model].filter(Boolean).join(' / ') || '—'}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {d.section_id ? `#${d.section_id}` : '—'}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {d.ip_count ?? 0}
                </td>
                <td className="px-4 py-3">
                  <span className={`inline-flex items-center gap-1.5 text-xs font-medium`}>
                    <span className={`w-2 h-2 rounded-full ${d.is_online ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                    <span className={d.is_online ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}>
                      {d.is_online ? 'Online' : 'Offline'}
                    </span>
                  </span>
                </td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(d)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                  {deleteConfirm === d.id ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => handleDelete(d.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteConfirm(d.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {!isFiltered && total > DEFAULT_LIMIT && (
        <Pagination
          page={page}
          limit={DEFAULT_LIMIT}
          total={total}
          onChange={handlePageChange}
        />
      )}

      {modal && (
        <Modal
          title={modal === 'create' ? 'Add Device' : 'Edit Device'}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Hostname <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="router01.example.com"
                value={form.hostname}
                onChange={e => setForm(f => ({ ...f, hostname: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.type_id}
                onChange={e => setForm(f => ({ ...f, type_id: e.target.value }))}
              >
                <option value="">No type</option>
                {deviceTypes.map(t => (
                  <option key={t.id} value={t.id}>{t.icon} {t.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={2}
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Vendor</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="Cisco"
                  value={form.vendor}
                  onChange={e => setForm(f => ({ ...f, vendor: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Model</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="ASR-1000"
                  value={form.model}
                  onChange={e => setForm(f => ({ ...f, model: e.target.value }))}
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">OS Version</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="IOS 15.4"
                value={form.os_version}
                onChange={e => setForm(f => ({ ...f, os_version: e.target.value }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
