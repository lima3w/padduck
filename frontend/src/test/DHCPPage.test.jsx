import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import DHCPPage from '../pages/DHCPPage'

vi.mock('../api/modules', () => ({
  getDHCPServers: vi.fn(),
  getDHCPLeases: vi.fn(),
  createDHCPServer: vi.fn(),
  updateDHCPServer: vi.fn(),
  deleteDHCPServer: vi.fn(),
  createDHCPLease: vi.fn(),
  updateDHCPLease: vi.fn(),
  deleteDHCPLease: vi.fn(),
}))
vi.mock('../api/locations', () => ({ getLocations: vi.fn() }))
vi.mock('../api/vlans', () => ({ getVlans: vi.fn() }))
vi.mock('../api/admin', () => ({ syncTechnitiumLeases: vi.fn() }))
vi.mock('../utils/storageKeys', () => ({ getCachedUser: () => ({ role: 'admin' }) }))

import { getDHCPServers, getDHCPLeases, updateDHCPServer } from '../api/modules'
import { getLocations } from '../api/locations'
import { getVlans } from '../api/vlans'
import { syncTechnitiumLeases } from '../api/admin'

const technitium = {
  id: 1, name: 'Technitium', address: 'http://10.0.0.5', vendor: 'technitium', status: 'active',
  autoSync: true, vlanId: 3, vlanName: '10 users', lastSyncedAt: '2026-07-17T10:00:00Z', lastSyncError: 'scope bad failed',
}
const lease = { id: 9, serverId: 1, serverName: 'Technitium', ipAddress: '192.0.2.10', macAddress: 'aa:bb', hostname: 'pc', state: 'active', vlanId: 3, vlanName: '10 users' }

describe('DHCPPage: Technitium sync and VLANs', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getDHCPServers.mockResolvedValue({ data: [technitium] })
    getDHCPLeases.mockResolvedValue({ data: [lease] })
    getLocations.mockResolvedValue([])
    getVlans.mockResolvedValue({ data: [{ id: 3, vlanId: 10, name: 'users' }, { id: 4, vlanId: 20, name: 'iot' }] })
  })

  it('shows VLANs and the sync status, including the last error', async () => {
    render(<DHCPPage />)
    expect(await screen.findByText('192.0.2.10')).toBeInTheDocument()
    expect(screen.getAllByText('10 users').length).toBeGreaterThanOrEqual(2) // server row + lease row
    expect(screen.getByText(/Last synced/)).toBeInTheDocument()
    expect(screen.getByText(/Last sync error: scope bad failed/)).toBeInTheDocument()
  })

  it('filters leases by VLAN via the API', async () => {
    const user = userEvent.setup()
    render(<DHCPPage />)
    await screen.findByText('192.0.2.10')
    await user.selectOptions(screen.getByLabelText('Filter leases by VLAN'), '4')
    await waitFor(() => expect(getDHCPLeases).toHaveBeenLastCalledWith({ vlan_id: '4' }))
  })

  it('runs a manual sync and reports the count', async () => {
    const user = userEvent.setup()
    syncTechnitiumLeases.mockResolvedValue({ data: { synced: 5 } })
    render(<DHCPPage />)
    await user.click(await screen.findByRole('button', { name: 'Sync now' }))
    expect(await screen.findByText('Synced 5 leases')).toBeInTheDocument()
  })

  it('saves the server VLAN and auto-sync opt-out', async () => {
    const user = userEvent.setup()
    updateDHCPServer.mockResolvedValue({ data: {} })
    render(<DHCPPage />)
    await screen.findByText('192.0.2.10')
    await user.click(screen.getAllByRole('button', { name: 'Edit' })[0])
    await user.selectOptions(screen.getByLabelText('VLAN'), '4')
    await user.click(screen.getByLabelText('Sync leases automatically'))
    await user.click(screen.getByRole('button', { name: 'Save' }))
    await waitFor(() => expect(updateDHCPServer).toHaveBeenCalled())
    expect(updateDHCPServer.mock.calls[0][1]).toMatchObject({ vlan_id: 4, auto_sync: false })
  })
})
