import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { requestPasswordReset } from '../api/auth'

export default function ForgotPasswordPage() {
  const { t } = useTranslation()
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await requestPasswordReset(email)
      setSubmitted(true)
    } catch (err) {
      setError(err.response?.data?.error || t('forgotPassword.genericError'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4">
      <div className="w-full max-w-sm">
        <h1 className="text-2xl font-bold text-center text-gray-900 dark:text-gray-100 mb-6">
          {t('forgotPassword.title')}
        </h1>

        {submitted ? (
          <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 text-center space-y-4">
            <p className="text-gray-700 dark:text-gray-300 text-sm">
              {t('forgotPassword.sentPrefix')}<strong>{email}</strong>{t('forgotPassword.sentSuffix')}
            </p>
            <Link to="/login" className="text-blue-600 dark:text-blue-400 text-sm hover:underline">
              {t('forgotPassword.backToLogin')}
            </Link>
          </div>
        ) : (
          <form
            onSubmit={handleSubmit}
            className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 space-y-4"
          >
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {t('forgotPassword.instructions')}
            </p>

            {error && (
              <p className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded px-3 py-2">
                {error}
              </p>
            )}

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('forgotPassword.emailAddress')}
              </label>
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder={t('forgotPassword.emailPlaceholder')}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 disabled:opacity-50 text-sm font-medium transition"
            >
              {loading ? t('forgotPassword.sending') : t('forgotPassword.sendResetLink')}
            </button>

            <p className="text-center text-sm text-gray-500 dark:text-gray-400">
              <Link to="/login" className="text-blue-600 dark:text-blue-400 hover:underline">
                {t('forgotPassword.backToLogin')}
              </Link>
            </p>
          </form>
        )}
      </div>
    </div>
  )
}
