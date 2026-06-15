import { useState, useEffect, useCallback } from 'react'
import * as client from '../api/admin'

const defaultConfig = {
  enabled: false,
  provider_name: '',
  discovery_url: '',
  authorization_url: '',
  token_url: '',
  userinfo_url: '',
  client_id: '',
  client_secret: '',
  scopes: 'openid email profile',
}

export default function AdminOAuth2Page() {
  const [config, setConfig] = useState(defaultConfig)
  const [secretSet, setSecretSet] = useState(false)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState({ text: '', type: '' })
  const [loading, setLoading] = useState(true)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await client.getOAuth2Config().catch(() => null)
      if (res) {
        const c = res.data
        setSecretSet(c.clientSecret === '****')
        setConfig({ ...defaultConfig, ...c, clientSecret: '', client_secret: '' })
      }
    } catch (err) {
      showMessage('Failed to load OAuth2 config: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  const showMessage = (text, type = 'success') => {
    setMessage({ text, type })
    setTimeout(() => setMessage({ text: '', type: '' }), 4000)
  }

  const handleChange = (key, value) => {
    setConfig((prev) => ({ ...prev, [key]: value }))
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = { ...config }
      if (!payload.client_secret && !payload.clientSecret) {
        delete payload.client_secret
        delete payload.clientSecret
      }
      await client.updateOAuth2Config(payload)
      showMessage('OAuth2 configuration saved')
      loadData()
    } catch (err) {
      showMessage('Save failed: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSaving(false)
    }
  }

  const redirectUri = `${window.location.origin}/auth/callback`
  const hasDiscovery = !!(config.discovery_url || config.discoveryUrl)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
        Loading OAuth2 settings...
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">OAuth2 / OIDC</h1>

      {message.text && (
        <div
          className={`mb-4 p-4 rounded text-sm ${
            message.type === 'error'
              ? 'bg-red-50 border border-red-200 text-red-700 dark:bg-red-900/30 dark:border-red-700 dark:text-red-300'
              : 'bg-green-50 border border-green-200 text-green-700 dark:bg-green-900/30 dark:border-green-700 dark:text-green-300'
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="space-y-6">
        {/* Provider Settings */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Provider Settings</h2>

          <label className="flex items-center gap-3 mb-4 cursor-pointer">
            <input
              type="checkbox"
              checked={!!config.enabled}
              onChange={(e) => handleChange('enabled', e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              <strong>Enable OAuth2 / OIDC authentication</strong>
              <span className="block text-gray-500 dark:text-gray-400">Allow users to sign in via an external OAuth2 provider</span>
            </span>
          </label>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Provider Name</label>
            <input
              type="text"
              value={config.provider_name || config.providerName || ''}
              onChange={(e) => handleChange('provider_name', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="Google"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Displayed on the login button, e.g. &ldquo;Sign in with Google&rdquo;.</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Redirect URI</label>
            <input
              type="text"
              value={redirectUri}
              readOnly
              className="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded text-sm bg-gray-50 dark:bg-gray-700/50 text-gray-600 dark:text-gray-400 cursor-default"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Register this URI with your OAuth2 provider.</p>
          </div>
        </div>

        {/* OIDC Discovery */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">OIDC Discovery</h2>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Discovery URL</label>
            <input
              type="url"
              value={config.discovery_url || config.discoveryUrl || ''}
              onChange={(e) => handleChange('discovery_url', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="https://accounts.google.com/.well-known/openid-configuration"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              When set, the authorization, token, and userinfo URLs are discovered automatically.
            </p>
          </div>
        </div>

        {/* Manual URLs */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-1 text-gray-900 dark:text-gray-100">Manual URLs</h2>
          {hasDiscovery && (
            <p className="text-sm text-amber-600 dark:text-amber-400 mb-4">
              A discovery URL is set — these fields are optional and will be overridden by auto-discovery.
            </p>
          )}
          {!hasDiscovery && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">Required when no discovery URL is provided.</p>
          )}

          <div className={`space-y-4 ${hasDiscovery ? 'opacity-60' : ''}`}>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Authorization URL</label>
              <input
                type="url"
                value={config.authorization_url || config.authorizationUrl || ''}
                onChange={(e) => handleChange('authorization_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="https://accounts.google.com/o/oauth2/auth"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Token URL</label>
              <input
                type="url"
                value={config.token_url || config.tokenUrl || ''}
                onChange={(e) => handleChange('token_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="https://oauth2.googleapis.com/token"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Userinfo URL</label>
              <input
                type="url"
                value={config.userinfo_url || config.userinfoUrl || ''}
                onChange={(e) => handleChange('userinfo_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="https://www.googleapis.com/oauth2/v3/userinfo"
              />
            </div>
          </div>
        </div>

        {/* Credentials */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Credentials</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Client ID</label>
            <input
              type="text"
              value={config.client_id || config.clientId || ''}
              onChange={(e) => handleChange('client_id', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="your-client-id"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Client Secret</label>
            <input
              type="password"
              value={config.client_secret || config.clientSecret || ''}
              onChange={(e) => handleChange('client_secret', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder={secretSet ? 'unchanged' : 'Enter client secret'}
            />
            {secretSet && (
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Leave blank to keep the existing secret.</p>
            )}
          </div>
        </div>

        {/* Scopes */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Scopes</h2>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Requested Scopes</label>
            <input
              type="text"
              value={config.scopes || ''}
              onChange={(e) => handleChange('scopes', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="openid email profile"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Space-separated list of OAuth2 scopes to request.</p>
          </div>
        </div>

        <button
          onClick={handleSave}
          disabled={saving}
          className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
        >
          {saving ? 'Saving...' : 'Save Settings'}
        </button>
      </div>
    </div>
  )
}
