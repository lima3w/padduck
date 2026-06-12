import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from '../pages/LoginPage'

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ login: vi.fn() }),
}))

vi.mock('../api/app', () => ({
  getPublicInfo: vi.fn(),
}))

vi.mock('../api/auth', () => ({
  getAuthProviders: vi.fn(),
}))

import { getAuthProviders } from '../api/auth'
import { getPublicInfo } from '../api/app'

describe('LoginPage', () => {
  beforeEach(() => {
    getAuthProviders.mockResolvedValue({ data: { ldap: false, oauth2: false, saml: false } })
    getPublicInfo.mockResolvedValue({ data: { registrationEnabled: true } })
  })

  it('hides the registration link when self-registration is disabled', async () => {
    getPublicInfo.mockResolvedValue({ data: { registrationEnabled: false } })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(getPublicInfo).toHaveBeenCalled()
    })
    expect(screen.queryByText("Don't have an account?")).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: 'Register' })).not.toBeInTheDocument()
  })

  it('shows the registration link when self-registration is enabled', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    expect(await screen.findByRole('link', { name: 'Register' })).toBeInTheDocument()
  })
})
