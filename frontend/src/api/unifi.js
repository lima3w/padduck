import { api } from './client'

export const testUniFiConnection = (params) => api.post('/admin/unifi/test', params || {})
export const syncUniFi = () => api.post('/admin/unifi/sync')
export const getUniFiStatus = () => api.get('/admin/unifi/status')
export const getUniFiConflicts = (status = 'open') => api.get('/admin/unifi/conflicts', { params: { status } })
export const resolveUniFiConflict = (id, action) => api.post(`/admin/unifi/conflicts/${id}/resolve`, { action })
export const getUniFiLinks = () => api.get('/unifi/links')
