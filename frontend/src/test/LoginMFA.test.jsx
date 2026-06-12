import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import LoginPage from '../pages/LoginPage'

const authLogin = vi.fn()

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ login: authLogin }),
}))

vi.mock('../api/client', () => ({
  getAuthProviders: vi.fn(),
  getPublicInfo: vi.fn(),
  login: vi.fn(),
  ldapLogin: vi.fn(),
  verifyMFA: vi.fn(),
  resendVerification: vi.fn(),
}))

import { getAuthProviders, getPublicInfo, login, verifyMFA } from '../api/client'

function renderLogin() {
  return render(
    <MemoryRouter initialEntries={['/login']}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<div>Dashboard Home</div>} />
      </Routes>
    </MemoryRouter>
  )
}

// Walks the password step so each test starts at the MFA prompt.
async function reachMFAPrompt(user) {
  login.mockResolvedValue({ data: { mfaRequired: true, mfaChallenge: 'challenge-token-1' } })
  renderLogin()
  await user.type(screen.getByLabelText('Username'), 'alice')
  await user.type(screen.getByLabelText('Password'), 'hunter2hunter2')
  await user.click(screen.getByRole('button', { name: /sign in/i }))
  await screen.findByText('Two-Factor Authentication')
}

describe('LoginPage MFA step', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getAuthProviders.mockResolvedValue({ data: { ldap: false, oauth2: false, saml: false } })
    getPublicInfo.mockResolvedValue({ data: { registrationEnabled: true } })
  })

  it('shows the MFA prompt after a password login that requires MFA', async () => {
    const user = userEvent.setup()
    await reachMFAPrompt(user)

    expect(login).toHaveBeenCalledWith('alice', 'hunter2hunter2')
    expect(screen.getByLabelText('Authentication Code')).toBeInTheDocument()
    // The user must not be considered logged in yet.
    expect(authLogin).not.toHaveBeenCalled()
  })

  it('completes login with a valid MFA code', async () => {
    const user = userEvent.setup()
    await reachMFAPrompt(user)

    verifyMFA.mockResolvedValue({ data: { user: { id: 7, username: 'alice' } } })
    await user.type(screen.getByLabelText('Authentication Code'), '123456')
    await user.click(screen.getByRole('button', { name: 'Verify' }))

    await waitFor(() => {
      expect(verifyMFA).toHaveBeenCalledWith('challenge-token-1', '123456')
      expect(authLogin).toHaveBeenCalledWith({ id: 7, username: 'alice' })
    })
    expect(await screen.findByText('Dashboard Home')).toBeInTheDocument()
  })

  it('shows the error and stays on the MFA form for an invalid code', async () => {
    const user = userEvent.setup()
    await reachMFAPrompt(user)

    verifyMFA.mockRejectedValue({ response: { status: 401, data: { error: 'invalid MFA code' } } })
    await user.type(screen.getByLabelText('Authentication Code'), '000000')
    await user.click(screen.getByRole('button', { name: 'Verify' }))

    expect(await screen.findByText('invalid MFA code')).toBeInTheDocument()
    expect(screen.getByLabelText('Authentication Code')).toBeInTheDocument()
    expect(authLogin).not.toHaveBeenCalled()
  })

  it('returns to the sign-in form when the MFA challenge expires', async () => {
    const user = userEvent.setup()
    await reachMFAPrompt(user)

    verifyMFA.mockRejectedValue({ response: { status: 401, data: { error: 'MFA challenge expired' } } })
    await user.type(screen.getByLabelText('Authentication Code'), '123456')
    await user.click(screen.getByRole('button', { name: 'Verify' }))

    expect(await screen.findByText('MFA session expired. Please sign in again.')).toBeInTheDocument()
    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.queryByLabelText('Authentication Code')).not.toBeInTheDocument()
  })

  it('allows backing out of the MFA prompt to the sign-in form', async () => {
    const user = userEvent.setup()
    await reachMFAPrompt(user)

    await user.click(screen.getByRole('button', { name: 'Back to sign in' }))
    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.queryByLabelText('Authentication Code')).not.toBeInTheDocument()
  })
})
