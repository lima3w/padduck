import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import ChoiceButtons from './ChoiceButtons'

// First-run wizard step: anonymous telemetry opt-in. Controlled: the wizard
// owns the value ('true' | 'false' | undefined) and saves it on Next.
export default function TelemetryStep({ values, onChange }) {
  const { t } = useTranslation()
  const [detailsOpen, setDetailsOpen] = useState(false)

  return (
    <div>
      <div className="mb-6 text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">{t('telemetrySetup.title')}</h2>
        <p className="text-gray-600 text-sm leading-relaxed">{t('telemetrySetup.subtitle')}</p>
      </div>

      <div className="bg-blue-50 border border-blue-100 rounded-lg p-4 mb-5 space-y-2">
        <p className="text-sm font-semibold text-blue-900">{t('telemetrySetup.commitmentTitle')}</p>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>✓ <strong>{t('telemetrySetup.commitment1Bold')}</strong>{t('telemetrySetup.commitment1Rest')}</li>
          <li>✓ <strong>{t('telemetrySetup.commitment2Bold')}</strong>{t('telemetrySetup.commitment2Rest')}</li>
          <li>✓ <strong>{t('telemetrySetup.commitment3Bold')}</strong>{t('telemetrySetup.commitment3Rest')}</li>
          <li>✓ <strong>{t('telemetrySetup.commitment4Bold')}</strong>{t('telemetrySetup.commitment4Rest')}</li>
        </ul>
      </div>

      <div className="border border-gray-200 rounded-lg overflow-hidden mb-6">
        <button
          type="button"
          onClick={() => setDetailsOpen(o => !o)}
          aria-expanded={detailsOpen}
          className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 text-sm font-medium text-gray-700 transition"
        >
          <span>{t('telemetrySetup.whatIsCollected')}</span>
          <span className="text-gray-400 text-xs">{detailsOpen ? `▲ ${t('telemetrySetup.hide')}` : `▼ ${t('telemetrySetup.show')}`}</span>
        </button>
        {detailsOpen && (
          <div className="px-4 py-4 text-sm text-gray-600 border-t border-gray-100 space-y-3">
            <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide">{t('telemetrySetup.countsTitle')}</p>
            <ul className="list-disc list-inside space-y-1 text-gray-700">
              <li>{t('telemetrySetup.countsItem1')}</li>
              <li>{t('telemetrySetup.countsItem2')}</li>
              <li>{t('telemetrySetup.countsItem3')}</li>
            </ul>
            <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide mt-3">{t('telemetrySetup.utilizationTitle')}</p>
            <ul className="list-disc list-inside space-y-1 text-gray-700">
              <li>{t('telemetrySetup.utilizationItem1')}</li>
              <li>{t('telemetrySetup.utilizationItem2')}</li>
            </ul>
            <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide mt-3">{t('telemetrySetup.featureFlagsTitle')}</p>
            <ul className="list-disc list-inside space-y-1 text-gray-700">
              <li>{t('telemetrySetup.featureFlagsItem1')}</li>
              <li>{t('telemetrySetup.featureFlagsItem2')}</li>
            </ul>
            <p className="text-xs text-gray-500 uppercase font-semibold tracking-wide mt-3">{t('telemetrySetup.deploymentMetadataTitle')}</p>
            <ul className="list-disc list-inside space-y-1 text-gray-700">
              <li>{t('telemetrySetup.deploymentItem1')}</li>
              <li>{t('telemetrySetup.deploymentItem2')}</li>
              <li>{t('telemetrySetup.deploymentItem3')}</li>
            </ul>
            <p className="text-xs text-gray-400 mt-3 italic">{t('telemetrySetup.neverIncludedNote')}</p>
          </div>
        )}
      </div>

      <ChoiceButtons
        value={values.telemetry_enabled}
        onChange={v => onChange({ telemetry_enabled: v })}
        yesLabel={t('telemetrySetup.enableTelemetry')}
        noLabel={t('telemetrySetup.noThanks')}
      />

      <p className="text-xs text-center text-gray-400 mt-5">
        {t('telemetrySetup.changeDecisionPrefix')}{' '}
        <a href="/admin/settings?tab=telemetry" className="underline hover:text-gray-600">
          {t('telemetrySetup.adminSettingsTelemetryLink')}
        </a>
        .
      </p>
    </div>
  )
}
