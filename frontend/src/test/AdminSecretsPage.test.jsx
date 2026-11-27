import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AdminSecretsPage from '../pages/AdminSecretsPage'

vi.mock('../api/admin', () => ({
  getSecrets: vi.fn(),
  createSecret: vi.fn(),
  rotateSecret: vi.fn(),
  deleteSecret: vi.fn(),
  getSecretVersions: vi.fn(),
  deleteSecretVersion: vi.fn(),
}))

import {
  getSecrets, createSecret, rotateSecret, deleteSecret, getSecretVersions, deleteSecretVersion,
} from '../api/admin'

const SECRET = { id: 7, name: 'core-snmp', description: 'core switches', version: 2, exists: true, createdAt: '2026-11-01T10:00:00Z', rotatedAt: '2026-11-20T10:00:00Z' }

describe('AdminSecretsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getSecrets.mockResolvedValue({ data: [SECRET] })
  })

  it('lists secrets with their reference and never shows a value', async () => {
    render(<AdminSecretsPage />)
    expect(await screen.findByText('core-snmp')).toBeInTheDocument()
    expect(screen.getByText('secret://core-snmp')).toBeInTheDocument()
    expect(screen.getByText('v2')).toBeInTheDocument()
    expect(screen.queryByText(/hunter2/)).not.toBeInTheDocument()
  })

  it('creates a secret with a masked value field', async () => {
    const user = userEvent.setup()
    createSecret.mockResolvedValue({ data: {} })
    render(<AdminSecretsPage />)
    await user.click(await screen.findByRole('button', { name: '+ New Secret' }))

    const value = screen.getByLabelText(/^Value/)
    expect(value).toHaveAttribute('type', 'password')
    await user.type(screen.getByLabelText(/^Name/), 'edge-snmp')
    await user.type(value, 'hunter2')
    await user.click(screen.getByRole('button', { name: 'Create' }))

    await waitFor(() => expect(createSecret).toHaveBeenCalledWith({ name: 'edge-snmp', description: '', value: 'hunter2' }))
    expect(await screen.findByText('Secret created')).toBeInTheDocument()
  })

  it('shows the server error when creating fails', async () => {
    const user = userEvent.setup()
    createSecret.mockRejectedValue({ response: { data: { error: 'secret already exists' } } })
    render(<AdminSecretsPage />)
    await user.click(await screen.findByRole('button', { name: '+ New Secret' }))
    await user.type(screen.getByLabelText(/^Name/), 'core-snmp')
    await user.type(screen.getByLabelText(/^Value/), 'x')
    await user.click(screen.getByRole('button', { name: 'Create' }))
    expect(await screen.findByText('secret already exists')).toBeInTheDocument()
  })

  it('rotates a secret', async () => {
    const user = userEvent.setup()
    rotateSecret.mockResolvedValue({ data: {} })
    render(<AdminSecretsPage />)
    await user.click(await screen.findByRole('button', { name: 'Rotate' }))
    await user.type(screen.getByLabelText(/^New value/), 'v3-value')
    const dialog = screen.getByRole('dialog')
    await user.click(within(dialog).getByRole('button', { name: 'Rotate' }))
    await waitFor(() => expect(rotateSecret).toHaveBeenCalledWith(7, 'v3-value'))
    expect(await screen.findByText('Secret rotated')).toBeInTheDocument()
  })

  it('lists versions and only offers delete on old ones', async () => {
    const user = userEvent.setup()
    getSecretVersions.mockResolvedValue({ data: [
      { version: 2, current: true, createdAt: '2026-11-20T10:00:00Z' },
      { version: 1, current: false, createdAt: '2026-11-01T10:00:00Z' },
    ] })
    deleteSecretVersion.mockResolvedValue({})
    render(<AdminSecretsPage />)
    await user.click(await screen.findByRole('button', { name: 'Versions' }))

    expect(await screen.findByText('Current')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Delete v2' })).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Delete v1' }))
    await user.click(screen.getByRole('button', { name: 'Yes' }))
    await waitFor(() => expect(deleteSecretVersion).toHaveBeenCalledWith(7, 1))
  })

  it('deletes a secret after confirmation', async () => {
    const user = userEvent.setup()
    deleteSecret.mockResolvedValue({})
    render(<AdminSecretsPage />)
    await user.click(await screen.findByRole('button', { name: 'Delete' }))
    expect(screen.getByText(/Deletes the secret and every version/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Yes' }))
    await waitFor(() => expect(deleteSecret).toHaveBeenCalledWith(7))
  })
})
