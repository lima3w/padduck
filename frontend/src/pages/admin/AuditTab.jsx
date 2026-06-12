import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { purgeAuditLogs } from '../../api/admin'

export default function AuditTab({ config, handleConfigChange, handleSaveConfig, saving, showMessage }) {
  const navigate = useNavigate()
  const [purging, setPurging] = useState(false)

  const handlePurgeAuditLogs = async () => {
    if (!window.confirm('Delete all audit log entries older than the configured retention period?')) return
    setPurging(true)
    try {
      const res = await purgeAuditLogs()
      showMessage(res.data.message || 'Audit logs purged')
    } catch (err) {
      showMessage('Purge failed: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setPurging(false)
    }
  }

  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Audit Log Retention</h2>
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Retention Period (days)
              </label>
              <input
                type="number"
                min="1"
                max="3650"
                value={config.audit_log_retention_days || '90'}
                onChange={(e) => handleConfigChange('audit_log_retention_days', e.target.value)}
                className="w-32 px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              />
              <p className="text-xs text-gray-500 mt-1">
                Audit logs older than this many days will be deleted when a purge is run. Default: 90 days.
              </p>
            </div>
            <div className="flex gap-3 items-center">
              <button
                onClick={handleSaveConfig}
                disabled={saving}
                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-2">Purge Old Logs</h2>
            <p className="text-sm text-gray-600 mb-4">
              Permanently delete audit log entries older than the configured retention period.
            </p>
            <div className="flex gap-3">
              <button
                onClick={handlePurgeAuditLogs}
                disabled={purging}
                className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 disabled:opacity-50 transition text-sm font-medium"
              >
                {purging ? 'Purging...' : 'Purge Old Logs'}
              </button>
              <button
                onClick={() => navigate('/admin/audit-log')}
                className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 transition text-sm font-medium"
              >
                View Audit Log
              </button>
            </div>
          </div>
        </div>
  )
}
