import { api } from './client'

export const getTopologyGraph = (params) => api.get('/topology/graph', { params: params || {} })
export const getTopologyPath = (from, to) =>
  api.get('/topology/path', { params: { from_type: from.type, from_id: from.id, to_type: to.type, to_id: to.id } })
export const getTopologyEdges = (status) => api.get('/topology/edges', { params: status ? { status } : {} })
export const setTopologyEdgeStatus = (id, status) => api.put(`/topology/edges/${id}/status`, { status })
export const syncTopologyGraph = () => api.post('/topology/sync')
