import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { listIntents, getIntentDiff, approveIntent, rejectIntent } from '../api/admin'

const STATUS_STYLES = {
  pending: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
  approved: 'bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300',
  applied: 'bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300',
  rejected: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
  failed: 'bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300',
}

const show = v => (v === null || v === undefined || v === '' ? '—' : typeof v === 'object' ? JSON.stringify(v) : String(v))

// Shows a resource's pending desired-state change (with a field diff) and its
// intent history. Admins can approve or reject a pending intent in place.
export default function IntentPanel({ resourceType, resourceId, isAdmin = false, onApplied }) {
  const { t } = useTranslation()
  const [intents, setIntents] = useState([])
  const [diff, setDiff] = useState(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    if (!resourceType || !resourceId) return
    listIntents({ resource_type: resourceType, resource_id: resourceId })
      .then(res => setIntents(Array.isArray(res.data) ? res.data : []))
      .catch(() => setIntents([]))
    getIntentDiff(resourceType, resourceId)
      .then(res => setDiff(res.data))
      .catch(() => setDiff(null))
  }, [resourceType, resourceId])

  useEffect(() => { load() }, [load])

  const review = async (fn, intent) => {
    setBusy(true)
    setError('')
    try {
      await fn(intent.id)
      load()
      if (onApplied) onApplied()
    } catch (e) {
      setError(e.response?.data?.error?.message || t('intentPanel.actionFailed'))
      load()
    } finally {
      setBusy(false)
    }
  }

  const pending = diff?.pending
  if (!pending && intents.length === 0) return null

  return (
    <div className="space-y-3">
      {pending && (
        <div role="status" className="rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-900/20 dark:border-amber-700 p-4 text-sm">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="font-semibold text-amber-900 dark:text-amber-200">
                {t('intentPanel.pendingTitle', { operation: t(`intentPanel.operation.${pending.operation}`) })}
              </p>
              <p className="text-amber-800 dark:text-amber-300 text-xs mt-0.5">
                {t('intentPanel.submittedAt', { date: new Date(pending.submittedAt).toLocaleString() })}
              </p>
            </div>
            {isAdmin && (
              <div className="flex gap-2 shrink-0">
                <button disabled={busy} onClick={() => review(id => approveIntent(id), pending)}
                  className="px-3 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50">
                  {t('intentPanel.approve')}
                </button>
                <button disabled={busy} onClick={() => review(id => rejectIntent(id), pending)}
                  className="px-3 py-1 text-xs bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50">
                  {t('intentPanel.reject')}
                </button>
              </div>
            )}
          </div>
          {diff.changes?.length > 0 && (
            <table className="mt-3 w-full text-xs">
              <thead>
                <tr className="text-left text-amber-900 dark:text-amber-200">
                  <th className="py-1 pr-3">{t('intentPanel.field')}</th>
                  <th className="py-1 pr-3">{t('intentPanel.current')}</th>
                  <th className="py-1">{t('intentPanel.desired')}</th>
                </tr>
              </thead>
              <tbody>
                {diff.changes.map(c => (
                  <tr key={c.field} className="border-t border-amber-200 dark:border-amber-800">
                    <td className="py-1 pr-3 font-mono">{c.field}</td>
                    <td className="py-1 pr-3 text-gray-700 dark:text-gray-300">{show(c.current)}</td>
                    <td className="py-1 font-medium text-gray-900 dark:text-gray-100">{show(c.desired)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          {error && <p className="mt-2 text-xs text-red-700 dark:text-red-400">{error}</p>}
        </div>
      )}

      {intents.length > 0 && (
        <details className="bg-white dark:bg-gray-800 rounded-lg shadow">
          <summary className="cursor-pointer px-4 py-3 text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">
            {t('intentPanel.historyTitle', { count: intents.length })}
          </summary>
          <ul className="px-4 pb-3 space-y-2">
            {intents.map(i => (
              <li key={i.id} className="flex items-start gap-3 text-xs">
                <span className={`px-2 py-0.5 rounded-full font-medium ${STATUS_STYLES[i.status] || STATUS_STYLES.rejected}`}>
                  {t(`intentPanel.status.${i.status}`, { defaultValue: i.status })}
                </span>
                <span className="flex-1 text-gray-700 dark:text-gray-300">
                  {t(`intentPanel.operation.${i.operation}`)}
                  {i.reviewerNote ? ` — ${i.reviewerNote}` : ''}
                </span>
                <span className="text-gray-400 dark:text-gray-500 whitespace-nowrap">{new Date(i.submittedAt).toLocaleString()}</span>
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  )
}
