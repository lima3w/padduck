import { useQuery } from '@tanstack/react-query'
import { getVlanUsageReport } from '../api/vlans'

function UtilBar({ pct }) {
  const clamped = Math.max(0, Math.min(100, pct || 0))
  let barClass = 'bg-green-500'
  if (clamped >= 90) barClass = 'bg-red-500'
  else if (clamped >= 75) barClass = 'bg-yellow-500'
  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 h-2 bg-gray-200 dark:bg-gray-600 rounded-full overflow-hidden min-w-16">
        <div
          className={`h-full rounded-full ${barClass}`}
          style={{ width: `${clamped}%` }}
        />
      </div>
      <span className="text-xs text-gray-600 dark:text-gray-400 w-10 text-right">
        {clamped.toFixed(1)}%
      </span>
    </div>
  )
}

export default function VlanUsageReportPage() {
  const usageQuery = useQuery({
    queryKey: ['vlans', 'usage-report'],
    queryFn: () => getVlanUsageReport().then(r => {
      const data = r.data
      return Array.isArray(data) ? data : (data?.vlans ?? data?.report ?? [])
    }),
  })
  const rows = usageQuery.data ?? []
  const loading = usageQuery.isLoading
  const error = usageQuery.isError
    ? (usageQuery.error?.response?.data?.error || 'Failed to load VLAN usage report')
    : null

  if (loading) return <p className="text-gray-500">Loading VLAN usage report...</p>

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">VLAN Usage Report</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Subnet and IP utilisation across all VLANs
          </p>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-3 rounded text-sm bg-red-50 text-red-700 border border-red-200">
          {error}
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        {rows.length === 0 && !error ? (
          <div className="px-6 py-12 text-center text-gray-400">
            <p className="text-lg font-medium mb-1">No data available</p>
            <p className="text-sm">No VLANs have been configured yet.</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">VLAN ID</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Domain</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Group</th>
                <th className="text-right px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Subnets</th>
                <th className="text-right px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">IPs</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium w-48">Utilisation</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row, idx) => {
                return (
                  <tr key={row.id ?? idx} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">
                      {row.vlanId}
                    </td>
                    <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">
                      {row.name}
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {row.domainName || '—'}
                    </td>
                    <td className="px-4 py-3">
                      {row.groupName ? (
                        <span
                          className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium text-white"
                          style={{ backgroundColor: row.groupColour || '#6B7280' }}
                        >
                          {row.groupName}
                        </span>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right text-gray-700 dark:text-gray-300">
                      {row.subnetCount ?? 0}
                    </td>
                    <td className="px-4 py-3 text-right text-gray-700 dark:text-gray-300">
                      {row.ipCount ?? 0}
                    </td>
                    <td className="px-4 py-3">
                      <UtilBar pct={row.utilisationPct ?? 0} />
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
