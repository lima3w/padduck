import { useState, useEffect } from 'react'
import * as client from '../api/client'
import { useNavigate, useLocation, Link } from 'react-router-dom'
import { testDnsConnection, testTechnitiumConnection } from '../api/client'

const CONFIG_KEYS_BY_TAB = {
  registration: [
    'app_url',
    'registration_enabled',
    'require_email_verification',
    'require_admin_approval',
  ],
  smtp: [
    'smtp_host',
    'smtp_port',
    'smtp_username',
    'smtp_password',
    'smtp_from',
    'smtp_tls',
  ],
  audit: ['audit_log_retention_days'],
  alerts: ['default_alert_threshold_pct'],
  dns: [
    'pdns_enabled',
    'pdns_api_url',
    'pdns_api_key',
    'pdns_default_zone',
    'pdns_ptr_zones',
    'technitium_url',
    'technitium_token',
    'technitium_default_zone',
    'technitium_skip_tls',
    'dns_zone_filter_mode',
    'dns_zone_filter_list',
    'dns_zone_filter_auto_allow',
    'dns_auto_add_ips_enabled',
    'dns_auto_remove_ips_enabled',
  ],
  scanner: [
    'scanner_resolve_hostnames',
    'scanner_snmp_community',
    'scanner_snmp_version',
    'scanner_port_scan_enabled',
    'scanner_port_list',
  ],
  features: [
    'feature_customers_enabled',
    'feature_vlans_enabled',
    'feature_vrfs_enabled',
    'feature_racks_enabled',
    'feature_locations_enabled',
    'feature_bgp_enabled',
    'feature_devices_enabled',
    'feature_nat_enabled',
    'feature_firewall_enabled',
    'feature_dhcp_enabled',
    'feature_circuits_enabled',
    'anonymous_api_enabled',
  ],
  updates: [
    'update_check_enabled',
  ],
}

const FEATURE_TOGGLES = [
  {
    key: 'feature_customers_enabled',
    title: 'Customers',
    description: 'Customer records and customer navigation.',
  },
  {
    key: 'feature_vlans_enabled',
    title: 'VLANs',
    description: 'VLANs, VLAN domains, VLAN groups, and VLAN usage reports.',
  },
  {
    key: 'feature_vrfs_enabled',
    title: 'VRFs',
    description: 'VRF records and VRF navigation.',
  },
  {
    key: 'feature_racks_enabled',
    title: 'Racks',
    description: 'Rack records, rack details, and rack device lists.',
  },
  {
    key: 'feature_locations_enabled',
    title: 'Locations',
    description: 'Location records, location hierarchy, and location details.',
  },
  {
    key: 'feature_bgp_enabled',
    title: 'BGP / AS Numbers',
    description: 'Autonomous system records and BGP navigation.',
  },
  {
    key: 'feature_devices_enabled',
    title: 'Devices',
    description: 'Device inventory, device types, interfaces, and device IP associations.',
  },
  {
    key: 'feature_nat_enabled',
    title: 'NAT Rules',
    description: 'NAT mapping records and NAT navigation.',
  },
  {
    key: 'feature_firewall_enabled',
    title: 'Firewall Zones',
    description: 'Firewall zone records, zone mappings, and firewall navigation.',
  },
  {
    key: 'feature_dhcp_enabled',
    title: 'DHCP',
    description: 'DHCP server and lease records.',
  },
  {
    key: 'feature_circuits_enabled',
    title: 'Circuits',
    description: 'Circuit providers, physical circuits, and logical circuits.',
  },
  {
    key: 'anonymous_api_enabled',
    title: 'Anonymous API Access',
    description: 'Allow unauthenticated read-only IP queries via GET /api/v1/query/ip. Additive — no existing behavior changes.',
  },
]

