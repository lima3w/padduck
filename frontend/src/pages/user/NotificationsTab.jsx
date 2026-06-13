import { useState, useEffect } from 'react'
import * as client from '../../api/auth'

const NOTIF_LABELS = {
  loginSuccess: 'Successful login',
  loginFailed: 'Failed login attempt',
  accountLocked: 'Account locked',
  passwordChanged: 'Password changed',
  mfaChanges: 'MFA changes',
  apiTokenChanges: 'API token changes',
  roleChanges: 'Role changes',
  sessionRevoked: 'Session revoked',
}

export default function NotificationsTab() {
  const [prefs, setPrefs] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    client.getNotificationPreferences()
      .then((res) => setPrefs(res.data))
      .catch(() => setError('Failed to load preferences.'))
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
      Object.keys(NOTIF_LABELS).forEach((k) => {
        const snakeKey = k.replace(/([A-Z])/g, '_$1').toLowerCase()
        updates[snakeKey] = prefs[k]
      })
      const res = await client.updateNotificationPreferences(updates)
      setPrefs(res.data)
      setSaved(true)
    } catch {
      setError('Failed to save preferences.')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <p className="text-sm text-gray-500">Loading…</p>

  return (
    <div className="max-w-lg space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Notification Preferences</h2>
        <p className="text-sm text-gray-600">Choose which security events send you an email.</p>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      <div className="divide-y divide-gray-100 dark:divide-gray-700 border border-gray-200 dark:border-gray-700 rounded">
        {Object.entries(NOTIF_LABELS).map(([key, label]) => (
          <label key={key} className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
            <span className="text-sm text-gray-800 dark:text-gray-100">{label}</span>
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
          {saving ? 'Saving…' : 'Save preferences'}
        </button>
        {saved && <span className="text-sm text-green-600">Saved.</span>}
      </div>
    </div>
  )
}
