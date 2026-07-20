import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
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
  const { t } = useTranslation()
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
      showMessage(t('adminOAuth2.loadFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setLoading(false)
    }
  }, [t])

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
      showMessage(t('adminOAuth2.configSaved'))
      loadData()
    } catch (err) {
      showMessage(t('adminOAuth2.saveFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSaving(false)
    }
  }

  const redirectUri = `${window.location.origin}/auth/callback`
  const hasDiscovery = !!(config.discovery_url || config.discoveryUrl)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
        {t('adminOAuth2.loadingSettings')}
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">{t('adminOAuth2.title')}</h1>

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
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{t('adminOAuth2.providerSettingsTitle')}</h2>

          <label className="flex items-center gap-3 mb-4 cursor-pointer">
            <input
              type="checkbox"
              checked={!!config.enabled}
              onChange={(e) => handleChange('enabled', e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              <strong>{t('adminOAuth2.enableOAuth2')}</strong>
              <span className="block text-gray-500 dark:text-gray-400">{t('adminOAuth2.enableOAuth2Hint')}</span>
            </span>
          </label>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.providerName')}</label>
            <input
              type="text"
              value={config.provider_name || config.providerName || ''}
              onChange={(e) => handleChange('provider_name', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="Google"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('adminOAuth2.providerNameHint')}</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.redirectUri')}</label>
            <input
              type="text"
              value={redirectUri}
              readOnly
              className="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded text-sm bg-gray-50 dark:bg-gray-700/50 text-gray-600 dark:text-gray-400 cursor-default"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('adminOAuth2.redirectUriHint')}</p>
          </div>
        </div>

        {/* OIDC Discovery */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{t('adminOAuth2.oidcDiscoveryTitle')}</h2>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.discoveryUrl')}</label>
            <input
              type="url"
              value={config.discovery_url || config.discoveryUrl || ''}
              onChange={(e) => handleChange('discovery_url', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="https://accounts.google.com/.well-known/openid-configuration"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {t('adminOAuth2.discoveryUrlHint')}
            </p>
          </div>
        </div>

        {/* Manual URLs */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-1 text-gray-900 dark:text-gray-100">{t('adminOAuth2.manualUrlsTitle')}</h2>
          {hasDiscovery && (
            <p className="text-sm text-amber-600 dark:text-amber-400 mb-4">
              {t('adminOAuth2.discoverySetWarning')}
            </p>
          )}
          {!hasDiscovery && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">{t('adminOAuth2.requiredWhenNoDiscovery')}</p>
          )}

          <div className={`space-y-4 ${hasDiscovery ? 'opacity-60' : ''}`}>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.authorizationUrl')}</label>
              <input
                type="url"
                value={config.authorization_url || config.authorizationUrl || ''}
                onChange={(e) => handleChange('authorization_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="https://accounts.google.com/o/oauth2/auth"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.tokenUrl')}</label>
              <input
                type="url"
                value={config.token_url || config.tokenUrl || ''}
                onChange={(e) => handleChange('token_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="https://oauth2.googleapis.com/token"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.userinfoUrl')}</label>
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
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{t('adminOAuth2.credentialsTitle')}</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.clientId')}</label>
            <input
              type="text"
              value={config.client_id || config.clientId || ''}
              onChange={(e) => handleChange('client_id', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="your-client-id"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.clientSecret')}</label>
            <input
              type="password"
              value={config.client_secret || config.clientSecret || ''}
              onChange={(e) => handleChange('client_secret', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder={secretSet ? t('adminOAuth2.unchanged') : t('adminOAuth2.enterClientSecret')}
            />
            {secretSet && (
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('adminOAuth2.leaveBlankToKeepSecret')}</p>
            )}
          </div>
        </div>

        {/* Scopes */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{t('adminOAuth2.scopesTitle')}</h2>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminOAuth2.requestedScopes')}</label>
            <input
              type="text"
              value={config.scopes || ''}
              onChange={(e) => handleChange('scopes', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="openid email profile"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('adminOAuth2.scopesHint')}</p>
          </div>
        </div>

        <button
          onClick={handleSave}
          disabled={saving}
          className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
        >
          {saving ? t('common.saving') : t('scanRetention.saveSettings')}
        </button>
      </div>
    </div>
  )
}
