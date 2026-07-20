import { useTranslation } from 'react-i18next'

export default function AlertsTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  const { t } = useTranslation()
  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">{t('alertsTab.title')}</h2>
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('alertsTab.thresholdLabel')}
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
                {t('alertsTab.thresholdHint')}
              </p>
            </div>
            <button
              onClick={handleSaveConfig}
              disabled={saving}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {saving ? t('common.saving') : t('common.save')}
            </button>
          </div>
        </div>
  )
}
