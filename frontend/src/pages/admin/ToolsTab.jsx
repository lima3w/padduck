import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function ToolsTab({ config }) {
  const { t } = useTranslation()
  const featureEnabled = (key) => config?.[key] !== 'false'
  const toolSections = [
    {
      title: t('toolsTab.sections.dataTools.title'),
      links: [
        { to: '/admin/overlap-report', title: t('toolsTab.sections.dataTools.links.overlapReport.title'), description: t('toolsTab.sections.dataTools.links.overlapReport.description') },
        { to: '/admin/import', title: t('toolsTab.sections.dataTools.links.dataImport.title'), description: t('toolsTab.sections.dataTools.links.dataImport.description') },
        { to: '/admin/export', title: t('toolsTab.sections.dataTools.links.dataExport.title'), description: t('toolsTab.sections.dataTools.links.dataExport.description') },
      ],
    },
    {
      title: t('toolsTab.sections.schemaTaxonomy.title'),
      links: [
        { to: '/admin/custom-fields', title: t('toolsTab.sections.schemaTaxonomy.links.customFields.title'), description: t('toolsTab.sections.schemaTaxonomy.links.customFields.description') },
        { to: '/admin/tags', title: t('toolsTab.sections.schemaTaxonomy.links.ipTags.title'), description: t('toolsTab.sections.schemaTaxonomy.links.ipTags.description') },
        { to: '/admin/vlan-domains', title: t('toolsTab.sections.schemaTaxonomy.links.vlanDomains.title'), description: t('toolsTab.sections.schemaTaxonomy.links.vlanDomains.description'), visible: featureEnabled('feature_vlans_enabled') },
        { to: '/admin/vlan-groups', title: t('toolsTab.sections.schemaTaxonomy.links.vlanGroups.title'), description: t('toolsTab.sections.schemaTaxonomy.links.vlanGroups.description'), visible: featureEnabled('feature_vlans_enabled') },
        { to: '/admin/vlans/usage-report', title: t('toolsTab.sections.schemaTaxonomy.links.vlanUsage.title'), description: t('toolsTab.sections.schemaTaxonomy.links.vlanUsage.description'), visible: featureEnabled('feature_vlans_enabled') },
      ],
    },
    {
      title: t('toolsTab.sections.discoveryAutomation.title'),
      links: [
        { to: '/admin/webhooks', title: t('toolsTab.sections.discoveryAutomation.links.webhooks.title'), description: t('toolsTab.sections.discoveryAutomation.links.webhooks.description') },
        { to: '/admin/integrations', title: t('toolsTab.sections.discoveryAutomation.links.integrations.title'), description: t('toolsTab.sections.discoveryAutomation.links.integrations.description') },
        { to: '/admin/grafana', title: t('toolsTab.sections.discoveryAutomation.links.grafana.title'), description: t('toolsTab.sections.discoveryAutomation.links.grafana.description') },
      ],
    },
    {
      title: t('toolsTab.sections.authentication.title'),
      links: [
        { to: '/admin/auth/ldap', title: t('toolsTab.sections.authentication.links.ldap.title'), description: t('toolsTab.sections.authentication.links.ldap.description') },
        { to: '/admin/auth/oauth2', title: t('toolsTab.sections.authentication.links.oauth2.title'), description: t('toolsTab.sections.authentication.links.oauth2.description') },
        { to: '/admin/auth/saml', title: t('toolsTab.sections.authentication.links.saml.title'), description: t('toolsTab.sections.authentication.links.saml.description') },
        { to: '/admin/identity-policies', title: t('toolsTab.sections.authentication.links.identityPolicies.title'), description: t('toolsTab.sections.authentication.links.identityPolicies.description') },
      ],
    },
  ]
    .map((section) => ({
      ...section,
      links: section.links.filter((link) => link.visible !== false),
    }))
    .filter((section) => section.links.length > 0)

  return (
        <div className="space-y-4">
          {toolSections.map((network) => (
            <div key={network.title} className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">{network.title}</h2>
              <div className="space-y-3">
                {network.links.map((link) => (
                  <Link
                    key={link.to}
                    to={link.to}
                    className="flex items-center gap-3 p-3 rounded border border-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 transition"
                  >
                    <div>
                      <p className="font-medium text-gray-900 dark:text-gray-100">{link.title}</p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">{link.description}</p>
                    </div>
                    <span className="ml-auto shrink-0 text-blue-600 dark:text-blue-400 text-sm">{t('toolsTab.openArrow')}</span>
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
  )
}
