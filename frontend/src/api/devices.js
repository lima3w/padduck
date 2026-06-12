// Devices, interfaces, and fingerprints.
import { api } from './client'

export const getDeviceSNMPCredentials = (id) => api.get(`/devices/${id}/snmp-credentials`)

export const getDevice = (id) => api.get(`/devices/${id}`)

export const updateDevice = (id, data) => api.put(`/devices/${id}`, data)

export const getDeviceTypes = () => api.get('/device-types')

export const getDeviceIPs = (id) => api.get(`/devices/${id}/ip-addresses`)

export const associateDeviceIP = (deviceId, ipId, data) =>
  api.post(`/devices/${deviceId}/ip-addresses/${ipId}/associate`, data)

export const disassociateDeviceIP = (deviceId, ipId) =>
  api.delete(`/devices/${deviceId}/ip-addresses/${ipId}`)

export const getDeviceInterfaces = (id) => api.get(`/devices/${id}/interfaces`)

export const createDeviceInterface = (id, data) => api.post(`/devices/${id}/interfaces`, data)

export const updateDeviceInterface = (deviceId, ifaceId, data) =>
  api.put(`/devices/${deviceId}/interfaces/${ifaceId}`, data)

export const deleteDeviceInterface = (deviceId, ifaceId) =>
  api.delete(`/devices/${deviceId}/interfaces/${ifaceId}`)

export const getDeviceFingerprint = (deviceId) => api.get(`/admin/devices/${deviceId}/fingerprint`)

export const buildDeviceFingerprint = (deviceId, data) => api.post(`/admin/devices/${deviceId}/fingerprint`, data)
