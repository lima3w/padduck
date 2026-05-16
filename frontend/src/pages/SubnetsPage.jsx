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
  getNameservers,
  getVlans,
  getCustomFields,
  api,
} from '../api/client'
import Modal from '../components/Modal'
import Pagination from '../components/Pagination'
import SubnetTree from '../components/SubnetTree'
import CustomFieldForm from '../components/CustomFieldForm'
import { getLocations } from '../api/locations'
import { downloadFile } from '../utils/download'

const DEFAULT_LIMIT = 25

const EMPTY_FORM = { network_address: '', prefix_length: '', description: '', gateway: '', auto_reserve_first: false, auto_reserve_last: false, location_id: '', nameserver_id: '', vlan_id: '', custom_fields: {}, alert_threshold_pct: '', alert_email_override: '' }

function splitCidrPreview(networkAddress, currentPrefix, newPrefix) {
  if (!networkAddress || isNaN(newPrefix) || newPrefix <= currentPrefix || newPrefix > 32) return []
  const parts = networkAddress.split('.').map(Number)
  if (parts.length !== 4 || parts.some(isNaN)) return []
  let base = 0
  for (const p of parts) base = (base << 8) | p
  base = base >>> 0
  const count = Math.pow(2, newPrefix - currentPrefix)
  const size = Math.pow(2, 32 - newPrefix)
  const results = []
  for (let i = 0; i < Math.min(count, 64); i++) {
    const net = (base + i * size) >>> 0
    const octets = [24, 16, 8, 0].map(s => (net >>> s) & 0xff)
    results.push(`${octets.join('.')}/${newPrefix}`)
  }
  return results
}

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
  const [nameservers, setNameservers] = useState([])
  const [vlans, setVlans] = useState([])

  const user = (() => { try { return JSON.parse(localStorage.getItem('current_user')) } catch { return null } })()
  const isAdmin = user?.role === 'admin'

  // Split modal state
  const [splitModal, setSplitModal] = useState(null) // null | { subnet }
  const [splitPrefix, setSplitPrefix] = useState('')
  const [splitting, setSplitting] = useState(false)
  const [splitError, setSplitError] = useState('')
  const [splitSuccess, setSplitSuccess] = useState(false)

  // Merge modal state
  const [mergeModal, setMergeModal] = useState(null) // null | { subnet, siblings: [] }
  const [mergeSelected, setMergeSelected] = useState([])
  const [merging, setMerging] = useState(false)
  const [mergeError, setMergeError] = useState('')

  // Resize modal state
  const [resizeModal, setResizeModal] = useState(null) // null | { subnet }
  const [resizePrefix, setResizePrefix] = useState('')
  const [resizing, setResizing] = useState(false)
  const [resizeError, setResizeError] = useState(null) // null | { message, conflictingIps, conflictingSubnets }
  const [resizeConfirmText, setResizeConfirmText] = useState('')
  const [toast, setToast] = useState('')

  const [downloading, setDownloading] = useState(false)

  useEffect(() => {
    setPage(1)
    setIsSearchActive(false)
    setSearchQuery('')
    load(1)
    loadCfDefs()
    loadLocations()
    loadNameservers()
    loadVlans()
  }, [sectionID])

  async function loadLocations() {
    try {
      const data = await getLocations()
      setLocations(Array.isArray(data) ? data : (data?.locations ?? []))
    } catch {}
  }

  async function loadNameservers() {
    try {
      const res = await getNameservers()
      const data = res.data
      setNameservers(Array.isArray(data) ? data : (data?.nameservers ?? []))
    } catch {}
  }

  async function loadVlans() {
    try {
      const res = await getVlans()
      const data = res.data
      setVlans(Array.isArray(data) ? data : (data?.vlans ?? []))
    } catch {}
  }

  async function loadCfDefs() {
    try {
      const res = await getCustomFields('subnet')
      setCfDefs(Array.isArray(res.data) ? res.data : [])
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

  const searchableFields = cfDefs.filter(d => d.isSearchable)

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
        nameserver_id: full.nameserverId ? String(full.nameserverId) : '',
        vlan_id: full.vlanId ? String(full.vlanId) : '',
        custom_fields: full.customFields || {},
        alert_threshold_pct: full.alertThresholdPct != null ? String(full.alertThresholdPct) : '',
        alert_email_override: full.alertEmailOverride || '',
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
          nameserver_id: form.nameserver_id ? parseInt(form.nameserver_id) : null,
          vlan_id: form.vlan_id ? parseInt(form.vlan_id) : null,
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
          nameserver_id: form.nameserver_id ? parseInt(form.nameserver_id) : null,
          vlan_id: form.vlan_id ? parseInt(form.vlan_id) : null,
          custom_fields: form.custom_fields || {},
          alert_threshold_pct: form.alert_threshold_pct ? parseInt(form.alert_threshold_pct) : null,
          alert_email_override: form.alert_email_override || null,
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

  function showToast(msg) {
    setToast(msg)
    setTimeout(() => setToast(''), 3000)
  }

  function openSplit(subnet) {
    setSplitModal({ subnet })
    setSplitPrefix(String(subnet.prefixLength + 1))
    setSplitError('')
    setSplitSuccess(false)
  }

  async function handleSplit() {
    if (!splitModal) return
    setSplitting(true)
    setSplitError('')
    try {
      await api.post(`/admin/subnets/${splitModal.subnet.id}/split`, { new_prefix_len: parseInt(splitPrefix) })
      setSplitSuccess(true)
      setSplitModal(null)
      showToast('Subnet split successfully')
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      setSplitError(err.response?.data?.error || 'Failed to split subnet')
    } finally {
      setSplitting(false)
    }
  }

  async function openMerge(subnet) {
    try {
      const res = await api.get(`/sections/${sectionID}/subnets`)
      const all = res.data?.data ?? res.data ?? []
      const siblings = all.filter(s => s.id !== subnet.id && s.prefixLength === subnet.prefixLength)
      setMergeModal({ subnet, siblings })
      setMergeSelected([])
      setMergeError('')
    } catch {
      setError('Failed to load siblings for merge')
    }
  }

  async function handleMerge() {
    if (!mergeModal) return
    setMerging(true)
    setMergeError('')
    try {
      const ids = [mergeModal.subnet.id, ...mergeSelected]
      await api.post('/admin/subnets/merge', { subnet_ids: ids })
      setMergeModal(null)
      showToast('Subnets merged successfully')
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      setMergeError(err.response?.data?.error || 'Failed to merge subnets')
    } finally {
      setMerging(false)
    }
  }

  function openResize(subnet) {
    setResizeModal({ subnet })
    setResizePrefix(`${subnet.networkAddress}/${subnet.prefixLength}`)
    setResizeError(null)
    setResizeConfirmText('')
  }

  async function handleResize() {
    if (!resizeModal) return
    setResizing(true)
    setResizeError(null)
    try {
      await api.post(`/admin/subnets/${resizeModal.subnet.id}/resize`, { new_prefix: resizePrefix })
      setResizeModal(null)
      showToast('Subnet resized successfully')
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      if (err.response?.status === 409) {
        const d = err.response.data
        setResizeError({
          message: d.error || 'Resize conflicts with existing data',
          conflictingIps: d.conflicting_ips || [],
          conflictingSubnets: d.conflicting_subnets || [],
        })
      } else {
        setResizeError({ message: err.response?.data?.error || 'Failed to resize subnet' })
      }
    } finally {
      setResizing(false)
    }
  }

  async function handleExportSubnets() {
    setDownloading(true)
    try {
      await downloadFile(`/api/v1/admin/reports/export/subnets?format=csv`, `subnets-section-${sectionID}.csv`)
    } catch {
      setError('Export failed')
    } finally {
      setDownloading(false)
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
          <button
            onClick={handleExportSubnets}
            disabled={downloading}
            className="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-sm disabled:opacity-50"
          >
            {downloading ? 'Exporting...' : 'Export CSV'}
          </button>
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
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Nameserver</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">VLAN</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
                  {searchableFields.map(d => (
                    <th key={d.name} className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{d.label}</th>
                  ))}
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {subnets.length === 0 && (
                  <tr><td colSpan={8 + searchableFields.length} className="px-4 py-6 text-center text-gray-400">No subnets yet</td></tr>
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
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                      {s.nameserverId ? (nameservers.find(ns => ns.id === s.nameserverId)?.name || `#${s.nameserverId}`) : '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                      {s.vlanId ? (
                        <Link to={`/vlans/${s.vlanId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                          VLAN {vlans.find(v => v.id === s.vlanId)?.vlanId || `#${s.vlanId}`}
                        </Link>
                      ) : '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {s.isContainer ? <span className="text-gray-400 italic text-xs">Container subnet</span> : s.description}
                    </td>
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
                      {isAdmin && (
                        <>
                          <button onClick={() => openSplit(s)} className="text-gray-400 hover:text-purple-600 text-xs">Split</button>
                          <button onClick={() => openMerge(s)} className="text-gray-400 hover:text-indigo-600 text-xs">Merge</button>
                          <button onClick={() => openResize(s)} className="text-gray-400 hover:text-teal-600 text-xs">Resize</button>
                        </>
                      )}
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

      {toast && (
        <div className="fixed bottom-4 right-4 bg-green-600 text-white px-4 py-2 rounded shadow-lg text-sm z-50 transition-opacity">
          {toast}
        </div>
      )}

      {splitModal && (
        <Modal title={`Split ${splitModal.subnet.networkAddress}/${splitModal.subnet.prefixLength}`} onClose={() => setSplitModal(null)}>
          <div className="space-y-4">
            {splitError && <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{splitError}</div>}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">New Prefix Length</label>
              <input
                type="number"
                min={splitModal.subnet.prefixLength + 1}
                max={32}
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={splitPrefix}
                onChange={e => setSplitPrefix(e.target.value)}
              />
            </div>
            {splitPrefix && !isNaN(parseInt(splitPrefix)) && parseInt(splitPrefix) > splitModal.subnet.prefixLength && (
              <div>
                <p className="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">Preview — child CIDRs to create:</p>
                <div className="grid grid-cols-2 gap-1 max-h-48 overflow-y-auto">
                  {splitCidrPreview(splitModal.subnet.networkAddress, splitModal.subnet.prefixLength, parseInt(splitPrefix)).map((c, i) => (
                    <span key={i} className="font-mono text-xs bg-gray-50 dark:bg-gray-700 rounded px-2 py-1">{c}</span>
                  ))}
                </div>
              </div>
            )}
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={() => setSplitModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button
                onClick={handleSplit}
                disabled={splitting || !splitPrefix || isNaN(parseInt(splitPrefix)) || parseInt(splitPrefix) <= splitModal.subnet.prefixLength}
                className="px-4 py-2 bg-purple-600 text-white rounded text-sm hover:bg-purple-700 disabled:opacity-50"
              >
                {splitting ? 'Splitting...' : 'Split'}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {mergeModal && (
        <Modal title={`Merge with ${mergeModal.subnet.networkAddress}/${mergeModal.subnet.prefixLength}`} onClose={() => setMergeModal(null)}>
          <div className="space-y-4">
            {mergeError && <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{mergeError}</div>}
            {mergeModal.siblings.length === 0 ? (
              <p className="text-sm text-gray-500">No sibling subnets with the same prefix length found in this section.</p>
            ) : (
              <>
                <p className="text-sm text-gray-600 dark:text-gray-400">Select subnets to merge with <strong className="font-mono">{mergeModal.subnet.networkAddress}/{mergeModal.subnet.prefixLength}</strong>:</p>
                <div className="space-y-1 max-h-48 overflow-y-auto">
                  {mergeModal.siblings.map(s => (
                    <label key={s.id} className="flex items-center gap-2 cursor-pointer p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-700">
                      <input
                        type="checkbox"
                        className="w-4 h-4"
                        checked={mergeSelected.includes(s.id)}
                        onChange={e => setMergeSelected(prev => e.target.checked ? [...prev, s.id] : prev.filter(id => id !== s.id))}
                      />
                      <span className="font-mono text-sm">{s.networkAddress}/{s.prefixLength}</span>
                      {s.description && <span className="text-xs text-gray-400">{s.description}</span>}
                    </label>
                  ))}
                </div>
                {mergeSelected.length > 0 && (
                  <div className="p-3 bg-blue-50 dark:bg-blue-900/20 rounded text-sm text-blue-800 dark:text-blue-300">
                    Merging {1 + mergeSelected.length} subnets with /{mergeModal.subnet.prefixLength - 1} prefix
                  </div>
                )}
              </>
            )}
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={() => setMergeModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              {mergeModal.siblings.length > 0 && (
                <button
                  onClick={handleMerge}
                  disabled={merging || mergeSelected.length === 0}
                  className="px-4 py-2 bg-indigo-600 text-white rounded text-sm hover:bg-indigo-700 disabled:opacity-50"
                >
                  {merging ? 'Merging...' : 'Merge'}
                </button>
              )}
            </div>
          </div>
        </Modal>
      )}

      {resizeModal && (
        <Modal title={`Resize ${resizeModal.subnet.networkAddress}/${resizeModal.subnet.prefixLength}`} onClose={() => setResizeModal(null)}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">New CIDR</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="192.168.0.0/23"
                value={resizePrefix}
                onChange={e => { setResizePrefix(e.target.value); setResizeError(null); setResizeConfirmText('') }}
              />
            </div>
            {resizeError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm dark:bg-red-900/20 dark:border-red-700 dark:text-red-400">
                <p className="font-medium mb-1">{resizeError.message}</p>
                {(resizeError.conflictingIps?.length > 0 || resizeError.conflictingSubnets?.length > 0) && (
                  <>
                    {resizeError.conflictingIps?.length > 0 && (
                      <div className="mt-2">
                        <p className="text-xs font-semibold">Conflicting IPs:</p>
                        <p className="font-mono text-xs">{resizeError.conflictingIps.join(', ')}</p>
                      </div>
                    )}
                    {resizeError.conflictingSubnets?.length > 0 && (
                      <div className="mt-2">
                        <p className="text-xs font-semibold">Conflicting Subnets:</p>
                        <p className="font-mono text-xs">{resizeError.conflictingSubnets.join(', ')}</p>
                      </div>
                    )}
                    <div className="mt-3">
                      <label className="block text-xs font-medium mb-1">Type CONFIRM to proceed anyway:</label>
                      <input
                        className="w-full border rounded px-2 py-1 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-red-400 dark:bg-gray-700 dark:border-gray-600"
                        placeholder="CONFIRM"
                        value={resizeConfirmText}
                        onChange={e => setResizeConfirmText(e.target.value)}
                      />
                    </div>
                  </>
                )}
              </div>
            )}
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={() => setResizeModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button
                onClick={handleResize}
                disabled={resizing || !resizePrefix || (resizeError?.conflictingIps?.length > 0 && resizeConfirmText !== 'CONFIRM')}
                className="px-4 py-2 bg-teal-600 text-white rounded text-sm hover:bg-teal-700 disabled:opacity-50"
              >
                {resizing ? 'Resizing...' : 'Resize'}
              </button>
            </div>
          </div>
        </Modal>
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
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Nameserver (optional)</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.nameserver_id}
                onChange={e => setForm(f => ({ ...f, nameserver_id: e.target.value }))}
              >
                <option value="">No nameserver</option>
                {nameservers.map(ns => (
                  <option key={ns.id} value={ns.id}>{ns.name} ({ns.server1})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">VLAN (optional)</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.vlan_id}
                onChange={e => setForm(f => ({ ...f, vlan_id: e.target.value }))}
              >
                <option value="">No VLAN</option>
                {vlans.map(vlan => (
                  <option key={vlan.id} value={vlan.id}>VLAN {vlan.vlanId} — {vlan.name}</option>
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
            <div className="border-t dark:border-gray-600 pt-4 space-y-4">
              <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Alert Settings</p>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Alert Threshold % (optional)</label>
                <input
                  type="number" min="1" max="100"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="e.g. 80"
                  value={form.alert_threshold_pct}
                  onChange={e => setForm(f => ({ ...f, alert_threshold_pct: e.target.value }))}
                />
                <p className="text-xs text-gray-400 mt-1">Send alert when utilisation exceeds this percentage</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Alert Email Override (optional)</label>
                <input
                  type="email"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="alerts@example.com"
                  value={form.alert_email_override}
                  onChange={e => setForm(f => ({ ...f, alert_email_override: e.target.value }))}
                />
                <p className="text-xs text-gray-400 mt-1">Override the default alert recipient for this subnet</p>
              </div>
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
