import { useState, useEffect, useCallback } from 'react'
import * as client from '../api/admin'

const defaultConfig = {
  enabled: false,
  idp_metadata_url: '',
  idp_metadata_xml: '',
  entity_id: '',
  sp_cert: '',
}

export default function AdminSamlPage() {
  const [config, setConfig] = useState(defaultConfig)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState({ text: '', type: '' })
  const [loading, setLoading] = useState(true)
  const [showXmlPaste, setShowXmlPaste] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await client.getSamlConfig().catch(() => null)
      if (res) {
        setConfig({ ...defaultConfig, ...res.data })
      }
    } catch (err) {
      showMessage('Failed to load SAML config: ' + (err.response?.data?.error || err.message), 'error')
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
      await client.updateSamlConfig(config)
      showMessage('SAML configuration saved')
      loadData()
    } catch (err) {
      showMessage('Save failed: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSaving(false)
    }
  }

  const acsUrl = `${window.location.origin}/auth/saml/acs`
  const spCert = config.sp_cert || config.spCert || ''
  const certFingerprint = spCert ? spCert.replace(/-----[^-]+-----/g, '').replace(/\s/g, '').slice(0, 20) + '...' : null

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
        Loading SAML settings...
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">SAML 2.0</h1>

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
        {/* Enable toggle */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={!!config.enabled}
              onChange={(e) => handleChange('enabled', e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              <strong>Enable SAML 2.0 authentication</strong>
              <span className="block text-gray-500 dark:text-gray-400">Allow users to sign in via a SAML identity provider</span>
            </span>
          </label>
        </div>

        {/* Identity Provider */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Identity Provider</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">IDP Metadata URL</label>
            <input
              type="url"
              value={config.idp_metadata_url || config.idpMetadataUrl || ''}
              onChange={(e) => handleChange('idp_metadata_url', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="https://idp.example.com/metadata"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              The backend will fetch and parse this XML on save.
            </p>
          </div>

          <div>
            <button
              type="button"
              onClick={() => setShowXmlPaste((v) => !v)}
              className="text-sm text-blue-600 dark:text-blue-400 hover:underline mb-2"
            >
              {showXmlPaste ? 'Hide XML paste' : 'Or paste IDP metadata XML manually'}
            </button>
            {showXmlPaste && (
              <textarea
                rows={8}
                value={config.idp_metadata_xml || config.idpMetadataXml || ''}
                onChange={(e) => handleChange('idp_metadata_xml', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 font-mono"
                placeholder="<?xml version=&quot;1.0&quot;?> <EntityDescriptor ...>"
              />
            )}
          </div>
        </div>

        {/* Service Provider */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Service Provider</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Entity ID</label>
            <input
              type="text"
              value={config.entity_id || config.entityId || ''}
              onChange={(e) => handleChange('entity_id', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="https://ipam.example.com"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Unique identifier for this service provider. Usually the application URL.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Assertion Consumer Service (ACS) URL</label>
            <input
              type="text"
              value={acsUrl}
              readOnly
              className="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded text-sm bg-gray-50 dark:bg-gray-700/50 text-gray-600 dark:text-gray-400 cursor-default"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Register this URL with your identity provider.</p>
          </div>
        </div>

        {/* SP Certificate */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-2 text-gray-900 dark:text-gray-100">SP Certificate</h2>
          {certFingerprint ? (
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">Certificate fingerprint (first 20 chars):</p>
              <code className="text-xs font-mono bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded text-gray-800 dark:text-gray-200">
                {certFingerprint}
              </code>
            </div>
          ) : (
            <p className="text-sm text-gray-500 dark:text-gray-400">Auto-generated on first use.</p>
          )}
        </div>

        {/* Download SP Metadata */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-2 text-gray-900 dark:text-gray-100">SP Metadata</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">
            Download the SP metadata XML to configure your identity provider.
          </p>
          <a
            href="/api/v1/auth/saml/metadata"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-block bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 transition text-sm font-medium"
          >
            Download SP Metadata
          </a>
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
