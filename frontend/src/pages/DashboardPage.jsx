import { useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import { getDashboardSummary, getDashboardRecentActivity } from '../api/app'
import { getAdminConfig } from '../api/admin'
import { getCachedUser } from '../utils/storageKeys'

function formatRelativeTime(isoString, t) {
  const now = Date.now()
  const then = new Date(isoString).getTime()
  const diff = Math.floor((now - then) / 1000)

  if (diff < 60) return t('dashboard.justNow')
  if (diff < 3600) return t('dashboard.minAgo', { count: Math.floor(diff / 60) })
  if (diff < 86400) return t('dashboard.hrAgo', { count: Math.floor(diff / 3600) })
  return t('dashboard.daysAgo', { count: Math.floor(diff / 86400) })
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

function UtilizationBar({ pct }) {
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
  const { t } = useTranslation()
  const navigate = useNavigate()
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  // Redirect admins to the telemetry setup page on first boot (when the
  // telemetry_enabled config key has never been set to "true" or "false").
  useEffect(() => {
    if (!isAdmin) return
    getAdminConfig()
      .then(res => {
        const val = res.data?.config?.telemetry_enabled
        if (val !== 'true' && val !== 'false') {
          navigate('/setup/telemetry', { replace: true })
        }
      })
      .catch(() => {})
  }, [isAdmin, navigate])

  const summaryQuery = useQuery({
    queryKey: ['dashboard', 'summary'],
    queryFn: () => getDashboardSummary().then(r => r.data),
  })
  const activityQuery = useQuery({
    queryKey: ['dashboard', 'activity'],
    queryFn: () => getDashboardRecentActivity().then(r => r.data),
  })
  // Best-effort panels: their failures must not take the dashboard down.
  const nearCapacityQuery = useQuery({
    queryKey: ['dashboard', 'near-capacity'],
    queryFn: () => api.get('/admin/reports/subnets-near-capacity')
      .then(r => (Array.isArray(r.data) ? r.data : []))
      .catch(() => []),
  })
  const driftedQuery = useQuery({
    queryKey: ['dashboard', 'drifted-ips'],
    queryFn: () => api.get('/admin/reports/inactive-ips', { params: { days: 30 } })
      .then(r => (r.data?.inactive ?? []).slice(0, 8))
      .catch(() => []),
  })

  const summary = summaryQuery.data ?? null
  const activity = activityQuery.data ?? []
  const nearCapacity = nearCapacityQuery.data ?? []
  const driftedIPs = driftedQuery.data ?? []
  const loading = summaryQuery.isLoading || activityQuery.isLoading
  const error = summaryQuery.isError || activityQuery.isError ? t('dashboard.loadError') : null

  const load = () => {
    summaryQuery.refetch()
    activityQuery.refetch()
    nearCapacityQuery.refetch()
    driftedQuery.refetch()
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('dashboard.title')}</h1>
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
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('dashboard.title')}</h1>
        <p className="text-red-600">{error}</p>
        <button onClick={load} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">
          {t('dashboard.retry')}
        </button>
      </div>
    )
  }

  const utilPct = summary?.utilizationPct ?? 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('dashboard.title')}</h1>
        <button
          onClick={load}
          className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
        >
          {t('dashboard.refresh')}
        </button>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <SummaryCard label={t('nav.networks')} value={summary?.totalNetworks ?? 0} />
        <SummaryCard label={t('dashboard.subnets')} value={summary?.totalSubnets ?? 0} />
        <SummaryCard
          label={t('dashboard.ipAddresses')}
          value={`${summary?.usedIps ?? 0} / ${summary?.totalIps ?? 0}`}
          sub={t('dashboard.assignedTotal')}
        />
        <SummaryCard
          label={t('dashboard.utilisation')}
          value={`${utilPct.toFixed(1)}%`}
          sub={t('dashboard.assignedIps')}
        />
      </div>

      {/* Pending Requests card (admin) */}
      {isAdmin && (() => {
        const pendingCount = (summary?.pendingSubnetRequests ?? 0) + (summary?.pendingIpRequests ?? 0)
        return (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <SummaryCard
              label={t('dashboard.pendingRequests')}
              value={pendingCount}
              sub={pendingCount > 0 ? t('dashboard.clickToReview') : t('dashboard.noPendingRequests')}
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
              {t('dashboard.subnetsNearCapacity')}
            </h2>
            <Link to="/reports/utilization-trends" className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400">
              {t('dashboard.viewTrends')}
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

      {/* Drifted IPs */}
      {isAdmin && driftedIPs.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">
              {t('dashboard.driftedIps')} <span className="ml-1 text-xs font-normal text-gray-400">{t('dashboard.inactiveDaysSuffix')}</span>
            </h2>
            <Link to="/reports/inactive-ips" className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400">
              {t('dashboard.viewAll')}
            </Link>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="text-gray-500 dark:text-gray-400 border-b dark:border-gray-700">
                  <th className="text-left pb-2 font-medium">{t('dashboard.ipAddressColumn')}</th>
                  <th className="text-left pb-2 font-medium">{t('dashboard.hostname')}</th>
                  <th className="text-left pb-2 font-medium">{t('dashboard.subnet')}</th>
                  <th className="text-right pb-2 font-medium">{t('dashboard.daysInactive')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                {driftedIPs.map(ip => (
                  <tr key={ip.ipId} className="hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="py-1.5 font-mono text-gray-800 dark:text-gray-200">{ip.ipAddress}</td>
                    <td className="py-1.5 text-gray-600 dark:text-gray-400">{ip.hostname || '—'}</td>
                    <td className="py-1.5 text-gray-500 dark:text-gray-500 font-mono">{ip.subnetCidr}</td>
                    <td className="py-1.5 text-right">
                      <span className={`inline-block px-1.5 py-0.5 rounded text-xs font-medium ${ip.daysInactive > 90 ? 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400' : 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400'}`}>
                        {ip.daysInactive}d
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top subnets */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
          <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4">
            {t('dashboard.topUtilisedSubnets')}
          </h2>
          {summary?.topSubnets?.length === 0 ? (
            <p className="text-sm text-gray-400">{t('dashboard.noUtilizationData')}</p>
          ) : (
            <div className="space-y-3">
              {summary?.topSubnets?.map((s) => (
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
                  <UtilizationBar pct={s.utilizationPct} />
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Recent activity */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-5">
          <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4">
            {t('dashboard.recentActivity')}
          </h2>
          {activity.length === 0 ? (
            <p className="text-sm text-gray-400">{t('dashboard.noRecentActivity')}</p>
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
                      <span className="text-gray-400 dark:text-gray-500"> {t('dashboard.byUser', { username: a.username })}</span>
                    )}
                  </div>
                  <span className="text-xs text-gray-400 dark:text-gray-500 whitespace-nowrap flex-shrink-0">
                    {formatRelativeTime(a.createdAt, t)}
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
