import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import * as client from '../api/admin'

const ACTION_KEYS = [
  'login', 'login_mfa', 'logout', 'token_created', 'token_revoked', 'mfa_enabled',
  'mfa_disabled', 'backup_codes_regenerated', 'user_created', 'user_role_updated',
  'user_deleted', 'user_approved', 'user_rejected', 'account_unlocked', 'config_updated',
  'section_created', 'section_updated', 'section_deleted', 'subnet_created',
  'subnet_updated', 'subnet_deleted', 'ip_address_created', 'ip_address_deleted',
  'ip_assigned', 'ip_released', 'ip_allocated', 'vrf_created', 'vrf_updated',
  'vrf_deleted', 'vlan_created', 'vlan_updated', 'vlan_deleted', 'audit_logs_purged',
]

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
  const { t } = useTranslation()
  const ACTION_LABELS = Object.fromEntries(ACTION_KEYS.map(k => [k, t(`auditLog.actions.${k}`)]))
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
      setError(t('auditLog.exportFailedPrefix') + (err.response?.data?.error || err.message))
    } finally {
      setExporting(false)
    }
  }

  const totalPages = Math.ceil(total / limit)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{t('audit.logTab')}</h1>
          <p className="text-sm text-gray-500 mt-1">{t('auditLog.totalEntries', { count: total, formattedCount: total.toLocaleString() })}</p>
        </div>
        <button
          onClick={handleExport}
          disabled={exporting}
          className="px-4 py-2 bg-gray-700 text-white text-sm rounded hover:bg-gray-800 disabled:opacity-50"
        >
          {exporting ? t('networks.exporting') : t('networks.exportCsv')}
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('auditLog.action')}</label>
            <select
              value={filters.action}
              onChange={(e) => handleFilterChange('action', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            >
              <option value="">{t('auditLog.allActions')}</option>
              {Object.entries(ACTION_LABELS).map(([k, v]) => (
                <option key={k} value={k}>{v}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('auditLog.resourceType')}</label>
            <select
              value={filters.resource_type}
              onChange={(e) => handleFilterChange('resource_type', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            >
              <option value="">{t('auditLog.allTypes')}</option>
              {['session', 'api_token', 'user', 'user_approval', 'config', 'network', 'subnet', 'ip_address', 'vrf', 'vlan', 'audit_log'].map((rt) => (
                <option key={rt} value={rt}>{rt}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('login.username')}</label>
            <input
              type="text"
              value={filters.username}
              onChange={(e) => handleFilterChange('username', e.target.value)}
              placeholder={t('auditLog.searchUsernamePlaceholder')}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('delegations.status')}</label>
            <select
              value={filters.status}
              onChange={(e) => handleFilterChange('status', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            >
              <option value="">{t('discoveryConflicts.all')}</option>
              <option value="success">{t('auditLog.success')}</option>
              <option value="failure">{t('auditLog.failure')}</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('associateIp.ipAddress')}</label>
            <input
              type="text"
              value={filters.ip}
              onChange={(e) => handleFilterChange('ip', e.target.value)}
              placeholder={t('auditLog.ipAddressPlaceholder')}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('auditLog.from')}</label>
            <input
              type="datetime-local"
              value={filters.since}
              onChange={(e) => handleFilterChange('since', e.target.value)}
              className="w-full text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('auditLog.until')}</label>
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
              {t('auditLog.clearFilters')}
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
          <div className="p-8 text-center text-gray-500">{t('common.loading')}</div>
        ) : logs.length === 0 ? (
          <div className="p-8 text-center text-gray-400">{t('auditLog.noEntriesFound')}</div>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">{t('auditLog.timestamp')}</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">{t('auditLog.user')}</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">{t('auditLog.action')}</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">{t('auditLog.resource')}</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">{t('associateIp.ipAddress')}</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">{t('delegations.status')}</th>
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
                  <td className="px-4 py-2.5 font-medium text-gray-900">{log.username || <span className="text-gray-400 italic">{t('auditLog.systemUser')}</span>}</td>
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
              {t('auditLog.showingRange', { start: page * limit + 1, end: Math.min((page + 1) * limit, total), total: total.toLocaleString() })}
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
                className="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-40 hover:bg-gray-100"
              >
                {t('auditLog.previous')}
              </button>
              <span className="px-3 py-1 text-sm text-gray-600">
                {t('auditLog.pageOf', { page: page + 1, totalPages })}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                disabled={page >= totalPages - 1}
                className="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-40 hover:bg-gray-100"
              >
                {t('auditLog.next')}
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
              <Row label={t('auditLog.timestamp')} value={formatTimestamp(selectedLog.timestamp)} />
              <Row label={t('auditLog.user')} value={selectedLog.username || t('auditLog.systemParens')} />
              <Row label={t('auditLog.action')} value={selectedLog.action} />
              <Row label={t('auditLog.resourceType')} value={selectedLog.resourceType} />
              {selectedLog.resourceId && <Row label={t('auditLog.resourceIdLabel')} value={selectedLog.resourceId} />}
              {selectedLog.resourceName && <Row label={t('auditLog.resourceNameLabel')} value={selectedLog.resourceName} />}
              <Row label={t('associateIp.ipAddress')} value={selectedLog.ipAddress} mono />
              <Row label={t('delegations.status')} value={selectedLog.status} />
              {selectedLog.errorMessage && <Row label={t('auditLog.errorLabel')} value={selectedLog.errorMessage} />}
              {selectedLog.newValues && (
                <div>
                  <span className="font-medium text-gray-600">{t('auditLog.newValues')}</span>
                  <pre className="mt-1 bg-gray-50 rounded p-2 text-xs overflow-auto">
                    {JSON.stringify(JSON.parse(selectedLog.newValues), null, 2)}
                  </pre>
                </div>
              )}
              {selectedLog.oldValues && (
                <div>
                  <span className="font-medium text-gray-600">{t('auditLog.oldValues')}</span>
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
