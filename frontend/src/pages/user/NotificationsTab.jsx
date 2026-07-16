import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import * as client from '../../api/auth'

const NOTIF_KEYS = [
  'loginSuccess',
  'loginFailed',
  'accountLocked',
  'passwordChanged',
  'mfaChanges',
  'apiTokenChanges',
  'roleChanges',
  'sessionRevoked',
]

export default function NotificationsTab() {
  const { t } = useTranslation()
  const [prefs, setPrefs] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    client.getNotificationPreferences()
      .then((res) => {
        const d = res.data
        setPrefs({
          loginSuccess:    d.login_success,
          loginFailed:     d.login_failed,
          accountLocked:   d.account_locked,
          passwordChanged: d.password_changed,
          mfaChanges:      d.mfa_changes,
          apiTokenChanges: d.api_token_changes,
          roleChanges:     d.role_changes,
          sessionRevoked:  d.session_revoked,
        })
      })
      .catch(() => setError(t('userTabs.notifications.loadError')))
      .finally(() => setLoading(false))
  }, [])

  const toggle = (key) => {
    setPrefs((prev) => ({ ...prev, [key]: !prev[key] }))
    setSaved(false)
  }

  const handleSave = async () => {
    setSaving(true)
    setError('')
    try {
      const updates = {}
      NOTIF_KEYS.forEach((k) => {
        const snakeKey = k.replace(/([A-Z])/g, '_$1').toLowerCase()
        updates[snakeKey] = prefs[k]
      })
      const res = await client.updateNotificationPreferences(updates)
      setPrefs(res.data)
      setSaved(true)
    } catch {
      setError(t('userTabs.notifications.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const labelFor = (key) => {
    if (key === 'loginSuccess') return t('userTabs.loginHistory.successful')
    if (key === 'loginFailed') return t('userTabs.loginHistory.failedAttempt')
    return t(`userTabs.notifications.labels.${key}`)
  }

  if (loading) return <p className="text-sm text-gray-500">{t('common.loading')}</p>

  return (
    <div className="max-w-lg space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">{t('userTabs.notifications.title')}</h2>
        <p className="text-sm text-gray-600">{t('userTabs.notifications.subtitle')}</p>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      <div className="divide-y divide-gray-100 dark:divide-gray-700 border border-gray-200 dark:border-gray-700 rounded">
        {NOTIF_KEYS.map((key) => (
          <label key={key} className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
            <span className="text-sm text-gray-800 dark:text-gray-100">{labelFor(key)}</span>
            <input
              type="checkbox"
              checked={prefs?.[key] ?? true}
              onChange={() => toggle(key)}
              className="h-4 w-4 text-blue-600 border-gray-300 rounded"
            />
          </label>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <button
          onClick={handleSave}
          disabled={saving}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition"
        >
          {saving ? t('common.saving') : t('userTabs.notifications.savePreferences')}
        </button>
        {saved && <span className="text-sm text-green-600">{t('common.saved')}</span>}
      </div>
    </div>
  )
}
