import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import NetworksPage from '../pages/NetworksPage'

vi.mock('../api/client', () => ({
  getNetworksPaginated: vi.fn(),
  createNetwork: vi.fn(),
  updateNetwork: vi.fn(),
  deleteNetwork: vi.fn(),
  searchNetworks: vi.fn(),
}))

vi.mock('../api/requests', () => ({
  submitSubnetRequest: vi.fn(),
}))

vi.mock('../utils/storageKeys', () => ({
  getCachedUser: vi.fn(),
}))

import { getNetworksPaginated, createNetwork, updateNetwork, deleteNetwork } from '../api/client'
import { getCachedUser } from '../utils/storageKeys'

const fixtures = [
  { id: 1, name: 'Lab', description: 'Lab network' },
  { id: 2, name: 'Production', description: 'Prod network' },
]

function renderPage() {
  return render(
    <MemoryRouter>
      <NetworksPage />
    </MemoryRouter>
  )
}

describe('NetworksPage CRUD', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getCachedUser.mockReturnValue({ id: 1, username: 'admin', role: 'admin' })
    getNetworksPaginated.mockResolvedValue({ data: { data: fixtures, total: fixtures.length } })
  })

  it('lists networks from the API', async () => {
    renderPage()
    expect(await screen.findByText('Lab')).toBeInTheDocument()
    expect(screen.getByText('Production')).toBeInTheDocument()
    expect(screen.getByText('2 networks')).toBeInTheDocument()
  })

  it('shows an empty state when there are no networks', async () => {
    getNetworksPaginated.mockResolvedValue({ data: { data: [], total: 0 } })
    renderPage()
    expect(await screen.findByText('No networks yet.')).toBeInTheDocument()
  })

  it('creates a network through the modal form', async () => {
    const user = userEvent.setup()
    createNetwork.mockResolvedValue({ data: { id: 3 } })
    renderPage()
    await screen.findByText('Lab')

    await user.click(screen.getByRole('button', { name: '+ New Network' }))
    await user.type(screen.getByLabelText('Name'), 'Staging')
    await user.type(screen.getByLabelText('Description'), 'Staging network')
    await user.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(createNetwork).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'Staging', description: 'Staging network' })
      )
    })
    // The list reloads after a successful create.
    expect(getNetworksPaginated).toHaveBeenCalledTimes(2)
  })

  it('edits a network with the form prefilled', async () => {
    const user = userEvent.setup()
    updateNetwork.mockResolvedValue({ data: {} })
    renderPage()
    await screen.findByText('Lab')

    await user.click(screen.getAllByRole('button', { name: 'Edit' })[0])
    const nameInput = screen.getByLabelText('Name')
    expect(nameInput).toHaveValue('Lab')

    await user.clear(nameInput)
    await user.type(nameInput, 'Lab Renamed')
    await user.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(updateNetwork).toHaveBeenCalledWith(
        1,
        expect.objectContaining({ name: 'Lab Renamed' })
      )
    })
  })

  it('deletes a network only after confirmation', async () => {
    const user = userEvent.setup()
    deleteNetwork.mockResolvedValue({ data: {} })
    renderPage()
    await screen.findByText('Lab')

    await user.click(screen.getAllByRole('button', { name: 'Delete' })[0])
    expect(deleteNetwork).not.toHaveBeenCalled()
    expect(screen.getByText('Confirm?')).toBeInTheDocument()

    // Backing out does not delete.
    await user.click(screen.getByRole('button', { name: 'No' }))
    expect(deleteNetwork).not.toHaveBeenCalled()

    // Confirming does.
    await user.click(screen.getAllByRole('button', { name: 'Delete' })[0])
    await user.click(screen.getByRole('button', { name: 'Yes' }))
    await waitFor(() => {
      expect(deleteNetwork).toHaveBeenCalledWith(1)
    })
  })

  it('shows an error banner when the API fails', async () => {
    getNetworksPaginated.mockRejectedValue(new Error('boom'))
    renderPage()
    expect(await screen.findByText('Failed to load networks')).toBeInTheDocument()
  })

  it('hides admin actions from non-admin users', async () => {
    getCachedUser.mockReturnValue({ id: 2, username: 'viewer', role: 'user' })
    renderPage()
    await screen.findByText('Lab')

    expect(screen.queryByRole('button', { name: '+ New Network' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Delete' })).not.toBeInTheDocument()
    // Non-admins get the request-subnet path instead.
    expect(screen.getAllByRole('button', { name: 'Request Subnet' }).length).toBeGreaterThan(0)
  })
})
