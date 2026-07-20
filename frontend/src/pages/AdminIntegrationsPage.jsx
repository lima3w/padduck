import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { generateTokenForMe } from '../api/auth'
import { testDnsConnection, testTechnitiumConnection } from '../api/dns'
import { getAdminConfig, getApiTokenAnalytics, getAutomationPolicies, getIntegrationTemplates, testLdapConnection } from '../api/admin'

const PLATFORM_META = [
  { id: 'n8n', name: 'n8n', stepCount: 4 },
  { id: 'zapier', name: 'Zapier', stepCount: 4 },
  { id: 'make', name: 'Make (Integromat)', stepCount: 4 },
]

const ENDPOINT_KEYS = [
  { method: 'GET', path: '/api/v1/subnets/{id}/next-available', descKey: 'previewNextAvailable' },
  { method: 'POST', path: '/api/v1/subnets/{subnetID}/ip-addresses/allocate', descKey: 'allocateNextAvailable' },
  { method: 'POST', path: '/api/v1/ip-addresses/{id}/assign', descKey: 'assignIp' },
  { method: 'POST', path: '/api/v1/ip-addresses/{id}/release', descKey: 'releaseIp' },
  { method: 'GET', path: '/api/v1/networks', descKey: 'listNetworks' },
  { method: 'GET', path: '/api/v1/networks/{id}/subnets', descKey: 'listSubnetsInNetwork' },
  { method: 'GET', path: '/api/v1/subnets/{subnetID}/ip-addresses', descKey: 'listIpsInSubnet' },
]

const METHOD_COLORS = {
  GET: 'bg-blue-100 text-blue-700',
  POST: 'bg-green-100 text-green-700',
  PUT: 'bg-amber-100 text-amber-700',
  DELETE: 'bg-red-100 text-red-700',
}

