import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getLocation } from '../api/locations'
import { getRacks, createRack, updateRack, deleteRack } from '../api/racks'
import { api } from '../api/client'
import Modal from '../components/Modal'

const RACK_EMPTY_FORM = { name: '', size_u: '42', description: '' }

export default function LocationDetailPage() {
  const { id } = useParams()
  const [location, setLocation] = useState(null)
  const [breadcrumb, setBreadcrumb] = useState([])
  const [subnets, setSubnets] = useState([])
  const [devices, setDevices] = useState([])
  const [racks, setRacks] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [rackModal, setRackModal] = useState(null)
  const [rackForm, setRackForm] = useState(RACK_EMPTY_FORM)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadAll()
  }, [id])

  async function loadAll() {
    try {
      setLoading(true)
      setError(null)
      const loc = await getLocation(id)
      setLocation(loc)

      // Build breadcrumb by traversing parent chain
      const crumbs = []
      let current = loc
      while (current) {
        crumbs.unshift({ id: current.id, name: current.name })
        if (current.parentId) {
          try {
            current = await getLocation(current.parentId)
          } catch {
            break
          }
        } else {
          break
        }
      }
      setBreadcrumb(crumbs)

      await Promise.all([loadSubnets(), loadDevices(), loadRacks()])
    } catch (err) {
      setError(err.message || 'Failed to load location')
    } finally {
      setLoading(false)
    }
  }

  async function loadSubnets() {
    try {
      const res = await api.get(`/locations/${id}/subnets`)
      setSubnets(res.data || [])
    } catch {}
  }

  async function loadDevices() {
    try {
      const res = await api.get(`/locations/${id}/devices`)
      setDevices(res.data || [])
    } catch {}
  }

  async function loadRacks() {
    try {
      const data = await getRacks(id)
      setRacks(Array.isArray(data) ? data : (data?.racks ?? []))
    } catch {}
  }

  function openCreateRack() {
    setRackForm(RACK_EMPTY_FORM)
    setRackModal('create')
  }

  function openEditRack(rack) {
    setRackForm({
      name: rack.name || '',
      size_u: rack.sizeU ? String(rack.sizeU) : '42',
      description: rack.description || '',
    })
    setRackModal({ edit: rack })
  }

  async function handleRackSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        name: rackForm.name,
        location_id: parseInt(id),
        size_u: parseInt(rackForm.size_u) || 42,
        description: rackForm.description || null,
      }
      if (rackModal === 'create') {
        await createRack(body)
      } else {
        await updateRack(rackModal.edit.id, body)
      }
      setRackModal(null)
      loadRacks()
    } catch (err) {
      setError(err.message || 'Failed to save rack')
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteRack(rackId) {
    try {
      await deleteRack(rackId)
      setDeleteConfirm(null)
      loadRacks()
    } catch (err) {
      setError(err.message || 'Failed to delete rack')
    }
  }

  if (loading) return <p className="text-gray-500">Loading location...</p>
  if (error && !location) return <p className="text-red-600">{error}</p>

  return (
    <div>
      {/* Breadcrumb */}
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1 flex-wrap">
        <Link to="/locations" className="hover:text-blue-600">Locations</Link>
        {breadcrumb.map((crumb, i) => (
          <span key={crumb.id} className="flex items-center gap-1">
            <span>/</span>
            {i < breadcrumb.length - 1 ? (
              <Link to={`/locations/${crumb.id}`} className="hover:text-blue-600">{crumb.name}</Link>
            ) : (
              <span className="text-gray-800 dark:text-gray-200 font-medium">{crumb.name}</span>
            )}
          </span>
        ))}
      </nav>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{location?.name}</h1>
          {location?.description && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{location.description}</p>
          )}
        </div>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      {/* Location details */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div>
            <dt className="text-gray-500 dark:text-gray-400">Type</dt>
            <dd className="text-gray-800 dark:text-gray-200 font-medium capitalize">{location?.type}</dd>
          </div>
          {location?.address && (
            <div>
              <dt className="text-gray-500 dark:text-gray-400">Address</dt>
              <dd className="text-gray-800 dark:text-gray-200">{location.address}</dd>
            </div>
          )}
        </dl>
      </div>

      {/* Racks */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">Racks</h2>
          <button onClick={openCreateRack} className="px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
            + Add Rack
          </button>
        </div>
        {racks.length === 0 ? (
          <p className="text-sm text-gray-400">No racks in this location.</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Size</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Utilization</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {racks.map(rack => {
                  const usedU = rack.usedU ?? 0
                  const sizeU = rack.sizeU ?? 42
                  const pct = sizeU > 0 ? Math.round((usedU / sizeU) * 100) : 0
                  return (
                    <tr key={rack.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                      <td className="px-4 py-3">
                        <Link to={`/racks/${rack.id}`} className="text-blue-600 dark:text-blue-400 hover:underline font-medium">
                          {rack.name}
                        </Link>
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{sizeU}U</td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <div className="flex-1 bg-gray-200 dark:bg-gray-600 rounded-full h-2 max-w-32">
                            <div
                              className={`h-2 rounded-full ${pct > 80 ? 'bg-red-500' : pct > 60 ? 'bg-yellow-500' : 'bg-green-500'}`}
                              style={{ width: `${pct}%` }}
                            ></div>
                          </div>
                          <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                            {usedU}/{sizeU}U ({pct}%)
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-right space-x-2">
                        <button onClick={() => openEditRack(rack)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                        {deleteConfirm === rack.id ? (
                          <>
                            <span className="text-red-600 text-xs">Confirm?</span>
                            <button onClick={() => handleDeleteRack(rack.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                            <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                          </>
                        ) : (
                          <button onClick={() => setDeleteConfirm(rack.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Subnets */}
      <div className="mb-6">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">Subnets</h2>
        {subnets.length === 0 ? (
          <p className="text-sm text-gray-400">No subnets assigned to this location.</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Network</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
                </tr>
              </thead>
              <tbody>
                {subnets.map(s => (
                  <tr key={s.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-mono font-medium">
                      <Link
                        to={`/subnets/${s.id}/ip-addresses`}
                        className="text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        {s.networkAddress}/{s.prefixLength}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{s.description || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Devices */}
      <div className="mb-6">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">Devices</h2>
        {devices.length === 0 ? (
          <p className="text-sm text-gray-400">No devices assigned to this location.</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Hostname</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {devices.map(d => (
                  <tr key={d.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-medium">
                      <Link to={`/devices/${d.id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                        {d.hostname}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{d.type?.name || '—'}</td>
                    <td className="px-4 py-3">
                      <span className="flex items-center gap-1.5 text-xs font-medium">
                        <span className={`w-2 h-2 rounded-full ${d.isOnline ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                        <span className={d.isOnline ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}>
                          {d.isOnline ? 'Online' : 'Offline'}
                        </span>
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Rack modal */}
      {rackModal && (
        <Modal
          title={rackModal === 'create' ? 'Add Rack' : 'Edit Rack'}
          onClose={() => setRackModal(null)}
        >
          <form onSubmit={handleRackSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="Rack A1"
                value={rackForm.name}
                onChange={e => setRackForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Size (U)</label>
              <input
                type="number"
                min="1"
                max="100"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={rackForm.size_u}
                onChange={e => setRackForm(f => ({ ...f, size_u: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={2}
                value={rackForm.description}
                onChange={e => setRackForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setRackModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
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
