import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import ScanJobsPage from './ScanJobsPage'
import ScanRetentionPage from './ScanRetentionPage'
import TopologyHintsPage from './TopologyHintsPage'
import DiscoveryConflictsPage from './DiscoveryConflictsPage'
import DriftReviewPage from './DriftReviewPage'
import AdminAgentsPage from './AdminAgentsPage'

export default function DiscoveryPage() {
  const { t } = useTranslation()
  const TABS = [
    { id: 'scan-jobs', label: t('discovery.scanJobsTab') },
    { id: 'scan-retention', label: t('discovery.scanRetentionTab') },
    { id: 'topology-hints', label: t('discovery.topologyHintsTab') },
    { id: 'conflicts', label: t('discovery.conflictsTab') },
    { id: 'drift', label: t('discovery.driftTab') },
    { id: 'scan-agents', label: t('discovery.scanAgentsTab') },
  ]
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = searchParams.get('tab') || 'scan-jobs'

  function setTab(id) {
    setSearchParams({ tab: id })
  }

  return (
    <div className="flex flex-col min-h-0 h-full">
      <div className="border-b border-gray-200 dark:border-gray-700">
        <div className="flex gap-0">
          {TABS.map((tab) => (
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
      <div className="flex-1 min-h-0">
        {activeTab === 'scan-jobs' && <ScanJobsPage />}
        {activeTab === 'scan-retention' && <ScanRetentionPage />}
        {activeTab === 'topology-hints' && <TopologyHintsPage />}
        {activeTab === 'conflicts' && <DiscoveryConflictsPage />}
        {activeTab === 'drift' && <DriftReviewPage />}
        {activeTab === 'scan-agents' && <AdminAgentsPage />}
      </div>
    </div>
  )
}
