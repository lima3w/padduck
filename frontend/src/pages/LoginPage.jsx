import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import * as client from '../api/client'

export default function LoginPage() {
  const navigate = useNavigate()
  const { login } = useAuth()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // MFA step state
  const [mfaChallenge, setMfaChallenge] = useState(null)
  const [mfaCode, setMfaCode] = useState('')

  // Email verification resend state
  const [showResend, setShowResend] = useState(false)
  const [resendEmail, setResendEmail] = useState('')
  const [resendStatus, setResendStatus] = useState('')

  // External auth providers
  const [providers, setProviders] = useState({ ldap: false, oauth2: false, saml: false })
  const [useLdap, setUseLdap] = useState(false)

  useEffect(() => {
    client.getAuthProviders()
      .then((res) => setProviders(res.data || {}))
      .catch(() => {}) // 404 or network error — silently ignore
  }, [])

  const handlePasswordLogin = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setShowResend(false)
    setResendStatus('')

    try {
      const response = useLdap
        ? await client.ldapLogin(username, password)
        : await client.login(username, password)
      const data = response.data

      if (data.mfa_required) {
        setMfaChallenge(data.mfa_challenge)
        setLoading(false)
        return
      }

      login(data.user)
      navigate('/')
    } catch (err) {
      const msg = err.response?.data?.error || 'Login failed'
      setError(msg)
      if (msg === 'email address not verified') {
        setShowResend(true)
      }
    } finally {
      setLoading(false)
    }
  }

  const handleMFASubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const response = await client.verifyMFA(mfaChallenge, mfaCode)
      login(response.data.user)
      navigate('/')
    } catch (err) {
      const msg = err.response?.data?.error || 'MFA verification failed'
      setError(msg)
      if (err.response?.status === 401 && msg.includes('expired')) {
        setMfaChallenge(null)
        setMfaCode('')
        setError('MFA session expired. Please sign in again.')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleResendVerification = async () => {
    if (!resendEmail) return
    setResendStatus('sending')
    try {
      await client.resendVerification(resendEmail)
      setResendStatus('sent')
    } catch {
      setResendStatus('error')
    }
  }

  if (mfaChallenge) {
    return (
      <div className="min-h-screen bg-[#07162b] flex items-center justify-center p-4">
        <div className="bg-white dark:bg-[#0a1f3a] rounded-lg shadow-xl p-8 max-w-md w-full border border-transparent dark:border-[#25364a]">
          <div className="flex justify-center mb-6">
            <img src="/logo.png" alt="Padduck" className="w-48 h-auto" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-[#f4f7fa] mb-1 text-center">Two-Factor Authentication</h1>
          <p className="text-gray-600 dark:text-[#a8b8cb] mb-6 text-center">Enter the 6-digit code from your authenticator app, or a backup code.</p>

          {error && (
            <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 rounded text-red-700 dark:text-red-400 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleMFASubmit}>
            <div className="mb-6">
              <label htmlFor="mfaCode" className="block text-sm font-medium text-gray-700 dark:text-[#a8b8cb] mb-2">
                Authentication Code
              </label>
              <input
                type="text"
                id="mfaCode"
                value={mfaCode}
                onChange={(e) => setMfaCode(e.target.value.replace(/\s/g, ''))}
                className="w-full px-4 py-2 border border-gray-300 dark:border-[#25364a] rounded-lg focus:ring-2 focus:ring-[#f5b800] focus:border-transparent text-center text-xl tracking-widest font-mono"
                placeholder="000000"
                maxLength={12}
                autoFocus
                autoComplete="one-time-code"
              />
            </div>

            <button
              type="submit"
              disabled={loading || !mfaCode}
              className="w-full bg-[#f5b800] text-[#07162b] py-2 px-4 rounded-lg hover:bg-[#ffcf33] disabled:opacity-50 transition font-semibold"
            >
              {loading ? 'Verifying…' : 'Verify'}
            </button>
          </form>

          <button
            type="button"
            onClick={() => { setMfaChallenge(null); setMfaCode(''); setError(''); }}
            className="w-full mt-3 text-sm text-[#a8b8cb] hover:text-[#f4f7fa] transition"
          >
            Back to sign in
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[#07162b] flex items-center justify-center p-4">
      <div className="bg-white dark:bg-[#0a1f3a] rounded-lg shadow-xl p-8 max-w-md w-full border border-transparent dark:border-[#25364a]">
        <div className="flex flex-col items-center mb-6">
          <img src="/logo.png" alt="Padduck" className="w-48 h-auto mb-1" />
          <p className="text-gray-600 dark:text-[#a8b8cb] text-sm">Sign in to continue</p>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 rounded text-red-700 dark:text-red-400 text-sm">
            {error}
          </div>
        )}

        {showResend && (
          <div className="mb-4 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded text-sm">
            <p className="text-yellow-800 dark:text-yellow-300 font-medium mb-2">Need a new verification link?</p>
            {resendStatus === 'sent' ? (
              <p className="text-green-700 dark:text-green-400">Verification email sent — check your inbox.</p>
            ) : resendStatus === 'error' ? (
              <p className="text-red-700 dark:text-red-400">Failed to send email. Try again later.</p>
            ) : (
              <div className="flex gap-2">
                <input
                  type="email"
                  value={resendEmail}
                  onChange={(e) => setResendEmail(e.target.value)}
                  placeholder="Your email address"
                  className="flex-1 px-3 py-1.5 border border-yellow-300 dark:border-yellow-700 rounded text-gray-800 dark:text-[#f4f7fa] text-sm focus:ring-2 focus:ring-[#f5b800] focus:border-transparent"
                />
                <button
                  type="button"
                  onClick={handleResendVerification}
                  disabled={resendStatus === 'sending' || !resendEmail}
                  className="px-3 py-1.5 bg-[#f5b800] text-[#07162b] rounded text-sm hover:bg-[#ffcf33] disabled:opacity-50 transition whitespace-nowrap font-semibold"
                >
                  {resendStatus === 'sending' ? 'Sending…' : 'Resend'}
                </button>
              </div>
            )}
          </div>
        )}

        {providers.ldap && (
          <div className="mb-4">
            <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-[#a8b8cb] cursor-pointer select-none">
              <input
                type="checkbox"
                checked={useLdap}
                onChange={(e) => setUseLdap(e.target.checked)}
                className="w-4 h-4 rounded accent-[#f5b800]"
              />
              Sign in with LDAP / Active Directory
            </label>
          </div>
        )}

        <form onSubmit={handlePasswordLogin}>
          <div className="mb-4">
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-[#a8b8cb] mb-2">
              Username
            </label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-[#25364a] rounded-lg focus:ring-2 focus:ring-[#f5b800] focus:border-transparent"
              placeholder="username"
              autoFocus
              autoComplete="username"
            />
          </div>

          <div className="mb-6">
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-[#a8b8cb] mb-2">
              Password
            </label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-[#25364a] rounded-lg focus:ring-2 focus:ring-[#f5b800] focus:border-transparent"
              placeholder="••••••••"
              autoComplete="current-password"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-[#f5b800] text-[#07162b] py-2 px-4 rounded-lg hover:bg-[#ffcf33] disabled:opacity-50 transition font-semibold"
          >
            {loading ? 'Signing in…' : 'Sign In'}
          </button>
        </form>

        <div className="mt-4 text-center space-y-2">
          <div>
            <span className="text-gray-600 dark:text-[#a8b8cb] text-sm">Don&apos;t have an account? </span>
            <Link to="/register" className="text-[#f5b800] text-sm hover:underline font-medium">
              Register
            </Link>
          </div>
          <div>
            <Link to="/forgot-password" className="text-[#f5b800] text-sm hover:underline">
              Forgot password?
            </Link>
          </div>
        </div>

        {(providers.oauth2 || providers.saml) && (
          <div className="mt-6 pt-6 border-t border-gray-200 dark:border-[#25364a]">
            <p className="text-xs text-gray-500 dark:text-[#a8b8cb]/60 text-center mb-3">Or continue with</p>
            <div className="flex flex-col gap-2">
              {providers.oauth2 && (
                <button
                  type="button"
                  onClick={() => { window.location.href = '/api/v1/auth/oauth2/login' }}
                  className="w-full border border-gray-300 dark:border-[#25364a] text-gray-700 dark:text-[#a8b8cb] py-2 px-4 rounded-lg hover:bg-gray-50 dark:hover:bg-[#0d2848] transition font-medium text-sm"
                >
                  Sign in with SSO
                </button>
              )}
              {providers.saml && (
                <button
                  type="button"
                  onClick={() => { window.location.href = '/api/v1/auth/saml/login' }}
                  className="w-full border border-gray-300 dark:border-[#25364a] text-gray-700 dark:text-[#a8b8cb] py-2 px-4 rounded-lg hover:bg-gray-50 dark:hover:bg-[#0d2848] transition font-medium text-sm"
                >
                  Sign in with SAML
                </button>
              )}
            </div>
          </div>
        )}

        {!providers.oauth2 && !providers.saml && (
          <div className="mt-6 pt-6 border-t border-gray-200 dark:border-[#25364a]">
            <p className="text-xs text-gray-500 dark:text-[#a8b8cb]/60">
              Your password is securely hashed and never transmitted in plain text.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
