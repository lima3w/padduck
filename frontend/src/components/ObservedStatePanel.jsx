import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getObservedState } from '../api/discovery'

export default function ObservedStatePanel({ resourceType, resourceId }) {
  const { t } = useTranslation()
  const [state, setState] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!resourceType || !resourceId) return
    setLoading(true)
    getObservedState(resourceType, resourceId)
      .then(res => setState(res.data))
      .catch(() => setState(null))
      .finally(() => setLoading(false))
  }, [resourceType, resourceId])

  if (loading) return null

  const data = state?.observedData || {}

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <h3 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-3">
        {t('observedStatePanel.title')}
      </h3>

      {!state ? (
        <p className="text-sm text-gray-400 dark:text-gray-500">{t('observedStatePanel.noScanData')}</p>
      ) : (
        <div className="space-y-2 text-sm">
          <div className="flex items-center gap-2">
            <span className={`inline-block w-2 h-2 rounded-full ${data.isAlive ? 'bg-green-500' : 'bg-gray-400'}`} />
            <span className="text-gray-600 dark:text-gray-400">{data.isAlive ? t('observedStatePanel.alive') : t('observedStatePanel.notResponding')} {t('observedStatePanel.lastSeenByScannerSuffix')}</span>
          </div>
          {data.ptrRecord && (
            <div className="text-gray-600 dark:text-gray-400">{t('observedStatePanel.ptrLabel')} <span className="font-mono text-xs text-gray-800 dark:text-gray-200">{data.ptrRecord}</span></div>
          )}
          {data.snmpHostname && (
            <div className="text-gray-600 dark:text-gray-400">{t('observedStatePanel.snmpHostnameLabel')} <span className="font-medium text-gray-800 dark:text-gray-200">{data.snmpHostname}</span></div>
          )}
          {data.snmpMacAddress && (
            <div className="text-gray-600 dark:text-gray-400">{t('observedStatePanel.snmpMacLabel')} <span className="font-mono text-xs text-gray-800 dark:text-gray-200">{data.snmpMacAddress}</span></div>
          )}
          {data.responseTimeMs > 0 && (
            <div className="text-gray-600 dark:text-gray-400">{t('observedStatePanel.responseTimeLabel')} <span className="text-gray-800 dark:text-gray-200">{data.responseTimeMs} ms</span></div>
          )}
          {data.openPorts?.length > 0 && (
            <div className="text-gray-600 dark:text-gray-400">
              {t('observedStatePanel.openPortsLabel')} <span className="font-mono text-xs text-gray-800 dark:text-gray-200">{data.openPorts.join(', ')}</span>
            </div>
          )}
          <div className="border-t dark:border-gray-700 pt-2 mt-2 text-xs text-gray-400 dark:text-gray-500 flex justify-between">
            <span>{t('observedStatePanel.sourceLabel')} {state.source}</span>
            <span>{t('observedStatePanel.lastSeenLabel')} {state.lastSeenAt ? new Date(state.lastSeenAt).toLocaleString() : '—'}</span>
          </div>
        </div>
      )}
    </div>
  )
}
