import { useState } from 'react'
import { checkAllDns, testDnsConnection, testTechnitiumConnection } from '../../api/dns'

export default function DnsTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  const [dnsTestStatus, setDnsTestStatus] = useState(null) // null | 'testing' | { ok, message }
  const [technitiumTestStatus, setTechnitiumTestStatus] = useState(null) // null | 'testing' | { ok, message }
  const [dnsBulkStatus, setDnsBulkStatus] = useState(null) // null | 'running' | { ok, message }

  const handleTestDns = async () => {
    setDnsTestStatus('testing')
    try {
      const res = await testDnsConnection()
      const msg = res.data?.message || 'Connected'
      setDnsTestStatus({ ok: true, message: msg })
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Connection failed'
      setDnsTestStatus({ ok: false, message: msg })
    }
  }

  const handleTestTechnitium = async () => {
    setTechnitiumTestStatus('testing')
    try {
      await testTechnitiumConnection({
        url: config?.technitium_url || '',
        token: config?.technitium_token || '',
        skip_tls: config?.technitium_skip_tls === 'true',
      })
      setTechnitiumTestStatus({ ok: true, message: 'Connected' })
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Connection failed'
      setTechnitiumTestStatus({ ok: false, message: msg })
    }
  }

  const handleDnsBulkCheck = async () => {
    setDnsBulkStatus('running')
    try {
      await checkAllDns()
      setDnsBulkStatus({ ok: true, message: 'DNS bulk check started in background' })
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Failed to start DNS check'
      setDnsBulkStatus({ ok: false, message: msg })
    }
  }

  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">PowerDNS Integration</h2>

            <label className="flex items-center gap-3 mb-4 cursor-pointer">
              <input
                type="checkbox"
                checked={config.pdns_enabled === 'true'}
                onChange={e => handleConfigChange('pdns_enabled', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">
                <strong>Enable PowerDNS integration</strong>
                <span className="block text-gray-500">Sync DNS records with PowerDNS server</span>
              </span>
            </label>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">API URL</label>
              <input
                type="url"
                value={config.pdns_api_url || ''}
                onChange={e => handleConfigChange('pdns_api_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="http://pdns.example.com:8081"
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">API Key</label>
              <input
                type="password"
                value={config.pdns_api_key || ''}
                onChange={e => handleConfigChange('pdns_api_key', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="••••••••"
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Default Zone</label>
              <input
                type="text"
                value={config.pdns_default_zone || ''}
                onChange={e => handleConfigChange('pdns_default_zone', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="example.com."
              />
              <p className="text-xs text-gray-500 mt-1">Include the trailing dot (FQDN format).</p>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">PTR Zones</label>
              <textarea
                rows={3}
                value={config.pdns_ptr_zones || ''}
                onChange={e => handleConfigChange('pdns_ptr_zones', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm font-mono"
                placeholder="0.168.192.in-addr.arpa., 1.0.10.in-addr.arpa."
              />
              <p className="text-xs text-gray-500 mt-1">Comma-separated list of reverse zones for PTR records.</p>
            </div>

            {dnsTestStatus && dnsTestStatus !== 'testing' && (
              <div className={`mb-4 px-3 py-2 rounded text-sm ${dnsTestStatus.ok ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
                {dnsTestStatus.ok ? `Connected: ${dnsTestStatus.message}` : `Error: ${dnsTestStatus.message}`}
              </div>
            )}
            <button
              onClick={handleTestDns}
              disabled={dnsTestStatus === 'testing'}
              className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
            >
              {dnsTestStatus === 'testing' ? 'Testing...' : 'Test PowerDNS Connection'}
            </button>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Technitium DNS Server</h2>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Server URL</label>
              <input
                type="url"
                value={config.technitium_url || ''}
                onChange={e => handleConfigChange('technitium_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="http://192.168.1.1"
              />
              <p className="text-xs text-gray-500 mt-1">Base URL of the Technitium DNS web interface (no trailing slash).</p>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">API Token</label>
              <input
                type="password"
                value={config.technitium_token || ''}
                onChange={e => handleConfigChange('technitium_token', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="••••••••"
              />
              <p className="text-xs text-gray-500 mt-1">API token from Technitium DNS administration panel.</p>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Default Zone (for DNS sync)</label>
              <input
                type="text"
                value={config.technitium_default_zone || ''}
                onChange={e => handleConfigChange('technitium_default_zone', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="example.com"
              />
              <p className="text-xs text-gray-500 mt-1">Zone where A records are created when an IP is assigned a DNS name.</p>
            </div>

            <div className="mb-4 flex items-center gap-2">
              <input
                type="checkbox"
                id="technitium_skip_tls"
                checked={config.technitium_skip_tls === 'true'}
                onChange={e => handleConfigChange('technitium_skip_tls', e.target.checked ? 'true' : 'false')}
                className="h-4 w-4 text-blue-600 rounded border-gray-300"
              />
              <label htmlFor="technitium_skip_tls" className="text-sm font-medium text-gray-700">
                Skip TLS certificate verification
              </label>
              <span className="text-xs text-yellow-600">(use only for self-signed certs)</span>
            </div>

            {technitiumTestStatus && technitiumTestStatus !== 'testing' && (
              <div className={`mb-4 px-3 py-2 rounded text-sm ${technitiumTestStatus.ok ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
                {technitiumTestStatus.ok ? `Connected: ${technitiumTestStatus.message}` : `Error: ${technitiumTestStatus.message}`}
              </div>
            )}

            <button
              onClick={handleTestTechnitium}
              disabled={technitiumTestStatus === 'testing'}
              className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
            >
              {technitiumTestStatus === 'testing' ? 'Testing...' : 'Test Connection'}
            </button>
          </div>

          <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-1 text-gray-900 dark:text-gray-100">DNS Zone Visibility</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
              Control which zones from the DNS provider are shown in the DNS Zones list.
            </p>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Filter Mode</label>
              <select
                value={config.dns_zone_filter_mode || 'allow_all'}
                onChange={e => handleConfigChange('dns_zone_filter_mode', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm dark:bg-gray-700 dark:text-gray-100"
              >
                <option value="allow_all">Allow all except — show every zone; listed zones are hidden</option>
                <option value="block_all">Block all except — hide every zone; only listed zones are shown</option>
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {(config.dns_zone_filter_mode || 'allow_all') === 'allow_all' ? 'Zones to hide' : 'Zones to show'}
              </label>
              <textarea
                rows={4}
                value={config.dns_zone_filter_list || ''}
                onChange={e => handleConfigChange('dns_zone_filter_list', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm font-mono dark:bg-gray-700 dark:text-gray-100"
                placeholder="example.com&#10;internal.lan&#10;10.in-addr.arpa."
              />
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">One zone name per line.</p>
            </div>

            {(config.dns_zone_filter_mode || 'allow_all') === 'block_all' && (
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.dns_zone_filter_auto_allow === 'true'}
                  onChange={e => handleConfigChange('dns_zone_filter_auto_allow', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">
                  Automatically allow new zones — when a zone is found that is not in the list above, add it automatically
                </span>
              </label>
            )}
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">DNS Auto-Sync</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
              Automatically synchronize IP address records in IPAM with A/AAAA records from the configured DNS provider.
            </p>
            <div className="space-y-3">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.dns_auto_add_ips_enabled === 'true'}
                  onChange={e => handleConfigChange('dns_auto_add_ips_enabled', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">
                  Auto-add discovered IPs to matching subnet — when a DNS A/AAAA record is found for an IP not already in IPAM, create the record automatically
                </span>
              </label>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.dns_auto_remove_ips_enabled === 'true'}
                  onChange={e => handleConfigChange('dns_auto_remove_ips_enabled', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">
                  Auto-remove IPs no longer in DNS — remove IPAM records that were added by DNS auto-sync but no longer appear in any DNS zone
                </span>
              </label>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-2">DNS Bulk Check</h2>
            <p className="text-sm text-gray-500 mb-4">
              Run a background check on all IP addresses that have a DNS name assigned, verifying that DNS records are in sync.
            </p>
            {dnsBulkStatus && dnsBulkStatus !== 'running' && (
              <div className={`mb-4 px-3 py-2 rounded text-sm ${dnsBulkStatus.ok ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
                {dnsBulkStatus.message}
              </div>
            )}
            <button
              onClick={handleDnsBulkCheck}
              disabled={dnsBulkStatus === 'running'}
              className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 disabled:opacity-50 transition text-sm font-medium"
            >
              {dnsBulkStatus === 'running' ? 'Starting...' : 'Run DNS Bulk Check'}
            </button>
          </div>

          <button
            onClick={handleSaveConfig}
            disabled={saving}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
  )
}
