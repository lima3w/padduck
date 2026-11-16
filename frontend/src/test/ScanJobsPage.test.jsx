import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ScanJobsPage from '../pages/ScanJobsPage'

vi.mock('../api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

vi.mock('../api/ipam', () => ({
  getNetworks: vi.fn(),
  createSubnet: vi.fn(),
}))

import { api } from '../api/client'
import { getNetworks, createSubnet } from '../api/ipam'

const notFound = () => Object.assign(new Error('bad request'), {
  response: { status: 400, data: { error: 'subnet not found for CIDR: 10.9.0.0/24' } },
})

async function openCreateAndSubmit(user, cidr = '10.9.0.0/24') {
  await user.click(await screen.findByRole('button', { name: '+ New Job' }))
  await user.type(screen.getByPlaceholderText('Office subnet scan'), 'Office scan')
  await user.type(screen.getByPlaceholderText('192.168.1.0/24'), cidr)
  await user.click(screen.getByRole('button', { name: 'Create' }))
}

describe('ScanJobsPage: unknown subnet CIDR', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.get.mockResolvedValue({ data: [] })
    getNetworks.mockResolvedValue({ data: [{ id: 7, name: 'Lab' }, { id: 8, name: 'Prod' }] })
  })

  it('offers to create the subnet, then creates it and retries the scan job', async () => {
    const user = userEvent.setup()
    api.post.mockRejectedValueOnce(notFound()).mockResolvedValueOnce({ data: {} })
    createSubnet.mockResolvedValue({ data: {} })

    render(<ScanJobsPage />)
    await openCreateAndSubmit(user)

    expect(await screen.findByText('Subnet not found')).toBeInTheDocument()
    await user.selectOptions(screen.getByLabelText('Create it in network'), '8')
    await user.click(screen.getByRole('button', { name: 'Create subnet and scan job' }))

    await waitFor(() => expect(createSubnet).toHaveBeenCalledWith('8', { network_address: '10.9.0.0', prefix_length: 24 }))
    await waitFor(() => expect(api.post).toHaveBeenCalledTimes(2))
    await waitFor(() => expect(screen.queryByText('Subnet not found')).not.toBeInTheDocument())
  })

  it('cancel backs out of scan job creation entirely', async () => {
    const user = userEvent.setup()
    api.post.mockRejectedValueOnce(notFound())

    render(<ScanJobsPage />)
    await openCreateAndSubmit(user)
    await screen.findByText('Subnet not found')
    await user.click(screen.getAllByRole('button', { name: 'Cancel' }).at(-1))

    expect(screen.queryByText('Subnet not found')).not.toBeInTheDocument()
    expect(screen.queryByText('New Scan Job')).not.toBeInTheDocument()
    expect(createSubnet).not.toHaveBeenCalled()
  })

  it('does not offer subnet creation for other create errors', async () => {
    const user = userEvent.setup()
    api.post.mockRejectedValueOnce(Object.assign(new Error('x'), { response: { status: 400, data: { error: 'name is required' } } }))

    render(<ScanJobsPage />)
    await openCreateAndSubmit(user)

    expect(await screen.findByText('name is required')).toBeInTheDocument()
    expect(screen.queryByText('Subnet not found')).not.toBeInTheDocument()
  })

  it('disables creation when there are no networks', async () => {
    const user = userEvent.setup()
    getNetworks.mockResolvedValue({ data: [] })
    api.post.mockRejectedValueOnce(notFound())

    render(<ScanJobsPage />)
    await openCreateAndSubmit(user)

    expect(await screen.findByText(/no networks yet/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Create subnet and scan job' })).toBeDisabled()
  })
})
