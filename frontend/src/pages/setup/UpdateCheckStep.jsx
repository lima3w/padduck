import { useTranslation } from 'react-i18next'
import ChoiceButtons from './ChoiceButtons'

// First-run wizard step: opt in to checking GitHub for new Padduck releases.
// Controlled: the wizard owns the value and saves it on Next.
export default function UpdateCheckStep({ values, onChange }) {
  const { t } = useTranslation()

  return (
    <div>
      <div className="mb-6 text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">{t('setupWizard.updates.title')}</h2>
        <p className="text-gray-600 text-sm leading-relaxed">{t('setupWizard.updates.subtitle')}</p>
      </div>

      <div className="bg-blue-50 border border-blue-100 rounded-lg p-4 mb-6 text-sm text-blue-800 space-y-1">
        <p>{t('setupWizard.updates.detail')}</p>
        <p>{t('setupWizard.updates.privacy')}</p>
      </div>

      <ChoiceButtons
        value={values.update_check_enabled}
        onChange={v => onChange({ update_check_enabled: v })}
        yesLabel={t('setupWizard.updates.enable')}
        noLabel={t('setupWizard.updates.disable')}
      />

      <p className="text-xs text-center text-gray-400 mt-5">
        {t('setupWizard.updates.changeLaterPrefix')}{' '}
        <a href="/admin/settings?tab=updates" className="underline hover:text-gray-600">
          {t('setupWizard.updates.adminSettingsLink')}
        </a>
        .
      </p>
    </div>
  )
}
