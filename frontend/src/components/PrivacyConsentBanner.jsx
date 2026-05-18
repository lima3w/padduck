import { useState, useEffect } from 'react'
import { getPrivacyPolicyVersion, acceptPrivacyPolicy, getCurrentUser } from '../api/client'

export default function PrivacyConsentBanner() {
  const [policyVersion, setPolicyVersion] = useState(null)
  const [userAcceptedVersion, setUserAcceptedVersion] = useState(undefined)
  const [accepting, setAccepting] = useState(false)
  const [error, setError] = useState(null)
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    const cached = localStorage.getItem('current_user')
    if (!cached) return // not authenticated

    try {
      const u = JSON.parse(cached)
      setUserAcceptedVersion(u.privacyAcceptedVersion || null)
    } catch {
      return
    }

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
      // Update cached user with new accepted version
      const res = await getCurrentUser()
      localStorage.setItem('current_user', JSON.stringify(res.data))
      setDismissed(true)
    } catch {
      setError('Failed to record consent. Please try again.')
    } finally {
      setAccepting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Privacy Policy</h2>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          {userAcceptedVersion
            ? `Our privacy policy has been updated to version ${policyVersion}. Please review and accept it to continue.`
            : `Please accept our privacy policy (version ${policyVersion}) to continue using this application.`}
        </p>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          By clicking &quot;Accept&quot;, you agree to our privacy policy regarding the collection and use of
          your data within this IPAM system.
        </p>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex justify-end pt-2">
          <button
            onClick={handleAccept}
            disabled={accepting}
            className="px-5 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {accepting ? 'Accepting…' : 'Accept Privacy Policy'}
          </button>
        </div>
      </div>
    </div>
  )
}
