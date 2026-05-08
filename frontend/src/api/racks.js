import { api } from './client'

export const getRacks = (locationId) =>
  api.get('/racks', locationId ? { params: { location_id: locationId } } : {}).then(r => r.data)
export const getRack = (id) => api.get(`/racks/${id}`).then(r => r.data)
export const createRack = (data) => api.post('/racks', data).then(r => r.data)
export const updateRack = (id, data) => api.put(`/racks/${id}`, data).then(r => r.data)
export const deleteRack = (id) => api.delete(`/racks/${id}`)
export const getRackDevices = (id) => api.get(`/racks/${id}/devices`).then(r => r.data)
