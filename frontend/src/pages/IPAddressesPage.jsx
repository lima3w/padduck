import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { getSubnet, getIPAddressesPaginated, createIPAddress, assignIPAddress, assignIPAddressWithLease, releaseIPAddress, releaseExpiredLease, deleteIPAddress, searchIPAddresses, getTags, updateIPMeta, bulkReleaseIPs, bulkDeleteIPs, pushDHCPReservation, removeDHCPReservation } from '../api/ipam'
import { getCustomFields } from '../api/admin'
import { submitIPRequest } from '../api/requests'
import { getDevices, associateDeviceIP, disassociateDeviceIP } from '../api/devices'
import Modal from '../components/Modal'
import Pagination from '../components/Pagination'
import TagBadge from '../components/TagBadge'
import CustomFieldForm from '../components/CustomFieldForm'
import ObservedStatePanel from '../components/ObservedStatePanel'
import { downloadFile } from '../utils/download'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'
import { getCachedUser, getStoredItem, LEGACY_STORAGE_KEYS, setStoredItem, STORAGE_KEYS } from '../utils/storageKeys'
import DelegationsTab from './ip/DelegationsTab'
import PortBadges from './ip/PortBadges'
import SortTh from './ip/SortTh'
import UtilisationHistorySection from './ip/UtilisationHistorySection'

const DEFAULT_LIMIT = 25
const PAGE_SIZE_OPTIONS = [10, 25, 50, 100]

const STATUS_COLORS = {
  available: 'bg-green-100 text-green-700',
  assigned: 'bg-blue-100 text-blue-700',
  reserved: 'bg-yellow-100 text-yellow-700',
}

const COLUMN_KEYS = ['address', 'hostname', 'status', 'tag', 'device', 'mac_address', 'dns_name', 'ptr_record', 'last_seen', 'services']
const COLUMN_LABEL_KEYS = {
  address: 'deviceIp.address',
  hostname: 'dashboard.hostname',
  status: 'delegations.status',
  tag: 'adminTags.tagColumn',
  device: 'ipAddressesPage.columnDevice',
  mac_address: 'ipAddressesPage.columnMacAddress',
  dns_name: 'myRequests.dnsName',
  ptr_record: 'ipAddressesPage.columnPickerPtrRecord',
  last_seen: 'adminAgents.lastSeen',
  services: 'ipAddressesPage.columnServices',
}
const DEFAULT_VISIBLE = ['address', 'hostname', 'status', 'tag', 'device']

const LS_KEY = STORAGE_KEYS.ipColumns
const LEGACY_LS_KEY = LEGACY_STORAGE_KEYS.ipColumns

function loadColumnVisibility() {
  try {
    const saved = JSON.parse(getStoredItem(LS_KEY, LEGACY_LS_KEY))
    if (saved && Array.isArray(saved)) return saved
  } catch {}
  return DEFAULT_VISIBLE
}

const IP_REQUEST_EMPTY = { specific_ip: '', dns_name: '', purpose: '' }

