import { useState, useEffect } from 'react'
import { api } from '../api/client'
import Modal from '../components/Modal'
import { downloadFile } from '../utils/download'

const DAYS_OPTIONS = [30, 60, 90, 180]

const SORT_KEYS = {
  ipAddress: (a, b) => {
    // Sort by IP numerically
    const toNum = ip => ip.split('.').reduce((acc, o) => (acc << 8) + parseInt(o), 0) >>> 0
    return toNum(a.ipAddress) - toNum(b.ipAddress)
  },
  hostname: (a, b) => (a.hostname || '').localeCompare(b.hostname || ''),
  subnetCidr: (a, b) => (a.subnetCidr || '').localeCompare(b.subnetCidr || ''),
  sectionName: (a, b) => (a.sectionName || '').localeCompare(b.sectionName || ''),
  assignedTo: (a, b) => (a.assignedTo || '').localeCompare(b.assignedTo || ''),
  lastSeen: (a, b) => {
    const ta = a.lastSeen ? new Date(a.lastSeen).getTime() : 0
    const tb = b.lastSeen ? new Date(b.lastSeen).getTime() : 0
    return ta - tb
  },
  daysInactive: (a, b) => (a.daysInactive ?? 0) - (b.daysInactive ?? 0),
}

export default function InactiveIPsPage() {
  const [days, setDays] = useState(90)
  const [sectionId, setSectionId] = useState('')
  const [sections, setSections] = useState([])
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState([])
  const [sortKey, setSortKey] = useState('daysInactive')
  const [sortDir, setSortDir] = useState('desc')
  const [confirmRelease, setConfirmRelease] = useState(false)
  const [releasing, setReleasing] = useState(false)
  const [releaseMsg, setReleaseMsg] = useState('')
  const [downloading, setDownloading] = useState(false)

  useEffect(() => { loadSections() }, [])
  useEffect(() => { load() }, [days, sectionId])

  async function loadSections() {
    try {
      const { data } = await api.get('/sections')
      setSections(Array.isArray(data) ? data : (data?.data ?? []))
    } catch {}
  }

  async function load() {
    setLoading(true)
    setError('')
    setSelected([])
    try {
      const params = { days }
      if (sectionId) params.section_id = sectionId
      const { data } = await api.get('/admin/reports/inactive-ips', { params })
      setRows(Array.isArray(data) ? data : [])
    } catch {
      setError('Failed to load inactive IPs')
    } finally {
      setLoading(false)
    }
  }

  function toggleSort(key) {
    if (sortKey === key) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDir('desc')
    }
  }

  const sorted = [...rows].sort((a, b) => {
    const fn = SORT_KEYS[sortKey] || SORT_KEYS.daysInactive
    return sortDir === 'asc' ? fn(a, b) : fn(b, a)
  })

  function toggleAll() {
    if (selected.length === rows.length) {
      setSelected([])
    } else {
      setSelected(rows.map(r => r.ipId))
    }
  }

  function toggleRow(id) {
    setSelected(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }

  async function handleRelease() {
    setReleasing(true)
    setReleaseMsg('')
    try {
      await api.post('/admin/ip-addresses/bulk-release', { ipIds: selected })
      setReleaseMsg(`Successfully released ${selected.length} IP address(es).`)
      setConfirmRelease(false)
      setSelected([])
      load()
    } catch (err) {
      setReleaseMsg(err.response?.data?.error || 'Failed to release IPs')
    } finally {
      setReleasing(false)
    }
  }

  async function handleExport() {
    setDownloading(true)
    try {
      const params = new URLSearchParams({ days, format: 'csv' })
      if (sectionId) params.set('section_id', sectionId)
      await downloadFile(`/api/v1/admin/reports/export/inactive-ips?${params}`, `inactive-ips-${days}d.csv`)
    } catch {
      setError('Export failed')
    } finally {
      setDownloading(false)
    }
  }

  function SortTh({ col, label }) {
    const active = sortKey === col
    return (
      <th
        className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium cursor-pointer select-none hover:text-blue-600"
        onClick={() => toggleSort(col)}
      >
        {label} {active ? (sortDir === 'asc' ? '▲' : '▼') : ''}
      </th>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Inactive IPs</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">IP addresses that have been inactive for the selected period</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleExport}
            disabled={downloading}
            className="px-3 py-1.5 bg-gray-600 text-white rounded text-sm hover:bg-gray-700 disabled:opacity-50"
          >
            {downloading ? 'Exporting...' : 'Export CSV'}
          </button>
          {selected.length > 0 && (
            <button
              onClick={() => setConfirmRelease(true)}
              className="px-3 py-1.5 bg-red-600 text-white rounded text-sm hover:bg-red-700"
            >
              Release Selected ({selected.length})
            </button>
          )}
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-4 flex flex-wrap gap-4 items-center">
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Inactive for at least</label>
          <div className="flex gap-1">
            {DAYS_OPTIONS.map(d => (
              <button
                key={d}
                onClick={() => setDays(d)}
                className={`px-3 py-1.5 rounded text-sm font-medium transition ${days === d ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}`}
              >
                {d}d
              </button>
            ))}
          </div>
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Section</label>
          <select
            value={sectionId}
            onChange={e => setSectionId(e.target.value)}
            className="border rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
          >
            <option value="">All Sections</option>
            {sections.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>
        </div>
      </div>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}
      {releaseMsg && (
        <div className={`mb-4 px-4 py-2 rounded text-sm ${releaseMsg.startsWith('Successfully') ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
          {releaseMsg}
        </div>
      )}

      {loading ? (
        <p className="text-gray-500 text-sm">Loading...</p>
      ) : (
        <>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">{rows.length} inactive IP{rows.length !== 1 ? 's' : ''}</p>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="px-4 py-3">
                    <input
                      type="checkbox"
                      checked={rows.length > 0 && selected.length === rows.length}
                      onChange={toggleAll}
                      className="w-4 h-4"
                    />
                  </th>
                  <SortTh col="ipAddress" label="IP Address" />
                  <SortTh col="hostname" label="Hostname" />
                  <SortTh col="subnetCidr" label="Subnet" />
                  <SortTh col="sectionName" label="Section" />
                  <SortTh col="assignedTo" label="Assigned To" />
                  <SortTh col="lastSeen" label="Last Seen" />
                  <SortTh col="daysInactive" label="Days Inactive" />
                </tr>
              </thead>
              <tbody>
                {sorted.length === 0 && (
                  <tr>
                    <td colSpan={8} className="px-4 py-8 text-center text-gray-400">No inactive IPs found for this period</td>
                  </tr>
                )}
                {sorted.map(row => (
                  <tr key={row.ipId} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selected.includes(row.ipId)}
                        onChange={() => toggleRow(row.ipId)}
                        className="w-4 h-4"
                      />
                    </td>
                    <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{row.ipAddress}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{row.hostname || '—'}</td>
                    <td className="px-4 py-3 font-mono text-blue-600 dark:text-blue-400 text-xs">{row.subnetCidr}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{row.sectionName || '—'}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{row.assignedTo || '—'}</td>
                    <td className="px-4 py-3 text-gray-400 text-xs">
                      {row.lastSeen ? new Date(row.lastSeen).toLocaleDateString() : 'Never'}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`font-medium ${row.daysInactive > 90 ? 'text-red-600' : row.daysInactive > 30 ? 'text-yellow-600' : 'text-gray-600 dark:text-gray-400'}`}>
                        {row.daysInactive ?? '—'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      {confirmRelease && (
        <Modal title="Confirm Bulk Release" onClose={() => setConfirmRelease(false)}>
          <div className="space-y-4">
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Are you sure you want to release <strong>{selected.length}</strong> IP address(es)? This will mark them as available.
            </p>
            {releaseMsg && !releaseMsg.startsWith('Successfully') && (
              <p className="text-red-600 text-sm">{releaseMsg}</p>
            )}
            <div className="flex justify-end gap-2">
              <button onClick={() => setConfirmRelease(false)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button
                onClick={handleRelease}
                disabled={releasing}
                className="px-4 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700 disabled:opacity-50"
              >
                {releasing ? 'Releasing...' : 'Release'}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
