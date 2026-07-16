import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import * as client from '../api/auth'

export default function VerifyEmailPage() {
  const { t } = useTranslation()
  const [searchParams] = useSearchParams()
  const [status, setStatus] = useState('verifying') // verifying, success, error
  const [message, setMessage] = useState('')

  useEffect(() => {
    const token = searchParams.get('token')
    if (!token) {
      setStatus('error')
      setMessage(t('verifyEmail.noToken'))
      return
    }

    client.verifyEmail(token)
      .then((response) => {
        setStatus('success')
        setMessage(response.data.message)
      })
      .catch((err) => {
        setStatus('error')
        setMessage(err.response?.data?.error || t('verifyEmail.failed'))
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams])

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full text-center">
        {status === 'verifying' && (
          <>
            <div className="text-blue-500 text-4xl mb-4 animate-spin">⟳</div>
            <h2 className="text-2xl font-bold text-gray-900 mb-2">{t('verifyEmail.verifying')}</h2>
            <p className="text-gray-600">{t('verifyEmail.verifyingDescription')}</p>
          </>
        )}

        {status === 'success' && (
          <>
            <div className="text-green-500 text-5xl mb-4">✓</div>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">{t('verifyEmail.successTitle')}</h2>
            <p className="text-gray-600 mb-6">{message}</p>
            <Link
              to="/login"
              className="block w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition font-medium"
            >
              {t('register.goToLogin')}
            </Link>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="text-red-500 text-5xl mb-4">✗</div>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">{t('verifyEmail.errorTitle')}</h2>
            <p className="text-gray-600 mb-6">{message}</p>
            <Link
              to="/login"
              className="block w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition font-medium"
            >
              {t('verifyEmail.backToLogin')}
            </Link>
          </>
        )}
      </div>
    </div>
  )
}
