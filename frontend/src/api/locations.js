import { api } from './client'

export const getLocations = (params = {}) => api.get('/locations', { params }).then(r => r.data)
export const getLocationTree = () => api.get('/locations/tree').then(r => r.data)
export const getLocation = (id) => api.get(`/locations/${id}`).then(r => r.data)
export const createLocation = (data) => api.post('/locations', data).then(r => r.data)
export const updateLocation = (id, data) => api.put(`/locations/${id}`, data).then(r => r.data)
export const deleteLocation = (id) => api.delete(`/locations/${id}`)
