// Observed state and unregistered-host discovery endpoints.
import { api } from './client'

export const getObservedState = (resourceType, resourceId) =>
  api.get('/admin/discovery/observed', { params: { resource_type: resourceType, resource_id: resourceId } })

export const listUnregisteredHosts = () => api.get('/admin/discovery/unregistered')
