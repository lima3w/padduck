import { NavLink, useLocation } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getPendingRequestCount } from '../api/requests'
import { listDriftItems } from '../api/discovery'
import { getDnsZones } from '../api/dns'
import { getFeatures } from '../api/app'
import { checkForUpdates } from '../api/admin'
import { DEFAULT_FEATURES, normalizeFeatures } from '../utils/features'
import { getCachedUser } from '../utils/storageKeys'

const SECTION_HEADER = 'mt-2 pt-3 mb-1 px-3 text-xs font-semibold text-[#a8b8cb]/50 uppercase tracking-wider border-t border-[#25364a]'

export default function Sidebar({ open, onClose }) {
  const { t } = useTranslation()
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'
  const location = useLocation()

  useEffect(() => {
    onClose?.()
  }, [location.pathname, onClose])

  const [pendingCount, setPendingCount] = useState(0)
  const [driftCount, setDriftCount] = useState(0)
  const [dnsConfigured, setDnsConfigured] = useState(true)
  const [features, setFeatures] = useState(null)
  const [version, setVersion] = useState(null)

  useEffect(() => {
    if (!isAdmin) return
    let cancelled = false
    async function fetchCount() {
      try {
        const res = await getPendingRequestCount()
        if (!cancelled) setPendingCount(res.data?.count ?? 0)
      } catch {}
    }
    fetchCount()
    const interval = setInterval(fetchCount, 30000)
    return () => { cancelled = true; clearInterval(interval) }
  }, [isAdmin])

  useEffect(() => {
    if (!isAdmin) return
    let cancelled = false
    async function fetchDriftCount() {
      try {
        const res = await listDriftItems('open')
        if (!cancelled) setDriftCount(Array.isArray(res.data) ? res.data.length : 0)
      } catch {}
    }
    fetchDriftCount()
    const interval = setInterval(fetchDriftCount, 30000)
    return () => { cancelled = true; clearInterval(interval) }
  }, [isAdmin])

  useEffect(() => {
    let cancelled = false
    async function checkDns() {
      try {
        const res = await getDnsZones()
        if (!cancelled) setDnsConfigured(res.data?.configured !== false)
      } catch {}
    }
    checkDns()
    return () => { cancelled = true }
  }, [])

  useEffect(() => {
    if (!isAdmin) return
    let cancelled = false
    async function fetchVersion() {
      try {
        const res = await checkForUpdates()
        if (!cancelled) setVersion(res.data?.currentVersion ?? null)
      } catch {}
    }
    fetchVersion()
    return () => { cancelled = true }
  }, [isAdmin])

  useEffect(() => {
    let cancelled = false
    async function loadFeatures() {
      try {
        const res = await getFeatures()
        if (!cancelled) setFeatures(normalizeFeatures(res.data))
      } catch {
        if (!cancelled) setFeatures(DEFAULT_FEATURES)
      }
    }
    loadFeatures()
    return () => { cancelled = true }
  }, [])

  return (
    <aside className={`fixed inset-y-0 left-0 z-40 w-64 bg-[#07162b] text-[#f4f7fa] flex flex-col border-r border-[#25364a] transition-transform duration-200 ease-in-out lg:relative lg:inset-auto lg:z-auto lg:w-48 lg:translate-x-0 ${open ? 'translate-x-0' : '-translate-x-full'}`}>
      <div className="flex items-center justify-between px-4 pt-4 pb-2 lg:hidden shrink-0">
        <span className="text-sm font-semibold text-[#a8b8cb]">{t('nav.navigation')}</span>
        <button
          type="button"
          onClick={onClose}
          aria-label={t('nav.closeNavigation')}
          className="p-1.5 rounded hover:bg-[#0d2848] text-[#a8b8cb] focus:outline-none focus:ring-2 focus:ring-[#f5b800]"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <nav className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto overscroll-contain p-4 pt-2 lg:pt-4">
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
            }`
          }
        >
          {t('nav.dashboard')}
        </NavLink>
        <NavLink
          to="/networks"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
            }`
          }
        >
          {t('nav.networks')}
        </NavLink>
        {features?.devices && (
          <NavLink
            to="/devices"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.devices')}
          </NavLink>
        )}
        {!isAdmin && (
          <NavLink
            to="/requests"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.myRequests')}
          </NavLink>
        )}

        {(features?.locations || features?.racks) && (
          <div className={SECTION_HEADER}>{t('nav.sections.physical')}</div>
        )}
        {features?.locations && (
          <NavLink
            to="/locations"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.locations')}
          </NavLink>
        )}
        {features?.racks && (
          <NavLink
            to="/racks"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.racks')}
          </NavLink>
        )}

        {features?.bgp && (
          <NavLink
            to="/autonomous-systems"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.bgpAsNumbers')}
          </NavLink>
        )}

        {(features?.nat || features?.firewall || features?.dhcp || features?.circuits) && (
          <div className={SECTION_HEADER}>{t('nav.sections.networkServices')}</div>
        )}
        {features?.nat && (
          <NavLink
            to="/nat-rules"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.natRules')}
          </NavLink>
        )}
        {features?.firewall && (
          <NavLink
            to="/firewall-zones"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.firewallZones')}
          </NavLink>
        )}
        {features?.dhcp && (
          <NavLink
            to="/dhcp"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.dhcp')}
          </NavLink>
        )}
        {features?.circuits && (
          <NavLink
            to="/circuits"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.circuits')}
          </NavLink>
        )}

        {features?.customers && (
          <>
            <div className={SECTION_HEADER}>{t('nav.sections.customers')}</div>
            <NavLink
              to="/customers"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.customers')}
            </NavLink>
          </>
        )}

        {(features?.vlans || features?.vrfs) && (
          <div className={SECTION_HEADER}>{t('nav.sections.vlansVrfs')}</div>
        )}
        {features?.vlans && (
          <NavLink
            to="/vlans"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.vlans')}
          </NavLink>
        )}
        {features?.vrfs && (
          <NavLink
            to="/vrfs"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.vrfs')}
          </NavLink>
        )}

        <div className={SECTION_HEADER}>{t('nav.sections.dns')}</div>
        <NavLink
          to="/dns/nameservers"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
            }`
          }
        >
          {t('nav.nameservers')}
        </NavLink>
        {dnsConfigured && (
          <NavLink
            to="/dns/zones"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            {t('nav.dnsZones')}
          </NavLink>
        )}

        <NavLink
          to="/reports"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
            }`
          }
        >
          {t('nav.reports')}
        </NavLink>
        {isAdmin && (
          <>
            <div className={SECTION_HEADER}>{t('nav.sections.admin')}</div>
            <NavLink
              to="/admin/requests"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors flex items-center justify-between ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              <span>{t('nav.requests')}</span>
              {pendingCount > 0 && (
                <span className="ml-1 inline-flex items-center justify-center px-1.5 py-0.5 text-xs font-bold leading-none text-white bg-yellow-500 rounded-full">
                  {pendingCount}
                </span>
              )}
            </NavLink>
            <NavLink
              to="/admin/users-roles"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'text-[#c8d8e8] hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.usersRoles')}
            </NavLink>
            <NavLink
              to="/admin/settings"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.settings')}
            </NavLink>
            <NavLink
              to="/admin/audit"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.audit')}
            </NavLink>
            <NavLink
              to="/admin/discovery"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors flex items-center justify-between ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              <span>{t('nav.discovery')}</span>
              {driftCount > 0 && (
                <span className="ml-1 inline-flex items-center justify-center px-1.5 py-0.5 text-xs font-bold leading-none text-white bg-yellow-500 rounded-full">
                  {driftCount}
                </span>
              )}
            </NavLink>
            <div className={`${SECTION_HEADER} mt-1`}>{t('nav.sections.automation')}</div>
            <NavLink
              to="/admin/automation/policies"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.policies')}
            </NavLink>
            <NavLink
              to="/admin/api-token-analytics"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.tokenAnalytics')}
            </NavLink>
            <NavLink
              to="/admin/system-health"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.systemHealth')}
            </NavLink>
            <NavLink
              to="/admin/backups"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              {t('nav.backups')}
            </NavLink>
          </>
        )}
      </nav>
      {isAdmin && version && (
        <div className="px-4 py-2 text-[10px] text-[#a8b8cb]/40 border-t border-[#25364a] shrink-0">
          {version}
        </div>
      )}
    </aside>
  )
}
