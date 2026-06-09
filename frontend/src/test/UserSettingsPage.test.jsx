import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import UserSettingsPage from '../pages/UserSettingsPage'

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    user: {
      id: 1,
      username: 'alice',
      email: 'alice@example.test',
      role: 'user',
      state: 'active',
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
  })

  it('exposes account settings as accessible tabs', () => {
    render(
      <MemoryRouter initialEntries={['/settings']}>
        <UserSettingsPage />
      </MemoryRouter>
    )

    expect(screen.getByRole('tablist', { name: 'Account settings networks' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Profile', selected: true })).toBeInTheDocument()
    // Privacy tab removed — users imply acceptance by using the system
    expect(screen.queryByRole('tab', { name: 'Privacy' })).not.toBeInTheDocument()
    // Remaining tabs should be present
    expect(screen.getByRole('tab', { name: 'Security' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Notifications' })).toBeInTheDocument()
  })
})
