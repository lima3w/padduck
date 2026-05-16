import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import * as client from '../api/client'

export default function AuthCallbackPage() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [error, setError] = useState(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const token = params.get('token')
    const errorMsg = params.get('error')

    if (errorMsg) {
      setError(decodeURIComponent(errorMsg))
      return
    }

    if (token) {
      // Store token first so the API call is authenticated
      localStorage.setItem('auth_token', token)
      // Fetch the user record, then fully login
      client.getCurrentUser()
        .then((res) => {
          login(token, res.data)
          navigate('/', { replace: true })
        })
        .catch(() => {
          // Partial login without user object — app will fetch on next load
          login(token, null)
          navigate('/', { replace: true })
        })
      return
    }

    // No token and no error — unexpected state
    setError('No authentication token received. Please try signing in again.')
  }, [])

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full text-center">
          <h1 className="text-xl font-bold text-gray-900 mb-4">Sign In Failed</h1>
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
            {error}
          </div>
          <Link
            to="/login"
            className="inline-block bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition font-medium"
          >
            Back to Sign In
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full text-center">
        <div className="flex justify-center mb-4">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
        </div>
        <p className="text-gray-600">Completing sign in...</p>
      </div>
    </div>
  )
}
