import { useState, useEffect } from 'react'
import * as client from '../../api/auth'

export default function SecurityTab() {
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

  // Change password
  const [showChangePassword, setShowChangePassword] = useState(false)
  const [changePwForm, setChangePwForm] = useState({ current: '', next: '', confirm: '' })
  const [changePwError, setChangePwError] = useState('')
  const [changePwSuccess, setChangePwSuccess] = useState(false)
  const [changePwLoading, setChangePwLoading] = useState(false)

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

  const handleChangePassword = async (e) => {
    e.preventDefault()
    setChangePwError('')
    setChangePwSuccess(false)
    if (changePwForm.next !== changePwForm.confirm) {
      setChangePwError('New passwords do not match')
      return
    }
    if (changePwForm.next.length < 8) {
      setChangePwError('New password must be at least 8 characters')
      return
    }
    setChangePwLoading(true)
    try {
      await client.changePassword(changePwForm.current, changePwForm.next)
      setChangePwSuccess(true)
      setChangePwForm({ current: '', next: '', confirm: '' })
      setShowChangePassword(false)
    } catch (err) {
      setChangePwError(err.response?.data?.error || 'Failed to change password')
    } finally {
      setChangePwLoading(false)
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
      setBackupCodes(res.data.backupCodes)
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
      setStatus({ totpEnabled: false, backupCodesLeft: 0 })
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
      setRegenResult(res.data.backupCodes)
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
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Password</h2>
        <p className="text-sm text-gray-600 mb-4">Change your account password.</p>
        {changePwSuccess && (
          <div className="mb-3 p-3 bg-green-50 border border-green-200 rounded text-sm text-green-800">
            Password changed successfully.
          </div>
        )}
        {!showChangePassword ? (
          <button
            type="button"
            onClick={() => { setShowChangePassword(true); setChangePwError(''); setChangePwSuccess(false); }}
            className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition"
          >
            Change Password
          </button>
        ) : (
          <form onSubmit={handleChangePassword} className="space-y-3">
            {changePwError && <p className="text-sm text-red-600">{changePwError}</p>}
            <div>
              <label className="block text-sm text-gray-700 mb-1">Current password</label>
              <input
                type="password"
                autoComplete="current-password"
                value={changePwForm.current}
                onChange={e => setChangePwForm(f => ({ ...f, current: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            <div>
              <label className="block text-sm text-gray-700 mb-1">New password</label>
              <input
                type="password"
                autoComplete="new-password"
                value={changePwForm.next}
                onChange={e => setChangePwForm(f => ({ ...f, next: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            <div>
              <label className="block text-sm text-gray-700 mb-1">Confirm new password</label>
              <input
                type="password"
                autoComplete="new-password"
                value={changePwForm.confirm}
                onChange={e => setChangePwForm(f => ({ ...f, confirm: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={changePwLoading}
                className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 transition"
              >
                {changePwLoading ? 'Saving…' : 'Update Password'}
              </button>
              <button
                type="button"
                onClick={() => { setShowChangePassword(false); setChangePwForm({ current: '', next: '', confirm: '' }); setChangePwError(''); }}
                className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition"
              >
                Cancel
              </button>
            </div>
          </form>
        )}
      </div>

      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Two-Factor Authentication</h2>
        <p className="text-sm text-gray-600 mb-4">
          Protect your account with a TOTP authenticator app (e.g. Google Authenticator, Authy).
        </p>

        {status?.totpEnabled ? (
          <div className="space-y-4">
            <div className="flex items-center gap-3 p-3 bg-green-50 border border-green-200 rounded">
              <span className="text-green-600 text-lg">✓</span>
              <div>
                <p className="text-sm font-medium text-green-800">MFA is enabled</p>
                <p className="text-xs text-green-700">{status.backupCodesLeft} backup code{status.backupCodesLeft !== 1 ? 's' : ''} remaining</p>
              </div>
            </div>

            {status.backupCodesLeft <= 2 && (
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
                <img src={setupData.qrCode} alt="TOTP QR code" className="w-48 h-48 border border-gray-200 rounded bg-white" />
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
