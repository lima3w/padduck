import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { getDnsZones } from '../api/client'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

export default function DnsZonesPage() {
  const navigate = useNavigate()
  const [zones, setZones] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [configured, setConfigured] = useState(true)

  useEffect(() => {
    load()
  }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const res = await getDnsZones()
      const data = res.data
      if (data?.configured === false) {
        setConfigured(false)
        setZones([])
      } else {
        setConfigured(true)
        setZones(Array.isArray(data) ? data : (data?.zones ?? []))
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load DNS zones')
    } finally {
      setLoading(false)
    }
  }

  if (loading) return <PageSpinner message="Loading DNS zones..." />

  if (!configured) {
    return (
      <div>
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-4">DNS Zones</h1>
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded-lg p-6 text-center">
          <p className="text-yellow-800 dark:text-yellow-200 font-medium mb-2">No DNS provider is configured.</p>
          <p className="text-yellow-700 dark:text-yellow-300 text-sm mb-4">
            Set up PowerDNS or Technitium in{' '}
            <Link to="/admin/settings" className="underline hover:text-yellow-900 dark:hover:text-yellow-100">
              Admin Settings &rarr; DNS
            </Link>
            .
          </p>
        </div>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-4">DNS Zones</h1>

      <ErrorBanner error={error} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Zone Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Kind</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Serial</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {zones.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-gray-400">
                  No DNS zones found.
                </td>
              </tr>
            )}
            {zones.map(zone => (
              <tr
                key={zone.id || zone.name}
                className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30 cursor-pointer"
                onClick={() => navigate(`/dns/zones/${encodeURIComponent(zone.name || zone.id)}`)}
              >
                <td className="px-4 py-3 font-mono font-medium text-blue-600 dark:text-blue-400">
                  {zone.name || zone.id}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400 capitalize">{zone.kind || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{zone.serial ?? '—'}</td>
                <td className="px-4 py-3 text-right">
                  <span className="text-blue-600 dark:text-blue-400 text-xs">View &rarr;</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
