import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import Modal from '../components/Modal'
import { getRacks, createRack, updateRack, deleteRack } from '../api/racks'
import { getLocations } from '../api/locations'

const EMPTY_FORM = { name: '', size_u: '42', description: '', location_id: '' }

export default function RacksPage() {
  const [racks, setRacks] = useState([])
  const [locations, setLocations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: rack }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [locationFilter, setLocationFilter] = useState('')

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const [racksData, locsData] = await Promise.all([
        getRacks(),
        getLocations().catch(() => []),
      ])
      setRacks(Array.isArray(racksData) ? racksData : (racksData?.racks ?? []))
      const locs = Array.isArray(locsData) ? locsData : (locsData?.locations ?? [])
      setLocations(locs)
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load racks.')
    } finally {
      setLoading(false)
    }
  }

  const showMessage = (text, type = 'success') => {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 4000)
  }

  const locationName = (id) => {
    if (!id) return null
    const loc = locations.find((l) => l.id === id)
    return loc?.name || `#${id}`
  }

  const openCreate = () => {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  const openEdit = (rack) => {
    setForm({
      name: rack.name,
      size_u: String(rack.sizeU || 42),
      description: rack.description || '',
      location_id: rack.locationId ? String(rack.locationId) : '',
    })
    setModal({ edit: rack })
  }

  const closeModal = () => { setModal(null); setForm(EMPTY_FORM) }

  const handleSave = async () => {
    if (!form.name.trim()) return
    setSaving(true)
    try {
      const payload = {
        name: form.name.trim(),
        size_u: parseInt(form.size_u) || 42,
        description: form.description.trim() || null,
        location_id: form.location_id ? parseInt(form.location_id) : null,
      }
      if (modal === 'create') {
        await createRack(payload)
        showMessage('Rack created.')
      } else {
        await updateRack(modal.edit.id, payload)
        showMessage('Rack updated.')
      }
      closeModal()
      await load()
    } catch (err) {
      showMessage(err.response?.data?.error || 'Failed to save rack.', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (rack) => {
    try {
      await deleteRack(rack.id)
      setDeleteConfirm(null)
      showMessage('Rack deleted.')
      await load()
    } catch (err) {
      showMessage(err.response?.data?.error || 'Failed to delete rack.', 'error')
    }
  }

  const filtered = locationFilter
    ? racks.filter((r) => r.locationId === parseInt(locationFilter))
    : racks

  const inputClass = 'w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100'

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Racks</h1>
        <button
          onClick={openCreate}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-blue-700 transition"
        >
          + Add Rack
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'}`}>
          {message.text}
        </div>
      )}

      {locations.length > 0 && (
        <div className="mb-4 flex items-center gap-2">
          <label className="text-sm text-gray-600 dark:text-gray-400">Filter by location:</label>
          <select
            value={locationFilter}
            onChange={(e) => setLocationFilter(e.target.value)}
            className="text-sm border border-gray-300 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-700 dark:text-gray-100"
          >
            <option value="">All locations</option>
            {locations.map((l) => (
              <option key={l.id} value={l.id}>{l.name}</option>
            ))}
          </select>
        </div>
      )}

      {loading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : filtered.length === 0 ? (
        <div className="text-center py-16 text-gray-500">
          <p className="text-lg font-medium mb-1">No racks found</p>
          <p className="text-sm">Add a rack to start tracking physical equipment.</p>
        </div>
      ) : (
        <div className="border border-gray-200 dark:border-gray-700 rounded overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800 text-gray-700 dark:text-gray-300">
              <tr>
                <th className="px-4 py-3 text-left font-medium">Name</th>
                <th className="px-4 py-3 text-left font-medium">Location</th>
                <th className="px-4 py-3 text-left font-medium">Size</th>
                <th className="px-4 py-3 text-left font-medium">Description</th>
                <th className="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {filtered.map((rack) => (
                <tr key={rack.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                  <td className="px-4 py-3">
                    <Link to={`/racks/${rack.id}`} className="font-medium text-blue-600 dark:text-blue-400 hover:underline">
                      {rack.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                    {rack.locationId ? (
                      <Link to={`/locations/${rack.locationId}`} className="hover:text-blue-600 dark:hover:text-blue-400 hover:underline">
                        {locationName(rack.locationId)}
                      </Link>
                    ) : (
                      <span className="text-gray-400">—</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{rack.sizeU}U</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                    {rack.description || <span className="text-gray-400">—</span>}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button onClick={() => openEdit(rack)} className="text-blue-600 hover:underline text-xs mr-3">Edit</button>
                    <button onClick={() => setDeleteConfirm(rack)} className="text-red-600 hover:underline text-xs">Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {modal && (
        <Modal title={modal === 'create' ? 'Add Rack' : 'Edit Rack'} onClose={closeModal}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input type="text" value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} className={inputClass} placeholder="e.g. Rack-A1" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Size (U)</label>
              <input type="number" min="1" max="100" value={form.size_u} onChange={(e) => setForm((p) => ({ ...p, size_u: e.target.value }))} className={inputClass} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Location</label>
              <select value={form.location_id} onChange={(e) => setForm((p) => ({ ...p, location_id: e.target.value }))} className={inputClass}>
                <option value="">No location</option>
                {locations.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <input type="text" value={form.description} onChange={(e) => setForm((p) => ({ ...p, description: e.target.value }))} className={inputClass} placeholder="Optional" />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={closeModal} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition">Cancel</button>
              <button onClick={handleSave} disabled={saving || !form.name.trim()} className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition">
                {saving ? 'Saving…' : modal === 'create' ? 'Create' : 'Save'}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {deleteConfirm && (
        <Modal title="Delete Rack" onClose={() => setDeleteConfirm(null)}>
          <p className="text-sm text-gray-700 dark:text-gray-300 mb-4">
            Delete rack <strong>{deleteConfirm.name}</strong>? This cannot be undone.
          </p>
          <div className="flex justify-end gap-2">
            <button onClick={() => setDeleteConfirm(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition">Cancel</button>
            <button onClick={() => handleDelete(deleteConfirm)} className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700 transition">Delete</button>
          </div>
        </Modal>
      )}
    </div>
  )
}
