import { useAuth } from '../hooks/useAuth'
import { Link } from 'react-router-dom'
import UserMenu from './UserMenu'

export default function Header({ darkMode, onSearchClick }) {
  const { user } = useAuth()

  return (
    <header className="bg-blue-700 dark:bg-gray-800 text-white px-6 py-3 flex items-center justify-between shadow">
      <div className="flex items-center">
        <span className="text-xl font-bold tracking-tight">IPAM Next</span>
        <span className="ml-3 text-blue-300 dark:text-gray-400 text-sm hidden sm:inline">IP Address Management</span>
      </div>
      <div className="flex items-center gap-3">
        <button
          onClick={onSearchClick}
          className="flex items-center gap-2 text-sm bg-blue-600/60 dark:bg-gray-700/60 hover:bg-blue-600 dark:hover:bg-gray-700 text-blue-100 dark:text-gray-300 border border-blue-500/40 dark:border-gray-600 px-3 py-1 rounded-md transition"
          title="Search (Ctrl+K)"
        >
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8" strokeWidth="2" />
            <path d="m21 21-4.35-4.35" strokeWidth="2" strokeLinecap="round" />
          </svg>
          <span className="hidden md:inline">Search</span>
          <kbd className="hidden md:inline-flex items-center text-xs text-blue-300 dark:text-gray-500 font-mono">⌘K</kbd>
        </button>
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
