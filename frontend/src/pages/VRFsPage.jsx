import { useState, useEffect } from 'react'
import Modal from '../components/Modal'
import { getVrfs, createVrf, updateVrf, deleteVrf } from '../api/vlans'
import { downloadFile } from '../utils/download'
import { getCachedUser } from '../utils/storageKeys'

// VRF model has no JSON tags — Go field names come through as PascalCase:
//   ID, Name, RouteDistinguisher, Description, CreatedAt, UpdatedAt

const EMPTY_FORM = { Name: '', RouteDistinguisher: '', Description: '' }

export default function VRFsPage() {
  const [vrfs, setVrfs] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: vrf }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [downloading, setDownloading] = useState(false)

  const isAdmin = getCachedUser()?.role === 'admin'

  async function handleExport() {
    setDownloading(true)
    try { await downloadFile('/api/v1/admin/reports/export/vrfs', 'vrfs.csv') }
    catch { setError('Export failed') }
    finally { setDownloading(false) }
  }

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const res = await getVrfs()
      setVrfs(Array.isArray(res.data) ? res.data : [])
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load VRFs.')
    } finally {
      setLoading(false)
    }
  }

  const showMessage = (text, type = 'success') => {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 4000)
  }

  const openCreate = () => {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  const openEdit = (vrf) => {
    setForm({ Name: vrf.Name, RouteDistinguisher: vrf.RouteDistinguisher || '', Description: vrf.Description || '' })
    setModal({ edit: vrf })
  }

  const closeModal = () => { setModal(null); setForm(EMPTY_FORM) }

  const handleSave = async () => {
    if (!form.Name.trim()) return
    setSaving(true)
    try {
      const payload = {
        name: form.Name.trim(),
        route_distinguisher: form.RouteDistinguisher.trim(),
        description: form.Description.trim(),
      }
      if (modal === 'create') {
        await createVrf(payload)
        showMessage('VRF created.')
      } else {
        await updateVrf(modal.edit.ID, payload)
        showMessage('VRF updated.')
      }
      closeModal()
      await load()
    } catch (err) {
      showMessage(err.response?.data?.error || 'Failed to save VRF.', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (vrf) => {
    try {
      await deleteVrf(vrf.ID)
      setDeleteConfirm(null)
      showMessage('VRF deleted.')
      await load()
    } catch (err) {
      showMessage(err.response?.data?.error || 'Failed to delete VRF.', 'error')
    }
  }

  const inputClass = 'w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100'

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">VRFs</h1>
        <div className="flex items-center gap-2">
          {isAdmin && (
            <button onClick={handleExport} disabled={downloading} className="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-sm disabled:opacity-50">
              {downloading ? 'Exporting...' : 'Export CSV'}
            </button>
          )}
          <button
            onClick={openCreate}
            className="bg-blue-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-blue-700 transition"
          >
            + Add VRF
          </button>
        </div>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'}`}>
          {message.text}
        </div>
      )}

      {loading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : vrfs.length === 0 ? (
        <div className="text-center py-16 text-gray-500">
          <p className="text-lg font-medium mb-1">No VRFs</p>
          <p className="text-sm">Create a VRF to separate routing domains.</p>
        </div>
      ) : (
        <div className="border border-gray-200 dark:border-gray-700 rounded overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800 text-gray-700 dark:text-gray-300">
              <tr>
                <th className="px-4 py-3 text-left font-medium">Name</th>
                <th className="px-4 py-3 text-left font-medium">Route Distinguisher</th>
                <th className="px-4 py-3 text-left font-medium">Description</th>
                <th className="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {vrfs.map((vrf) => (
                <tr key={vrf.ID} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                  <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{vrf.Name}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                    {vrf.RouteDistinguisher || <span className="text-gray-400">—</span>}
                  </td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                    {vrf.Description || <span className="text-gray-400">—</span>}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(vrf)}
                      className="text-blue-600 hover:underline text-xs mr-3"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => setDeleteConfirm(vrf)}
                      className="text-red-600 hover:underline text-xs"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {modal && (
        <Modal
          title={modal === 'create' ? 'Add VRF' : 'Edit VRF'}
          onClose={closeModal}
        >
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={form.Name}
                onChange={(e) => setForm((p) => ({ ...p, Name: e.target.value }))}
                className={inputClass}
                placeholder="e.g. management"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Route Distinguisher
              </label>
              <input
                type="text"
                value={form.RouteDistinguisher}
                onChange={(e) => setForm((p) => ({ ...p, RouteDistinguisher: e.target.value }))}
                className={inputClass}
                placeholder="e.g. 65000:100"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Description
              </label>
              <input
                type="text"
                value={form.Description}
                onChange={(e) => setForm((p) => ({ ...p, Description: e.target.value }))}
                className={inputClass}
                placeholder="Optional description"
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                onClick={closeModal}
                className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition"
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={saving || !form.Name.trim()}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition"
              >
                {saving ? 'Saving…' : modal === 'create' ? 'Create' : 'Save'}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {deleteConfirm && (
        <Modal title="Delete VRF" onClose={() => setDeleteConfirm(null)}>
          <p className="text-sm text-gray-700 dark:text-gray-300 mb-4">
            Delete <strong>{deleteConfirm.Name}</strong>? This cannot be undone.
          </p>
          <div className="flex justify-end gap-2">
            <button
              onClick={() => setDeleteConfirm(null)}
              className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition"
            >
              Cancel
            </button>
            <button
              onClick={() => handleDelete(deleteConfirm)}
              className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700 transition"
            >
              Delete
            </button>
          </div>
        </Modal>
      )}
    </div>
  )
}
