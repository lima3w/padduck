import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import DevicesPage from '../pages/DevicesPage'

vi.mock('../api/app', () => ({
  getFeatures: vi.fn(),
}))

vi.mock('../api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

vi.mock('../api/locations', () => ({ getLocations: vi.fn(() => Promise.resolve([])) }))
vi.mock('../api/racks', () => ({ getRacks: vi.fn(() => Promise.resolve([])) }))

vi.mock('../utils/listPrefs', () => ({
  loadPrefs: vi.fn((key, defaults) => defaults ?? {}),
  savePrefs: vi.fn(),
  loadColPrefs: vi.fn((key, defaults) => defaults ?? {}),
  saveColPrefs: vi.fn(),
}))

vi.mock('../utils/storageKeys', () => ({
  getCachedUser: vi.fn(() => ({ id: 1, username: 'admin', role: 'admin' })),
  STORAGE_KEYS: {},
  LEGACY_STORAGE_KEYS: {},
}))

import { api } from '../api/client'
import { getFeatures } from '../api/app'

const deviceRows = [
  { id: 1, hostname: 'edge-router-01', vendor: 'Cisco', model: 'ASR1001', isOnline: true, ipCount: 2 },
  { id: 2, hostname: 'core-switch-01', vendor: 'Arista', model: '7050X', isOnline: false, ipCount: 5 },
]

// Routes the api.get mock by URL so each endpoint the page touches on mount
// gets a sensible response.
function wireApiGet(devices = deviceRows) {
  api.get.mockImplementation((url) => {
    if (url === '/devices') return Promise.resolve({ data: { data: devices, total: devices.length } })
    if (url === '/device-types') return Promise.resolve({ data: [] })
    if (url === '/admin/custom-fields') return Promise.resolve({ data: [] })
    return Promise.resolve({ data: [] })
  })
}

function renderPage() {
  return render(
    <MemoryRouter>
      <DevicesPage />
    </MemoryRouter>
  )
}

describe('DevicesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getFeatures.mockResolvedValue({ data: {} })
    wireApiGet()
  })

  it('lists devices from the API', async () => {
    renderPage()
    expect(await screen.findByText('edge-router-01')).toBeInTheDocument()
    expect(screen.getByText('core-switch-01')).toBeInTheDocument()
  })

  it('shows the empty state when there are no devices', async () => {
    wireApiGet([])
    renderPage()
    expect(await screen.findByText('No devices yet.')).toBeInTheDocument()
  })

  it('runs a filtered search through the search endpoint', async () => {
    const user = userEvent.setup()
    api.post.mockResolvedValue({ data: [deviceRows[0]] })
    renderPage()
    await screen.findByText('edge-router-01')

    await user.type(screen.getByPlaceholderText(/hostname/i), 'edge')
    await user.click(screen.getByRole('button', { name: /^search$/i }))

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/devices/search', expect.objectContaining({ query: 'edge' }))
    })
    expect(screen.getByText('edge-router-01')).toBeInTheDocument()
    expect(screen.queryByText('core-switch-01')).not.toBeInTheDocument()
  })

  it('shows an error banner when loading fails', async () => {
    api.get.mockRejectedValue(new Error('boom'))
    renderPage()
    expect(await screen.findByText('Failed to load devices')).toBeInTheDocument()
  })
})
