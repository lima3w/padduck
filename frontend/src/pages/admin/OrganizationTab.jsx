import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getOrgSettings, updateOrgSettings, testOrgSMTP } from '../../api/admin'

const QUOTAS = [
  ['max_subnets', 'subnets'],
  ['max_ip_addresses', 'ipAddresses'],
  ['max_users', 'users'],
  ['max_webhooks', 'webhooks'],
  ['max_api_tokens', 'apiTokens'],
]

// Organization limits (read-only: only a platform admin can change them) and
// the org's own SMTP overrides, with a test-send.
export default function OrganizationTab({ showMessage }) {
  const { t } = useTranslation()
  const [settings, setSettings] = useState(null)
  const [form, setForm] = useState({ smtp_host: '', smtp_port: '', smtp_from: '' })
  const [saving, setSaving] = useState(false)
  const [testEmail, setTestEmail] = useState('')

  useEffect(() => {
    getOrgSettings()
      .then((res) => {
        setSettings(res.data)
        setForm({
          smtp_host: res.data.smtp_host || '',
          smtp_port: res.data.smtp_port ?? '',
          smtp_from: res.data.smtp_from || '',
        })
      })
      .catch((err) => showMessage(t('orgTab.loadFailed') + (err.response?.data?.error || err.message), 'error'))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      const res = await updateOrgSettings({
        smtp_host: form.smtp_host || null,
        smtp_port: form.smtp_port === '' ? null : Number(form.smtp_port),
        smtp_from: form.smtp_from || null,
      })
      setSettings(res.data)
      showMessage(t('orgTab.saved'))
    } catch (err) {
      showMessage(t('orgTab.saveFailed') + (err.response?.data?.error || err.message), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleTest = async () => {
    if (!testEmail) {
      showMessage(t('smtpTab.enterEmailError'), 'error')
      return
    }
    try {
      await testOrgSMTP(testEmail)
      showMessage(t('smtpTab.testEmailSentPrefix') + testEmail)
    } catch (err) {
      showMessage(t('smtpTab.testFailedPrefix') + (err.response?.data?.error || err.message), 'error')
    }
  }

  if (!settings) return null

  const input = 'w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm'

  return (
    <div className="space-y-4">
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">{t('orgTab.limitsTitle')}</h2>
        <p className="text-sm text-gray-500 mb-4">{t('orgTab.limitsHelp')}</p>
        <dl className="grid grid-cols-2 md:grid-cols-5 gap-4">
          {QUOTAS.map(([key, label]) => (
            <div key={key}>
              <dt className="text-sm text-gray-600">{t(`orgTab.${label}`)}</dt>
              <dd className="text-lg font-semibold">{settings[key] ?? t('orgTab.unlimited')}</dd>
            </div>
          ))}
        </dl>
        <p className="text-sm text-gray-600 mt-4">
          {t('orgTab.auditRetention')}: {settings.audit_retention_days ? t('orgTab.days', { count: settings.audit_retention_days }) : t('orgTab.globalDefault')}
        </p>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-1">{t('orgTab.smtpTitle')}</h2>
        <p className="text-sm text-gray-500 mb-4">{t('orgTab.smtpHelp')}</p>
        <div className="grid grid-cols-3 gap-4 mb-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('smtpTab.smtpHost')}</label>
            <input type="text" className={input} value={form.smtp_host} placeholder={t('orgTab.useGlobal')}
              onChange={(e) => setForm({ ...form, smtp_host: e.target.value })} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('smtpTab.port')}</label>
            <input type="number" min="0" className={input} value={form.smtp_port} placeholder={t('orgTab.useGlobal')}
              onChange={(e) => setForm({ ...form, smtp_port: e.target.value })} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('orgTab.fromAddress')}</label>
            <input type="text" className={input} value={form.smtp_from} placeholder={t('orgTab.useGlobal')}
              onChange={(e) => setForm({ ...form, smtp_from: e.target.value })} />
          </div>
        </div>
        <button onClick={handleSave} disabled={saving}
          className="px-4 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50">
          {saving ? t('orgTab.saving') : t('orgTab.save')}
        </button>

        <div className="flex gap-2 mt-6 pt-4 border-t border-gray-200">
          <input type="email" className={input + ' max-w-xs'} value={testEmail} placeholder="you@example.com"
            onChange={(e) => setTestEmail(e.target.value)} />
          <button onClick={handleTest}
            className="px-4 py-2 border border-gray-300 rounded text-sm font-medium hover:bg-gray-50">
            {t('orgTab.sendTest')}
          </button>
        </div>
      </div>
    </div>
  )
}
