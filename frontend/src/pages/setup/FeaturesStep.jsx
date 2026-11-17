import { useTranslation } from 'react-i18next'
import { FEATURE_SECTIONS } from './steps'

// First-run wizard step: choose which modules (sidebar sections) this
// deployment uses. Reuses the existing feature_*_enabled flags, which hide the
// module from navigation and make its API routes reject direct access. A flag
// that has never been changed counts as enabled. Controlled: the wizard owns
// `values` and saves them on Next.
export default function FeaturesStep({ values, onChange }) {
  const { t } = useTranslation()

  return (
    <div>
      <div className="mb-6 text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">{t('setupWizard.features.title')}</h2>
        <p className="text-gray-600 text-sm leading-relaxed">{t('setupWizard.features.subtitle')}</p>
      </div>

      <div className="space-y-5">
        {FEATURE_SECTIONS.map(section => (
          <fieldset key={section.id}>
            <legend className="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2">{t(section.labelKey)}</legend>
            <div className="space-y-2">
              {section.keys.map(key => (
                <label key={key} className="flex items-start gap-3 rounded border border-gray-200 p-3 cursor-pointer hover:bg-gray-50">
                  <input
                    type="checkbox"
                    checked={values[key] !== 'false'}
                    onChange={e => onChange({ [key]: e.target.checked ? 'true' : 'false' })}
                    className="mt-1 h-4 w-4 rounded text-blue-600"
                  />
                  <span>
                    <span className="block text-sm font-medium text-gray-900">{t(`featuresTab.toggles.${key}.title`)}</span>
                    <span className="block text-xs text-gray-500">{t(`featuresTab.toggles.${key}.description`)}</span>
                  </span>
                </label>
              ))}
            </div>
          </fieldset>
        ))}
      </div>

      <p className="text-xs text-center text-gray-400 mt-5">
        {t('setupWizard.features.changeLaterPrefix')}{' '}
        <a href="/admin/settings?tab=features" className="underline hover:text-gray-600">
          {t('setupWizard.features.adminSettingsLink')}
        </a>
        .
      </p>
    </div>
  )
}
