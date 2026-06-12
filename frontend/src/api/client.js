import axios from 'axios'
import { clearCachedUser } from '../utils/storageKeys'

function getCookie(name) {
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : null
}

export const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Non-authenticated endpoints. Responses still need the same key normalization
// as authenticated responses because login/MFA cache user data in localStorage.
export const noAuthApi = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

const MUTATING_METHODS = new Set(['post', 'put', 'patch', 'delete'])

async function ensureCSRFToken({ forceRefresh = false } = {}) {
  let csrfToken = getCookie('csrf-token')
  if (csrfToken && !forceRefresh) return csrfToken

  const response = await axios.get('/api/v1/csrf-token')
  csrfToken = response.data?.csrf_token || getCookie('csrf-token')
  return csrfToken
}

// Keys whose values contain user-defined data and must not have their keys transformed.
const OPAQUE_FIELDS = new Set(['config', 'custom_fields'])

function snakeToCamel(str) {
  return str.replace(/_([a-z0-9])/g, (_, c) => c.toUpperCase())
}

function deepCamelKeys(obj, opaque = false) {
  if (Array.isArray(obj)) return obj.map(item => deepCamelKeys(item, opaque))
  if (obj !== null && typeof obj === 'object') {
    if (opaque) return obj
    return Object.fromEntries(
      Object.entries(obj).map(([k, v]) => [snakeToCamel(k), deepCamelKeys(v, OPAQUE_FIELDS.has(k))])
    )
  }
  return obj
}

function normalizeResponseData(response) {
  if (response.data && typeof response.data === 'object' && !(response.data instanceof Blob)) {
    response.data = deepCamelKeys(response.data)
  }
  return response
}

function isCSRFValidationError(error) {
  return error.response?.status === 403 && error.response?.data?.error === 'csrf validation failed'
}

function isMutatingRequest(config) {
  return MUTATING_METHODS.has((config?.method || 'get').toLowerCase())
}

// Add CSRF token to every mutating request (session cookie is sent automatically by the browser).
api.interceptors.request.use(async (config) => {
  const method = (config.method || 'get').toLowerCase()
  const csrfToken = MUTATING_METHODS.has(method) ? await ensureCSRFToken() : getCookie('csrf-token')
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken
  }
  return config
})

// Normalise response data to camelCase and handle 401s.
api.interceptors.response.use(
  normalizeResponseData,
  async (error) => {
    const originalRequest = error.config || {}
    if (isCSRFValidationError(error) && isMutatingRequest(originalRequest) && !originalRequest._csrfRetried) {
      const csrfToken = await ensureCSRFToken({ forceRefresh: true })
      originalRequest._csrfRetried = true
      originalRequest.headers = {
        ...(originalRequest.headers || {}),
        'X-CSRF-Token': csrfToken,
      }
      return api.request(originalRequest)
    }

    if (error.response?.status === 401) {
      clearCachedUser()
      const publicPaths = ['/login', '/register', '/forgot-password', '/reset-password', '/verify-email', '/auth/']
      const onPublicPage = publicPaths.some(p => window.location.pathname.startsWith(p))
      if (!onPublicPage) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

noAuthApi.interceptors.response.use(normalizeResponseData)
