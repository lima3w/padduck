import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import Modal from '../components/Modal'
import { getLocationTree, createLocation, updateLocation, deleteLocation } from '../api/locations'

const LOCATION_TYPES = ['site', 'building', 'floor', 'room', 'cage', 'other']

const EMPTY_FORM = { name: '', type: 'site', parent_id: '', address: '', description: '' }

function LocationRow({ node, allNodes, depth, onEdit, onDelete, deleteConfirm, setDeleteConfirm }) {
  const [expanded, setExpanded] = useState(true)
  const hasChildren = node.children && node.children.length > 0
  const indent = depth * 20

  return (
    <>
      <tr className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
        <td className="px-4 py-3">
          <div className="flex items-center gap-1" style={{ paddingLeft: indent }}>
            {hasChildren ? (
              <button
                onClick={() => setExpanded(!expanded)}
                className="text-gray-400 hover:text-gray-600 w-4 text-xs"
              >
                {expanded ? '▼' : '▶'}
              </button>
            ) : (
              <span className="w-4"></span>
            )}
            <Link to={`/locations/${node.id}`} className="text-blue-600 dark:text-blue-400 hover:underline font-medium">
              {node.name}
            </Link>
          </div>
        </td>
        <td className="px-4 py-3 text-gray-500 dark:text-gray-400 capitalize">{node.type}</td>
        <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{node.address || '—'}</td>
        <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{node.description || '—'}</td>
        <td className="px-4 py-3 text-right space-x-2">
          <button onClick={() => onEdit(node)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
          {deleteConfirm === node.id ? (
            <>
              <span className="text-red-600 text-xs">Confirm?</span>
              <button onClick={() => onDelete(node.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
              <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
            </>
          ) : (
            <button onClick={() => setDeleteConfirm(node.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
          )}
        </td>
      </tr>
      {hasChildren && expanded && node.children.map(child => (
        <LocationRow
          key={child.id}
          node={child}
          allNodes={allNodes}
          depth={depth + 1}
          onEdit={onEdit}
          onDelete={onDelete}
          deleteConfirm={deleteConfirm}
          setDeleteConfirm={setDeleteConfirm}
        />
      ))}
    </>
  )
}

export default function LocationsPage() {
  const [tree, setTree] = useState([])
  const [allLocations, setAllLocations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [modal, setModal] = useState(null)
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
      const [treeData, flatData] = await Promise.all([
        getLocationTree(),
        fetch('/api/v1/locations', {
          headers: { Authorization: `Bearer ${localStorage.getItem('auth_token')}` }
        }).then(r => r.ok ? r.json() : [])
      ])
      setTree(Array.isArray(treeData) ? treeData : [])
      setAllLocations(Array.isArray(flatData) ? flatData : (flatData?.locations ?? []))
    } catch (err) {
      setError(err.message || 'Failed to load locations')
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(loc) {
    setForm({
      name: loc.name || '',
      type: loc.type || 'site',
      parent_id: loc.parent_id ? String(loc.parent_id) : '',
      address: loc.address || '',
      description: loc.description || '',
    })
    setModal({ edit: loc })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        name: form.name,
        type: form.type,
        parent_id: form.parent_id ? parseInt(form.parent_id) : null,
        address: form.address || null,
        description: form.description || null,
      }
      if (modal === 'create') {
        await createLocation(body)
      } else {
        await updateLocation(modal.edit.id, body)
      }
      setModal(null)
      load()
    } catch (err) {
      setError(err.message || 'Failed to save location')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteLocation(id)
      setDeleteConfirm(null)
      load()
    } catch (err) {
      setError(err.message || 'Failed to delete location')
    }
  }

  if (loading) return <p className="text-gray-500">Loading locations...</p>

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Locations</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + Add Location
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Address</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {tree.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-6 text-center text-gray-400">
                  No locations yet. Add your first location to get started.
                </td>
              </tr>
            )}
            {tree.map(node => (
              <LocationRow
                key={node.id}
                node={node}
                allNodes={allLocations}
                depth={0}
                onEdit={openEdit}
                onDelete={handleDelete}
                deleteConfirm={deleteConfirm}
                setDeleteConfirm={setDeleteConfirm}
              />
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <Modal
          title={modal === 'create' ? 'Add Location' : 'Edit Location'}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="Main Data Center"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.type}
                onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
              >
                {LOCATION_TYPES.map(t => (
                  <option key={t} value={t} className="capitalize">{t.charAt(0).toUpperCase() + t.slice(1)}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Parent Location</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.parent_id}
                onChange={e => setForm(f => ({ ...f, parent_id: e.target.value }))}
              >
                <option value="">None (top-level)</option>
                {allLocations
                  .filter(l => modal === 'create' || l.id !== modal?.edit?.id)
                  .map(l => (
                    <option key={l.id} value={l.id}>{l.name}</option>
                  ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Address</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="123 Main St, City, State"
                value={form.address}
                onChange={e => setForm(f => ({ ...f, address: e.target.value }))}
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
