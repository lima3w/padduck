import { useState, useEffect, useRef } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import Avatar from './Avatar'

export default function UserMenu({ darkMode }) {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const menuRef = useRef(null)

  // Close on outside click
  useEffect(() => {
    if (!open) return
    function handleClick(e) {
      if (menuRef.current && !menuRef.current.contains(e.target)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  // Close on Escape key
  useEffect(() => {
    if (!open) return
    function handleKey(e) {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [open])

  async function handleLogout() {
    setOpen(false)
    await logout()
    navigate('/login')
  }

  function toggleDarkMode() {
    if (!darkMode) return
    const modes = ['system', 'light', 'dark']
    const next = modes[(modes.indexOf(darkMode.mode) + 1) % modes.length]
    darkMode.setPreference(next)
  }

  const modeLabel = darkMode
    ? { system: 'System', light: 'Light', dark: 'Dark' }[darkMode.mode]
    : null

  const modeIcon = darkMode
    ? { system: '⚙', light: '☀', dark: '🌙' }[darkMode.mode]
    : null

  if (!user) return null

  return (
    <div className="relative" ref={menuRef}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-2 rounded px-2 py-1 hover:bg-blue-600 dark:hover:bg-gray-700 transition focus:outline-none focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-blue-700 dark:focus:ring-offset-gray-800"
        aria-label="Open user menu"
        aria-haspopup="menu"
        aria-expanded={open}
      >
        <Avatar
          email={user.email}
          username={user.username}
          gravatarUrl={user.gravatarUrl}
          avatarSource={user.avatarSource}
          avatarBust={user.updatedAt}
          size={32}
        />
        <span className="text-sm text-blue-100 dark:text-gray-300">{user.username}</span>
        <svg className="w-3 h-3 text-blue-200 dark:text-gray-400" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
        </svg>
      </button>

      {open && (
        <div
          role="menu"
          aria-label="User menu"
          className="absolute right-0 mt-2 w-52 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50 py-1"
        >
          <Link
            role="menuitem"
            to="/settings"
            onClick={() => setOpen(false)}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none dark:hover:bg-gray-700 dark:focus:bg-gray-700"
          >
            <span>My Settings</span>
          </Link>
          <Link
            role="menuitem"
            to="/requests"
            onClick={() => setOpen(false)}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none dark:hover:bg-gray-700 dark:focus:bg-gray-700"
          >
            <span>My Requests</span>
          </Link>
          <Link
            role="menuitem"
            to="/settings?tab=sessions"
            onClick={() => setOpen(false)}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none dark:hover:bg-gray-700 dark:focus:bg-gray-700"
          >
            <span>Active Sessions</span>
          </Link>
          <Link
            role="menuitem"
            to="/settings?tab=history"
            onClick={() => setOpen(false)}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none dark:hover:bg-gray-700 dark:focus:bg-gray-700"
          >
            <span>Login History</span>
          </Link>

          {darkMode && (
            <button
              type="button"
              role="menuitem"
              onClick={toggleDarkMode}
              className="w-full flex items-center justify-between gap-2 px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none dark:hover:bg-gray-700 dark:focus:bg-gray-700"
            >
              <span>Dark Mode</span>
              <span className="text-xs font-medium text-gray-500 dark:text-gray-400">
                {modeIcon} {modeLabel}
              </span>
            </button>
          )}

          <div className="my-1 border-t border-gray-200 dark:border-gray-600" />

          <Link
            role="menuitem"
            to="/admin/privacy/consent-report"
            onClick={() => setOpen(false)}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none dark:hover:bg-gray-700 dark:focus:bg-gray-700"
          >
            <span>Privacy Policy</span>
          </Link>

          <div className="my-1 border-t border-gray-200 dark:border-gray-600" />

          <button
            type="button"
            role="menuitem"
            onClick={handleLogout}
            className="w-full text-left px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none dark:hover:bg-gray-700 dark:focus:bg-gray-700"
          >
            Logout
          </button>
        </div>
      )}
    </div>
  )
}
