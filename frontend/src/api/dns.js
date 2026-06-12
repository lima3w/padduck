// Nameservers, DNS zones, and DNS admin checks.
import { api } from './client'

export const getNameservers = () => api.get('/nameservers')

export const getNameserver = (id) => api.get(`/nameservers/${id}`)

export const createNameserver = (data) => api.post('/nameservers', data)

export const updateNameserver = (id, data) => api.put(`/nameservers/${id}`, data)

export const deleteNameserver = (id) => api.delete(`/nameservers/${id}`)

export const getDnsZones = () => api.get('/dns/zones')

export const getDnsZoneRecords = (zone, type) =>
  api.get(`/dns/zones/${encodeURIComponent(zone)}/records`, { params: type ? { type } : {} })

export const testDnsConnection = () => api.post('/admin/dns/test')

export const testTechnitiumConnection = (params) => api.post('/admin/dns/technitium/test', params || {})

export const checkAllDns = () => api.post('/admin/dns/check-all')
