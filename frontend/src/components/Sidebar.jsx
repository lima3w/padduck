import { NavLink } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { getPendingRequestCount } from '../api/requests'
import { getDnsZones, getFeatures, checkForUpdates } from '../api/client'
import { DEFAULT_FEATURES, normalizeFeatures } from '../utils/features'
import { getCachedUser } from '../utils/storageKeys'

const SECTION_HEADER = 'mt-2 pt-3 mb-1 px-3 text-xs font-semibold text-[#a8b8cb]/50 uppercase tracking-wider border-t border-[#25364a]'

export default function Sidebar() {
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  const [pendingCount, setPendingCount] = useState(0)
  const [dnsConfigured, setDnsConfigured] = useState(true)
  const [features, setFeatures] = useState(DEFAULT_FEATURES)
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
    <aside className="w-48 bg-[#07162b] text-[#f4f7fa] h-full min-h-0 flex flex-col border-r border-[#25364a]">
      <nav className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto overscroll-contain p-4">
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
            }`
          }
        >
          Dashboard
        </NavLink>
        <NavLink
          to="/networks"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
            }`
          }
        >
          Networks
        </NavLink>
        {features.devices && (
          <NavLink
            to="/devices"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            Devices
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
            My Requests
          </NavLink>
        )}

        {(features.locations || features.racks) && (
          <div className={SECTION_HEADER}>Physical</div>
        )}
        {features.locations && (
          <NavLink
            to="/locations"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            Locations
          </NavLink>
        )}
        {features.racks && (
          <NavLink
            to="/racks"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            Racks
          </NavLink>
        )}

        {features.bgp && (
          <NavLink
            to="/autonomous-systems"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            BGP / AS Numbers
          </NavLink>
        )}

        {(features.nat || features.firewall || features.dhcp || features.circuits) && (
          <div className={SECTION_HEADER}>Network Services</div>
        )}
        {features.nat && (
          <NavLink
            to="/nat-rules"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            NAT Rules
          </NavLink>
        )}
        {features.firewall && (
          <NavLink
            to="/firewall-zones"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            Firewall Zones
          </NavLink>
        )}
        {features.dhcp && (
          <NavLink
            to="/dhcp"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            DHCP
          </NavLink>
        )}
        {features.circuits && (
          <NavLink
            to="/circuits"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            Circuits
          </NavLink>
        )}

        {features.customers && (
          <>
            <div className={SECTION_HEADER}>Customers</div>
            <NavLink
              to="/customers"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              Customers
            </NavLink>
          </>
        )}

        {(features.vlans || features.vrfs) && (
          <div className={SECTION_HEADER}>VLANs &amp; VRFs</div>
        )}
        {features.vlans && (
          <NavLink
            to="/vlans"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            VLANs
          </NavLink>
        )}
        {features.vrfs && (
          <NavLink
            to="/vrfs"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
              }`
            }
          >
            VRFs
          </NavLink>
        )}

        <div className={SECTION_HEADER}>DNS</div>
        <NavLink
          to="/dns/nameservers"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
            }`
          }
        >
          Nameservers
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
            DNS Zones
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
          Reports
        </NavLink>

        {isAdmin && (
          <>
            <div className={SECTION_HEADER}>Admin</div>
            <NavLink
              to="/admin/requests"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors flex items-center justify-between ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              <span>Requests</span>
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
              Users &amp; Roles
            </NavLink>
            <NavLink
              to="/admin/settings"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              Settings
            </NavLink>
            <NavLink
              to="/admin/audit"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              Audit
            </NavLink>
            <NavLink
              to="/admin/discovery"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              Discovery
            </NavLink>
            <div className={`${SECTION_HEADER} mt-1`}>Automation</div>
            <NavLink
              to="/admin/automation/policies"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              Policies
            </NavLink>
            <NavLink
              to="/admin/api-token-analytics"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              Token Analytics
            </NavLink>
            <NavLink
              to="/admin/system-health"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              System Health
            </NavLink>
            <NavLink
              to="/admin/backups"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-[#f5b800] text-[#07162b]' : 'hover:bg-[#0d2848]'
                }`
              }
            >
              Backups
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
