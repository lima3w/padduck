import { NavLink } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { getPendingRequestCount } from '../api/requests'

export default function Sidebar() {
  const user = (() => {
    try { return JSON.parse(localStorage.getItem('current_user')) } catch { return null }
  })()
  const isAdmin = user?.role === 'admin'

  const [pendingCount, setPendingCount] = useState(0)

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
          Devices
        </NavLink>
        {!isAdmin && (
          <NavLink
            to="/requests"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
              }`
            }
          >
            My Requests
          </NavLink>
        )}

        <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
          Physical
        </div>
        <NavLink
          to="/locations"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Locations
        </NavLink>

        <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
          VLANs
        </div>
        <NavLink
          to="/vlans"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          VLANs
        </NavLink>

        <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
          DNS
        </div>
        <NavLink
          to="/dns/nameservers"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Nameservers
        </NavLink>
        <NavLink
          to="/dns/zones"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          DNS Zones
        </NavLink>

        <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
          Reports
        </div>
        <NavLink
          to="/reports/utilization-trends"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Utilization Trends
        </NavLink>
        <NavLink
          to="/reports/inactive-ips"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Inactive IPs
        </NavLink>

        <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
          Tools
        </div>
        <NavLink
          to="/tools/calculator"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          IP Calculator
        </NavLink>

        {isAdmin && (
          <>
            <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              Admin
            </div>
            <NavLink
              to="/admin/requests"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors flex items-center justify-between ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
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
              to="/admin/users"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Users &amp; Roles
            </NavLink>
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
            <NavLink
              to="/admin/custom-fields"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Custom Fields
            </NavLink>
            <NavLink
              to="/admin/vlan-domains"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              VLAN Domains
            </NavLink>
            <NavLink
              to="/admin/vlan-groups"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              VLAN Groups
            </NavLink>
            <NavLink
              to="/admin/vlans/usage-report"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              VLAN Usage
            </NavLink>
            <NavLink
              to="/admin/scan-jobs"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Scan Jobs
            </NavLink>
            <NavLink
              to="/admin/scan-agents"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Scan Agents
            </NavLink>
            <NavLink
              to="/admin/reports/scheduled"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Scheduled Reports
            </NavLink>
            <NavLink
              to="/admin/import"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Data Import
            </NavLink>
            <NavLink
              to="/admin/export"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Data Export
            </NavLink>
            <p className="px-3 mt-3 mb-1 text-xs text-gray-400 uppercase tracking-wider">Authentication</p>
            <NavLink
              to="/admin/auth/ldap"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              LDAP / AD
            </NavLink>
            <NavLink
              to="/admin/auth/oauth2"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              OAuth2 / OIDC
            </NavLink>
            <NavLink
              to="/admin/auth/saml"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              SAML SSO
            </NavLink>
          </>
        )}
      </nav>
    </aside>
  )
}
