import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import SubnetsPage from '../pages/SubnetsPage'

vi.mock('../api/admin', () => ({
  getCustomFields: vi.fn(),
}))

vi.mock('../api/client', () => ({
  api: { get: vi.fn(), post: vi.fn() },
}))

vi.mock('../api/dns', () => ({
  getNameservers: vi.fn(),
}))

vi.mock('../api/ipam', () => ({
  getNetwork: vi.fn(),
  getSubnet: vi.fn(),
  getSubnetsPaginated: vi.fn(),
  createSubnet: vi.fn(),
  updateSubnet: vi.fn(),
  deleteSubnet: vi.fn(),
  searchSubnets: vi.fn(),
  getSubnetTree: vi.fn(),
}))

vi.mock('../api/vlans', () => ({
  getVlans: vi.fn(),
}))

vi.mock('../api/locations', () => ({ getLocations: vi.fn(() => Promise.resolve([])) }))

vi.mock('../utils/listPrefs', () => ({
  loadPrefs: vi.fn((key, defaults) => defaults ?? {}),
  savePrefs: vi.fn(),
}))

vi.mock('../utils/storageKeys', () => ({
  getCachedUser: vi.fn(() => ({ id: 1, username: 'admin', role: 'admin' })),
  STORAGE_KEYS: {},
  LEGACY_STORAGE_KEYS: {},
}))

import { getNetwork, getSubnetsPaginated } from '../api/ipam'
import { getNameservers } from '../api/dns'
import { getVlans } from '../api/vlans'
import { getCustomFields } from '../api/admin'

const subnets = [
  { id: 11, networkAddress: '10.10.0.0', prefixLength: 24, description: 'office lan', totalIps: 254, usedIps: 10 },
  { id: 12, networkAddress: '10.20.0.0', prefixLength: 25, description: 'lab lan', totalIps: 126, usedIps: 5 },
]

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/networks/7/subnets']}>
      <Routes>
        <Route path="/networks/:networkID/subnets" element={<SubnetsPage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('SubnetsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getNetwork.mockResolvedValue({ data: { id: 7, name: 'Lab Network' } })
    getSubnetsPaginated.mockResolvedValue({ data: { data: subnets, total: subnets.length } })
    getNameservers.mockResolvedValue({ data: [] })
    getVlans.mockResolvedValue({ data: [] })
    getCustomFields.mockResolvedValue({ data: [] })
  })

  it('lists the subnets of the network', async () => {
    renderPage()
    expect(await screen.findByText(/10\.10\.0\.0/)).toBeInTheDocument()
    expect(screen.getByText(/10\.20\.0\.0/)).toBeInTheDocument()
    expect(screen.getByText('Lab Network')).toBeInTheDocument()
    expect(getSubnetsPaginated).toHaveBeenCalledWith('7', 1, expect.any(Number))
  })

  it('shows the empty state when the network has no subnets', async () => {
    getSubnetsPaginated.mockResolvedValue({ data: { data: [], total: 0 } })
    renderPage()
    expect(await screen.findByText('No subnets yet.')).toBeInTheDocument()
  })

  it('shows an error banner when loading fails', async () => {
    getSubnetsPaginated.mockRejectedValue(new Error('boom'))
    getNetwork.mockRejectedValue(new Error('boom'))
    renderPage()
    expect(await screen.findByText('Failed to load subnets')).toBeInTheDocument()
  })
})
