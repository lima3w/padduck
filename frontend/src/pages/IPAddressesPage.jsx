import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getSubnet, getIPAddressesPaginated, createIPAddress, assignIPAddress, releaseIPAddress, deleteIPAddress, searchIPAddresses, getTags, updateIPMeta } from '../api/client'
import { submitIPRequest } from '../api/requests'
import Modal from '../components/Modal'
import Pagination from '../components/Pagination'
import TagBadge from '../components/TagBadge'
import CustomFieldForm from '../components/CustomFieldForm'

const DEFAULT_LIMIT = 25

const STATUS_COLORS = {
  available: 'bg-green-100 text-green-700',
  assigned: 'bg-blue-100 text-blue-700',
  reserved: 'bg-yellow-100 text-yellow-700',
}

const COLUMN_KEYS = ['address', 'hostname', 'status', 'tag', 'assigned_to', 'device', 'mac_address', 'dns_name', 'ptr_record', 'last_seen']
const COLUMN_LABELS = {
  address: 'Address',
  hostname: 'Hostname',
  status: 'Status',
  tag: 'Tag',
  assigned_to: 'Assigned To',
  device: 'Device',
  mac_address: 'MAC Address',
  dns_name: 'DNS Name',
  ptr_record: 'Hostname/PTR',
  last_seen: 'Last Seen',
}
const DEFAULT_VISIBLE = ['address', 'hostname', 'status', 'tag', 'assigned_to']

const LS_KEY = 'ipam_ip_columns'

function loadColumnVisibility() {
  try {
    const saved = JSON.parse(localStorage.getItem(LS_KEY))
    if (saved && Array.isArray(saved)) return saved
  } catch {}
  return DEFAULT_VISIBLE
}

const IP_REQUEST_EMPTY = { specific_ip: '', dns_name: '', purpose: '' }

