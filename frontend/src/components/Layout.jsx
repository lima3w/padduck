import { Outlet, useLocation } from 'react-router-dom'
import { useState, useEffect } from 'react'
import Header from './Header'
import Sidebar from './Sidebar'
import CommandPalette from './CommandPalette'
import { useDarkMode } from '../hooks/useDarkMode'

export default function Layout() {
  const darkMode = useDarkMode()
  const [paletteOpen, setPaletteOpen] = useState(false)
  const [navOpen, setNavOpen] = useState(false)
  const location = useLocation()

  // close mobile nav on route change
  useEffect(() => {
    setNavOpen(false)
  }, [location.pathname])

  useEffect(() => {
    function onKey(e) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setPaletteOpen((o) => !o)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  return (
    <div className="flex flex-col h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded focus:bg-white focus:px-3 focus:py-2 focus:text-sm focus:font-medium focus:text-blue-700 focus:ring-2 focus:ring-blue-500 dark:focus:bg-gray-800 dark:focus:text-blue-300"
      >
        Skip to main content
      </a>
      <Header darkMode={darkMode} onSearchClick={() => setPaletteOpen(true)} onNavToggle={() => setNavOpen(o => !o)} />
      <div className="flex flex-1 overflow-hidden">
        {/* mobile overlay */}
        {navOpen && (
          <div
            className="fixed inset-0 z-30 bg-black/50 lg:hidden"
            onClick={() => setNavOpen(false)}
            aria-hidden="true"
          />
        )}
        <Sidebar open={navOpen} onClose={() => setNavOpen(false)} />
        <main id="main-content" tabIndex={-1} className="flex-1 overflow-auto p-3 sm:p-6 focus:outline-none">
          <Outlet />
        </main>
      </div>
      <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
    </div>
  )
}
