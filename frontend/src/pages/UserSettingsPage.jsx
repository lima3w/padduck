import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../hooks/useAuth'
import ProfileTab from './user/ProfileTab'
import SecurityTab from './user/SecurityTab'
import TokensTab from './user/TokensTab'
import SessionsTab from './user/SessionsTab'
import NotificationsTab from './user/NotificationsTab'
import LoginHistoryTab from './user/LoginHistoryTab'

const TAB_PARAM_MAP = { history: 'login-history', notif: 'notifications' }
const VALID_TABS = new Set(['profile', 'security', 'tokens', 'login-history', 'sessions', 'notifications'])

export default function UserSettingsPage() {
  const { t } = useTranslation()
  // setUser updates the shared auth context — header avatar updates immediately
  const { user, setUser } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()

  const rawTab = searchParams.get('tab') || 'profile'
  const resolvedTab = TAB_PARAM_MAP[rawTab] || rawTab
  const tab = VALID_TABS.has(resolvedTab) ? resolvedTab : 'profile'

  const setTab = (id) => setSearchParams({ tab: id }, { replace: true })

  const tabs = [
    { id: 'profile', label: t('settings.tabs.profile') },
    { id: 'security', label: t('settings.tabs.security') },
    { id: 'tokens', label: t('settings.tabs.apiTokens') },
    { id: 'sessions', label: t('settings.tabs.sessions') },
    { id: 'notifications', label: t('settings.tabs.notifications') },
    { id: 'login-history', label: t('settings.tabs.loginHistory') },
  ]

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">{t('settings.title')}</h1>

      <div className="flex flex-wrap gap-1 mb-6 border-b border-gray-200 dark:border-gray-700" role="tablist" aria-label={t('settings.tablistLabel')}>
        {tabs.map((t) => (
          <button
            key={t.id}
            type="button"
            role="tab"
            id={`account-tab-${t.id}`}
            aria-selected={tab === t.id}
            aria-controls={`account-panel-${t.id}`}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition ${
              tab === t.id
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div role="tabpanel" id={`account-panel-${tab}`} aria-labelledby={`account-tab-${tab}`}>
        {tab === 'profile' && <ProfileTab user={user} onAvatarChange={setUser} />}
        {tab === 'security' && <SecurityTab />}
        {tab === 'tokens' && <TokensTab />}
        {tab === 'sessions' && <SessionsTab />}
        {tab === 'notifications' && <NotificationsTab />}
        {tab === 'login-history' && <LoginHistoryTab />}
      </div>
    </div>
  )
}
