import { useSearchParams } from 'react-router-dom'
import AdminUsersPage from './AdminUsersPage'
import AdminRolesPage from './AdminRolesPage'
import RolePresetsPage from './RolePresetsPage'

const TABS = [
  { id: 'users', label: 'Users' },
  { id: 'roles', label: 'Roles' },
  { id: 'presets', label: 'Permission Presets' },
]

export default function UsersRolesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = searchParams.get('tab') || 'users'

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
      <div className="flex-1 min-h-0 overflow-auto">
        {activeTab === 'users' && <AdminUsersPage />}
        {activeTab === 'roles' && <AdminRolesPage />}
        {activeTab === 'presets' && <RolePresetsPage />}
      </div>
    </div>
  )
}
