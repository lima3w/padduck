import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { getOverlapReport } from '../api/client'

export default function OverlapReportPage() {
  const [overlaps, setOverlaps] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function load() {
      try {
        const res = await getOverlapReport()
        setOverlaps(res.data.overlaps || [])
      } catch {
        setError('Failed to load overlap report')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) return <p className="text-gray-500">Loading overlap report...</p>

  return (
    <div className="max-w-5xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Subnet Overlap Report</h1>
          <p className="text-sm text-gray-500 mt-1">All overlapping subnet pairs across the system</p>
        </div>
        <Link to="/admin/settings" className="text-sm text-blue-600 hover:underline">Back to Settings</Link>
      </div>

      {error && <p className="text-red-600 mb-4 text-sm">{error}</p>}

      {overlaps.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center text-gray-400">
          <p className="text-lg font-medium text-green-600 mb-1">No overlaps detected</p>
          <p className="text-sm">All subnets are properly separated.</p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="px-4 py-3 bg-red-50 border-b border-red-100 text-sm text-red-700 font-medium">
            {overlaps.length} overlapping pair{overlaps.length !== 1 ? 's' : ''} found
          </div>
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="text-left px-4 py-3 text-gray-600 font-medium">Subnet A</th>
                <th className="text-left px-4 py-3 text-gray-600 font-medium">Subnet B</th>
              </tr>
            </thead>
            <tbody>
              {overlaps.map((pair, i) => (
                <tr key={i} className="border-b last:border-0 hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link
                      to={`/subnets/${pair.subnet_a.ID}/ip-addresses`}
                      className="font-mono text-blue-600 hover:underline font-medium"
                    >
                      {pair.subnet_a.NetworkAddress}/{pair.subnet_a.PrefixLength}
                    </Link>
                    {pair.subnet_a.Description && (
                      <span className="ml-2 text-gray-400 text-xs">{pair.subnet_a.Description}</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <Link
                      to={`/subnets/${pair.subnet_b.ID}/ip-addresses`}
                      className="font-mono text-blue-600 hover:underline font-medium"
                    >
                      {pair.subnet_b.NetworkAddress}/{pair.subnet_b.PrefixLength}
                    </Link>
                    {pair.subnet_b.Description && (
                      <span className="ml-2 text-gray-400 text-xs">{pair.subnet_b.Description}</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
