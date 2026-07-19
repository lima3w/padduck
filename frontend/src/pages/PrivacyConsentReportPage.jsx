import { useEffect, useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { listPrivacyVersions, createPrivacyVersion, getConsentReport } from '../api/admin'

function formatDate(str) {
  if (!str) return '—'
  return new Date(str).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function ConsentBadge({ hasConsent }) {
  const { t } = useTranslation()
  return hasConsent ? (
    <span className="px-2 py-0.5 rounded text-xs font-semibold bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-300">
      {t('privacyConsentReport.consented')}
    </span>
  ) : (
    <span className="px-2 py-0.5 rounded text-xs font-semibold bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300">
      {t('privacyConsentReport.noConsent')}
    </span>
  )
}

function AddVersionModal({ onClose, onCreated }) {
  const { t } = useTranslation()
  const [form, setForm] = useState({ version: '', effective_date: '', summary: '' })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)

  function handleChange(e) {
    setForm(f => ({ ...f, [e.target.name]: e.target.value }))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    if (!form.version.trim()) { setError(t('privacyConsentReport.versionRequired')); return }
    if (!form.effective_date) { setError(t('privacyConsentReport.effectiveDateRequired')); return }
    setSaving(true)
    try {
      const payload = {
        version: form.version.trim(),
        effective_date: form.effective_date,
        summary: form.summary.trim() || undefined,
      }
      const res = await createPrivacyVersion(payload)
      onCreated(res.data?.version)
      onClose()
    } catch (err) {
      setError(err.response?.data?.error || t('privacyConsentReport.createVersionFailed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{t('privacyConsentReport.addModalTitle')}</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
          >
            ✕
          </button>
        </div>
        <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
          {error && (
            <div className="rounded bg-red-50 dark:bg-red-900/30 px-3 py-2 text-sm text-red-700 dark:text-red-300">
              {error}
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('privacyConsentReport.versionLabel')} <span className="text-red-500">*</span>
            </label>
            <input
              name="version"
              value={form.version}
              onChange={handleChange}
              placeholder={t('privacyConsentReport.versionPlaceholder')}
              className="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('privacyConsentReport.effectiveDateLabel')} <span className="text-red-500">*</span>
            </label>
            <input
              type="date"
              name="effective_date"
              value={form.effective_date}
              onChange={handleChange}
              className="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('privacyConsentReport.summaryLabel')}
            </label>
            <textarea
              name="summary"
              value={form.summary}
              onChange={handleChange}
              rows={3}
              placeholder={t('privacyConsentReport.summaryPlaceholder')}
              className="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              {saving ? t('common.saving') : t('privacyConsentReport.addVersion')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function PrivacyConsentReportPage() {
  const { t } = useTranslation()
  const [versions, setVersions] = useState([])
  const [users, setUsers] = useState([])
  const [noConsentCount, setNoConsentCount] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showModal, setShowModal] = useState(false)

  const fetchAll = useCallback(() => {
    setLoading(true)
    setError(null)
    Promise.all([listPrivacyVersions(), getConsentReport()])
      .then(([vRes, cRes]) => {
        setVersions(vRes.data?.versions ?? [])
        setUsers(cRes.data?.users ?? [])
        setNoConsentCount(cRes.data?.noConsentCount ?? 0)
      })
      .catch(() => setError(t('privacyConsentReport.loadError')))
      .finally(() => setLoading(false))
  }, [t])

  useEffect(() => {
    fetchAll()
  }, [fetchAll])

  function handleVersionCreated(v) {
    if (v) setVersions(prev => [v, ...prev])
    else fetchAll()
  }

  if (loading) {
    return (
      <div className="flex min-h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
        {t('reconciliation.loading')}
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-6 text-center text-sm text-red-700 dark:text-red-300">
          {error}
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-8">
      {showModal && (
        <AddVersionModal
          onClose={() => setShowModal(false)}
          onCreated={handleVersionCreated}
        />
      )}

      {/* Policy Versions Panel */}
      <div>
        <div className="mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('privacyConsentReport.title')}</h1>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {t('privacyConsentReport.subtitle')}
            </p>
          </div>
          <button
            onClick={() => setShowModal(true)}
            className="px-4 py-2 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition-colors"
          >
            {t('privacyConsentReport.addVersionButton')}
          </button>
        </div>

        {versions.length === 0 ? (
          <div className="rounded bg-white dark:bg-gray-800 shadow px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
            {t('privacyConsentReport.noPolicyVersions')}
          </div>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.versionLabel')}</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.effectiveDateLabel')}</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.summaryLabel')}</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.createdColumn')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                  {versions.map(v => (
                    <tr key={v.id} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                      <td className="px-4 py-3 font-semibold text-gray-900 dark:text-gray-100">{v.version}</td>
                      <td className="px-4 py-3 text-gray-700 dark:text-gray-300">{formatDate(v.effectiveDate)}</td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-md">
                        {v.summary || <span className="text-gray-400 dark:text-gray-600">—</span>}
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                        {formatDate(v.createdAt)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>

      {/* User Consent Status Panel */}
      <div>
        <div className="mb-4">
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">{t('privacyConsentReport.userConsentStatusTitle')}</h2>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {t('privacyConsentReport.userConsentStatusSubtitle')}
          </p>
        </div>

        {noConsentCount > 0 && (
          <div className="mb-4 rounded bg-yellow-50 dark:bg-yellow-900/30 px-4 py-3 text-sm text-yellow-800 dark:text-yellow-300">
            {t('privacyConsentReport.noConsentWarning', { count: noConsentCount })}
          </div>
        )}

        {users.length === 0 ? (
          <div className="rounded bg-white dark:bg-gray-800 shadow px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
            {t('privacyConsentReport.noUsersFound')}
          </div>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.usernameColumn')}</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('customers.email')}</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.consentStatusColumn')}</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.versionAcceptedColumn')}</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.dateAcceptedColumn')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                  {users.map(u => (
                    <tr key={u.userId} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                      <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{u.username}</td>
                      <td className="px-4 py-3 text-gray-700 dark:text-gray-300">{u.email}</td>
                      <td className="px-4 py-3"><ConsentBadge hasConsent={u.hasConsent} /></td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                        {u.privacyAcceptedVersion || <span className="text-gray-400 dark:text-gray-600">—</span>}
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                        {u.privacyAcceptedAt ? formatDate(u.privacyAcceptedAt) : <span className="text-gray-400 dark:text-gray-600">—</span>}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
