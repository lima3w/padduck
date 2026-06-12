import { useEffect, useState } from 'react'
import { createAutonomousSystem, deleteAutonomousSystem, getAutonomousSystems, updateAutonomousSystem } from '../api/modules'
import Modal from '../components/Modal'

const EMPTY_FORM = { asn: '', name: '', description: '', type: 'external', rir: '' }
const RIR_OPTIONS = ['', 'ARIN', 'RIPE NCC', 'APNIC', 'LACNIC', 'AFRINIC']

export default function AutonomousSystemsPage() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [modal, setModal] = useState(null)
  const [editing, setEditing] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      const res = await getAutonomousSystems()
      setItems(res.data || [])
    } catch {
      setError('Failed to load autonomous systems')
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setEditing(null)
    setModal('create')
  }

  function openEdit(item) {
    setForm({ asn: String(item.asn), name: item.name, description: item.description, type: item.type, rir: item.rir })
    setEditing(item)
    setModal('edit')
  }

  function closeModal() {
    setModal(null)
    setEditing(null)
    setForm(EMPTY_FORM)
    setError('')
  }

  async function handleSave(e) {
    e.preventDefault()
    setSaving(true)
    setError('')
    const payload = { ...form, asn: parseInt(form.asn, 10) }
    try {
      if (modal === 'edit') {
        await updateAutonomousSystem(editing.id, payload)
        setMessage('Autonomous system updated')
      } else {
        await createAutonomousSystem(payload)
        setMessage('Autonomous system created')
      }
      closeModal()
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteAutonomousSystem(id)
      setDeleteConfirm(null)
      setMessage('Autonomous system deleted')
      await load()
    } catch {
      setError('Delete failed')
    }
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">BGP Autonomous Systems</h1>
        <button
          onClick={openCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
        >
          + New AS
        </button>
      </div>

      {message && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded text-sm">
          {message}
        </div>
      )}
      {error && !modal && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-gray-500 text-sm">Loading…</div>
      ) : items.length === 0 ? (
        <div className="text-gray-500 text-sm">No autonomous systems yet.</div>
      ) : (
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600">ASN</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Name</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Type</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">RIR</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Description</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {items.map(item => (
                <tr key={item.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-mono font-medium text-gray-900">AS{item.asn}</td>
                  <td className="px-4 py-3 text-gray-800">{item.name || '—'}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                      item.type === 'internal' ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-600'
                    }`}>
                      {item.type}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-600">{item.rir || '—'}</td>
                  <td className="px-4 py-3 text-gray-500 max-w-xs truncate">{item.description || '—'}</td>
                  <td className="px-4 py-3 text-right space-x-2 whitespace-nowrap">
                    <button onClick={() => openEdit(item)} className="text-blue-600 hover:underline text-xs">Edit</button>
                    <button onClick={() => setDeleteConfirm(item)} className="text-red-600 hover:underline text-xs">Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {modal && (
        <Modal onClose={closeModal}>
          <h2 className="text-lg font-semibold mb-4">{modal === 'edit' ? 'Edit Autonomous System' : 'New Autonomous System'}</h2>
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ASN *</label>
              <input
                type="number"
                min="1"
                max="4294967295"
                value={form.asn}
                onChange={e => setForm(f => ({ ...f, asn: e.target.value }))}
                required
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="e.g. 64512"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                type="text"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
              <select
                value={form.type}
                onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="external">External</option>
                <option value="internal">Internal</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">RIR</label>
              <select
                value={form.rir}
                onChange={e => setForm(f => ({ ...f, rir: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {RIR_OPTIONS.map(r => <option key={r} value={r}>{r || '— None —'}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                type="text"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            {error && <p className="text-red-600 text-sm">{error}</p>}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={closeModal} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving…' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {deleteConfirm && (
        <Modal onClose={() => setDeleteConfirm(null)}>
          <h2 className="text-lg font-semibold mb-2">Delete Autonomous System</h2>
          <p className="text-sm text-gray-600 mb-4">
            Delete <strong>AS{deleteConfirm.asn}</strong>
            {deleteConfirm.name ? ` (${deleteConfirm.name})` : ''}? This cannot be undone.
          </p>
          <div className="flex justify-end gap-2">
            <button onClick={() => setDeleteConfirm(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">Cancel</button>
            <button onClick={() => handleDelete(deleteConfirm.id)} className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700">Delete</button>
          </div>
        </Modal>
      )}
    </div>
  )
}
