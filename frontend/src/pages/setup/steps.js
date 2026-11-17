import TelemetryStep from './TelemetryStep'
import UpdateCheckStep from './UpdateCheckStep'

// Ordered registry of first-run wizard steps. To add a step, append an entry:
//   id        unique id, used for the step indicator
//   labelKey  i18n key for the short label in the progress indicator
//   keys      admin config keys this step edits; saved together on "Next"
//   Component controlled component receiving { values, onChange(partial) }
export const WIZARD_STEPS = [
  { id: 'telemetry', labelKey: 'setupWizard.steps.telemetry', keys: ['telemetry_enabled'], Component: TelemetryStep },
  { id: 'updates', labelKey: 'setupWizard.steps.updates', keys: ['update_check_enabled'], Component: UpdateCheckStep },
]

// Admin config key set once the wizard is finished or dismissed.
export const WIZARD_DONE_KEY = 'setup_wizard_completed'
