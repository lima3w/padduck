import { useState, useEffect } from 'react'
import * as client from '../../api/auth'

export default function TokensTab() {
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
                    Created {new Date(t.createdAt).toLocaleDateString()}
                    {t.lastUsedAt ? ` · Last used ${new Date(t.lastUsedAt).toLocaleDateString()}` : ' · Never used'}
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
