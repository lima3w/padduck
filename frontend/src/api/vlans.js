// VRFs, VLAN domains/groups, and VLANs.
import { api } from './client'

export const getVrfs = () => api.get('/vrfs')

export const getVrf = (id) => api.get(`/vrfs/${id}`)

export const createVrf = (data) => api.post('/vrfs', data)

export const updateVrf = (id, data) => api.put(`/vrfs/${id}`, data)

export const deleteVrf = (id) => api.delete(`/vrfs/${id}`)

export const getVrfVlans = (id) => api.get(`/vrfs/${id}/vlans`)

export const getVlanDomains = () => api.get('/vlan-domains')

export const getVlanDomain = (id) => api.get(`/vlan-domains/${id}`)

export const createVlanDomain = (data) => api.post('/vlan-domains', data)

export const updateVlanDomain = (id, data) => api.put(`/vlan-domains/${id}`, data)

export const deleteVlanDomain = (id) => api.delete(`/vlan-domains/${id}`)

export const getVlanGroups = () => api.get('/vlan-groups')

export const getVlanGroup = (id) => api.get(`/vlan-groups/${id}`)

export const createVlanGroup = (data) => api.post('/vlan-groups', data)

export const updateVlanGroup = (id, data) => api.put(`/vlan-groups/${id}`, data)

export const deleteVlanGroup = (id) => api.delete(`/vlan-groups/${id}`)

export const getVlans = () => api.get('/vlans')

export const getVlan = (id) => api.get(`/vlans/${id}`)

export const createVlan = (data) => api.post('/vlans', data)

export const updateVlan = (id, data) => api.put(`/vlans/${id}`, data)

export const deleteVlan = (id) => api.delete(`/vlans/${id}`)

export const getVlanSubnets = (id) => api.get(`/vlans/${id}/subnets`)

export const assignSubnetToVlan = (id, subnetId) => api.post(`/vlans/${id}/subnets`, { subnet_id: subnetId })

export const removeSubnetFromVlan = (id, subnetId) => api.delete(`/vlans/${id}/subnets/${subnetId}`)

export const getVlanUsageReport = () => api.get('/admin/vlans/usage-report')
