import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { checkForUpdates } from '../../api/admin'

export default function UpdatesTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  const { t } = useTranslation()
  const [updateStatus, setUpdateStatus] = useState(null)
  const [checkingUpdates, setCheckingUpdates] = useState(false)

  const handleUpdateCheck = async () => {
    setCheckingUpdates(true)
    setUpdateStatus(null)
    try {
      const res = await checkForUpdates()
      setUpdateStatus({ ok: true, data: res.data })
    } catch (err) {
      setUpdateStatus({
        ok: false,
        message: err.response?.data?.error || err.message || t('updatesTab.checkFailed'),
      })
    } finally {
      setCheckingUpdates(false)
    }
  }

  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">{t('updatesTab.title')}</h2>

            <label className="flex items-center gap-3 mb-4 cursor-pointer">
              <input
                type="checkbox"
                checked={config.update_check_enabled === 'true'}
                onChange={(e) => handleConfigChange('update_check_enabled', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700">
                <strong>{t('updatesTab.enableUpdateChecks')}</strong>
                <span className="block text-gray-500">{t('updatesTab.enableUpdateChecksHint')}</span>
              </span>
            </label>
          </div>

          {updateStatus && (
            <div
              className={`rounded border p-4 text-sm ${
                updateStatus.ok
                  ? updateStatus.data?.updateAvailable
                    ? 'bg-yellow-50 border-yellow-200 text-yellow-800'
                    : 'bg-green-50 border-green-200 text-green-700'
                  : 'bg-red-50 border-red-200 text-red-700'
              }`}
            >
              {updateStatus.ok ? (
                <div className="space-y-1">
                  <p className="font-medium">
                    {updateStatus.data?.enabled === false
                      ? t('updatesTab.checksDisabled')
                      : updateStatus.data?.updateAvailable
                      ? t('updatesTab.updateAvailable', { version: updateStatus.data.latestVersion })
                      : t('updatesTab.noUpdateAvailable')}
                  </p>
                  <p>
                    {t('updatesTab.currentPrefix')}{updateStatus.data?.currentVersion || t('updatesTab.unknown')}
                    {updateStatus.data?.latestVersion ? t('updatesTab.latestSuffix', { version: updateStatus.data.latestVersion }) : ''}
                  </p>
                  {updateStatus.data?.releaseUrl && (
                    <a
                      href={updateStatus.data.releaseUrl}
                      className="inline-block text-blue-600 hover:underline"
                      target="_blank"
                      rel="noreferrer"
                    >
                      {t('updatesTab.viewRelease')}
                    </a>
                  )}
                </div>
              ) : (
                <p>{updateStatus.message}</p>
              )}
            </div>
          )}

          <div className="flex gap-3">
            <button
              onClick={handleSaveConfig}
              disabled={saving}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {saving ? t('common.saving') : t('common.save')}
            </button>
            <button
              onClick={handleUpdateCheck}
              disabled={checkingUpdates}
              className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
            >
              {checkingUpdates ? t('updatesTab.checking') : t('updatesTab.checkNow')}
            </button>
          </div>
        </div>
  )
}
