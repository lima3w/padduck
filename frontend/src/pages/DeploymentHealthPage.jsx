import { useState, useEffect, useCallback } from 'react'
import { getSystemHealth } from '../api/client'

function StatusBadge({ status }) {
  const s = (status || '').toLowerCase()
  let cls = 'inline-block px-2 py-0.5 rounded text-xs font-semibold '
  if (s === 'ok' || s === 'healthy') {
    cls += 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
  } else if (s === 'degraded' || s === 'unknown') {
    cls += 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
  } else {
    cls += 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  }
  return <span className={cls}>{status}</span>
}

function SectionHeading({ children }) {
  return (
    <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
      {children}
    </h2>
  )
}

function Card({ children, className = '' }) {
  return (
    <div className={`bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-5 ${className}`}>
      {children}
    </div>
  )
}

export default function DeploymentHealthPage() {
  const [health, setHealth] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const fetchHealth = useCallback(() => {
    setLoading(true)
    setError(null)
    getSystemHealth()
      .then((res) => {
        setHealth(res.data)
      })
      .catch((err) => {
        setError(err?.response?.data?.error || err.message || 'Failed to load system health')
      })
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    fetchHealth()
  }, [fetchHealth])

  return (
    <div className="max-w-4xl mx-auto px-4 py-8 space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          Deployment Health
        </h1>
        <button
          onClick={fetchHealth}
          disabled={loading}
          className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors"
        >
          {loading ? 'Refreshing...' : 'Refresh'}
        </button>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 rounded-lg p-4 text-sm text-red-700 dark:text-red-300">
          {error}
        </div>
      )}

      {/* Panel 1: System Health */}
      <network>
        <SectionHeading>System Health</SectionHeading>
        {loading && !health ? (
          <div className="text-sm text-gray-500 dark:text-gray-400">Loading...</div>
        ) : health ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/* Database */}
            <Card>
              <div className="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-2">
                Database
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={health.database?.status || 'unknown'} />
                {health.database?.detail && (
                  <span className="text-xs text-red-600 dark:text-red-400 truncate" title={health.database.detail}>
                    {health.database.detail}
                  </span>
                )}
              </div>
            </Card>

            {/* Scan Agents */}
            <Card>
              <div className="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-2">
                Scan Agents
              </div>
              {health.scanAgents?.total != null ? (
                health.scanAgents.total === 0 ? (
                  <p className="text-sm text-gray-500 dark:text-gray-400">No agents registered.</p>
                ) : (
                  <div className="space-y-1 text-sm text-gray-700 dark:text-gray-300">
                    <div>Total: <span className="font-medium">{health.scanAgents.total}</span></div>
                    <div className="flex items-center gap-1">
                      Healthy: <StatusBadge status={health.scanAgents.healthy > 0 ? 'healthy' : 'ok'} />
                      <span className="font-medium ml-1">{health.scanAgents.healthy}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      Offline:
                      <span className={`ml-1 font-medium ${health.scanAgents.offline > 0 ? 'text-red-600 dark:text-red-400' : ''}`}>
                        {health.scanAgents.offline}
                      </span>
                    </div>
                  </div>
                )
              ) : (
                <p className="text-sm text-gray-500 dark:text-gray-400">No agents registered.</p>
              )}
            </Card>
          </div>
        ) : null}
      </network>

    </div>
  )
}
