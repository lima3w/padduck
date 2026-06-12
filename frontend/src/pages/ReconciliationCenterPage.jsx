import { useEffect, useState } from 'react'
import { getReconciliationReport } from '../api/admin'

function Panel({ title, count, children }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden mb-4">
      <div className="px-4 py-3 border-b dark:border-gray-700 flex items-center justify-between">
        <h2 className="font-semibold text-gray-800 dark:text-gray-100">{title}</h2>
        <span
          className={`px-2 py-0.5 rounded text-xs font-medium ${
            count > 0 ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'
          }`}
        >
          {count > 0 ? `${count} issue${count !== 1 ? 's' : ''}` : 'Clean'}
        </span>
      </div>
      {children}
    </div>
  )
}

function EmptyState() {
  return (
    <p className="px-4 py-6 text-center text-sm text-green-600 dark:text-green-400">
      No issues found
    </p>
  )
}

function StaleIPsPanel({ items }) {
  return (
    <Panel title="Stale IP Assignments" count={items.length}>
      {items.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">IP Address</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Hostname</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Subnet</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Last Seen</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Days Inactive</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {items.map((ip, i) => (
                <tr key={ip.ipId ?? i} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                  <td className="px-4 py-2 font-mono text-gray-900 dark:text-gray-100">{ip.ipAddress}</td>
                  <td className="px-4 py-2 text-gray-700 dark:text-gray-300">{ip.hostname || '—'}</td>
                  <td className="px-4 py-2 text-gray-700 dark:text-gray-300">{ip.subnetCidr}</td>
                  <td className="px-4 py-2 text-gray-700 dark:text-gray-300">
                    {ip.lastSeen ? new Date(ip.lastSeen).toLocaleDateString() : 'Never'}
                  </td>
                  <td className="px-4 py-2 text-red-600 dark:text-red-400 font-medium">{ip.daysInactive}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  )
}

function DNSDriftPanel({ items }) {
  return (
    <Panel title="DNS Drift" count={items.length}>
      {items.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">IP Address</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">DNS Name</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">PTR Record</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Last Checked</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {items.map((entry, i) => (
                <tr key={entry.ipId ?? i} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                  <td className="px-4 py-2 font-mono text-gray-900 dark:text-gray-100">{entry.address}</td>
                  <td className="px-4 py-2 text-gray-700 dark:text-gray-300">{entry.dnsName || '—'}</td>
                  <td className="px-4 py-2 text-gray-700 dark:text-gray-300">{entry.ptrRecord || '—'}</td>
                  <td className="px-4 py-2 text-gray-700 dark:text-gray-300">{entry.dnsLastChecked || 'never'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  )
}

function SubnetOverlapsPanel({ items }) {
  return (
    <Panel title="Subnet Overlaps" count={items.length}>
      {items.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Subnet A</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">CIDR A</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Subnet B</th>
                <th className="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">CIDR B</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {items.map((pair, i) => {
                const a = pair.subnetA
                const b = pair.subnetB
                return (
                  <tr key={i} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                    <td className="px-4 py-2 text-gray-700 dark:text-gray-300">{a?.description || a?.networkAddress}</td>
                    <td className="px-4 py-2 font-mono text-gray-900 dark:text-gray-100">
                      {a?.networkAddress}/{a?.prefixLength}
                    </td>
                    <td className="px-4 py-2 text-gray-700 dark:text-gray-300">{b?.description || b?.networkAddress}</td>
                    <td className="px-4 py-2 font-mono text-gray-900 dark:text-gray-100">
                      {b?.networkAddress}/{b?.prefixLength}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  )
}

export default function ReconciliationCenterPage() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    getReconciliationReport()
      .then(res => setData(res.data))
      .catch(err => setError(err?.response?.data?.error || 'Failed to load reconciliation data'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="max-w-6xl mx-auto px-4 py-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Reconciliation Center</h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Operational drift across stale IP assignments, DNS records, and subnet overlaps.
        </p>
      </div>

      {loading && (
        <p className="text-sm text-gray-500 dark:text-gray-400">Loading...</p>
      )}

      {error && (
        <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 px-4 py-3 text-sm text-red-700 dark:text-red-400 mb-4">
          {error}
        </div>
      )}

      {data && (
        <>
          <StaleIPsPanel items={data.staleIps ?? []} />
          <DNSDriftPanel items={data.dnsDrift ?? []} />
          <SubnetOverlapsPanel items={data.overlaps ?? []} />
        </>
      )}
    </div>
  )
}
