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
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full">
          <h1 className="text-2xl font-bold text-gray-900 mb-1">Two-Factor Authentication</h1>
          <p className="text-gray-600 mb-6">Enter the 6-digit code from your authenticator app, or a backup code.</p>

          {error && (
            <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleMFASubmit}>
            <div className="mb-6">
              <label htmlFor="mfaCode" className="block text-sm font-medium text-gray-700 mb-2">
                Authentication Code
              </label>
              <input
                type="text"
                id="mfaCode"
                value={mfaCode}
                onChange={(e) => setMfaCode(e.target.value.replace(/\s/g, ''))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-center text-xl tracking-widest font-mono"
                placeholder="000000"
                maxLength={12}
                autoFocus
                autoComplete="one-time-code"
              />
            </div>

            <button
              type="submit"
              disabled={loading || !mfaCode}
              className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {loading ? 'Verifying…' : 'Verify'}
            </button>
          </form>

          <button
            type="button"
            onClick={() => { setMfaChallenge(null); setMfaCode(''); setError(''); }}
            className="w-full mt-3 text-sm text-gray-500 hover:text-gray-700 transition"
          >
            Back to sign in
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">IPAM Next</h1>
        <p className="text-gray-600 mb-6">Sign in to continue</p>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
            {error}
          </div>
        )}

        {showResend && (
          <div className="mb-4 p-4 bg-yellow-50 border border-yellow-200 rounded text-sm">
            <p className="text-yellow-800 font-medium mb-2">Need a new verification link?</p>
            {resendStatus === 'sent' ? (
              <p className="text-green-700">Verification email sent — check your inbox.</p>
            ) : resendStatus === 'error' ? (
              <p className="text-red-700">Failed to send email. Try again later.</p>
            ) : (
              <div className="flex gap-2">
                <input
                  type="email"
                  value={resendEmail}
                  onChange={(e) => setResendEmail(e.target.value)}
                  placeholder="Your email address"
                  className="flex-1 px-3 py-1.5 border border-yellow-300 rounded text-gray-800 text-sm focus:ring-2 focus:ring-yellow-400 focus:border-transparent"
                />
                <button
                  type="button"
                  onClick={handleResendVerification}
                  disabled={resendStatus === 'sending' || !resendEmail}
                  className="px-3 py-1.5 bg-yellow-600 text-white rounded text-sm hover:bg-yellow-700 disabled:opacity-50 transition whitespace-nowrap"
                >
                  {resendStatus === 'sending' ? 'Sending…' : 'Resend'}
                </button>
              </div>
            )}
          </div>
        )}

        {providers.ldap && (
          <div className="mb-4">
            <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={useLdap}
                onChange={(e) => setUseLdap(e.target.checked)}
                className="w-4 h-4 text-blue-600 rounded"
              />
              Sign in with LDAP / Active Directory
            </label>
          </div>
        )}

        <form onSubmit={handlePasswordLogin}>
          <div className="mb-4">
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-2">
              Username
            </label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="username"
              autoFocus
              autoComplete="username"
            />
          </div>

          <div className="mb-6">
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-2">
              Password
            </label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="••••••••"
              autoComplete="current-password"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {loading ? 'Signing in…' : 'Sign In'}
          </button>
        </form>

        <div className="mt-4 text-center">
          <span className="text-gray-600 text-sm">Don't have an account? </span>
          <Link to="/register" className="text-blue-600 text-sm hover:underline font-medium">
            Register
          </Link>
        </div>

        {(providers.oauth2 || providers.saml) && (
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-xs text-gray-500 text-center mb-3">Or continue with</p>
            <div className="flex flex-col gap-2">
              {providers.oauth2 && (
                <button
                  type="button"
                  onClick={() => { window.location.href = '/api/v1/auth/oauth2/login' }}
                  className="w-full border border-gray-300 text-gray-700 py-2 px-4 rounded-lg hover:bg-gray-50 transition font-medium text-sm"
                >
                  Sign in with SSO
                </button>
              )}
              {providers.saml && (
                <button
                  type="button"
                  onClick={() => { window.location.href = '/api/v1/auth/saml/login' }}
                  className="w-full border border-gray-300 text-gray-700 py-2 px-4 rounded-lg hover:bg-gray-50 transition font-medium text-sm"
                >
                  Sign in with SAML
                </button>
              )}
            </div>
          </div>
        )}

        {!providers.oauth2 && !providers.saml && (
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-xs text-gray-500">
              Your password is securely hashed and never transmitted in plain text.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
