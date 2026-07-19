import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { sendTelemetryNow } from '../../api/admin'

export default function TelemetryTab({ config, handleConfigChange, handleSaveConfig, saving, showMessage }) {
  const { t } = useTranslation()
  const [sending, setSending] = useState(false)
  const [detailsOpen, setDetailsOpen] = useState(false)

  const handleSendNow = async () => {
    setSending(true)
    try {
      await sendTelemetryNow()
      showMessage(t('telemetryTab.sentSuccess'))
    } catch (err) {
      showMessage(t('telemetryTab.sendFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSending(false)
    }
  }

  const enabled = config.telemetry_enabled === 'true'

  return (
    <div className="space-y-4">
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">{t('telemetryTab.title')}</h2>
        <p className="text-sm text-gray-500 mb-4">
          {t('telemetryTab.subtitle')}
        </p>

        <label className="flex items-center gap-3 mb-4 cursor-pointer">
          <input
            type="checkbox"
            checked={enabled}
            onChange={(e) => handleConfigChange('telemetry_enabled', e.target.checked ? 'true' : 'false')}
            className="w-4 h-4 text-blue-600 rounded"
          />
          <span className="text-sm text-gray-700">
            <strong>{t('telemetryTab.enableTelemetry')}</strong>
            <span className="block text-gray-500">{t('telemetryTab.enableTelemetryHint')}</span>
          </span>
        </label>

        <div className="border border-gray-100 rounded-lg overflow-hidden">
          <button
            type="button"
            onClick={() => setDetailsOpen((o) => !o)}
            className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 text-sm font-medium text-gray-700 transition"
          >
            <span>{t('telemetryTab.whatIsCollected')}</span>
            <span className="text-gray-400">{detailsOpen ? '▲' : '▼'}</span>
          </button>
          {detailsOpen && (
            <div className="px-4 py-3 text-sm text-gray-600 space-y-2 border-t border-gray-100">
              <ul className="list-disc list-inside space-y-1">
                <li>{t('telemetryTab.collectedObjectCounts')}</li>
                <li>{t('telemetryTab.collectedActiveUsers')}</li>
                <li>{t('telemetryTab.collectedUtilization')}</li>
                <li>{t('telemetryTab.collectedFeatureFlags')}</li>
                <li>{t('telemetryTab.collectedVersion')}</li>
                <li>{t('telemetryTab.collectedLocale')}</li>
              </ul>
              <p className="text-gray-500 text-xs mt-2">
                {t('telemetryTab.noPiiNotice')}
              </p>
            </div>
          )}
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">{t('telemetryTab.scheduleDeploymentTitle')}</h2>
        <p className="text-sm text-gray-500 mb-4">
          {t('telemetryTab.snapshotsSentToPrefix')}<span className="font-mono text-xs">base.lima3.dev</span>{t('telemetryTab.snapshotsSentToSuffix')}
        </p>

        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 mb-1">{t('telemetryTab.snapshotPeriod')}</label>
          <select
            value={config.telemetry_snapshot_period || 'daily'}
            onChange={(e) => handleConfigChange('telemetry_snapshot_period', e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
          >
            <option value="daily">{t('telemetryTab.daily')}</option>
            <option value="weekly">{t('telemetryTab.weekly')}</option>
          </select>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('telemetryTab.deploymentType')}</label>
            <select
              value={config.telemetry_deployment_type || ''}
              onChange={(e) => handleConfigChange('telemetry_deployment_type', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
            >
              <option value="">{t('telemetryTab.unknown')}</option>
              <option value="docker">{t('telemetryTab.docker')}</option>
              <option value="docker_compose">{t('telemetryTab.dockerCompose')}</option>
              <option value="kubernetes">{t('telemetryTab.kubernetes')}</option>
              <option value="baremetal">{t('telemetryTab.bareMetal')}</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('telemetryTab.deploymentMode')}</label>
            <select
              value={config.telemetry_deployment_mode || ''}
              onChange={(e) => handleConfigChange('telemetry_deployment_mode', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
            >
              <option value="">{t('telemetryTab.unknown')}</option>
              <option value="self_hosted">{t('telemetryTab.selfHosted')}</option>
              <option value="on_prem">{t('telemetryTab.onPrem')}</option>
              <option value="dev">{t('telemetryTab.development')}</option>
              <option value="test">{t('telemetryTab.testStaging')}</option>
            </select>
          </div>
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">{t('telemetryTab.localeTitle')}</h2>
        <p className="text-sm text-gray-500 mb-4">
          {t('telemetryTab.localeSubtitle')}
        </p>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('telemetryTab.uiLocale')}</label>
            <input
              type="text"
              value={config.telemetry_ui_locale || ''}
              onChange={(e) => handleConfigChange('telemetry_ui_locale', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="en-US"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('telemetryTab.timezoneRegion')}</label>
            <input
              type="text"
              value={config.telemetry_timezone_region || ''}
              onChange={(e) => handleConfigChange('telemetry_timezone_region', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="America/New_York"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('telemetryTab.countryCode')}</label>
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
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('telemetryTab.regionCode')}</label>
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
          {saving ? t('common.saving') : t('common.save')}
        </button>
        <button
          onClick={handleSendNow}
          disabled={sending}
          className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
        >
          {sending ? t('telemetryTab.sending') : t('telemetryTab.sendTestSnapshotNow')}
        </button>
      </div>
    </div>
  )
}
