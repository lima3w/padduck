import { NavLink } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { getPendingRequestCount } from '../api/requests'
import { getDnsZones, getFeatures } from '../api/client'
import { DEFAULT_FEATURES, normalizeFeatures } from '../utils/features'
import { getCachedUser } from '../utils/storageKeys'

export default function Sidebar() {
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  const [pendingCount, setPendingCount] = useState(0)
  const [dnsConfigured, setDnsConfigured] = useState(true)
  const [features, setFeatures] = useState(DEFAULT_FEATURES)

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
    <aside className="w-48 bg-gray-800 dark:bg-gray-900 text-gray-200 dark:text-gray-300 h-full min-h-0 flex flex-col border-r border-gray-700 dark:border-gray-700">
      <nav className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto overscroll-contain p-4">
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
        {features.devices && (
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
        )}
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

        {(features.locations || features.racks) && (
          <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
            Physical
          </div>
        )}
        {features.locations && (
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
        )}
        {features.racks && (
          <NavLink
            to="/racks"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
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
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
              }`
            }
          >
            BGP / AS Numbers
          </NavLink>
        )}

        {(features.nat || features.firewall || features.dhcp || features.circuits) && (
          <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
            Network Services
          </div>
        )}
        {features.nat && (
          <NavLink
            to="/nat-rules"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
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
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
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
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
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
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
              }`
            }
          >
            Circuits
          </NavLink>
        )}

        {features.customers && (
          <>
            <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              Customers
            </div>
            <NavLink
              to="/customers"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Customers
            </NavLink>
          </>
        )}

        {(features.vlans || features.vrfs) && (
          <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
            VLANs &amp; VRFs
          </div>
        )}
        {features.vlans && (
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
        )}
        {features.vrfs && (
          <NavLink
            to="/vrfs"
            className={({ isActive }) =>
              `px-3 py-2 rounded text-sm font-medium transition-colors ${
                isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
              }`
            }
          >
            VRFs
          </NavLink>
        )}

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
        {dnsConfigured && (
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
        )}

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
        <NavLink
          to="/reports/duplicates"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Duplicate Detection
        </NavLink>
        <NavLink
          to="/reports/reconciliation"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
            }`
          }
        >
          Reconciliation Center
        </NavLink>

        {isAdmin && (
          <>
            <div className="mt-4 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              Admin
            </div>
            <NavLink
              to="/admin"
              end
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Overview
            </NavLink>
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
              to="/admin/roles"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Role Management
            </NavLink>
            <NavLink
              to="/admin/roles/presets"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Permission Presets
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
              to="/admin/break-glass"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Break-Glass
            </NavLink>
            <NavLink
              to="/admin/identity-policies"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Identity Policies
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
              to="/admin/audit/retention"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Audit Retention
            </NavLink>
            <NavLink
              to="/admin/discovery/conflicts"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Discovery Conflicts
            </NavLink>
            <div className="mt-2 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              Discovery
            </div>
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
              to="/admin/scan-profiles"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Scan Profiles
            </NavLink>
            <NavLink
              to="/admin/scan-retention"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Scan Retention
            </NavLink>
            <NavLink
              to="/admin/topology/hints"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Topology Hints
            </NavLink>
            <div className="mt-2 mb-1 px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              Automation
            </div>
            <NavLink
              to="/admin/automation/policies"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Policies
            </NavLink>
            <NavLink
              to="/admin/integration-templates"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Integration Templates
            </NavLink>
            <NavLink
              to="/admin/api-token-analytics"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Token Analytics
            </NavLink>
            <NavLink
              to="/admin/privacy/consent-report"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              Privacy Consent
            </NavLink>
            <NavLink
              to="/admin/system-health"
              className={({ isActive }) =>
                `px-3 py-2 rounded text-sm font-medium transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700 dark:hover:bg-gray-700'
                }`
              }
            >
              System Health
            </NavLink>
          </>
        )}
      </nav>
    </aside>
  )
}
