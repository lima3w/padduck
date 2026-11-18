import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { listAutomationExecutions, rollbackAutomationExecution } from '../api/admin'
import { getCachedUser } from '../utils/storageKeys'

const fmt = v => (v ? new Date(v).toLocaleString() : '—')
const REVERSIBLE = new Set(['tag', 'scan'])

const decisionClass = {
  allow: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
  deny: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  manual_review: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
}

export function canRollback(execution) {
  return !execution.rolledBackAt
    && (execution.actionsTaken || []).some(a => a.status === 'success' && REVERSIBLE.has(a.type))
}

// Automation history for one policy (policyId) or one resource (resourceType +
// resourceId). Admin only. With hideWhenEmpty it renders nothing until there is
// history, so it can sit on resource pages without adding noise.
export default function AutomationExecutions({ policyId, resourceType, resourceId, hideWhenEmpty = false }) {
  const { t } = useTranslation()
  const isAdmin = getCachedUser()?.role === 'admin'
  const [items, setItems] = useState(null)
  const [error, setError] = useState('')
  const [confirming, setConfirming] = useState(null)

  const load = useCallback(() => {
    const params = {}
    if (policyId) params.policy_id = policyId
    if (resourceType) params.resource_type = resourceType
    if (resourceId) params.resource_id = resourceId
    return listAutomationExecutions(params)
      .then(res => { setItems(res.data?.executions || []); setError('') })
      .catch(() => { setItems([]); setError(t('automationExecutions.loadError')) })
  }, [policyId, resourceType, resourceId, t])

  useEffect(() => {
    if (isAdmin) load()
  }, [isAdmin, load])

  async function rollback(id) {
    setError('')
    try {
      await rollbackAutomationExecution(id)
      setConfirming(null)
      await load()
    } catch (e) {
      setConfirming(null)
      setError(e?.response?.data?.error?.message || t('automationExecutions.rollbackError'))
    }
  }

  if (!isAdmin || items === null) return null
  if (hideWhenEmpty && items.length === 0 && !error) return null

  return (
    <section aria-label={t('automationExecutions.title')} className="mt-6">
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">{t('automationExecutions.title')}</h3>
      {error && <p className="mb-2 text-sm text-red-600" role="alert">{error}</p>}
      {items.length === 0 ? (
        <p className="text-sm text-gray-500 dark:text-gray-400">{t('automationExecutions.empty')}</p>
      ) : (
        <div className="overflow-x-auto rounded border border-gray-200 dark:border-gray-700">
          <table className="min-w-full text-xs">
            <thead className="bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 uppercase">
              <tr>
                {['when', 'policy', 'resource', 'decision', 'actions', 'source', ''].map(k => (
                  <th key={k} className="px-3 py-2 text-left font-semibold">{k ? t(`automationExecutions.col.${k}`) : ''}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800 bg-white dark:bg-gray-900 text-gray-800 dark:text-gray-200">
              {items.map(e => (
                <tr key={e.id}>
                  <td className="px-3 py-2 whitespace-nowrap">{fmt(e.executedAt)}</td>
                  <td className="px-3 py-2">{e.policyName}</td>
                  <td className="px-3 py-2">{e.resourceName || e.resourceType}</td>
                  <td className="px-3 py-2">
                    <span className={`px-2 py-0.5 rounded-full ${decisionClass[e.decision] || ''}`}>{e.decision.replace('_', ' ')}</span>
                  </td>
                  <td className="px-3 py-2">
                    {(e.actionsTaken || []).length === 0 && '—'}
                    <ul className="space-y-0.5">
                      {(e.actionsTaken || []).map((a, i) => (
                        <li key={i}>
                          <span className="font-mono">{a.type}</span>{' '}
                          <span className={a.status === 'success' ? 'text-green-700 dark:text-green-400' : 'text-red-600'}>
                            {t(`automationExecutions.status.${a.status}`, a.status)}
                          </span>
                          {a.error && <span className="text-gray-500"> ({a.error})</span>}
                          {a.rollback && (
                            <span className="text-gray-500"> — {t(`automationExecutions.rollbackStatus.${a.rollback.status}`, a.rollback.status)}{a.rollback.message ? `: ${a.rollback.message}` : ''}</span>
                          )}
                        </li>
                      ))}
                    </ul>
                  </td>
                  <td className="px-3 py-2">{e.triggeredBy}</td>
                  <td className="px-3 py-2 text-right whitespace-nowrap">
                    {e.rolledBackAt && <span className="text-gray-500">{t('automationExecutions.rolledBack')}</span>}
                    {canRollback(e) && (confirming === e.id ? (
                      <span className="space-x-2">
                        <button onClick={() => rollback(e.id)} className="text-red-600 hover:text-red-800">{t('automationExecutions.confirmRollback')}</button>
                        <button onClick={() => setConfirming(null)} className="text-gray-500 hover:text-gray-700">{t('common.cancel')}</button>
                      </span>
                    ) : (
                      <button onClick={() => setConfirming(e.id)} className="text-blue-600 hover:text-blue-800 dark:text-blue-400">{t('automationExecutions.rollback')}</button>
                    ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
