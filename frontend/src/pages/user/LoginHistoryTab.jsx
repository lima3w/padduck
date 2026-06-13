import { useState, useEffect } from 'react'
import * as client from '../../api/auth'

export default function LoginHistoryTab() {
  const [history, setHistory] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    loadHistory()
  }, [])

  const loadHistory = async () => {
    setLoading(true)
    setError('')
    try {
      const res = await client.getLoginHistory()
      setHistory(res.data)
    } catch {
      setError('Failed to load login history.')
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (iso) => {
    const d = new Date(iso)
    return d.toLocaleString()
  }

  const knownIPs = history.filter((a) => a.success).map((a) => a.ipAddress).filter(Boolean)
  const knownIPSet = new Set(knownIPs)

  return (
    <div className="max-w-2xl space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Login History</h2>
        <p className="text-sm text-gray-600 mb-4">Recent login attempts on your account (last 20).</p>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      {loading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : history.length === 0 ? (
        <p className="text-sm text-gray-500">No login history yet.</p>
      ) : (
        <div className="divide-y divide-gray-200 border border-gray-200 rounded">
          {history.map((attempt) => {
            const isNew = attempt.ipAddress && !knownIPSet.has(attempt.ipAddress) && !attempt.success
            return (
              <div key={attempt.id} className={`px-4 py-3 flex items-start gap-3 ${!attempt.success ? 'bg-red-50' : ''}`}>
                <div className={`mt-0.5 w-2 h-2 rounded-full flex-shrink-0 ${attempt.success ? 'bg-green-500' : 'bg-red-500'}`} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className={`text-sm font-medium ${attempt.success ? 'text-green-800' : 'text-red-800'}`}>
                      {attempt.success ? 'Successful login' : 'Failed login attempt'}
                    </span>
                    {isNew && (
                      <span className="text-xs px-1.5 py-0.5 bg-yellow-100 text-yellow-800 border border-yellow-200 rounded">New IP</span>
                    )}
                  </div>
                  <p className="text-xs text-gray-500 mt-0.5">
                    {formatDate(attempt.createdAt)}
                    {attempt.ipAddress ? ` · ${attempt.ipAddress}` : ''}
                  </p>
                  {attempt.failureReason && (
                    <p className="text-xs text-red-600 mt-0.5">{attempt.failureReason}</p>
                  )}
                  {attempt.userAgent && (
                    <p className="text-xs text-gray-400 mt-0.5 truncate">{attempt.userAgent}</p>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
