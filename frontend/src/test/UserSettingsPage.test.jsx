import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import UserSettingsPage from '../pages/UserSettingsPage'
import * as client from '../api/client'

vi.mock('../api/client', () => ({
  getCurrentUser: vi.fn(),
  getPrivacyPolicyVersion: vi.fn(),
  acceptPrivacyPolicy: vi.fn(),
}))

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    user: {
      id: 1,
      username: 'alice',
      email: 'alice@example.test',
      role: 'user',
      state: 'active',
      privacyAcceptedVersion: '1.0',
    },
  }),
}))

function installLocalStorage() {
  let store = {}
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: {
      clear: () => { store = {} },
      getItem: (key) => store[key] ?? null,
      removeItem: (key) => { delete store[key] },
      setItem: (key, value) => { store[key] = String(value) },
    },
  })
}

describe('UserSettingsPage', () => {
  beforeEach(() => {
    if (!globalThis.localStorage) installLocalStorage()
    localStorage.clear()
    client.getPrivacyPolicyVersion.mockResolvedValue({ data: { version: '1.1' } })
    client.acceptPrivacyPolicy.mockResolvedValue({})
  })

  it('exposes account settings as accessible tabs', () => {
    render(
      <MemoryRouter initialEntries={['/settings']}>
        <UserSettingsPage />
      </MemoryRouter>
    )

    expect(screen.getByRole('tablist', { name: 'Account settings sections' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Profile', selected: true })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Privacy' })).toBeInTheDocument()
  })

  it('shows and records privacy consent from the privacy tab', async () => {
    render(
      <MemoryRouter initialEntries={['/settings?tab=privacy']}>
        <UserSettingsPage />
      </MemoryRouter>
    )

    expect(await screen.findByText('Current policy version')).toBeInTheDocument()
    expect(screen.getByText('1.1')).toBeInTheDocument()
    expect(screen.getByText('1.0')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Accept current policy' }))

    await waitFor(() => expect(client.acceptPrivacyPolicy).toHaveBeenCalledTimes(1))
    expect(await screen.findByText('Privacy consent recorded.')).toBeInTheDocument()
    expect(screen.getByText('Current')).toBeInTheDocument()
  })
})
