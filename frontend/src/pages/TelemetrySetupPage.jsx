import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { updateAdminConfig } from '../api/admin'

export default function TelemetrySetupPage() {
  const navigate = useNavigate()
  const [saving, setSaving] = useState(false)
  const [detailsOpen, setDetailsOpen] = useState(false)

  async function handleChoice(enabled) {
    setSaving(true)
    try {
      await updateAdminConfig({ telemetry_enabled: enabled ? 'true' : 'false' })
    } catch {
      // If the save fails (e.g. non-admin), just navigate away silently.
    }
    navigate('/dashboard', { replace: true })
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-lg max-w-lg w-full p-8">

        <div className="mb-6 text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Help Make Padduck Better</h1>
          <p className="text-gray-600 text-sm leading-relaxed">
            You can opt in to share anonymous usage statistics. This is entirely
            optional — Padduck works the same either way.
          </p>
        </div>

        <div className="bg-blue-50 border border-blue-100 rounded-lg p-4 mb-5 space-y-2">
          <p className="text-sm font-semibold text-blue-900">Our commitment to you</p>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>✓ <strong>Never used for marketing or sales</strong> — data is only used to prioritize features and fix bugs</li>
            <li>✓ <strong>Never sold or shared</strong> — data stays between you and the Padduck project, full stop</li>
            <li>✓ <strong>Completely anonymous</strong> — no IP addresses, hostnames, user names, or network identifiers of any kind</li>
            <li>✓ <strong>You can turn it off at any time</strong> — the toggle is in Admin Settings → Telemetry</li>
          </ul>
        </div>

        <div className="border border-gray-200 rounded-lg overflow-hidden mb-6">
          <button
            type="button"
            onClick={() => setDetailsOpen(o => !o)}
            className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 text-sm font-medium text-gray-700 transition"
          >
            <span>What exactly is collected?</span>
            <span className="text-gray-400 text-xs">{detailsOpen ? '▲ hide' : '▼ show'}</span>
          </button>
          {detailsOpen && (
            <div className="px-4 py-4 text-sm text-gray-600 border-t border-gray-100 space-y-3">
              <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide">Counts (numbers only — no names)</p>
              <ul className="list-disc list-inside space-y-1 text-gray-700">
                <li>Total users, active users in the last 7 and 30 days</li>
                <li>Total subnets (IPv4 / IPv6 split), VLANs, customers, locations</li>
                <li>Subnet size distribution across five prefix-length buckets</li>
              </ul>
              <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide mt-3">Utilization statistics</p>
              <ul className="list-disc list-inside space-y-1 text-gray-700">
                <li>Average, median, 75th, 90th, and 95th percentile subnet utilization</li>
                <li>Counts of empty, half-full, near-full, and fully-used subnets</li>
              </ul>
              <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide mt-3">Feature flags</p>
              <ul className="list-disc list-inside space-y-1 text-gray-700">
                <li>Which optional features are enabled (e.g. VLANs, Devices, DNS integration) — on/off only</li>
                <li>Whether LDAP, OIDC, or SAML authentication is configured</li>
              </ul>
              <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide mt-3">Deployment metadata</p>
              <ul className="list-disc list-inside space-y-1 text-gray-700">
                <li>App version, a random anonymous install ID (never tied to your server or network)</li>
                <li>Deployment type and mode if you choose to fill them in (e.g. Docker Compose, Self-Hosted)</li>
                <li>Optional locale fields if configured (UI locale, timezone region, country/region code)</li>
              </ul>
              <p className="text-xs text-gray-400 mt-3 italic">
                No subnet CIDRs, IP addresses, MAC addresses, hostnames, usernames, emails, or any names from your data are ever included.
              </p>
            </div>
          )}
        </div>

        <div className="flex flex-col gap-3">
          <button
            onClick={() => handleChoice(true)}
            disabled={saving}
            className="w-full bg-blue-600 text-white py-3 rounded-lg font-semibold hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {saving ? 'Saving...' : 'Enable Telemetry'}
          </button>
          <button
            onClick={() => handleChoice(false)}
            disabled={saving}
            className="w-full bg-white text-gray-600 py-3 rounded-lg font-medium border border-gray-300 hover:bg-gray-50 disabled:opacity-50 transition"
          >
            No Thanks
          </button>
        </div>

        <p className="text-xs text-center text-gray-400 mt-5">
          You can change this decision at any time in{' '}
          <a href="/admin/settings?tab=telemetry" className="underline hover:text-gray-600">
            Admin Settings → Telemetry
          </a>
          .
        </p>
      </div>
    </div>
  )
}
