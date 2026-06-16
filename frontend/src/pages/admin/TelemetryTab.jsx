import { useState } from 'react'
import { sendTelemetryNow } from '../../api/admin'

export default function TelemetryTab({ config, handleConfigChange, handleSaveConfig, saving, showMessage }) {
  const [sending, setSending] = useState(false)
  const [detailsOpen, setDetailsOpen] = useState(false)

  const handleSendNow = async () => {
    setSending(true)
    try {
      await sendTelemetryNow()
      showMessage('Telemetry snapshot sent successfully')
    } catch (err) {
      showMessage('Send failed: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSending(false)
    }
  }

  const enabled = config.telemetry_enabled === 'true'

  return (
    <div className="space-y-4">
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">Telemetry</h2>
        <p className="text-sm text-gray-500 mb-4">
          Padduck can send periodic usage snapshots to a PocketBase instance you control.
          No data leaves your infrastructure without your configuration — the destination URL and token
          are set below and are never shared with third parties.
        </p>

        <label className="flex items-center gap-3 mb-4 cursor-pointer">
          <input
            type="checkbox"
            checked={enabled}
            onChange={(e) => handleConfigChange('telemetry_enabled', e.target.checked ? 'true' : 'false')}
            className="w-4 h-4 text-blue-600 rounded"
          />
          <span className="text-sm text-gray-700">
            <strong>Enable telemetry</strong>
            <span className="block text-gray-500">Send usage snapshots on the configured schedule.</span>
          </span>
        </label>

        <div className="border border-gray-100 rounded-lg overflow-hidden">
          <button
            type="button"
            onClick={() => setDetailsOpen((o) => !o)}
            className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 text-sm font-medium text-gray-700 transition"
          >
            <span>What is collected?</span>
            <span className="text-gray-400">{detailsOpen ? '▲' : '▼'}</span>
          </button>
          {detailsOpen && (
            <div className="px-4 py-3 text-sm text-gray-600 space-y-2 border-t border-gray-100">
              <ul className="list-disc list-inside space-y-1">
                <li>Object counts: subnets, IP addresses, VLANs, devices, users (no names or IPs)</li>
                <li>Active user counts over 7 and 30 days (count only)</li>
                <li>Subnet utilization percentiles and threshold bucket counts (no addresses or hostnames)</li>
                <li>Feature flag states (enabled/disabled per feature)</li>
                <li>Version and instance identifier</li>
                <li>Locale fields you configure below (UI locale, timezone, country/region codes)</li>
              </ul>
              <p className="text-gray-500 text-xs mt-2">
                No IP addresses, hostnames, user names, or any personally identifiable information is ever included.
              </p>
            </div>
          )}
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-4">PocketBase Destination</h2>

        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 mb-1">PocketBase URL</label>
          <input
            type="url"
            value={config.telemetry_pocketbase_url || ''}
            onChange={(e) => handleConfigChange('telemetry_pocketbase_url', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
            placeholder="https://pb.example.com"
          />
          <p className="text-xs text-gray-500 mt-1">Base URL of your PocketBase instance (no trailing slash).</p>
        </div>

        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 mb-1">Service Token</label>
          <input
            type="password"
            value={config.telemetry_pocketbase_token || ''}
            onChange={(e) => handleConfigChange('telemetry_pocketbase_token', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
            placeholder="••••••••"
          />
          <p className="text-xs text-gray-500 mt-1">Bearer token used to authenticate with the PocketBase REST API.</p>
        </div>

        <div className="mb-2">
          <label className="block text-sm font-medium text-gray-700 mb-1">Snapshot Period</label>
          <select
            value={config.telemetry_snapshot_period || 'daily'}
            onChange={(e) => handleConfigChange('telemetry_snapshot_period', e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
          >
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
          </select>
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">Locale (optional)</h2>
        <p className="text-sm text-gray-500 mb-4">
          These values are included in snapshots to help aggregate data by region. All fields are optional.
        </p>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">UI Locale</label>
            <input
              type="text"
              value={config.telemetry_ui_locale || ''}
              onChange={(e) => handleConfigChange('telemetry_ui_locale', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="en-US"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Timezone Region</label>
            <input
              type="text"
              value={config.telemetry_timezone_region || ''}
              onChange={(e) => handleConfigChange('telemetry_timezone_region', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="America/New_York"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Country Code</label>
            <input
              type="text"
              value={config.telemetry_country_code || ''}
              onChange={(e) => handleConfigChange('telemetry_country_code', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="US"
              maxLength={2}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Region Code</label>
            <input
              type="text"
              value={config.telemetry_region_code || ''}
              onChange={(e) => handleConfigChange('telemetry_region_code', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="NY"
              maxLength={3}
            />
          </div>
        </div>
      </div>

      <div className="flex gap-3 items-center">
        <button
          onClick={handleSaveConfig}
          disabled={saving}
          className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
        >
          {saving ? 'Saving...' : 'Save'}
        </button>
        <button
          onClick={handleSendNow}
          disabled={sending}
          className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
        >
          {sending ? 'Sending...' : 'Send Test Snapshot Now'}
        </button>
      </div>
    </div>
  )
}
