import { useState, useEffect, useMemo, useCallback } from 'react'
import { Link } from 'react-router-dom'
import Modal from '../components/Modal'
import Pagination from '../components/Pagination'
import CustomFieldForm from '../components/CustomFieldForm'
import { api } from '../api/client'
import { getFeatures } from '../api/app'
import { normalizeFeatures } from '../utils/features'
import { getLocations } from '../api/locations'
import { getRacks } from '../api/racks'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'
import { downloadFile } from '../utils/download'
import { loadPrefs, savePrefs, loadColPrefs, saveColPrefs } from '../utils/listPrefs'
import { getCachedUser, LEGACY_STORAGE_KEYS, STORAGE_KEYS } from '../utils/storageKeys'
import vendorCatalog from '../data/vendors.json'

const FILTER_KEY = STORAGE_KEYS.deviceFilters
const LEGACY_FILTER_KEY = LEGACY_STORAGE_KEYS.deviceFilters
const COL_KEY = STORAGE_KEYS.deviceColumns
const LEGACY_COL_KEY = LEGACY_STORAGE_KEYS.deviceColumns
const DEFAULT_COLS = { vendor_model: true, location: true, ips: true, status: true }

const DEFAULT_LIMIT = 50

const EMPTY_FORM = { hostname: '', type_id: '', description: '', vendor: '', model: '', os_version: '', location_id: '', rack_id: '', rack_unit_start: '', rack_unit_size: '', custom_fields: {}, snmp_version: 'v2c', snmp_community: '', snmp_v3_user: '', snmp_v3_auth_proto: 'SHA', snmp_v3_auth_pass: '', snmp_v3_priv_proto: 'AES', snmp_v3_priv_pass: '' }

const COL_LABELS = { vendor_model: 'Vendor / Model', location: 'Location', ips: 'IPs', status: 'Status' }

function ColToggle({ cols, setCols, labels }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="relative">
      <button onClick={() => setOpen(o => !o)} className="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-sm flex items-center gap-1">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h8m-8 6h16" /></svg>
        Columns
      </button>
      {open && (
        <div className="absolute right-0 mt-1 w-44 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-lg z-10 py-1">
          {Object.entries(labels).map(([key, label]) => (
            <label key={key} className="flex items-center gap-2 px-3 py-1.5 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer text-sm text-gray-700 dark:text-gray-300">
              <input type="checkbox" checked={cols[key] !== false} onChange={e => setCols(c => ({ ...c, [key]: e.target.checked }))} className="rounded" />
              {label}
            </label>
          ))}
        </div>
      )}
    </div>
  )
}

