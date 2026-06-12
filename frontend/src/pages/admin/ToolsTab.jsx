import { Link } from 'react-router-dom'

export default function ToolsTab({ config }) {
  const featureEnabled = (key) => config?.[key] !== 'false'
  const toolSections = [
    {
      title: 'Data Tools',
      links: [
        { to: '/admin/overlap-report', title: 'Subnet Overlap Check', description: 'Find overlapping subnets across all networks' },
        { to: '/admin/import', title: 'Data Import', description: 'Import subnets, IP addresses, or phpIPAM data' },
        { to: '/admin/export', title: 'Data Export', description: 'Export a full data backup' },
      ],
    },
    {
      title: 'Schema & Taxonomy',
      links: [
        { to: '/admin/custom-fields', title: 'Custom Fields', description: 'Manage extra fields for subnets, IPs, and devices' },
        { to: '/admin/tags', title: 'IP Tags', description: 'Create and manage IP address tags' },
        { to: '/admin/vlan-domains', title: 'VLAN Domains', description: 'Manage VLAN namespace boundaries', visible: featureEnabled('feature_vlans_enabled') },
        { to: '/admin/vlan-groups', title: 'VLAN Groups', description: 'Group VLANs for organization and reporting', visible: featureEnabled('feature_vlans_enabled') },
        { to: '/admin/vlans/usage-report', title: 'VLAN Usage', description: 'Review VLAN allocation and utilization', visible: featureEnabled('feature_vlans_enabled') },
      ],
    },
    {
      title: 'Discovery & Automation',
      links: [
        { to: '/admin/webhooks', title: 'Webhooks', description: 'Configure outbound event delivery' },
        { to: '/admin/integrations', title: 'Integrations', description: 'Integration setup notes and connection checks' },
        { to: '/admin/grafana', title: 'Grafana', description: 'Configure the Grafana datasource integration' },
      ],
    },
    {
      title: 'Authentication',
      links: [
        { to: '/admin/auth/ldap', title: 'LDAP / AD', description: 'Configure LDAP authentication and group mappings' },
        { to: '/admin/auth/oauth2', title: 'OAuth2 / OIDC', description: 'Configure OAuth2 or OpenID Connect login' },
        { to: '/admin/auth/saml', title: 'SAML SSO', description: 'Configure SAML single sign-on' },
        { to: '/admin/identity-policies', title: 'Identity Policies', description: 'Manage IP-based access control and identity rules' },
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
                    <span className="ml-auto shrink-0 text-blue-600 dark:text-blue-400 text-sm">Open →</span>
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
  )
}
