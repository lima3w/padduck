import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import * as client from '../../api/auth'

export default function TokensTab() {
  const { t } = useTranslation()
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
      setError(err.response?.data?.error || t('userTabs.tokens.createFailed'))
    } finally {
      setCreating(false)
    }
  }

  const handleRevoke = async (id) => {
    try {
      await client.revokeToken(id)
      setTokens((prev) => prev.filter((tok) => tok.id !== id))
    } catch (err) {
      setError(err.response?.data?.error || t('userTabs.tokens.revokeFailed'))
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">{t('settings.tabs.apiTokens')}</h2>
        <p className="text-sm text-gray-600 mb-4">
          {t('userTabs.tokens.subtitle')}
        </p>

        {newToken && (
          <div className="mb-4 p-4 bg-green-50 border border-green-200 rounded">
            <p className="text-sm font-medium text-green-800 mb-2">{t('userTabs.tokens.createdCopyNow')}</p>
            <code className="block p-2 bg-white border border-green-200 rounded font-mono text-xs break-all text-gray-700">{newToken}</code>
            <button
              type="button"
              onClick={() => setNewToken(null)}
              className="mt-2 text-xs text-green-700 hover:underline"
            >
              {t('common.dismiss')}
            </button>
          </div>
        )}

        {error && <p className="mb-4 text-sm text-red-600">{error}</p>}

        <form onSubmit={handleCreate} className="flex gap-2 mb-6">
          <input
            type="text"
            value={tokenName}
            onChange={(e) => setTokenName(e.target.value)}
            placeholder={t('userTabs.tokens.namePlaceholder')}
            className="flex-1 px-3 py-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <button
            type="submit"
            disabled={creating || !tokenName.trim()}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {creating ? t('userTabs.tokens.creating') : t('userTabs.tokens.createToken')}
          </button>
        </form>

        {loading ? (
          <p className="text-sm text-gray-500">{t('common.loading')}</p>
        ) : tokens.length === 0 ? (
          <p className="text-sm text-gray-500">{t('userTabs.tokens.empty')}</p>
        ) : (
          <div className="divide-y divide-gray-200 border border-gray-200 rounded">
            {tokens.map((tok) => (
              <div key={tok.id} className="flex items-center justify-between px-4 py-3">
                <div>
                  <p className="text-sm font-medium text-gray-900">{tok.name}</p>
                  <p className="text-xs text-gray-500">
                    {t('userTabs.tokens.created', { date: new Date(tok.createdAt).toLocaleDateString() })}
                    {tok.lastUsedAt ? ` · ${t('userTabs.tokens.lastUsed', { date: new Date(tok.lastUsedAt).toLocaleDateString() })}` : ` · ${t('userTabs.tokens.neverUsed')}`}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => handleRevoke(tok.id)}
                  className="text-sm text-red-600 hover:text-red-800 transition"
                >
                  {t('common.revoke')}
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
