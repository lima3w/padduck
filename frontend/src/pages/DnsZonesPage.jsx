import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { getDnsZones } from '../api/dns'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'

export default function DnsZonesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const zonesQuery = useQuery({
    queryKey: ['dns', 'zones'],
    queryFn: () => getDnsZones().then(r => r.data),
  })

  const data = zonesQuery.data
  const configured = data?.configured !== false
  const zones = configured ? (Array.isArray(data) ? data : (data?.zones ?? [])) : []
  const loading = zonesQuery.isLoading
  const error = zonesQuery.isError
    ? (zonesQuery.error?.response?.data?.error || t('dnsZones.loadError'))
    : null

  if (loading) return <PageSpinner message={t('dnsZones.loading')} />

  if (!configured) {
    return (
      <div>
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-4">{t('nav.dnsZones')}</h1>
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded-lg p-6 text-center">
          <p className="text-yellow-800 dark:text-yellow-200 font-medium mb-2">{t('dnsZones.noProviderConfigured')}</p>
          <p className="text-yellow-700 dark:text-yellow-300 text-sm mb-4">
            {t('dnsZones.setupPrefix')}{' '}
            <Link to="/admin/settings" className="underline hover:text-yellow-900 dark:hover:text-yellow-100">
              {t('dnsZones.adminSettingsDnsLink')}
            </Link>
            {t('dnsZones.setupSuffix')}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-4">{t('nav.dnsZones')}</h1>

      <ErrorBanner error={error} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('dnsZones.zoneName')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('dnsZones.kind')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('dnsZones.serial')}</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {zones.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-gray-400">
                  {t('dnsZones.noZonesFound')}
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
                  <span className="text-blue-600 dark:text-blue-400 text-xs">{t('dnsZones.viewArrow')}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>
    </div>
  )
}
