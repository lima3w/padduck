import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import * as client from '../../api/auth'

export default function SessionsTab() {
  const { t } = useTranslation()
  const [sessions, setSessions] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revoking, setRevoking] = useState(null)
  const [logoutAllLoading, setLogoutAllLoading] = useState(false)

  useEffect(() => {
    load()
  }, [])

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      const res = await client.listMySessions()
      setSessions(res.data)
    } catch {
      setError(t('userTabs.sessions.loadError'))
    } finally {
      setLoading(false)
    }
  }

  const handleRevoke = async (id) => {
    setRevoking(id)
    try {
      await client.revokeMySession(id)
      setSessions((prev) => prev.filter((s) => s.id !== id))
    } catch {
      setError(t('userTabs.sessions.revokeFailed'))
    } finally {
      setRevoking(null)
    }
  }

  const handleLogoutAll = async () => {
    if (!confirm(t('userTabs.sessions.confirmSignOutAll'))) return
    setLogoutAllLoading(true)
    try {
      await client.logoutAllDevices()
      await load()
    } catch {
      setError(t('userTabs.sessions.signOutAllFailed'))
    } finally {
      setLogoutAllLoading(false)
    }
  }

  const fmt = (iso) => {
    try { return new Date(iso).toLocaleString() } catch { return iso }
  }

  return (
    <div className="max-w-2xl space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-1">{t('userMenu.activeSessions')}</h2>
          <p className="text-sm text-gray-600">{t('userTabs.sessions.subtitle')}</p>
        </div>
        <button
          onClick={handleLogoutAll}
          disabled={logoutAllLoading || loading}
          className="flex-shrink-0 text-sm text-red-600 border border-red-300 px-3 py-1.5 rounded hover:bg-red-50 disabled:opacity-50 transition"
        >
          {logoutAllLoading ? t('userTabs.sessions.signingOut') : t('userTabs.sessions.signOutAllDevices')}
        </button>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      {loading ? (
        <p className="text-sm text-gray-500">{t('common.loading')}</p>
      ) : sessions.length === 0 ? (
        <p className="text-sm text-gray-500">{t('userTabs.sessions.empty')}</p>
      ) : (
        <div className="divide-y divide-gray-200 border border-gray-200 rounded">
          {sessions.map((s) => (
            <div key={s.id} className="px-4 py-3 flex items-start gap-3">
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {s.deviceName || t('userTabs.sessions.unknownDevice')}
                </p>
                <p className="text-xs text-gray-500 mt-0.5">
                  {s.ipAddress || t('userTabs.sessions.unknownIp')}
                  {' · '}{t('userTabs.sessions.lastActive', { date: fmt(s.lastUsedAt) })}
                </p>
                <p className="text-xs text-gray-400 mt-0.5">
                  {t('userTabs.sessions.expires', { date: fmt(s.absoluteExpiresAt) })}
                </p>
              </div>
              <button
                onClick={() => handleRevoke(s.id)}
                disabled={revoking === s.id}
                className="flex-shrink-0 text-xs text-red-600 hover:underline disabled:opacity-50"
              >
                {revoking === s.id ? t('userTabs.sessions.revoking') : t('common.revoke')}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