export default function IPAddressesPage() {
  const { t } = useTranslation()
  const COLUMN_LABELS = Object.fromEntries(Object.entries(COLUMN_LABEL_KEYS).map(([k, key]) => [k, t(key)]))
  const { subnetID } = useParams()
  const user = getCachedUser()
  const canAssignIP = user?.role === 'admin'
  const isAdmin = canAssignIP

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
  const [createError, setCreateError] = useState(null)
  const [ipReqSuccess, setIPReqSuccess] = useState(false)
  const [activeTab, setActiveTab] = useState('ips') // 'ips' | 'delegations'
  const [downloading, setDownloading] = useState(false)
  const [selected, setSelected] = useState(new Set())
  const [bulkReleasing, setBulkReleasing] = useState(false)
  const [bulkDeleting, setBulkDeleting] = useState(false)
  const [bulkDeleteConfirm, setBulkDeleteConfirm] = useState(false)
  const [devices, setDevices] = useState([])
  const [sortCol, setSortCol] = useState('')
  const [sortDir, setSortDir] = useState('asc')
  const [fullRange, setFullRange] = useState(false)
  const [pageSize, setPageSize] = useState(DEFAULT_LIMIT)
  const [metaDeviceSearch, setMetaDeviceSearch] = useState('')

  const loadCfDefs = useCallback(async () => {
    try {
      const res = await getCustomFields('ip_address')
      setCfDefs(Array.isArray(res.data) ? res.data : [])
    } catch {}
  }, [])

  const loadDevices = useCallback(async () => {
    try {
      const res = await getDevices({ limit: 1000 })
      setDevices(res.data?.data ?? res.data ?? [])
    } catch {}
  }, [])

  const load = useCallback(async (p = page, col = sortCol, dir = sortDir, full = fullRange, limit = pageSize) => {
    try {
      setLoading(true)
      setSelected(new Set())
      setSearchQuery('')
      setSearchStatus('')
      setIsSearchActive(false)
      const [subRes, ipRes, tagRes] = await Promise.all([
        getSubnet(subnetID),
        getIPAddressesPaginated(subnetID, p, limit, col, dir, true, full),
        getTags(),
      ])
      setSubnet(subRes.data)
      const data = ipRes.data
      setIPs(data.data ?? data)
      setTotal(data.total ?? (Array.isArray(data) ? data.length : 0))
      setTags(tagRes.data || [])
    } catch {
      setError(t('ipAddressesPage.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [subnetID, page, sortCol, sortDir, fullRange, pageSize, t])

  useEffect(() => {
    setPage(1)
    setIsSearchActive(false)
    load(1, sortCol, sortDir, fullRange, pageSize)
    loadCfDefs()
    loadDevices()
    // Intentionally keyed on subnetID only: load/loadCfDefs/loadDevices now change identity
    // whenever page/sort/fullRange/pageSize change (load reads them as default params), and
    // this effect must only reset to page 1 when navigating to a different subnet, not on
    // every such change.
  }, [subnetID])

  function handlePageChange(newPage) {
    setPage(newPage)
    load(newPage, sortCol, sortDir, fullRange, pageSize)
  }

  function handleSort(col) {
    const newDir = sortCol === col && sortDir === 'asc' ? 'desc' : 'asc'
    setSortCol(col)
    setSortDir(newDir)
    setFullRange(false)
    setPage(1)
    load(1, col, newDir, false, pageSize)
  }

  function handleFullRange(checked) {
    setFullRange(checked)
    setPage(1)
    load(1, sortCol, sortDir, checked, pageSize)
  }

  function handlePageSize(size) {
    setPageSize(size)
    setPage(1)
    load(1, sortCol, sortDir, fullRange, size)
  }

  function toggleColumn(col) {
    const next = visibleCols.includes(col)
      ? visibleCols.filter(c => c !== col)
      : [...visibleCols, col]
    // always keep address
    const final = next.includes('address') ? next : ['address', ...next]
    setVisibleCols(final)
    setStoredItem(LS_KEY, JSON.stringify(final), LEGACY_LS_KEY)
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
      setError(t('ipAddressesPage.searchFailed'))
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
    setSelected(new Set())
    load(1)
  }

  function filterMACInput(val) {
    // Strip invalid chars, collapse consecutive separators, cap at 17 chars (aa:bb:cc:dd:ee:ff)
    return val
      .replace(/[^0-9a-fA-F:\-.\s]/g, '')
      .replace(/([:\-.\s]){2,}/g, '$1')
      .slice(0, 17)
  }

  function normalizeMAC(val) {
    const stripped = val.replace(/[:\-.\s]/g, '').toLowerCase()
    if (stripped === '') return ''
    if (stripped.length !== 12 || !/^[0-9a-f]{12}$/.test(stripped)) return val
    return stripped.match(/.{2}/g).join(':')
  }

  function ipInSubnet(ip, networkAddress, prefixLength) {
    const toNum = (addr) => addr.split('.').reduce((acc, o) => ((acc << 8) | Number(o)) >>> 0, 0)
    const mask = prefixLength === 0 ? 0 : ((-1 << (32 - prefixLength)) >>> 0)
    return (toNum(ip) & mask) === (toNum(networkAddress) & mask)
  }

  function networkPrefix(networkAddress, prefixLength) {
    if (!networkAddress) return ''
    if (networkAddress.includes(':')) return networkAddress  // IPv6: use as-is
    const octets = networkAddress.split('.')
    if (prefixLength <= 8) return octets[0] + '.'
    if (prefixLength <= 16) return octets.slice(0, 2).join('.') + '.'
    return octets.slice(0, 3).join('.') + '.'
  }

  function openCreate(prefillAddress) {
    const addr = prefillAddress || networkPrefix(subnet?.networkAddress, subnet?.prefixLength)
    setForm({ address: addr, hostname: '', status: 'available', device_id: '', tag_id: '', mac_address: '', ptr_record: '', dns_name: '', custom_fields: {} })
    setCreateError(null)
    setModal('create')
  }

  function openAssign(ip) {
    setForm({ device_id: '', tag_id: '', mac_address: '', ptr_record: '', lease_duration_days: '' })
    setModal({ assign: ip })
  }

  function openMeta(ip) {
    setForm({
      hostname: ip.hostname || '',
      tag_id: ip.tagId ? String(ip.tagId) : '',
      mac_address: ip.macAddress || '',
      ptr_record: ip.ptrRecord || '',
      dns_name: ip.dnsName || '',
      custom_fields: ip.customFields || {},
      device_id: ip.deviceId ? String(ip.deviceId) : '',
    })
    setMetaDeviceSearch('')
    setModal({ meta: ip })
  }

  async function handleCreate(e) {
    e.preventDefault()
    if (subnet && !form.address.includes(':') && !ipInSubnet(form.address, subnet.networkAddress, subnet.prefixLength)) {
      setCreateError(t('ipAddressesPage.ipMustBeWithin', { network: subnet.networkAddress, prefix: subnet.prefixLength }))
      return
    }
    setCreateError(null)
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
      setCreateError(err.response?.data?.error || t('ipAddressesPage.createFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleAssign(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const days = parseInt(form.lease_duration_days)
      const deviceId = form.device_id ? parseInt(form.device_id) : null
      if (days > 0) {
        await assignIPAddressWithLease(modal.assign.id, { device_id: deviceId, lease_duration_days: days })
      } else {
        await assignIPAddress(modal.assign.id, { device_id: deviceId })
      }
      setModal(null)
      load(page)
    } catch {
      setError(t('ipAddressesPage.assignFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleReleaseExpired(id) {
    try {
      await releaseExpiredLease(id)
      load(page)
    } catch {
      setError(t('ipAddressesPage.releaseExpiredFailed'))
    }
  }

  async function handleUpdateMeta(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const oldDeviceId = modal.meta.deviceId || null
      const newDeviceId = form.device_id ? parseInt(form.device_id) : null
      if (newDeviceId !== oldDeviceId) {
        if (newDeviceId) {
          await associateDeviceIP(newDeviceId, modal.meta.id, {})
        } else if (oldDeviceId) {
          await disassociateDeviceIP(oldDeviceId, modal.meta.id)
        }
      }
      await updateIPMeta(modal.meta.id, {
        hostname: form.hostname || '',
        tag_id: form.tag_id ? parseInt(form.tag_id) : null,
        mac_address: form.mac_address || null,
        ptr_record: form.ptr_record || null,
        dns_name: form.dns_name || null,
        custom_fields: form.custom_fields || {},
      })
      setModal(null)
      load(page)
    } catch(err) {
      setError(err.response?.data?.error || t('ipAddressesPage.updateFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handlePushReservation(id) {
    try {
      await pushDHCPReservation(id)
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('ipAddressesPage.pushReservationFailed'))
    }
  }

  async function handleRemoveReservation(id) {
    try {
      await removeDHCPReservation(id)
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('ipAddressesPage.removeReservationFailed'))
    }
  }

  async function handleRelease(id) {
    try {
      await releaseIPAddress(id)
      load(page)
    } catch {
      setError(t('ipAddressesPage.releaseFailed'))
    }
  }

  async function handleDelete(id) {
    try {
      await deleteIPAddress(id)
      setDeleteConfirm(null)
      load(page)
    } catch {
      setError(t('ipAddressesPage.deleteFailed'))
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
        setIPReqError(t('ipAddressesPage.ipTakenError'))
      } else {
        setIPReqError(err.response?.data?.error || t('ipAddressesPage.submitRequestFailed'))
      }
    } finally {
      setSaving(false)
    }
  }

  async function handleExportIPs() {
    setDownloading(true)
    try {
      await downloadFile(`/api/v1/admin/reports/export/ips?format=csv&subnet_id=${subnetID}`, `ips-subnet-${subnetID}.csv`)
    } catch {
      setError(t('exportData.exportFailed'))
    } finally {
      setDownloading(false)
    }
  }

  function toggleSelect(id) {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  async function handleBulkRelease() {
    if (selected.size === 0) return
    setBulkReleasing(true)
    try {
      await bulkReleaseIPs(Array.from(selected))
      setSelected(new Set())
      load(page)
    } catch {
      setError(t('ipAddressesPage.bulkReleaseFailed'))
    } finally {
      setBulkReleasing(false)
    }
  }

  async function handleBulkDelete() {
    if (selected.size === 0) return
    setBulkDeleting(true)
    setBulkDeleteConfirm(false)
    try {
      await bulkDeleteIPs(Array.from(selected))
      setSelected(new Set())
      load(page)
    } catch {
      setError(t('ipAddressesPage.bulkDeleteFailed'))
    } finally {
      setBulkDeleting(false)
    }
  }

  if (loading) return <PageSpinner message={t('ipAddressesPage.loadingIPs')} />

  const col = (key) => visibleCols.includes(key)

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/networks" className="hover:text-blue-600">{t('nav.networks')}</Link>
        <span>/</span>
        {subnet && (
          <Link to={`/networks/${subnet.networkId}/subnets`} className="hover:text-blue-600">{t('dashboard.subnets')}</Link>
        )}
        <span>/</span>
        <span className="text-gray-800 font-medium font-mono">{subnet?.networkAddress}/{subnet?.prefixLength}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('ipAddressesPage.title')}</h1>
          <div className="flex border border-gray-200 dark:border-gray-600 rounded overflow-hidden text-xs">
            <button
              onClick={() => setActiveTab('ips')}
              className={`px-3 py-1.5 transition ${activeTab === 'ips' ? 'bg-blue-600 text-white' : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'}`}
            >
              {t('ipAddressesPage.ipsTab')}
            </button>
            <button
              onClick={() => setActiveTab('delegations')}
              className={`px-3 py-1.5 transition border-l border-gray-200 dark:border-gray-600 ${activeTab === 'delegations' ? 'bg-blue-600 text-white' : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'}`}
            >
              {t('ipAddressesPage.delegationsTab')}
            </button>
          </div>
        </div>
        <div className="flex gap-2 items-center">
          <div className="relative">
            <button
              onClick={() => setShowColPicker(v => !v)}
              className="px-3 py-2 bg-gray-100 text-gray-600 rounded hover:bg-gray-200 text-sm"
              title={t('ipAddressesPage.toggleColumnsTitle')}
            >
              {t('devicesPage.columns')}
            </button>
            {showColPicker && (
              <div className="absolute right-0 top-9 bg-white border rounded shadow-lg z-10 p-3 min-w-max">
                <p className="text-xs font-medium text-gray-500 mb-2">{t('ipAddressesPage.showHideColumns')}</p>
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
          <button
            onClick={handleExportIPs}
            disabled={downloading}
            className="px-3 py-2 bg-gray-100 text-gray-600 rounded hover:bg-gray-200 text-sm disabled:opacity-50"
            title={t('ipAddressesPage.exportCsvTitle')}
          >
            {downloading ? t('devicesPage.exporting') : t('devicesPage.exportCsv')}
          </button>
          {!canAssignIP && (
            <button onClick={openIPRequest} className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 text-sm font-medium">
              {t('ipAddressesPage.requestIp')}
            </button>
          )}
          {canAssignIP && (
            <button onClick={() => openCreate()} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
              {t('ipAddressesPage.newIp')}
            </button>
          )}
        </div>
      </div>

      <ErrorBanner error={error} />

      {activeTab === 'delegations' && <DelegationsTab subnetId={subnetID} />}

      {activeTab === 'ips' && <>{/* data quality network removed */}
      <div className="mb-4 space-y-2">
        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            type="text"
            placeholder={t('ipAddressesPage.searchPlaceholder')}
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="flex-1 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <select
            value={searchStatus}
            onChange={e => setSearchStatus(e.target.value)}
            className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">{t('adminRequests.allStatuses')}</option>
            <option value="available">{t('ipAddressesPage.available')}</option>
            <option value="assigned">{t('adminRoles.assignedLabel')}</option>
            <option value="reserved">{t('dhcp.reserved')}</option>
          </select>
          <button
            type="button"
            onClick={() => setShowAdvanced(v => !v)}
            className="px-3 py-2 text-sm border rounded hover:bg-gray-50 text-gray-600"
          >
            {showAdvanced ? t('ipAddressesPage.hideFilters') : t('ipAddressesPage.moreFilters')}
          </button>
          {searchableFields.length > 0 && (
            <button
              type="button"
              onClick={addCfFilterRow}
              className="px-3 py-2 text-sm border rounded hover:bg-gray-50 text-gray-600"
            >
              {t('devicesPage.addFilter')}
            </button>
          )}
          <button
            type="submit"
            disabled={searching}
            className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm font-medium disabled:opacity-50"
          >
            {searching ? t('ipAddressesPage.searching') : t('devicesPage.search')}
          </button>
          {(isSearchActive || searchQuery || searchStatus || Object.values(advFilters).some(Boolean) || cfFilterRows.length > 0) && (
            <button
              type="button"
              onClick={handleClearSearch}
              className="px-4 py-2 bg-gray-400 text-white rounded hover:bg-gray-500 text-sm font-medium"
            >
              {t('common.clear')}
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
              <option value="is">{t('devicesPage.operatorIs')}</option>
              <option value="contains">{t('devicesPage.operatorContains')}</option>
              <option value="is not">{t('devicesPage.operatorIsNot')}</option>
            </select>
            <input
              type="text"
              value={row.value}
              onChange={e => updateCfFilterRow(idx, { value: e.target.value })}
              placeholder={t('adminCustomFields.valuePlaceholder')}
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
              <label className="block text-gray-600 mb-1">{t('adminTags.tagColumn')}</label>
              <select
                value={advFilters.tag_id}
                onChange={e => setAdvFilters(f => ({ ...f, tag_id: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm"
              >
                <option value="">{t('ipAddressesPage.anyTag')}</option>
                {tags.map(tag => <option key={tag.id} value={tag.id}>{tag.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-gray-600 mb-1">{t('ipAddressesPage.assignedStatusLabel')}</label>
              <select
                value={advFilters.is_assigned}
                onChange={e => setAdvFilters(f => ({ ...f, is_assigned: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm"
              >
                <option value="">{t('ipAddressesPage.any')}</option>
                <option value="true">{t('ipAddressesPage.assignedOnly')}</option>
                <option value="false">{t('ipAddressesPage.notAssigned')}</option>
              </select>
            </div>
            <div>
              <label className="block text-gray-600 mb-1">{t('ipAddressesPage.columnMacAddress')}</label>
              <input
                type="text"
                placeholder={t('ipAddressesPage.partialMatchPlaceholder')}
                value={advFilters.mac_address}
                onChange={e => setAdvFilters(f => ({ ...f, mac_address: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm font-mono"
              />
            </div>
            <div>
              <label className="block text-gray-600 mb-1">{t('ipAddressesPage.hostnamePtrLabel')}</label>
              <input
                type="text"
                placeholder={t('ipAddressesPage.partialMatchPlaceholder')}
                value={advFilters.ptr_record}
                onChange={e => setAdvFilters(f => ({ ...f, ptr_record: e.target.value }))}
                className="w-full border rounded px-3 py-1.5 text-sm"
              />
            </div>
          </div>
        )}
      </div>

      {selected.size > 0 && isAdmin && (
        <div className="mb-3 flex items-center gap-3 p-2 bg-blue-50 dark:bg-blue-900/20 rounded border border-blue-200 dark:border-blue-800 text-sm">
          <span className="text-blue-700 dark:text-blue-300 font-medium">{t('adminUsersPage.selectedCount', { count: selected.size })}</span>
          <button onClick={handleBulkRelease} disabled={bulkReleasing || bulkDeleting} className="px-3 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700 disabled:opacity-50">
            {bulkReleasing ? t('ipAddressesPage.releasing') : t('ipAddressesPage.releaseSelected')}
          </button>
          {bulkDeleteConfirm ? (
            <>
              <span className="text-red-600 dark:text-red-400 text-xs font-medium">{t('ipAddressesPage.deleteCount', { count: selected.size })}</span>
              <button onClick={handleBulkDelete} disabled={bulkDeleting} className="px-3 py-1 bg-red-600 text-white rounded text-xs hover:bg-red-700 disabled:opacity-50">
                {bulkDeleting ? t('ipAddressesPage.deleting') : t('ipAddressesPage.confirmDeleteButton')}
              </button>
              <button onClick={() => setBulkDeleteConfirm(false)} className="px-3 py-1 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 text-xs">{t('common.cancel')}</button>
            </>
          ) : (
            <button onClick={() => setBulkDeleteConfirm(true)} disabled={bulkDeleting} className="px-3 py-1 bg-red-600 text-white rounded text-xs hover:bg-red-700 disabled:opacity-50">
              {t('ipAddressesPage.deleteSelected')}
            </button>
          )}
          <button onClick={() => { setSelected(new Set()); setBulkDeleteConfirm(false) }} className="px-3 py-1 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 text-xs">{t('common.clear')}</button>
        </div>
      )}

      {!isSearchActive && (
        <div className="flex items-center justify-between mb-2">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {t('ipAddressesPage.addressesCount', { count: total })}
            {fullRange && subnet && !subnet.networkAddress?.includes(':') && (
              <span className="ml-2 text-xs text-blue-500 dark:text-blue-400">{t('ipAddressesPage.fullRangeLabel')}</span>
            )}
          </p>
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
              <span>{t('ipAddressesPage.perPage')}</span>
              <select
                value={pageSize}
                onChange={e => handlePageSize(Number(e.target.value))}
                className="border rounded px-2 py-0.5 text-sm bg-white dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {PAGE_SIZE_OPTIONS.map(n => <option key={n} value={n}>{n}</option>)}
              </select>
            </label>
            {subnet && !subnet.networkAddress?.includes(':') && (
              <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer select-none">
                <span>{t('ipAddressesPage.showAllIPs')}</span>
                <button
                  role="switch"
                  aria-checked={fullRange}
                  onClick={() => handleFullRange(!fullRange)}
                  className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 ${fullRange ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'}`}
                >
                  <span className={`inline-block h-3 w-3 transform rounded-full bg-white transition-transform ${fullRange ? 'translate-x-5' : 'translate-x-1'}`} />
                </button>
              </label>
            )}
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              {isAdmin && (
                <th className="px-3 py-3 w-8">
                  <input type="checkbox" checked={ips.length > 0 && ips.every(ip => selected.has(ip.id))} onChange={e => e.target.checked ? setSelected(new Set(ips.map(ip => ip.id))) : setSelected(new Set())} />
                </th>
              )}
              {col('address') && <SortTh col="address" label={t('deviceIp.address')} sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />}
              {col('hostname') && <SortTh col="hostname" label={t('dashboard.hostname')} sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />}
              {col('status') && <SortTh col="status" label={t('delegations.status')} sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />}
              {col('tag') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminTags.tagColumn')}</th>}
              {col('device') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('ipAddressesPage.columnDevice')}</th>}
              {col('mac_address') && <SortTh col="mac_address" label={t('ipAddressesPage.columnMacAddress')} sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />}
              {col('dns_name') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('myRequests.dnsName')}</th>}
              {col('ptr_record') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('ipAddressesPage.ptrHostnameColumn')}</th>}
              {col('last_seen') && <SortTh col="last_seen" label={t('adminAgents.lastSeen')} sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />}
              {col('services') && <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('ipAddressesPage.columnServices')}</th>}
              {searchableFields.map(d => (
                <th key={d.name} className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{d.label}</th>
              ))}
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {ips.length === 0 && (
              <EmptyRow colSpan={visibleCols.length + searchableFields.length + 1} message={t('ipAddressesPage.noIpAddressesYet')} />
            )}
            {ips.map(ip => (
              <tr key={ip.virtual ? `v-${ip.address}` : ip.id} className={`border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30 ${ip.virtual ? 'opacity-50' : ''}`}>
                {isAdmin && (
                  <td className="px-3 py-2 w-8">
                    {!ip.virtual && <input type="checkbox" checked={selected.has(ip.id)} onChange={() => toggleSelect(ip.id)} />}
                  </td>
                )}
                {col('address') && <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{ip.address}</td>}
                {col('hostname') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.hostname || '—'}</td>}
                {col('status') && (
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${ip.virtual ? 'bg-gray-100 text-gray-400 dark:bg-gray-700 dark:text-gray-500' : STATUS_COLORS[ip.status] || 'bg-gray-100 text-gray-600'}`}>
                      {ip.status}
                    </span>
                  </td>
                )}
                {col('tag') && <td className="px-4 py-3"><TagBadge tag={ip.tag} /></td>}
                {col('device') && (
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    {ip.deviceId ? (
                      <Link to={`/devices/${ip.deviceId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                        {ip.device?.hostname || `#${ip.deviceId}`}
                      </Link>
                    ) : '—'}
                    {ip.expiresAt && (
                      <span className={`ml-1.5 text-xs px-1.5 py-0.5 rounded ${new Date(ip.expiresAt) < new Date() ? 'bg-red-100 text-red-700' : 'bg-yellow-50 text-yellow-700'}`}>
                        {new Date(ip.expiresAt) < new Date() ? t('ipAddressesPage.expired') : `${t('ipAddressesPage.expiresPrefix')}${new Date(ip.expiresAt).toLocaleDateString()}`}
                      </span>
                    )}
                  </td>
                )}
                {col('mac_address') && <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400 text-xs">{ip.macAddress || '—'}</td>}
                {col('dns_name') && (
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    <span className="flex items-center gap-1">
                      {ip.dnsName || '—'}
                      {ip.dnsName && ip.dnsRecords && !ip.dnsRecords.includes(ip.address) && (
                        <span
                          title={t('ipAddressesPage.dnsMismatchTooltip')}
                          className="text-yellow-500 cursor-help"
                        >
                          &#9888;
                        </span>
                      )}
                    </span>
                  </td>
                )}
                {col('ptr_record') && <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.ptrRecord || '—'}</td>}
                {col('last_seen') && (
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                    {ip.lastSeen ? new Date(ip.lastSeen).toLocaleString() : '—'}
                  </td>
                )}
                {col('services') && (
                  <td className="px-4 py-3">
                    <PortBadges portOpen={ip.portOpen} />
                  </td>
                )}
                {searchableFields.map(d => {
                  const val = ip.customFields?.[d.name]
                  return (
                    <td key={d.name} className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {val ? (
                        <button
                          className="hover:text-blue-600 dark:hover:text-blue-400 underline decoration-dotted text-left"
                          onClick={() => addCfFilterFromValue(d.name, val)}
                          title={t('devicesPage.filterByThisValue')}
                        >
                          {val}
                        </button>
                      ) : '—'}
                    </td>
                  )
                })}
                <td className="px-4 py-3 text-right space-x-2">
                  {ip.virtual ? (
                    isAdmin && <button onClick={() => openCreate(ip.address)} className="text-gray-400 hover:text-green-600 text-xs">{t('vrfs.create')}</button>
                  ) : (
                    <>
                      <button onClick={() => openMeta(ip)} className="text-gray-400 hover:text-indigo-600 text-xs">{t('common.edit')}</button>
                      {ip.status !== 'assigned' && (
                        <button onClick={() => openAssign(ip)} className="text-gray-400 hover:text-blue-600 text-xs">{t('ipAddressesPage.assign')}</button>
                      )}
                      {ip.status === 'assigned' && (
                        <>
                          <button onClick={() => handleRelease(ip.id)} className="text-gray-400 hover:text-yellow-600 text-xs">{t('ipAddressesPage.release')}</button>
                          {ip.expiresAt && new Date(ip.expiresAt) < new Date() && (
                            <button onClick={() => handleReleaseExpired(ip.id)} className="text-red-500 hover:text-red-700 text-xs">{t('ipAddressesPage.releaseExpired')}</button>
                          )}
                        </>
                      )}
                      {subnet?.technitiumScopeName && ip.macAddress && (
                        <button onClick={() => handlePushReservation(ip.id)} className="text-gray-400 hover:text-purple-600 text-xs">{t('ipAddressesPage.reserve')}</button>
                      )}
                      {subnet?.technitiumScopeName && (
                        <button onClick={() => handleRemoveReservation(ip.id)} className="text-gray-400 hover:text-orange-600 text-xs">{t('ipAddressesPage.unreserve')}</button>
                      )}
                      {deleteConfirm === ip.id ? (
                        <>
                          <span className="text-red-600 text-xs">{t('subnets.confirmDelete')}</span>
                          <button onClick={() => handleDelete(ip.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">{t('common.yes')}</button>
                          <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">{t('common.no')}</button>
                        </>
                      ) : (
                        <button onClick={() => setDeleteConfirm(ip.id)} className="text-gray-400 hover:text-red-600 text-xs">{t('common.delete')}</button>
                      )}
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {!isSearchActive && total > pageSize && (
        <Pagination
          page={page}
          limit={pageSize}
          total={total}
          onChange={handlePageChange}
        />
      )}
      <UtilisationHistorySection subnetId={subnetID} />
      </>}

      {modal === 'create' && (
        <Modal title={t('ipAddressesPage.newIpAddressModalTitle')} onClose={() => setModal(null)}>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('ipAddressesPage.ipAddressLabel')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="192.168.0.10"
                value={form.address}
                onChange={e => setForm(f => ({ ...f, address: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('dashboard.hostname')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="server01.example.com"
                value={form.hostname}
                onChange={e => setForm(f => ({ ...f, hostname: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('delegations.status')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.status}
                onChange={e => setForm(f => ({ ...f, status: e.target.value }))}
              >
                <option value="available">{t('ipAddressesPage.available')}</option>
                <option value="assigned">{t('adminRoles.assignedLabel')}</option>
                <option value="reserved">{t('dhcp.reserved')}</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminTags.tagColumn')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.tag_id}
                onChange={e => setForm(f => ({ ...f, tag_id: e.target.value }))}
              >
                <option value="">{t('ipAddressesPage.noTag')}</option>
                {tags.map(tag => <option key={tag.id} value={tag.id}>{tag.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('ipAddressesPage.columnMacAddress')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="aa:bb:cc:dd:ee:ff"
                value={form.mac_address}
                onChange={e => setForm(f => ({ ...f, mac_address: filterMACInput(e.target.value) }))}
                onBlur={e => setForm(f => ({ ...f, mac_address: normalizeMAC(e.target.value) }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('ipAddressesPage.ptrHostnameLabel')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="host.example.com"
                value={form.ptr_record}
                onChange={e => setForm(f => ({ ...f, ptr_record: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('myRequests.dnsName')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="web01.example.com"
                value={form.dns_name}
                onChange={e => setForm(f => ({ ...f, dns_name: e.target.value }))}
              />
            </div>
            {cfDefs.length > 0 && (
              <div className="border-t pt-4">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">{t('subnetForm.customFields')}</p>
                <CustomFieldForm
                  definitions={cfDefs}
                  values={form.custom_fields}
                  onChange={(name, value) => setForm(f => ({ ...f, custom_fields: { ...f.custom_fields, [name]: value } }))}
                />
              </div>
            )}
            {createError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{createError}</div>
            )}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('ipAddressesPage.addIp')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal?.assign && (
        <Modal title={t('ipAddressesPage.assignModalTitle', { address: modal.assign.Address })} onClose={() => setModal(null)}>
          <form onSubmit={handleAssign} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('ipAddressesPage.columnDevice')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
                value={form.device_id}
                onChange={e => setForm(f => ({ ...f, device_id: e.target.value }))}
              >
                <option value="">{t('ipAddressesPage.noneOption')}</option>
                {devices.map(d => (
                  <option key={d.id} value={d.id}>{d.hostname}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('ipAddressesPage.leaseDurationLabel')} <span className="text-gray-400 font-normal">{t('ipAddressesPage.optionalSuffix')}</span>
              </label>
              <input
                type="number"
                min="1"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder={t('ipAddressesPage.leaveBlankPermanentPlaceholder')}
                value={form.lease_duration_days}
                onChange={e => setForm(f => ({ ...f, lease_duration_days: e.target.value }))}
              />
              <p className="text-xs text-gray-400 mt-1">{t('ipAddressesPage.leaseDurationHint')}</p>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('ipAddressesPage.assign')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal === 'requestIP' && (
        <Modal title={t('ipAddressesPage.requestIpAddressModalTitle')} onClose={() => setModal(null)}>
          {ipReqSuccess ? (
            <div className="py-4 text-center text-green-600 font-medium">{t('ipAddressesPage.requestSubmittedSuccess')}</div>
          ) : (
            <form onSubmit={handleIPRequestSubmit} className="space-y-4">
              {ipReqError && (
                <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{ipReqError}</div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">{t('ipAddressesPage.subnetLabel')}</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm font-mono bg-gray-50 text-gray-500"
                  value={subnet ? `${subnet.networkAddress}/${subnet.prefixLength}` : t('ipAddressesPage.subnetHashLabel', { id: subnetID })}
                  readOnly
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('ipAddressesPage.specificIpLabel')} <span className="text-gray-400 font-normal">{t('ipAddressesPage.optionalAutoAssignHint')}</span>
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
                  {t('myRequests.dnsName')} <span className="text-gray-400 font-normal">{t('ipAddressesPage.optionalHint')}</span>
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
                  {t('ipAddressesPage.purposeLabel')} <span className="text-red-500">*</span>
                </label>
                <textarea
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={3}
                  placeholder={t('ipAddressesPage.purposePlaceholder')}
                  value={ipReqForm.purpose}
                  onChange={e => setIPReqForm(f => ({ ...f, purpose: e.target.value }))}
                  required
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
                <button type="submit" disabled={saving} className="px-4 py-2 bg-green-600 text-white rounded text-sm hover:bg-green-700 disabled:opacity-50">
                  {saving ? t('ipAddressesPage.submitting') : t('ipAddressesPage.submitRequest')}
                </button>
              </div>
            </form>
          )}
        </Modal>
      )}

      {modal?.meta && (
        <Modal title={t('ipAddressesPage.editIpModalTitle', { address: modal.meta.Address })} onClose={() => setModal(null)}>
          <form onSubmit={handleUpdateMeta} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('ipAddressesPage.columnDevice')}</label>
              {form.device_id ? (
                <div className="flex items-center gap-2 px-3 py-2 border rounded bg-gray-50 dark:bg-gray-700 dark:border-gray-600">
                  <span className="text-sm text-gray-700 dark:text-gray-200 flex-1">
                    {devices.find(d => d.id === parseInt(form.device_id))?.hostname || t('ipAddressesPage.deviceHashLabel', { id: form.device_id })}
                  </span>
                  <button type="button" onClick={() => setForm(f => ({ ...f, device_id: '' }))} className="text-gray-400 hover:text-red-500 text-xs">{t('ipAddressesPage.clearSelection')}</button>
                </div>
              ) : (
                <div className="relative">
                  <input
                    type="text"
                    placeholder={t('ipAddressesPage.searchDevicesPlaceholder')}
                    value={metaDeviceSearch}
                    onChange={e => setMetaDeviceSearch(e.target.value)}
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  />
                  {metaDeviceSearch && (
                    <div className="absolute top-full left-0 right-0 bg-white dark:bg-gray-800 border dark:border-gray-600 rounded shadow-lg z-20 max-h-40 overflow-y-auto">
                      {devices
                        .filter(d => d.hostname?.toLowerCase().includes(metaDeviceSearch.toLowerCase()))
                        .slice(0, 10)
                        .map(d => (
                          <button
                            key={d.id}
                            type="button"
                            className="w-full text-left px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-800 dark:text-gray-200"
                            onClick={() => { setForm(f => ({ ...f, device_id: String(d.id) })); setMetaDeviceSearch('') }}
                          >
                            {d.hostname}
                          </button>
                        ))}
                      {devices.filter(d => d.hostname?.toLowerCase().includes(metaDeviceSearch.toLowerCase())).length === 0 && (
                        <div className="px-3 py-2 text-sm text-gray-400">{t('ipAddressesPage.noDevicesFound')}</div>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminTags.tagColumn')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.tag_id}
                onChange={e => setForm(f => ({ ...f, tag_id: e.target.value }))}
              >
                <option value="">{t('ipAddressesPage.noTag')}</option>
                {tags.map(tag => <option key={tag.id} value={tag.id}>{tag.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('ipAddressesPage.columnMacAddress')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="aa:bb:cc:dd:ee:ff"
                value={form.mac_address}
                onChange={e => setForm(f => ({ ...f, mac_address: filterMACInput(e.target.value) }))}
                onBlur={e => setForm(f => ({ ...f, mac_address: normalizeMAC(e.target.value) }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('ipAddressesPage.ptrHostnameLabel')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="host.example.com"
                value={form.ptr_record}
                onChange={e => setForm(f => ({ ...f, ptr_record: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('myRequests.dnsName')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="web01.example.com"
                value={form.dns_name}
                onChange={e => setForm(f => ({ ...f, dns_name: e.target.value }))}
              />
            </div>
            {modal.meta.portOpen && Object.values(modal.meta.portOpen).some(Boolean) && (
              <div className="border-t pt-3">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{t('ipAddressesPage.openPortsReadOnly')}</p>
                <PortBadges portOpen={modal.meta.portOpen} />
              </div>
            )}
            {(modal.meta.dnsLastChecked) && (
              <div className="border-t pt-3">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{t('ipAddressesPage.dnsInfoReadOnly')}</p>
                <div className="grid grid-cols-2 gap-2 text-xs text-gray-600">
                  <span className="font-medium">{t('ipAddressesPage.lastDnsCheckLabel')}</span>
                  <span>{modal.meta.dnsLastChecked ? new Date(modal.meta.dnsLastChecked).toLocaleString() : '—'}</span>
                </div>
              </div>
            )}
            {isAdmin && (
              <div className="border-t pt-3">
                <ObservedStatePanel resourceType="ip_address" resourceId={modal.meta.id} />
              </div>
            )}
            {cfDefs.length > 0 && (
              <div className="border-t pt-4">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">{t('subnetForm.customFields')}</p>
                <CustomFieldForm
                  definitions={cfDefs}
                  values={form.custom_fields}
                  onChange={(name, value) => setForm(f => ({ ...f, custom_fields: { ...f.custom_fields, [name]: value } }))}
                />
              </div>
            )}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('common.save')}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
