import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { getDuplicates } from '../api/client'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

export default function DuplicatesPage() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    getDuplicates()
      .then(res => setData(res.data))
      .catch(() => setError('Failed to load duplicates report'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <PageSpinner message="Checking for duplicates..." />

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Duplicate Detection</h1>
      <ErrorBanner error={error} />

      {/* Duplicate Hostnames */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
        <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4">
          Duplicate Device Hostnames
          {data?.duplicateHostnames?.length > 0 && (
            <span className="ml-2 px-2 py-0.5 rounded-full text-xs bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 font-normal normal-case">
              {data.duplicateHostnames.length} found
            </span>
          )}
        </h2>
        {!data?.duplicateHostnames?.length ? (
          <p className="text-sm text-green-600 dark:text-green-400">No duplicate hostnames found.</p>
        ) : (
          <table className="w-full text-sm">
            <thead className="text-left text-gray-500 dark:text-gray-400 border-b dark:border-gray-700">
              <tr>
                <th className="pb-2 font-medium">Hostname</th>
                <th className="pb-2 font-medium">Count</th>
                <th className="pb-2 font-medium">Device IDs</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {data.duplicateHostnames.map((d, i) => (
                <tr key={i} className="hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td className="py-2 font-mono text-gray-800 dark:text-gray-200">{d.hostname}</td>
                  <td className="py-2">
                    <span className="px-2 py-0.5 rounded-full text-xs bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">
                      {d.count}
                    </span>
                  </td>
                  <td className="py-2 text-gray-500 dark:text-gray-400 text-xs">
                    {d.deviceIds?.map(id => (
                      <Link
                        key={id}
                        to={`/devices/${id}`}
                        className="mr-2 text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        #{id}
                      </Link>
                    ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Conflicting IPs */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
        <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4">
          Conflicting IP Assignments
          {data?.conflictingIps?.length > 0 && (
            <span className="ml-2 px-2 py-0.5 rounded-full text-xs bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 font-normal normal-case">
              {data.conflictingIps.length} found
            </span>
          )}
        </h2>
        {!data?.conflictingIps?.length ? (
          <p className="text-sm text-green-600 dark:text-green-400">No conflicting IP assignments found.</p>
        ) : (
          <table className="w-full text-sm">
            <thead className="text-left text-gray-500 dark:text-gray-400 border-b dark:border-gray-700">
              <tr>
                <th className="pb-2 font-medium">IP Address</th>
                <th className="pb-2 font-medium">Subnet</th>
                <th className="pb-2 font-medium">Assigned Hostnames</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {data.conflictingIps.map((c, i) => (
                <tr key={i} className="hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td className="py-2 font-mono text-gray-800 dark:text-gray-200">{c.ipAddress}</td>
                  <td className="py-2 font-mono text-gray-500 dark:text-gray-400 text-xs">{c.subnetCidr}</td>
                  <td className="py-2 text-gray-500 dark:text-gray-400 text-xs">
                    {c.hostnames?.filter(Boolean).join(', ') || '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
