import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getPrivacyPolicyVersion, acceptPrivacyPolicy, getCurrentUser } from '../api/auth'
import { getCachedUser, setCachedUser } from '../utils/storageKeys'

export default function PrivacyConsentBanner() {
  const { t } = useTranslation()
  const [policyVersion, setPolicyVersion] = useState(null)
  const [userAcceptedVersion, setUserAcceptedVersion] = useState(undefined)
  const [accepting, setAccepting] = useState(false)
  const [error, setError] = useState(null)
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    const user = getCachedUser()
    if (!user) return // not authenticated
    setUserAcceptedVersion(user.privacyAcceptedVersion || null)

    getPrivacyPolicyVersion()
      .then((res) => setPolicyVersion(res.data?.version || '1.0'))
      .catch(() => setPolicyVersion('1.0'))
  }, [])

  if (dismissed || policyVersion === null || userAcceptedVersion === undefined) return null

  if (userAcceptedVersion === policyVersion) return null

  const handleAccept = async () => {
    setAccepting(true)
    setError(null)
    try {
      await acceptPrivacyPolicy()
      const cached = getCachedUser() || {}
      setCachedUser({
        ...cached,
        privacyAcceptedVersion: policyVersion,
        privacy_accepted_version: undefined,
      })
      setUserAcceptedVersion(policyVersion)
      setDismissed(true)
      getCurrentUser()
        .then((res) => setCachedUser(res.data))
        .catch(() => {})
    } catch {
      setError(t('privacyConsent.acceptError'))
    } finally {
      setAccepting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{t('privacyConsent.title')}</h2>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          {userAcceptedVersion
            ? t('privacyConsent.updatedNotice', { version: policyVersion })
            : t('privacyConsent.acceptNotice', { version: policyVersion })}
        </p>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          {t('privacyConsent.agreementText')}
        </p>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex justify-end pt-2">
          <button
            onClick={handleAccept}
            disabled={accepting}
            className="px-5 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {accepting ? t('privacyConsent.accepting') : t('privacyConsent.acceptPrivacyPolicy')}
          </button>
        </div>
      </div>
    </div>
  )
}
