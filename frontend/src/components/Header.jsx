import { useAuth } from '../hooks/useAuth'
import { Link } from 'react-router-dom'
import UserMenu from './UserMenu'

export default function Header({ darkMode }) {
  const { user } = useAuth()

  return (
    <header className="bg-blue-700 dark:bg-gray-800 text-white px-6 py-3 flex items-center justify-between shadow">
      <div className="flex items-center">
        <span className="text-xl font-bold tracking-tight">IPAM Next</span>
        <span className="ml-3 text-blue-300 dark:text-gray-400 text-sm">IP Address Management</span>
      </div>
      <div className="flex items-center gap-4">
        {user?.role === 'admin' && (
          <Link
            to="/admin/settings"
            className="text-sm bg-blue-600 dark:bg-gray-700 hover:bg-blue-800 dark:hover:bg-gray-600 px-3 py-1 rounded transition"
          >
            Admin
          </Link>
        )}
        {user && <UserMenu darkMode={darkMode} />}
      </div>
    </header>
  )
}
