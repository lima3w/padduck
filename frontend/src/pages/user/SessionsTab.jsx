import { useState, useEffect } from 'react'
import * as client from '../../api/auth'

export default function SessionsTab() {
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
      setError('Failed to load sessions.')
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
      setError('Failed to revoke session.')
    } finally {
      setRevoking(null)
    }
  }

  const handleLogoutAll = async () => {
    if (!confirm('Sign out of all other devices?')) return
    setLogoutAllLoading(true)
    try {
      await client.logoutAllDevices()
      await load()
    } catch {
      setError('Failed to sign out all devices.')
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
          <h2 className="text-lg font-semibold text-gray-900 mb-1">Active Sessions</h2>
          <p className="text-sm text-gray-600">Devices currently signed in to your account.</p>
        </div>
        <button
          onClick={handleLogoutAll}
          disabled={logoutAllLoading || loading}
          className="flex-shrink-0 text-sm text-red-600 border border-red-300 px-3 py-1.5 rounded hover:bg-red-50 disabled:opacity-50 transition"
        >
          {logoutAllLoading ? 'Signing out…' : 'Sign out all devices'}
        </button>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      {loading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : sessions.length === 0 ? (
        <p className="text-sm text-gray-500">No active sessions.</p>
      ) : (
        <div className="divide-y divide-gray-200 border border-gray-200 rounded">
          {sessions.map((s) => (
            <div key={s.id} className="px-4 py-3 flex items-start gap-3">
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {s.deviceName || 'Unknown device'}
                </p>
                <p className="text-xs text-gray-500 mt-0.5">
                  {s.ipAddress || 'Unknown IP'}
                  {' · '}Last active {fmt(s.lastUsedAt)}
                </p>
                <p className="text-xs text-gray-400 mt-0.5">
                  Expires {fmt(s.absoluteExpiresAt)}
                </p>
              </div>
              <button
                onClick={() => handleRevoke(s.id)}
                disabled={revoking === s.id}
                className="flex-shrink-0 text-xs text-red-600 hover:underline disabled:opacity-50"
              >
                {revoking === s.id ? 'Revoking…' : 'Revoke'}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
