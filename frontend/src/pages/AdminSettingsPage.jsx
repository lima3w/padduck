import { useState, useEffect } from 'react'
import * as client from '../api/client'
import { useNavigate, Link } from 'react-router-dom'
import { testDnsConnection } from '../api/client'

export default function AdminSettingsPage() {
  const navigate = useNavigate()
  const [config, setConfig] = useState(null)
  const [approvals, setApprovals] = useState([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [purging, setPurging] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [message, setMessage] = useState({ text: '', type: '' })
  const [activeTab, setActiveTab] = useState('registration')
  const [dnsTestStatus, setDnsTestStatus] = useState(null) // null | 'testing' | 'ok' | { error: string }

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [configRes, approvalsRes] = await Promise.all([
        client.getAdminConfig(),
        client.listPendingApprovals(),
      ])
      setConfig(configRes.data.config)
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
      await client.updateAdminConfig(config)
      showMessage('Settings saved successfully')
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

  const tabs = [
    { id: 'registration', label: 'Registration' },
    { id: 'smtp', label: 'SMTP / Email' },
    { id: 'approvals', label: `Approvals${approvals.length > 0 ? ` (${approvals.length})` : ''}` },
    { id: 'audit', label: 'Audit' },
    { id: 'dns', label: 'DNS' },
    { id: 'scanner', label: 'Scanner' },
    { id: 'tools', label: 'Tools' },
  ]

  return (
    <div className="max-w-3xl mx-auto p-6">
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

      <div className="flex gap-1 mb-6 border-b border-gray-200">
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
          </div>

          <div className="flex gap-3">
            <button
              onClick={handleSaveConfig}
              disabled={saving}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
            <button
              onClick={handleTestDns}
              disabled={dnsTestStatus === 'testing'}
              className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
            >
              {dnsTestStatus === 'testing' ? 'Testing...' : 'Test Connection'}
            </button>
          </div>
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
            <div className="flex gap-3 items-center mt-6">
              <button
                onClick={handleSaveConfig}
                disabled={saving}
                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'tools' && (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Subnet Tools</h2>
            <div className="space-y-3">
              <Link
                to="/admin/overlap-report"
                className="flex items-center gap-3 p-3 rounded border hover:bg-gray-50 transition"
              >
                <div>
                  <p className="font-medium text-gray-900">Subnet Overlap Check</p>
                  <p className="text-sm text-gray-500">Find all overlapping subnets across all sections</p>
                </div>
                <span className="ml-auto text-blue-600 text-sm">Open →</span>
              </Link>
            </div>
          </div>
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">IP Tags</h2>
            <div className="space-y-3">
              <Link
                to="/admin/tags"
                className="flex items-center gap-3 p-3 rounded border hover:bg-gray-50 transition"
              >
                <div>
                  <p className="font-medium text-gray-900">Manage IP Tags</p>
                  <p className="text-sm text-gray-500">Create, edit, and delete IP address tags with colour coding</p>
                </div>
                <span className="ml-auto text-blue-600 text-sm">Open →</span>
              </Link>
            </div>
          </div>
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
                      Registered {new Date(approval.created_at).toLocaleDateString()}
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
