import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getAdminRoles, listRolePresets, getRolePresetDiff } from '../api/admin'

export default function RolePresetsPage() {
  const { t } = useTranslation()
  const [presets, setPresets] = useState([])
  const [roles, setRoles] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Compare state
  const [selectedRoleId, setSelectedRoleId] = useState('')
  const [selectedPresetId, setSelectedPresetId] = useState('')
  const [diff, setDiff] = useState(null)
  const [diffLoading, setDiffLoading] = useState(false)
  const [diffError, setDiffError] = useState(null)

  useEffect(() => {
    let cancelled = false
    Promise.all([
      listRolePresets().then(res => Array.isArray(res.data) ? res.data : []),
      getAdminRoles().then(res => Array.isArray(res.data) ? res.data : []),
    ])
      .then(([p, r]) => {
        if (!cancelled) {
          setPresets(p)
          setRoles(r)
        }
      })
      .catch(err => {
        if (!cancelled) setError(err.response?.data?.message || err.message || t('rolePresets.loadDataFailed'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [t])

  async function handleCompare() {
    if (!selectedRoleId || !selectedPresetId) return
    setDiffLoading(true)
    setDiffError(null)
    setDiff(null)
    try {
      const res = await getRolePresetDiff(selectedRoleId, selectedPresetId)
      setDiff(res.data)
    } catch (err) {
      setDiffError(err.response?.data?.message || err.message || t('rolePresets.loadDiffFailed'))
    } finally {
      setDiffLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="p-6 text-sm text-gray-500 dark:text-gray-400">{t('reconciliation.loading')}</div>
    )
  }

  if (error) {
    return (
      <div className="p-6 text-sm text-red-600 dark:text-red-400">{error}</div>
    )
  }

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-10">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('usersRoles.presetsTab')}</h1>

      {/* Preset cards */}
      <network>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          {t('rolePresets.subtitle')}
        </p>
        <div className="grid gap-4 sm:grid-cols-2">
          {presets.map(preset => (
            <div
              key={preset.id}
              className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 space-y-3"
            >
              <div>
                <h2 className="text-base font-semibold text-gray-900 dark:text-gray-100">{preset.name}</h2>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{preset.description}</p>
              </div>
              <div className="flex flex-wrap gap-1">
                {(preset.permissions || []).map(perm => (
                  <span
                    key={perm}
                    className="inline-block text-xs font-mono px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300"
                  >
                    {perm}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </network>

      {/* Compare role with preset */}
      <network className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-6 space-y-4">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{t('rolePresets.compareTitle')}</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {t('rolePresets.compareSubtitle')}
        </p>

        <div className="flex flex-wrap items-end gap-3">
          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-gray-600 dark:text-gray-400" htmlFor="role-select">
              {t('rolePresets.role')}
            </label>
            <select
              id="role-select"
              value={selectedRoleId}
              onChange={e => { setSelectedRoleId(e.target.value); setDiff(null) }}
              className="border border-gray-300 dark:border-gray-600 rounded px-3 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">{t('rolePresets.selectRolePlaceholder')}</option>
              {roles.map(r => (
                <option key={r.id} value={r.id}>{r.name}</option>
              ))}
            </select>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-gray-600 dark:text-gray-400" htmlFor="preset-select">
              {t('rolePresets.preset')}
            </label>
            <select
              id="preset-select"
              value={selectedPresetId}
              onChange={e => { setSelectedPresetId(e.target.value); setDiff(null) }}
              className="border border-gray-300 dark:border-gray-600 rounded px-3 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">{t('rolePresets.selectPresetPlaceholder')}</option>
              {presets.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>

          <button
            onClick={handleCompare}
            disabled={!selectedRoleId || !selectedPresetId || diffLoading}
            className="px-4 py-1.5 text-sm font-medium rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {diffLoading ? t('rolePresets.comparing') : t('rolePresets.compare')}
          </button>
        </div>

        {diffError && (
          <p className="text-sm text-red-600 dark:text-red-400">{diffError}</p>
        )}

        {diff && (
          <div className="mt-4 space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {t('rolePresets.comparingPrefix')}<strong className="text-gray-900 dark:text-gray-100">{diff.role?.name}</strong>{t('rolePresets.comparingMiddle')}<strong className="text-gray-900 dark:text-gray-100">{diff.preset?.name}</strong>{t('rolePresets.comparingSuffix')}
            </p>

            {diff.added?.length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-green-700 dark:text-green-400 mb-1">
                  {t('rolePresets.addedHeading', { count: diff.added.length })}
                </h3>
                <div className="flex flex-wrap gap-1">
                  {diff.added.map(p => (
                    <span
                      key={p}
                      className="inline-block text-xs font-mono px-2 py-0.5 rounded bg-green-100 dark:bg-green-900/40 text-green-800 dark:text-green-300"
                    >
                      {p}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {diff.removed?.length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-red-700 dark:text-red-400 mb-1">
                  {t('rolePresets.removedHeading', { count: diff.removed.length })}
                </h3>
                <div className="flex flex-wrap gap-1">
                  {diff.removed.map(p => (
                    <span
                      key={p}
                      className="inline-block text-xs font-mono px-2 py-0.5 rounded bg-red-100 dark:bg-red-900/40 text-red-800 dark:text-red-300"
                    >
                      {p}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {diff.unchanged?.length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-1">
                  {t('rolePresets.unchangedHeading', { count: diff.unchanged.length })}
                </h3>
                <div className="flex flex-wrap gap-1">
                  {diff.unchanged.map(p => (
                    <span
                      key={p}
                      className="inline-block text-xs font-mono px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400"
                    >
                      {p}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {diff.added?.length === 0 && diff.removed?.length === 0 && (
              <p className="text-sm text-gray-500 dark:text-gray-400 italic">
                {t('rolePresets.matchesExactly')}
              </p>
            )}
          </div>
        )}
      </network>
    </div>
  )
}
