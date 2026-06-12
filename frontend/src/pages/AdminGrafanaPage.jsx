import { useState } from 'react'
import { generateTokenForMe } from '../api/auth'

const METRICS = [
  { name: 'subnet_utilization', desc: 'All subnets with CIDR, network, used/total IPs, and utilisation %' },
  { name: 'ip_by_status', desc: 'IP address counts grouped by status (assigned, available, reserved, …)' },
  { name: 'section_summary', desc: 'Per-network subnet count, total IPs, and used IPs' },
]

export default function AdminGrafanaPage() {
  const [token, setToken] = useState('')
  const [tokenName, setTokenName] = useState('grafana-datasource')
  const [generating, setGenerating] = useState(false)
  const [error, setError] = useState('')

  const baseUrl = window.location.origin + '/api/grafana'

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
    <div className="p-6 max-w-3xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-1">Grafana Data Source</h1>
        <p className="text-sm text-gray-500">
          Configure Grafana to query IPAM data using the built-in SimpleJSON datasource endpoint.
        </p>
      </div>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">1. Datasource URL</h2>
        <div className="bg-gray-50 border border-gray-200 rounded px-4 py-3 font-mono text-sm text-gray-800 select-all">
          {baseUrl}
        </div>
        <p className="text-xs text-gray-500">
          Use the <strong>JSON API</strong> or <strong>SimpleJSON</strong> Grafana plugin and set this as the base URL.
        </p>
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">2. Generate an API Token</h2>
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
            <p className="text-xs text-gray-500 mt-1">
              Add this as a <strong>Bearer</strong> token in the Grafana datasource HTTP Headers:{' '}
              <code className="bg-gray-100 px-1 rounded">Authorization: Bearer {'<token>'}</code>
            </p>
          </div>
        )}
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">3. Available Metrics</h2>
        <div className="rounded border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Metric</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Description</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {METRICS.map(m => (
                <tr key={m.name}>
                  <td className="px-4 py-3 font-mono text-gray-900">{m.name}</td>
                  <td className="px-4 py-3 text-gray-600">{m.desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </network>

      <network className="space-y-2">
        <h2 className="text-base font-semibold text-gray-800">4. Grafana Plugin Setup</h2>
        <ol className="list-decimal list-inside space-y-1 text-sm text-gray-700">
          <li>Install the <strong>Marcusolsson JSON API</strong> or <strong>SimpleJSON</strong> plugin in Grafana.</li>
          <li>Add a new datasource using the URL above.</li>
          <li>Under <em>Custom HTTP Headers</em>, add <code className="bg-gray-100 px-1 rounded">Authorization</code> with value <code className="bg-gray-100 px-1 rounded">Bearer &lt;your-token&gt;</code>.</li>
          <li>Save &amp; test the connection (the health check endpoint should return <code className="bg-gray-100 px-1 rounded">ok</code>).</li>
          <li>In a panel, select a metric from the table above and render as a Table or Stat visualization.</li>
        </ol>
      </network>
    </div>
  )
}
