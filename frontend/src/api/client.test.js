import axios from 'axios'
import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('axios', () => {
  const instance = {
    interceptors: {
      request: {
        use: vi.fn((handler) => {
          instance.requestInterceptor = handler
        }),
      },
      response: { use: vi.fn() },
    },
  }

  return {
    default: {
      create: vi.fn(() => instance),
      get: vi.fn(),
    },
  }
})

describe('api client CSRF handling', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.clearAllMocks()
    document.cookie = 'csrf-token=; Max-Age=0; path=/'
  })

  it('fetches a csrf token before mutating authenticated requests when the cookie is missing', async () => {
    axios.get.mockResolvedValue({ data: { csrf_token: 'fresh-token' } })

    const client = await import('./client')
    const config = await client.api.requestInterceptor({ method: 'post', headers: {} })

    expect(axios.get).toHaveBeenCalledWith('/api/v1/csrf-token')
    expect(config.headers['X-CSRF-Token']).toBe('fresh-token')
  })

  it('uses the existing csrf cookie without fetching a new token', async () => {
    document.cookie = 'csrf-token=existing-token; path=/'

    const client = await import('./client')
    const config = await client.api.requestInterceptor({ method: 'delete', headers: {} })

    expect(axios.get).not.toHaveBeenCalled()
    expect(config.headers['X-CSRF-Token']).toBe('existing-token')
  })

  it('does not fetch csrf tokens for read requests', async () => {
    const client = await import('./client')
    const config = await client.api.requestInterceptor({ method: 'get', headers: {} })

    expect(axios.get).not.toHaveBeenCalled()
    expect(config.headers['X-CSRF-Token']).toBeUndefined()
  })
})
