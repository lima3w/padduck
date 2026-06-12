import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listTopologyHints, updateTopologyHintStatus } from '../api/admin'

const STATUS_FILTERS = [
  { label: 'All', value: '' },
  { label: 'Suggested', value: 'suggested' },
  { label: 'Confirmed', value: 'confirmed' },
  { label: 'Dismissed', value: 'dismissed' },
]

function confidenceBadge(score) {
  const pct = Math.round(score * 100)
  let cls = 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
  if (score >= 0.8) cls = 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
  else if (score >= 0.5) cls = 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-semibold ${cls}`}>
      {pct}%
    </span>
  )
}

function statusBadge(status) {
  const map = {
    suggested: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
    confirmed: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
    dismissed: 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400',
  }
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-semibold capitalize ${map[status] || ''}`}>
      {status}
    </span>
  )
}

export default function TopologyHintsPage() {
  const queryClient = useQueryClient()
  const [statusFilter, setStatusFilter] = useState('')
  const [actionError, setActionError] = useState(null)

  const hintsQuery = useQuery({
    queryKey: ['topology', 'hints', statusFilter],
    queryFn: () => listTopologyHints(statusFilter).then(res => res.data?.hints ?? []),
  })
  const hints = hintsQuery.data ?? []
  const loading = hintsQuery.isLoading
  const error = hintsQuery.isError ? 'Failed to load topology hints' : null

  const statusMutation = useMutation({
    mutationFn: ({ id, newStatus }) => updateTopologyHintStatus(id, newStatus),
    onSuccess: () => {
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['topology', 'hints'] })
    },
    onError: (err, { id }) => {
      setActionError(err.response?.data?.error || `Failed to update hint #${id}`)
    },
  })

  function handleStatusUpdate(id, newStatus) {
    setActionError(null)
    statusMutation.mutate({ id, newStatus })
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Topology Hints
            {!loading && (
              <span className="ml-2 text-base font-normal text-gray-500 dark:text-gray-400">
                ({hints.length})
              </span>
            )}
          </h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Suggested relationships inferred from discovery and inventory data
          </p>
        </div>
        <div className="flex gap-1">
          {STATUS_FILTERS.map(f => (
            <button
              key={f.value}
              onClick={() => setStatusFilter(f.value)}
              className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                statusFilter === f.value
                  ? 'bg-blue-600 text-white'
                  : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {actionError && (
        <div className="mb-4 rounded bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-red-700 dark:text-red-300">
          {actionError}
        </div>
      )}

      {loading ? (
        <div className="flex min-h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
          Loading...
        </div>
      ) : error ? (
        <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-6 text-center text-sm text-red-700 dark:text-red-300">
          {error}
        </div>
      ) : hints.length === 0 ? (
        <div className="rounded bg-white dark:bg-gray-800 shadow px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
          No topology hints found
          {statusFilter ? ` with status "${statusFilter}"` : ''}.
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Source</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Target</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Hint Type</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Confidence</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Evidence</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Status</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                {hints.map(hint => (
                  <tr key={hint.id} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                    <td className="px-4 py-3 text-gray-900 dark:text-gray-100 font-mono text-xs">
                      <span className="font-semibold">{hint.sourceType}</span>:{hint.sourceId}
                    </td>
                    <td className="px-4 py-3 text-gray-900 dark:text-gray-100 font-mono text-xs">
                      <span className="font-semibold">{hint.targetType}</span>:{hint.targetId}
                    </td>
                    <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                      {hint.hintType}
                    </td>
                    <td className="px-4 py-3">
                      {confidenceBadge(hint.confidenceScore)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-xs truncate" title={hint.evidence ?? ''}>
                      {hint.evidence || <span className="text-gray-400 dark:text-gray-600">—</span>}
                    </td>
                    <td className="px-4 py-3">
                      {statusBadge(hint.status)}
                    </td>
                    <td className="px-4 py-3">
                      {hint.status === 'suggested' && (
                        <div className="flex gap-2">
                          <button
                            onClick={() => handleStatusUpdate(hint.id, 'confirmed')}
                            className="px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-700 hover:bg-green-200 dark:bg-green-900/40 dark:text-green-300 dark:hover:bg-green-900/70 transition-colors"
                          >
                            Confirm
                          </button>
                          <button
                            onClick={() => handleStatusUpdate(hint.id, 'dismissed')}
                            className="px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-400 dark:hover:bg-gray-600 transition-colors"
                          >
                            Dismiss
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
