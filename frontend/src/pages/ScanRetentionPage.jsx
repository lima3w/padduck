import { useState, useEffect } from 'react'
import { getScanRetention, updateScanRetention, runScanRetentionPrune } from '../api/client'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

export default function ScanRetentionPage() {
  const [settings, setSettings] = useState(null)
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
        setSettings(res.data)
        setForm({
          rawHistoryDays: res.data.rawHistoryDays,
          rollupEnabled: res.data.rollupEnabled,
          rollupAfterDays: res.data.rollupAfterDays,
        })
      })
      .catch(() => setError('Failed to load retention settings'))
      .finally(() => setLoading(false))
  }, [])

  async function handleSave(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const res = await updateScanRetention({
        raw_history_days: form.rawHistoryDays,
        rollup_enabled: form.rollupEnabled,
        rollup_after_days: form.rollupAfterDays,
      })
      setSettings(res.data)
      showMsg('Settings saved')
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  async function handlePrune() {
    if (!confirm('Run prune now? This will delete scan data older than the configured retention period.')) return
    setPruning(true)
    try {
      const res = await runScanRetentionPrune()
      showMsg(`Pruned ${res.data.pruned} records`)
    } catch {
      setError('Prune failed')
    } finally {
      setPruning(false)
    }
  }

  if (loading) return <PageSpinner message="Loading retention settings..." />

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Scan Retention</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Control how long raw scan history is retained before automatic cleanup.
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
              Raw History Retention (days)
            </label>
            <input
              type="number" min="1" max="3650"
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              value={form.rawHistoryDays}
              onChange={e => setForm(f => ({ ...f, rawHistoryDays: parseInt(e.target.value) || 1 }))}
            />
            <p className="text-xs text-gray-500 mt-1">Scan results and run history older than this will be deleted nightly.</p>
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
              Enable rollup compression
            </label>
          </div>
          {form.rollupEnabled && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Rollup after (days)
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
              {saving ? 'Saving...' : 'Save Settings'}
            </button>
            <button type="button" onClick={handlePrune} disabled={pruning} className="px-4 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700 disabled:opacity-50">
              {pruning ? 'Pruning...' : 'Run Prune Now'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
