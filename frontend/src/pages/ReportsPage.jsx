import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import UtilizationTrendsPage from './UtilizationTrendsPage'
import InactiveIPsPage from './InactiveIPsPage'
import DuplicatesPage from './DuplicatesPage'
import ReconciliationCenterPage from './ReconciliationCenterPage'
import ScheduledReportsPage from './ScheduledReportsPage'
import { getCachedUser } from '../utils/storageKeys'

export default function ReportsPage() {
  const { t } = useTranslation()
  const BASE_TABS = [
    { id: 'utilization', label: t('reports.utilizationTrendsTab') },
    { id: 'inactive', label: t('reports.inactiveIpsTab') },
    { id: 'duplicates', label: t('reports.duplicateDetectionTab') },
    { id: 'reconciliation', label: t('reports.reconciliationCenterTab') },
  ]
  const [searchParams, setSearchParams] = useSearchParams()
  const isAdmin = getCachedUser()?.role === 'admin'
  const tabs = isAdmin ? [...BASE_TABS, { id: 'scheduled', label: t('reports.scheduledReportsTab') }] : BASE_TABS
  const activeTab = searchParams.get('tab') || 'utilization'

  function setTab(id) {
    setSearchParams({ tab: id })
  }

  return (
    <div className="flex flex-col min-h-0 h-full">
      <div className="border-b border-gray-200 dark:border-gray-700">
        <div className="flex gap-0">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setTab(tab.id)}
              className={`px-4 py-3 text-sm font-medium transition-colors border-b-2 -mb-px ${
                activeTab === tab.id
                  ? 'border-blue-600 text-blue-600 dark:border-blue-400 dark:text-blue-400'
                  : 'border-transparent text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-auto">
        {activeTab === 'utilization' && <UtilizationTrendsPage />}
        {activeTab === 'inactive' && <InactiveIPsPage />}
        {activeTab === 'duplicates' && <DuplicatesPage />}
        {activeTab === 'reconciliation' && <ReconciliationCenterPage />}
        {activeTab === 'scheduled' && isAdmin && <ScheduledReportsPage />}
      </div>
    </div>
  )
}
