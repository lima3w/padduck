import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getApiTokenAnalytics } from '../api/admin'

export default function APITokenAnalyticsPage() {
  const { t } = useTranslation()
  const [tokens, setTokens] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    getApiTokenAnalytics()
      .then(res => setTokens(res.data || []))
      .catch(err => setError(err.response?.data?.error || t('apiTokenAnalyticsPage.loadFailed')))
      .finally(() => setLoading(false))
  }, [t])

  function fmtDate(val) {
    if (!val) return '—'
    return new Date(val).toLocaleString()
  }

  const columns = [
    t('apiTokenAnalyticsPage.tokenNameColumn'),
    t('auditLog.user'),
    t('adminIntegrations.scopeColumn'),
    t('apiTokenAnalyticsPage.usageCountColumn'),
    t('identityPolicies.lastUsed'),
    t('apiTokenAnalyticsPage.lastIpColumn'),
    t('apiTokenAnalyticsPage.expiresColumn'),
    t('apiTokenAnalyticsPage.rateLimitColumn'),
    t('apiTokenAnalyticsPage.rotatedColumn'),
  ]

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('apiTokenAnalyticsPage.title')}</h1>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        {t('apiTokenAnalyticsPage.subtitle')}
      </p>

      {loading && <p className="text-sm text-gray-500 dark:text-gray-400">{t('common.loading')}</p>}
      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      {!loading && !error && (
        <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800">
              <tr>
                {columns.map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {tokens.length === 0 ? (
                <tr><td colSpan={9} className="px-4 py-6 text-center text-gray-400">{t('apiTokenAnalyticsPage.noTokensFound')}</td></tr>
              ) : tokens.map(tok => (
                <tr key={tok.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{tok.name}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{tok.username || tok.userId}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{tok.scope || t('apiTokenAnalyticsPage.fullScope')}</td>
                  <td className="px-4 py-3 text-gray-900 dark:text-gray-100 font-mono">{tok.usageCount ?? 0}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">{fmtDate(tok.lastUsedAt)}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 font-mono text-xs">{tok.lastUsedIp || '—'}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">{fmtDate(tok.expiresAt)}</td>
                  <td className="px-4 py-3 text-gray-900 dark:text-gray-100 font-mono">{tok.rateLimitPerMinute ?? '—'}</td>
                  <td className="px-4 py-3">
                    {tok.isRotated
                      ? <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-700">{t('common.yes')}</span>
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
