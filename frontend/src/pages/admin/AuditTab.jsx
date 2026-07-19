import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { purgeAuditLogs } from '../../api/admin'

export default function AuditTab({ config, handleConfigChange, handleSaveConfig, saving, showMessage }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [purging, setPurging] = useState(false)

  const handlePurgeAuditLogs = async () => {
    if (!window.confirm(t('audit.purgeConfirm'))) return
    setPurging(true)
    try {
      const res = await purgeAuditLogs()
      showMessage(res.data.message || t('audit.purgedDefault'))
    } catch (err) {
      showMessage(t('audit.purgeFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setPurging(false)
    }
  }

  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">{t('audit.retentionTitle')}</h2>
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('audit.retentionPeriodDays')}
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
                {t('audit.retentionHint')}
              </p>
            </div>
            <div className="flex gap-3 items-center">
              <button
                onClick={handleSaveConfig}
                disabled={saving}
                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
              >
                {saving ? t('common.saving') : t('common.save')}
              </button>
            </div>
          </div>
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-2">{t('audit.purgeOldLogsTitle')}</h2>
            <p className="text-sm text-gray-600 mb-4">
              {t('audit.purgeOldLogsSubtitle')}
            </p>
            <div className="flex gap-3">
              <button
                onClick={handlePurgeAuditLogs}
                disabled={purging}
                className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 disabled:opacity-50 transition text-sm font-medium"
              >
                {purging ? t('audit.purging') : t('audit.purgeOldLogs')}
              </button>
              <button
                onClick={() => navigate('/admin/audit-log')}
                className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 transition text-sm font-medium"
              >
                {t('audit.viewAuditLog')}
              </button>
            </div>
          </div>
        </div>
  )
}
