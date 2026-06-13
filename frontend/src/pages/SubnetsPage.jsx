import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { api } from '../api/client'
import { getNetwork, getSubnet, getSubnetsPaginated, createSubnet, updateSubnet, deleteSubnet, searchSubnets, getSubnetTree } from '../api/ipam'
import { getNameservers } from '../api/dns'
import { getVlans } from '../api/vlans'
import { getCustomFields } from '../api/admin'
import Pagination from '../components/Pagination'
import SubnetTree from '../components/SubnetTree'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'
import { getLocations } from '../api/locations'
import { downloadFile } from '../utils/download'
import { loadPrefs, savePrefs } from '../utils/listPrefs'
import { getCachedUser, LEGACY_STORAGE_KEYS, STORAGE_KEYS } from '../utils/storageKeys'
import SplitSubnetModal from './subnet/SplitSubnetModal'
import MergeSubnetModal from './subnet/MergeSubnetModal'
import ResizeSubnetModal from './subnet/ResizeSubnetModal'
import SubnetFormModal from './subnet/SubnetFormModal'

const DEFAULT_LIMIT = 25
const FILTER_KEY = STORAGE_KEYS.subnetFilters
const LEGACY_FILTER_KEY = LEGACY_STORAGE_KEYS.subnetFilters

const EMPTY_FORM = { network_address: '', prefix_length: '24', description: '', gateway: '', auto_reserve_first: false, auto_reserve_last: false, location_id: '', nameserver_id: '', vlan_id: '', custom_fields: {}, alert_threshold_pct: '', alert_email_override: '' }

export default function SubnetsPage() {
  const { networkID } = useParams()
  const navigate = useNavigate()
  const [network, setSection] = useState(null)
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
  const [filterLocationId, setFilterLocationId] = useState(() => loadPrefs(FILTER_KEY, { filterLocationId: '' }, LEGACY_FILTER_KEY).filterLocationId)
  const [nameservers, setNameservers] = useState([])
  const [vlans, setVlans] = useState([])

  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  // Split modal state
  const [splitModal, setSplitModal] = useState(null) // null | { subnet }
  const [splitPrefix, setSplitPrefix] = useState('')
  const [splitting, setSplitting] = useState(false)
  const [splitError, setSplitError] = useState('')
  const [splitBlockingIPs, setSplitBlockingIPs] = useState([])
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
  }, [networkID])

  useEffect(() => {
    savePrefs(FILTER_KEY, { filterLocationId }, LEGACY_FILTER_KEY)
  }, [filterLocationId])

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
        getNetwork(networkID),
        getSubnetsPaginated(networkID, p, DEFAULT_LIMIT),
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
      const res = await getSubnetTree(networkID)
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
      const res = await searchSubnets(networkID, body)
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
        prefix_length: full.prefixLength != null ? String(full.prefixLength) : '24',
        description: full.description || '',
        gateway: full.gateway || '',
        auto_reserve_first: full.autoReserveFirst || false,
        auto_reserve_last: full.autoReserveLast || false,
        location_id: full.locationId ? String(full.locationId) : '',
        nameserver_id: full.nameserverId ? String(full.nameserverId) : '',
        vlan_id: full.vlanId != null ? String(full.vlanId) : '',
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
        await createSubnet(networkID, {
          network_address: form.network_address,
          prefix_length: form.prefix_length !== '' ? parseInt(form.prefix_length) : 24,
          description: form.description,
          gateway: form.gateway || null,
          auto_reserve_first: form.auto_reserve_first,
          auto_reserve_last: form.auto_reserve_last,
          location_id: form.location_id ? parseInt(form.location_id) : null,
          nameserver_id: form.nameserver_id ? parseInt(form.nameserver_id) : null,
          vlan_id: form.vlan_id !== '' ? parseInt(form.vlan_id) : null,
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
          vlan_id: form.vlan_id !== '' ? parseInt(form.vlan_id) : null,
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
    setSplitBlockingIPs([])
    setSplitSuccess(false)
  }

  async function handleSplit() {
    if (!splitModal) return
    setSplitting(true)
    setSplitError('')
    setSplitBlockingIPs([])
    try {
      await api.post(`/admin/subnets/${splitModal.subnet.id}/split`, { new_prefix_len: parseInt(splitPrefix) })
      setSplitSuccess(true)
      setSplitModal(null)
      showToast('Subnet split successfully')
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      const data = err.response?.data
      if (data?.blocking_ips?.length) {
        setSplitBlockingIPs(data.blocking_ips)
        setSplitError('Split blocked: the following IPs fall on network or broadcast addresses and must be removed first.')
      } else {
        setSplitError(data?.error || 'Failed to split subnet')
      }
    } finally {
      setSplitting(false)
    }
  }

  async function openMerge(subnet) {
    try {
      const res = await api.get(`/networks/${networkID}/subnets`)
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
      await downloadFile(`/api/v1/admin/reports/export/subnets?format=csv`, `subnets-network-${networkID}.csv`)
    } catch {
      setError('Export failed')
    } finally {
      setDownloading(false)
    }
  }

  if (loading) return <PageSpinner message="Loading subnets..." />

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/networks" className="hover:text-blue-600">Networks</Link>
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-medium">{network?.name}</span>
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

      <ErrorBanner error={error} />

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
            <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Network</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Prefix</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Location</th>
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
                  <EmptyRow colSpan={6 + searchableFields.length} message="No subnets yet." />
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
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {s.locationId ? (
                        <Link to={`/locations/${s.locationId}`} className="text-blue-600 dark:text-blue-400 hover:underline text-xs">
                          {locations.find(l => l.id === s.locationId)?.name || `#${s.locationId}`}
                        </Link>
                      ) : '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                      {s.vlanId != null ? (
                        <Link to={`/vlans/${s.vlanId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                          VLAN {vlans.find(v => v.id === s.vlanId)?.vlanId ?? `#${s.vlanId}`}
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
              <div className="overflow-x-auto">
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
        <SplitSubnetModal
          splitModal={splitModal}
          splitPrefix={splitPrefix}
          setSplitPrefix={setSplitPrefix}
          splitting={splitting}
          splitError={splitError}
          splitBlockingIPs={splitBlockingIPs}
          onSplit={handleSplit}
          onClose={() => setSplitModal(null)}
        />
      )}

      {mergeModal && (
        <MergeSubnetModal
          mergeModal={mergeModal}
          mergeSelected={mergeSelected}
          setMergeSelected={setMergeSelected}
          merging={merging}
          mergeError={mergeError}
          onMerge={handleMerge}
          onClose={() => setMergeModal(null)}
        />
      )}

      {resizeModal && (
        <ResizeSubnetModal
          resizeModal={resizeModal}
          resizePrefix={resizePrefix}
          setResizePrefix={setResizePrefix}
          resizing={resizing}
          resizeError={resizeError}
          setResizeError={setResizeError}
          resizeConfirmText={resizeConfirmText}
          setResizeConfirmText={setResizeConfirmText}
          onResize={handleResize}
          onClose={() => setResizeModal(null)}
        />
      )}

      {modal && (
        <SubnetFormModal
          modal={modal}
          form={form}
          setForm={setForm}
          overlapError={overlapError}
          saving={saving}
          locations={locations}
          nameservers={nameservers}
          vlans={vlans}
          cfDefs={cfDefs}
          onSubmit={handleSubmit}
          onClose={() => setModal(null)}
        />
      )}
    </div>
  )
}
