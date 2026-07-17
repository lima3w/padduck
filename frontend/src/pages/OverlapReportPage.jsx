import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { getOverlapReport } from '../api/admin'

export default function OverlapReportPage() {
  const { t } = useTranslation()
  const [overlaps, setOverlaps] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function load() {
      try {
        const res = await getOverlapReport()
        setOverlaps(res.data.overlaps || [])
      } catch {
        setError(t('overlapReport.loadError'))
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [t])

  if (loading) return <p className="text-gray-500">{t('overlapReport.loading')}</p>

  return (
    <div className="max-w-5xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{t('overlapReport.title')}</h1>
          <p className="text-sm text-gray-500 mt-1">{t('overlapReport.subtitle')}</p>
        </div>
        <Link to="/admin/settings" className="text-sm text-blue-600 hover:underline">{t('overlapReport.backToSettings')}</Link>
      </div>

      {error && <p className="text-red-600 mb-4 text-sm">{error}</p>}

      {overlaps.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center text-gray-400">
          <p className="text-lg font-medium text-green-600 mb-1">{t('overlapReport.noOverlapsDetected')}</p>
          <p className="text-sm">{t('overlapReport.allSubnetsSeparated')}</p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="px-4 py-3 bg-red-50 border-b border-red-100 text-sm text-red-700 font-medium">
            {t('overlapReport.pairsFound', { count: overlaps.length })}
          </div>
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="text-left px-4 py-3 text-gray-600 font-medium">{t('overlapReport.subnetA')}</th>
                <th className="text-left px-4 py-3 text-gray-600 font-medium">{t('overlapReport.subnetB')}</th>
              </tr>
            </thead>
            <tbody>
              {overlaps.map((pair, i) => (
                <tr key={i} className="border-b last:border-0 hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link
                      to={`/subnets/${pair.subnetA.id}/ip-addresses`}
                      className="font-mono text-blue-600 hover:underline font-medium"
                    >
                      {pair.subnetA.networkAddress}/{pair.subnetA.prefixLength}
                    </Link>
                    {pair.subnetA.description && (
                      <span className="ml-2 text-gray-400 text-xs">{pair.subnetA.description}</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <Link
                      to={`/subnets/${pair.subnetB.id}/ip-addresses`}
                      className="font-mono text-blue-600 hover:underline font-medium"
                    >
                      {pair.subnetB.networkAddress}/{pair.subnetB.prefixLength}
                    </Link>
                    {pair.subnetB.description && (
                      <span className="ml-2 text-gray-400 text-xs">{pair.subnetB.description}</span>
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
