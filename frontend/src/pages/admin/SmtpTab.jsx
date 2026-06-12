import { useState } from 'react'
import { testSMTP } from '../../api/admin'

export default function SmtpTab({ config, handleConfigChange, handleSaveConfig, saving, showMessage }) {
  const [testEmail, setTestEmail] = useState('')

  const handleTestSMTP = async () => {
    if (!testEmail) {
      showMessage('Enter an email address to send test to', 'error')
      return
    }
    try {
      await testSMTP(testEmail)
      showMessage('Test email sent to ' + testEmail)
    } catch (err) {
      showMessage('SMTP test failed: ' + (err.response?.data?.error || err.message), 'error')
    }
  }

  return (
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
  )
}