export default function DevicesPage() {
  const [devices, setDevices] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [deviceTypes, setDeviceTypes] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [savedFilters] = useState(() => loadPrefs(FILTER_KEY, { filterHostname: '', filterTypeId: '', filterOnline: '', filterLocationId: '' }, LEGACY_FILTER_KEY))
  const [filterHostname, setFilterHostname] = useState(savedFilters.filterHostname)
  const [filterTypeId, setFilterTypeId] = useState(savedFilters.filterTypeId)
  const [filterOnline, setFilterOnline] = useState(savedFilters.filterOnline)
  const [isFiltered, setIsFiltered] = useState(false)
  const [cols, setCols] = useState(() => loadColPrefs(COL_KEY, DEFAULT_COLS, LEGACY_COL_KEY))
  const [locationsEnabled, setLocationsEnabled] = useState(true)
  const col = (name) => cols[name] !== false && (name !== 'location' || locationsEnabled)
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)
  const [downloading, setDownloading] = useState(false)

  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  async function handleExport() {
    setDownloading(true)
    try { await downloadFile('/api/v1/admin/reports/export/devices', 'devices.csv') }
    catch { setError('Export failed') }
    finally { setDownloading(false) }
  }
  const [cfDefs, setCfDefs] = useState([])
  const [cfFilterRows, setCfFilterRows] = useState([])
  const [locations, setLocations] = useState([])
  const [racks, setRacks] = useState([])
  const [filterLocationId, setFilterLocationId] = useState(savedFilters.filterLocationId)

  useEffect(() => {
    savePrefs(FILTER_KEY, { filterHostname, filterTypeId, filterOnline, filterLocationId }, LEGACY_FILTER_KEY)
  }, [filterHostname, filterTypeId, filterOnline, filterLocationId])

  useEffect(() => {
    saveColPrefs(COL_KEY, cols, LEGACY_COL_KEY)
  }, [cols])

  const loadLocations = useCallback(async () => {
    try {
      const data = await getLocations()
      setLocations(Array.isArray(data) ? data : (data?.locations ?? []))
    } catch {}
  }, [])

  async function loadRacksForLocation(locationId) {
    if (!locationId) { setRacks([]); return }
    try {
      const data = await getRacks(locationId)
      setRacks(Array.isArray(data) ? data : (data?.racks ?? []))
    } catch { setRacks([]) }
  }

  const loadCfDefs = useCallback(async () => {
    try {
      const res = await api.get('/admin/custom-fields', { params: { entity_type: 'device' } })
      setCfDefs(normalizeCustomFieldDefs(res.data || []))
    } catch {}
  }, [])

  const loadDeviceTypes = useCallback(async () => {
    try {
      const res = await api.get('/device-types')
      setDeviceTypes(Array.isArray(res.data) ? res.data : [])
    } catch {}
  }, [])

  const vendorSuggestions = useMemo(() => {
    const typeName = (deviceTypes.find(t => String(t.id) === String(form.type_id))?.name || '').toLowerCase()
    const cats = Object.values(vendorCatalog.categories)
    const matched = cats.find(c => c.keywords.some(k => typeName.includes(k))) || vendorCatalog.categories.other
    return Object.keys(matched.vendors)
  }, [form.type_id, deviceTypes])

  const modelSuggestions = useMemo(() => {
    const typeName = (deviceTypes.find(t => String(t.id) === String(form.type_id))?.name || '').toLowerCase()
    const cats = Object.values(vendorCatalog.categories)
    const matched = cats.find(c => c.keywords.some(k => typeName.includes(k))) || vendorCatalog.categories.other
    const models = matched.vendors[form.vendor] || []
    return models
  }, [form.type_id, form.vendor, deviceTypes])

  const load = useCallback(async (p) => {
    try {
      setLoading(true)
      const res = await api.get('/devices', { params: { page: p, limit: DEFAULT_LIMIT } })
      const data = res.data
      setDevices(getDeviceRows(data))
      setTotal(data.total ?? 0)
      setPage(p)
    } catch {
      setError('Failed to load devices')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadDeviceTypes()
    load(1)
    loadCfDefs()
    getFeatures().then(res => {
      const f = normalizeFeatures(res.data)
      setLocationsEnabled(f.locations !== false)
      if (f.locations !== false) loadLocations()
    }).catch(() => loadLocations())
  }, [loadDeviceTypes, load, loadCfDefs, loadLocations])

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
    const body = {}
    if (filterHostname.trim()) body.query = filterHostname.trim()
    if (filterTypeId) body.type_id = parseInt(filterTypeId)
    if (filterOnline !== '') body.is_online = filterOnline === 'true'
    if (filterLocationId) body.location_id = parseInt(filterLocationId)
    const cfFilters = {}
    cfFilterRows.forEach(r => { if (r.value.trim()) cfFilters[r.field] = r.value.trim() })
    if (Object.keys(cfFilters).length) body.custom_fields = cfFilters

    if (!Object.keys(body).length) {
      setIsFiltered(false)
      load(1)
      return
    }

    try {
      setLoading(true)
      setIsFiltered(true)
      const res = await api.post('/devices/search', body)
      const rows = getDeviceRows(res.data)
      setDevices(rows)
      setTotal(rows.length)
      setPage(1)
    } catch {
      setError('Failed to search devices')
    } finally {
      setLoading(false)
    }
  }

  function handleClearSearch() {
    setFilterHostname('')
    setFilterTypeId('')
    setFilterOnline('')
    setFilterLocationId('')
    setCfFilterRows([])
    setIsFiltered(false)
    savePrefs(FILTER_KEY, { filterHostname: '', filterTypeId: '', filterOnline: '', filterLocationId: '' }, LEGACY_FILTER_KEY)
    load(1)
  }

  function handlePageChange(newPage) {
    setPage(newPage)
    load(newPage)
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(device) {
    const locId = device.locationId ? String(device.locationId) : ''
    if (locId) loadRacksForLocation(locId)
    setForm({
      hostname: device.hostname || '',
      type_id: device.typeId ? String(device.typeId) : '',
      description: device.description || '',
      vendor: device.vendor || '',
      model: device.model || '',
      os_version: device.osVersion || '',
      location_id: locId,
      rack_id: device.rackId ? String(device.rackId) : '',
      rack_unit_start: device.rackUnitStart != null ? String(device.rackUnitStart) : '',
      rack_unit_size: device.rackUnitSize != null ? String(device.rackUnitSize) : '',
      custom_fields: device.customFields || {},
    })
    setModal({ edit: device })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        hostname: form.hostname,
        type_id: form.type_id ? parseInt(form.type_id) : null,
        description: form.description || null,
        vendor: form.vendor || null,
        model: form.model || null,
        os_version: form.os_version || null,
        location_id: form.location_id ? parseInt(form.location_id) : null,
        rack_id: form.rack_id ? parseInt(form.rack_id) : null,
        rack_unit_start: form.rack_unit_start ? parseInt(form.rack_unit_start) : null,
        rack_unit_size: form.rack_unit_size ? parseInt(form.rack_unit_size) : null,
        custom_fields: form.custom_fields || {},
        snmp_version: form.snmp_version || 'v2c',
        snmp_community: form.snmp_community || null,
        snmp_v3_user: form.snmp_v3_user || null,
        snmp_v3_auth_proto: form.snmp_v3_auth_proto || null,
        snmp_v3_auth_pass: form.snmp_v3_auth_pass || null,
        snmp_v3_priv_proto: form.snmp_v3_priv_proto || null,
        snmp_v3_priv_pass: form.snmp_v3_priv_pass || null,
      }
      if (modal === 'create') {
        await api.post('/devices', body)
      } else {
        const id = modal.edit.id
        await api.put(`/devices/${id}`, body)
      }
      setModal(null)
      load(page)
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to save device')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await api.delete(`/devices/${id}`)
      setDeleteConfirm(null)
      load(page)
    } catch {
      setError('Failed to delete device')
    }
  }

  if (loading && devices.length === 0) return <PageSpinner message="Loading devices..." />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Devices</h1>
        <div className="flex items-center gap-2">
          <ColToggle cols={cols} setCols={setCols} labels={locationsEnabled ? COL_LABELS : Object.fromEntries(Object.entries(COL_LABELS).filter(([k]) => k !== 'location'))} />
          {isAdmin && (
            <button onClick={handleExport} disabled={downloading} className="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-sm disabled:opacity-50">
              {downloading ? 'Exporting...' : 'Export CSV'}
            </button>
          )}
          <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
            + Add Device
          </button>
        </div>
      </div>

      <ErrorBanner error={error} />

      <div className="mb-4 space-y-2">
        <form onSubmit={handleSearch} className="flex gap-2 flex-wrap">
          <input
            type="text"
            placeholder="Search hostname..."
            value={filterHostname}
            onChange={e => setFilterHostname(e.target.value)}
            className="flex-1 min-w-40 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
          />
          <select
            value={filterTypeId}
            onChange={e => setFilterTypeId(e.target.value)}
            className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
          >
            <option value="">All Types</option>
            {deviceTypes.map(t => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
          <select
            value={filterOnline}
            onChange={e => setFilterOnline(e.target.value)}
            className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
          >
            <option value="">Online/Offline</option>
            <option value="true">Online only</option>
            <option value="false">Offline only</option>
          </select>
          {locationsEnabled && locations.length > 0 && (
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
            disabled={loading}
            className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm font-medium disabled:opacity-50"
          >
            Search
          </button>
          {(isFiltered || filterHostname || filterTypeId || filterOnline || filterLocationId || cfFilterRows.length > 0) && (
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

      {!isFiltered && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">
          {total} device{total !== 1 ? 's' : ''}
        </p>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Hostname</th>
              {col('vendor_model') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Vendor / Model</th>}
              {col('location') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Location</th>}
              {col('ips') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">IPs</th>}
              {col('status') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>}
              {searchableFields.map(f => (
                <th key={f.name} className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{f.label}</th>
              ))}
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {devices.length === 0 && (
              <EmptyRow colSpan={2 + [col('vendor_model'),col('location'),col('ips'),col('status')].filter(Boolean).length + 1 + searchableFields.length} message="No devices yet." />
            )}

            {devices.map(d => (
              <tr key={d.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  <span title={d.type?.name || ''}>
                    {d.type?.icon || '🖥️'}
                  </span>
                </td>
                <td className="px-4 py-3 font-medium">
                  <Link to={`/devices/${d.id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                    {d.hostname}
                  </Link>
                </td>
                {col('vendor_model') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {[d.vendor, d.model].filter(Boolean).join(' / ') || '—'}
                </td>}
                {col('location') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {d.locationId ? (
                    <Link to={`/locations/${d.locationId}`} className="text-blue-600 dark:text-blue-400 hover:underline text-xs">
                      {locations.find(l => l.id === d.locationId)?.name || `#${d.locationId}`}
                    </Link>
                  ) : '—'}
                </td>}
                {col('ips') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {d.ipCount ?? 0}
                </td>}
                {col('status') && <td className="px-4 py-3">
                  <span className={`inline-flex items-center gap-1.5 text-xs font-medium`}>
                    <span className={`w-2 h-2 rounded-full ${d.isOnline ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                    <span className={d.isOnline ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}>
                      {d.isOnline ? 'Online' : 'Offline'}
                    </span>
                  </span>
                </td>}
                {searchableFields.map(f => {
                  const val = d.customFields?.[f.name]
                  return (
                    <td key={f.name} className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {val ? (
                        <button
                          className="hover:text-blue-600 dark:hover:text-blue-400 underline decoration-dotted text-left"
                          onClick={() => addCfFilterFromValue(f.name, val)}
                          title="Filter by this value"
                        >
                          {val}
                        </button>
                      ) : '—'}
                    </td>
                  )
                })}
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(d)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                  {deleteConfirm === d.id ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => handleDelete(d.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteConfirm(d.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {!isFiltered && total > DEFAULT_LIMIT && (
        <Pagination
          page={page}
          limit={DEFAULT_LIMIT}
          total={total}
          onChange={handlePageChange}
        />
      )}

      {modal && (
        <Modal
          title={modal === 'create' ? 'Add Device' : 'Edit Device'}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Hostname <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="router01.example.com"
                value={form.hostname}
                onChange={e => setForm(f => ({ ...f, hostname: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.type_id}
                onChange={e => setForm(f => ({ ...f, type_id: e.target.value }))}
              >
                <option value="">No type</option>
                {deviceTypes.map(t => (
                  <option key={t.id} value={t.id}>{t.name}</option>
                ))}
              </select>
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
            <datalist id="vendor-list">
              {vendorSuggestions.map(v => <option key={v} value={v} />)}
            </datalist>
            <datalist id="model-list">
              {modelSuggestions.map(m => <option key={m} value={m} />)}
            </datalist>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Vendor</label>
                <input
                  list="vendor-list"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="Cisco"
                  value={form.vendor}
                  onChange={e => setForm(f => ({ ...f, vendor: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Model</label>
                <input
                  list="model-list"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="ASR-1000"
                  value={form.model}
                  onChange={e => setForm(f => ({ ...f, model: e.target.value }))}
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">OS Version</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="IOS 15.4"
                value={form.os_version}
                onChange={e => setForm(f => ({ ...f, os_version: e.target.value }))}
              />
            </div>
            {locationsEnabled && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Location (optional)</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.location_id}
                onChange={e => {
                  const locId = e.target.value
                  setForm(f => ({ ...f, location_id: locId, rack_id: '', rack_unit_start: '', rack_unit_size: '' }))
                  loadRacksForLocation(locId)
                }}
              >
                <option value="">No location</option>
                {locations.map(l => (
                  <option key={l.id} value={l.id}>{l.name}</option>
                ))}
              </select>
            </div>
            )}
            {locationsEnabled && form.location_id && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Rack (optional)</label>
                  <select
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    value={form.rack_id}
                    onChange={e => setForm(f => ({ ...f, rack_id: e.target.value }))}
                  >
                    <option value="">No rack</option>
                    {racks.map(r => (
                      <option key={r.id} value={r.id}>{r.name}</option>
                    ))}
                  </select>
                </div>
                {form.rack_id && (
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Rack Unit Start</label>
                      <input
                        type="number"
                        min="1"
                        className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                        placeholder="1"
                        value={form.rack_unit_start}
                        onChange={e => setForm(f => ({ ...f, rack_unit_start: e.target.value }))}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Rack Unit Size</label>
                      <input
                        type="number"
                        min="1"
                        className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                        placeholder="1"
                        value={form.rack_unit_size}
                        onChange={e => setForm(f => ({ ...f, rack_unit_size: e.target.value }))}
                      />
                    </div>
                  </div>
                )}
              </>
            )}
            <div className="border-t dark:border-gray-600 pt-4">
              <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">SNMP</p>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">SNMP Version</label>
                  <select
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    value={form.snmp_version}
                    onChange={e => setForm(f => ({ ...f, snmp_version: e.target.value }))}
                  >
                    <option value="v1">v1</option>
                    <option value="v2c">v2c</option>
                    <option value="v3">v3</option>
                  </select>
                </div>
                {(form.snmp_version === 'v1' || form.snmp_version === 'v2c') && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Community String</label>
                    <input
                      type="text"
                      className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                      placeholder="public"
                      value={form.snmp_community}
                      onChange={e => setForm(f => ({ ...f, snmp_community: e.target.value }))}
                    />
                  </div>
                )}
                {form.snmp_version === 'v3' && (
                  <>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username</label>
                      <input
                        type="text"
                        className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                        value={form.snmp_v3_user}
                        onChange={e => setForm(f => ({ ...f, snmp_v3_user: e.target.value }))}
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Auth Protocol</label>
                        <select
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          value={form.snmp_v3_auth_proto}
                          onChange={e => setForm(f => ({ ...f, snmp_v3_auth_proto: e.target.value }))}
                        >
                          <option value="SHA">SHA</option>
                          <option value="MD5">MD5</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Auth Password</label>
                        <input
                          type="password"
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          value={form.snmp_v3_auth_pass}
                          onChange={e => setForm(f => ({ ...f, snmp_v3_auth_pass: e.target.value }))}
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Priv Protocol</label>
                        <select
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          value={form.snmp_v3_priv_proto}
                          onChange={e => setForm(f => ({ ...f, snmp_v3_priv_proto: e.target.value }))}
                        >
                          <option value="AES">AES</option>
                          <option value="DES">DES</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Priv Password</label>
                        <input
                          type="password"
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          value={form.snmp_v3_priv_pass}
                          onChange={e => setForm(f => ({ ...f, snmp_v3_priv_pass: e.target.value }))}
                        />
                      </div>
                    </div>
                  </>
                )}
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

function getDeviceRows(data) {
  if (Array.isArray(data)) return data
  return data?.data ?? data?.devices ?? []
}

function normalizeCustomFieldDefs(defs) {
  const rows = Array.isArray(defs) ? defs : []
  return rows.map(def => ({
    ...def,
    // snake_case aliases (used by filters and internal logic)
    field_type: def.field_type ?? def.fieldType,
    is_required: def.is_required ?? def.isRequired,
    default_value: def.default_value ?? def.defaultValue,
    display_order: def.display_order ?? def.displayOrder,
    is_searchable: def.is_searchable ?? def.isSearchable,
    // camelCase aliases (used by CustomFieldForm)
    fieldType: def.fieldType ?? def.field_type,
    isRequired: def.isRequired ?? def.is_required,
    defaultValue: def.defaultValue ?? def.default_value,
    displayOrder: def.displayOrder ?? def.display_order,
    isSearchable: def.isSearchable ?? def.is_searchable,
  }))
}
