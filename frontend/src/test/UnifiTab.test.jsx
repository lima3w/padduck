import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import UnifiTab from '../pages/admin/UnifiTab'

vi.mock('../api/unifi', () => ({
  testUniFiConnection: vi.fn(),
  syncUniFi: vi.fn(),
  getUniFiStatus: vi.fn(),
  getUniFiConflicts: vi.fn(),
  resolveUniFiConflict: vi.fn(),
}))

import { testUniFiConnection, syncUniFi, getUniFiStatus, getUniFiConflicts, resolveUniFiConflict } from '../api/unifi'

const conflict = { id: 4, ipAddress: '10.0.10.7', field: 'hostname', padduckValue: 'printer', unifiValue: 'laptop' }

function renderTab(config = {}) {
  const handleConfigChange = vi.fn()
  const handleSaveConfig = vi.fn()
  render(<UnifiTab config={config} handleConfigChange={handleConfigChange} handleSaveConfig={handleSaveConfig} saving={false} />)
  return { handleConfigChange, handleSaveConfig }
}

describe('UnifiTab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getUniFiStatus.mockResolvedValue({ data: { configured: true, lastSyncedAt: null, lastError: '', openConflicts: 1 } })
    getUniFiConflicts.mockResolvedValue({ data: [conflict] })
  })

  it('reports field changes under the unifi_* config keys', async () => {
    const user = userEvent.setup()
    const { handleConfigChange } = renderTab()
    await user.type(screen.getByLabelText('Controller URL'), 'h')
    expect(handleConfigChange).toHaveBeenCalledWith('unifi_url', 'h')
    await user.click(screen.getByLabelText('Skip TLS certificate verification'))
    expect(handleConfigChange).toHaveBeenCalledWith('unifi_skip_tls', 'true')
  })

  it('defaults the scheduled interval to 0 (off)', async () => {
    renderTab()
    expect(screen.getByLabelText('Scheduled sync interval (minutes)')).toHaveValue(0)
    await screen.findByText('10.0.10.7')
  })

  it('tests the connection with the current form values', async () => {
    const user = userEvent.setup()
    testUniFiConnection.mockResolvedValue({ data: { status: 'ok' } })
    renderTab({ unifi_url: 'https://u', unifi_api_key: '********', unifi_site: 'lab', unifi_skip_tls: 'true' })
    await user.click(screen.getByRole('button', { name: 'Test Connection' }))
    expect(await screen.findByText('Connected')).toBeInTheDocument()
    expect(testUniFiConnection).toHaveBeenCalledWith({
      url: 'https://u', api_key: '********', username: '', password: '', site: 'lab', skip_tls: true,
    })
  })

  it('shows the server error when the connection test fails', async () => {
    const user = userEvent.setup()
    testUniFiConnection.mockRejectedValue({ response: { data: { error: 'UniFi connection failed: site "x" not found' } } })
    renderTab({ unifi_url: 'https://u' })
    await user.click(screen.getByRole('button', { name: 'Test Connection' }))
    expect(await screen.findByText(/site "x" not found/)).toBeInTheDocument()
  })

  it('runs a sync and reports the totals', async () => {
    const user = userEvent.setup()
    syncUniFi.mockResolvedValue({ data: { vlansLinked: 3, clientsCreated: 2, clientsUpdated: 1, devicesCreated: 1, devicesLinked: 1, devicesUpdated: 0, conflictsRaised: 2 } })
    renderTab()
    await user.click(await screen.findByRole('button', { name: 'Sync now' }))
    expect(await screen.findByText('Sync complete: 3 VLANs, 3 clients, 2 devices, 2 new conflicts')).toBeInTheDocument()
  })

  it('disables sync until UniFi is configured', async () => {
    getUniFiStatus.mockResolvedValue({ data: { configured: false, lastError: '' } })
    renderTab()
    await waitFor(() => expect(screen.getByRole('button', { name: 'Sync now' })).toBeDisabled())
  })

  it('lists conflicts and resolves them either way', async () => {
    const user = userEvent.setup()
    resolveUniFiConflict.mockResolvedValue({ data: {} })
    renderTab()
    expect(await screen.findByText('printer')).toBeInTheDocument()
    expect(screen.getByText('laptop')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Use UniFi value' }))
    await waitFor(() => expect(resolveUniFiConflict).toHaveBeenCalledWith(4, 'accept'))
    await user.click(screen.getByRole('button', { name: 'Keep Padduck value' }))
    await waitFor(() => expect(resolveUniFiConflict).toHaveBeenCalledWith(4, 'dismiss'))
  })

  it('shows the last sync error and per-item errors', async () => {
    getUniFiStatus.mockResolvedValue({ data: { configured: true, lastSyncedAt: '2026-07-18T10:00:00Z', lastError: 'listing clients: boom', lastResult: { errors: ['creating IP 10.0.0.1: nope'], clientsUnmatched: 4 } } })
    renderTab()
    expect(await screen.findByText(/Last sync error: listing clients: boom/)).toBeInTheDocument()
    expect(screen.getByText('creating IP 10.0.0.1: nope')).toBeInTheDocument()
    expect(screen.getByText(/4 clients were skipped/)).toBeInTheDocument()
  })
})
