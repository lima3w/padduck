import axios from 'axios'

function getCookie(name) {
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : null
}

export const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Keys whose values contain user-defined data and must not have their keys transformed.
const OPAQUE_FIELDS = new Set(['custom_fields'])

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

// Add auth token and CSRF token to every request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  const csrfToken = getCookie('csrf-token')
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken
  }
  return config
})

// Normalise response data to camelCase and handle 401s.
api.interceptors.response.use(
  (response) => {
    if (response.data && typeof response.data === 'object' && !(response.data instanceof Blob)) {
      response.data = deepCamelKeys(response.data)
    }
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('current_user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Sections
export const getSections = () => api.get('/sections')
export const getSection = (id) => api.get(`/sections/${id}`)
export const createSection = (data) => api.post('/sections', data)
export const updateSection = (id, data) => api.put(`/sections/${id}`, data)
export const deleteSection = (id) => api.delete(`/sections/${id}`)

// Subnets
export const getSubnets = (sectionID) => api.get(`/sections/${sectionID}/subnets`)
export const getSubnet = (id) => api.get(`/subnets/${id}`)
export const createSubnet = (sectionID, data) => api.post(`/sections/${sectionID}/subnets`, data)
export const updateSubnet = (id, data) => api.put(`/subnets/${id}`, data)
export const deleteSubnet = (id) => api.delete(`/subnets/${id}`)

// IP Addresses
export const getIPAddresses = (subnetID) => api.get(`/subnets/${subnetID}/ip-addresses`)
export const createIPAddress = (subnetID, data) => api.post(`/subnets/${subnetID}/ip-addresses`, data)
export const assignIPAddress = (id, data) => api.post(`/ip-addresses/${id}/assign`, data)
export const releaseIPAddress = (id) => api.post(`/ip-addresses/${id}/release`)
export const deleteIPAddress = (id) => api.delete(`/ip-addresses/${id}`)

// Logout (POST /auth/me/logout)
export const logout = () => api.post('/auth/me/logout')

// Dashboard
export const getDashboardSummary = () => api.get('/dashboard/summary')
export const getDashboardRecentActivity = () => api.get('/dashboard/recent-activity')

// Subnet tree
export const getSubnetTree = (sectionID) => api.get(`/sections/${sectionID}/subnets/tree`)

// Paginated lists
export const getSectionsPaginated = (page = 1, limit = 25) =>
  api.get('/sections', { params: { page, limit } })
export const getSubnetsPaginated = (sectionID, page = 1, limit = 25) =>
  api.get(`/sections/${sectionID}/subnets`, { params: { page, limit } })
export const getIPAddressesPaginated = (subnetID, page = 1, limit = 25) =>
  api.get(`/subnets/${subnetID}/ip-addresses`, { params: { page, limit } })

// Authentication
export const generateToken = (userId, tokenName) =>
  api.post(`/auth/tokens/${userId}`, { token_name: tokenName })

export const generateTokenForMe = (tokenName) =>
  api.post('/auth/me/tokens', { token_name: tokenName })

export const getCurrentUser = () => api.get('/auth/me')

export const listUserTokens = (userId) => api.get(`/auth/tokens/${userId}`)

export const listMyTokens = () => api.get('/auth/me/tokens')

export const revokeToken = (tokenId) => api.delete(`/auth/tokens/${tokenId}`)

// Non-authenticated endpoints (no interceptor needed)
const noAuthApi = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

export const generateTokenAnonymous = (userId, tokenName) =>
  noAuthApi.post(`/auth/tokens/${userId}`, { token_name: tokenName })

export const login = (username, password) =>
  noAuthApi.post('/auth/login', { username, password })

export const register = (username, email, password) =>
  noAuthApi.post('/auth/register', { username, email, password })

export const verifyEmail = (token) =>
  noAuthApi.get(`/auth/verify-email?token=${encodeURIComponent(token)}`)

export const resendVerification = (email) =>
  noAuthApi.post('/auth/resend-verification', { email })

export const verifyMFA = (mfaChallenge, code) =>
  noAuthApi.post('/auth/verify-mfa', { mfa_challenge: mfaChallenge, code })

// MFA (authenticated)
export const getMFAStatus = () => api.get('/auth/me/mfa')
export const setupTOTP = () => api.post('/auth/me/mfa/setup')
export const confirmTOTP = (code) => api.post('/auth/me/mfa/confirm', { code })
export const disableTOTP = (code) => api.delete('/auth/me/mfa', { data: { code } })
export const regenerateBackupCodes = (code) => api.post('/auth/me/mfa/backup-codes', { code })

// Admin config
export const getAdminConfig = () => api.get('/admin/config')
export const updateAdminConfig = (updates) => api.put('/admin/config', updates)
export const testSMTP = (to) => api.post('/admin/config/test-email', { to })

// Security / login history
export const getLoginHistory = () => api.get('/user/login-history')
export const requestAccountUnlock = (username) => noAuthApi.post('/auth/unlock', { username })
export const verifyAccountUnlock = (token) => noAuthApi.get(`/auth/unlock?token=${encodeURIComponent(token)}`)

// Admin approvals
export const listPendingApprovals = () => api.get('/admin/approvals')
export const approveUser = (id) => api.post(`/admin/approvals/${id}/approve`)
export const rejectUser = (id, reason) => api.post(`/admin/approvals/${id}/reject`, { reason })

// Audit logs
export const getAuditLogs = (params = {}) => api.get('/admin/audit-logs', { params })
export const exportAuditLogs = (params = {}) =>
  api.get('/admin/audit-logs/export', { params, responseType: 'blob' })
export const purgeAuditLogs = () => api.post('/admin/audit-logs/purge')

// Admin user management
export const adminUnlockUser = (id) => api.post(`/admin/users/${id}/unlock`)

// Search
export const searchSections = (query, limit = 50, offset = 0) =>
  api.post('/sections/search', { query, limit, offset })

export const searchSubnets = (sectionID, body) =>
  api.post(`/subnets/search/${sectionID}`, body)

export const searchIPAddresses = (subnetID, query, status = '', limit = 50, offset = 0, filters = {}) =>
  api.post(`/ip-addresses/search/${subnetID}`, { query, status, limit, offset, ...filters })

// IP Tags
export const getTags = () => api.get('/tags')
export const createTag = (data) => api.post('/tags', data)
export const updateTag = (id, data) => api.put(`/tags/${id}`, data)
export const deleteTag = (id) => api.delete(`/tags/${id}`)

// IP address meta
export const updateIPMeta = (id, data) => api.put(`/ip-addresses/${id}`, data)

// Overlap report
export const getOverlapReport = () => api.get('/admin/subnets/overlap-report')
