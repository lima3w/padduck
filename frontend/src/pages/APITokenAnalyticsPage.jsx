import { useState, useEffect } from 'react'
import { getApiTokenAnalytics } from '../api/admin'

export default function APITokenAnalyticsPage() {
  const [tokens, setTokens] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    getApiTokenAnalytics()
      .then(res => setTokens(res.data || []))
      .catch(err => setError(err.response?.data?.error || 'Failed to load analytics'))
      .finally(() => setLoading(false))
  }, [])

  function fmtDate(val) {
    if (!val) return '—'
    return new Date(val).toLocaleString()
  }

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">API Token Analytics</h1>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Usage statistics for all API tokens. Rate limit shown is the global per-minute cap.
      </p>

      {loading && <p className="text-sm text-gray-500 dark:text-gray-400">Loading…</p>}
      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      {!loading && !error && (
        <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800">
              <tr>
                {['Token Name', 'User', 'Scope', 'Usage Count', 'Last Used', 'Last IP', 'Expires', 'Rate Limit/min', 'Rotated'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {tokens.length === 0 ? (
                <tr><td colSpan={9} className="px-4 py-6 text-center text-gray-400">No tokens found.</td></tr>
              ) : tokens.map(t => (
                <tr key={t.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{t.name}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{t.username || t.userId}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{t.scope || 'full'}</td>
                  <td className="px-4 py-3 text-gray-900 dark:text-gray-100 font-mono">{t.usageCount ?? 0}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">{fmtDate(t.lastUsedAt)}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 font-mono text-xs">{t.lastUsedIp || '—'}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">{fmtDate(t.expiresAt)}</td>
                  <td className="px-4 py-3 text-gray-900 dark:text-gray-100 font-mono">{t.rateLimitPerMinute ?? '—'}</td>
                  <td className="px-4 py-3">
                    {t.isRotated
                      ? <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-700">Yes</span>
                      : <span className="text-gray-400">—</span>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
