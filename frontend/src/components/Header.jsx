import { useAuth } from '../hooks/useAuth'
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import UserMenu from './UserMenu'
import { api } from '../api/client'
import { isSafeHttpUrl } from '../utils/url'

export default function Header({ darkMode, onSearchClick, onNavToggle }) {
  const { t } = useTranslation()
  const { user } = useAuth()
  const [appUrl, setAppUrl] = useState(null)

  useEffect(() => {
    api.get('/public-info').then(res => {
      // Admin-configured value; only accept http(s) so a misconfigured
      // javascript:/data: URL can never become the logo link.
      if (isSafeHttpUrl(res.data?.appUrl)) setAppUrl(res.data.appUrl)
    }).catch(() => {})
  }, [])

  return (
    <header className="bg-[#07162b] text-white px-4 py-3 flex items-center justify-between shadow border-b border-[#25364a] shrink-0">
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onNavToggle}
          aria-label={t('header.toggleNavigation')}
          className="lg:hidden p-1.5 rounded hover:bg-[#25364a] text-[#a8b8cb] focus:outline-none focus:ring-2 focus:ring-[#f5b800]"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <a
          href={appUrl || '/'}
          className="flex items-center gap-3 hover:opacity-90 transition-opacity"
          title={appUrl || t('nav.dashboard')}
        >
          <img src="/favicon.svg" alt="Padduck" className="w-8 h-8" />
          <span className="text-xl font-bold tracking-tight">Padduck</span>
          <span className="text-[#a8b8cb] text-sm hidden sm:inline">{t('header.tagline')}</span>
        </a>
      </div>
      <div className="flex items-center gap-2 sm:gap-3">
        <button
          type="button"
          onClick={onSearchClick}
          aria-label={t('header.search')}
          className="flex items-center gap-2 text-sm bg-[#0a1f3a] hover:bg-[#25364a] text-[#a8b8cb] border border-[#25364a] px-3 py-1 rounded-md transition focus:outline-none focus:ring-2 focus:ring-[#f5b800] focus:ring-offset-2 focus:ring-offset-[#07162b]"
          title={t('header.searchTitle')}
        >
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8" strokeWidth="2" />
            <path d="m21 21-4.35-4.35" strokeWidth="2" strokeLinecap="round" />
          </svg>
          <span className="hidden md:inline">{t('header.search')}</span>
          <kbd className="hidden md:inline-flex items-center text-xs text-[#3a4f65] font-mono">⌘K</kbd>
        </button>
        {user?.role === 'admin' && (
          <Link
            to="/admin"
            className="hidden sm:block text-sm bg-[#0a1f3a] hover:bg-[#25364a] px-3 py-1 rounded transition border border-[#25364a]"
          >
            {t('header.admin')}
          </Link>
        )}
        {user && <UserMenu darkMode={darkMode} />}
      </div>
    </header>
  )
}
