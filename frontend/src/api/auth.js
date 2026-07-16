// Authentication, account, sessions, and MFA.
import { api, noAuthApi } from './client'

export const logout = () => api.post('/auth/me/logout')

export const generateToken = (userId, tokenName) =>
  api.post(`/auth/tokens/${userId}`, { token_name: tokenName })

export const generateTokenForMe = (tokenName) =>
  api.post('/auth/me/tokens', { token_name: tokenName })

export const getCurrentUser = () => api.get('/auth/me')

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

export const changePassword = (currentPassword, newPassword) =>
  api.post('/auth/me/change-password', { current_password: currentPassword, new_password: newPassword })

export const verifyMFA = (mfaChallenge, code) =>
  noAuthApi.post('/auth/verify-mfa', { mfa_challenge: mfaChallenge, code })

export const getMFAStatus = () => api.get('/auth/me/mfa')

export const setupTOTP = () => api.post('/auth/me/mfa/setup')

export const confirmTOTP = (code) => api.post('/auth/me/mfa/confirm', { code })

export const disableTOTP = (code) => api.delete('/auth/me/mfa', { data: { code } })

export const regenerateBackupCodes = (code) => api.post('/auth/me/mfa/backup-codes', { code })

export const updateMyAvatar = (source, data) => api.put('/auth/me/avatar', { source, data })

export const updateMyLocale = (locale) => api.put('/auth/me/locale', { locale })

export const getNotificationPreferences = () => api.get('/user/notification-preferences')

export const updateNotificationPreferences = (data) => api.put('/user/notification-preferences', data)

export const listMySessions = () => api.get('/auth/me/sessions')

export const revokeMySession = (sessionId) => api.delete(`/auth/me/sessions/${sessionId}`)

export const logoutAllDevices = () => api.delete('/auth/me/sessions')

export const getLoginHistory = () => api.get('/user/login-history')

export const requestAccountUnlock = (username) => noAuthApi.post('/auth/unlock', { username })

export const verifyAccountUnlock = (token) => noAuthApi.get(`/auth/unlock?token=${encodeURIComponent(token)}`)

export const getPrivacyPolicyVersion = () => noAuthApi.get('/privacy-policy/version')

export const acceptPrivacyPolicy = () => api.post('/auth/me/accept-privacy')

export const getAuthProviders = () => noAuthApi.get('/auth/providers')

export const ldapLogin = (username, password) =>
  noAuthApi.post('/auth/ldap/login', { username, password })
