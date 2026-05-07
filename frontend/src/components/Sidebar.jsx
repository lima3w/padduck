import { NavLink } from 'react-router-dom'

export default function Sidebar() {
  const user = (() => {
    try { return JSON.parse(localStorage.getItem('current_user')) } catch { return null }
  })()
  const isAdmin = user?.role === 'admin'

  return (
    <aside className="w-48 bg-gray-800 dark:bg-gray-900 text-gray-200 dark:text-gray-300 min-h-full flex flex-col border-r border-gray-700 dark:border-gray-700">
      <nav className="flex flex-col p-4 gap-1">
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Dashboard
        </NavLink>
        <NavLink
          to="/sections"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Sections
        </NavLink>
        <NavLink
          to="/devices"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          🖥️ Devices
        </NavLink>

        {isAdmin && (
          <>
            <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              Admin
            </div>
            <NavLink
              to="/admin/settings"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Settings
            </NavLink>
            <NavLink
              to="/admin/audit-log"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Audit Log
            </NavLink>
          </>
        )}
      </nav>
    </aside>
  )
}
