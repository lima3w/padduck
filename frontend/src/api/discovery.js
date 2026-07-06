// Observed state and unregistered-host discovery endpoints.
import { api } from './client'

export const getObservedState = (resourceType, resourceId) =>
  api.get('/admin/discovery/observed', { params: { resource_type: resourceType, resource_id: resourceId } })

export const listUnregisteredHosts = () => api.get('/admin/discovery/unregistered')

export const listDriftItems = (status) => api.get('/admin/drift', { params: status ? { status } : {} })

export const getDriftItem = (id) => api.get(`/admin/drift/${id}`)

export const acceptDrift = (id) => api.post(`/admin/drift/${id}/accept`)

export const dismissDrift = (id) => api.post(`/admin/drift/${id}/dismiss`)

export const escalateDrift = (id, note) => api.post(`/admin/drift/${id}/escalate`, note ? { note } : {})
