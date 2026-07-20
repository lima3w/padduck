import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'
import { Link } from 'react-router-dom'

function formatRelativeTime(isoString, t) {
  const diff = Math.floor((Date.now() - new Date(isoString).getTime()) / 1000)
  if (diff < 60) return t('changeHistory.justNow')
  if (diff < 3600) return t('changeHistory.minutesAgo', { count: Math.floor(diff / 60) })
  if (diff < 86400) return t('changeHistory.hoursAgo', { count: Math.floor(diff / 3600) })
  return t('changeHistory.daysAgo', { count: Math.floor(diff / 86400) })
}

export default function ChangeHistory({ resourceType, resourceId }) {
  const { t } = useTranslation()
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
        <h3 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">{t('changeHistory.title')}</h3>
        <Link to="/admin/audit-log" className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400">{t('changeHistory.viewAll')}</Link>
      </div>
      <div className="space-y-2">
        {entries.map(e => (
          <div key={e.id} className="flex items-start gap-2 text-xs">
            <span className="text-gray-400 dark:text-gray-500 whitespace-nowrap mt-0.5">{formatRelativeTime(e.timestamp, t)}</span>
            <span className="flex-1 text-gray-700 dark:text-gray-300">
              <span className="font-medium">{e.username || t('changeHistory.systemFallback')}</span>
              {' '}{e.action.replace(/_/g, ' ')}
              {e.resourceName ? <span className="text-gray-500"> &mdash; {e.resourceName}</span> : null}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
