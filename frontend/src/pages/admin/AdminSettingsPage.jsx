import { useState, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { approveUser, getAdminConfig, listPendingApprovals, rejectUser, updateAdminConfig } from '../../api/admin'
import { CONFIG_KEYS_BY_TAB } from './settingsShared'
import RegistrationTab from './RegistrationTab'
import SmtpTab from './SmtpTab'
import ApprovalsTab from './ApprovalsTab'
import AuditTab from './AuditTab'
import AlertsTab from './AlertsTab'
import DnsTab from './DnsTab'
import ScannerTab from './ScannerTab'
import FeaturesTab from './FeaturesTab'
import UpdatesTab from './UpdatesTab'
import NotificationsTab from './NotificationsTab'
import ToolsTab from './ToolsTab'

// Shell: owns the config object, save/messaging, and tab navigation. Each
// tab is its own component under src/pages/admin/ and keeps tab-specific
// state and handlers to itself.
export default function AdminSettingsPage() {
  const location = useLocation()
  const [config, setConfig] = useState(null)
  const [approvals, setApprovals] = useState([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState({ text: '', type: '' })
  const initialTab = new URLSearchParams(location.search).get('tab') || 'registration'
  const [activeTab, setActiveTab] = useState(initialTab)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [configRes, approvalsRes] = await Promise.all([
        getAdminConfig(),
        listPendingApprovals(),
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
      await updateAdminConfig(updates)
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

  const handleApprove = async (id) => {
    try {
      await approveUser(id)
      showMessage('User approved')
      setApprovals((prev) => prev.filter((a) => a.id !== id))
    } catch (err) {
      showMessage('Failed to approve: ' + (err.response?.data?.error || err.message), 'error')
    }
  }

  const handleReject = async (id) => {
    const reason = window.prompt('Rejection reason (optional):') ?? ''
    try {
      await rejectUser(id, reason)
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

  const configProps = { config, handleConfigChange, handleSaveConfig, saving }

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

      {activeTab === 'registration' && config && <RegistrationTab {...configProps} />}
      {activeTab === 'smtp' && config && <SmtpTab {...configProps} showMessage={showMessage} />}
      {activeTab === 'approvals' && <ApprovalsTab approvals={approvals} handleApprove={handleApprove} handleReject={handleReject} />}
      {activeTab === 'audit' && config && <AuditTab {...configProps} showMessage={showMessage} />}
      {activeTab === 'alerts' && config && <AlertsTab {...configProps} />}
      {activeTab === 'dns' && config && <DnsTab {...configProps} />}
      {activeTab === 'scanner' && config && <ScannerTab {...configProps} />}
      {activeTab === 'features' && config && <FeaturesTab {...configProps} />}
      {activeTab === 'updates' && config && <UpdatesTab {...configProps} />}
      {activeTab === 'notifications' && <NotificationsTab />}
      {activeTab === 'tools' && <ToolsTab config={config} />}
    </div>
  )
}
