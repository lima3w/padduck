import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listDriftItems, acceptDrift, dismissDrift, escalateDrift } from '../api/discovery'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

const STATUS_STYLES = {
  open: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  accepted: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  dismissed: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
  escalated: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
}

export default function DriftReviewPage() {
  const queryClient = useQueryClient()
  const [statusFilter, setStatusFilter] = useState('open')
  const [message, setMessage] = useState(null)

  function showMsg(text, type = 'success') {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 3000)
  }

  const driftQuery = useQuery({
    queryKey: ['drift', 'items', statusFilter],
    queryFn: () => listDriftItems(statusFilter).then(res => res.data || []),
  })
  const items = driftQuery.data ?? []
  const loading = driftQuery.isLoading
  const error = driftQuery.isError ? 'Failed to load drift items' : null

  const resolveMutation = useMutation({
    mutationFn: ({ id, action, note }) => {
      if (action === 'accept') return acceptDrift(id)
      if (action === 'dismiss') return dismissDrift(id)
      return escalateDrift(id, note)
    },
    onSuccess: (_res, { action }) => {
      showMsg(`Drift item ${action === 'accept' ? 'accepted' : action === 'dismiss' ? 'dismissed' : 'escalated'}`)
      queryClient.invalidateQueries({ queryKey: ['drift', 'items'] })
    },
    onError: () => showMsg('Failed to resolve drift item', 'error'),
  })
  const resolving = resolveMutation.isPending ? resolveMutation.variables?.id : null

  function handleAccept(id) {
    resolveMutation.mutate({ id, action: 'accept' })
  }

  function handleDismiss(id) {
    resolveMutation.mutate({ id, action: 'dismiss' })
  }

  function handleEscalate(id) {
    const note = window.prompt('Escalation note (optional):') ?? ''
    resolveMutation.mutate({ id, action: 'escalate', note })
  }

  const openCount = items.filter(i => i.status === 'open').length

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          Drift Review
          {statusFilter === 'open' && openCount > 0 && (
            <span className="ml-2 px-2 py-0.5 text-sm rounded-full bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
              {openCount}
            </span>
          )}
        </h1>
        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-600 dark:text-gray-400">Status:</label>
          <select
            value={statusFilter}
            onChange={e => setStatusFilter(e.target.value)}
            className="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            {/* No "All" option: the backend's status query param always defaults to
                "open" when empty (fiber's Query(key, "open") treats "" as absent),
                so there's no way to request every status through this endpoint. */}
            <option value="open">Open</option>
            <option value="accepted">Accepted</option>
            <option value="dismissed">Dismissed</option>
            <option value="escalated">Escalated</option>
          </select>
        </div>
      </div>

      <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
        Differences between what the scanner last observed and Padduck's authoritative records.
      </p>

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
      ) : items.length === 0 ? (
        <p className="text-sm text-gray-500 dark:text-gray-400">No drift items found.</p>
      ) : (
        <div className="space-y-4">
          {items.map(item => (
            <div key={item.id} className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 overflow-hidden">
              <div className="flex items-center justify-between px-4 py-3 bg-gray-50 dark:bg-gray-800">
                <div className="text-sm text-gray-700 dark:text-gray-200">
                  <span className="font-semibold">{item.resourceType}</span>
                  <span className="text-gray-500 dark:text-gray-400"> #{item.resourceId}</span>
                </div>
                <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_STYLES[item.status] || STATUS_STYLES.dismissed}`}>
                  {item.status}
                </span>
              </div>

              <table className="min-w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-100 dark:border-gray-800">
                    <th className="px-4 py-2 text-left font-semibold text-gray-600 dark:text-gray-300">Field</th>
                    <th className="px-4 py-2 text-left font-semibold text-gray-600 dark:text-gray-300">Authoritative</th>
                    <th className="px-4 py-2 text-left font-semibold text-gray-600 dark:text-gray-300">Observed</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                  {item.fieldDiffs?.map((diff, i) => (
                    <tr key={i}>
                      <td className="px-4 py-2 font-mono text-gray-800 dark:text-gray-200">{diff.field}</td>
                      <td className="px-4 py-2 text-gray-600 dark:text-gray-400">
                        {diff.authoritative || <span className="italic text-gray-400">(none)</span>}
                      </td>
                      <td className="px-4 py-2 text-gray-900 dark:text-gray-100">{diff.observed}</td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {item.status === 'open' && (
                <div className="flex gap-2 px-4 py-3 border-t border-gray-100 dark:border-gray-800">
                  <button
                    onClick={() => handleAccept(item.id)}
                    disabled={resolving === item.id}
                    className="px-3 py-1 text-xs rounded bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
                  >
                    Accept
                  </button>
                  <button
                    onClick={() => handleDismiss(item.id)}
                    disabled={resolving === item.id}
                    className="px-3 py-1 text-xs rounded bg-gray-500 text-white hover:bg-gray-600 disabled:opacity-50"
                  >
                    Dismiss
                  </button>
                  <button
                    onClick={() => handleEscalate(item.id)}
                    disabled={resolving === item.id}
                    className="px-3 py-1 text-xs rounded bg-red-600 text-white hover:bg-red-700 disabled:opacity-50"
                  >
                    Escalate
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
