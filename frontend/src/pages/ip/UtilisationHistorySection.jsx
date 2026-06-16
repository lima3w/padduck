import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

const HISTORY_DAYS_OPTIONS = [7, 30, 90, 365]

export default function UtilisationHistorySection({ subnetId }) {
  const [historyDays, setHistoryDays] = useState(30)
  const [historyData, setHistoryData] = useState([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [historyError, setHistoryError] = useState('')

  useEffect(() => {
    async function fetchHistory() {
      setHistoryLoading(true)
      setHistoryError('')
      try {
        const { data } = await api.get(`/subnets/${subnetId}/utilization/history`, { params: { days: historyDays } })
        setHistoryData(Array.isArray(data) ? data : [])
      } catch {
        setHistoryData([])
        setHistoryError('Failed to load utilization history.')
      } finally {
        setHistoryLoading(false)
      }
    }
    if (subnetId) fetchHistory()
  }, [subnetId, historyDays])

  const chartData = historyData.map(d => ({
    date: new Date(d.recordedAt).toLocaleDateString(),
    pct: d.utilizationPct != null ? parseFloat(d.utilizationPct.toFixed(1)) : 0,
  }))

  return (
    <div className="mt-6 bg-white dark:bg-gray-800 rounded-lg shadow p-5">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">Utilization History</h2>
        <div className="flex gap-1">
          {HISTORY_DAYS_OPTIONS.map(d => (
            <button
              key={d}
              onClick={() => setHistoryDays(d)}
              className={`px-2 py-1 rounded text-xs font-medium transition ${historyDays === d ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}`}
            >
              {d}d
            </button>
          ))}
        </div>
      </div>
      {historyLoading && <p className="text-gray-400 text-sm">Loading history...</p>}
      {!historyLoading && historyError && <p className="text-red-500 text-sm">{historyError}</p>}
      {!historyLoading && !historyError && chartData.length === 0 && (
        <p className="text-gray-400 text-sm">No utilization history available for this period.</p>
      )}
      {!historyLoading && chartData.length > 0 && (
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis dataKey="date" tick={{ fontSize: 11 }} />
            <YAxis domain={[0, 100]} tickFormatter={v => `${v}%`} tick={{ fontSize: 11 }} />
            <Tooltip formatter={v => [`${v}%`, 'Utilization']} />
            <Line type="monotone" dataKey="pct" stroke="#3b82f6" strokeWidth={2} dot={false} />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
