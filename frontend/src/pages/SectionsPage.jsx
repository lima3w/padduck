import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { getSections, createSection, updateSection, deleteSection, searchSections } from '../api/client'
import Modal from '../components/Modal'

export default function SectionsPage() {
  const navigate = useNavigate()
  const [sections, setSections] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [searching, setSearching] = useState(false)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: section }
  const [form, setForm] = useState({ name: '', description: '' })
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setSearchQuery('')
      const res = await getSections()
      setSections(res.data)
    } catch {
      setError('Failed to load sections')
    } finally {
      setLoading(false)
    }
  }

  async function handleSearch(e) {
    e.preventDefault()
    if (!searchQuery.trim()) {
      load()
      return
    }
    try {
      setSearching(true)
      const res = await searchSections(searchQuery)
      setSections(res.data)
    } catch {
      setError('Failed to search sections')
    } finally {
      setSearching(false)
    }
  }

  function handleClearSearch() {
    setSearchQuery('')
    load()
  }

  function openCreate() {
    setForm({ name: '', description: '' })
    setModal('create')
  }

  function openEdit(section) {
    setForm({ name: section.Name, description: section.Description })
    setModal({ edit: section })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      if (modal === 'create') {
        await createSection({ name: form.name, description: form.description, created_by: 1 })
      } else {
        await updateSection(modal.edit.ID, { name: form.name, description: form.description })
      }
      setModal(null)
      load()
    } catch {
      setError('Failed to save section')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteSection(id)
      setDeleteConfirm(null)
      load()
    } catch {
      setError('Failed to delete section')
    }
  }

  if (loading) return <p className="text-gray-500">Loading sections...</p>

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800">Sections</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + New Section
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="mb-4">
        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            type="text"
            placeholder="Search sections..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="flex-1 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            disabled={searching}
            className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm font-medium disabled:opacity-50"
          >
            {searching ? 'Searching...' : 'Search'}
          </button>
          {searchQuery && (
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

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Description</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {sections.length === 0 && (
              <tr><td colSpan={3} className="px-4 py-6 text-center text-gray-400">No sections yet</td></tr>
            )}
            {sections.map(s => (
              <tr key={s.ID} className="border-b last:border-0 hover:bg-gray-50">
                <td
                  className="px-4 py-3 font-medium text-blue-600 cursor-pointer hover:underline"
                  onClick={() => navigate(`/sections/${s.ID}/subnets`)}
                >
                  {s.Name}
                </td>
                <td className="px-4 py-3 text-gray-500">{s.Description}</td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(s)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                  {deleteConfirm === s.ID ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => handleDelete(s.ID)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteConfirm(s.ID)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <Modal title={modal === 'create' ? 'New Section' : 'Edit Section'} onClose={() => setModal(null)}>
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
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
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
