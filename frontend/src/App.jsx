import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Suspense, lazy, useEffect } from 'react'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'

const LoginPage = lazy(() => import('./pages/LoginPage'))
const RegisterPage = lazy(() => import('./pages/RegisterPage'))
const VerifyEmailPage = lazy(() => import('./pages/VerifyEmailPage'))
const DashboardPage = lazy(() => import('./pages/DashboardPage'))
const SectionsPage = lazy(() => import('./pages/SectionsPage'))
const SubnetsPage = lazy(() => import('./pages/SubnetsPage'))
const IPAddressesPage = lazy(() => import('./pages/IPAddressesPage'))
const AdminSettingsPage = lazy(() => import('./pages/AdminSettingsPage'))
const AuditLogPage = lazy(() => import('./pages/AuditLogPage'))
const UserSettingsPage = lazy(() => import('./pages/UserSettingsPage'))
const AdminTagsPage = lazy(() => import('./pages/AdminTagsPage'))
const OverlapReportPage = lazy(() => import('./pages/OverlapReportPage'))
const DevicesPage = lazy(() => import('./pages/DevicesPage'))
const DeviceDetailPage = lazy(() => import('./pages/DeviceDetailPage'))
const AdminCustomFieldsPage = lazy(() => import('./pages/AdminCustomFieldsPage'))
const AdminUsersPage = lazy(() => import('./pages/AdminUsersPage'))
const LocationsPage = lazy(() => import('./pages/LocationsPage'))
const LocationDetailPage = lazy(() => import('./pages/LocationDetailPage'))
const RackDetailPage = lazy(() => import('./pages/RackDetailPage'))
const NameserversPage = lazy(() => import('./pages/NameserversPage'))
const DnsZonesPage = lazy(() => import('./pages/DnsZonesPage'))
const DnsZoneDetailPage = lazy(() => import('./pages/DnsZoneDetailPage'))
const AdminRequestsPage = lazy(() => import('./pages/AdminRequestsPage'))
const MyRequestsPage = lazy(() => import('./pages/MyRequestsPage'))
const VlansPage = lazy(() => import('./pages/VlansPage'))
const VlanDetailPage = lazy(() => import('./pages/VlanDetailPage'))
const VlanDomainsPage = lazy(() => import('./pages/VlanDomainsPage'))
const VlanGroupsPage = lazy(() => import('./pages/VlanGroupsPage'))
const VlanUsageReportPage = lazy(() => import('./pages/VlanUsageReportPage'))
const ScanJobsPage = lazy(() => import('./pages/ScanJobsPage'))
const AdminAgentsPage = lazy(() => import('./pages/AdminAgentsPage'))
const AdminWebhooksPage = lazy(() => import('./pages/AdminWebhooksPage'))
const CalculatorPage = lazy(() => import('./pages/CalculatorPage'))
const TopologyPage = lazy(() => import('./pages/TopologyPage'))
const UtilizationTrendsPage = lazy(() => import('./pages/UtilizationTrendsPage'))
const ScheduledReportsPage = lazy(() => import('./pages/ScheduledReportsPage'))
const InactiveIPsPage = lazy(() => import('./pages/InactiveIPsPage'))
const ImportDataPage = lazy(() => import('./pages/ImportDataPage'))
const ExportDataPage = lazy(() => import('./pages/ExportDataPage'))
const AdminLdapPage = lazy(() => import('./pages/AdminLdapPage'))
const AdminOAuth2Page = lazy(() => import('./pages/AdminOAuth2Page'))
const AdminSamlPage = lazy(() => import('./pages/AdminSamlPage'))
const AuthCallbackPage = lazy(() => import('./pages/AuthCallbackPage'))
const AdminIntegrationsPage = lazy(() => import('./pages/AdminIntegrationsPage'))
const CustomersPage = lazy(() => import('./pages/CustomersPage'))

// Apply system dark preference immediately on app mount (before useDarkMode hook runs)
function DarkModeBootstrap() {
  useEffect(() => {
    const stored = localStorage.getItem('ipam-color-scheme')
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const html = document.documentElement
    if (!stored || stored === 'system') {
      html.classList.toggle('dark', mq.matches)
      html.classList.toggle('light', !mq.matches)
    } else {
      html.classList.toggle('dark', stored === 'dark')
      html.classList.toggle('light', stored === 'light')
    }
  }, [])
  return null
}

function PageLoadingFallback() {
  return (
    <div className="flex min-h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      Loading...
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <DarkModeBootstrap />
      <Suspense fallback={<PageLoadingFallback />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/verify-email" element={<VerifyEmailPage />} />
          <Route path="/auth/callback" element={<AuthCallbackPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="sections" element={<SectionsPage />} />
            <Route path="sections/:sectionID/subnets" element={<SubnetsPage />} />
            <Route path="subnets/:subnetID/ip-addresses" element={<IPAddressesPage />} />
            <Route path="admin/settings" element={<AdminSettingsPage />} />
            <Route path="admin/audit-log" element={<AuditLogPage />} />
            <Route path="admin/tags" element={<AdminTagsPage />} />
            <Route path="admin/overlap-report" element={<OverlapReportPage />} />
            <Route path="settings" element={<UserSettingsPage />} />
            <Route path="devices" element={<DevicesPage />} />
            <Route path="devices/:id" element={<DeviceDetailPage />} />
            <Route path="admin/custom-fields" element={<AdminCustomFieldsPage />} />
            <Route path="admin/users" element={<AdminUsersPage />} />
            <Route path="locations" element={<LocationsPage />} />
            <Route path="locations/:id" element={<LocationDetailPage />} />
            <Route path="racks/:id" element={<RackDetailPage />} />
            <Route path="dns/nameservers" element={<NameserversPage />} />
            <Route path="dns/zones" element={<DnsZonesPage />} />
            <Route path="dns/zones/:zone" element={<DnsZoneDetailPage />} />
            <Route path="admin/requests" element={<AdminRequestsPage />} />
            <Route path="requests" element={<MyRequestsPage />} />
            <Route path="vlans" element={<VlansPage />} />
            <Route path="vlans/:id" element={<VlanDetailPage />} />
            <Route path="admin/vlan-domains" element={<VlanDomainsPage />} />
            <Route path="admin/vlan-groups" element={<VlanGroupsPage />} />
            <Route path="admin/vlans/usage-report" element={<VlanUsageReportPage />} />
            <Route path="admin/scan-jobs" element={<ScanJobsPage />} />
            <Route path="admin/scan-agents" element={<AdminAgentsPage />} />
            <Route path="admin/webhooks" element={<AdminWebhooksPage />} />
            <Route path="tools/calculator" element={<CalculatorPage />} />
            <Route path="sections/:id/topology" element={<TopologyPage />} />
            <Route path="reports/utilization-trends" element={<UtilizationTrendsPage />} />
            <Route path="reports/inactive-ips" element={<InactiveIPsPage />} />
            <Route path="admin/reports/scheduled" element={<ScheduledReportsPage />} />
            <Route path="admin/import" element={<ImportDataPage />} />
            <Route path="admin/export" element={<ExportDataPage />} />
            <Route path="admin/auth/ldap" element={<AdminLdapPage />} />
            <Route path="admin/auth/oauth2" element={<AdminOAuth2Page />} />
            <Route path="admin/auth/saml" element={<AdminSamlPage />} />
            <Route path="admin/integrations" element={<AdminIntegrationsPage />} />
            <Route path="customers" element={<CustomersPage />} />
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}
