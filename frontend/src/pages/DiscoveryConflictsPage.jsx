import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listDiscoveryConflicts, resolveDiscoveryConflict } from '../api/admin'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

function confidenceColor(score) {
  if (score >= 0.8) return 'text-green-600 dark:text-green-400'
  if (score >= 0.5) return 'text-yellow-600 dark:text-yellow-400'
  return 'text-red-600 dark:text-red-400'
}

export default function DiscoveryConflictsPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [statusFilter, setStatusFilter] = useState('pending')
  const [message, setMessage] = useState(null)

  function showMsg(text, type = 'success') {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 3000)
  }

  const conflictsQuery = useQuery({
    queryKey: ['discovery', 'conflicts', statusFilter],
    queryFn: () => listDiscoveryConflicts(statusFilter || '').then(res => res.data || []),
  })
  const conflicts = conflictsQuery.data ?? []
  const loading = conflictsQuery.isLoading
  const error = conflictsQuery.isError ? t('discoveryConflicts.loadError') : null

  const resolveMutation = useMutation({
    mutationFn: ({ id, action }) => resolveDiscoveryConflict(id, action),
    onSuccess: (_res, { action }) => {
      showMsg(t('discoveryConflicts.resolvedSuccess', { action }))
      queryClient.invalidateQueries({ queryKey: ['discovery', 'conflicts'] })
    },
    onError: () => showMsg(t('discoveryConflicts.resolveError'), 'error'),
  })
  const resolving = resolveMutation.isPending ? resolveMutation.variables?.id : null

  function handleResolve(id, action) {
    resolveMutation.mutate({ id, action })
  }

  const pendingCount = conflicts.filter(c => c.status === 'pending').length

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          {t('discoveryConflicts.title')}
          {statusFilter === 'pending' && pendingCount > 0 && (
            <span className="ml-2 px-2 py-0.5 text-sm rounded-full bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
              {pendingCount}
            </span>
          )}
        </h1>
        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-600 dark:text-gray-400">{t('discoveryConflicts.statusLabel')}</label>
          <select
            value={statusFilter}
            onChange={e => setStatusFilter(e.target.value)}
            className="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            <option value="">{t('discoveryConflicts.all')}</option>
            <option value="pending">{t('discoveryConflicts.pending')}</option>
            <option value="accepted">{t('discoveryConflicts.accepted')}</option>
            <option value="rejected">{t('discoveryConflicts.rejected')}</option>
          </select>
        </div>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${
          message.type === 'error'
            ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
            : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
        }`}>
          {message.text}
        </div>
      )}

      {error && <ErrorBanner message={error} />}

      {loading ? (
        <PageSpinner />
      ) : conflicts.length === 0 ? (
        <p className="text-sm text-gray-500 dark:text-gray-400">{t('discoveryConflicts.noConflictsFound')}</p>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('discoveryConflicts.deviceId')}</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('discoveryConflicts.field')}</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('discoveryConflicts.currentValue')}</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('discoveryConflicts.discoveredValue')}</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('discoveryConflicts.confidence')}</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('discoveryConflicts.source')}</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('delegations.status')}</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300">{t('vrfs.actions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {conflicts.map(conflict => (
                <tr key={conflict.id}>
                  <td className="px-4 py-3 text-gray-900 dark:text-gray-100">{conflict.deviceId}</td>
                  <td className="px-4 py-3 font-mono text-gray-800 dark:text-gray-200">{conflict.fieldName}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                    {conflict.currentValue ?? <span className="italic text-gray-400">{t('discoveryConflicts.none')}</span>}
                  </td>
                  <td className="px-4 py-3 text-gray-900 dark:text-gray-100">{conflict.discoveredValue}</td>
                  <td className={`px-4 py-3 font-semibold ${confidenceColor(conflict.confidenceScore)}`}>
                    {Math.round(conflict.confidenceScore * 100)}%
                  </td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{conflict.source}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${
                      conflict.status === 'pending'
                        ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
                        : conflict.status === 'accepted'
                        ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                        : 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'
                    }`}>
                      {conflict.status}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    {conflict.status === 'pending' && (
                      <div className="flex gap-2">
                        <button
                          onClick={() => handleResolve(conflict.id, 'accepted')}
                          disabled={resolving === conflict.id}
                          className="px-2 py-1 text-xs rounded bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
                        >
                          {t('discoveryConflicts.accept')}
                        </button>
                        <button
                          onClick={() => handleResolve(conflict.id, 'rejected')}
                          disabled={resolving === conflict.id}
                          className="px-2 py-1 text-xs rounded bg-red-600 text-white hover:bg-red-700 disabled:opacity-50"
                        >
                          {t('approvals.reject')}
                        </button>
                      </div>
                    )}
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
