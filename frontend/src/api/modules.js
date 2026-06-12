// Network modules: customers, AS, NAT, firewall zones, DHCP, circuits.
import { api } from './client'

export const getCustomers = () => api.get('/customers')

export const getCustomer = (id) => api.get(`/customers/${id}`)

export const createCustomer = (data) => api.post('/customers', data)

export const updateCustomer = (id, data) => api.put(`/customers/${id}`, data)

export const deleteCustomer = (id) => api.delete(`/customers/${id}`)

export const getAutonomousSystems = () => api.get('/autonomous-systems')

export const getAutonomousSystem = (id) => api.get(`/autonomous-systems/${id}`)

export const createAutonomousSystem = (data) => api.post('/autonomous-systems', data)

export const updateAutonomousSystem = (id, data) => api.put(`/autonomous-systems/${id}`, data)

export const deleteAutonomousSystem = (id) => api.delete(`/autonomous-systems/${id}`)

export const getNATRules = () => api.get('/nat-rules')

export const createNATRule = (data) => api.post('/nat-rules', data)

export const updateNATRule = (id, data) => api.put(`/nat-rules/${id}`, data)

export const deleteNATRule = (id) => api.delete(`/nat-rules/${id}`)

export const getFirewallZones = () => api.get('/firewall-zones')

export const createFirewallZone = (data) => api.post('/firewall-zones', data)

export const updateFirewallZone = (id, data) => api.put(`/firewall-zones/${id}`, data)

export const deleteFirewallZone = (id) => api.delete(`/firewall-zones/${id}`)

export const getFirewallZoneMappings = (params = {}) => api.get('/firewall-zone-mappings', { params })

export const createFirewallZoneMapping = (data) => api.post('/firewall-zone-mappings', data)

export const updateFirewallZoneMapping = (id, data) => api.put(`/firewall-zone-mappings/${id}`, data)

export const deleteFirewallZoneMapping = (id) => api.delete(`/firewall-zone-mappings/${id}`)

export const getDHCPServers = () => api.get('/dhcp-servers')

export const createDHCPServer = (data) => api.post('/dhcp-servers', data)

export const updateDHCPServer = (id, data) => api.put(`/dhcp-servers/${id}`, data)

export const deleteDHCPServer = (id) => api.delete(`/dhcp-servers/${id}`)

export const getDHCPLeases = (params = {}) => api.get('/dhcp-leases', { params })

export const createDHCPLease = (data) => api.post('/dhcp-leases', data)

export const updateDHCPLease = (id, data) => api.put(`/dhcp-leases/${id}`, data)

export const deleteDHCPLease = (id) => api.delete(`/dhcp-leases/${id}`)

export const getCircuitProviders = () => api.get('/circuit-providers')

export const createCircuitProvider = (data) => api.post('/circuit-providers', data)

export const updateCircuitProvider = (id, data) => api.put(`/circuit-providers/${id}`, data)

export const deleteCircuitProvider = (id) => api.delete(`/circuit-providers/${id}`)

export const getPhysicalCircuits = () => api.get('/physical-circuits')

export const createPhysicalCircuit = (data) => api.post('/physical-circuits', data)

export const updatePhysicalCircuit = (id, data) => api.put(`/physical-circuits/${id}`, data)

export const deletePhysicalCircuit = (id) => api.delete(`/physical-circuits/${id}`)

export const getLogicalCircuits = () => api.get('/logical-circuits')

export const createLogicalCircuit = (data) => api.post('/logical-circuits', data)

export const updateLogicalCircuit = (id, data) => api.put(`/logical-circuits/${id}`, data)

export const deleteLogicalCircuit = (id) => api.delete(`/logical-circuits/${id}`)

export const getCustomerAssociations = (params = {}) => api.get('/customer-associations', { params })

export const createCustomerAssociation = (data) => api.post('/customer-associations', data)

export const deleteCustomerAssociation = (id) => api.delete(`/customer-associations/${id}`)
