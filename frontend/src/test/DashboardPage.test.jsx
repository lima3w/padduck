import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import DashboardPage from '../pages/DashboardPage'

vi.mock('../api/admin', () => ({
  getInactiveIPs: vi.fn(),
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
