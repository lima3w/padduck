import { useState } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { resetPassword } from '../api/client'

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') || ''
  const navigate = useNavigate()

  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4">
        <div className="w-full max-w-sm bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 text-center space-y-4">
          <p className="text-red-600 dark:text-red-400 text-sm">
            Invalid or missing reset token. Please request a new password reset link.
          </p>
          <Link to="/forgot-password" className="text-blue-600 dark:text-blue-400 text-sm hover:underline">
            Request new link
          </Link>
        </div>
      </div>
    )
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    if (password !== confirm) {
      setError('Passwords do not match.')
      return
    }
    if (password.length < 8) {
      setError('Password must be at least 8 characters.')
      return
    }
    setLoading(true)
    try {
      await resetPassword(token, password)
      navigate('/login', { state: { message: 'Password reset successful. You can now log in.' } })
    } catch (err) {
      const msg = err.response?.data?.error || ''
      if (msg.toLowerCase().includes('expired') || msg.toLowerCase().includes('invalid')) {
        setError('This reset link has expired or is invalid. Please request a new one.')
      } else {
        setError(msg || 'Failed to reset password. Please try again.')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4">
      <div className="w-full max-w-sm">
        <h1 className="text-2xl font-bold text-center text-gray-900 dark:text-gray-100 mb-6">
          Set a new password
        </h1>

        <form
          onSubmit={handleSubmit}
          className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 space-y-4"
        >
          {error && (
            <p className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded px-3 py-2">
              {error}
              {error.includes('expired') && (
                <> {' '}
                  <Link to="/forgot-password" className="underline">Request new link</Link>
                </>
              )}
            </p>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              New password
            </label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Confirm password
            </label>
            <input
              type="password"
              required
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              placeholder="••••••••"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 disabled:opacity-50 text-sm font-medium transition"
          >
            {loading ? 'Saving…' : 'Set new password'}
          </button>
        </form>
      </div>
    </div>
  )
}
