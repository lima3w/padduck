import { useSearchParams } from 'react-router-dom'
import AuditLogPage from './AuditLogPage'
import AuditRetentionPage from './AuditRetentionPage'

const TABS = [
  { id: 'log', label: 'Audit Log' },
  { id: 'retention', label: 'Retention' },
]

const VALID_TABS = new Set(TABS.map((t) => t.id))

export default function AuditPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const tab = VALID_TABS.has(searchParams.get('tab')) ? searchParams.get('tab') : 'log'

  function selectTab(id) {
    setSearchParams({ tab: id }, { replace: true })
  }

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div
        role="tablist"
        aria-label="Audit sections"
        className="flex gap-1 mb-6 border-b border-gray-200 dark:border-gray-700"
      >
        {TABS.map((t) => (
          <button
            key={t.id}
            role="tab"
            aria-selected={tab === t.id}
            onClick={() => selectTab(t.id)}
            className={`px-4 py-2 text-sm font-medium rounded-t transition-colors focus:outline-none ${
              tab === t.id
                ? 'bg-white dark:bg-gray-800 border border-b-white dark:border-gray-700 dark:border-b-gray-800 -mb-px text-blue-600 dark:text-blue-400'
                : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'log' && <AuditLogPage />}
      {tab === 'retention' && <AuditRetentionPage />}
    </div>
  )
}
