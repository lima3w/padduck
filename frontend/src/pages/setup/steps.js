import TelemetryStep from './TelemetryStep'
import UpdateCheckStep from './UpdateCheckStep'
import FeaturesStep from './FeaturesStep'

// Sidebar sections and the feature flags that control them, in nav order.
// DNS, Admin and Automation have no feature flag and are always shown.
export const FEATURE_SECTIONS = [
  {
    id: 'physical',
    labelKey: 'nav.sections.physical',
    keys: ['feature_devices_enabled', 'feature_locations_enabled', 'feature_racks_enabled', 'feature_bgp_enabled'],
  },
  {
    id: 'networkServices',
    labelKey: 'nav.sections.networkServices',
    keys: ['feature_nat_enabled', 'feature_firewall_enabled', 'feature_dhcp_enabled', 'feature_circuits_enabled'],
  },
  { id: 'customers', labelKey: 'nav.sections.customers', keys: ['feature_customers_enabled'] },
  { id: 'vlansVrfs', labelKey: 'nav.sections.vlansVrfs', keys: ['feature_vlans_enabled', 'feature_vrfs_enabled'] },
]

const FEATURE_KEYS = FEATURE_SECTIONS.flatMap(s => s.keys)

const isChosen = (values, key) => values[key] === 'true' || values[key] === 'false'

// Ordered registry of first-run wizard steps. To add a step, append an entry:
//   id          unique id, used for the step indicator
//   labelKey    i18n key for the short label in the progress indicator
//   keys        admin config keys this step edits; saved together on "Next"
//   Component   controlled component receiving { values, onChange(partial) }
//   isComplete  optional (values) => boolean gating "Next"; defaults to every
//               key in `keys` having been explicitly chosen
export const WIZARD_STEPS = [
  { id: 'telemetry', labelKey: 'setupWizard.steps.telemetry', keys: ['telemetry_enabled'], Component: TelemetryStep },
  { id: 'updates', labelKey: 'setupWizard.steps.updates', keys: ['update_check_enabled'], Component: UpdateCheckStep },
  // Every module already has a value (new installs start with the homelab-unneeded ones off), so this step is always ready to continue.
  { id: 'features', labelKey: 'setupWizard.steps.features', keys: FEATURE_KEYS, Component: FeaturesStep, isComplete: () => true },
]

export function isStepComplete(step, values) {
  return step.isComplete ? step.isComplete(values) : step.keys.every(k => isChosen(values, k))
}

// Admin config key set once the wizard is finished or dismissed.
export const WIZARD_DONE_KEY = 'setup_wizard_completed'
