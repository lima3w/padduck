import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getAuditRetention, updateAuditRetention, pruneAuditLogs } from '../api/admin'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

export default function AuditRetentionPage() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [pruning, setPruning] = useState(false)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [form, setForm] = useState({ retentionDays: 365, archiveEnabled: false })

  function showMsg(text, type = 'success') {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 4000)
  }

  useEffect(() => {
    getAuditRetention()
      .then(res => {
        setForm({
          retentionDays: res.data.retention_days ?? 365,
          archiveEnabled: res.data.archive_enabled ?? false,
        })
      })
      .catch(() => setError(t('auditRetention.loadError')))
      .finally(() => setLoading(false))
  }, [t])

  async function handleSave(e) {
    e.preventDefault()
    if (form.retentionDays < 30) {
      setError(t('auditRetention.retentionMinError'))
      return
    }
    setSaving(true)
    try {
      await updateAuditRetention({
        retention_days: form.retentionDays,
        archive_enabled: form.archiveEnabled,
      })
      showMsg(t('auditRetention.settingsSaved'))
    } catch (err) {
      setError(err.response?.data?.error || t('auditRetention.saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handlePrune() {
    if (!confirm(t('auditRetention.pruneConfirm'))) return
    setPruning(true)
    try {
      const res = await pruneAuditLogs()
      showMsg(t('auditRetention.pruneSuccess', { count: res.data.deleted }))
    } catch {
      setError(t('scanRetention.pruneError'))
    } finally {
      setPruning(false)
    }
  }



  if (loading) return <PageSpinner message={t('auditRetention.loadingSettings')} />

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('auditRetention.title')}</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          {t('auditRetention.subtitle')}
        </p>
      </div>

      <ErrorBanner error={error} onDismiss={() => setError(null)} />
      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}

      {/* Retention Settings Panel */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 max-w-lg mb-6">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">{t('auditRetention.retentionSettingsTitle')}</h2>
        <form onSubmit={handleSave} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('auditRetention.retentionPeriodDaysLower')}
            </label>
            <input
              type="number"
              min="30"
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              value={form.retentionDays}
              onChange={e => setForm(f => ({ ...f, retentionDays: parseInt(e.target.value) || 1 }))}
            />
            <p className="text-xs text-gray-500 mt-1">{t('auditRetention.retentionHint')}</p>
          </div>
          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="archiveEnabled"
              checked={form.archiveEnabled}
              onChange={e => setForm(f => ({ ...f, archiveEnabled: e.target.checked }))}
              className="accent-blue-600"
            />
            <label htmlFor="archiveEnabled" className="text-sm font-medium text-gray-700 dark:text-gray-300">
              {t('auditRetention.archiveEnabled')}
            </label>
          </div>
          <div className="flex items-center gap-3 pt-2">
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
            >
              {saving ? t('common.saving') : t('scanRetention.saveSettings')}
            </button>
            <button
              type="button"
              onClick={handlePrune}
              disabled={pruning}
              className="px-4 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700 disabled:opacity-50"
            >
              {pruning ? t('scanRetention.pruning') : t('scanRetention.runPruneNow')}
            </button>
          </div>
        </form>
      </div>

    </div>
  )
}