export default function AdminSettingsPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const [config, setConfig] = useState(null)
  const [approvals, setApprovals] = useState([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [purging, setPurging] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [message, setMessage] = useState({ text: '', type: '' })
  const initialTab = new URLSearchParams(location.search).get('tab') || 'registration'
  const [activeTab, setActiveTab] = useState(initialTab)
  const [dnsTestStatus, setDnsTestStatus] = useState(null) // null | 'testing' | 'ok' | { error: string }
  const [technitiumTestStatus, setTechnitiumTestStatus] = useState(null) // null | 'testing' | { ok, message }
  const [dnsBulkStatus, setDnsBulkStatus] = useState(null) // null | 'running' | { ok, message }
  const [notifStats, setNotifStats] = useState(null)
  const [notifStatsLoading, setNotifStatsLoading] = useState(false)
  const [updateStatus, setUpdateStatus] = useState(null)
  const [checkingUpdates, setCheckingUpdates] = useState(false)
  const [showSnmpCommunity, setShowSnmpCommunity] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  useEffect(() => {
    if (activeTab === 'notifications') loadNotifStats()
  }, [activeTab])

  const loadData = async () => {
    setLoading(true)
    try {
      const [configRes, approvalsRes] = await Promise.all([
        client.getAdminConfig(),
        client.listPendingApprovals(),
      ])
      const loadedConfig = configRes.data.config || {}
      const featureDefaults = {}
      CONFIG_KEYS_BY_TAB.features.forEach(key => {
        if (!Object.prototype.hasOwnProperty.call(loadedConfig, key)) {
          featureDefaults[key] = 'true'
        }
      })
      setConfig({ ...loadedConfig, ...featureDefaults })
      setApprovals(approvalsRes.data.approvals || [])
    } catch (err) {
      showMessage('Failed to load settings: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setLoading(false)
    }
  }

  const showMessage = (text, type = 'success') => {
    setMessage({ text, type })
    setTimeout(() => setMessage({ text: '', type: '' }), 4000)
  }

  const handleConfigChange = (key, value) => {
    setConfig((prev) => ({ ...prev, [key]: value }))
  }

  const handleSaveConfig = async () => {
    setSaving(true)
    try {
      const keys = CONFIG_KEYS_BY_TAB[activeTab] || []
      const updates = Object.fromEntries(
        keys
          .filter((key) => Object.prototype.hasOwnProperty.call(config, key))
          .map((key) => [key, config[key]])
      )
      await client.updateAdminConfig(updates)
      showMessage('Settings saved successfully')
      if (activeTab === 'features') {
        window.setTimeout(() => {
          window.location.href = window.location.pathname + '?tab=features'
        }, 250)
      }
    } catch (err) {
      showMessage('Failed to save: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleTestSMTP = async () => {
    if (!testEmail) {
      showMessage('Enter an email address to send test to', 'error')
      return
    }
    try {
      await client.testSMTP(testEmail)
      showMessage('Test email sent to ' + testEmail)
    } catch (err) {
      showMessage('SMTP test failed: ' + (err.response?.data?.error || err.message), 'error')
    }
  }

  const handleApprove = async (id) => {
    try {
      await client.approveUser(id)
      showMessage('User approved')
      setApprovals((prev) => prev.filter((a) => a.id !== id))
    } catch (err) {
      showMessage('Failed to approve: ' + (err.response?.data?.error || err.message), 'error')
    }
  }

  const handleReject = async (id) => {
    const reason = window.prompt('Rejection reason (optional):') ?? ''
    try {
      await client.rejectUser(id, reason)
      showMessage('User rejected')
      setApprovals((prev) => prev.filter((a) => a.id !== id))
    } catch (err) {
      showMessage('Failed to reject: ' + (err.response?.data?.error || err.message), 'error')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500">
        Loading settings...
      </div>
    )
  }

  const handlePurgeAuditLogs = async () => {
    if (!window.confirm('Delete all audit log entries older than the configured retention period?')) return
    setPurging(true)
    try {
      const res = await client.purgeAuditLogs()
      showMessage(res.data.message || 'Audit logs purged')
    } catch (err) {
      showMessage('Purge failed: ' + (err.response?.data?.error || err.message), 'error')
    } finally {
      setPurging(false)
    }
  }

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
      await client.checkAllDns()
      setDnsBulkStatus({ ok: true, message: 'DNS bulk check started in background' })
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Failed to start DNS check'
      setDnsBulkStatus({ ok: false, message: msg })
    }
  }

  const loadNotifStats = async () => {
    setNotifStatsLoading(true)
    try {
      const res = await client.getNotificationStats()
      setNotifStats(res.data)
    } catch {
      setNotifStats(null)
    } finally {
      setNotifStatsLoading(false)
    }
  }

  const handleUpdateCheck = async () => {
    setCheckingUpdates(true)
    setUpdateStatus(null)
    try {
      const res = await client.checkForUpdates()
      setUpdateStatus({ ok: true, data: res.data })
    } catch (err) {
      setUpdateStatus({
        ok: false,
        message: err.response?.data?.error || err.message || 'Update check failed',
      })
    } finally {
      setCheckingUpdates(false)
    }
  }

  const tabs = [
    { id: 'registration', label: 'Registration' },
    { id: 'smtp', label: 'SMTP / Email' },
    { id: 'approvals', label: `Approvals${approvals.length > 0 ? ` (${approvals.length})` : ''}` },
    { id: 'audit', label: 'Audit' },
    { id: 'alerts', label: 'Alerts' },
    { id: 'dns', label: 'DNS' },
    { id: 'scanner', label: 'Scanner' },
    { id: 'features', label: 'Features' },
    { id: 'updates', label: 'Updates' },
    { id: 'notifications', label: 'Notifications' },
    { id: 'tools', label: 'Tools' },
  ]

  const featureEnabled = (key) => config?.[key] !== 'false'
  const toolSections = [
    {
      title: 'Data Tools',
      links: [
        { to: '/admin/overlap-report', title: 'Subnet Overlap Check', description: 'Find overlapping subnets across all networks' },
        { to: '/admin/import', title: 'Data Import', description: 'Import subnets, IP addresses, or phpIPAM data' },
        { to: '/admin/export', title: 'Data Export', description: 'Export a full data backup' },
      ],
    },
    {
      title: 'Schema & Taxonomy',
      links: [
        { to: '/admin/custom-fields', title: 'Custom Fields', description: 'Manage extra fields for subnets, IPs, and devices' },
        { to: '/admin/tags', title: 'IP Tags', description: 'Create and manage IP address tags' },
        { to: '/admin/vlan-domains', title: 'VLAN Domains', description: 'Manage VLAN namespace boundaries', visible: featureEnabled('feature_vlans_enabled') },
        { to: '/admin/vlan-groups', title: 'VLAN Groups', description: 'Group VLANs for organization and reporting', visible: featureEnabled('feature_vlans_enabled') },
        { to: '/admin/vlans/usage-report', title: 'VLAN Usage', description: 'Review VLAN allocation and utilization', visible: featureEnabled('feature_vlans_enabled') },
      ],
    },
    {
      title: 'Discovery & Automation',
      links: [
        { to: '/admin/webhooks', title: 'Webhooks', description: 'Configure outbound event delivery' },
        { to: '/admin/integrations', title: 'Integrations', description: 'Integration setup notes and connection checks' },
        { to: '/admin/grafana', title: 'Grafana', description: 'Configure the Grafana datasource integration' },
      ],
    },
    {
      title: 'Reports',
      links: [
        { to: '/admin/reports/scheduled', title: 'Scheduled Reports', description: 'Manage recurring emailed reports' },
      ],
    },
    {
      title: 'Authentication',
      links: [
        { to: '/admin/auth/ldap', title: 'LDAP / AD', description: 'Configure LDAP authentication and group mappings' },
        { to: '/admin/auth/oauth2', title: 'OAuth2 / OIDC', description: 'Configure OAuth2 or OpenID Connect login' },
        { to: '/admin/auth/saml', title: 'SAML SSO', description: 'Configure SAML single sign-on' },
        { to: '/admin/identity-policies', title: 'Identity Policies', description: 'Manage IP-based access control and identity rules' },
      ],
    },
  ]
    .map((network) => ({
      ...network,
      links: network.links.filter((link) => link.visible !== false),
    }))
    .filter((network) => network.links.length > 0)

  return (
    <div className="w-full max-w-7xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Admin Settings</h1>

      {message.text && (
        <div
          className={`mb-4 p-4 rounded text-sm ${
            message.type === 'error'
              ? 'bg-red-50 border border-red-200 text-red-700'
              : 'bg-green-50 border border-green-200 text-green-700'
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="flex flex-wrap gap-1 mb-6 border-b border-gray-200">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm font-medium rounded-t transition ${
              activeTab === tab.id
                ? 'bg-white border border-b-white border-gray-200 text-blue-600 -mb-px'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'registration' && config && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Application URL</h2>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Public URL
              </label>
              <input
                type="url"
                value={config.app_url || ''}
                onChange={(e) => handleConfigChange('app_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="http://localhost:3000"
              />
              <p className="text-xs text-gray-500 mt-1">
                Used in verification and notification emails. Include scheme and port if non-standard (e.g. <code>http://ipam.example.com:8080</code>).
              </p>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Registration Settings</h2>

            <label className="flex items-center gap-3 mb-4 cursor-pointer">
              <input
                type="checkbox"
                checked={config.registration_enabled !== 'false'}
                onChange={(e) => handleConfigChange('registration_enabled', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">
                <strong>Enable self-registration</strong>
                <span className="block text-gray-500">Allow anyone to create an account</span>
              </span>
            </label>

            {config.require_email_verification === 'true' && !config.smtp_host && (
              <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded text-yellow-800 text-sm">
                Email verification is enabled but SMTP is not configured — verification emails will not be sent and new users will be stuck. Configure SMTP on the Email tab first.
              </div>
            )}

            <label className="flex items-center gap-3 mb-4 cursor-pointer">
              <input
                type="checkbox"
                checked={config.require_email_verification === 'true'}
                onChange={(e) => handleConfigChange('require_email_verification', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">
                <strong>Require email verification</strong>
                <span className="block text-gray-500">Users must verify their email before logging in</span>
              </span>
            </label>

            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={config.require_admin_approval === 'true'}
                onChange={(e) => handleConfigChange('require_admin_approval', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">
                <strong>Require admin approval</strong>
                <span className="block text-gray-500">New accounts must be approved by an admin</span>
              </span>
            </label>
          </div>

          <button
            onClick={handleSaveConfig}
            disabled={saving}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      )}

      {activeTab === 'smtp' && config && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">SMTP Configuration</h2>

            <div className="grid grid-cols-2 gap-4 mb-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Host</label>
                <input
                  type="text"
                  value={config.smtp_host || ''}
                  onChange={(e) => handleConfigChange('smtp_host', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  placeholder="smtp.example.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
                <input
                  type="number"
                  value={config.smtp_port || '587'}
                  onChange={(e) => handleConfigChange('smtp_port', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  placeholder="587"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4 mb-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
                <input
                  type="text"
                  value={config.smtp_username || ''}
                  onChange={(e) => handleConfigChange('smtp_username', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  placeholder="user@example.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
                <input
                  type="password"
                  value={config.smtp_password || ''}
                  onChange={(e) => handleConfigChange('smtp_password', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  placeholder="••••••••"
                />
              </div>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">From Address</label>
              <input
                type="email"
                value={config.smtp_from || ''}
                onChange={(e) => handleConfigChange('smtp_from', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="noreply@example.com"
              />
            </div>

            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={config.smtp_tls !== 'false'}
                onChange={(e) => handleConfigChange('smtp_tls', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">Use TLS</span>
            </label>
          </div>

          <div className="flex gap-3 items-center">
            <button
              onClick={handleSaveConfig}
              disabled={saving}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
            <input
              type="email"
              value={testEmail}
              onChange={(e) => setTestEmail(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded text-sm"
              placeholder="test@example.com"
            />
            <button
              onClick={handleTestSMTP}
              className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 transition text-sm font-medium"
            >
              Send Test Email
            </button>
          </div>
        </div>
      )}

      {activeTab === 'audit' && config && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Audit Log Retention</h2>
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Retention Period (days)
              </label>
              <input
                type="number"
                min="1"
                max="3650"
                value={config.audit_log_retention_days || '90'}
                onChange={(e) => handleConfigChange('audit_log_retention_days', e.target.value)}
                className="w-32 px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              />
              <p className="text-xs text-gray-500 mt-1">
                Audit logs older than this many days will be deleted when a purge is run. Default: 90 days.
              </p>
            </div>
            <div className="flex gap-3 items-center">
              <button
                onClick={handleSaveConfig}
                disabled={saving}
                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-2">Purge Old Logs</h2>
            <p className="text-sm text-gray-600 mb-4">
              Permanently delete audit log entries older than the configured retention period.
            </p>
            <div className="flex gap-3">
              <button
                onClick={handlePurgeAuditLogs}
                disabled={purging}
                className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 disabled:opacity-50 transition text-sm font-medium"
              >
                {purging ? 'Purging...' : 'Purge Old Logs'}
              </button>
              <button
                onClick={() => navigate('/admin/audit-log')}
                className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 transition text-sm font-medium"
              >
                View Audit Log
              </button>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'alerts' && config && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Utilisation Alerts</h2>
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Default Alert Threshold (%)
              </label>
              <input
                type="number"
                min="1"
                max="100"
                value={config.default_alert_threshold_pct || ''}
                onChange={(e) => handleConfigChange('default_alert_threshold_pct', e.target.value)}
                className="w-32 px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="80"
              />
              <p className="text-xs text-gray-500 mt-1">
                Send an alert when a subnet&apos;s utilisation exceeds this percentage. Individual subnets can override this value.
              </p>
            </div>
            <button
              onClick={handleSaveConfig}
              disabled={saving}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
      )}

      {activeTab === 'dns' && config && (
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
      )}

      {activeTab === 'scanner' && config && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Discovery Scanner</h2>
            <div className="space-y-4">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.scanner_resolve_hostnames !== 'false'}
                  onChange={(e) =>
                    handleConfigChange('scanner_resolve_hostnames', e.target.checked ? 'true' : 'false')
                  }
                  className="w-4 h-4 text-blue-600"
                />
                <div>
                  <span className="font-medium text-gray-900">Resolve hostnames via reverse DNS</span>
                  <p className="text-xs text-gray-500 mt-0.5">
                    After a successful ping, the scanner performs a PTR lookup and stores the result in the
                    IP address record. A forward lookup is also done to flag mismatches. Adds up to 2 s per
                    alive host. Default: enabled.
                  </p>
                </div>
              </label>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-base font-semibold mb-4">SNMP</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Global community string</label>
                <div className="relative">
                  <input
                    type={showSnmpCommunity ? 'text' : 'password'}
                    value={config.scanner_snmp_community ?? ''}
                    onChange={(e) => handleConfigChange('scanner_snmp_community', e.target.value)}
                    placeholder="public"
                    className="w-full px-3 py-2 pr-10 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => setShowSnmpCommunity(v => !v)}
                    className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-400 hover:text-gray-600"
                    aria-label={showSnmpCommunity ? 'Hide community string' : 'Show community string'}
                  >
                    {showSnmpCommunity ? (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 4.411m0 0L21 21" /></svg>
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                    )}
                  </button>
                </div>
                <p className="text-xs text-gray-500 mt-1">Used when no per-device community is configured. Default: public.</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">SNMP version</label>
                <select
                  value={config.scanner_snmp_version ?? '2c'}
                  onChange={(e) => handleConfigChange('scanner_snmp_version', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                >
                  <option value="2c">v2c</option>
                  <option value="3">v3</option>
                </select>
                <p className="text-xs text-gray-500 mt-1">Global default version. Per-device credentials override this.</p>
              </div>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-base font-semibold mb-4">Port Scanning</h3>
            <div className="space-y-4">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.scanner_port_scan_enabled === 'true'}
                  onChange={(e) =>
                    handleConfigChange('scanner_port_scan_enabled', e.target.checked ? 'true' : 'false')
                  }
                  className="w-4 h-4 text-blue-600"
                />
                <div>
                  <span className="font-medium text-gray-900">Enable TCP port scanning</span>
                  <p className="text-xs text-gray-500 mt-0.5">
                    After a successful ping, probe the ports listed below on each alive host. Default: disabled.
                  </p>
                </div>
              </label>
              {config.scanner_port_scan_enabled === 'true' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Port list</label>
                  <input
                    type="text"
                    value={config.scanner_port_list ?? ''}
                    onChange={(e) => handleConfigChange('scanner_port_list', e.target.value)}
                    placeholder="22,80,443,3306,5432,8080,8443"
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  />
                  <p className="text-xs text-gray-500 mt-1">Comma-separated port numbers. Default: 22,80,443,3306,5432,8080,8443.</p>
                </div>
              )}
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
          </div>
        </div>
      )}

      {activeTab === 'tools' && (
        <div className="space-y-4">
          {toolSections.map((network) => (
            <div key={network.title} className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{network.title}</h2>
              <div className="space-y-3">
                {network.links.map((link) => (
                  <Link
                    key={link.to}
                    to={link.to}
                    className="flex items-center gap-3 p-3 rounded border border-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 transition"
                  >
                    <div>
                      <p className="font-medium text-gray-900 dark:text-gray-100">{link.title}</p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">{link.description}</p>
                    </div>
                    <span className="ml-auto shrink-0 text-blue-600 dark:text-blue-400 text-sm">Open →</span>
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {activeTab === 'features' && config && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-2">Enabled Modules</h2>
            <p className="text-sm text-gray-600 mb-5">
              Disabled modules are removed from navigation and their API routes reject direct access.
            </p>
            <div className="grid gap-4 md:grid-cols-2">
              {FEATURE_TOGGLES.map((feature) => (
                <label
                  key={feature.key}
                  className="flex items-start gap-3 rounded border border-gray-200 dark:border-gray-700 p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50"
                >
                  <input
                    type="checkbox"
                    checked={config[feature.key] !== 'false'}
                    onChange={(e) => handleConfigChange(feature.key, e.target.checked ? 'true' : 'false')}
                    className="mt-1 h-4 w-4 rounded text-blue-600"
                  />
                  <span>
                    <span className="block font-medium text-gray-900 dark:text-gray-100">{feature.title}</span>
                    <span className="block text-sm text-gray-500 dark:text-gray-400">{feature.description}</span>
                  </span>
                </label>
              ))}
            </div>
          </div>

          <button
            onClick={handleSaveConfig}
            disabled={saving}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      )}

      {activeTab === 'updates' && config && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Update Check</h2>

            <label className="flex items-center gap-3 mb-4 cursor-pointer">
              <input
                type="checkbox"
                checked={config.update_check_enabled === 'true'}
                onChange={(e) => handleConfigChange('update_check_enabled', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">
                <strong>Enable update checks</strong>
                <span className="block text-gray-500">Checks the GitHub releases API for new versions of Padduck.</span>
              </span>
            </label>
          </div>

          {updateStatus && (
            <div
              className={`rounded border p-4 text-sm ${
                updateStatus.ok
                  ? updateStatus.data?.updateAvailable
                    ? 'bg-yellow-50 border-yellow-200 text-yellow-800'
                    : 'bg-green-50 border-green-200 text-green-700'
                  : 'bg-red-50 border-red-200 text-red-700'
              }`}
            >
              {updateStatus.ok ? (
                <div className="space-y-1">
                  <p className="font-medium">
                    {updateStatus.data?.enabled === false
                      ? 'Update checks are disabled.'
                      : updateStatus.data?.updateAvailable
                      ? `Update available: ${updateStatus.data.latestVersion}`
                      : 'No update available.'}
                  </p>
                  <p>
                    Current: {updateStatus.data?.currentVersion || 'unknown'}
                    {updateStatus.data?.latestVersion ? ` · Latest: ${updateStatus.data.latestVersion}` : ''}
                  </p>
                  {updateStatus.data?.releaseUrl && (
                    <a
                      href={updateStatus.data.releaseUrl}
                      className="inline-block text-blue-600 hover:underline"
                      target="_blank"
                      rel="noreferrer"
                    >
                      View release
                    </a>
                  )}
                </div>
              ) : (
                <p>{updateStatus.message}</p>
              )}
            </div>
          )}

          <div className="flex gap-3">
            <button
              onClick={handleSaveConfig}
              disabled={saving}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
            <button
              onClick={handleUpdateCheck}
              disabled={checkingUpdates}
              className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
            >
              {checkingUpdates ? 'Checking...' : 'Check Now'}
            </button>
          </div>
        </div>
      )}

      {activeTab === 'notifications' && (
        <div className="space-y-6">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-1">Notification Stats</h2>
            <p className="text-sm text-gray-600 mb-4">
              Counts of notification emails sent by type. Users can control their preferences under Account Settings.
            </p>
          </div>

          {notifStatsLoading ? (
            <p className="text-sm text-gray-500">Loading…</p>
          ) : notifStats === null ? (
            <p className="text-sm text-red-500">Failed to load notification stats.</p>
          ) : Object.keys(notifStats).length === 0 ? (
            <p className="text-sm text-gray-500">No notifications have been sent yet.</p>
          ) : (
            <div className="border border-gray-200 rounded divide-y divide-gray-100">
              {Object.entries(notifStats).map(([key, count]) => (
                <div key={key} className="flex items-center justify-between px-4 py-3">
                  <span className="text-sm text-gray-700">
                    {key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())}
                  </span>
                  <span className="text-sm font-medium text-gray-900">{count.toLocaleString()}</span>
                </div>
              ))}
            </div>
          )}

          <button
            onClick={loadNotifStats}
            disabled={notifStatsLoading}
            className="text-sm text-blue-600 hover:underline disabled:opacity-50"
          >
            Refresh
          </button>
        </div>
      )}

      {activeTab === 'approvals' && (
        <div>
          {approvals.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              No pending approvals
            </div>
          ) : (
            <div className="space-y-3">
              {approvals.map((approval) => (
                <div
                  key={approval.id}
                  className="bg-white border border-gray-200 rounded-lg p-4 flex items-center justify-between"
                >
                  <div>
                    <p className="font-medium text-gray-900">{approval.username}</p>
                    <p className="text-sm text-gray-500">{approval.email}</p>
                    <p className="text-xs text-gray-400">
                      Registered {new Date(approval.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleApprove(approval.id)}
                      className="bg-green-600 text-white px-4 py-1.5 rounded text-sm hover:bg-green-700 transition font-medium"
                    >
                      Approve
                    </button>
                    <button
                      onClick={() => handleReject(approval.id)}
                      className="bg-red-600 text-white px-4 py-1.5 rounded text-sm hover:bg-red-700 transition font-medium"
                    >
                      Reject
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
