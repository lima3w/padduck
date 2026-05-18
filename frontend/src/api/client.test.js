import axios from 'axios'
import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('axios', () => {
  const instances = []
  const makeInstance = () => {
    const instance = {
      interceptors: {
        request: {
          use: vi.fn((handler) => {
            instance.requestInterceptor = handler
          }),
        },
        response: {
          use: vi.fn((handler) => {
            instance.responseInterceptor = handler
          }),
        },
      },
    }
    instances.push(instance)
    return instance
  }

  return {
    default: {
      create: vi.fn(() => makeInstance()),
      get: vi.fn(),
      __instances: instances,
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

  it('normalizes unauthenticated auth response user fields before caching callers receive them', async () => {
    await import('./client')

    const noAuthApi = axios.__instances[1]
    const response = noAuthApi.responseInterceptor({
      data: {
        user: {
          id: 1,
          privacy_accepted_version: '1.0',
        },
      },
    })

    expect(response.data.user.privacyAcceptedVersion).toBe('1.0')
    expect(response.data.user.privacy_accepted_version).toBeUndefined()
  })
})
