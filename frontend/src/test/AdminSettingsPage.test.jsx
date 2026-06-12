import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import AdminSettingsPage from '../pages/admin/AdminSettingsPage'

vi.mock('../api/admin', () => ({
  getAdminConfig: vi.fn(),
  updateAdminConfig: vi.fn(),
  listPendingApprovals: vi.fn(),
  approveUser: vi.fn(),
  rejectUser: vi.fn(),
  testSMTP: vi.fn(),
  purgeAuditLogs: vi.fn(),
  getNotificationStats: vi.fn(),
  checkForUpdates: vi.fn(),
}))

vi.mock('../api/dns', () => ({
  checkAllDns: vi.fn(),
  testDnsConnection: vi.fn(),
  testTechnitiumConnection: vi.fn(),
}))

import { getAdminConfig, updateAdminConfig, listPendingApprovals, approveUser } from '../api/admin'

const config = {
  app_url: 'https://padduck.example.com',
  registration_enabled: 'true',
  smtp_host: 'mail.example.com',
  audit_log_retention_days: '90',
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminSettingsPage />
    </MemoryRouter>
  )
}

describe('AdminSettingsPage shell', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getAdminConfig.mockResolvedValue({ data: { config } })
    listPendingApprovals.mockResolvedValue({ data: { approvals: [] } })
  })

  it('renders the registration tab by default with loaded config', async () => {
    renderPage()
    expect(await screen.findByText('Registration Settings')).toBeInTheDocument()
    expect(screen.getByDisplayValue('https://padduck.example.com')).toBeInTheDocument()
  })

  it('switches tabs and renders the matching tab component', async () => {
    const user = userEvent.setup()
    renderPage()
    await screen.findByText('Registration Settings')

    await user.click(screen.getByRole('button', { name: 'SMTP / Email' }))
    expect(screen.getByText('SMTP Configuration')).toBeInTheDocument()
    expect(screen.queryByText('Registration Settings')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Audit' }))
    expect(screen.getByText('Audit Log Retention')).toBeInTheDocument()
  })

  it('saves only the active tab keys', async () => {
    const user = userEvent.setup()
    updateAdminConfig.mockResolvedValue({ data: {} })
    renderPage()
    await screen.findByText('Registration Settings')

    await user.click(screen.getByRole('button', { name: 'Save Settings' }))
    await waitFor(() => expect(updateAdminConfig).toHaveBeenCalledTimes(1))
    const payload = updateAdminConfig.mock.calls[0][0]
    expect(payload).toHaveProperty('app_url')
    expect(payload).toHaveProperty('registration_enabled')
    expect(payload).not.toHaveProperty('smtp_host', undefined)
    expect(Object.keys(payload)).not.toContain('smtp_host')
    expect(await screen.findByText('Settings saved successfully')).toBeInTheDocument()
  })

  it('shows the pending approval count in the tab label and approves users', async () => {
    const user = userEvent.setup()
    listPendingApprovals.mockResolvedValue({
      data: { approvals: [{ id: 9, username: 'newbie', email: 'n@example.com', createdAt: new Date().toISOString() }] },
    })
    approveUser.mockResolvedValue({ data: {} })
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Approvals (1)' }))
    expect(screen.getByText('newbie')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Approve' }))
    await waitFor(() => expect(approveUser).toHaveBeenCalledWith(9))
    expect(await screen.findByText('No pending approvals')).toBeInTheDocument()
  })
})
