import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getUniFiConflicts, getUniFiStatus, resolveUniFiConflict, syncUniFi, testUniFiConnection } from '../../api/unifi'

const INPUT = 'w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm'

export default function UnifiTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  const { t } = useTranslation()
  const [testStatus, setTestStatus] = useState(null) // null | 'testing' | { ok, message }
  const [syncStatus, setSyncStatus] = useState(null) // null | 'syncing' | { ok, message }
  const [status, setStatus] = useState(null)
  const [conflicts, setConflicts] = useState([])
  const [conflictError, setConflictError] = useState('')

  const refresh = useCallback(async () => {
    try {
      const [st, cf] = await Promise.all([getUniFiStatus(), getUniFiConflicts('open')])
      setStatus(st.data)
      setConflicts(Array.isArray(cf.data) ? cf.data : [])
    } catch {
      // Status is informational; the settings form still works without it.
    }
  }, [])

  useEffect(() => { refresh() }, [refresh])

  const connection = () => ({
    url: config.unifi_url || '',
    api_key: config.unifi_api_key || '',
    username: config.unifi_username || '',
    password: config.unifi_password || '',
    site: config.unifi_site || '',
    skip_tls: config.unifi_skip_tls === 'true',
  })

  async function handleTest() {
    setTestStatus('testing')
    try {
      await testUniFiConnection(connection())
      setTestStatus({ ok: true, message: t('unifiTab.connected') })
    } catch (err) {
      setTestStatus({ ok: false, message: err.response?.data?.error || err.message || t('unifiTab.connectionFailed') })
    }
  }

  async function handleSync() {
    setSyncStatus('syncing')
    try {
      const res = await syncUniFi()
      const r = res.data || {}
      setSyncStatus({ ok: true, message: t('unifiTab.syncDone', {
        vlans: (r.vlansLinked ?? 0), clients: (r.clientsCreated ?? 0) + (r.clientsUpdated ?? 0), devices: (r.devicesCreated ?? 0) + (r.devicesLinked ?? 0) + (r.devicesUpdated ?? 0), conflicts: r.conflictsRaised ?? 0,
      }) })
    } catch (err) {
      setSyncStatus({ ok: false, message: err.response?.data?.error || err.message || t('unifiTab.syncFailed') })
    } finally {
      await refresh()
    }
  }

  async function handleResolve(id, action) {
    setConflictError('')
    try {
      await resolveUniFiConflict(id, action)
      await refresh()
    } catch (err) {
      setConflictError(err.response?.data?.error || t('unifiTab.resolveFailed'))
    }
  }

  const fieldLabel = f => (f === 'mac_address' ? t('unifiTab.fieldMac') : t('unifiTab.fieldHostname'))

  return (
    <div className="space-y-4">
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">{t('unifiTab.title')}</h2>
        <p className="text-sm text-gray-500 mb-4">{t('unifiTab.intro')}</p>

        <div className="mb-4">
          <label htmlFor="unifi_url" className="block text-sm font-medium text-gray-700 mb-1">{t('unifiTab.url')}</label>
          <input id="unifi_url" type="text" value={config.unifi_url || ''} onChange={e => handleConfigChange('unifi_url', e.target.value)} className={INPUT} placeholder="https://192.168.1.1" />
        </div>

        <div className="mb-4">
          <label htmlFor="unifi_api_key" className="block text-sm font-medium text-gray-700 mb-1">{t('unifiTab.apiKey')}</label>
          <input id="unifi_api_key" type="password" value={config.unifi_api_key || ''} onChange={e => handleConfigChange('unifi_api_key', e.target.value)} className={INPUT} placeholder="••••••••" autoComplete="off" />
          <p className="text-xs text-gray-500 mt-1">{t('unifiTab.apiKeyHint')}</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div>
            <label htmlFor="unifi_username" className="block text-sm font-medium text-gray-700 mb-1">{t('unifiTab.username')}</label>
            <input id="unifi_username" type="text" value={config.unifi_username || ''} onChange={e => handleConfigChange('unifi_username', e.target.value)} className={INPUT} autoComplete="off" />
          </div>
          <div>
            <label htmlFor="unifi_password" className="block text-sm font-medium text-gray-700 mb-1">{t('unifiTab.password')}</label>
            <input id="unifi_password" type="password" value={config.unifi_password || ''} onChange={e => handleConfigChange('unifi_password', e.target.value)} className={INPUT} placeholder="••••••••" autoComplete="new-password" />
          </div>
          <p className="md:col-span-2 text-xs text-gray-500 -mt-2">{t('unifiTab.loginHint')}</p>
        </div>

        <div className="mb-4">
          <label htmlFor="unifi_site" className="block text-sm font-medium text-gray-700 mb-1">{t('unifiTab.site')}</label>
          <input id="unifi_site" type="text" value={config.unifi_site || ''} onChange={e => handleConfigChange('unifi_site', e.target.value)} className={INPUT} placeholder="default" />
          <p className="text-xs text-gray-500 mt-1">{t('unifiTab.siteHint')}</p>
        </div>

        <div className="mb-4 flex items-center gap-2">
          <input id="unifi_skip_tls" type="checkbox" checked={config.unifi_skip_tls === 'true'} onChange={e => handleConfigChange('unifi_skip_tls', e.target.checked ? 'true' : 'false')} className="h-4 w-4 text-blue-600 rounded border-gray-300" />
          <label htmlFor="unifi_skip_tls" className="text-sm font-medium text-gray-700">{t('unifiTab.skipTls')}</label>
          <span className="text-xs text-yellow-600">{t('unifiTab.skipTlsWarning')}</span>
        </div>

        <div className="mb-4">
          <label htmlFor="unifi_sync_interval_minutes" className="block text-sm font-medium text-gray-700 mb-1">{t('unifiTab.interval')}</label>
          <input id="unifi_sync_interval_minutes" type="number" min="0" step="1" value={config.unifi_sync_interval_minutes ?? '0'} onChange={e => handleConfigChange('unifi_sync_interval_minutes', e.target.value)} className="w-32 px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm" />
          <p className="text-xs text-gray-500 mt-1">{t('unifiTab.intervalHint')}</p>
        </div>

        {testStatus && testStatus !== 'testing' && (
          <div role="status" className={`mb-4 px-3 py-2 rounded text-sm ${testStatus.ok ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
            {testStatus.message}
          </div>
        )}

        <div className="flex gap-3">
          <button onClick={handleTest} disabled={testStatus === 'testing'} className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium">
            {testStatus === 'testing' ? t('unifiTab.testing') : t('unifiTab.testConnection')}
          </button>
          <button onClick={handleSaveConfig} disabled={saving} className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700 disabled:bg-blue-400 transition text-sm font-medium">
            {saving ? t('common.saving') : t('common.save')}
          </button>
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-2">{t('unifiTab.syncTitle')}</h2>
        <p className="text-sm text-gray-500 mb-3">{t('unifiTab.syncIntro')}</p>
        {status && (
          <div className="text-sm text-gray-600 mb-3 space-y-1">
            <p>{status.lastSyncedAt ? t('unifiTab.lastSynced', { time: new Date(status.lastSyncedAt).toLocaleString() }) : t('unifiTab.neverSynced')}</p>
            {status.lastError && <p className="text-red-600">{t('unifiTab.lastError', { error: status.lastError })}</p>}
            {status.lastResult?.errors?.length > 0 && (
              <ul className="list-disc list-inside text-red-600 text-xs">
                {status.lastResult.errors.map((e, i) => <li key={i}>{e}</li>)}
              </ul>
            )}
            {status.lastResult?.clientsUnmatched > 0 && <p className="text-gray-500 text-xs">{t('unifiTab.unmatched', { count: status.lastResult.clientsUnmatched })}</p>}
          </div>
        )}
        {syncStatus && syncStatus !== 'syncing' && (
          <div role="status" className={`mb-3 px-3 py-2 rounded text-sm ${syncStatus.ok ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
            {syncStatus.message}
          </div>
        )}
        <button onClick={handleSync} disabled={syncStatus === 'syncing' || (status && !status.configured)} className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium">
          {syncStatus === 'syncing' ? t('unifiTab.syncing') : t('unifiTab.syncNow')}
        </button>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">{t('unifiTab.conflictsTitle')}</h2>
        <p className="text-sm text-gray-500 mb-3">{t('unifiTab.conflictsIntro')}</p>
        {conflictError && <p role="alert" className="text-sm text-red-600 mb-2">{conflictError}</p>}
        {conflicts.length === 0 ? (
          <p className="text-sm text-gray-500">{t('unifiTab.noConflicts')}</p>
        ) : (
          <div className="overflow-x-auto rounded border border-gray-200">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left">{t('associateIp.ipAddress')}</th>
                  <th className="px-3 py-2 text-left">{t('unifiTab.field')}</th>
                  <th className="px-3 py-2 text-left">{t('unifiTab.padduckValue')}</th>
                  <th className="px-3 py-2 text-left">{t('unifiTab.unifiValue')}</th>
                  <th />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {conflicts.map(c => (
                  <tr key={c.id}>
                    <td className="px-3 py-2 font-mono">{c.ipAddress}</td>
                    <td className="px-3 py-2">{fieldLabel(c.field)}</td>
                    <td className="px-3 py-2 font-mono">{c.padduckValue || '—'}</td>
                    <td className="px-3 py-2 font-mono">{c.unifiValue || '—'}</td>
                    <td className="px-3 py-2 text-right space-x-3 whitespace-nowrap">
                      <button className="text-blue-600 text-xs" onClick={() => handleResolve(c.id, 'accept')}>{t('unifiTab.useUnifi')}</button>
                      <button className="text-gray-600 text-xs" onClick={() => handleResolve(c.id, 'dismiss')}>{t('unifiTab.keepPadduck')}</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
