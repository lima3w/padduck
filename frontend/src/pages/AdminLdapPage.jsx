import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import * as client from '../api/admin'

const defaultConfig = {
  enabled: false,
  host: '',
  port: 389,
  tls_mode: 'none',
  skip_cert_verify: false,
  bind_dn: '',
  bind_password: '',
  base_dn: '',
  user_filter: '(sAMAccountName=%s)',
  username_attr: 'sAMAccountName',
  email_attr: 'mail',
}

export default function AdminLdapPage() {
  const { t } = useTranslation()
  const [config, setConfig] = useState(defaultConfig)
  const [passwordSet, setPasswordSet] = useState(false)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState(null)
  const [message, setMessage] = useState({ text: '', type: '' })
  const [loading, setLoading] = useState(true)

  const [mappings, setMappings] = useState([])
  const [roles, setRoles] = useState([])
  const [newMapping, setNewMapping] = useState({ ldap_group_dn: '', role_id: '' })
  const [addingMapping, setAddingMapping] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [cfgRes, mappingsRes, rolesRes] = await Promise.all([
        client.getLdapConfig().catch(() => null),
        client.getLdapGroupMappings().catch(() => ({ data: [] })),
        client.getAdminRoles().catch(() => ({ data: [] })),
      ])
      if (cfgRes) {
        const c = cfgRes.data
        setPasswordSet(c.bindPassword === '****')
        setConfig({
          enabled: c.enabled ?? defaultConfig.enabled,
          host: c.host ?? defaultConfig.host,
          port: c.port ?? defaultConfig.port,
          tls_mode: c.tlsMode ?? defaultConfig.tls_mode,
          skip_cert_verify: c.skipCertVerify ?? defaultConfig.skip_cert_verify,
          bind_dn: c.bindDn ?? defaultConfig.bind_dn,
          bind_password: '',
          base_dn: c.baseDn ?? defaultConfig.base_dn,
          user_filter: c.userFilter ?? defaultConfig.user_filter,
          username_attr: c.usernameAttr ?? defaultConfig.username_attr,
          email_attr: c.emailAttr ?? defaultConfig.email_attr,
        })
      }
      setMappings(mappingsRes.data || [])
      setRoles(rolesRes.data || [])
    } catch (err) {
      showMessage(t('adminLdap.loadFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadData()
  }, [loadData])

  const showMessage = (text, type = 'success') => {
    setMessage({ text, type })
    setTimeout(() => setMessage({ text: '', type: '' }), 4000)
  }

  const handleChange = (key, value) => {
    setConfig((prev) => ({ ...prev, [key]: value }))
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = { ...config }
      if (!payload.bind_password) delete payload.bind_password
      await client.updateLdapConfig(payload)
      showMessage(t('adminLdap.configSaved'))
      loadData()
    } catch (err) {
      showMessage(t('adminLdap.saveFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleTest = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      const res = await client.testLdapConnection()
      setTestResult({ ok: true, message: res.data?.message || t('adminLdap.connectionSuccessful') })
    } catch (err) {
      const msg = err.response?.data?.error || err.message || t('adminLdap.connectionFailed')
      setTestResult({ ok: false, message: msg })
    } finally {
      setTesting(false)
    }
  }

  const handleAddMapping = async () => {
    if (!newMapping.ldap_group_dn || !newMapping.role_id) {
      showMessage(t('adminLdap.bothFieldsRequired'), 'error')
      return
    }
    setAddingMapping(true)
    try {
      await client.createLdapGroupMapping(newMapping)
      setNewMapping({ ldap_group_dn: '', role_id: '' })
      const res = await client.getLdapGroupMappings()
      setMappings(res.data || [])
      showMessage(t('adminLdap.mappingAdded'))
    } catch (err) {
      showMessage(t('adminLdap.addMappingFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setAddingMapping(false)
    }
  }

  const handleDeleteMapping = async (id) => {
    if (!window.confirm(t('adminLdap.deleteMappingConfirm'))) return
    try {
      await client.deleteLdapGroupMapping(id)
      setMappings((prev) => prev.filter((m) => m.id !== id))
      showMessage(t('adminLdap.mappingDeleted'))
    } catch (err) {
      showMessage(t('adminLdap.deleteMappingFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
        {t('adminLdap.loadingSettings')}
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">{t('adminLdap.title')}</h1>

      {message.text && (
        <div
          className={`mb-4 p-4 rounded text-sm ${
            message.type === 'error'
              ? 'bg-red-50 border border-red-200 text-red-700 dark:bg-red-900/30 dark:border-red-700 dark:text-red-300'
              : 'bg-green-50 border border-green-200 text-green-700 dark:bg-green-900/30 dark:border-green-700 dark:text-green-300'
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="space-y-6">
        {/* Connection Settings */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{t('adminLdap.connectionSettingsTitle')}</h2>

          <label className="flex items-center gap-3 mb-4 cursor-pointer">
            <input
              type="checkbox"
              checked={!!config.enabled}
              onChange={(e) => handleChange('enabled', e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              <strong>{t('adminLdap.enableLdap')}</strong>
              <span className="block text-gray-500 dark:text-gray-400">{t('adminLdap.enableLdapHint')}</span>
            </span>
          </label>

          <div className="grid grid-cols-3 gap-4 mb-4">
            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.host')}</label>
              <input
                type="text"
                value={config.host || ''}
                onChange={(e) => handleChange('host', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="ldap.example.com"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.port')}</label>
              <input
                type="number"
                value={config.port || 389}
                onChange={(e) => handleChange('port', parseInt(e.target.value, 10) || 389)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="389"
              />
            </div>
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.tlsMode')}</label>
            <select
              value={config.tls_mode || 'none'}
              onChange={(e) => handleChange('tls_mode', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            >
              <option value="none">{t('adminLdap.tlsModeNone')}</option>
              <option value="starttls">{t('adminLdap.tlsModeStarttls')}</option>
              <option value="tls">{t('adminLdap.tlsModeTls')}</option>
            </select>
          </div>

          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={!!config.skip_cert_verify}
              onChange={(e) => handleChange('skip_cert_verify', e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">{t('adminLdap.skipCertVerify')}</span>
          </label>
        </div>

        {/* Bind Credentials */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{t('adminLdap.bindCredentialsTitle')}</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.bindDn')}</label>
            <input
              type="text"
              value={config.bind_dn || ''}
              onChange={(e) => handleChange('bind_dn', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="cn=ldapbind,dc=example,dc=com"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.bindPassword')}</label>
            <input
              type="password"
              value={config.bind_password || ''}
              onChange={(e) => handleChange('bind_password', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder={passwordSet ? t('adminLdap.unchanged') : t('adminLdap.enterBindPassword')}
            />
            {passwordSet && (
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('adminLdap.leaveBlankToKeepPassword')}</p>
            )}
          </div>
        </div>

        {/* Search Settings */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{t('adminLdap.searchSettingsTitle')}</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.baseDn')}</label>
            <input
              type="text"
              value={config.base_dn || ''}
              onChange={(e) => handleChange('base_dn', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="dc=example,dc=com"
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.userFilter')}</label>
            <input
              type="text"
              value={config.user_filter || ''}
              onChange={(e) => handleChange('user_filter', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 font-mono"
              placeholder="(sAMAccountName=%s)"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('adminLdap.userFilterHintPrefix')}<code>%s</code>{t('adminLdap.userFilterHintSuffix')}</p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.usernameAttr')}</label>
              <input
                type="text"
                value={config.username_attr || ''}
                onChange={(e) => handleChange('username_attr', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="sAMAccountName"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminLdap.emailAttr')}</label>
              <input
                type="text"
                value={config.email_attr || ''}
                onChange={(e) => handleChange('email_attr', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="mail"
              />
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-3 items-start flex-wrap">
          <button
            onClick={handleSave}
            disabled={saving}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
          >
            {saving ? t('common.saving') : t('scanRetention.saveSettings')}
          </button>
          <button
            onClick={handleTest}
            disabled={testing}
            className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50 transition text-sm font-medium"
          >
            {testing ? t('adminLdap.testing') : t('adminLdap.testConnection')}
          </button>
          {testResult && (
            <div
              className={`px-3 py-2 rounded text-sm ${
                testResult.ok
                  ? 'bg-green-50 border border-green-200 text-green-700 dark:bg-green-900/30 dark:border-green-700 dark:text-green-300'
                  : 'bg-red-50 border border-red-200 text-red-700 dark:bg-red-900/30 dark:border-red-700 dark:text-red-300'
              }`}
            >
              {testResult.ok ? testResult.message : `${t('adminLdap.errorPrefix')}${testResult.message}`}
            </div>
          )}
        </div>

        {/* Group → Role Mappings */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-1 text-gray-900 dark:text-gray-100">{t('adminLdap.groupRoleMappingsTitle')}</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            {t('adminLdap.groupRoleMappingsSubtitle')}
          </p>

          {mappings.length > 0 ? (
            <table className="w-full text-sm mb-4">
              <thead>
                <tr className="text-left text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
                  <th className="pb-2 font-medium">{t('adminLdap.ldapGroupDnColumn')}</th>
                  <th className="pb-2 font-medium">{t('rolePresets.role')}</th>
                  <th className="pb-2"></th>
                </tr>
              </thead>
              <tbody>
                {mappings.map((m) => {
                  const role = roles.find((r) => r.id === m.roleId)
                  return (
                    <tr key={m.id} className="border-b border-gray-100 dark:border-gray-700">
                      <td className="py-2 font-mono text-xs text-gray-800 dark:text-gray-200 pr-4">{m.ldapGroupDn}</td>
                      <td className="py-2 text-gray-700 dark:text-gray-300 pr-4">{role?.name || m.roleId}</td>
                      <td className="py-2 text-right">
                        <button
                          onClick={() => handleDeleteMapping(m.id)}
                          className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 text-xs font-medium"
                        >
                          {t('common.delete')}
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          ) : (
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">{t('adminLdap.noGroupMappings')}</p>
          )}

          <div className="flex gap-3 items-end">
            <div className="flex-1">
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">{t('adminLdap.ldapGroupDnColumn')}</label>
              <input
                type="text"
                value={newMapping.ldap_group_dn}
                onChange={(e) => setNewMapping((prev) => ({ ...prev, ldap_group_dn: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 font-mono"
                placeholder="CN=ipam-admins,OU=Groups,DC=example,DC=com"
              />
            </div>
            <div className="w-40">
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">{t('rolePresets.role')}</label>
              <select
                value={newMapping.role_id}
                onChange={(e) => setNewMapping((prev) => ({ ...prev, role_id: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              >
                <option value="">{t('adminLdap.selectRole')}</option>
                {roles.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </div>
            <button
              onClick={handleAddMapping}
              disabled={addingMapping}
              className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50 transition text-sm font-medium whitespace-nowrap"
            >
              {addingMapping ? t('adminLdap.adding') : t('adminLdap.addMapping')}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
