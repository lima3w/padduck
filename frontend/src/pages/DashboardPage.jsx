import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { getDashboardSummary, getDashboardRecentActivity, api } from '../api/client'

function formatRelativeTime(isoString) {
  const now = Date.now()
  const then = new Date(isoString).getTime()
  const diff = Math.floor((now - then) / 1000)

  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)} min ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)} hr ago`
  return `${Math.floor(diff / 86400)} days ago`
}

function actionIcon(action) {
  switch (action) {
    case 'ip_assigned': return '+'
    case 'ip_released': return '-'
    case 'subnet_created': return '+'
    case 'subnet_deleted': return 'x'
    case 'subnet_updated': return '~'
    default: return '•'
  }
}

function actionLabel(action) {
  return action.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}

function UtilisationBar({ pct }) {
  const colour =
    pct > 90 ? 'bg-red-500' :
    pct > 70 ? 'bg-yellow-500' :
    'bg-green-500'
  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
        <div
          className={`${colour} h-2 rounded-full transition-all`}
          style={{ width: `${Math.min(pct, 100)}%` }}
        />
      </div>
      <span className="text-xs text-gray-500 dark:text-gray-400 w-10 text-right">
        {pct.toFixed(0)}%
      </span>
    </div>
  )
}

function SummaryCard({ label, value, sub, onClick, highlight }) {
  return (
    <div
      className={`bg-white dark:bg-gray-800 rounded-lg shadow p-5 flex flex-col gap-1 ${onClick ? 'cursor-pointer hover:shadow-md transition-shadow' : ''} ${highlight ? 'ring-2 ring-yellow-400' : ''}`}
      onClick={onClick}
    >
      <span className="text-sm text-gray-500 dark:text-gray-400">{label}</span>
      <span className="text-3xl font-bold text-gray-800 dark:text-gray-100">{value}</span>
      {sub && <span className="text-xs text-gray-400 dark:text-gray-500">{sub}</span>}
    </div>
  )
}

export default function DashboardPage() {
  const navigate = useNavigate()
  const user = (() => { try { return JSON.parse(localStorage.getItem('current_user')) } catch { return null } })()
  const isAdmin = user?.role === 'admin'

  const [summary, setSummary] = useState(null)
  const [activity, setActivity] = useState([])
  const [nearCapacity, setNearCapacity] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const [sumRes, actRes] = await Promise.all([
        getDashboardSummary(),
        getDashboardRecentActivity(),
      ])
      setSummary(sumRes.data)
      setActivity(actRes.data)
      // Load subnets near capacity (best-effort, non-blocking)
      try {
        const capRes = await api.get('/admin/reports/subnets-near-capacity')
        setNearCapacity(Array.isArray(capRes.data) ? capRes.data : [])
      } catch {}
    } catch {
      setError('Failed to load dashboard data')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Dashboard</h1>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="bg-white dark:bg-gray-800 rounded-lg shadow p-5 h-24 animate-pulse bg-gray-100 dark:bg-gray-700" />
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Dashboard</h1>
        <p className="text-red-600">{error}</p>
        <button onClick={load} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">
          Retry
        </button>
      </div>
    )
  }

  const utilPct = summary?.utilisation_pct ?? 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Dashboard</h1>
        <button
          onClick={load}
          className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
        >
          Refresh
        </button>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <SummaryCard label="Sections" value={summary?.total_sections ?? 0} />
        <SummaryCard label="Subnets" value={summary?.total_subnets ?? 0} />
        <SummaryCard
          label="IP Addresses"
          value={`${summary?.used_ips ?? 0} / ${summary?.total_ips ?? 0}`}
          sub="assigned / total"
        />
        <SummaryCard
          label="Utilisation"
          value={`${utilPct.toFixed(1)}%`}
          sub="assigned IPs"
        />
      </div>

      {/* Pending Requests card (admin) */}
      {isAdmin && (() => {
        const pendingCount = (summary?.pending_subnet_requests ?? 0) + (summary?.pending_ip_requests ?? 0)
        return (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <SummaryCard
              label="Pending Requests"
              value={pendingCount}
              sub={pendingCount > 0 ? 'Click to review' : 'No pending requests'}
              onClick={() => navigate('/admin/requests')}
              highlight={pendingCount > 0}
            />
          </div>
        )
      })()}

      {/* Subnets Near Capacity */}
      {isAdmin && nearCapacity.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">
              Subnets Near Capacity
            </h2>
            <Link to="/reports/utilisation-trends" className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400">
              View trends →
            </Link>
          </div>
          <div className="space-y-3">
            {nearCapacity.map(s => (
              <div key={s.subnetId}>
                <div className="flex justify-between text-sm mb-1">
                  <Link
                    to={`/subnets/${s.subnetId}/ip-addresses`}
                    className="font-mono text-blue-600 dark:text-blue-400 hover:underline truncate"
                  >
                    {s.cidr}
                  </Link>
                  <span className="text-gray-500 dark:text-gray-400 ml-2 whitespace-nowrap text-xs">
                    {(s.currentPct ?? 0).toFixed(1)}%
                  </span>
                </div>
                {s.description && <p className="text-xs text-gray-400 mb-1">{s.description}</p>}
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className={`${s.currentPct >= 90 ? 'bg-red-500' : 'bg-yellow-500'} h-2 rounded-full transition-all`}
                    style={{ width: `${Math.min(s.currentPct ?? 0, 100)}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top subnets */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
          <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4">
            Top Utilised Subnets
          </h2>
          {summary?.top_subnets?.length === 0 ? (
            <p className="text-sm text-gray-400">No subnet utilisation data yet</p>
          ) : (
            <div className="space-y-3">
              {summary?.top_subnets?.map((s) => (
                <div key={s.id}>
                  <div className="flex justify-between text-sm mb-1">
                    <Link
                      to={`/subnets/${s.id}/ip-addresses`}
                      className="font-mono text-blue-600 dark:text-blue-400 hover:underline truncate"
                    >
                      {s.cidr}
                    </Link>
                    <span className="text-gray-500 dark:text-gray-400 ml-2 whitespace-nowrap">
                      {s.used}/{s.total}
                    </span>
                  </div>
                  {s.description && (
                    <p className="text-xs text-gray-400 mb-1">{s.description}</p>
                  )}
                  <UtilisationBar pct={s.utilisationPct} />
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Recent activity */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
          <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4">
            Recent Activity
          </h2>
          {activity.length === 0 ? (
            <p className="text-sm text-gray-400">No recent activity</p>
          ) : (
            <div className="space-y-2">
              {activity.map((a) => (
                <div key={a.id} className="flex items-start gap-3 text-sm">
                  <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 font-bold text-xs flex-shrink-0 mt-0.5">
                    {actionIcon(a.action)}
                  </span>
                  <div className="flex-1 min-w-0">
                    <span className="text-gray-700 dark:text-gray-300 font-medium">
                      {actionLabel(a.action)}
                    </span>
                    {a.description && (
                      <span className="text-gray-500 dark:text-gray-400"> — {
                        a.entityType === 'subnet' && a.entityId
                          ? <Link to={`/subnets/${a.entityId}/ip-addresses`} className="font-mono text-blue-600 dark:text-blue-400 hover:underline">{a.description}</Link>
                          : a.description
                      }</span>
                    )}
                    {a.username && (
                      <span className="text-gray-400 dark:text-gray-500"> by {a.username}</span>
                    )}
                  </div>
                  <span className="text-xs text-gray-400 dark:text-gray-500 whitespace-nowrap flex-shrink-0">
                    {formatRelativeTime(a.createdAt)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
