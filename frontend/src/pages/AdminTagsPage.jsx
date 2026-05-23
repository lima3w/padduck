import { useState, useEffect } from 'react'
import { getTags, createTag, updateTag, deleteTag } from '../api/client'
import Modal from '../components/Modal'
import TagBadge from '../components/TagBadge'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const EMPTY_FORM = { name: '', colour: '#6B7280', description: '' }

export default function AdminTagsPage() {
  const [tags, setTags] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: tag }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      const res = await getTags()
      setTags(res.data || [])
    } catch {
      setError('Failed to load tags')
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

  function openEdit(tag) {
    setForm({ name: tag.name, colour: tag.colour, description: tag.description || '' })
    setModal({ edit: tag })
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
        await createTag(payload)
        showMsg('Tag created')
      } else {
        await updateTag(modal.edit.id, payload)
        showMsg('Tag updated')
      }
      setModal(null)
      load()
    } catch(err) {
      setError(err.response?.data?.error || 'Failed to save tag')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteTag(id)
      setDeleteConfirm(null)
      showMsg('Tag deleted')
      load()
    } catch(err) {
      const msg = err.response?.data?.error || 'Failed to delete tag'
      setError(msg)
      setDeleteConfirm(null)
    }
  }

  if (loading) return <PageSpinner message="Loading tags..." />

  return (
    <div className="max-w-3xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">IP Tags</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + New Tag
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      <ErrorBanner error={error} onDismiss={() => setError(null)} />

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Tag</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Colour</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Description</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Type</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {tags.length === 0 && (
              <EmptyRow colSpan={5} message="No tags yet." />
            )}
            {tags.map(tag => (
              <tr key={tag.id} className="border-b last:border-0 hover:bg-gray-50">
                <td className="px-4 py-3"><TagBadge tag={tag} /></td>
                <td className="px-4 py-3 font-mono text-xs text-gray-500">{tag.colour}</td>
                <td className="px-4 py-3 text-gray-500">{tag.description || '—'}</td>
                <td className="px-4 py-3">
                  {tag.isSystem ? (
                    <span className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">System</span>
                  ) : (
                    <span className="text-xs bg-blue-50 text-blue-600 px-2 py-0.5 rounded">Custom</span>
                  )}
                </td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(tag)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                  {!tag.isSystem && (
                    deleteConfirm === tag.id ? (
                      <>
                        <span className="text-red-600 text-xs">Confirm?</span>
                        <button onClick={() => handleDelete(tag.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                        <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                      </>
                    ) : (
                      <button onClick={() => setDeleteConfirm(tag.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                    )
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <Modal title={modal === 'create' ? 'New Tag' : 'Edit Tag'} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Colour</label>
              <div className="flex gap-2 items-center">
                <input
                  type="color"
                  value={form.colour}
                  onChange={e => setForm(f => ({ ...f, colour: e.target.value }))}
                  className="w-10 h-10 rounded cursor-pointer border"
                />
                <input
                  className="flex-1 border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={form.colour}
                  onChange={e => setForm(f => ({ ...f, colour: e.target.value }))}
                  pattern="^#[0-9A-Fa-f]{6}$"
                  placeholder="#6B7280"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="pt-1">
              <p className="text-xs text-gray-500 mb-1">Preview:</p>
              <TagBadge tag={{ name: form.name || 'Preview', colour: form.colour }} />
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
