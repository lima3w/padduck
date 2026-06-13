// Networks, subnets, IP addresses, tags, and search.
import { api } from './client'

export const getNetworks = () => api.get('/networks')

export const getNetwork = (id) => api.get(`/networks/${id}`)

export const createNetwork = (data) => api.post('/networks', data)

export const updateNetwork = (id, data) => api.put(`/networks/${id}`, data)

export const deleteNetwork = (id) => api.delete(`/networks/${id}`)

export const getSubnets = (networkID) => api.get(`/networks/${networkID}/subnets`)

export const getSubnet = (id) => api.get(`/subnets/${id}`)

export const createSubnet = (networkID, data) => api.post(`/networks/${networkID}/subnets`, data)

export const updateSubnet = (id, data) => api.put(`/subnets/${id}`, data)

export const deleteSubnet = (id) => api.delete(`/subnets/${id}`)

export const getIPAddresses = (subnetID) => api.get(`/subnets/${subnetID}/ip-addresses`)

export const createIPAddress = (subnetID, data) => api.post(`/subnets/${subnetID}/ip-addresses`, data)

export const assignIPAddress = (id, data) => api.post(`/ip-addresses/${id}/assign`, data)

export const assignIPAddressWithLease = (id, data) => api.post(`/ip-addresses/${id}/assign-with-lease`, data)

export const releaseIPAddress = (id) => api.post(`/ip-addresses/${id}/release`)

export const getIPLeaseStatus = (id) => api.get(`/ip-addresses/${id}/lease-status`)

export const releaseExpiredLease = (id) => api.post(`/ip-addresses/${id}/release-expired`)

export const deleteIPAddress = (id) => api.delete(`/ip-addresses/${id}`)

export const getSubnetTree = (networkID) => api.get(`/networks/${networkID}/subnets/tree`)

export const getNetworksPaginated = (page = 1, limit = 25) =>
  api.get('/networks', { params: { page, limit } })

export const getSubnetsPaginated = (networkID, page = 1, limit = 25) =>
  api.get(`/networks/${networkID}/subnets`, { params: { page, limit } })

export const getIPAddressesPaginated = (subnetID, page = 1, limit = 25, sort = '', order = 'asc', hideAvailable = false, fullRange = false) =>
  api.get(`/subnets/${subnetID}/ip-addresses`, {
    params: {
      page,
      limit,
      ...(fullRange ? { full_range: true } : {
        ...(sort && { sort, order }),
        ...(hideAvailable && { hide_available: true }),
      }),
    },
  })

export const searchNetworks = (query, limit = 50, offset = 0) =>
  api.post('/networks/search', { query, limit, offset })

export const searchSubnets = (networkID, body) =>
  api.post(`/subnets/search/${networkID}`, body)

export const searchIPAddresses = (subnetID, query, status = '', limit = 50, offset = 0, filters = {}) =>
  api.post(`/ip-addresses/search/${subnetID}`, { query, status, limit, offset, ...filters })

export const searchIPAddressesGlobal = (q) =>
  api.get('/ip-addresses/search', { params: { q } })

export const quickCreateIPAddress = (address) =>
  api.post('/ip-addresses/quick-create', { address })

export const globalSearch = (q) =>
  api.get('/search', { params: { q } })

export const getTags = () => api.get('/tags')

export const createTag = (data) => api.post('/tags', data)

export const updateTag = (id, data) => api.put(`/tags/${id}`, data)

export const deleteTag = (id) => api.delete(`/tags/${id}`)

export const updateIPMeta = (id, data) => api.put(`/ip-addresses/${id}`, data)

export const bulkReleaseIPs = (ipIds) => api.post('/admin/ip-addresses/bulk-release', { ip_ids: ipIds })

export const bulkDeleteIPs = (ipIds) => api.post('/admin/ip-addresses/bulk-delete', { ip_ids: ipIds })
