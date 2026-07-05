import { useState, useEffect } from 'react'
import { getObservedState } from '../api/discovery'

export default function ObservedStatePanel({ resourceType, resourceId }) {
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
        Observed State
      </h3>

      {!state ? (
        <p className="text-sm text-gray-400 dark:text-gray-500">No scan data observed for this resource yet.</p>
      ) : (
        <div className="space-y-2 text-sm">
          <div className="flex items-center gap-2">
            <span className={`inline-block w-2 h-2 rounded-full ${data.isAlive ? 'bg-green-500' : 'bg-gray-400'}`} />
            <span className="text-gray-600 dark:text-gray-400">{data.isAlive ? 'Alive' : 'Not responding'} (last seen by scanner)</span>
          </div>
          {data.ptrRecord && (
            <div className="text-gray-600 dark:text-gray-400">PTR: <span className="font-mono text-xs text-gray-800 dark:text-gray-200">{data.ptrRecord}</span></div>
          )}
          {data.snmpHostname && (
            <div className="text-gray-600 dark:text-gray-400">SNMP hostname: <span className="font-medium text-gray-800 dark:text-gray-200">{data.snmpHostname}</span></div>
          )}
          {data.snmpMacAddress && (
            <div className="text-gray-600 dark:text-gray-400">SNMP MAC: <span className="font-mono text-xs text-gray-800 dark:text-gray-200">{data.snmpMacAddress}</span></div>
          )}
          {data.responseTimeMs > 0 && (
            <div className="text-gray-600 dark:text-gray-400">Response time: <span className="text-gray-800 dark:text-gray-200">{data.responseTimeMs} ms</span></div>
          )}
          {data.openPorts?.length > 0 && (
            <div className="text-gray-600 dark:text-gray-400">
              Open ports: <span className="font-mono text-xs text-gray-800 dark:text-gray-200">{data.openPorts.join(', ')}</span>
            </div>
          )}
          <div className="border-t dark:border-gray-700 pt-2 mt-2 text-xs text-gray-400 dark:text-gray-500 flex justify-between">
            <span>Source: {state.source}</span>
            <span>Last seen: {state.lastSeenAt ? new Date(state.lastSeenAt).toLocaleString() : '—'}</span>
          </div>
        </div>
      )}
    </div>
  )
}
