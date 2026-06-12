import { useEffect, useMemo, useState } from 'react'
import { getV2CompatibilityWarnings, getV2DeprecationReport, getV2MigrationReadiness } from '../api/admin'
import { downloadFile } from '../utils/download'

const STATUS_STYLES = {
  pass: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-200',
  warn: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-200',
  fail: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-200',
}

const SEVERITY_STYLES = {
  info: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-200',
  warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-200',
  critical: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-200',
}

function Badge({ value, styles }) {
  return (
    <span className={`inline-flex min-w-16 justify-center rounded px-2 py-1 text-xs font-semibold uppercase ${styles[value] || styles.info || styles.pass}`}>
      {value}
    </span>
  )
}

function SummaryTile({ label, value }) {
  return (
    <div className="rounded border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
      <div className="text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">{label}</div>
      <div className="mt-2 text-2xl font-bold text-gray-900 dark:text-gray-100">{value ?? 0}</div>
    </div>
  )
}

export default function AdminCompatibilityPage() {
  const [readiness, setReadiness] = useState(null)
  const [warnings, setWarnings] = useState(null)
  const [deprecations, setDeprecations] = useState(null)
  const [error, setError] = useState('')
  const [downloading, setDownloading] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setError('')
      try {
        const [readinessRes, warningsRes, deprecationsRes] = await Promise.all([
          getV2MigrationReadiness(),
          getV2CompatibilityWarnings(),
          getV2DeprecationReport(),
        ])
        if (!cancelled) {
          setReadiness(readinessRes.data)
          setWarnings(warningsRes.data)
          setDeprecations(deprecationsRes.data)
        }
      } catch (err) {
        if (!cancelled) setError(err.response?.data?.error || err.message || 'Compatibility report failed')
      }
    }
    load()
    return () => { cancelled = true }
  }, [])

  const statusCounts = readiness?.summary?.byStatus || {}
  const severityCounts = warnings?.summary?.bySeverity || {}
  const sortedChecks = useMemo(() => {
    const order = { fail: 0, warn: 1, pass: 2 }
    return [...(readiness?.checks || [])].sort((a, b) => (order[a.status] ?? 9) - (order[b.status] ?? 9))
  }, [readiness])

  async function downloadBundle() {
    setDownloading(true)
    setError('')
    try {
      await downloadFile('/api/v1/admin/export/v2-migration-bundle', 'ipam-v2-migration-bundle.zip')
    } catch (err) {
      setError(err.message || 'Migration bundle download failed')
    } finally {
      setDownloading(false)
    }
  }

  return (
    <div className="w-full max-w-7xl mx-auto p-6">
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">V2 Compatibility</h1>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            Migration readiness, deprecation reporting, and v2 export preparation.
          </p>
        </div>
        <button
          type="button"
          onClick={downloadBundle}
          disabled={downloading}
          className="inline-flex items-center justify-center gap-2 rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {downloading ? (
            <span className="inline-block h-4 w-4 rounded-full border-2 border-white border-t-transparent animate-spin" />
          ) : (
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v12m0 0l4-4m-4 4l-4-4M4 17v1a3 3 0 003 3h10a3 3 0 003-3v-1" />
            </svg>
          )}
          Migration Bundle
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300">
          {error}
        </div>
      )}

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <SummaryTile label="Checks" value={readiness?.summary?.total} />
        <SummaryTile label="Warnings" value={(statusCounts.warn || 0) + (severityCounts.warning || 0)} />
        <SummaryTile label="Failures" value={statusCounts.fail} />
        <SummaryTile label="Deprecations" value={deprecations?.summary?.total} />
      </div>

      <network className="mt-8">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold uppercase text-gray-500 dark:text-gray-400">Migration Readiness</h2>
          {readiness?.summary && (
            <Badge value={readiness.summary.ready ? 'pass' : 'fail'} styles={STATUS_STYLES} />
          )}
        </div>
        <div className="overflow-hidden rounded border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-900/40">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">Status</th>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">Area</th>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">Check</th>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              {sortedChecks.map((check) => (
                <tr key={check.id}>
                  <td className="px-4 py-3 align-top"><Badge value={check.status} styles={STATUS_STYLES} /></td>
                  <td className="px-4 py-3 align-top text-sm text-gray-700 dark:text-gray-300">{check.area}</td>
                  <td className="px-4 py-3 align-top">
                    <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{check.summary}</div>
                    <div className="mt-1 text-sm text-gray-600 dark:text-gray-400">{check.detail}</div>
                    {check.signals?.length > 0 && (
                      <div className="mt-2 flex flex-wrap gap-2">
                        {check.signals.map((signal) => (
                          <span key={signal} className="rounded bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-gray-900 dark:text-gray-300">{signal}</span>
                        ))}
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-3 align-top text-sm text-gray-700 dark:text-gray-300">{check.recommendedWork || 'No action required.'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </network>

      <network className="mt-8">
        <h2 className="mb-3 text-sm font-semibold uppercase text-gray-500 dark:text-gray-400">Deprecation Report</h2>
        <div className="grid gap-3 lg:grid-cols-2">
          {(deprecations?.deprecations || []).map((item) => (
            <article key={item.id} className="rounded border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="text-sm font-semibold text-gray-900 dark:text-gray-100">{item.summary}</div>
                  <div className="mt-1 text-sm text-gray-600 dark:text-gray-400">{item.detail}</div>
                </div>
                <Badge value={item.severity} styles={SEVERITY_STYLES} />
              </div>
              <dl className="mt-4 grid gap-3 text-sm">
                <div>
                  <dt className="text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">V1 Surface</dt>
                  <dd className="mt-1 text-gray-800 dark:text-gray-200">{item.v1Surface}</dd>
                </div>
                <div>
                  <dt className="text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">V2 Change</dt>
                  <dd className="mt-1 text-gray-800 dark:text-gray-200">{item.v2Change}</dd>
                </div>
                <div>
                  <dt className="text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">Recommended Work</dt>
                  <dd className="mt-1 text-gray-800 dark:text-gray-200">{item.recommendedWork}</dd>
                </div>
              </dl>
            </article>
          ))}
        </div>
      </network>
    </div>
  )
}
