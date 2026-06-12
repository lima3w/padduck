import { test, expect, request } from '@playwright/test'
import { ADMIN_USER, ADMIN_PASSWORD } from './helpers'

// Mutations authenticated by session cookie must carry a CSRF token.
test('cookie-authenticated mutations require a CSRF token', async ({ baseURL }) => {
  const api = await request.newContext({ baseURL })

  const login = await api.post('/api/v1/auth/login', {
    data: { username: ADMIN_USER, password: ADMIN_PASSWORD },
  })
  expect(login.ok()).toBeTruthy()

  // Session cookie alone: rejected.
  const noToken = await api.post('/api/v1/networks', {
    data: { name: 'csrf-probe', description: '' },
  })
  expect(noToken.status()).toBe(403)

  // With the cookie + header pair: accepted.
  const tokenRes = await api.get('/api/v1/csrf-token')
  const { csrf_token: csrfToken } = await tokenRes.json()
  const withToken = await api.post('/api/v1/networks', {
    headers: { 'X-CSRF-Token': csrfToken },
    data: { name: `csrf-ok-${Date.now()}`, description: 'created by e2e' },
  })
  expect(withToken.ok()).toBeTruthy()

  // Clean up the created network.
  const created = await withToken.json()
  if (created?.id) {
    await api.delete(`/api/v1/networks/${created.id}`, {
      headers: { 'X-CSRF-Token': csrfToken },
    })
  }
  await api.dispose()
})