export default function IPAddressesPage() {
  const { subnetID } = useParams()
  const user = (() => { try { return JSON.parse(localStorage.getItem('current_user')) } catch { return null } })()
  const canAssignIP = user?.role === 'admin'

  const [subnet, setSubnet] = useState(null)
  const [ips, setIPs] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [tags, setTags] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [searchStatus, setSearchStatus] = useState('')
  const [searching, setSearching] = useState(false)
  const [isSearchActive, setIsSearchActive] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [advFilters, setAdvFilters] = useState({ tag_id: '', mac_address: '', ptr_record: '', is_assigned: '' })
  const [modal, setModal] = useState(null) // null | 'create' | { assign: ip } | { meta: ip } | 'requestIP'
  const [form, setForm] = useState({ address: '', hostname: '', status: 'available', assigned_to: '', tag_id: '', mac_address: '', ptr_record: '', dns_name: '' })
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)
  const [visibleCols, setVisibleCols] = useState(loadColumnVisibility)
  const [showColPicker, setShowColPicker] = useState(false)
  const [cfDefs, setCfDefs] = useState([])
  const [cfFilterRows, setCfFilterRows] = useState([])
  const [ipReqForm, setIPReqForm] = useState(IP_REQUEST_EMPTY)
  const [ipReqError, setIPReqError] = useState(null)
  const [ipReqSuccess, setIPReqSuccess] = useState(false)

  const token = localStorage.getItem('token')
  const cfHeaders = { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }

  useEffect(() => {
    setPage(1)
    setIsSearchActive(false)
    load(1)
    loadCfDefs()
  }, [subnetID])

  async function loadCfDefs() {
    try {
      const res = await fetch('/api/v1/admin/custom-fields?entity_type=ip_address', { headers: cfHeaders })
      if (res.ok) setCfDefs(await res.json() || [])
    } catch {}
  }

  async function load(p = page) {
    try {
      setLoading(true)
      setSearchQuery('')
      setSearchStatus('')
      setIsSearchActive(false)
      const [subRes, ipRes, tagRes] = await Promise.all([
        getSubnet(subnetID),
        getIPAddressesPaginated(subnetID, p, DEFAULT_LIMIT),
        getTags(),
      ])
      setSubnet(subRes.data)
      const data = ipRes.data
      setIPs(data.data ?? data)
      setTotal(data.total ?? (Array.isArray(data) ? data.length : 0))
      setTags(tagRes.data || [])
    } catch {
      setError('Failed to load IP addresses')
    } finally {
      setLoading(false)
    }
  }

  function handlePageChange(newPage) {
    setPage(newPage)
    load(newPage)
  }

  function toggleColumn(col) {
    const next = visibleCols.includes(col)
      ? visibleCols.filter(c => c !== col)
      : [...visibleCols, col]
    // always keep address
    const final = next.includes('address') ? next : ['address', ...next]
    setVisibleCols(final)
    localStorage.setItem(LS_KEY, JSON.stringify(final))
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
    const cfFilters = {}
    cfFilterRows.forEach(r => { if (r.value.trim()) cfFilters[r.field] = r.value.trim() })
    const hasCf = Object.keys(cfFilters).length > 0
    if (!searchQuery.trim() && !searchStatus && !Object.values(advFilters).some(Boolean) && !hasCf) {
      setIsSearchActive(false)
      load(1)
      return
    }
    try {
      setSearching(true)
      setIsSearchActive(true)
      const filters = {}
      if (advFilters.tag_id) filters.tag_id = parseInt(advFilters.tag_id)
      if (advFilters.mac_address) filters.mac_address = advFilters.mac_address
      if (advFilters.ptr_record) filters.ptr_record = advFilters.ptr_record
      if (advFilters.is_assigned !== '') filters.is_assigned = advFilters.is_assigned === 'true'
      if (hasCf) filters.custom_fields = cfFilters
      const res = await searchIPAddresses(subnetID, searchQuery, searchStatus, 50, 0, filters)
      const data = res.data
      setIPs(Array.isArray(data) ? data : (data.data ?? []))
      setTotal(Array.isArray(data) ? data.length : (data.total ?? 0))
      setPage(1)
    } catch {
      setError('Failed to search IP addresses')
    } finally {
      setSearching(false)
    }
  }

  function handleClearSearch() {
    setSearchQuery('')
    setSearchStatus('')
    setIsSearchActive(false)
    setCfFilterRows([])
    setAdvFilters({ tag_id: '', mac_address: '', ptr_record: '', is_assigned: '' })
    load(1)
  }

  function openCreate() {
    setForm({ address: '', hostname: '', status: 'available', assigned_to: '', tag_id: '', mac_address: '', ptr_record: '', dns_name: '', custom_fields: {} })
    setModal('create')
  }

  function openAssign(ip) {
    setForm({ assigned_to: '', tag_id: '', mac_address: '', ptr_record: '' })
    setModal({ assign: ip })
  }

  function openMeta(ip) {
    setForm({
      tag_id: ip.TagID ? String(ip.TagID) : '',
      mac_address: ip.MACAddress || '',
      ptr_record: ip.PTRRecord || ip.ptrRecord || '',
      dns_name: ip.dnsName || '',
      custom_fields: ip.custom_fields || {},
    })
    setModal({ meta: ip })
  }

  async function handleCreate(e) {
    e.preventDefault()
    setSaving(true)
    try {
      await createIPAddress(subnetID, {
        address: form.address,
        hostname: form.hostname,
        status: form.status,
        tag_id: form.tag_id ? parseInt(form.tag_id) : null,
        mac_address: form.mac_address || null,
        ptr_record: form.ptr_record || null,
        dns_name: form.dns_name || null,
        custom_fields: form.custom_fields || {},
      })
      setModal(null)
      load(page)
    } catch(err) {
      setError(err.response?.data?.error || 'Failed to create IP address')
    } finally {
      setSaving(false)
    }
  }

  async function handleAssign(e) {
    e.preventDefault()
    setSaving(true)
    try {
      await assignIPAddress(modal.assign.id, { assigned_to: form.assigned_to })
      setModal(null)
      load(page)
    } catch {
      setError('Failed to assign IP address')
    } finally {
      setSaving(false)
    }
  }

  async function handleUpdateMeta(e) {
    e.preventDefault()
    setSaving(true)
    try {
      await updateIPMeta(modal.meta.id, {
        tag_id: form.tag_id ? parseInt(form.tag_id) : null,
        mac_address: form.mac_address || null,
        ptr_record: form.ptr_record || null,
        dns_name: form.dns_name || null,
        custom_fields: form.custom_fields || {},
      })
      setModal(null)
      load(page)
    } catch {
      setError('Failed to update IP')
    } finally {
      setSaving(false)
    }
  }

  async function handleRelease(id) {
    try {
      await releaseIPAddress(id)
      load(page)
    } catch {
      setError('Failed to release IP address')
    }
  }

  async function handleDelete(id) {
    try {
      await deleteIPAddress(id)
      setDeleteConfirm(null)
      load(page)
    } catch {
      setError('Failed to delete IP address')
    }
  }

  function openIPRequest() {
    setIPReqForm(IP_REQUEST_EMPTY)
    setIPReqError(null)
    setIPReqSuccess(false)
    setModal('requestIP')
  }

  async function handleIPRequestSubmit(e) {
    e.preventDefault()
    setIPReqError(null)
    setSaving(true)
    try {
      await submitIPRequest({
        subnet_id: parseInt(subnetID),
        specific_ip: ipReqForm.specific_ip || null,
        dns_name: ipReqForm.dns_name || null,
        purpose: ipReqForm.purpose,
      })
      setIPReqSuccess(true)
      setTimeout(() => setModal(null), 1500)
    } catch (err) {
      if (err.response?.status === 409) {
        setIPReqError('That IP address is already taken. Please choose a different one or leave blank for auto-assign.')
      } else {
        setIPReqError(err.response?.data?.error || 'Failed to submit request')
      }
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <p className="text-gray-500">Loading IP addresses...</p>

  const col = (key) => visibleCols.includes(key)

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/sections" className="hover:text-blue-600">Sections</Link>
        <span>/</span>
        {subnet && (
          <Link to={`/sections/${subnet.sectionId}/subnets`} className="hover:text-blue-600">Subnets</Link>
        )}
        <span>/</span>
        <span className="text-gray-800 font-medium font-mono">{subnet?.networkAddress}/{subnet?.prefixLength}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800">IP Addresses</h1>
        <div className="flex gap-2 items-center">
          <div className="relative">
            <button
              onClick={() => setShowColPicker(v => !v)}
              className="px-3 py-2 bg-gray-100 text-gray-600 rounded hover:bg-gray-200 text-sm"
              title="Toggle columns"
            >
              Columns
            </button>
            {showColPicker && (
              <div className="absolute right-0 top-9 bg-white border rounded shadow-lg z-10 p-3 min-w-max">
                <p className="text-xs font-medium text-gray-500 mb-2">Show/hide columns</p>
                {COLUMN_KEYS.filter(k => k !== 'address').map(k => (
                  <label key={k} className="flex items-center gap-2 cursor-pointer py-0.5">
                    <input
                      type="checkbox"
                      checked={visibleCols.includes(k)}
                      onChange={() => toggleColumn(k)}
                      className="w-3.5 h-3.5"
                    />
                    <span className="text-sm text-gray-700">{COLUMN_LABELS[k]}</span>
                  </label>
                ))}
              </div>
            )}
          </div>
          {!canAssignIP && (
            <button onClick={openIPRequest} className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 text-sm font-medium">
              Request IP
            </button>
          )}
          {canAssignIP && (
            <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
              + New IP
            </button>
          )}
        </div>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="mb-4 space-y-2">
        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            type="text"
            placeholder="Search IP addresses..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="flex-1 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <select
            value={searchStatus}
            onChange={e => setSearchStatus(e.target.value)}
            className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Statuses</option>
            <option value="available">Available</option>
            <option value="assigned">Assigned</option>
            <option value="reserved">Reserved</option>
          </select>
          <button
            type="button"
            onClick={() => setShowAdvanced(v => !v)}
            className="px-3 py-2 text-sm border rounded hover:bg-gray-50 text-gray-600"
          >
            {showAdvanced ? 'Hide Filters' : 'More Filters'}
          </button>
          {searchableFields.length > 0 && (
            <button
              type="button"
              onClick={addCfFilterRow}
              className="px-3 py-2 text-sm border rounded hover:bg-gray-50 text-gray-600"
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
          {(isSearchActive || searchQuery || searchStatus || Object.values(advFilters).some(Boolean) || cfFilterRows.length > 0) && (
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
              className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {searchableFields.map(d => <option key={d.name} value={d.name}>{d.label}</option>)}
            </select>
            <select
              value={row.op}
              onChange={e => updateCfFilterRow(idx, { op: e.target.value })}
              className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
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
              className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
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

        {showAdvanced && (
          <div className="border rounded p-4 bg-gray-50 grid grid-cols-2 gap-4 text-sm">
            <div>
              <label className="block text-gray-600 mb-1">Tag</label>
              <select
                value={advFilters.tag_id}
                onChange={e => setAdvFilters(f => ({ ...f, tag_id: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm"
              >
                <option value="">Any tag</option>
                {tags.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-gray-600 mb-1">Assigned Status</label>
              <select
                value={advFilters.is_assigned}
                onChange={e => setAdvFilters(f => ({ ...f, is_assigned: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm"
              >
                <option value="">Any</option>
                <option value="true">Assigned only</option>
                <option value="false">Not assigned</option>
              </select>
            </div>
            <div>
              <label className="block text-gray-600 mb-1">MAC Address</label>
              <input
                type="text"
                placeholder="partial match"
                value={advFilters.mac_address}
                onChange={e => setAdvFilters(f => ({ ...f, mac_address: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm font-mono"
              />
            </div>
            <div>
              <label className="block text-gray-600 mb-1">Hostname / PTR</label>
              <input
                type="text"
                placeholder="partial match"
                value={advFilters.ptr_record}
                onChange={e => setAdvFilters(f => ({ ...f, ptr_record: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm"
              />
            </div>
          </div>
        )}
      </div>

      {!isSearchActive && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">
          {total} address{total !== 1 ? 'es' : ''}
        </p>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              {col('address') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Address</th>}
              {col('hostname') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Hostname</th>}
              {col('status') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>}
              {col('tag') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Tag</th>}
              {col('assigned_to') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Assigned To</th>}
              {col('device') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Device</th>}
              {col('mac_address') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">MAC Address</th>}
              {col('dns_name') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">DNS Name</th>}
              {col('ptr_record') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">PTR / Hostname</th>}
              {col('last_seen') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Last Seen</th>}
              {searchableFields.map(d => (
                <th key={d.name} className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{d.label}</th>
              ))}
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {ips.length === 0 && (
              <tr><td colSpan={visibleCols.length + searchableFields.length + 1} className="px-4 py-6 text-center text-gray-400">No IP addresses yet</td></tr>
            )}
            {ips.map(ip => (
              <tr key={ip.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                {col('address') && <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{ip.Address}</td>}
                {col('hostname') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.Hostname || '—'}</td>}
                {col('status') && (
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[ip.Status] || 'bg-gray-100 text-gray-600'}`}>
                      {ip.Status}
                    </span>
                  </td>
                )}
                {col('tag') && <td className="px-4 py-3"><TagBadge tag={ip.Tag} /></td>}
                {col('assigned_to') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.AssignedTo || '—'}</td>}
                {col('device') && (
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    {ip.device_id ? (
                      <Link to={`/devices/${ip.device_id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                        {ip.device?.hostname || `#${ip.device_id}`}
                      </Link>
                    ) : '—'}
                  </td>
                )}
                {col('mac_address') && <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400 text-xs">{ip.MACAddress || '—'}</td>}
                {col('dns_name') && (
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    <span className="flex items-center gap-1">
                      {ip.dnsName || '—'}
                      {ip.dnsName && ip.dnsRecords && !ip.dnsRecords.includes(ip.Address) && (
                        <span
                          title="DNS mismatch: DNS records do not include this IP's address"
                          className="text-yellow-500 cursor-help"
                        >
                          &#9888;
                        </span>
                      )}
                    </span>
                  </td>
                )}
                {col('ptr_record') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.PTRRecord || ip.ptrRecord || '—'}</td>}
                {col('last_seen') && (
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                    {ip.LastSeen ? new Date(ip.LastSeen).toLocaleString() : '—'}
                  </td>
                )}
                {searchableFields.map(d => {
                  const val = ip.custom_fields?.[d.name]
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
                  <button onClick={() => openMeta(ip)} className="text-gray-400 hover:text-indigo-600 text-xs">Edit</button>
                  {ip.Status !== 'assigned' && (
                    <button onClick={() => openAssign(ip)} className="text-gray-400 hover:text-blue-600 text-xs">Assign</button>
                  )}
                  {ip.Status === 'assigned' && (
                    <button onClick={() => handleRelease(ip.id)} className="text-gray-400 hover:text-yellow-600 text-xs">Release</button>
                  )}
                  {deleteConfirm === ip.id ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => handleDelete(ip.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteConfirm(ip.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
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

      {modal === 'create' && (
        <Modal title="New IP Address" onClose={() => setModal(null)}>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">IP Address</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="192.168.0.10"
                value={form.address}
                onChange={e => setForm(f => ({ ...f, address: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Hostname</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="server01.example.com"
                value={form.hostname}
                onChange={e => setForm(f => ({ ...f, hostname: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.status}
                onChange={e => setForm(f => ({ ...f, status: e.target.value }))}
              >
                <option value="available">Available</option>
                <option value="assigned">Assigned</option>
                <option value="reserved">Reserved</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tag</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.tag_id}
                onChange={e => setForm(f => ({ ...f, tag_id: e.target.value }))}
              >
                <option value="">No tag</option>
                {tags.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">MAC Address</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="aa:bb:cc:dd:ee:ff"
                value={form.mac_address}
                onChange={e => setForm(f => ({ ...f, mac_address: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">PTR / Hostname</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="host.example.com"
                value={form.ptr_record}
                onChange={e => setForm(f => ({ ...f, ptr_record: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">DNS Name</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="web01.example.com"
                value={form.dns_name}
                onChange={e => setForm(f => ({ ...f, dns_name: e.target.value }))}
              />
            </div>
            {cfDefs.length > 0 && (
              <div className="border-t pt-4">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">Custom Fields</p>
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
                {saving ? 'Saving...' : 'Add IP'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal?.assign && (
        <Modal title={`Assign ${modal.assign.Address}`} onClose={() => setModal(null)}>
          <form onSubmit={handleAssign} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Assign To</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="server name or user"
                value={form.assigned_to}
                onChange={e => setForm(f => ({ ...f, assigned_to: e.target.value }))}
                required
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Assign'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal === 'requestIP' && (
        <Modal title="Request IP Address" onClose={() => setModal(null)}>
          {ipReqSuccess ? (
            <div className="py-4 text-center text-green-600 font-medium">Request submitted successfully!</div>
          ) : (
            <form onSubmit={handleIPRequestSubmit} className="space-y-4">
              {ipReqError && (
                <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{ipReqError}</div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Subnet</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm font-mono bg-gray-50 text-gray-500"
                  value={subnet ? `${subnet.networkAddress}/${subnet.prefixLength}` : `Subnet #${subnetID}`}
                  readOnly
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Specific IP <span className="text-gray-400 font-normal">(optional — leave blank for auto-assign)</span>
                </label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g. 192.168.1.42"
                  value={ipReqForm.specific_ip}
                  onChange={e => setIPReqForm(f => ({ ...f, specific_ip: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  DNS Name <span className="text-gray-400 font-normal">(optional)</span>
                </label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g. myserver.example.com"
                  value={ipReqForm.dns_name}
                  onChange={e => setIPReqForm(f => ({ ...f, dns_name: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Purpose <span className="text-red-500">*</span>
                </label>
                <textarea
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={3}
                  placeholder="Describe why you need this IP address..."
                  value={ipReqForm.purpose}
                  onChange={e => setIPReqForm(f => ({ ...f, purpose: e.target.value }))}
                  required
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
                <button type="submit" disabled={saving} className="px-4 py-2 bg-green-600 text-white rounded text-sm hover:bg-green-700 disabled:opacity-50">
                  {saving ? 'Submitting...' : 'Submit Request'}
                </button>
              </div>
            </form>
          )}
        </Modal>
      )}

      {modal?.meta && (
        <Modal title={`Edit ${modal.meta.Address}`} onClose={() => setModal(null)}>
          <form onSubmit={handleUpdateMeta} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tag</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.tag_id}
                onChange={e => setForm(f => ({ ...f, tag_id: e.target.value }))}
              >
                <option value="">No tag</option>
                {tags.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">MAC Address</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="aa:bb:cc:dd:ee:ff"
                value={form.mac_address}
                onChange={e => setForm(f => ({ ...f, mac_address: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">PTR / Hostname</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="host.example.com"
                value={form.ptr_record}
                onChange={e => setForm(f => ({ ...f, ptr_record: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">DNS Name</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="web01.example.com"
                value={form.dns_name}
                onChange={e => setForm(f => ({ ...f, dns_name: e.target.value }))}
              />
            </div>
            {(modal.meta.dnsLastChecked) && (
              <div className="border-t pt-3">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">DNS Info (read-only)</p>
                <div className="grid grid-cols-2 gap-2 text-xs text-gray-600">
                  <span className="font-medium">Last DNS Check:</span>
                  <span>{modal.meta.dnsLastChecked ? new Date(modal.meta.dnsLastChecked).toLocaleString() : '—'}</span>
                </div>
              </div>
            )}
            {cfDefs.length > 0 && (
              <div className="border-t pt-4">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">Custom Fields</p>
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
