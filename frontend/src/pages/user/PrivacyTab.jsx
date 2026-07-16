import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import * as client from '../../api/auth'
import { getCachedUser, setCachedUser } from '../../utils/storageKeys'

export default function PrivacyTab({ user }) {
  const { t } = useTranslation()
  const [policyVersion, setPolicyVersion] = useState(null)
  const [acceptedVersion, setAcceptedVersion] = useState(user?.privacyAcceptedVersion || user?.privacy_accepted_version || null)
  const [loading, setLoading] = useState(true)
  const [accepting, setAccepting] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    client.getPrivacyPolicyVersion()
      .then((res) => {
        if (!cancelled) setPolicyVersion(res.data?.version || '1.0')
      })
      .catch(() => {
        if (!cancelled) setError(t('userTabs.privacy.loadError'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [])

  useEffect(() => {
    setAcceptedVersion(user?.privacyAcceptedVersion || user?.privacy_accepted_version || null)
  }, [user])

  const handleAccept = async () => {
    setAccepting(true)
    setError('')
    setSaved(false)
    try {
      await client.acceptPrivacyPolicy()
      const nextVersion = policyVersion || '1.0'
      const cached = getCachedUser() || {}
      setCachedUser({
        ...cached,
        privacyAcceptedVersion: nextVersion,
        privacy_accepted_version: undefined,
      })
      setAcceptedVersion(nextVersion)
      setSaved(true)
    } catch {
      setError(t('userTabs.privacy.recordError'))
    } finally {
      setAccepting(false)
    }
  }

  if (loading) return <p className="text-sm text-gray-500">{t('common.loading')}</p>

  const currentAccepted = acceptedVersion && acceptedVersion === policyVersion

  return (
    <div className="max-w-lg space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">{t('userTabs.privacy.title')}</h2>
        <p className="text-sm text-gray-600">
          {t('userTabs.privacy.subtitle')}
        </p>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      <dl className="divide-y divide-gray-200 border border-gray-200 rounded">
        <div className="flex items-center justify-between gap-4 px-4 py-3">
          <dt className="text-sm font-medium text-gray-600">{t('userTabs.privacy.currentVersion')}</dt>
          <dd className="text-sm text-gray-900">{policyVersion || t('userTabs.privacy.unknown')}</dd>
        </div>
        <div className="flex items-center justify-between gap-4 px-4 py-3">
          <dt className="text-sm font-medium text-gray-600">{t('userTabs.privacy.acceptedVersion')}</dt>
          <dd className="text-sm text-gray-900">{acceptedVersion || t('userTabs.privacy.notAccepted')}</dd>
        </div>
        <div className="flex items-center justify-between gap-4 px-4 py-3">
          <dt className="text-sm font-medium text-gray-600">{t('userTabs.privacy.status')}</dt>
          <dd className={currentAccepted ? 'text-sm font-medium text-green-700' : 'text-sm font-medium text-yellow-700'}>
            {currentAccepted ? t('userTabs.privacy.current') : t('userTabs.privacy.actionRequired')}
          </dd>
        </div>
      </dl>

      {!currentAccepted && (
        <button
          type="button"
          onClick={handleAccept}
          disabled={accepting || !policyVersion}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {accepting ? t('userTabs.privacy.accepting') : t('userTabs.privacy.acceptCurrentPolicy')}
        </button>
      )}
      {saved && <p className="text-sm text-green-600">{t('userTabs.privacy.consentRecorded')}</p>}
    </div>
  )
}
