import { api } from './client'

// Subnet requests (user)
export const submitSubnetRequest = (data) => api.post('/requests/subnets', data)
export const getMySubnetRequests = () => api.get('/requests/subnets')
export const cancelSubnetRequest = (id) => api.delete(`/requests/subnets/${id}`)

// IP requests (user)
export const submitIPRequest = (data) => api.post('/requests/ips', data)
export const getMyIPRequests = () => api.get('/requests/ips')
export const cancelIPRequest = (id) => api.delete(`/requests/ips/${id}`)

// Comments
export const getRequestComments = (type, id) => api.get(`/requests/${type}/${id}/comments`)
export const addRequestComment = (type, id, body) => api.post(`/requests/${type}/${id}/comments`, { body })

// Admin
export const adminGetSubnetRequests = (params = {}) => api.get('/admin/requests/subnets', { params })
export const adminGetIPRequests = (params = {}) => api.get('/admin/requests/ips', { params })
export const adminApproveSubnetRequest = (id, reviewerNote) =>
  api.post(`/admin/requests/subnets/${id}/approve`, { reviewer_note: reviewerNote })
export const adminRejectSubnetRequest = (id, reviewerNote) =>
  api.post(`/admin/requests/subnets/${id}/reject`, { reviewer_note: reviewerNote })
export const adminApproveIPRequest = (id, reviewerNote) =>
  api.post(`/admin/requests/ips/${id}/approve`, { reviewer_note: reviewerNote })
export const adminRejectIPRequest = (id, reviewerNote) =>
  api.post(`/admin/requests/ips/${id}/reject`, { reviewer_note: reviewerNote })
export const getPendingRequestCount = () => api.get('/admin/requests/pending-count')
