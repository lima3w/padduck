import { useEffect, useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { getIdentityPolicies, updateIdentityPolicies, getSessionRisk } from '../api/admin'

function formatDatetime(str) {
  if (!str) return '—'
  return new Date(str).toLocaleString(undefined, {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

export default function IdentityPoliciesPage() {
  const { t } = useTranslation()
  // Policy form state
  const [form, setForm] = useState({
    enforce_mfa: false,
    session_max_age_hours: 24,
    api_token_max_age_days: 0,
    inactive_user_days: 0,
  })
  const [loadingPolicies, setLoadingPolicies] = useState(true)
  const [savingPolicies, setSavingPolicies] = useState(false)
  const [policyError, setPolicyError] = useState(null)
  const [policySuccess, setPolicySuccess] = useState(false)

  // Session risk state
  const [sessions, setSessions] = useState([])
  const [loadingSessions, setLoadingSessions] = useState(true)
  const [sessionError, setSessionError] = useState(null)

  useEffect(() => {
    setLoadingPolicies(true)
    getIdentityPolicies()
      .then(res => {
        const d = res.data ?? {}
        const loaded = {
          enforce_mfa: d.enforceMfa ?? false,
          session_max_age_hours: d.sessionMaxAgeHours ?? 24,
          api_token_max_age_days: d.apiTokenMaxAgeDays ?? 0,
          inactive_user_days: d.inactiveUserDays ?? 0,
        }
        setForm(loaded)
      })
      .catch(() => setPolicyError(t('identityPolicies.loadPoliciesFailed')))
      .finally(() => setLoadingPolicies(false))
  }, [t])

  const fetchSessions = useCallback(() => {
    setLoadingSessions(true)
    setSessionError(null)
    getSessionRisk()
      .then(res => setSessions(res.data?.sessions ?? []))
      .catch(() => setSessionError(t('identityPolicies.loadSessionsFailed')))
      .finally(() => setLoadingSessions(false))
  }, [t])

  useEffect(() => {
    fetchSessions()
  }, [fetchSessions])

  function handleChange(e) {
    const { name, value, type, checked } = e.target
    setForm(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : Number(value),
    }))
  }

  async function handleSave(e) {
    e.preventDefault()
    setPolicyError(null)
    setPolicySuccess(false)
    setSavingPolicies(true)
    try {
      await updateIdentityPolicies({
        enforce_mfa: form.enforce_mfa,
        session_max_age_hours: form.session_max_age_hours,
        api_token_max_age_days: form.api_token_max_age_days,
        inactive_user_days: form.inactive_user_days,
      })
      setPolicySuccess(true)
      setTimeout(() => setPolicySuccess(false), 3000)
    } catch (err) {
      setPolicyError(err.response?.data?.error || t('identityPolicies.savePoliciesFailed'))
    } finally {
      setSavingPolicies(false)
    }
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('identityPolicies.title')}</h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {t('identityPolicies.subtitle')}
        </p>
      </div>

      {/* Identity Policies Panel */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">{t('identityPolicies.policySettingsTitle')}</h2>

        {loadingPolicies ? (
          <div className="text-sm text-gray-500 dark:text-gray-400">{t('common.loading')}</div>
        ) : (
          <form onSubmit={handleSave} className="space-y-6">
            {policyError && (
              <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-red-700 dark:text-red-300">
                {policyError}
              </div>
            )}
            {policySuccess && (
              <div className="rounded bg-green-50 dark:bg-green-900/30 px-4 py-3 text-sm text-green-700 dark:text-green-300">
                {t('identityPolicies.policiesSaved')}
              </div>
            )}

            {/* Enforce MFA */}
            <div className="flex items-center gap-3">
              <input
                id="enforce_mfa"
                name="enforce_mfa"
                type="checkbox"
                checked={form.enforce_mfa}
                onChange={handleChange}
                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <label htmlFor="enforce_mfa" className="text-sm font-medium text-gray-700 dark:text-gray-300">
                {t('identityPolicies.enforceMfa')}
              </label>
            </div>

            {/* Session Max Age */}
            <div>
              <label htmlFor="session_max_age_hours" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('identityPolicies.sessionMaxAge')}
                <span className="ml-1 text-gray-400 dark:text-gray-500 font-normal">{t('identityPolicies.sessionMaxAgeRange')}</span>
              </label>
              <input
                id="session_max_age_hours"
                name="session_max_age_hours"
                type="number"
                min={1}
                max={8760}
                value={form.session_max_age_hours}
                onChange={handleChange}
                className="w-40 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* API Token Max Age */}
            <div>
              <label htmlFor="api_token_max_age_days" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('identityPolicies.apiTokenMaxAge')}
                <span className="ml-1 text-gray-400 dark:text-gray-500 font-normal">{t('identityPolicies.noLimit')}</span>
              </label>
              <input
                id="api_token_max_age_days"
                name="api_token_max_age_days"
                type="number"
                min={0}
                max={3650}
                value={form.api_token_max_age_days}
                onChange={handleChange}
                className="w-40 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Inactive User Days */}
            <div>
              <label htmlFor="inactive_user_days" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('identityPolicies.inactiveUserAutoDisable')}
                <span className="ml-1 text-gray-400 dark:text-gray-500 font-normal">{t('identityPolicies.noLimit')}</span>
              </label>
              <input
                id="inactive_user_days"
                name="inactive_user_days"
                type="number"
                min={0}
                max={3650}
                value={form.inactive_user_days}
                onChange={handleChange}
                className="w-40 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <button
              type="submit"
              disabled={savingPolicies}
              className="px-4 py-2 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              {savingPolicies ? t('common.saving') : t('identityPolicies.savePolicies')}
            </button>
          </form>
        )}
      </div>

      {/* Active Session Risk Panel */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{t('identityPolicies.activeSessionRiskTitle')}</h2>
          <button
            onClick={fetchSessions}
            disabled={loadingSessions}
            className="px-3 py-1.5 rounded text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors"
          >
            {loadingSessions ? t('identityPolicies.refreshing') : t('dashboard.refresh')}
          </button>
        </div>

        {sessionError && (
          <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-red-700 dark:text-red-300 mb-4">
            {sessionError}
          </div>
        )}

        {loadingSessions ? (
          <div className="text-sm text-gray-500 dark:text-gray-400">{t('identityPolicies.loadingSessions')}</div>
        ) : sessions.length === 0 ? (
          <div className="text-sm text-gray-500 dark:text-gray-400 py-6 text-center">{t('identityPolicies.noActiveSessions')}</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('auditLog.user')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('myRequests.ipType')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('privacyConsentReport.createdColumn')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('identityPolicies.lastUsed')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('identityPolicies.riskFlags')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                {sessions.map((s, i) => {
                  const hasRisk = s.riskFlags && s.riskFlags.length > 0
                  return (
                    <tr
                      key={i}
                      className={hasRisk
                        ? 'bg-yellow-50 dark:bg-yellow-900/20 hover:bg-yellow-100 dark:hover:bg-yellow-900/30'
                        : 'hover:bg-gray-50 dark:hover:bg-gray-750'}
                    >
                      <td className="px-4 py-3 text-gray-900 dark:text-gray-100">
                        {s.username || t('identityPolicies.userFallback', { id: s.userId })}
                      </td>
                      <td className="px-4 py-3 text-gray-700 dark:text-gray-300 font-mono text-xs">
                        {s.ipAddress || '—'}
                      </td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                        {formatDatetime(s.createdAt)}
                      </td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                        {formatDatetime(s.lastUsedAt)}
                      </td>
                      <td className="px-4 py-3">
                        {hasRisk ? (
                          <div className="flex flex-wrap gap-1">
                            {s.riskFlags.map(flag => (
                              <span
                                key={flag}
                                className="px-2 py-0.5 rounded text-xs font-semibold bg-yellow-100 text-yellow-800 dark:bg-yellow-800/40 dark:text-yellow-300"
                              >
                                {flag}
                              </span>
                            ))}
                          </div>
                        ) : (
                          <span className="text-gray-400 dark:text-gray-500 text-xs">—</span>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
