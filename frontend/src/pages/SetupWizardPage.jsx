import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'
import { getAdminConfig, updateAdminConfig } from '../api/admin'
import { notifyFeaturesChanged } from '../context/FeaturesContext'
import { WIZARD_STEPS, WIZARD_DONE_KEY, isStepComplete } from './setup/steps'

// First-run setup wizard shell: step navigation, per-step save, and skip
// affordances. Steps live in ./setup/steps.js.
export default function SetupWizardPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [index, setIndex] = useState(0)
  const [values, setValues] = useState({})
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  // Prefill with any answers already stored (e.g. returning to the wizard).
  useEffect(() => {
    let cancelled = false
    getAdminConfig()
      .then(res => {
        if (cancelled) return
        const config = res.data?.config || {}
        const known = {}
        for (const step of WIZARD_STEPS) {
          for (const key of step.keys) {
            if (config[key] === 'true' || config[key] === 'false') known[key] = config[key]
          }
        }
        // Local choices made while loading win over stored values.
        setValues(v => ({ ...known, ...v }))
      })
      .catch(() => {})
    return () => { cancelled = true }
  }, [])

  const step = WIZARD_STEPS[index]
  const isLast = index === WIZARD_STEPS.length - 1
  const answered = isStepComplete(step, values)

  async function finish(extra = {}) {
    setSaving(true)
    setError('')
    try {
      await updateAdminConfig({ ...extra, [WIZARD_DONE_KEY]: 'true' })
      notifyFeaturesChanged()
      navigate('/dashboard', { replace: true })
    } catch (err) {
      setError(err.response?.data?.error || t('setupWizard.saveFailed'))
      setSaving(false)
    }
  }

  function stepUpdates() {
    return Object.fromEntries(step.keys.filter(k => values[k] !== undefined).map(k => [k, values[k]]))
  }

  async function handleNext() {
    if (isLast) {
      await finish(stepUpdates())
      return
    }
    setSaving(true)
    setError('')
    try {
      await updateAdminConfig(stepUpdates())
      notifyFeaturesChanged()
      setIndex(i => i + 1)
    } catch (err) {
      setError(err.response?.data?.error || t('setupWizard.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleSkipStep() {
    setError('')
    if (isLast) {
      await finish()
    } else {
      setIndex(i => i + 1)
    }
  }

  const StepComponent = step.Component

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-lg max-w-lg w-full p-8">
        <ol className="flex items-center justify-center gap-2 mb-6" aria-label={t('setupWizard.progressLabel')}>
          {WIZARD_STEPS.map((s, i) => (
            <li
              key={s.id}
              aria-current={i === index ? 'step' : undefined}
              className={`flex items-center gap-2 text-xs ${i === index ? 'text-blue-700 font-semibold' : i < index ? 'text-gray-500' : 'text-gray-400'}`}
            >
              <span className={`w-6 h-6 rounded-full flex items-center justify-center border ${i === index ? 'bg-blue-600 text-white border-blue-600' : i < index ? 'bg-blue-50 border-blue-200' : 'border-gray-300'}`}>
                {i < index ? '✓' : i + 1}
              </span>
              <span>{t(s.labelKey)}</span>
              {i < WIZARD_STEPS.length - 1 && <span className="w-6 border-t border-gray-300" aria-hidden="true" />}
            </li>
          ))}
        </ol>

        <StepComponent values={values} onChange={partial => setValues(v => ({ ...v, ...partial }))} />

        {error && <p role="alert" className="text-sm text-red-600 mt-4">{error}</p>}

        <div className="flex items-center justify-between gap-3 mt-6">
          <button
            type="button"
            onClick={() => setIndex(i => i - 1)}
            disabled={saving || index === 0}
            className="px-4 py-2 text-sm border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50 disabled:opacity-40 transition"
          >
            {t('setupWizard.back')}
          </button>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={handleSkipStep}
              disabled={saving}
              className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700 disabled:opacity-40 transition"
            >
              {t('setupWizard.skipStep')}
            </button>
            <button
              type="button"
              onClick={handleNext}
              disabled={saving || !answered}
              className="px-5 py-2 text-sm font-semibold bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition"
            >
              {saving ? t('setupWizard.saving') : isLast ? t('setupWizard.finish') : t('setupWizard.next')}
            </button>
          </div>
        </div>

        <p className="text-center mt-6">
          <button
            type="button"
            onClick={() => finish()}
            disabled={saving}
            className="text-xs text-gray-400 underline hover:text-gray-600 disabled:opacity-40"
          >
            {t('setupWizard.skipAll')}
          </button>
        </p>
      </div>
    </div>
  )
}
