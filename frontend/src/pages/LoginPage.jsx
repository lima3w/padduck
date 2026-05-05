import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import * as client from '../api/client'

export default function LoginPage() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [userId, setUserId] = useState('1')
  const [tokenName, setTokenName] = useState('CLI Token')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [showToken, setShowToken] = useState(false)
  const [generatedToken, setGeneratedToken] = useState('')

  const handleGenerateToken = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const response = await client.generateTokenAnonymous(userId, tokenName)
      const { token } = response.data

      // Fetch user data
      const userResponse = await client.generateTokenForMe(tokenName)
      const userData = {
        id: parseInt(userId),
        username: `User ${userId}`,
        email: `user${userId}@localhost`,
      }

      // Login
      login(token, userData)
      setGeneratedToken(token)
      setShowToken(true)
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to generate token')
    } finally {
      setLoading(false)
    }
  }

  const handleContinue = () => {
    navigate('/')
  }

  if (showToken) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Token Generated</h1>
          <p className="text-gray-600 mb-4">
            Your API token has been generated. Store it securely—you won't see it again.
          </p>

          <div className="bg-gray-50 border border-gray-200 rounded p-3 mb-4 break-all">
            <code className="text-sm text-gray-700 font-mono">{generatedToken}</code>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded p-3 mb-6">
            <p className="text-sm text-blue-800">
              Use this token in the Authorization header: <code>Bearer {generatedToken}</code>
            </p>
          </div>

          <button
            onClick={handleContinue}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 transition"
          >
            Continue to Dashboard
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">IPAM Next</h1>
        <p className="text-gray-600 mb-8">Generate an API token to get started</p>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleGenerateToken}>
          <div className="mb-4">
            <label htmlFor="userId" className="block text-sm font-medium text-gray-700 mb-2">
              User ID
            </label>
            <input
              type="number"
              id="userId"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="1"
            />
            <p className="text-xs text-gray-500 mt-1">Default admin user is ID 1</p>
          </div>

          <div className="mb-6">
            <label htmlFor="tokenName" className="block text-sm font-medium text-gray-700 mb-2">
              Token Name
            </label>
            <input
              type="text"
              id="tokenName"
              value={tokenName}
              onChange={(e) => setTokenName(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="My API Token"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {loading ? 'Generating...' : 'Generate Token'}
          </button>
        </form>

        <div className="mt-6 pt-6 border-t border-gray-200">
          <p className="text-xs text-gray-500">
            Tokens are stored securely using SHA-256 hashing. Keep your token safe and never commit it to version control.
          </p>
        </div>
      </div>
    </div>
  )
}
