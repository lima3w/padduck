import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import * as client from '../api/client'

function ProfileTab({ user }) {
  return (
    <div className="max-w-lg">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Profile</h2>
      <dl className="space-y-3">
        <div>
          <dt className="text-sm font-medium text-gray-500">Username</dt>
          <dd className="mt-1 text-sm text-gray-900">{user?.username}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-gray-500">Email</dt>
          <dd className="mt-1 text-sm text-gray-900">{user?.email}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-gray-500">Role</dt>
          <dd className="mt-1 text-sm text-gray-900 capitalize">{user?.role}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-gray-500">Account State</dt>
          <dd className="mt-1 text-sm text-gray-900 capitalize">{user?.state?.replace(/_/g, ' ')}</dd>
        </div>
      </dl>
    </div>
  )
}

function SecurityTab() {
  const [status, setStatus] = useState(null)
  const [loadingStatus, setLoadingStatus] = useState(true)

  // Setup flow
  const [setupData, setSetupData] = useState(null)
  const [confirmCode, setConfirmCode] = useState('')
  const [backupCodes, setBackupCodes] = useState(null)
  const [setupError, setSetupError] = useState('')
  const [setupLoading, setSetupLoading] = useState(false)

  // Disable flow
  const [disableCode, setDisableCode] = useState('')
  const [disableError, setDisableError] = useState('')
  const [disableLoading, setDisableLoading] = useState(false)
  const [showDisable, setShowDisable] = useState(false)

  // Backup code regen
  const [regenCode, setRegenCode] = useState('')
  const [regenError, setRegenError] = useState('')
  const [regenLoading, setRegenLoading] = useState(false)
  const [regenResult, setRegenResult] = useState(null)
  const [showRegen, setShowRegen] = useState(false)

  useEffect(() => {
    loadStatus()
  }, [])

  const loadStatus = async () => {
    setLoadingStatus(true)
    try {
      const res = await client.getMFAStatus()
      setStatus(res.data)
    } catch {
      setStatus(null)
    } finally {
      setLoadingStatus(false)
    }
  }

  const handleStartSetup = async () => {
    setSetupLoading(true)
    setSetupError('')
    try {
      const res = await client.setupTOTP()
      setSetupData(res.data)
      setBackupCodes(null)
      setConfirmCode('')
    } catch (err) {
      setSetupError(err.response?.data?.error || 'Failed to start MFA setup')
    } finally {
      setSetupLoading(false)
    }
  }

  const handleConfirmTOTP = async (e) => {
    e.preventDefault()
    setSetupLoading(true)
    setSetupError('')
    try {
      const res = await client.confirmTOTP(confirmCode)
      setBackupCodes(res.data.backup_codes)
      setSetupData(null)
      await loadStatus()
    } catch (err) {
      setSetupError(err.response?.data?.error || 'Invalid code')
    } finally {
      setSetupLoading(false)
    }
  }

  const handleDisable = async (e) => {
    e.preventDefault()
    setDisableLoading(true)
    setDisableError('')
    try {
      await client.disableTOTP(disableCode)
      setShowDisable(false)
      setDisableCode('')
      setStatus({ totp_enabled: false, backup_codes_left: 0 })
    } catch (err) {
      setDisableError(err.response?.data?.error || 'Failed to disable MFA')
    } finally {
      setDisableLoading(false)
    }
  }

  const handleRegen = async (e) => {
    e.preventDefault()
    setRegenLoading(true)
    setRegenError('')
    try {
      const res = await client.regenerateBackupCodes(regenCode)
      setRegenResult(res.data.backup_codes)
      setRegenCode('')
      setShowRegen(false)
      await loadStatus()
    } catch (err) {
      setRegenError(err.response?.data?.error || 'Failed to regenerate codes')
    } finally {
      setRegenLoading(false)
    }
  }

  if (loadingStatus) {
    return <p className="text-sm text-gray-500">Loading…</p>
  }

  return (
    <div className="max-w-lg space-y-8">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Two-Factor Authentication</h2>
        <p className="text-sm text-gray-600 mb-4">
          Protect your account with a TOTP authenticator app (e.g. Google Authenticator, Authy).
        </p>

        {status?.totp_enabled ? (
          <div className="space-y-4">
            <div className="flex items-center gap-3 p-3 bg-green-50 border border-green-200 rounded">
              <span className="text-green-600 text-lg">✓</span>
              <div>
                <p className="text-sm font-medium text-green-800">MFA is enabled</p>
                <p className="text-xs text-green-700">{status.backup_codes_left} backup code{status.backup_codes_left !== 1 ? 's' : ''} remaining</p>
              </div>
            </div>

            {status.backup_codes_left <= 2 && (
              <div className="p-3 bg-yellow-50 border border-yellow-200 rounded text-sm text-yellow-800">
                You&apos;re running low on backup codes. Consider regenerating them.
              </div>
            )}

            {regenResult && (
              <div className="p-4 bg-gray-50 border border-gray-200 rounded">
                <p className="text-sm font-medium text-gray-800 mb-2">New backup codes — save these now:</p>
                <ul className="space-y-1">
                  {regenResult.map((code) => (
                    <li key={code} className="font-mono text-sm text-gray-700">{code}</li>
                  ))}
                </ul>
              </div>
            )}

            {!showRegen && !showDisable && (
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setShowRegen(true)}
                  className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition"
                >
                  Regenerate Backup Codes
                </button>
                <button
                  type="button"
                  onClick={() => setShowDisable(true)}
                  className="px-4 py-2 text-sm border border-red-300 text-red-700 rounded hover:bg-red-50 transition"
                >
                  Disable MFA
                </button>
              </div>
            )}

            {showRegen && (
              <form onSubmit={handleRegen} className="space-y-3">
                <p className="text-sm text-gray-700">Enter your current TOTP code to regenerate backup codes:</p>
                {regenError && <p className="text-sm text-red-600">{regenError}</p>}
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={regenCode}
                    onChange={(e) => setRegenCode(e.target.value.replace(/\s/g, ''))}
                    placeholder="000000"
                    maxLength={12}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded font-mono text-center"
                  />
                  <button
                    type="submit"
                    disabled={regenLoading || !regenCode}
                    className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 transition"
                  >
                    {regenLoading ? 'Regenerating…' : 'Confirm'}
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowRegen(false); setRegenCode(''); setRegenError(''); }}
                    className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            )}

            {showDisable && (
              <form onSubmit={handleDisable} className="space-y-3">
                <p className="text-sm text-gray-700">Enter your TOTP code or a backup code to disable MFA:</p>
                {disableError && <p className="text-sm text-red-600">{disableError}</p>}
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={disableCode}
                    onChange={(e) => setDisableCode(e.target.value.replace(/\s/g, ''))}
                    placeholder="000000"
                    maxLength={12}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded font-mono text-center"
                  />
                  <button
                    type="submit"
                    disabled={disableLoading || !disableCode}
                    className="px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700 disabled:opacity-50 transition"
                  >
                    {disableLoading ? 'Disabling…' : 'Disable MFA'}
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowDisable(false); setDisableCode(''); setDisableError(''); }}
                    className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            {backupCodes ? (
              <div className="space-y-4">
                <div className="p-4 bg-green-50 border border-green-200 rounded">
                  <p className="text-sm font-medium text-green-800 mb-3">MFA enabled! Save your backup codes:</p>
                  <ul className="space-y-1">
                    {backupCodes.map((code) => (
                      <li key={code} className="font-mono text-sm text-gray-700">{code}</li>
                    ))}
                  </ul>
                  <p className="text-xs text-gray-600 mt-3">These codes will not be shown again.</p>
                </div>
              </div>
            ) : setupData ? (
              <div className="space-y-4">
                <p className="text-sm text-gray-700">Scan this QR code with your authenticator app:</p>
                <img src={setupData.qr_code} alt="TOTP QR code" className="w-48 h-48 border border-gray-200 rounded" />
                <details className="text-sm">
                  <summary className="text-gray-500 cursor-pointer">Can&apos;t scan? Enter the secret manually</summary>
                  <code className="block mt-2 p-2 bg-gray-100 rounded font-mono text-xs break-all">{setupData.secret}</code>
                </details>
                <form onSubmit={handleConfirmTOTP} className="space-y-3">
                  <p className="text-sm text-gray-700">Then enter the 6-digit code to confirm:</p>
                  {setupError && <p className="text-sm text-red-600">{setupError}</p>}
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={confirmCode}
                      onChange={(e) => setConfirmCode(e.target.value.replace(/\s/g, ''))}
                      placeholder="000000"
                      maxLength={6}
                      className="flex-1 px-3 py-2 border border-gray-300 rounded font-mono text-center text-xl tracking-widest"
                      autoFocus
                    />
                    <button
                      type="submit"
                      disabled={setupLoading || confirmCode.length < 6}
                      className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 transition"
                    >
                      {setupLoading ? 'Verifying…' : 'Enable MFA'}
                    </button>
                  </div>
                </form>
              </div>
            ) : (
              <div className="space-y-3">
                <div className="p-3 bg-gray-50 border border-gray-200 rounded text-sm text-gray-600">
                  MFA is not enabled on your account.
                </div>
                {setupError && <p className="text-sm text-red-600">{setupError}</p>}
                <button
                  type="button"
                  onClick={handleStartSetup}
                  disabled={setupLoading}
                  className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 transition"
                >
                  {setupLoading ? 'Setting up…' : 'Set Up MFA'}
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function TokensTab() {
  const [tokens, setTokens] = useState([])
  const [loading, setLoading] = useState(true)
  const [tokenName, setTokenName] = useState('')
  const [newToken, setNewToken] = useState(null)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    loadTokens()
  }, [])

  const loadTokens = async () => {
    setLoading(true)
    try {
      const res = await client.listMyTokens()
      setTokens(res.data)
    } catch {
      setTokens([])
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (e) => {
    e.preventDefault()
    if (!tokenName.trim()) return
    setCreating(true)
    setError('')
    try {
      const res = await client.generateTokenForMe(tokenName.trim())
      setNewToken(res.data.token)
      setTokenName('')
      await loadTokens()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create token')
    } finally {
      setCreating(false)
    }
  }

  const handleRevoke = async (id) => {
    try {
      await client.revokeToken(id)
      setTokens((prev) => prev.filter((t) => t.id !== id))
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to revoke token')
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">API Tokens</h2>
        <p className="text-sm text-gray-600 mb-4">
          Tokens authenticate API requests. Treat them like passwords.
        </p>

        {newToken && (
          <div className="mb-4 p-4 bg-green-50 border border-green-200 rounded">
            <p className="text-sm font-medium text-green-800 mb-2">Token created — copy it now, it won&apos;t be shown again:</p>
            <code className="block p-2 bg-white border border-green-200 rounded font-mono text-xs break-all text-gray-700">{newToken}</code>
            <button
              type="button"
              onClick={() => setNewToken(null)}
              className="mt-2 text-xs text-green-700 hover:underline"
            >
              Dismiss
            </button>
          </div>
        )}

        {error && <p className="mb-4 text-sm text-red-600">{error}</p>}

        <form onSubmit={handleCreate} className="flex gap-2 mb-6">
          <input
            type="text"
            value={tokenName}
            onChange={(e) => setTokenName(e.target.value)}
            placeholder="Token name (e.g. CLI, Terraform)"
            className="flex-1 px-3 py-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <button
            type="submit"
            disabled={creating || !tokenName.trim()}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {creating ? 'Creating…' : 'Create Token'}
          </button>
        </form>

        {loading ? (
          <p className="text-sm text-gray-500">Loading…</p>
        ) : tokens.length === 0 ? (
          <p className="text-sm text-gray-500">No tokens yet.</p>
        ) : (
          <div className="divide-y divide-gray-200 border border-gray-200 rounded">
            {tokens.map((t) => (
              <div key={t.id} className="flex items-center justify-between px-4 py-3">
                <div>
                  <p className="text-sm font-medium text-gray-900">{t.name}</p>
                  <p className="text-xs text-gray-500">
                    Created {new Date(t.created_at).toLocaleDateString()}
                    {t.last_used_at ? ` · Last used ${new Date(t.last_used_at).toLocaleDateString()}` : ' · Never used'}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => handleRevoke(t.id)}
                  className="text-sm text-red-600 hover:text-red-800 transition"
                >
                  Revoke
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function LoginHistoryTab() {
  const [history, setHistory] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    loadHistory()
  }, [])

  const loadHistory = async () => {
    setLoading(true)
    setError('')
    try {
      const res = await client.getLoginHistory()
      setHistory(res.data)
    } catch {
      setError('Failed to load login history.')
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (iso) => {
    const d = new Date(iso)
    return d.toLocaleString()
  }

  const knownIPs = history.filter((a) => a.success).map((a) => a.ipAddress).filter(Boolean)
  const knownIPSet = new Set(knownIPs)

  return (
    <div className="max-w-2xl space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Login History</h2>
        <p className="text-sm text-gray-600 mb-4">Recent login attempts on your account (last 20).</p>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      {loading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : history.length === 0 ? (
        <p className="text-sm text-gray-500">No login history yet.</p>
      ) : (
        <div className="divide-y divide-gray-200 border border-gray-200 rounded">
          {history.map((attempt) => {
            const isNew = attempt.ipAddress && !knownIPSet.has(attempt.ipAddress) && !attempt.success
            return (
              <div key={attempt.id} className={`px-4 py-3 flex items-start gap-3 ${!attempt.success ? 'bg-red-50' : ''}`}>
                <div className={`mt-0.5 w-2 h-2 rounded-full flex-shrink-0 ${attempt.success ? 'bg-green-500' : 'bg-red-500'}`} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className={`text-sm font-medium ${attempt.success ? 'text-green-800' : 'text-red-800'}`}>
                      {attempt.success ? 'Successful login' : 'Failed login attempt'}
                    </span>
                    {isNew && (
                      <span className="text-xs px-1.5 py-0.5 bg-yellow-100 text-yellow-800 border border-yellow-200 rounded">New IP</span>
                    )}
                  </div>
                  <p className="text-xs text-gray-500 mt-0.5">
                    {formatDate(attempt.createdAt)}
                    {attempt.ipAddress ? ` · ${attempt.ipAddress}` : ''}
                  </p>
                  {attempt.failureReason && (
                    <p className="text-xs text-red-600 mt-0.5">{attempt.failureReason}</p>
                  )}
                  {attempt.userAgent && (
                    <p className="text-xs text-gray-400 mt-0.5 truncate">{attempt.userAgent}</p>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

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

function NotificationsTab() {
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

      <div className="divide-y divide-gray-100 border border-gray-200 rounded">
        {Object.entries(NOTIF_LABELS).map(([key, label]) => (
          <label key={key} className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-gray-50">
            <span className="text-sm text-gray-800">{label}</span>
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

function SessionsTab() {
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

function PrivacyTab({ user }) {
  const [policyVersion, setPolicyVersion] = useState(null)
  const [acceptedVersion, setAcceptedVersion] = useState(user?.privacyAcceptedVersion || user?.privacy_accepted_version || null)
  const [loading, setLoading] = useState(true)
  const [accepting, setAccepting] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    client.getPrivacyPolicyVersion()
      .then((res) => {
        if (!cancelled) setPolicyVersion(res.data?.version || '1.0')
      })
      .catch(() => {
        if (!cancelled) setError('Failed to load the current privacy policy version.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [])

  useEffect(() => {
    setAcceptedVersion(user?.privacyAcceptedVersion || user?.privacy_accepted_version || null)
  }, [user])

  const handleAccept = async () => {
    setAccepting(true)
    setError('')
    setSaved(false)
    try {
      await client.acceptPrivacyPolicy()
      const nextVersion = policyVersion || '1.0'
      const cached = JSON.parse(localStorage.getItem('current_user') || '{}')
      localStorage.setItem('current_user', JSON.stringify({
        ...cached,
        privacyAcceptedVersion: nextVersion,
        privacy_accepted_version: undefined,
      }))
      setAcceptedVersion(nextVersion)
      setSaved(true)
    } catch {
      setError('Failed to record privacy consent.')
    } finally {
      setAccepting(false)
    }
  }

  if (loading) return <p className="text-sm text-gray-500">Loading...</p>

  const currentAccepted = acceptedVersion && acceptedVersion === policyVersion

  return (
    <div className="max-w-lg space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Privacy</h2>
        <p className="text-sm text-gray-600">
          Review the privacy policy version recorded for your account.
        </p>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      <dl className="divide-y divide-gray-200 border border-gray-200 rounded">
        <div className="flex items-center justify-between gap-4 px-4 py-3">
          <dt className="text-sm font-medium text-gray-600">Current policy version</dt>
          <dd className="text-sm text-gray-900">{policyVersion || 'Unknown'}</dd>
        </div>
        <div className="flex items-center justify-between gap-4 px-4 py-3">
          <dt className="text-sm font-medium text-gray-600">Accepted version</dt>
          <dd className="text-sm text-gray-900">{acceptedVersion || 'Not accepted'}</dd>
        </div>
        <div className="flex items-center justify-between gap-4 px-4 py-3">
          <dt className="text-sm font-medium text-gray-600">Status</dt>
          <dd className={currentAccepted ? 'text-sm font-medium text-green-700' : 'text-sm font-medium text-yellow-700'}>
            {currentAccepted ? 'Current' : 'Action required'}
          </dd>
        </div>
      </dl>

      {!currentAccepted && (
        <button
          type="button"
          onClick={handleAccept}
          disabled={accepting || !policyVersion}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {accepting ? 'Accepting...' : 'Accept current policy'}
        </button>
      )}
      {saved && <p className="text-sm text-green-600">Privacy consent recorded.</p>}
    </div>
  )
}

const TAB_PARAM_MAP = { history: 'login-history', notif: 'notifications' }
const VALID_TABS = new Set(['profile', 'security', 'tokens', 'login-history', 'sessions', 'notifications', 'privacy'])

export default function UserSettingsPage() {
  const { user } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()

  const rawTab = searchParams.get('tab') || 'profile'
  const resolvedTab = TAB_PARAM_MAP[rawTab] || rawTab
  const tab = VALID_TABS.has(resolvedTab) ? resolvedTab : 'profile'

  const setTab = (id) => setSearchParams({ tab: id }, { replace: true })

  const tabs = [
    { id: 'profile', label: 'Profile' },
    { id: 'security', label: 'Security' },
    { id: 'tokens', label: 'API Tokens' },
    { id: 'sessions', label: 'Sessions' },
    { id: 'notifications', label: 'Notifications' },
    { id: 'login-history', label: 'Login History' },
    { id: 'privacy', label: 'Privacy' },
  ]

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Account Settings</h1>

      <div className="flex gap-1 mb-6 border-b border-gray-200 overflow-x-auto" role="tablist" aria-label="Account settings sections">
        {tabs.map((t) => (
          <button
            key={t.id}
            type="button"
            role="tab"
            id={`account-tab-${t.id}`}
            aria-selected={tab === t.id}
            aria-controls={`account-panel-${t.id}`}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition ${
              tab === t.id
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div role="tabpanel" id={`account-panel-${tab}`} aria-labelledby={`account-tab-${tab}`}>
        {tab === 'profile' && <ProfileTab user={user} />}
        {tab === 'security' && <SecurityTab />}
        {tab === 'tokens' && <TokensTab />}
        {tab === 'sessions' && <SessionsTab />}
        {tab === 'notifications' && <NotificationsTab />}
        {tab === 'login-history' && <LoginHistoryTab />}
        {tab === 'privacy' && <PrivacyTab user={user} />}
      </div>
    </div>
  )
}
