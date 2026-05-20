import { useState, useEffect } from 'react'
import Modal from '../components/Modal'
import {
  getVlanDomains,
  createVlanDomain,
  updateVlanDomain,
  deleteVlanDomain,
} from '../api/client'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import TableActions from '../components/TableActions'
import { getCachedUser } from '../utils/storageKeys'

const EMPTY_FORM = { name: '', description: '' }

export default function VlanDomainsPage() {
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  const [domains, setDomains] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: domain }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const res = await getVlanDomains()
      const data = res.data
      setDomains(Array.isArray(data) ? data : (data?.domains ?? []))
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load VLAN domains')
    } finally {
      setLoading(false)
    }
  }

  function showMsg(text, type = 'success') {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 3000)
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(domain) {
    setForm({ name: domain.name || '', description: domain.description || '' })
    setModal({ edit: domain })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const payload = {
        name: form.name,
        description: form.description || null,
      }
      if (modal === 'create') {
        await createVlanDomain(payload)
        showMsg('VLAN domain created')
      } else {
        await updateVlanDomain(modal.edit.id, payload)
        showMsg('VLAN domain updated')
      }
      setModal(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save VLAN domain')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteVlanDomain(id)
      setDeleteConfirm(null)
      showMsg('VLAN domain deleted')
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to delete VLAN domain')
      setDeleteConfirm(null)
    }
  }

  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-red-600 font-semibold text-lg mb-2">Access Denied</p>
          <p className="text-gray-500 text-sm">You need admin privileges to manage VLAN domains.</p>
        </div>
      </div>
    )
  }

  if (loading) return <PageSpinner message="Loading VLAN domains..." />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">VLAN Domains</h1>
        <button
          onClick={openCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
        >
          + New Domain
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      <ErrorBanner error={error} onDismiss={() => setError(null)} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">VLANs</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {domains.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-gray-400">
                  No VLAN domains yet. Create your first L2 domain to get started.
                </td>
              </tr>
            )}
            {domains.map(domain => (
              <tr key={domain.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{domain.name}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{domain.description || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
                    {domain.vlanCount ?? domain.vlan_count ?? 0}
                  </span>
                </td>
                <td className="px-4 py-3 text-right">
                  <TableActions
                    onEdit={() => openEdit(domain)}
                    onDelete={() => handleDelete(domain.id)}
                    confirming={deleteConfirm === domain.id}
                    onRequestDelete={() => setDeleteConfirm(domain.id)}
                    onCancelDelete={() => setDeleteConfirm(null)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <Modal
          title={modal === 'create' ? 'New VLAN Domain' : 'Edit VLAN Domain'}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. Campus LAN"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={2}
                placeholder="Optional description"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setModal(null)}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
