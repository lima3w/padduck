import { useState, useEffect } from 'react'
import Modal from '../components/Modal'
import { getNameservers, createNameserver, updateNameserver, deleteNameserver } from '../api/client'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import { getCachedUser } from '../utils/storageKeys'

const EMPTY_FORM = { name: '', server1: '', server2: '', server3: '', description: '' }

export default function NameserversPage() {
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  const [nameservers, setNameservers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: ns }
  const [form, setForm] = useState(EMPTY_FORM)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    load()
  }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const res = await getNameservers()
      const data = res.data
      setNameservers(Array.isArray(data) ? data : (data?.nameservers ?? []))
    } catch (err) {
      if (err.response?.status === 403) {
        setError('You do not have permission to view nameservers.')
      } else {
        setError(err.response?.data?.error || 'Failed to load nameservers')
      }
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(ns) {
    setForm({
      name: ns.name || '',
      server1: ns.server1 || '',
      server2: ns.server2 || '',
      server3: ns.server3 || '',
      description: ns.description || '',
    })
    setModal({ edit: ns })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        name: form.name,
        server1: form.server1,
        server2: form.server2 || null,
        server3: form.server3 || null,
        description: form.description || null,
      }
      if (modal === 'create') {
        await createNameserver(body)
      } else {
        await updateNameserver(modal.edit.id, body)
      }
      setModal(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save nameserver')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteNameserver(id)
      setDeleteConfirm(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to delete nameserver')
    }
  }

  if (loading) return <PageSpinner message="Loading nameservers..." />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Nameservers</h1>
        {isAdmin && (
          <button
            onClick={openCreate}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
          >
            + Add Nameserver
          </button>
        )}
      </div>

      <ErrorBanner error={error} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Primary</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Secondary</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Tertiary</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {nameservers.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-6 text-center text-gray-400">
                  No nameservers yet. Add your first nameserver to get started.
                </td>
              </tr>
            )}
            {nameservers.map(ns => (
              <tr key={ns.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{ns.name}</td>
                <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">{ns.server1}</td>
                <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">{ns.server2 || '—'}</td>
                <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">{ns.server3 || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ns.description || '—'}</td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button
                    onClick={() => openEdit(ns)}
                    className="text-gray-400 hover:text-blue-600 text-xs"
                  >
                    Edit
                  </button>
                  {deleteConfirm === ns.id ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button
                        onClick={() => handleDelete(ns.id)}
                        className="text-red-600 hover:text-red-800 text-xs font-medium"
                      >
                        Yes
                      </button>
                      <button
                        onClick={() => setDeleteConfirm(null)}
                        className="text-gray-400 hover:text-gray-600 text-xs"
                      >
                        No
                      </button>
                    </>
                  ) : (
                    <button
                      onClick={() => setDeleteConfirm(ns.id)}
                      className="text-gray-400 hover:text-red-600 text-xs"
                    >
                      Delete
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <Modal
          title={modal === 'create' ? 'Add Nameserver' : 'Edit Nameserver'}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. Corporate DNS"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Primary <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="8.8.8.8"
                value={form.server1}
                onChange={e => setForm(f => ({ ...f, server1: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Secondary</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="8.8.4.4"
                value={form.server2}
                onChange={e => setForm(f => ({ ...f, server2: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Tertiary</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder=""
                value={form.server3}
                onChange={e => setForm(f => ({ ...f, server3: e.target.value }))}
              />
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
