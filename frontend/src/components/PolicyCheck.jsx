import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { checkPolicy } from '../api/admin'

// Shows whether a write would pass the automation policies before the form is
// submitted: allowed, needs manual review, or blocked. The check is advisory;
// the server evaluates again on submit. Values are sent as strings.
export default function PolicyCheck({ workflow, action, values, enabled = true }) {
  const { t } = useTranslation()
  const [result, setResult] = useState(null)
  const seq = useRef(0)
  const key = JSON.stringify(values)

  useEffect(() => {
    if (!enabled) return undefined
    const mine = ++seq.current
    const timer = setTimeout(() => {
      const strValues = Object.fromEntries(
        Object.entries(JSON.parse(key)).filter(([, v]) => v !== '' && v != null).map(([k, v]) => [k, String(v)]),
      )
      checkPolicy(workflow, action, strValues)
        .then(res => { if (mine === seq.current) setResult(res.data) })
        .catch(() => { if (mine === seq.current) setResult(null) })
    }, 400)
    return () => clearTimeout(timer)
  }, [workflow, action, key, enabled])

  if (!enabled || !result) return null

  const styles = {
    allow: 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800 dark:text-green-300',
    manual_review: 'bg-amber-50 border-amber-200 text-amber-800 dark:bg-amber-900/20 dark:border-amber-800 dark:text-amber-300',
    deny: 'bg-red-50 border-red-200 text-red-700 dark:bg-red-900/20 dark:border-red-800 dark:text-red-300',
  }
  const text = {
    allow: result.bypassed ? t('policyCheck.bypassed') : t('policyCheck.passed'),
    manual_review: t('policyCheck.review', { message: result.message || '' }),
    deny: t('policyCheck.blocked', { message: result.message || '' }),
  }[result.result]
  if (!text) return null

  return (
    <div role="status" className={`p-3 border rounded text-sm ${styles[result.result]}`}>
      {text}
    </div>
  )
}
