import { useState, useEffect } from 'react'
import {
  generateTokenForMe,
  getAdminConfig,
  getApiTokenAnalytics,
  getAutomationPolicies,
  getIntegrationTemplates,
  testDnsConnection,
  testTechnitiumConnection,
  testLdapConnection,
} from '../api/client'

const PLATFORMS = [
  {
    id: 'n8n',
    name: 'n8n',
    description: 'Self-hosted workflow automation',
    steps: [
      'In n8n, create an HTTP Request node and set the URL to your IPAM instance.',
      'Under Authentication, choose "Header Auth" and add the header Authorization: Bearer <your-token>.',
      'Use the Webhook Trigger node to receive IPAM events (configure the webhook URL in IPAM → Admin → Webhooks).',
      'Pair trigger nodes with HTTP Request action nodes to read/write IPAM data.',
    ],
  },
  {
    id: 'zapier',
    name: 'Zapier',
    description: 'Cloud workflow automation',
    steps: [
      'In Zapier, create a Zap with a "Webhook by Zapier" trigger to receive IPAM events.',
      'Copy the Zapier webhook URL into IPAM → Admin → Webhooks to push events.',
      'For action steps, use the "HTTP by Zapier" action with your IPAM API URL.',
      'Set the Authorization header to Bearer <your-token> in the HTTP action settings.',
    ],
  },
  {
    id: 'make',
    name: 'Make (Integromat)',
    description: 'Visual automation platform',
    steps: [
      'In Make, add a "Webhooks" module as a trigger and copy the URL into IPAM → Admin → Webhooks.',
      'Add an "HTTP" module for actions, set the URL to your IPAM API endpoint.',
      'Under Headers, add Authorization: Bearer <your-token>.',
      'Map the response fields to downstream modules in your scenario.',
    ],
  },
]

const KEY_ENDPOINTS = [
  { method: 'GET', path: '/api/v1/subnets/{id}/next-available', desc: 'Preview the next free IP without allocating it' },
  { method: 'POST', path: '/api/v1/subnets/{subnetID}/ip-addresses/allocate', desc: 'Allocate the next available IP address' },
  { method: 'POST', path: '/api/v1/ip-addresses/{id}/assign', desc: 'Assign an IP to a host' },
  { method: 'POST', path: '/api/v1/ip-addresses/{id}/release', desc: 'Release an IP address' },
  { method: 'GET', path: '/api/v1/networks', desc: 'List all networks' },
  { method: 'GET', path: '/api/v1/networks/{id}/subnets', desc: 'List subnets in a network' },
  { method: 'GET', path: '/api/v1/subnets/{subnetID}/ip-addresses', desc: 'List IPs in a subnet' },
]

const METHOD_COLORS = {
  GET: 'bg-blue-100 text-blue-700',
  POST: 'bg-green-100 text-green-700',
  PUT: 'bg-amber-100 text-amber-700',
  DELETE: 'bg-red-100 text-red-700',
}

function IntegrationHealthPanel() {
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
      setStatus(s => ({ ...s, [key]: { ok: true, message: 'Connected' } }))
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Connection failed'
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
      <h2 className="text-base font-semibold text-gray-800">Integration Health</h2>
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
                  {intg.enabled && intg.configured ? 'Enabled' : intg.configured ? 'Disabled' : 'Not configured'}
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
                {st === 'testing' ? 'Testing…' : 'Test Connection'}
              </button>
            </div>
          )
        })}
      </div>
    </network>
  )
}

function AutomationOverview() {
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
      <h2 className="text-base font-semibold text-gray-800">Automation Control Plane</h2>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div className="border border-gray-200 rounded-lg p-4 bg-white">
          <div className="text-2xl font-semibold text-gray-900">{templates.length}</div>
          <div className="text-xs text-gray-500">Integration templates</div>
        </div>
        <div className="border border-gray-200 rounded-lg p-4 bg-white">
          <div className="text-2xl font-semibold text-gray-900">{tokens.length}</div>
          <div className="text-xs text-gray-500">API tokens tracked</div>
        </div>
        <div className="border border-gray-200 rounded-lg p-4 bg-white">
          <div className="text-2xl font-semibold text-gray-900">{policies.filter(p => p.enabled).length}</div>
          <div className="text-xs text-gray-500">Active policies</div>
        </div>
      </div>
      {tokens.length > 0 && (
        <div className="rounded border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Token</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Owner</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Scope</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Usage</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Rate Limit</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {tokens.slice(0, 6).map(t => (
                <tr key={t.id}>
                  <td className="px-4 py-3 text-gray-800">{t.name}</td>
                  <td className="px-4 py-3 text-gray-600">{t.username || t.userId}</td>
                  <td className="px-4 py-3 text-gray-600">{t.scope}</td>
                  <td className="px-4 py-3 text-gray-600">{t.usageCount}</td>
                  <td className="px-4 py-3 text-gray-600">{t.rateLimitPerMinute}/min</td>
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
      setError('Failed to generate token')
    } finally {
      setGenerating(false)
    }
  }

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-1">Automation Integrations</h1>
        <p className="text-sm text-gray-500">
          Connect IPAM to n8n, Zapier, Make, or any HTTP-capable automation platform using API tokens and webhooks.
        </p>
      </div>

      <IntegrationHealthPanel />
      <AutomationOverview />

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">1. Generate an API Token</h2>
        <form onSubmit={handleGenerate} className="flex gap-2 items-end">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-600 mb-1">Token name</label>
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
            {generating ? 'Generating…' : 'Generate'}
          </button>
        </form>
        {error && <p className="text-red-600 text-sm">{error}</p>}
        {token && (
          <div>
            <p className="text-xs text-amber-600 font-medium mb-1">Copy this token now — it will not be shown again.</p>
            <div className="bg-yellow-50 border border-yellow-200 rounded px-4 py-3 font-mono text-sm break-all select-all">
              {token}
            </div>
          </div>
        )}
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">2. Key API Endpoints</h2>
        <div className="rounded border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600 w-16">Method</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Endpoint</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Description</th>
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
          Base URL: <code className="bg-gray-100 px-1 rounded">{baseUrl}</code>.
          All requests require <code className="bg-gray-100 px-1 rounded">Authorization: Bearer &lt;token&gt;</code>.
        </p>
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">3. Platform Setup</h2>
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
        <h2 className="text-base font-semibold text-gray-800">4. Webhook Events</h2>
        <p className="text-sm text-gray-600">
          Configure outbound webhooks in <a href="/admin/webhooks" className="text-blue-600 hover:underline">Admin → Webhooks</a>.
          IPAM will POST a JSON payload to your automation platform URL on every IP, subnet, or network change.
        </p>
        <p className="text-sm text-gray-500">
          In n8n, use a <strong>Webhook</strong> trigger node. In Zapier, use <strong>Webhooks by Zapier</strong>.
          In Make, use the <strong>Webhooks</strong> module.
        </p>
      </network>
    </div>
  )
}
