import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import DashboardPage from '../pages/DashboardPage'

vi.mock('../api/admin', () => ({
  getInactiveIPs: vi.fn(),
  getAdminConfig: vi.fn(() => Promise.resolve({ data: { config: { telemetry_enabled: 'false' } } })),
}))

vi.mock('../api/app', () => ({
  getDashboardSummary: vi.fn(),
  getDashboardRecentActivity: vi.fn(),
}))

vi.mock('../api/client', () => ({
  api: { get: vi.fn() },
}))

vi.mock('../utils/storageKeys', () => ({
  getCachedUser: vi.fn(() => ({ id: 1, username: 'admin', role: 'admin' })),
}))

import { api } from '../api/client'
import { getDashboardSummary, getDashboardRecentActivity } from '../api/app'
import { getAdminConfig } from '../api/admin'

const summary = {
  totalNetworks: 3,
  totalSubnets: 12,
  totalIPs: 500,
  assignedIPs: 250,
  utilizationPct: 50,
}

const activity = [
  { id: 1, action: 'ip_assigned', description: '10.0.0.5 assigned', createdAt: new Date().toISOString() },
  { id: 2, action: 'subnet_created', description: '10.1.0.0/24 created', createdAt: new Date().toISOString() },
]

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <DashboardPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getDashboardSummary.mockResolvedValue({ data: summary })
    getDashboardRecentActivity.mockResolvedValue({ data: activity })
    api.get.mockResolvedValue({ data: [] })
  })

  it('renders summary cards from the API', async () => {
    renderPage()
    expect(await screen.findByText('Networks')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('Subnets')).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
  })

  it('renders recent activity entries', async () => {
    renderPage()
    expect(await screen.findByText(/10\.0\.0\.5 assigned/)).toBeInTheDocument()
    expect(screen.getByText(/10\.1\.0\.0\/24 created/)).toBeInTheDocument()
  })

  it('shows an error state with a working retry button', async () => {
    const user = userEvent.setup()
    getDashboardSummary.mockRejectedValueOnce(new Error('boom'))
    renderPage()

    expect(await screen.findByText('Failed to load dashboard data')).toBeInTheDocument()

    // Retry refetches and recovers.
    await user.click(screen.getByRole('button', { name: 'Retry' }))
    expect(await screen.findByText('Networks')).toBeInTheDocument()
    await waitFor(() => expect(getDashboardSummary).toHaveBeenCalledTimes(2))
  })
})

describe('DashboardPage first-run wizard redirect', () => {
  function renderWithSetupRoute() {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    return render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/dashboard']}>
          <Routes>
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/setup" element={<div>Setup wizard</div>} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    )
  }

  beforeEach(() => {
    vi.clearAllMocks()
    getDashboardSummary.mockResolvedValue({ data: summary })
    getDashboardRecentActivity.mockResolvedValue({ data: activity })
    api.get.mockResolvedValue({ data: [] })
  })

  it('sends admins to /setup on a fresh install', async () => {
    getAdminConfig.mockResolvedValue({ data: { config: {} } })
    renderWithSetupRoute()
    expect(await screen.findByText('Setup wizard')).toBeInTheDocument()
  })

  it('does not redirect once the wizard was completed or dismissed', async () => {
    getAdminConfig.mockResolvedValue({ data: { config: { setup_wizard_completed: 'true' } } })
    renderWithSetupRoute()
    await waitFor(() => expect(getAdminConfig).toHaveBeenCalled())
    expect(screen.queryByText('Setup wizard')).not.toBeInTheDocument()
  })

  it('does not redirect existing installs that already answered telemetry', async () => {
    getAdminConfig.mockResolvedValue({ data: { config: { telemetry_enabled: 'false' } } })
    renderWithSetupRoute()
    await waitFor(() => expect(getAdminConfig).toHaveBeenCalled())
    expect(screen.queryByText('Setup wizard')).not.toBeInTheDocument()
  })
})
