import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { getSection, getSubnets, createSubnet, updateSubnet, deleteSubnet, searchSubnets } from '../api/client'
import Modal from '../components/Modal'

export default function SubnetsPage() {
  const { sectionID } = useParams()
  const navigate = useNavigate()
  const [section, setSection] = useState(null)
  const [subnets, setSubnets] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [searching, setSearching] = useState(false)
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState({ network_address: '', prefix_length: '', description: '' })
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => { load() }, [sectionID])

  async function load() {
    try {
      setLoading(true)
      setSearchQuery('')
      const [secRes, subRes] = await Promise.all([getSection(sectionID), getSubnets(sectionID)])
      setSection(secRes.data)
      setSubnets(subRes.data)
    } catch {
      setError('Failed to load subnets')
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
      const res = await searchSubnets(sectionID, searchQuery)
      setSubnets(res.data)
    } catch {
      setError('Failed to search subnets')
    } finally {
      setSearching(false)
    }
  }

  function handleClearSearch() {
    setSearchQuery('')
    load()
  }

  function openCreate() {
    setForm({ network_address: '', prefix_length: '', description: '' })
    setModal('create')
  }

  function openEdit(subnet) {
    setForm({ network_address: subnet.NetworkAddress, prefix_length: subnet.PrefixLength, description: subnet.Description })
    setModal({ edit: subnet })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      if (modal === 'create') {
        await createSubnet(sectionID, {
          network_address: form.network_address,
          prefix_length: parseInt(form.prefix_length),
          description: form.description,
        })
      } else {
        await updateSubnet(modal.edit.ID, { description: form.description })
      }
      setModal(null)
      load()
    } catch {
      setError('Failed to save subnet')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteSubnet(id)
      setDeleteConfirm(null)
      load()
    } catch {
      setError('Failed to delete subnet')
    }
  }

  if (loading) return <p className="text-gray-500">Loading subnets...</p>

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/sections" className="hover:text-blue-600">Sections</Link>
        <span>/</span>
        <span className="text-gray-800 font-medium">{section?.Name}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800">Subnets</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + New Subnet
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="mb-4">
        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            type="text"
            placeholder="Search subnets..."
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
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Network</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Prefix</th>
              <th className="text-left px-4 py-3 text-gray-600 font-medium">Description</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {subnets.length === 0 && (
              <tr><td colSpan={4} className="px-4 py-6 text-center text-gray-400">No subnets yet</td></tr>
            )}
            {subnets.map(s => (
              <tr key={s.ID} className="border-b last:border-0 hover:bg-gray-50">
                <td
                  className="px-4 py-3 font-mono font-medium text-blue-600 cursor-pointer hover:underline"
                  onClick={() => navigate(`/subnets/${s.ID}/ip-addresses`)}
                >
                  {s.NetworkAddress}
                </td>
                <td className="px-4 py-3 text-gray-600">/{s.PrefixLength}</td>
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
        <Modal title={modal === 'create' ? 'New Subnet' : 'Edit Subnet'} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            {modal === 'create' && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Network Address</label>
                  <input
                    className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="192.168.0.0"
                    value={form.network_address}
                    onChange={e => setForm(f => ({ ...f, network_address: e.target.value }))}
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Prefix Length</label>
                  <input
                    type="number" min="0" max="32"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="24"
                    value={form.prefix_length}
                    onChange={e => setForm(f => ({ ...f, prefix_length: e.target.value }))}
                    required
                  />
                </div>
              </>
            )}
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
