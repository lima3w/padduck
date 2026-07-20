import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export const ADMIN_SURFACE_SECTION_KEYS = [
  {
    titleKey: 'identityAccessTitle',
    links: [
      { to: '/admin/users', titleKey: 'usersTitle', descriptionKey: 'usersDescription' },
      { to: '/admin/roles', titleKey: 'rolesTitle', descriptionKey: 'rolesDescription' },
      { to: '/admin/roles/presets', titleKey: 'permissionPresetsTitle', descriptionKey: 'permissionPresetsDescription' },
      { to: '/admin/requests', titleKey: 'approvalsTitle', descriptionKey: 'approvalsDescription' },
      { to: '/admin/identity-policies', titleKey: 'identityPoliciesTitle', descriptionKey: 'identityPoliciesDescription' },
      { to: '/admin/break-glass', titleKey: 'breakGlassTitle', descriptionKey: 'breakGlassDescription' },
    ],
  },
  {
    titleKey: 'configurationTitle',
    links: [
      { to: '/admin/settings', titleKey: 'applicationSettingsTitle', descriptionKey: 'applicationSettingsDescription' },
      { to: '/admin/auth/ldap', titleKey: 'ldapAdTitle', descriptionKey: 'ldapAdDescription' },
      { to: '/admin/auth/oauth2', titleKey: 'oauth2Title', descriptionKey: 'oauth2Description' },
      { to: '/admin/auth/saml', titleKey: 'samlSsoTitle', descriptionKey: 'samlSsoDescription' },
      { to: '/admin/custom-fields', titleKey: 'customFieldsTitle', descriptionKey: 'customFieldsDescription' },
      { to: '/admin/tags', titleKey: 'ipTagsTitle', descriptionKey: 'ipTagsDescription' },
    ],
  },
  {
    titleKey: 'integrationsAutomationTitle',
    links: [
      { to: '/admin/integrations', titleKey: 'integrationsTitle', descriptionKey: 'integrationsDescription' },
      { to: '/admin/webhooks', titleKey: 'webhooksTitle', descriptionKey: 'webhooksDescription' },
      { to: '/admin/integration-templates', titleKey: 'integrationTemplatesTitle', descriptionKey: 'integrationTemplatesDescription' },
      { to: '/admin/automation/policies', titleKey: 'automationPoliciesTitle', descriptionKey: 'automationPoliciesDescription' },
      { to: '/admin/api-token-analytics', titleKey: 'tokenAnalyticsTitle', descriptionKey: 'tokenAnalyticsDescription' },
      { to: '/admin/grafana', titleKey: 'grafanaTitle', descriptionKey: 'grafanaDescription' },
    ],
  },
  {
    titleKey: 'discoveryOperationsTitle',
    links: [
      { to: '/admin/scan-jobs', titleKey: 'scanJobsTitle', descriptionKey: 'scanJobsDescription' },
      { to: '/admin/scan-agents', titleKey: 'scanAgentsTitle', descriptionKey: 'scanAgentsDescription' },
      { to: '/admin/scan-profiles', titleKey: 'scanProfilesTitle', descriptionKey: 'scanProfilesDescription' },
      { to: '/admin/scan-retention', titleKey: 'scanRetentionTitle', descriptionKey: 'scanRetentionDescription' },
      { to: '/admin/discovery/conflicts', titleKey: 'discoveryConflictsTitle', descriptionKey: 'discoveryConflictsDescription' },
      { to: '/admin/topology/hints', titleKey: 'topologyHintsTitle', descriptionKey: 'topologyHintsDescription' },
    ],
  },
  {
    titleKey: 'auditReportsDataTitle',
    links: [
      { to: '/admin/audit-log', titleKey: 'auditLogTitle', descriptionKey: 'auditLogDescription' },
      { to: '/admin/audit/retention', titleKey: 'auditRetentionTitle', descriptionKey: 'auditRetentionDescription' },
      { to: '/admin/privacy/consent-report', titleKey: 'privacyConsentTitle', descriptionKey: 'privacyConsentDescription' },
      { to: '/admin/reports/scheduled', titleKey: 'scheduledReportsTitle', descriptionKey: 'scheduledReportsDescription' },
      { to: '/admin/backups', titleKey: 'backupsTitle', descriptionKey: 'backupsDescription' },
      { to: '/admin/compatibility', titleKey: 'compatibilityTitle', descriptionKey: 'compatibilityDescription' },
      { to: '/admin/overlap-report', titleKey: 'subnetOverlapCheckTitle', descriptionKey: 'subnetOverlapCheckDescription' },
      { to: '/admin/system-health', titleKey: 'systemHealthTitle', descriptionKey: 'systemHealthDescription' },
    ],
  },
]

export function buildAdminSurfaceSections(t) {
  return ADMIN_SURFACE_SECTION_KEYS.map((section) => ({
    title: t(`adminOverviewPage.${section.titleKey}`),
    links: section.links.map((link) => ({
      to: link.to,
      title: t(`adminOverviewPage.${link.titleKey}`),
      description: t(`adminOverviewPage.${link.descriptionKey}`),
    })),
  }))
}

export default function AdminOverviewPage() {
  const { t } = useTranslation()
  const sections = buildAdminSurfaceSections(t)

  return (
    <div className="w-full max-w-7xl mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('adminOverviewPage.title')}</h1>
        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
          {t('adminOverviewPage.subtitle')}
        </p>
      </div>

      <div className="space-y-8">
        {sections.map((network) => (
          <network key={network.title} aria-labelledby={`admin-network-${network.title.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}>
            <h2 id={`admin-network-${network.title.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`} className="text-sm font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {network.title}
            </h2>
            <div className="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {network.links.map((link) => (
                <Link
                  key={link.to}
                  to={link.to}
                  className="block rounded border border-gray-200 bg-white p-4 transition hover:border-blue-300 hover:bg-blue-50 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-700 dark:bg-gray-800 dark:hover:border-blue-500 dark:hover:bg-gray-750"
                >
                  <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{link.title}</span>
                  <span className="mt-1 block text-sm text-gray-600 dark:text-gray-400">{link.description}</span>
                </Link>
              ))}
            </div>
          </network>
        ))}
      </div>
    </div>
  )
}
