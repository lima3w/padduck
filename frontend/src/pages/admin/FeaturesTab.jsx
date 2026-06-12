import { FEATURE_TOGGLES } from './settingsShared'

export default function FeaturesTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-2">Enabled Modules</h2>
            <p className="text-sm text-gray-600 mb-5">
              Disabled modules are removed from navigation and their API routes reject direct access.
            </p>
            <div className="grid gap-4 md:grid-cols-2">
              {FEATURE_TOGGLES.map((feature) => (
                <label
                  key={feature.key}
                  className="flex items-start gap-3 rounded border border-gray-200 dark:border-gray-700 p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50"
                >
                  <input
                    type="checkbox"
                    checked={config[feature.key] !== 'false'}
                    onChange={(e) => handleConfigChange(feature.key, e.target.checked ? 'true' : 'false')}
                    className="mt-1 h-4 w-4 rounded text-blue-600"
                  />
                  <span>
                    <span className="block font-medium text-gray-900 dark:text-gray-100">{feature.title}</span>
                    <span className="block text-sm text-gray-500 dark:text-gray-400">{feature.description}</span>
                  </span>
                </label>
              ))}
            </div>
          </div>

          <button
            onClick={handleSaveConfig}
            disabled={saving}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
  )
}
