import { useState, useEffect } from 'react'
import Modal from '../components/Modal'
import {
  getVlanGroups,
  createVlanGroup,
  updateVlanGroup,
  deleteVlanGroup,
} from '../api/client'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import TableActions from '../components/TableActions'
import { getCachedUser } from '../utils/storageKeys'

const EMPTY_FORM = { name: '', colour: '#6B7280', description: '' }

function GroupSwatch({ colour }) {
  return (
    <span
      className="inline-block w-5 h-5 rounded border border-gray-300 dark:border-gray-600 align-middle"
      style={{ backgroundColor: colour || '#6B7280' }}
    />
  )
}

export default function VlanGroupsPage() {
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: group }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const res = await getVlanGroups()
      const data = res.data
      setGroups(Array.isArray(data) ? data : (data?.groups ?? []))
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load VLAN groups')
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

  function openEdit(group) {
    setForm({
      name: group.name || '',
      colour: group.colour || '#6B7280',
      description: group.description || '',
    })
    setModal({ edit: group })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const payload = {
        name: form.name,
        colour: form.colour,
        description: form.description || null,
      }
      if (modal === 'create') {
        await createVlanGroup(payload)
        showMsg('VLAN group created')
      } else {
        await updateVlanGroup(modal.edit.id, payload)
        showMsg('VLAN group updated')
      }
      setModal(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save VLAN group')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteVlanGroup(id)
      setDeleteConfirm(null)
      showMsg('VLAN group deleted')
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to delete VLAN group')
      setDeleteConfirm(null)
    }
  }

  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-red-600 font-semibold text-lg mb-2">Access Denied</p>
          <p className="text-gray-500 text-sm">You need admin privileges to manage VLAN groups.</p>
        </div>
      </div>
    )
  }

  if (loading) return <PageSpinner message="Loading VLAN groups..." />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">VLAN Groups</h1>
        <button
          onClick={openCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
        >
          + New Group
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      <ErrorBanner error={error} onDismiss={() => setError(null)} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Colour</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {groups.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-gray-400">
                  No VLAN groups yet. Create a group to categorise VLANs.
                </td>
              </tr>
            )}
            {groups.map(group => (
              <tr key={group.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <GroupSwatch colour={group.colour} />
                    <span className="font-mono text-xs text-gray-500 dark:text-gray-400">{group.colour}</span>
                  </div>
                </td>
                <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{group.name}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{group.description || '—'}</td>
                <td className="px-4 py-3 text-right">
                  <TableActions
                    onEdit={() => openEdit(group)}
                    onDelete={() => handleDelete(group.id)}
                    confirming={deleteConfirm === group.id}
                    onRequestDelete={() => setDeleteConfirm(group.id)}
                    onCancelDelete={() => setDeleteConfirm(null)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {modal && (
        <Modal
          title={modal === 'create' ? 'New VLAN Group' : 'Edit VLAN Group'}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. Production"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Colour</label>
              <div className="flex gap-2 items-center">
                <input
                  type="color"
                  value={form.colour}
                  onChange={e => setForm(f => ({ ...f, colour: e.target.value }))}
                  className="w-10 h-10 rounded cursor-pointer border dark:border-gray-600"
                />
                <input
                  className="flex-1 border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.colour}
                  onChange={e => setForm(f => ({ ...f, colour: e.target.value }))}
                  pattern="^#[0-9A-Fa-f]{6}$"
                  placeholder="#6B7280"
                />
              </div>
              <div className="mt-2 flex items-center gap-2">
                <GroupSwatch colour={form.colour} />
                <span className="text-xs text-gray-500">Preview</span>
              </div>
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
