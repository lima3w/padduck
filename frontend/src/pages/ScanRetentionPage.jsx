import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getScanRetention, updateScanRetention, runScanRetentionPrune } from '../api/admin'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

export default function ScanRetentionPage() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [pruning, setPruning] = useState(false)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [form, setForm] = useState({ rawHistoryDays: 90, rollupEnabled: false, rollupAfterDays: 30 })

  function showMsg(text, type = 'success') {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 3000)
  }

  useEffect(() => {
    getScanRetention()
      .then(res => {
        setForm({
          rawHistoryDays: res.data.rawHistoryDays,
          rollupEnabled: res.data.rollupEnabled,
          rollupAfterDays: res.data.rollupAfterDays,
        })
      })
      .catch(() => setError(t('scanRetention.loadError')))
      .finally(() => setLoading(false))
  }, [t])

  async function handleSave(e) {
    e.preventDefault()
    setSaving(true)
    try {
      await updateScanRetention({
        raw_history_days: form.rawHistoryDays,
        rollup_enabled: form.rollupEnabled,
        rollup_after_days: form.rollupAfterDays,
      })
      showMsg(t('scanRetention.settingsSaved'))
    } catch (err) {
      setError(err.response?.data?.error || t('scanRetention.saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handlePrune() {
    if (!confirm(t('scanRetention.pruneConfirm'))) return
    setPruning(true)
    try {
      const res = await runScanRetentionPrune()
      showMsg(t('scanRetention.pruneSuccess', { count: res.data.pruned }))
    } catch {
      setError(t('scanRetention.pruneError'))
    } finally {
      setPruning(false)
    }
  }

  if (loading) return <PageSpinner message={t('scanRetention.loadingSettings')} />

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('scanRetention.title')}</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          {t('scanRetention.subtitle')}
        </p>
      </div>
      <ErrorBanner error={error} onDismiss={() => setError(null)} />
      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 max-w-lg">
        <form onSubmit={handleSave} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('scanRetention.rawHistoryRetentionDays')}
            </label>
            <input
              type="number" min="1" max="3650"
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              value={form.rawHistoryDays}
              onChange={e => setForm(f => ({ ...f, rawHistoryDays: parseInt(e.target.value) || 1 }))}
            />
            <p className="text-xs text-gray-500 mt-1">{t('scanRetention.rawHistoryHint')}</p>
          </div>
          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="rollupEnabled"
              checked={form.rollupEnabled}
              onChange={e => setForm(f => ({ ...f, rollupEnabled: e.target.checked }))}
              className="accent-blue-600"
            />
            <label htmlFor="rollupEnabled" className="text-sm font-medium text-gray-700 dark:text-gray-300">
              {t('scanRetention.enableRollupCompression')}
            </label>
          </div>
          {form.rollupEnabled && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('scanRetention.rollupAfterDays')}
              </label>
              <input
                type="number" min="1"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.rollupAfterDays}
                onChange={e => setForm(f => ({ ...f, rollupAfterDays: parseInt(e.target.value) || 1 }))}
              />
            </div>
          )}
          <div className="flex items-center gap-3 pt-2">
            <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
              {saving ? t('common.saving') : t('scanRetention.saveSettings')}
            </button>
            <button type="button" onClick={handlePrune} disabled={pruning} className="px-4 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700 disabled:opacity-50">
              {pruning ? t('scanRetention.pruning') : t('scanRetention.runPruneNow')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
