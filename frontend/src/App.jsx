import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import VerifyEmailPage from './pages/VerifyEmailPage'
import DashboardPage from './pages/DashboardPage'
import SectionsPage from './pages/SectionsPage'
import SubnetsPage from './pages/SubnetsPage'
import IPAddressesPage from './pages/IPAddressesPage'
import AdminSettingsPage from './pages/AdminSettingsPage'
import AuditLogPage from './pages/AuditLogPage'
import UserSettingsPage from './pages/UserSettingsPage'
import AdminTagsPage from './pages/AdminTagsPage'
import OverlapReportPage from './pages/OverlapReportPage'
import DevicesPage from './pages/DevicesPage'
import DeviceDetailPage from './pages/DeviceDetailPage'
import AdminCustomFieldsPage from './pages/AdminCustomFieldsPage'
import AdminUsersPage from './pages/AdminUsersPage'
import LocationsPage from './pages/LocationsPage'
import LocationDetailPage from './pages/LocationDetailPage'
import RackDetailPage from './pages/RackDetailPage'
import NameserversPage from './pages/NameserversPage'
import DnsZonesPage from './pages/DnsZonesPage'
import DnsZoneDetailPage from './pages/DnsZoneDetailPage'
import AdminRequestsPage from './pages/AdminRequestsPage'
import MyRequestsPage from './pages/MyRequestsPage'
import VlansPage from './pages/VlansPage'
import VlanDetailPage from './pages/VlanDetailPage'
import VlanDomainsPage from './pages/VlanDomainsPage'
import VlanGroupsPage from './pages/VlanGroupsPage'
import VlanUsageReportPage from './pages/VlanUsageReportPage'
import ScanJobsPage from './pages/ScanJobsPage'
import AdminAgentsPage from './pages/AdminAgentsPage'
import CalculatorPage from './pages/CalculatorPage'
import TopologyPage from './pages/TopologyPage'
import UtilisationTrendsPage from './pages/UtilisationTrendsPage'
import ScheduledReportsPage from './pages/ScheduledReportsPage'
import InactiveIPsPage from './pages/InactiveIPsPage'
import ProtectedRoute from './components/ProtectedRoute'

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

export default function App() {
  return (
    <BrowserRouter>
      <DarkModeBootstrap />
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/verify-email" element={<VerifyEmailPage />} />
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
          <Route path="tools/calculator" element={<CalculatorPage />} />
          <Route path="sections/:id/topology" element={<TopologyPage />} />
          <Route path="reports/utilisation-trends" element={<UtilisationTrendsPage />} />
          <Route path="reports/inactive-ips" element={<InactiveIPsPage />} />
          <Route path="admin/reports/scheduled" element={<ScheduledReportsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
