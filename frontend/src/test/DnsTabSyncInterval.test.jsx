import { render, screen, fireEvent } from '@testing-library/react'
import DnsTab from '../pages/admin/DnsTab'

vi.mock('../api/dns', () => ({
  checkAllDns: vi.fn(), createNameserver: vi.fn(), testDnsConnection: vi.fn(), testTechnitiumConnection: vi.fn(),
}))
vi.mock('../api/admin', () => ({
  getTechnitiumDHCPScopes: vi.fn(), importTechnitiumScope: vi.fn(), syncTechnitiumLeases: vi.fn(),
}))
vi.mock('../api/ipam', () => ({ getNetworks: vi.fn() }))

function renderTab(config) {
  const handleConfigChange = vi.fn()
  render(<DnsTab config={config} handleConfigChange={handleConfigChange} handleSaveConfig={vi.fn()} saving={false} showMessage={vi.fn()} />)
  return handleConfigChange
}

describe('DnsTab: Technitium DHCP sync interval', () => {
  it('shows the 15 minute default when unset', () => {
    renderTab({})
    expect(screen.getByLabelText('DHCP lease sync interval (minutes)')).toHaveValue(15)
  })

  it('shows a stored value, including 0 (disabled)', () => {
    renderTab({ technitium_dhcp_sync_interval_minutes: '0' })
    expect(screen.getByLabelText('DHCP lease sync interval (minutes)')).toHaveValue(0)
  })

  it('reports changes under the config key', () => {
    const onChange = renderTab({})
    fireEvent.change(screen.getByLabelText('DHCP lease sync interval (minutes)'), { target: { value: '30' } })
    expect(onChange).toHaveBeenCalledWith('technitium_dhcp_sync_interval_minutes', '30')
  })
})
