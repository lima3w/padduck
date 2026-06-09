import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Suspense, lazy, useEffect, useState } from 'react'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import { getFeatures } from './api/client'
import { normalizeFeatures } from './utils/features'
import { getStoredItem, LEGACY_STORAGE_KEYS, STORAGE_KEYS } from './utils/storageKeys'

const LoginPage = lazy(() => import('./pages/LoginPage'))
const RegisterPage = lazy(() => import('./pages/RegisterPage'))
const VerifyEmailPage = lazy(() => import('./pages/VerifyEmailPage'))
const DashboardPage = lazy(() => import('./pages/DashboardPage'))
const NetworksPage = lazy(() => import('./pages/NetworksPage'))
const SubnetsPage = lazy(() => import('./pages/SubnetsPage'))
const IPAddressesPage = lazy(() => import('./pages/IPAddressesPage'))
const AdminOverviewPage = lazy(() => import('./pages/AdminOverviewPage'))
const AdminSettingsPage = lazy(() => import('./pages/AdminSettingsPage'))
const AuditLogPage = lazy(() => import('./pages/AuditLogPage'))
const AuditRetentionPage = lazy(() => import('./pages/AuditRetentionPage'))
const AuditPage = lazy(() => import('./pages/AuditPage'))
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
const ScanProfilesPage = lazy(() => import('./pages/ScanProfilesPage'))
const ScanRetentionPage = lazy(() => import('./pages/ScanRetentionPage'))
const AdminAgentsPage = lazy(() => import('./pages/AdminAgentsPage'))
const AdminWebhooksPage = lazy(() => import('./pages/AdminWebhooksPage'))
const TopologyPage = lazy(() => import('./pages/TopologyPage'))
const UtilizationTrendsPage = lazy(() => import('./pages/UtilizationTrendsPage'))
const ScheduledReportsPage = lazy(() => import('./pages/ScheduledReportsPage'))
const InactiveIPsPage = lazy(() => import('./pages/InactiveIPsPage'))
const DuplicatesPage = lazy(() => import('./pages/DuplicatesPage'))
const ReconciliationCenterPage = lazy(() => import('./pages/ReconciliationCenterPage'))
const ImportDataPage = lazy(() => import('./pages/ImportDataPage'))
const ForgotPasswordPage = lazy(() => import('./pages/ForgotPasswordPage'))
const ResetPasswordPage = lazy(() => import('./pages/ResetPasswordPage'))
const VRFsPage = lazy(() => import('./pages/VRFsPage'))
const AdminRolesPage = lazy(() => import('./pages/AdminRolesPage'))
const RolePresetsPage = lazy(() => import('./pages/RolePresetsPage'))
const RacksPage = lazy(() => import('./pages/RacksPage'))
const ExportDataPage = lazy(() => import('./pages/ExportDataPage'))
const BackupsPage = lazy(() => import('./pages/BackupsPage'))
const AdminLdapPage = lazy(() => import('./pages/AdminLdapPage'))
const AdminOAuth2Page = lazy(() => import('./pages/AdminOAuth2Page'))
const AdminSamlPage = lazy(() => import('./pages/AdminSamlPage'))
const AuthCallbackPage = lazy(() => import('./pages/AuthCallbackPage'))
const AdminIntegrationsPage = lazy(() => import('./pages/AdminIntegrationsPage'))
const AdminGrafanaPage = lazy(() => import('./pages/AdminGrafanaPage'))
const CustomersPage = lazy(() => import('./pages/CustomersPage'))
const AutonomousSystemsPage = lazy(() => import('./pages/AutonomousSystemsPage'))
const NATRulesPage = lazy(() => import('./pages/NATRulesPage'))
const FirewallZonesPage = lazy(() => import('./pages/FirewallZonesPage'))
const DHCPPage = lazy(() => import('./pages/DHCPPage'))
const CircuitsPage = lazy(() => import('./pages/CircuitsPage'))
const DiscoveryConflictsPage = lazy(() => import('./pages/DiscoveryConflictsPage'))
const TopologyHintsPage = lazy(() => import('./pages/TopologyHintsPage'))
const APITokenAnalyticsPage = lazy(() => import('./pages/APITokenAnalyticsPage'))
const IntegrationTemplatesPage = lazy(() => import('./pages/IntegrationTemplatesPage'))
const AutomationPoliciesPage = lazy(() => import('./pages/AutomationPoliciesPage'))
const DeploymentHealthPage = lazy(() => import('./pages/DeploymentHealthPage'))
const PrivacyConsentReportPage = lazy(() => import('./pages/PrivacyConsentReportPage'))
const BreakGlassPage = lazy(() => import('./pages/BreakGlassPage'))
const IdentityPoliciesPage = lazy(() => import('./pages/IdentityPoliciesPage'))
const AdminCompatibilityPage = lazy(() => import('./pages/AdminCompatibilityPage'))
const DiscoveryPage = lazy(() => import('./pages/DiscoveryPage'))
const UsersRolesPage = lazy(() => import('./pages/UsersRolesPage'))
const ReportsPage = lazy(() => import('./pages/ReportsPage'))

