import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import DeviceDetailPage from '../pages/DeviceDetailPage'

vi.mock('../api/client', () => ({
  getDevice: vi.fn(),
  updateDevice: vi.fn(),
  getDeviceTypes: vi.fn(),
  getDeviceIPs: vi.fn(),
  associateDeviceIP: vi.fn(),
  disassociateDeviceIP: vi.fn(),
  getDeviceInterfaces: vi.fn(),
  createDeviceInterface: vi.fn(),
  updateDeviceInterface: vi.fn(),
  deleteDeviceInterface: vi.fn(),
  getCustomFields: vi.fn(),
  getDeviceSNMPCredentials: vi.fn(),
  searchIPAddressesGlobal: vi.fn(),
}))

vi.mock('../api/locations', () => ({ getLocations: vi.fn(() => Promise.resolve([])) }))
vi.mock('../api/racks', () => ({ getRacks: vi.fn(() => Promise.resolve([])) }))
vi.mock('../utils/storageKeys', () => ({
  getCachedUser: vi.fn(() => ({ id: 1, username: 'admin', role: 'admin' })),
}))

// Peripheral panels make their own API calls and have their own tests.
vi.mock('../components/ChangeHistory', () => ({ default: () => null }))
vi.mock('../components/FingerprintPanel', () => ({ default: () => null }))
vi.mock('../components/ObjectRelationshipsPanel', () => ({ default: () => null }))

import { getDevice, getDeviceTypes, getDeviceIPs, getDeviceInterfaces, getCustomFields } from '../api/client'

const device = {
  id: 5,
  hostname: 'core-sw-01',
  vendor: 'Juniper',
  model: 'EX4300',
  isOnline: true,
  customFields: {
    runbook: 'https://wiki.example.com/core-sw-01',
    evil: 'javascript:alert(1)',
  },
}

const cfDefs = [
  { id: 1, name: 'runbook', label: 'Runbook', fieldType: 'url' },
  { id: 2, name: 'evil', label: 'Evil Link', fieldType: 'url' },
]

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/devices/5']}>
      <Routes>
        <Route path="/devices/:id" element={<DeviceDetailPage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('DeviceDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getDevice.mockResolvedValue({ data: device })
    getDeviceTypes.mockResolvedValue({ data: [] })
    getDeviceIPs.mockResolvedValue({ data: [] })
    getDeviceInterfaces.mockResolvedValue({ data: [] })
    getCustomFields.mockResolvedValue({ data: cfDefs })
  })

  it('renders the device summary', async () => {
    renderPage()
    expect(await screen.findByRole('heading', { name: 'core-sw-01' })).toBeInTheDocument()
    expect(screen.getByText(/Juniper/)).toBeInTheDocument()
  })

  it('links http(s) URL custom fields but renders javascript: values inert', async () => {
    renderPage()
    await screen.findByRole('heading', { name: 'core-sw-01' })

    // Safe URL renders as a real link.
    const link = screen.getByRole('link', { name: 'https://wiki.example.com/core-sw-01' })
    expect(link).toHaveAttribute('href', 'https://wiki.example.com/core-sw-01')

    // The javascript: value is plain text — no anchor anywhere points at it.
    expect(screen.getByText('javascript:alert(1)')).toBeInTheDocument()
    const links = screen.getAllByRole('link')
    for (const a of links) {
      expect(a).not.toHaveAttribute('href', 'javascript:alert(1)')
    }
  })

  it('shows the API error when the device fails to load', async () => {
    getDevice.mockRejectedValue({ response: { data: { error: 'device not found' } } })
    renderPage()
    expect(await screen.findByText(/device not found/)).toBeInTheDocument()
  })
})
