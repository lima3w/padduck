import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  getSection,
  getSubnet,
  getSubnetsPaginated,
  createSubnet,
  updateSubnet,
  deleteSubnet,
  searchSubnets,
  getSubnetTree,
} from '../api/client'
import Modal from '../components/Modal'
import Pagination from '../components/Pagination'
import SubnetTree from '../components/SubnetTree'
import CustomFieldForm from '../components/CustomFieldForm'
import { getLocations } from '../api/locations'

const DEFAULT_LIMIT = 25

const EMPTY_FORM = { network_address: '', prefix_length: '', description: '', gateway: '', auto_reserve_first: false, auto_reserve_last: false, location_id: '', custom_fields: {} }

export default function SubnetsPage() {
  const { sectionID } = useParams()
  const navigate = useNavigate()
  const [section, setSection] = useState(null)
  const [subnets, setSubnets] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [searching, setSearching] = useState(false)
  const [isSearchActive, setIsSearchActive] = useState(false)
  const [viewMode, setViewMode] = useState('list') // 'list' | 'tree'
  const [treeData, setTreeData] = useState([])
  const [treeLoading, setTreeLoading] = useState(false)
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [overlapError, setOverlapError] = useState(null)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)
  const [cfDefs, setCfDefs] = useState([])
  const [cfFilterRows, setCfFilterRows] = useState([])
  const [locations, setLocations] = useState([])
  const [filterLocationId, setFilterLocationId] = useState('')

  const token = localStorage.getItem('token')
  const cfHeaders = { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }

  useEffect(() => {
    setPage(1)
    setIsSearchActive(false)
    setSearchQuery('')
    load(1)
    loadCfDefs()
    loadLocations()
  }, [sectionID])

  async function loadLocations() {
    try {
      const data = await getLocations()
      setLocations(Array.isArray(data) ? data : (data?.locations ?? []))
    } catch {}
  }

  async function loadCfDefs() {
    try {
      const res = await fetch('/api/v1/admin/custom-fields?entity_type=subnet', { headers: cfHeaders })
      if (res.ok) setCfDefs(await res.json() || [])
    } catch {}
  }

  async function load(p = page) {
    try {
      setLoading(true)
      const [secRes, subRes] = await Promise.all([
        getSection(sectionID),
        getSubnetsPaginated(sectionID, p, DEFAULT_LIMIT),
      ])
      setSection(secRes.data)
      const data = subRes.data
      setSubnets(data.data ?? data)
      setTotal(data.total ?? (Array.isArray(data) ? data.length : 0))
    } catch {
      setError('Failed to load subnets')
    } finally {
      setLoading(false)
    }
  }

  async function loadTree() {
    try {
      setTreeLoading(true)
      const res = await getSubnetTree(sectionID)
      setTreeData(res.data)
    } catch {
      setError('Failed to load subnet tree')
    } finally {
      setTreeLoading(false)
    }
  }

  function handleViewMode(mode) {
    setViewMode(mode)
    if (mode === 'tree' && treeData.length === 0) {
      loadTree()
    }
  }

  const searchableFields = cfDefs.filter(d => d.is_searchable)

  function addCfFilterRow() {
    if (searchableFields.length === 0) return
    setCfFilterRows(rows => [...rows, { field: searchableFields[0].name, op: 'is', value: '' }])
  }

  function updateCfFilterRow(idx, patch) {
    setCfFilterRows(rows => rows.map((r, i) => i === idx ? { ...r, ...patch } : r))
  }

  function removeCfFilterRow(idx) {
    setCfFilterRows(rows => rows.filter((_, i) => i !== idx))
  }

  function addCfFilterFromValue(fieldName, value) {
    setCfFilterRows(rows => {
      const existing = rows.findIndex(r => r.field === fieldName)
      if (existing >= 0) {
        return rows.map((r, i) => i === existing ? { ...r, value } : r)
      }
      return [...rows, { field: fieldName, op: 'is', value }]
    })
  }

  async function handleSearch(e) {
    e.preventDefault()
    const hasQuery = searchQuery.trim()
    const cfFilters = {}
    cfFilterRows.forEach(r => { if (r.value.trim()) cfFilters[r.field] = r.value.trim() })
    const hasCf = Object.keys(cfFilters).length > 0
    const hasLoc = Boolean(filterLocationId)
    if (!hasQuery && !hasCf && !hasLoc) {
      setIsSearchActive(false)
      load(1)
      return
    }
    try {
      setSearching(true)
      setIsSearchActive(true)
      const body = { query: searchQuery || '', limit: 100, offset: 0 }
      if (hasCf) body.custom_fields = cfFilters
      if (hasLoc) body.location_id = parseInt(filterLocationId)
      const res = await searchSubnets(sectionID, body)
      const data = res.data
      setSubnets(Array.isArray(data) ? data : (data.data ?? []))
      setTotal(Array.isArray(data) ? data.length : (data.total ?? 0))
      setPage(1)
    } catch {
      setError('Failed to search subnets')
    } finally {
      setSearching(false)
    }
  }

  function handleClearSearch() {
    setSearchQuery('')
    setCfFilterRows([])
    setFilterLocationId('')
    setIsSearchActive(false)
    load(1)
  }

  function handlePageChange(newPage) {
    setPage(newPage)
    load(newPage)
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setOverlapError(null)
    setModal('create')
  }

  async function openEdit(subnet) {
    try {
      const res = await getSubnet(subnet.id)
      const full = res.data
      setForm({
        network_address: full.networkAddress || '',
        prefix_length: full.prefixLength != null ? String(full.prefixLength) : '',
        description: full.description || '',
        gateway: full.gateway || '',
        auto_reserve_first: full.autoReserveFirst || false,
        auto_reserve_last: full.autoReserveLast || false,
        location_id: full.locationId ? String(full.locationId) : '',
        custom_fields: full.customFields || {},
      })
      setOverlapError(null)
      setModal({ edit: full })
    } catch {
      setError('Failed to load subnet details')
    }
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    setOverlapError(null)
    try {
      if (modal === 'create') {
        await createSubnet(sectionID, {
          network_address: form.network_address,
          prefix_length: parseInt(form.prefix_length),
          description: form.description,
          gateway: form.gateway || null,
          auto_reserve_first: form.auto_reserve_first,
          auto_reserve_last: form.auto_reserve_last,
          location_id: form.location_id ? parseInt(form.location_id) : null,
          custom_fields: form.custom_fields || {},
        })
      } else {
        const id = modal.edit.id
        await updateSubnet(id, {
          description: form.description,
          gateway: form.gateway || null,
          auto_reserve_first: form.auto_reserve_first,
          auto_reserve_last: form.auto_reserve_last,
          location_id: form.location_id ? parseInt(form.location_id) : null,
          custom_fields: form.custom_fields || {},
        })
      }
      setModal(null)
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch(err) {
      if (err.response?.status === 409) {
        const conflicting = err.response.data.conflicting_cidr
        setOverlapError(`Subnet overlaps with existing subnet${conflicting ? ': ' + conflicting : ''}`)
      } else {
        setError(err.response?.data?.error || 'Failed to save subnet')
      }
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteSubnet(id)
      setDeleteConfirm(null)
      load(page)
      if (viewMode === 'tree') loadTree()
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
        <span className="text-gray-800 dark:text-gray-200 font-medium">{section?.name}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Subnets</h1>
        <div className="flex items-center gap-2">
          {/* View toggle */}
          <div className="flex rounded overflow-hidden border border-gray-300 dark:border-gray-600">
            <button
              onClick={() => handleViewMode('list')}
              className={`px-3 py-1.5 text-sm font-medium transition ${
                viewMode === 'list'
                  ? 'bg-blue-600 text-white'
                  : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
              }`}
            >
              List
            </button>
            <button
              onClick={() => handleViewMode('tree')}
              className={`px-3 py-1.5 text-sm font-medium transition border-l border-gray-300 dark:border-gray-600 ${
                viewMode === 'tree'
                  ? 'bg-blue-600 text-white'
                  : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
              }`}
            >
              Tree
            </button>
          </div>
          <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
            + New Subnet
          </button>
        </div>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      {viewMode === 'list' && (
        <>
          <div className="mb-4 space-y-2">
            <form onSubmit={handleSearch} className="flex gap-2 flex-wrap">
              <input
                type="text"
                placeholder="Search subnets..."
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                className="flex-1 min-w-40 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
              />
              {locations.length > 0 && (
                <select
                  value={filterLocationId}
                  onChange={e => setFilterLocationId(e.target.value)}
                  className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
                >
                  <option value="">All Locations</option>
                  {locations.map(l => (
                    <option key={l.id} value={l.id}>{l.name}</option>
                  ))}
                </select>
              )}
              {searchableFields.length > 0 && (
                <button
                  type="button"
                  onClick={addCfFilterRow}
                  className="px-3 py-2 text-sm border rounded hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-300"
                >
                  + Filter
                </button>
              )}
              <button
                type="submit"
                disabled={searching}
                className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm font-medium disabled:opacity-50"
              >
                {searching ? 'Searching...' : 'Search'}
              </button>
              {(isSearchActive || cfFilterRows.length > 0 || filterLocationId) && (
                <button
                  type="button"
                  onClick={handleClearSearch}
                  className="px-4 py-2 bg-gray-400 text-white rounded hover:bg-gray-500 text-sm font-medium"
                >
                  Clear
                </button>
              )}
            </form>
            {cfFilterRows.map((row, idx) => (
              <div key={idx} className="flex gap-2 items-center">
                <select
                  value={row.field}
                  onChange={e => updateCfFilterRow(idx, { field: e.target.value })}
                  className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
                >
                  {searchableFields.map(d => <option key={d.name} value={d.name}>{d.label}</option>)}
                </select>
                <select
                  value={row.op}
                  onChange={e => updateCfFilterRow(idx, { op: e.target.value })}
                  className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
                >
                  <option value="is">is</option>
                  <option value="contains">contains</option>
                  <option value="is not">is not</option>
                </select>
                <input
                  type="text"
                  value={row.value}
                  onChange={e => updateCfFilterRow(idx, { value: e.target.value })}
                  placeholder="value"
                  className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
                />
                <button
                  type="button"
                  onClick={() => removeCfFilterRow(idx)}
                  className="text-gray-400 hover:text-red-600 text-sm px-1"
                >
                  &times;
                </button>
              </div>
            ))}
          </div>

          {!isSearchActive && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">
              {total} subnet{total !== 1 ? 's' : ''}
            </p>
          )}

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Network</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Prefix</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Gateway</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Location</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
                  {searchableFields.map(d => (
                    <th key={d.name} className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{d.label}</th>
                  ))}
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {subnets.length === 0 && (
                  <tr><td colSpan={6 + searchableFields.length} className="px-4 py-6 text-center text-gray-400">No subnets yet</td></tr>
                )}
                {subnets.map(s => (
                  <tr key={s.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td
                      className="px-4 py-3 font-mono font-medium text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"
                      onClick={() => navigate(`/subnets/${s.id}/ip-addresses`)}
                    >
                      {s.networkAddress}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">/{s.prefixLength}</td>
                    <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">{s.gateway || '—'}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {s.locationId ? (
                        <Link to={`/locations/${s.locationId}`} className="text-blue-600 dark:text-blue-400 hover:underline text-xs">
                          {locations.find(l => l.id === s.locationId)?.name || `#${s.locationId}`}
                        </Link>
                      ) : '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{s.description}</td>
                    {searchableFields.map(d => {
                      const val = s.customFields?.[d.name]
                      return (
                        <td key={d.name} className="px-4 py-3 text-gray-500 dark:text-gray-400">
                          {val ? (
                            <button
                              className="hover:text-blue-600 dark:hover:text-blue-400 underline decoration-dotted text-left"
                              onClick={() => addCfFilterFromValue(d.name, val)}
                              title="Filter by this value"
                            >
                              {val}
                            </button>
                          ) : '—'}
                        </td>
                      )
                    })}
                    <td className="px-4 py-3 text-right space-x-2">
                      <button onClick={() => openEdit(s)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                      {deleteConfirm === s.id ? (
                        <>
                          <span className="text-red-600 text-xs">Confirm?</span>
                          <button onClick={() => handleDelete(s.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                          <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                        </>
                      ) : (
                        <button onClick={() => setDeleteConfirm(s.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {!isSearchActive && total > DEFAULT_LIMIT && (
            <Pagination
              page={page}
              limit={DEFAULT_LIMIT}
              total={total}
              onChange={handlePageChange}
            />
          )}
        </>
      )}

      {viewMode === 'tree' && (
        <>
          {treeLoading ? (
            <p className="text-gray-500">Loading tree...</p>
          ) : (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                  <tr>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Network</th>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Used/Total</th>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Utilisation</th>
                    <th className="px-4 py-3"></th>
                  </tr>
                </thead>
                <tbody>
                  <SubnetTree
                    nodes={treeData}
                    onEdit={openEdit}
                    onDelete={handleDelete}
                    navigate={navigate}
                  />
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {modal && (
        <Modal title={modal === 'create' ? 'New Subnet' : 'Edit Subnet'} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            {overlapError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{overlapError}</div>
            )}
            {modal === 'create' && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Network Address</label>
                  <input
                    className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder="192.168.0.0"
                    value={form.network_address}
                    onChange={e => setForm(f => ({ ...f, network_address: e.target.value }))}
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Prefix Length</label>
                  <input
                    type="number" min="0" max="32"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder="24"
                    value={form.prefix_length}
                    onChange={e => setForm(f => ({ ...f, prefix_length: e.target.value }))}
                    required
                  />
                </div>
              </>
            )}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Gateway (optional)</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="192.168.0.1"
                value={form.gateway}
                onChange={e => setForm(f => ({ ...f, gateway: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Location (optional)</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.location_id}
                onChange={e => setForm(f => ({ ...f, location_id: e.target.value }))}
              >
                <option value="">No location</option>
                {locations.map(l => (
                  <option key={l.id} value={l.id}>{l.name}</option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.auto_reserve_first}
                  onChange={e => setForm(f => ({ ...f, auto_reserve_first: e.target.checked }))}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700">Auto-reserve first IP (network address)</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.auto_reserve_last}
                  onChange={e => setForm(f => ({ ...f, auto_reserve_last: e.target.checked }))}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700">Auto-reserve last IP (broadcast address)</span>
              </label>
            </div>
            {cfDefs.length > 0 && (
              <div className="border-t dark:border-gray-600 pt-4">
                <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Custom Fields</p>
                <CustomFieldForm
                  definitions={cfDefs}
                  values={form.custom_fields}
                  onChange={(name, value) => setForm(f => ({ ...f, custom_fields: { ...f.custom_fields, [name]: value } }))}
                />
              </div>
            )}
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
