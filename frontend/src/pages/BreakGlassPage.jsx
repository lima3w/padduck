import { useEffect, useState, useCallback } from 'react'
import { getBreakGlassStatus, activateBreakGlass, endBreakGlass } from '../api/admin'

function formatDatetime(str) {
  if (!str) return '—'
  return new Date(str).toLocaleString(undefined, {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

function formatDuration(startStr, endStr) {
  if (!endStr) return 'Active'
  const ms = new Date(endStr) - new Date(startStr)
  if (ms < 0) return '—'
  const totalSecs = Math.floor(ms / 1000)
  const h = Math.floor(totalSecs / 3600)
  const m = Math.floor((totalSecs % 3600) / 60)
  const s = totalSecs % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

function sessionStatus(session) {
  if (session.isActive) {
    return (
      <span className="px-2 py-0.5 rounded text-xs font-semibold bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300">
        Active
      </span>
    )
  }
  if (session.endedAt) {
    return (
      <span className="px-2 py-0.5 rounded text-xs font-semibold bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400">
        Ended
      </span>
    )
  }
  return (
    <span className="px-2 py-0.5 rounded text-xs font-semibold bg-yellow-100 text-yellow-700 dark:bg-yellow-900/50 dark:text-yellow-300">
      Expired
    </span>
  )
}

function TimeRemaining({ expiresAt }) {
  const [remaining, setRemaining] = useState('')

  useEffect(() => {
    function calc() {
      const diff = new Date(expiresAt) - Date.now()
      if (diff <= 0) { setRemaining('Expired'); return }
      const totalSecs = Math.floor(diff / 1000)
      const h = Math.floor(totalSecs / 3600)
      const m = Math.floor((totalSecs % 3600) / 60)
      const s = totalSecs % 60
      if (h > 0) setRemaining(`${h}h ${m}m ${s}s`)
      else if (m > 0) setRemaining(`${m}m ${s}s`)
      else setRemaining(`${s}s`)
    }
    calc()
    const id = setInterval(calc, 1000)
    return () => clearInterval(id)
  }, [expiresAt])

  return <span className="font-mono font-semibold">{remaining}</span>
}

export default function BreakGlassPage() {
  const [active, setActive] = useState(null)
  const [history, setHistory] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [justification, setJustification] = useState('')
  const [activating, setActivating] = useState(false)
  const [ending, setEnding] = useState(false)
  const [actionError, setActionError] = useState(null)

  const fetchStatus = useCallback(() => {
    setLoading(true)
    setError(null)
    getBreakGlassStatus()
      .then(res => {
        setActive(res.data?.active ?? null)
        setHistory(res.data?.history ?? [])
      })
      .catch(() => setError('Failed to load break-glass status'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    fetchStatus()
  }, [fetchStatus])

  async function handleActivate(e) {
    e.preventDefault()
    setActionError(null)
    if (justification.trim().length < 10) {
      setActionError('Justification must be at least 10 characters')
      return
    }
    if (!window.confirm('This action is fully audited. A break-glass session will be recorded and visible to all administrators. Continue?')) {
      return
    }
    setActivating(true)
    try {
      await activateBreakGlass(justification.trim())
      setJustification('')
      fetchStatus()
    } catch (err) {
      setActionError(err.response?.data?.error || 'Failed to activate break-glass session')
    } finally {
      setActivating(false)
    }
  }

  async function handleEnd() {
    setActionError(null)
    if (!window.confirm('End the current break-glass session?')) return
    setEnding(true)
    try {
      await endBreakGlass()
      fetchStatus()
    } catch (err) {
      setActionError(err.response?.data?.error || 'Failed to end break-glass session')
    } finally {
      setEnding(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
        Loading...
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 max-w-4xl mx-auto">
        <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-6 text-center text-sm text-red-700 dark:text-red-300">
          {error}
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Break-Glass Emergency Access</h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Enable a temporary emergency access session. All actions are fully audited.
        </p>
      </div>

      {actionError && (
        <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-red-700 dark:text-red-300">
          {actionError}
        </div>
      )}

      {/* Current Status */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Current Status</h2>
        {active ? (
          <div className="rounded-lg border-2 border-red-500 bg-red-50 dark:bg-red-900/20 p-4 space-y-3">
            <div className="flex items-center gap-3">
              <span className="text-lg font-bold text-red-700 dark:text-red-300 uppercase tracking-wide">
                BREAK-GLASS ACTIVE
              </span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">Activated by user ID:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{active.initiatedByUserId}</span>
              </div>
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">Started:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{formatDatetime(active.createdAt)}</span>
              </div>
              <div className="sm:col-span-2">
                <span className="font-medium text-gray-700 dark:text-gray-300">Justification:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{active.justification}</span>
              </div>
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">Time remaining:</span>{' '}
                <TimeRemaining expiresAt={active.expiresAt} />
              </div>
            </div>
            <div className="pt-2">
              <button
                onClick={handleEnd}
                disabled={ending}
                className="px-4 py-2 rounded text-sm font-medium text-white bg-red-600 hover:bg-red-700 disabled:opacity-50 transition-colors"
              >
                {ending ? 'Ending...' : 'End Session'}
              </button>
            </div>
          </div>
        ) : (
          <div className="rounded-lg border border-green-300 dark:border-green-700 bg-green-50 dark:bg-green-900/20 px-4 py-3 text-sm text-green-800 dark:text-green-300 font-medium">
            No active break-glass session.
          </div>
        )}
      </div>

      {/* Activate Form (only shown when no active session) */}
      {!active && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">Activate Break-Glass</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            The session will last 1 hour. Provide a justification explaining why emergency access is required.
          </p>
          <form onSubmit={handleActivate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Justification <span className="text-red-500">*</span>
                <span className="ml-1 text-gray-400 dark:text-gray-500 font-normal">(minimum 10 characters)</span>
              </label>
              <textarea
                value={justification}
                onChange={e => setJustification(e.target.value)}
                rows={4}
                placeholder="Describe the emergency requiring break-glass access..."
                className="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-red-500"
              />
            </div>
            <button
              type="submit"
              disabled={activating}
              className="px-4 py-2 rounded text-sm font-medium text-white bg-red-600 hover:bg-red-700 disabled:opacity-50 transition-colors"
            >
              {activating ? 'Activating...' : 'Activate Break-Glass'}
            </button>
          </form>
        </div>
      )}

      {/* Session History */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Session History</h2>
        {history.length === 0 ? (
          <div className="rounded bg-white dark:bg-gray-800 shadow px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
            No break-glass sessions recorded.
          </div>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Date</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Initiated By</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Justification</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Duration</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                  {history.map(s => (
                    <tr key={s.id} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                      <td className="px-4 py-3 text-gray-700 dark:text-gray-300 whitespace-nowrap">
                        {formatDatetime(s.createdAt)}
                      </td>
                      <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                        User {s.initiatedByUserId}
                      </td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-xs truncate" title={s.justification}>
                        {s.justification}
                      </td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                        {formatDuration(s.createdAt, s.endedAt || (s.isActive ? null : s.expiresAt))}
                      </td>
                      <td className="px-4 py-3">
                        {sessionStatus(s)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
