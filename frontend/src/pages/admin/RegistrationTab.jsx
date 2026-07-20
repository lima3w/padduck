import { useTranslation } from 'react-i18next'

export default function RegistrationTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  const { t } = useTranslation()
  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">{t('registrationTab.appUrlTitle')}</h2>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('registrationTab.publicUrlLabel')}
              </label>
              <input
                type="url"
                value={config.app_url || ''}
                onChange={(e) => handleConfigChange('app_url', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="http://localhost:3000"
              />
              <p className="text-xs text-gray-500 mt-1">
                {t('registrationTab.publicUrlHintPrefix')}<code>http://ipam.example.com:8080</code>{t('registrationTab.publicUrlHintSuffix')}
              </p>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">{t('registrationTab.registrationSettingsTitle')}</h2>

            <label className="flex items-center gap-3 mb-4 cursor-pointer">
              <input
                type="checkbox"
                checked={config.registration_enabled !== 'false'}
                onChange={(e) => handleConfigChange('registration_enabled', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">
                <strong>{t('registrationTab.enableSelfRegistration')}</strong>
                <span className="block text-gray-500">{t('registrationTab.enableSelfRegistrationHint')}</span>
              </span>
            </label>

            {config.require_email_verification === 'true' && !config.smtp_host && (
              <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded text-yellow-800 text-sm">
                {t('registrationTab.smtpNotConfiguredWarning')}
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
                <strong>{t('registrationTab.requireEmailVerification')}</strong>
                <span className="block text-gray-500">{t('registrationTab.requireEmailVerificationHint')}</span>
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
                <strong>{t('registrationTab.requireAdminApproval')}</strong>
                <span className="block text-gray-500">{t('registrationTab.requireAdminApprovalHint')}</span>
              </span>
            </label>
          </div>

          <button
            onClick={handleSaveConfig}
            disabled={saving}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {saving ? t('common.saving') : t('scanRetention.saveSettings')}
          </button>
        </div>
  )
}