// Apply system dark preference immediately on app mount (before useDarkMode hook runs)
function DarkModeBootstrap() {
  useEffect(() => {
    const stored = getStoredItem(STORAGE_KEYS.colorScheme, LEGACY_STORAGE_KEYS.colorScheme)
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

function FeatureGate({ feature, children }) {
  const [features, setFeatures] = useState(null)

  useEffect(() => {
    let cancelled = false
    getFeatures()
      .then((res) => {
        if (!cancelled) setFeatures(normalizeFeatures(res.data))
      })
      .catch(() => {
        if (!cancelled) setFeatures(normalizeFeatures())
      })
    return () => { cancelled = true }
  }, [])

  if (!features) return <PageLoadingFallback />
  if (features[feature] === false) {
    return (
      <div className="p-6 max-w-lg">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">Feature disabled</h1>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
          This feature has been disabled. An administrator can enable it in{' '}
          <a href="/admin/settings?tab=features" className="text-blue-600 dark:text-blue-400 underline hover:no-underline">
            Admin → Settings → Features
          </a>
          .
        </p>
      </div>
    )
  }
  return children
}

export default function App() {
  const gated = (feature, element) => <FeatureGate feature={feature}>{element}</FeatureGate>

  return (
    <BrowserRouter>
      <AuthProvider>
      <DarkModeBootstrap />
      <Suspense fallback={<PageLoadingFallback />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/verify-email" element={<VerifyEmailPage />} />
          <Route path="/forgot-password" element={<ForgotPasswordPage />} />
          <Route path="/reset-password" element={<ResetPasswordPage />} />
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
            <Route path="networks" element={<NetworksPage />} />
            <Route path="networks/:networkID/subnets" element={<SubnetsPage />} />
            <Route path="subnets/:subnetID/ip-addresses" element={<IPAddressesPage />} />
            <Route path="admin" element={<AdminOverviewPage />} />
            <Route path="admin/settings" element={<AdminSettingsPage />} />
            <Route path="admin/audit" element={<AuditPage />} />
            <Route path="admin/audit-log" element={<Navigate to="/admin/audit?tab=log" replace />} />
            <Route path="admin/audit/retention" element={<Navigate to="/admin/audit?tab=retention" replace />} />
            <Route path="admin/tags" element={<AdminTagsPage />} />
            <Route path="admin/overlap-report" element={<OverlapReportPage />} />
            <Route path="settings" element={<UserSettingsPage />} />
            <Route path="devices" element={gated('devices', <DevicesPage />)} />
            <Route path="devices/:id" element={gated('devices', <DeviceDetailPage />)} />
            <Route path="admin/custom-fields" element={<AdminCustomFieldsPage />} />
            <Route path="admin/users-roles" element={<UsersRolesPage />} />
            <Route path="admin/users" element={<Navigate to="/admin/users-roles" replace />} />
            <Route path="admin/roles" element={<Navigate to="/admin/users-roles?tab=roles" replace />} />
            <Route path="admin/roles/presets" element={<Navigate to="/admin/users-roles?tab=presets" replace />} />
            <Route path="locations" element={gated('locations', <LocationsPage />)} />
            <Route path="locations/:id" element={gated('locations', <LocationDetailPage />)} />
            <Route path="racks" element={gated('racks', <RacksPage />)} />
            <Route path="racks/:id" element={gated('racks', <RackDetailPage />)} />
            <Route path="dns/nameservers" element={<NameserversPage />} />
            <Route path="dns/zones" element={<DnsZonesPage />} />
            <Route path="dns/zones/:zone" element={<DnsZoneDetailPage />} />
            <Route path="admin/requests" element={<AdminRequestsPage />} />
            <Route path="requests" element={<MyRequestsPage />} />
            <Route path="vrfs" element={gated('vrfs', <VRFsPage />)} />
            <Route path="vlans" element={gated('vlans', <VlansPage />)} />
            <Route path="vlans/:id" element={gated('vlans', <VlanDetailPage />)} />
            <Route path="admin/vlan-domains" element={gated('vlans', <VlanDomainsPage />)} />
            <Route path="admin/vlan-groups" element={gated('vlans', <VlanGroupsPage />)} />
            <Route path="admin/vlans/usage-report" element={gated('vlans', <VlanUsageReportPage />)} />
            <Route path="admin/discovery" element={<DiscoveryPage />} />
            <Route path="admin/scan-jobs" element={<Navigate to="/admin/discovery?tab=scan-jobs" replace />} />
            <Route path="admin/scan-profiles" element={<Navigate to="/admin/discovery?tab=scan-profiles" replace />} />
            <Route path="admin/scan-retention" element={<Navigate to="/admin/discovery?tab=scan-retention" replace />} />
            <Route path="admin/discovery/conflicts" element={<Navigate to="/admin/discovery?tab=conflicts" replace />} />
            <Route path="admin/scan-agents" element={<AdminAgentsPage />} />
            <Route path="admin/webhooks" element={<AdminWebhooksPage />} />
            <Route path="admin/api-token-analytics" element={<APITokenAnalyticsPage />} />
            <Route path="admin/integration-templates" element={<IntegrationTemplatesPage />} />
            <Route path="admin/automation/policies" element={<AutomationPoliciesPage />} />
            <Route path="networks/:id/topology" element={<TopologyPage />} />
            <Route path="reports" element={<ReportsPage />} />
            <Route path="reports/utilization-trends" element={<Navigate to="/reports?tab=utilization" replace />} />
            <Route path="reports/inactive-ips" element={<Navigate to="/reports?tab=inactive" replace />} />
            <Route path="reports/duplicates" element={<Navigate to="/reports?tab=duplicates" replace />} />
            <Route path="reports/reconciliation" element={<Navigate to="/reports?tab=reconciliation" replace />} />
            <Route path="admin/reports/scheduled" element={<ScheduledReportsPage />} />
            <Route path="admin/import" element={<ImportDataPage />} />
            <Route path="admin/export" element={<ExportDataPage />} />
            <Route path="admin/backups" element={<BackupsPage />} />
            <Route path="admin/auth/ldap" element={<AdminLdapPage />} />
            <Route path="admin/auth/oauth2" element={<AdminOAuth2Page />} />
            <Route path="admin/auth/saml" element={<AdminSamlPage />} />
            <Route path="admin/integrations" element={<AdminIntegrationsPage />} />
            <Route path="admin/grafana" element={<AdminGrafanaPage />} />
            <Route path="customers" element={gated('customers', <CustomersPage />)} />
            <Route path="autonomous-systems" element={gated('bgp', <AutonomousSystemsPage />)} />
            <Route path="nat-rules" element={gated('nat', <NATRulesPage />)} />
            <Route path="firewall-zones" element={gated('firewall', <FirewallZonesPage />)} />
            <Route path="dhcp" element={gated('dhcp', <DHCPPage />)} />
            <Route path="circuits" element={gated('circuits', <CircuitsPage />)} />
            <Route path="admin/topology/hints" element={<Navigate to="/admin/discovery?tab=topology-hints" replace />} />
            <Route path="admin/system-health" element={<DeploymentHealthPage />} />
            <Route path="admin/privacy/consent-report" element={<PrivacyConsentReportPage />} />
            <Route path="admin/break-glass" element={<Navigate to="/admin/users" replace />} />
            <Route path="admin/identity-policies" element={<IdentityPoliciesPage />} />
            <Route path="admin/compatibility" element={<AdminCompatibilityPage />} />
          </Route>
        </Routes>
      </Suspense>
      </AuthProvider>
    </BrowserRouter>
  )
}
