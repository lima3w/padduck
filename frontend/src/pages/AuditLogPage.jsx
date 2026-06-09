import { useState, useEffect, useCallback } from 'react'
import * as client from '../api/client'

const ACTION_LABELS = {
  login: 'Login',
  login_mfa: 'Login (MFA)',
  logout: 'Logout',
  token_created: 'Token Created',
  token_revoked: 'Token Revoked',
  mfa_enabled: 'MFA Enabled',
  mfa_disabled: 'MFA Disabled',
  backup_codes_regenerated: 'Backup Codes Regenerated',
  user_created: 'User Created',
  user_role_updated: 'Role Changed',
  user_deleted: 'User Deleted',
  user_approved: 'User Approved',
  user_rejected: 'User Rejected',
  account_unlocked: 'Account Unlocked',
  config_updated: 'Config Updated',
  section_created: 'Network Created',
  section_updated: 'Network Updated',
  section_deleted: 'Network Deleted',
  subnet_created: 'Subnet Created',
  subnet_updated: 'Subnet Updated',
  subnet_deleted: 'Subnet Deleted',
  ip_address_created: 'IP Created',
  ip_address_deleted: 'IP Deleted',
  ip_assigned: 'IP Assigned',
  ip_released: 'IP Released',
  ip_allocated: 'IP Allocated',
  vrf_created: 'VRF Created',
  vrf_updated: 'VRF Updated',
  vrf_deleted: 'VRF Deleted',
  vlan_created: 'VLAN Created',
  vlan_updated: 'VLAN Updated',
  vlan_deleted: 'VLAN Deleted',
  audit_logs_purged: 'Audit Logs Purged',
}

const ACTION_COLORS = {
  login: 'bg-green-100 text-green-800',
  login_mfa: 'bg-green-100 text-green-800',
  logout: 'bg-gray-100 text-gray-700',
  user_deleted: 'bg-red-100 text-red-800',
  section_deleted: 'bg-red-100 text-red-800',
  subnet_deleted: 'bg-red-100 text-red-800',
  ip_address_deleted: 'bg-red-100 text-red-800',
  vrf_deleted: 'bg-red-100 text-red-800',
  vlan_deleted: 'bg-red-100 text-red-800',
  account_unlocked: 'bg-yellow-100 text-yellow-800',
  config_updated: 'bg-purple-100 text-purple-800',
  audit_logs_purged: 'bg-orange-100 text-orange-800',
}

function getActionColor(action) {
  return ACTION_COLORS[action] || 'bg-blue-100 text-blue-800'
}

function formatTimestamp(ts) {
  return new Date(ts).toLocaleString()
}