function IntegrationHealthPanel() {
  const { t } = useTranslation()
  const [config, setConfig] = useState(null)
  const [status, setStatus] = useState({}) // key → null | 'testing' | { ok, message }

  useEffect(() => {
    getAdminConfig()
      .then(res => setConfig(res.data?.config || {}))
      .catch(() => {})
  }, [])

  const test = async (key, fn) => {
    setStatus(s => ({ ...s, [key]: 'testing' }))
    try {
      await fn()
      setStatus(s => ({ ...s, [key]: { ok: true, message: t('adminIntegrations.connected') } }))
    } catch (err) {
      const msg = err.response?.data?.error || err.message || t('adminIntegrations.connectionFailed')
      setStatus(s => ({ ...s, [key]: { ok: false, message: msg } }))
    }
  }

  if (!config) return null

  const integrations = [
    {
      key: 'pdns',
      name: 'PowerDNS',
      enabled: config.pdns_enabled === 'true',
      configured: !!(config.pdns_api_url && config.pdns_api_key),
      onTest: () => test('pdns', testDnsConnection),
    },
    {
      key: 'technitium',
      name: 'Technitium DNS',
      enabled: !!(config.technitium_url && config.technitium_token),
      configured: !!(config.technitium_url && config.technitium_token),
      onTest: () => test('technitium', testTechnitiumConnection),
    },
    {
      key: 'ldap',
      name: 'LDAP',
      enabled: config.ldap_enabled === 'true',
      configured: !!(config.ldap_url),
      onTest: () => test('ldap', testLdapConnection),
    },
  ]

  return (
    <network className="space-y-3">
      <h2 className="text-base font-semibold text-gray-800">{t('adminIntegrations.healthTitle')}</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {integrations.map(intg => {
          const st = status[intg.key]
          const canTest = intg.configured && intg.enabled
          return (
            <div key={intg.key} className="border border-gray-200 rounded-lg p-4 space-y-2 bg-white">
              <div className="flex items-center justify-between">
                <span className="font-medium text-sm text-gray-800">{intg.name}</span>
                <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                  intg.enabled && intg.configured
                    ? 'bg-green-100 text-green-700'
                    : intg.configured
                    ? 'bg-yellow-100 text-yellow-700'
                    : 'bg-gray-100 text-gray-500'
                }`}>
                  {intg.enabled && intg.configured ? t('adminIntegrations.enabled') : intg.configured ? t('adminIntegrations.disabled') : t('adminIntegrations.notConfigured')}
                </span>
              </div>
              {st && st !== 'testing' && (
                <p className={`text-xs ${st.ok ? 'text-green-600' : 'text-red-600'}`}>
                  {st.ok ? `✓ ${st.message}` : `✗ ${st.message}`}
                </p>
              )}
              <button
                onClick={intg.onTest}
                disabled={!canTest || st === 'testing'}
                className="text-xs px-3 py-1.5 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 disabled:opacity-40 transition"
              >
                {st === 'testing' ? t('adminIntegrations.testing') : t('adminIntegrations.testConnection')}
              </button>
            </div>
          )
        })}
      </div>
    </network>
  )
}

function AutomationOverview() {
  const { t } = useTranslation()
  const [templates, setTemplates] = useState([])
  const [tokens, setTokens] = useState([])
  const [policies, setPolicies] = useState([])

  useEffect(() => {
    Promise.all([
      getIntegrationTemplates(),
      getApiTokenAnalytics(),
      getAutomationPolicies(),
    ]).then(([templateRes, tokenRes, policyRes]) => {
      setTemplates(templateRes.data || [])
      setTokens(tokenRes.data || [])
      setPolicies(policyRes.data || [])
    }).catch(() => {})
  }, [])

  return (
    <network className="space-y-3">
      <h2 className="text-base font-semibold text-gray-800">{t('adminIntegrations.controlPlaneTitle')}</h2>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div className="border border-gray-200 rounded-lg p-4 bg-white">
          <div className="text-2xl font-semibold text-gray-900">{templates.length}</div>
          <div className="text-xs text-gray-500">{t('adminIntegrations.integrationTemplates')}</div>
        </div>
        <div className="border border-gray-200 rounded-lg p-4 bg-white">
          <div className="text-2xl font-semibold text-gray-900">{tokens.length}</div>
          <div className="text-xs text-gray-500">{t('adminIntegrations.apiTokensTracked')}</div>
        </div>
        <div className="border border-gray-200 rounded-lg p-4 bg-white">
          <div className="text-2xl font-semibold text-gray-900">{policies.filter(p => p.enabled).length}</div>
          <div className="text-xs text-gray-500">{t('adminIntegrations.activePolicies')}</div>
        </div>
      </div>
      {tokens.length > 0 && (
        <div className="rounded border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('adminIntegrations.tokenColumn')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('adminIntegrations.ownerColumn')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('adminIntegrations.scopeColumn')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('adminIntegrations.usageColumn')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('adminIntegrations.rateLimitColumn')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {tokens.slice(0, 6).map(tok => (
                <tr key={tok.id}>
                  <td className="px-4 py-3 text-gray-800">{tok.name}</td>
                  <td className="px-4 py-3 text-gray-600">{tok.username || tok.userId}</td>
                  <td className="px-4 py-3 text-gray-600">{tok.scope}</td>
                  <td className="px-4 py-3 text-gray-600">{tok.usageCount}</td>
                  <td className="px-4 py-3 text-gray-600">{tok.rateLimitPerMinute}/min</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </network>
  )
}

export default function AdminIntegrationsPage() {
  const { t } = useTranslation()
  const PLATFORMS = PLATFORM_META.map(({ id, name, stepCount }) => ({
    id,
    name,
    description: t(`adminIntegrations.platforms.${id}.description`),
    steps: Array.from({ length: stepCount }, (_, i) => t(`adminIntegrations.platforms.${id}.steps.${i}`)),
  }))
  const KEY_ENDPOINTS = ENDPOINT_KEYS.map(ep => ({ ...ep, desc: t(`adminIntegrations.endpoints.${ep.descKey}`) }))
  const [token, setToken] = useState('')
  const [tokenName, setTokenName] = useState('automation-token')
  const [generating, setGenerating] = useState(false)
  const [error, setError] = useState('')
  const [activePlatform, setActivePlatform] = useState('n8n')

  const platform = PLATFORMS.find(p => p.id === activePlatform)
  const baseUrl = window.location.origin

  async function handleGenerate(e) {
    e.preventDefault()
    setGenerating(true)
    setError('')
    try {
      const res = await generateTokenForMe(tokenName)
      setToken(res.data?.token || res.data?.rawToken || '')
    } catch {
      setError(t('adminIntegrations.generateFailed'))
    } finally {
      setGenerating(false)
    }
  }

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-1">{t('adminIntegrations.title')}</h1>
        <p className="text-sm text-gray-500">
          {t('adminIntegrations.subtitle')}
        </p>
      </div>

      <IntegrationHealthPanel />
      <AutomationOverview />

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">{t('adminIntegrations.step1Title')}</h2>
        <form onSubmit={handleGenerate} className="flex gap-2 items-end">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('adminIntegrations.tokenNameLabel')}</label>
            <input
              type="text"
              value={tokenName}
              onChange={e => setTokenName(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <button
            type="submit"
            disabled={generating}
            className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {generating ? t('adminIntegrations.generating') : t('adminIntegrations.generate')}
          </button>
        </form>
        {error && <p className="text-red-600 text-sm">{error}</p>}
        {token && (
          <div>
            <p className="text-xs text-amber-600 font-medium mb-1">{t('adminIntegrations.copyTokenNotice')}</p>
            <div className="bg-yellow-50 border border-yellow-200 rounded px-4 py-3 font-mono text-sm break-all select-all">
              {token}
            </div>
          </div>
        )}
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">{t('adminIntegrations.step2Title')}</h2>
        <div className="rounded border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600 w-16">{t('adminIntegrations.methodColumn')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('adminIntegrations.endpointColumn')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('common.description')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {KEY_ENDPOINTS.map(ep => (
                <tr key={ep.path}>
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-0.5 rounded text-xs font-bold ${METHOD_COLORS[ep.method]}`}>
                      {ep.method}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-mono text-gray-800 text-xs">{ep.path}</td>
                  <td className="px-4 py-3 text-gray-600">{ep.desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p className="text-xs text-gray-500">
          {t('adminIntegrations.baseUrlLabel')}<code className="bg-gray-100 px-1 rounded">{baseUrl}</code>{t('adminIntegrations.baseUrlSuffix')}
          {' '}{t('adminIntegrations.allRequestsRequirePrefix')}<code className="bg-gray-100 px-1 rounded">Authorization: Bearer &lt;token&gt;</code>{t('adminIntegrations.allRequestsRequireSuffix')}
        </p>
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">{t('adminIntegrations.step3Title')}</h2>
        <div className="flex gap-2">
          {PLATFORMS.map(p => (
            <button
              key={p.id}
              onClick={() => setActivePlatform(p.id)}
              className={`px-4 py-2 rounded text-sm font-medium transition-colors ${
                activePlatform === p.id
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {p.name}
            </button>
          ))}
        </div>
        {platform && (
          <div className="bg-gray-50 border border-gray-200 rounded p-4 space-y-3">
            <p className="text-sm font-medium text-gray-700">{platform.description}</p>
            <ol className="list-decimal list-inside space-y-1.5 text-sm text-gray-700">
              {platform.steps.map((step, i) => (
                <li key={i}>{step}</li>
              ))}
            </ol>
          </div>
        )}
      </network>

      <network className="space-y-2">
        <h2 className="text-base font-semibold text-gray-800">{t('adminIntegrations.step4Title')}</h2>
        <p className="text-sm text-gray-600">
          {t('adminIntegrations.webhookEventsPrefix')}<a href="/admin/webhooks" className="text-blue-600 hover:underline">{t('adminIntegrations.webhookEventsLinkText')}</a>{t('adminIntegrations.webhookEventsSuffix')}
        </p>
        <p className="text-sm text-gray-500">
          {t('adminIntegrations.webhookEventsN8nPrefix')}<strong>{t('adminIntegrations.webhookEventsN8nStrong')}</strong>{t('adminIntegrations.webhookEventsN8nSuffix')}
          {t('adminIntegrations.webhookEventsZapierPrefix')}<strong>{t('adminIntegrations.webhookEventsZapierStrong')}</strong>{t('adminIntegrations.webhookEventsZapierSuffix')}
          {t('adminIntegrations.webhookEventsMakePrefix')}<strong>{t('adminIntegrations.webhookEventsMakeStrong')}</strong>{t('adminIntegrations.webhookEventsMakeSuffix')}
        </p>
      </network>
    </div>
  )
}
