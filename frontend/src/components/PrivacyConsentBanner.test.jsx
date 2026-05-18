import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PrivacyConsentBanner from './PrivacyConsentBanner'
import { acceptPrivacyPolicy, getCurrentUser, getPrivacyPolicyVersion } from '../api/client'

vi.mock('../api/client', () => ({
  acceptPrivacyPolicy: vi.fn(),
  getCurrentUser: vi.fn(),
  getPrivacyPolicyVersion: vi.fn(),
}))

function installLocalStorage() {
  const store = new Map()
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: {
      clear: vi.fn(() => store.clear()),
      getItem: vi.fn((key) => store.get(key) ?? null),
      removeItem: vi.fn((key) => store.delete(key)),
      setItem: vi.fn((key, value) => store.set(key, String(value))),
    },
  })
}

describe('PrivacyConsentBanner', () => {
  beforeEach(() => {
    installLocalStorage()
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('records consent and dismisses even if the follow-up user refresh fails', async () => {
    localStorage.setItem('current_user', JSON.stringify({ id: 1, username: 'admin' }))
    getPrivacyPolicyVersion.mockResolvedValue({ data: { version: '1.0' } })
    acceptPrivacyPolicy.mockResolvedValue({ data: { message: 'privacy policy accepted' } })
    getCurrentUser.mockRejectedValue(new Error('temporary refresh failure'))

    render(<PrivacyConsentBanner />)

    const button = await screen.findByRole('button', { name: /accept privacy policy/i })
    fireEvent.click(button)

    await waitFor(() => expect(acceptPrivacyPolicy).toHaveBeenCalledTimes(1))
    await waitFor(() => expect(screen.queryByText('Privacy Policy')).not.toBeInTheDocument())

    const cached = JSON.parse(localStorage.getItem('current_user'))
    expect(cached.privacyAcceptedVersion).toBe('1.0')
    expect(screen.queryByText(/failed to record consent/i)).not.toBeInTheDocument()
  })
})
