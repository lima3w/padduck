import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getDnsZoneRecords } from '../api/dns'

const RECORD_TYPES = ['All', 'A', 'AAAA', 'PTR', 'CNAME', 'MX']

export default function DnsZoneDetailPage() {
  const { zone } = useParams()
  const [records, setRecords] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [typeFilter, setTypeFilter] = useState('All')

  useEffect(() => {
    load(typeFilter)
  }, [zone])

  async function load(type) {
    try {
      setLoading(true)
      setError(null)
      const res = await getDnsZoneRecords(zone, type === 'All' ? '' : type)
      const data = res.data
      setRecords(Array.isArray(data) ? data : (data?.records ?? []))
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load zone records')
    } finally {
      setLoading(false)
    }
  }

  function handleTypeChange(type) {
    setTypeFilter(type)
    load(type)
  }

  const filteredRecords = typeFilter === 'All'
    ? records
    : records.filter(r => r.type === typeFilter)

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/dns/zones" className="hover:text-blue-600">DNS Zones</Link>
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-mono font-medium">{zone}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100 font-mono">{zone}</h1>
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500 dark:text-gray-400">Filter by type:</span>
          <div className="flex rounded overflow-hidden border border-gray-300 dark:border-gray-600">
            {RECORD_TYPES.map(t => (
              <button
                key={t}
                onClick={() => handleTypeChange(t)}
                className={`px-3 py-1.5 text-xs font-medium transition border-l first:border-l-0 border-gray-300 dark:border-gray-600 ${
                  typeFilter === t
                    ? 'bg-blue-600 text-white'
                    : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      {loading ? (
        <p className="text-gray-500">Loading records...</p>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">TTL</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Content / Value</th>
              </tr>
            </thead>
            <tbody>
              {filteredRecords.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-6 text-center text-gray-400">
                    No records found{typeFilter !== 'All' ? ` for type ${typeFilter}` : ''}.
                  </td>
                </tr>
              )}
              {filteredRecords.map((record, idx) => (
                <tr key={`${record.name}-${record.type}-${idx}`} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td className="px-4 py-3">
                    <span className="inline-block px-2 py-0.5 rounded text-xs font-mono font-medium bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">
                      {record.type}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-mono text-gray-700 dark:text-gray-300">{record.name}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{record.ttl ?? '—'}</td>
                  <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">
                    {(record.type === 'A' || record.type === 'AAAA') && record.content ? (
                      <Link
                        to={`/ip-addresses?highlight=${encodeURIComponent(record.content)}`}
                        className="text-blue-600 dark:text-blue-400 hover:underline"
                        title="View this IP in IPAM"
                      >
                        {record.content}
                      </Link>
                    ) : (
                      record.content || '—'
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          </div>
        </div>
      )}
    </div>
  )
}
