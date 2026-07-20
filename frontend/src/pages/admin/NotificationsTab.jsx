import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getNotificationStats } from '../../api/admin'

export default function NotificationsTab() {
  const { t } = useTranslation()
  const [notifStats, setNotifStats] = useState(null)
  const [notifStatsLoading, setNotifStatsLoading] = useState(false)

  const loadNotifStats = async () => {
    setNotifStatsLoading(true)
    try {
      const res = await getNotificationStats()
      setNotifStats(res.data)
    } catch {
      setNotifStats(null)
    } finally {
      setNotifStatsLoading(false)
    }
  }

  useEffect(() => { loadNotifStats() }, [])

  return (
        <div className="space-y-6">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-1">{t('notificationsTab.title')}</h2>
            <p className="text-sm text-gray-600 mb-4">
              {t('notificationsTab.subtitle')}
            </p>
          </div>

          {notifStatsLoading ? (
            <p className="text-sm text-gray-500">{t('common.loading')}</p>
          ) : notifStats === null ? (
            <p className="text-sm text-red-500">{t('notificationsTab.loadFailed')}</p>
          ) : Object.keys(notifStats).length === 0 ? (
            <p className="text-sm text-gray-500">{t('notificationsTab.noneSent')}</p>
          ) : (
            <div className="border border-gray-200 rounded divide-y divide-gray-100">
              {Object.entries(notifStats).map(([key, count]) => (
                <div key={key} className="flex items-center justify-between px-4 py-3">
                  <span className="text-sm text-gray-700">
                    {key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())}
                  </span>
                  <span className="text-sm font-medium text-gray-900">{count.toLocaleString()}</span>
                </div>
              ))}
            </div>
          )}

          <button
            onClick={loadNotifStats}
            disabled={notifStatsLoading}
            className="text-sm text-blue-600 hover:underline disabled:opacity-50"
          >
            {t('dashboard.refresh')}
          </button>
        </div>
  )
}
