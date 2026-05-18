import { useState, useEffect } from 'react'
import { api } from '../api/client'
import { Link } from 'react-router-dom'

function formatRelativeTime(isoString) {
  const diff = Math.floor((Date.now() - new Date(isoString).getTime()) / 1000)
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

export default function ChangeHistory({ resourceType, resourceId }) {
  const [entries, setEntries] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!resourceId) return
    api.get('/admin/audit-logs', { params: { resource_type: resourceType, resource_id: resourceId, limit: 8 } })
      .then(res => setEntries(res.data?.logs ?? []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [resourceType, resourceId])

  if (loading) return null
  if (entries.length === 0) return null

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">Change History</h3>
        <Link to="/admin/audit-log" className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400">View all</Link>
      </div>
      <div className="space-y-2">
        {entries.map(e => (
          <div key={e.id} className="flex items-start gap-2 text-xs">
            <span className="text-gray-400 dark:text-gray-500 whitespace-nowrap mt-0.5">{formatRelativeTime(e.timestamp)}</span>
            <span className="flex-1 text-gray-700 dark:text-gray-300">
              <span className="font-medium">{e.username || 'system'}</span>
              {' '}{e.action.replace(/_/g, ' ')}
              {e.resource_name ? <span className="text-gray-500"> &mdash; {e.resource_name}</span> : null}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
