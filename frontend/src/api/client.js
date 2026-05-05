import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Add token to requests if it exists
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle 401 responses (token expired)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('current_user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Sections
export const getSections = () => api.get('/sections')
export const getSection = (id) => api.get(`/sections/${id}`)
export const createSection = (data) => api.post('/sections', data)
export const updateSection = (id, data) => api.put(`/sections/${id}`, data)
export const deleteSection = (id) => api.delete(`/sections/${id}`)

// Subnets
export const getSubnets = (sectionID) => api.get(`/sections/${sectionID}/subnets`)
export const getSubnet = (id) => api.get(`/subnets/${id}`)
export const createSubnet = (sectionID, data) => api.post(`/sections/${sectionID}/subnets`, data)
export const updateSubnet = (id, data) => api.put(`/subnets/${id}`, data)
export const deleteSubnet = (id) => api.delete(`/subnets/${id}`)

// IP Addresses
export const getIPAddresses = (subnetID) => api.get(`/subnets/${subnetID}/ip-addresses`)
export const createIPAddress = (subnetID, data) => api.post(`/subnets/${subnetID}/ip-addresses`, data)
export const assignIPAddress = (id, data) => api.post(`/ip-addresses/${id}/assign`, data)
export const releaseIPAddress = (id) => api.post(`/ip-addresses/${id}/release`)
export const deleteIPAddress = (id) => api.delete(`/ip-addresses/${id}`)

// Authentication
export const generateToken = (userId, tokenName) =>
  api.post(`/auth/tokens/${userId}`, { token_name: tokenName })

export const generateTokenForMe = (tokenName) =>
  api.post('/auth/me/tokens', { token_name: tokenName })

export const getCurrentUser = () => api.get('/auth/me')

export const listUserTokens = (userId) => api.get(`/auth/tokens/${userId}`)

export const listMyTokens = () => api.get('/auth/me/tokens')

export const revokeToken = (tokenId) => api.delete(`/auth/tokens/${tokenId}`)

// Non-authenticated endpoints (no interceptor needed)
const noAuthApi = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

export const generateTokenAnonymous = (userId, tokenName) =>
  noAuthApi.post(`/auth/tokens/${userId}`, { token_name: tokenName })

// Search
export const searchSections = (query, limit = 50, offset = 0) =>
  api.post('/sections/search', { query, limit, offset })

export const searchSubnets = (sectionID, query, limit = 50, offset = 0) =>
  api.post(`/subnets/search/${sectionID}`, { query, limit, offset })

export const searchIPAddresses = (subnetID, query, status = '', limit = 50, offset = 0) =>
  api.post(`/ip-addresses/search/${subnetID}`, { query, status, limit, offset })
