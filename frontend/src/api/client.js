import axios from 'axios'

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
const noAuthApi = axios.create({
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
      localStorage.removeItem('current_user')
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
export const assignIPAddressWithLease = (id, data) => api.post(`/ip-addresses/${id}/assign-with-lease`, data)
export const releaseIPAddress = (id) => api.post(`/ip-addresses/${id}/release`)
export const getIPLeaseStatus = (id) => api.get(`/ip-addresses/${id}/lease-status`)
export const releaseExpiredLease = (id) => api.post(`/ip-addresses/${id}/release-expired`)
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

export const getFeatures = () => api.get('/features')

export const listUserTokens = (userId) => api.get(`/auth/tokens/${userId}`)

export const listMyTokens = () => api.get('/auth/me/tokens')

export const revokeToken = (tokenId) => api.delete(`/auth/tokens/${tokenId}`)

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

export const requestPasswordReset = (email) =>
  noAuthApi.post('/auth/request-password-reset', { email })

export const resetPassword = (token, newPassword) =>
  noAuthApi.post('/auth/reset-password', { token, new_password: newPassword })

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
export const checkForUpdates = () => api.get('/admin/updates/check')

// Notification preferences
export const getNotificationPreferences = () => api.get('/user/notification-preferences')
export const updateNotificationPreferences = (data) => api.put('/user/notification-preferences', data)
export const getNotificationStats = () => api.get('/admin/notification-stats')

// Sessions
export const listMySessions = () => api.get('/auth/me/sessions')
export const revokeMySession = (sessionId) => api.delete(`/auth/me/sessions/${sessionId}`)
export const logoutAllDevices = () => api.delete('/auth/me/sessions')

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

// Webhooks
export const getWebhookEndpoints = () => api.get('/admin/webhooks')
export const createWebhookEndpoint = (data) => api.post('/admin/webhooks', data)
export const updateWebhookEndpoint = (id, data) => api.put(`/admin/webhooks/${id}`, data)
export const deleteWebhookEndpoint = (id) => api.delete(`/admin/webhooks/${id}`)
export const getWebhookDeliveries = (limit = 50) => api.get('/admin/webhooks/deliveries', { params: { limit } })
export const getWebhookFailureGroups = (limit = 50) => api.get('/admin/webhooks/deliveries/failures', { params: { limit } })
export const getWebhookDelivery = (id) => api.get(`/admin/webhooks/deliveries/${id}`)
export const replayWebhookDelivery = (id) => api.post(`/admin/webhooks/deliveries/${id}/replay`)
export const getApiTokenAnalytics = () => api.get('/admin/api-tokens/analytics')
export const getIntegrationTemplates = () => api.get('/admin/integration-templates')
export const getAutomationPolicies = () => api.get('/admin/automation/policies')
export const createAutomationPolicy = (data) => api.post('/admin/automation/policies', data)
export const updateAutomationPolicy = (id, data) => api.put(`/admin/automation/policies/${id}`, data)
export const deleteAutomationPolicy = (id) => api.delete(`/admin/automation/policies/${id}`)

// Admin user management
export const adminUnlockUser = (id) => api.post(`/admin/users/${id}/unlock`)
export const suspendUser = (id, reason) => api.post(`/admin/users/${id}/suspend`, { reason })
export const unsuspendUser = (id) => api.post(`/admin/users/${id}/unsuspend`)
export const impersonateUser = (id) => api.post(`/admin/users/${id}/impersonate`)
export const sendPasswordResetEmail = (id) => api.post(`/admin/users/${id}/send-password-reset`)
export const updateUserEmail = (id, email) => api.put(`/admin/users/${id}/email`, { email })
export const gdprDeleteUser = (id) => api.post(`/admin/users/${id}/gdpr-delete`)
export const bulkSuspendUsers = (userIds, reason) => api.post('/admin/users/bulk-suspend', { user_ids: userIds, reason })
export const bulkActivateUsers = (userIds) => api.post('/admin/users/bulk-activate', { user_ids: userIds })
export const bulkDeleteUsers = (userIds) => api.post('/admin/users/bulk-delete', { user_ids: userIds })

// Search
export const searchSections = (query, limit = 50, offset = 0) =>
  api.post('/sections/search', { query, limit, offset })

export const searchSubnets = (sectionID, body) =>
  api.post(`/subnets/search/${sectionID}`, body)

export const searchIPAddresses = (subnetID, query, status = '', limit = 50, offset = 0, filters = {}) =>
  api.post(`/ip-addresses/search/${subnetID}`, { query, status, limit, offset, ...filters })

export const globalSearch = (q) =>
  api.get('/search', { params: { q } })

// IP Tags
export const getTags = () => api.get('/tags')
export const createTag = (data) => api.post('/tags', data)
export const updateTag = (id, data) => api.put(`/tags/${id}`, data)
export const deleteTag = (id) => api.delete(`/tags/${id}`)

// IP address meta
export const updateIPMeta = (id, data) => api.put(`/ip-addresses/${id}`, data)

// Overlap report
export const getOverlapReport = () => api.get('/admin/subnets/overlap-report')

// Nameservers
export const getNameservers = () => api.get('/nameservers')
export const getNameserver = (id) => api.get(`/nameservers/${id}`)
export const createNameserver = (data) => api.post('/nameservers', data)
export const updateNameserver = (id, data) => api.put(`/nameservers/${id}`, data)
export const deleteNameserver = (id) => api.delete(`/nameservers/${id}`)

// DNS zones
export const getDnsZones = () => api.get('/dns/zones')
export const getDnsZoneRecords = (zone, type) =>
  api.get(`/dns/zones/${encodeURIComponent(zone)}/records`, { params: type ? { type } : {} })

// DNS admin
export const testDnsConnection = () => api.post('/admin/dns/test')
export const testTechnitiumConnection = (params) => api.post('/admin/dns/technitium/test', params || {})
export const checkAllDns = () => api.post('/admin/dns/check-all')

// VRFs
export const getVrfs = () => api.get('/vrfs')
export const getVrf = (id) => api.get(`/vrfs/${id}`)
export const createVrf = (data) => api.post('/vrfs', data)
export const updateVrf = (id, data) => api.put(`/vrfs/${id}`, data)
export const deleteVrf = (id) => api.delete(`/vrfs/${id}`)
export const getVrfVlans = (id) => api.get(`/vrfs/${id}/vlans`)

// VLAN Domains (#206)
export const getVlanDomains = () => api.get('/vlan-domains')
export const getVlanDomain = (id) => api.get(`/vlan-domains/${id}`)
export const createVlanDomain = (data) => api.post('/vlan-domains', data)
export const updateVlanDomain = (id, data) => api.put(`/vlan-domains/${id}`, data)
export const deleteVlanDomain = (id) => api.delete(`/vlan-domains/${id}`)

// VLAN Groups (#207)
export const getVlanGroups = () => api.get('/vlan-groups')
export const getVlanGroup = (id) => api.get(`/vlan-groups/${id}`)
export const createVlanGroup = (data) => api.post('/vlan-groups', data)
export const updateVlanGroup = (id, data) => api.put(`/vlan-groups/${id}`, data)
export const deleteVlanGroup = (id) => api.delete(`/vlan-groups/${id}`)

// VLANs (#206 #207 #208)
export const getVlans = () => api.get('/vlans')
export const getVlan = (id) => api.get(`/vlans/${id}`)
export const createVlan = (data) => api.post('/vlans', data)
export const updateVlan = (id, data) => api.put(`/vlans/${id}`, data)
export const deleteVlan = (id) => api.delete(`/vlans/${id}`)
export const getVlanSubnets = (id) => api.get(`/vlans/${id}/subnets`)
export const assignSubnetToVlan = (id, subnetId) => api.post(`/vlans/${id}/subnets`, { subnet_id: subnetId })
export const removeSubnetFromVlan = (id, subnetId) => api.delete(`/vlans/${id}/subnets/${subnetId}`)

// VLAN usage report (#209)
export const getVlanUsageReport = () => api.get('/admin/vlans/usage-report')

// Admin roles (for LDAP group mappings)
export const getAdminRoles = () => api.get('/admin/roles')
export const createRole = (data) => api.post('/admin/roles', data)
export const getRole = (id) => api.get(`/admin/roles/${id}`)
export const updateRole = (id, data) => api.put(`/admin/roles/${id}`, data)
export const deleteRole = (id) => api.delete(`/admin/roles/${id}`)
export const addPermissionToRole = (roleId, data) => api.post(`/admin/roles/${roleId}/permissions`, data)
export const removePermissionFromRole = (roleId, permId) => api.delete(`/admin/roles/${roleId}/permissions/${permId}`)
export const listAvailablePermissions = () => api.get('/admin/permissions')

// LDAP (#229 #232)
export const getLdapConfig = () => api.get('/admin/auth/ldap')
export const updateLdapConfig = (config) => api.put('/admin/auth/ldap', config)
export const testLdapConnection = () => api.post('/admin/auth/ldap/test')
export const getLdapGroupMappings = () => api.get('/admin/auth/ldap/group-mappings')
export const createLdapGroupMapping = (data) => api.post('/admin/auth/ldap/group-mappings', data)
export const deleteLdapGroupMapping = (id) => api.delete(`/admin/auth/ldap/group-mappings/${id}`)

// OAuth2 / OIDC (#231)
export const getOAuth2Config = () => api.get('/admin/auth/oauth2')
export const updateOAuth2Config = (config) => api.put('/admin/auth/oauth2', config)

// SAML 2.0 (#230)
export const getSamlConfig = () => api.get('/admin/auth/saml')
export const updateSamlConfig = (config) => api.put('/admin/auth/saml', config)

// Privacy policy
export const getPrivacyPolicyVersion = () => noAuthApi.get('/privacy-policy/version')
export const acceptPrivacyPolicy = () => api.post('/auth/me/accept-privacy')

// Public auth providers info
export const getAuthProviders = () => noAuthApi.get('/auth/providers')

// LDAP login (public)
export const ldapLogin = (username, password) =>
  noAuthApi.post('/auth/ldap/login', { username, password })

// Custom fields
export const getCustomFields = (entityType) =>
  api.get('/admin/custom-fields', { params: entityType ? { entity_type: entityType } : {} })
export const createCustomField = (data) => api.post('/admin/custom-fields', data)
export const updateCustomField = (id, data) => api.put(`/admin/custom-fields/${id}`, data)
export const deleteCustomField = (id) => api.delete(`/admin/custom-fields/${id}`)
export const reorderCustomFields = (ids) => api.put('/admin/custom-fields/reorder', { ids })

// Devices
export const getDeviceSNMPCredentials = (id) => api.get(`/devices/${id}/snmp-credentials`)
export const getDevice = (id) => api.get(`/devices/${id}`)
export const updateDevice = (id, data) => api.put(`/devices/${id}`, data)
export const getDeviceTypes = () => api.get('/device-types')
export const getDeviceIPs = (id) => api.get(`/devices/${id}/ip-addresses`)
export const associateDeviceIP = (deviceId, ipId, data) =>
  api.post(`/devices/${deviceId}/ip-addresses/${ipId}/associate`, data)
export const disassociateDeviceIP = (deviceId, ipId) =>
  api.delete(`/devices/${deviceId}/ip-addresses/${ipId}`)
export const getDeviceInterfaces = (id) => api.get(`/devices/${id}/interfaces`)
export const createDeviceInterface = (id, data) => api.post(`/devices/${id}/interfaces`, data)
export const updateDeviceInterface = (deviceId, ifaceId, data) =>
  api.put(`/devices/${deviceId}/interfaces/${ifaceId}`, data)
export const deleteDeviceInterface = (deviceId, ifaceId) =>
  api.delete(`/devices/${deviceId}/interfaces/${ifaceId}`)

// Device fingerprints (#430)
export const getDeviceFingerprint = (deviceId) => api.get(`/admin/devices/${deviceId}/fingerprint`)
export const buildDeviceFingerprint = (deviceId, data) => api.post(`/admin/devices/${deviceId}/fingerprint`, data)

// Admin users & roles
export const getAdminUsers = () => api.get('/admin/users')
export const createUser = (data) => api.post('/users', data)
export const getUserRoles = (userId) => api.get(`/admin/users/${userId}/roles`)
export const assignUserRole = (userId, data) => api.post(`/admin/users/${userId}/roles`, data)
export const removeUserRole = (userId, roleId) => api.delete(`/admin/users/${userId}/roles/${roleId}`)

export const getCustomers = () => api.get('/customers')
export const getCustomer = (id) => api.get(`/customers/${id}`)
export const createCustomer = (data) => api.post('/customers', data)
export const updateCustomer = (id, data) => api.put(`/customers/${id}`, data)
export const deleteCustomer = (id) => api.delete(`/customers/${id}`)

export const getAutonomousSystems = () => api.get('/autonomous-systems')
export const getAutonomousSystem = (id) => api.get(`/autonomous-systems/${id}`)
export const createAutonomousSystem = (data) => api.post('/autonomous-systems', data)
export const updateAutonomousSystem = (id, data) => api.put(`/autonomous-systems/${id}`, data)
export const deleteAutonomousSystem = (id) => api.delete(`/autonomous-systems/${id}`)

// Bulk IP actions
export const bulkReleaseIPs = (ipIds) => api.post('/admin/ip-addresses/bulk-release', { ip_ids: ipIds })


// Inactive IPs report
export const getInactiveIPs = (days = 30, limit = 10) => api.get('/admin/reports/inactive-ips', { params: { days, limit } })

// Duplicate detection report (#425)
export const getDuplicates = () => api.get('/admin/reports/duplicates')

// Reconciliation center (#424)
export const getReconciliationReport = () => api.get('/admin/reports/reconciliation')

// Scan profiles (#432)
export const getScanProfiles = () => api.get('/admin/scan-profiles')
export const createScanProfile = (data) => api.post('/admin/scan-profiles', data)
export const updateScanProfile = (id, data) => api.put(`/admin/scan-profiles/${id}`, data)
export const deleteScanProfile = (id) => api.delete(`/admin/scan-profiles/${id}`)
export const getSubnetScanProfile = (subnetId) => api.get(`/admin/subnets/${subnetId}/scan-profile`)
export const setSubnetScanProfile = (subnetId, profileId) => api.put(`/admin/subnets/${subnetId}/scan-profile`, { profile_id: profileId })

// Scan retention (#435)
export const getScanRetention = () => api.get('/admin/scan-retention')
export const updateScanRetention = (data) => api.put('/admin/scan-retention', data)
export const runScanRetentionPrune = () => api.post('/admin/scan-retention/prune')

// Discovery conflicts (#431)
export const listDiscoveryConflicts = (status) => api.get('/admin/discovery/conflicts', { params: status ? { status } : {} })
export const resolveDiscoveryConflict = (id, action) => api.post(`/admin/discovery/conflicts/${id}/resolve`, { action })

// Topology hints (#434)
export const listTopologyHints = (status) => api.get('/admin/topology/hints', { params: status ? { status } : {} })
export const updateTopologyHintStatus = (id, status) => api.put(`/admin/topology/hints/${id}/status`, { status })
