import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getDHCPLeases } from '../../api/modules'
import { useFeatures } from '../../context/FeaturesContext'

const fmt = v => (v ? new Date(v).toLocaleString() : '—')

// Read-only list of DHCP leases linked to a subnet (pass subnetId) or to a
// single IP record (pass ipId). Renders nothing when the DHCP module is off,
// the lookup fails, or there are no linked leases.
export default function DHCPLeasesSection({ subnetId, ipId, compact = false }) {
  const { t } = useTranslation()
  const features = useFeatures()
  const enabled = features?.dhcp !== false
  const [leases, setLeases] = useState([])

  useEffect(() => {
    if (!enabled || (!subnetId && !ipId)) return undefined
    let cancelled = false
    const params = ipId ? { ip_id: ipId } : { subnet_id: subnetId }
    getDHCPLeases(params)
      .then(res => { if (!cancelled) setLeases(Array.isArray(res.data) ? res.data : []) })
      .catch(() => { if (!cancelled) setLeases([]) })
    return () => { cancelled = true }
  }, [enabled, subnetId, ipId])

  if (!enabled || leases.length === 0) return null

  return (
    <section aria-label={t('dhcp.linkedLeasesTitle')} className={compact ? '' : 'mt-8'}>
      <p className={compact ? 'text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2' : 'text-lg font-semibold mb-3 text-gray-800 dark:text-gray-100'}>
        {t('dhcp.linkedLeasesTitle')}
      </p>
      <div className="overflow-x-auto rounded border border-gray-200 dark:border-gray-700">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
          <thead className="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th className="px-3 py-2 text-left">{t('associateIp.ipAddress')}</th>
              <th className="px-3 py-2 text-left">{t('dhcp.mac')}</th>
              <th className="px-3 py-2 text-left">{t('dashboard.hostname')}</th>
              <th className="px-3 py-2 text-left">{t('dhcp.state')}</th>
              <th className="px-3 py-2 text-left">{t('dhcp.vlan')}</th>
              <th className="px-3 py-2 text-left">{t('dhcp.expires')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
            {leases.map(l => (
              <tr key={l.id}>
                <td className="px-3 py-2 font-mono">{l.ipAddress}</td>
                <td className="px-3 py-2 font-mono">{l.macAddress}</td>
                <td className="px-3 py-2">{l.hostname || '—'}</td>
                <td className="px-3 py-2">{l.state}</td>
                <td className="px-3 py-2">{l.vlanName || '—'}</td>
                <td className="px-3 py-2">{fmt(l.endsAt)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
