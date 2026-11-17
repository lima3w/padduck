import { render, screen, waitFor } from '@testing-library/react'
import DHCPLeasesSection from '../pages/ip/DHCPLeasesSection'
import { FeaturesProvider } from '../context/FeaturesContext'

vi.mock('../api/modules', () => ({ getDHCPLeases: vi.fn() }))
vi.mock('../api/app', () => ({ getFeatures: vi.fn() }))

import { getDHCPLeases } from '../api/modules'
import { getFeatures } from '../api/app'

const lease = { id: 1, ipAddress: '192.0.2.10', macAddress: 'aa:bb', hostname: 'pc', state: 'active', vlanName: '10 users', endsAt: null }

function renderSection(props) {
  return render(<FeaturesProvider><DHCPLeasesSection {...props} /></FeaturesProvider>)
}

describe('DHCPLeasesSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getFeatures.mockResolvedValue({ data: { features: { dhcp: true } } })
  })

  it('lists leases linked to a subnet', async () => {
    getDHCPLeases.mockResolvedValue({ data: [lease] })
    renderSection({ subnetId: '5' })
    expect(await screen.findByText('192.0.2.10')).toBeInTheDocument()
    expect(getDHCPLeases).toHaveBeenCalledWith({ subnet_id: '5' })
    expect(screen.getByText('10 users')).toBeInTheDocument()
  })

  it('queries by IP when given an ipId', async () => {
    getDHCPLeases.mockResolvedValue({ data: [lease] })
    renderSection({ ipId: 12, compact: true })
    await screen.findByText('192.0.2.10')
    expect(getDHCPLeases).toHaveBeenCalledWith({ ip_id: 12 })
  })

  it('renders nothing when there are no leases', async () => {
    getDHCPLeases.mockResolvedValue({ data: [] })
    const { container } = renderSection({ subnetId: '5' })
    await waitFor(() => expect(getDHCPLeases).toHaveBeenCalled())
    expect(container).toBeEmptyDOMElement()
  })

  it('renders nothing when the DHCP module is disabled', async () => {
    getDHCPLeases.mockResolvedValue({ data: [lease] })
    getFeatures.mockResolvedValue({ data: { features: { dhcp: false } } })
    const { container } = renderSection({ subnetId: '5' })
    await waitFor(() => expect(getFeatures).toHaveBeenCalled())
    await waitFor(() => expect(container).toBeEmptyDOMElement())
    expect(screen.queryByText('192.0.2.10')).not.toBeInTheDocument()
  })
})
