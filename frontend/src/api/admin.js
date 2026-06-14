// Admin: config, users/roles, audit, webhooks, reports, scan settings.
import { api } from './client'

export const getAdminConfig = () => api.get('/admin/config')

export const revealAdminConfigValue = (key) => api.get('/admin/config/reveal', { params: { key } })

export const updateAdminConfig = (updates) => api.put('/admin/config', updates)

export const testSMTP = (to) => api.post('/admin/config/test-email', { to })

export const checkForUpdates = () => api.get('/admin/updates/check')

export const getNotificationStats = () => api.get('/admin/notification-stats')

export const listPendingApprovals = () => api.get('/admin/approvals')

export const approveUser = (id) => api.post(`/admin/approvals/${id}/approve`)

export const rejectUser = (id, reason) => api.post(`/admin/approvals/${id}/reject`, { reason })

export const getAuditLogs = (params = {}) => api.get('/admin/audit-logs', { params })

export const exportAuditLogs = (params = {}) =>
  api.get('/admin/audit-logs/export', { params, responseType: 'blob' })

export const purgeAuditLogs = () => api.post('/admin/audit-logs/purge')

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

export const getOverlapReport = () => api.get('/admin/subnets/overlap-report')

export const getAdminRoles = () => api.get('/admin/roles')

export const createRole = (data) => api.post('/admin/roles', data)

export const getRole = (id) => api.get(`/admin/roles/${id}`)

export const updateRole = (id, data) => api.put(`/admin/roles/${id}`, data)

export const deleteRole = (id) => api.delete(`/admin/roles/${id}`)

export const addPermissionToRole = (roleId, data) => api.post(`/admin/roles/${roleId}/permissions`, data)

export const removePermissionFromRole = (roleId, permId) => api.delete(`/admin/roles/${roleId}/permissions/${permId}`)

export const listAvailablePermissions = () => api.get('/admin/permissions')

export const listRolePresets = () => api.get('/admin/roles/presets')

export const getRolePresetDiff = (roleId, presetId) => api.get(`/admin/roles/${roleId}/diff`, { params: { preset: presetId } })

export const getLdapConfig = () => api.get('/admin/auth/ldap')

export const updateLdapConfig = (config) => api.put('/admin/auth/ldap', config)

export const testLdapConnection = () => api.post('/admin/auth/ldap/test')

export const getLdapGroupMappings = () => api.get('/admin/auth/ldap/group-mappings')

export const createLdapGroupMapping = (data) => api.post('/admin/auth/ldap/group-mappings', data)

export const deleteLdapGroupMapping = (id) => api.delete(`/admin/auth/ldap/group-mappings/${id}`)

export const getOAuth2Config = () => api.get('/admin/auth/oauth2')

export const updateOAuth2Config = (config) => api.put('/admin/auth/oauth2', config)

export const getSamlConfig = () => api.get('/admin/auth/saml')

export const updateSamlConfig = (config) => api.put('/admin/auth/saml', config)

export const getCustomFields = (entityType) =>
  api.get('/admin/custom-fields', { params: entityType ? { entity_type: entityType } : {} })

export const createCustomField = (data) => api.post('/admin/custom-fields', data)

export const updateCustomField = (id, data) => api.put(`/admin/custom-fields/${id}`, data)

export const deleteCustomField = (id) => api.delete(`/admin/custom-fields/${id}`)

export const reorderCustomFields = (ids) => api.put('/admin/custom-fields/reorder', { ids })

export const getAdminUsers = () => api.get('/admin/users')

export const createUser = (data) => api.post('/users', data)

export const getUserRoles = (userId) => api.get(`/admin/users/${userId}/roles`)

export const assignUserRole = (userId, data) => api.post(`/admin/users/${userId}/roles`, data)

export const removeUserRole = (userId, roleId) => api.delete(`/admin/users/${userId}/roles/${roleId}`)

export const getInactiveIPs = (days = 30, limit = 10) => api.get('/admin/reports/inactive-ips', { params: { days, limit } })

export const getDuplicates = () => api.get('/admin/reports/duplicates')

export const getReconciliationReport = () => api.get('/admin/reports/reconciliation')

export const getScanProfiles = () => api.get('/admin/scan-profiles')

export const createScanProfile = (data) => api.post('/admin/scan-profiles', data)

export const updateScanProfile = (id, data) => api.put(`/admin/scan-profiles/${id}`, data)

export const deleteScanProfile = (id) => api.delete(`/admin/scan-profiles/${id}`)

export const getSubnetScanProfile = (subnetId) => api.get(`/admin/subnets/${subnetId}/scan-profile`)

export const setSubnetScanProfile = (subnetId, profileId) => api.put(`/admin/subnets/${subnetId}/scan-profile`, { profile_id: profileId })

export const getScanRetention = () => api.get('/admin/scan-retention')

export const updateScanRetention = (data) => api.put('/admin/scan-retention', data)

export const runScanRetentionPrune = () => api.post('/admin/scan-retention/prune')

export const getAuditRetention = () => api.get('/admin/audit/retention')

export const updateAuditRetention = (data) => api.put('/admin/audit/retention', data)

export const pruneAuditLogs = () => api.post('/admin/audit/prune')

export const exportAuditLog = (params) => api.get('/admin/audit/export', { params, responseType: 'blob' })

export const listDiscoveryConflicts = (status) => api.get('/admin/discovery/conflicts', { params: status ? { status } : {} })

export const resolveDiscoveryConflict = (id, action) => api.post(`/admin/discovery/conflicts/${id}/resolve`, { action })

export const listTopologyHints = (status) => api.get('/admin/topology/hints', { params: status ? { status } : {} })

export const updateTopologyHintStatus = (id, status) => api.put(`/admin/topology/hints/${id}/status`, { status })

export const listPrivacyVersions = () => api.get('/admin/privacy/versions')

export const createPrivacyVersion = (data) => api.post('/admin/privacy/versions', data)

export const getConsentReport = () => api.get('/admin/privacy/consent-report')

export const getSystemHealth = () => api.get('/admin/system-health')

export const downloadBackup = () =>
  api.get('/admin/backup/download', { responseType: 'blob' })

export const getV2CompatibilityWarnings = () => api.get('/admin/compatibility/v2-warnings')

export const getV2MigrationReadiness = () => api.get('/admin/compatibility/v2-readiness')

export const getV2DeprecationReport = () => api.get('/admin/compatibility/deprecations')

export const getBreakGlassStatus = () => api.get('/admin/break-glass')

export const activateBreakGlass = (justification) => api.post('/admin/break-glass/activate', { justification })

export const endBreakGlass = () => api.post('/admin/break-glass/end')

export const getIdentityPolicies = () => api.get('/admin/identity-policies')

export const updateIdentityPolicies = (data) => api.put('/admin/identity-policies', data)

export const getSessionRisk = () => api.get('/admin/session-risk')

export const getTechnitiumDHCPScopes = () => api.get('/admin/dhcp/technitium/scopes')
export const syncTechnitiumLeases = () => api.post('/admin/dhcp/technitium/sync')
export const importTechnitiumScope = (data) => api.post('/admin/dhcp/technitium/import-scope', data)