export default function AuditLogPage() {
  const [logs, setLogs] = useState([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [exporting, setExporting] = useState(false)
  const [error, setError] = useState('')
  const [selectedLog, setSelectedLog] = useState(null)

  const [filters, setFilters] = useState({
    action: '',
    resource_type: '',
    username: '',
    ip: '',
    status: '',
    since: '',
    until: '',
  })
  const [page, setPage] = useState(0)
  const limit = 50

  const loadLogs = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const params = { limit, offset: page * limit }
      if (filters.action) params.action = filters.action
      if (filters.resource_type) params.resource_type = filters.resource_type
      if (filters.username) params.username = filters.username
      if (filters.ip) params.ip = filters.ip
      if (filters.status) params.status = filters.status
      if (filters.since) params.since = new Date(filters.since).toISOString()
      if (filters.until) params.until = new Date(filters.until).toISOString()

      const res = await client.getAuditLogs(params)
      setLogs(res.data.logs || [])
      setTotal(res.data.total || 0)
    } catch (err) {
      setError(err.response?.data?.error || err.message)
    } finally {
      setLoading(false)
    }
  }, [filters, page])

  useEffect(() => {
    loadLogs()
  }, [loadLogs])

  const handleFilterChange = (key, value) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
    setPage(0)
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const params = {}
      if (filters.action) params.action = filters.action
      if (filters.resource_type) params.resource_type = filters.resource_type
      if (filters.username) params.username = filters.username
      if (filters.ip) params.ip = filters.ip
      if (filters.status) params.status = filters.status
      if (filters.since) params.since = new Date(filters.since).toISOString()
      if (filters.until) params.until = new Date(filters.until).toISOString()

      const res = await client.exportAuditLogs(params)
      const url = URL.createObjectURL(new Blob([res.data], { type: 'text/csv' }))
      const a = document.createElement('a')
      a.href = url
      a.download = `audit-log-${new Date().toISOString().slice(0, 10)}.csv`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      setError('Export failed: ' + (err.response?.data?.error || err.message))
    } finally {
      setExporting(false)
    }
  }

  const totalPages = Math.ceil(total / limit)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Audit Log</h1>
          <p className="text-sm text-gray-500 mt-1">{total.toLocaleString()} total entries</p>
        </div>
        <button
          onClick={handleExport}
          disabled={exporting}
          className="px-4 py-2 bg-gray-700 text-white text-sm rounded hover:bg-gray-800 disabled:opacity-50"
        >
          {exporting ? 'Exporting…' : 'Export CSV'}
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Action</label>
            <select
              value={filters.action}
              onChange={(e) => handleFilterChange('action', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            >
              <option value="">All actions</option>
              {Object.entries(ACTION_LABELS).map(([k, v]) => (
                <option key={k} value={k}>{v}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Resource Type</label>
            <select
              value={filters.resource_type}
              onChange={(e) => handleFilterChange('resource_type', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            >
              <option value="">All types</option>
              {['session', 'api_token', 'user', 'user_approval', 'config', 'network', 'subnet', 'ip_address', 'vrf', 'vlan', 'audit_log'].map((t) => (
                <option key={t} value={t}>{t}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Username</label>
            <input
              type="text"
              value={filters.username}
              onChange={(e) => handleFilterChange('username', e.target.value)}
              placeholder="Search username…"
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Status</label>
            <select
              value={filters.status}
              onChange={(e) => handleFilterChange('status', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            >
              <option value="">All</option>
              <option value="success">Success</option>
              <option value="failure">Failure</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">IP Address</label>
            <input
              type="text"
              value={filters.ip}
              onChange={(e) => handleFilterChange('ip', e.target.value)}
              placeholder="e.g. 192.168.1.1"
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">From</label>
            <input
              type="datetime-local"
              value={filters.since}
              onChange={(e) => handleFilterChange('since', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Until</label>
            <input
              type="datetime-local"
              value={filters.until}
              onChange={(e) => handleFilterChange('until', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>
          <div className="flex items-end">
            <button
              onClick={() => { setFilters({ action: '', resource_type: '', username: '', ip: '', status: '', since: '', until: '' }); setPage(0) }}
              className="w-full px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50"
            >
              Clear Filters
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">{error}</div>
      )}

      {/* Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-gray-500">Loading…</div>
        ) : logs.length === 0 ? (
          <div className="p-8 text-center text-gray-400">No audit log entries found</div>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Timestamp</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">User</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Action</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Resource</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">IP Address</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {logs.map((log) => (
                <tr
                  key={log.id}
                  className="hover:bg-gray-50 cursor-pointer"
                  onClick={() => setSelectedLog(log)}
                >
                  <td className="px-4 py-2.5 text-gray-600 whitespace-nowrap">{formatTimestamp(log.timestamp)}</td>
                  <td className="px-4 py-2.5 font-medium text-gray-900">{log.username || <span className="text-gray-400 italic">system</span>}</td>
                  <td className="px-4 py-2.5">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${getActionColor(log.action)}`}>
                      {ACTION_LABELS[log.action] || log.action}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-gray-600">
                    {log.resourceType && (
                      <span>
                        <span className="text-gray-400">{log.resourceType}</span>
                        {log.resourceName && <> · <span className="font-mono text-xs">{log.resourceName}</span></>}
                        {log.resourceId && !log.resourceName && <> #{log.resourceId}</>}
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-2.5 font-mono text-xs text-gray-500">{log.ipAddress}</td>
                  <td className="px-4 py-2.5">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                      log.status === 'failure' ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'
                    }`}>
                      {log.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200 bg-gray-50">
            <span className="text-sm text-gray-600">
              Showing {page * limit + 1}–{Math.min((page + 1) * limit, total)} of {total.toLocaleString()}
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
                className="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-40 hover:bg-gray-100"
              >
                Previous
              </button>
              <span className="px-3 py-1 text-sm text-gray-600">
                Page {page + 1} of {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                disabled={page >= totalPages - 1}
                className="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-40 hover:bg-gray-100"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Detail modal */}
      {selectedLog && (
        <div
          className="fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center z-50 p-4"
          onClick={() => setSelectedLog(null)}
        >
          <div
            className="bg-white rounded-lg shadow-xl max-w-lg w-full max-h-[80vh] overflow-auto"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-gray-200">
              <h2 className="font-semibold text-gray-900">
                {ACTION_LABELS[selectedLog.action] || selectedLog.action}
              </h2>
              <button onClick={() => setSelectedLog(null)} className="text-gray-400 hover:text-gray-600 text-xl leading-none">&times;</button>
            </div>
            <div className="px-5 py-4 space-y-3 text-sm">
              <Row label="Timestamp" value={formatTimestamp(selectedLog.timestamp)} />
              <Row label="User" value={selectedLog.username || '(system)'} />
              <Row label="Action" value={selectedLog.action} />
              <Row label="Resource Type" value={selectedLog.resourceType} />
              {selectedLog.resourceId && <Row label="Resource ID" value={selectedLog.resourceId} />}
              {selectedLog.resourceName && <Row label="Resource Name" value={selectedLog.resourceName} />}
              <Row label="IP Address" value={selectedLog.ipAddress} mono />
              <Row label="Status" value={selectedLog.status} />
              {selectedLog.errorMessage && <Row label="Error" value={selectedLog.errorMessage} />}
              {selectedLog.newValues && (
                <div>
                  <span className="font-medium text-gray-600">New Values</span>
                  <pre className="mt-1 bg-gray-50 rounded p-2 text-xs overflow-auto">
                    {JSON.stringify(JSON.parse(selectedLog.newValues), null, 2)}
                  </pre>
                </div>
              )}
              {selectedLog.oldValues && (
                <div>
                  <span className="font-medium text-gray-600">Old Values</span>
                  <pre className="mt-1 bg-gray-50 rounded p-2 text-xs overflow-auto">
                    {JSON.stringify(JSON.parse(selectedLog.oldValues), null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function Row({ label, value, mono }) {
  if (!value && value !== 0) return null
  return (
    <div className="flex gap-2">
      <span className="font-medium text-gray-600 w-32 shrink-0">{label}</span>
      <span className={mono ? 'font-mono text-xs' : ''}>{String(value)}</span>
    </div>
  )
}
