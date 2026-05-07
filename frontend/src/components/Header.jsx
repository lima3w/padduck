import { useAuth } from '../hooks/useAuth'
import { useNavigate, Link } from 'react-router-dom'

const modeIcons = { system: '⚙', light: '☀', dark: '🌙' }
const nextMode = { system: 'light', light: 'dark', dark: 'system' }

export default function Header({ darkMode }) {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <header className="bg-blue-700 dark:bg-gray-800 text-white px-6 py-3 flex items-center justify-between shadow">
      <div className="flex items-center">
        <span className="text-xl font-bold tracking-tight">IPAM Next</span>
        <span className="ml-3 text-blue-300 dark:text-gray-400 text-sm">IP Address Management</span>
      </div>
      <div className="flex items-center gap-4">
        {darkMode && (
          <button
            onClick={() => darkMode.setPreference(nextMode[darkMode.mode])}
            title={`Color scheme: ${darkMode.mode} (click to change)`}
            className="text-lg w-8 h-8 flex items-center justify-center rounded hover:bg-blue-600 dark:hover:bg-gray-700 transition"
            aria-label="Toggle color scheme"
          >
            {modeIcons[darkMode.mode]}
          </button>
        )}
        {user && (
          <>
            <span className="text-sm text-blue-100 dark:text-gray-300">{user.username}</span>
            <Link
              to="/settings"
              className="text-sm bg-blue-600 dark:bg-gray-700 hover:bg-blue-800 dark:hover:bg-gray-600 px-3 py-1 rounded transition"
            >
              Settings
            </Link>
            {user.role === 'admin' && (
              <Link
                to="/admin/settings"
                className="text-sm bg-blue-600 dark:bg-gray-700 hover:bg-blue-800 dark:hover:bg-gray-600 px-3 py-1 rounded transition"
              >
                Admin
              </Link>
            )}
            <button
              onClick={handleLogout}
              className="text-sm bg-blue-600 dark:bg-gray-700 hover:bg-blue-800 dark:hover:bg-gray-600 px-3 py-1 rounded transition"
            >
              Logout
            </button>
          </>
        )}
      </div>
    </header>
  )
}
