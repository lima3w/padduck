export default function RegistrationTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  return (
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
  )
}
