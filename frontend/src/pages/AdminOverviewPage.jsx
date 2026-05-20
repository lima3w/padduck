import { Link } from 'react-router-dom'

export const ADMIN_SURFACE_SECTIONS = [
  {
    title: 'Identity & Access',
    links: [
      { to: '/admin/users', title: 'Users', description: 'Manage accounts, states, and user roles.' },
      { to: '/admin/roles', title: 'Roles', description: 'Create and maintain custom role permissions.' },
      { to: '/admin/roles/presets', title: 'Permission Presets', description: 'Review role preset changes before applying them.' },
      { to: '/admin/requests', title: 'Approvals', description: 'Review pending subnet and IP address requests.' },
      { to: '/admin/identity-policies', title: 'Identity Policies', description: 'Set MFA, token, session, and inactive-user policies.' },
      { to: '/admin/break-glass', title: 'Break-Glass', description: 'Start and audit emergency administrator access.' },
    ],
  },
  {
    title: 'Configuration',
    links: [
      { to: '/admin/settings', title: 'Application Settings', description: 'Configure registration, email, DNS, scanner, features, and updates.' },
      { to: '/admin/auth/ldap', title: 'LDAP / AD', description: 'Configure LDAP authentication and group mappings.' },
      { to: '/admin/auth/oauth2', title: 'OAuth2 / OIDC', description: 'Configure OAuth2 or OpenID Connect login.' },
      { to: '/admin/auth/saml', title: 'SAML SSO', description: 'Configure SAML single sign-on.' },
      { to: '/admin/custom-fields', title: 'Custom Fields', description: 'Manage extra fields for subnets, IPs, and devices.' },
      { to: '/admin/tags', title: 'IP Tags', description: 'Create and maintain IP address tags.' },
    ],
  },
  {
    title: 'Integrations & Automation',
    links: [
      { to: '/admin/integrations', title: 'Integrations', description: 'Review integration setup notes and connection checks.' },
      { to: '/admin/webhooks', title: 'Webhooks', description: 'Configure outbound event delivery.' },
      { to: '/admin/integration-templates', title: 'Integration Templates', description: 'Use common automation platform templates.' },
      { to: '/admin/automation/policies', title: 'Automation Policies', description: 'Evaluate approval and validation rules.' },
      { to: '/admin/api-token-analytics', title: 'Token Analytics', description: 'Review API token usage and rate-limit visibility.' },
      { to: '/admin/grafana', title: 'Grafana', description: 'Configure the Grafana datasource integration.' },
    ],
  },
  {
    title: 'Discovery & Operations',
    links: [
      { to: '/admin/scan-jobs', title: 'Scan Jobs', description: 'Schedule and run network discovery scans.' },
      { to: '/admin/scan-agents', title: 'Scan Agents', description: 'Manage remote discovery agents and tokens.' },
      { to: '/admin/scan-profiles', title: 'Scan Profiles', description: 'Maintain reusable scan configurations.' },
      { to: '/admin/scan-retention', title: 'Scan Retention', description: 'Configure scan result retention and pruning.' },
      { to: '/admin/discovery/conflicts', title: 'Discovery Conflicts', description: 'Review observed data conflicts.' },
      { to: '/admin/topology/hints', title: 'Topology Hints', description: 'Manage topology relationships discovered from scans.' },
    ],
  },
  {
    title: 'Audit, Reports & Data',
    links: [
      { to: '/admin/audit-log', title: 'Audit Log', description: 'Search and export administrative activity.' },
      { to: '/admin/audit/retention', title: 'Audit Retention', description: 'Configure audit retention, pruning, and archive exports.' },
      { to: '/admin/privacy/consent-report', title: 'Privacy Consent', description: 'Track privacy versions and user consent status.' },
      { to: '/admin/reports/scheduled', title: 'Scheduled Reports', description: 'Manage recurring emailed reports.' },
      { to: '/admin/import', title: 'Data Import', description: 'Import subnets, IP addresses, or phpIPAM data.' },
      { to: '/admin/export', title: 'Data Export', description: 'Export a full data backup.' },
      { to: '/admin/overlap-report', title: 'Subnet Overlap Check', description: 'Find overlapping subnets across all sections.' },
      { to: '/admin/system-health', title: 'System Health', description: 'Review deployment, backup, and dependency health.' },
    ],
  },
]

export default function AdminOverviewPage() {
  return (
    <div className="w-full max-w-7xl mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Admin</h1>
        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Central access to administrative workflows.
        </p>
      </div>

      <div className="space-y-8">
        {ADMIN_SURFACE_SECTIONS.map((section) => (
          <section key={section.title} aria-labelledby={`admin-section-${section.title.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}>
            <h2 id={`admin-section-${section.title.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`} className="text-sm font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {section.title}
            </h2>
            <div className="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {section.links.map((link) => (
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
          </section>
        ))}
      </div>
    </div>
  )
}
