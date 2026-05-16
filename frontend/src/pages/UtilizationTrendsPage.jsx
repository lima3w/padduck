import { useState, useEffect } from 'react'
import { api } from '../api/client'

function DeltaBadge({ delta }) {
  if (delta == null) return <span className="text-gray-400">—</span>
  const isPositive = delta > 0
  const isNeutral = delta === 0
  return (
    <span className={`font-medium ${isNeutral ? 'text-gray-500' : isPositive ? 'text-red-600' : 'text-green-600'}`}>
      {isPositive ? '▲' : delta < 0 ? '▼' : ''} {Math.abs(delta).toFixed(1)}%
    </span>
  )
}

function PctBar({ pct }) {
  const colour = pct > 90 ? 'bg-red-500' : pct > 70 ? 'bg-yellow-500' : 'bg-green-500'
  return (
    <div className="flex items-center gap-2">
      <div className="w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
        <div className={`${colour} h-1.5 rounded-full`} style={{ width: `${Math.min(pct ?? 0, 100)}%` }} />
      </div>
      <span className="text-sm">{(pct ?? 0).toFixed(1)}%</span>
    </div>
  )
}

export default function UtilizationTrendsPage() {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [sortKey, setSortKey] = useState('currentPct')
  const [sortDir, setSortDir] = useState('desc')

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError('')
      const { data } = await api.get('/admin/reports/utilization-trends')
      setRows(Array.isArray(data) ? data : (data?.trends ?? []))
    } catch {
      setError('Failed to load utilization trends')
    } finally {
      setLoading(false)
    }
  }

  function toggleSort(key) {
    if (sortKey === key) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDir('desc')
    }
  }

  const sorted = [...rows].sort((a, b) => {
    const av = a[sortKey] ?? 0
    const bv = b[sortKey] ?? 0
    return sortDir === 'asc' ? av - bv : bv - av
  })

  function SortHeader({ col, label }) {
    const active = sortKey === col
    return (
      <th
        className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium cursor-pointer select-none hover:text-blue-600"
        onClick={() => toggleSort(col)}
      >
        {label} {active ? (sortDir === 'asc' ? '▲' : '▼') : ''}
      </th>
    )
  }

  if (loading) return <p className="text-gray-500">Loading utilization trends...</p>

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Utilization Trends</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Subnet utilization compared to 7 days ago</p>
        </div>
        <button onClick={load} className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">
          Refresh
        </button>
      </div>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Subnet</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <SortHeader col="currentPct" label="Current %" />
              <SortHeader col="weekAgoPct" label="Week Ago %" />
              <SortHeader col="deltaPct" label="Delta" />
            </tr>
          </thead>
          <tbody>
            {sorted.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-gray-400">No trend data available</td>
              </tr>
            )}
            {sorted.map((row) => (
              <tr key={row.subnetId} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-mono text-blue-600 dark:text-blue-400">{row.cidr}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{row.description || '—'}</td>
                <td className="px-4 py-3"><PctBar pct={row.currentPct} /></td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {row.weekAgoPct != null ? `${row.weekAgoPct.toFixed(1)}%` : '—'}
                </td>
                <td className="px-4 py-3"><DeltaBadge delta={row.deltaPct} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
