import { useState, useEffect, useCallback } from 'react'
import { getSystemHealth, downloadBackup } from '../api/client'

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
  const [backingUp, setBackingUp] = useState(false)

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

  async function handleDownloadBackup() {
    setBackingUp(true)
    try {
      const res = await downloadBackup()
      const url = URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement('a')
      const cd = res.headers['content-disposition'] || ''
      const match = cd.match(/filename="([^"]+)"/)
      a.href = url
      a.download = match ? match[1] : 'padduck-backup.sql'
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      alert('Backup failed: ' + (err?.response?.data?.error || err.message))
    } finally {
      setBackingUp(false)
    }
  }

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
      <section>
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
              {health.scan_agents?.total != null ? (
                health.scan_agents.total === 0 ? (
                  <p className="text-sm text-gray-500 dark:text-gray-400">No agents registered.</p>
                ) : (
                  <div className="space-y-1 text-sm text-gray-700 dark:text-gray-300">
                    <div>Total: <span className="font-medium">{health.scan_agents.total}</span></div>
                    <div className="flex items-center gap-1">
                      Healthy: <StatusBadge status={health.scan_agents.healthy > 0 ? 'healthy' : 'ok'} />
                      <span className="font-medium ml-1">{health.scan_agents.healthy}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      Offline:
                      <span className={`ml-1 font-medium ${health.scan_agents.offline > 0 ? 'text-red-600 dark:text-red-400' : ''}`}>
                        {health.scan_agents.offline}
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
      </section>

      {/* Panel 2: Backup & Restore */}
      <section>
        <SectionHeading>Backup &amp; Restore</SectionHeading>
        <Card>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
            Follow these steps when conducting a restore rehearsal to verify backup integrity
            and validate your recovery procedure.
          </p>
          <div className="mb-4">
            <button
              onClick={handleDownloadBackup}
              disabled={backingUp}
              className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              {backingUp ? 'Generating...' : 'Download Backup (.sql)'}
            </button>
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Downloads a full pg_dump of the database.
            </p>
          </div>
          {loading && !health ? (
            <div className="text-sm text-gray-500 dark:text-gray-400">Loading...</div>
          ) : health?.backup_notes ? (
            <ol className="space-y-3">
              {health.backup_notes.map((note) => (
                <li key={note.step} className="flex gap-3">
                  <span className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-600 text-white text-xs font-bold flex items-center justify-center">
                    {note.step}
                  </span>
                  <div>
                    <div className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                      {note.action}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">
                      {note.detail}
                    </div>
                  </div>
                </li>
              ))}
            </ol>
          ) : (
            <div className="text-sm text-gray-500 dark:text-gray-400">No backup notes available.</div>
          )}
        </Card>
      </section>

    </div>
  )
}
