import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import Modal from '../components/Modal'
import CustomFieldForm from '../components/CustomFieldForm'
import ChangeHistory from '../components/ChangeHistory'
import FingerprintPanel from '../components/FingerprintPanel'
import ObjectRelationshipsPanel from '../components/ObjectRelationshipsPanel'
import { getLocations } from '../api/locations'
import { getRacks } from '../api/racks'
import {
  getDevice, updateDevice, getDeviceTypes,
  getDeviceIPs, associateDeviceIP, disassociateDeviceIP,
  getDeviceInterfaces, createDeviceInterface, updateDeviceInterface, deleteDeviceInterface,
  getCustomFields, getDeviceSNMPCredentials,
  searchIPAddressesGlobal,
} from '../api/client'
import { getCachedUser } from '../utils/storageKeys'

const MEDIA_TYPES = ['copper', 'fiber', 'SFP', 'SFP+', 'QSFP', 'other']

const EMPTY_IFACE_FORM = { name: '', description: '', speed_mbps: '', media_type: '', vlan_id: '' }

export default function DeviceDetailPage() {
  const { id } = useParams()
  const [device, setDevice] = useState(null)
  const [deviceTypes, setDeviceTypes] = useState([])
  const [ipAddresses, setIpAddresses] = useState([])
  const [interfaces, setInterfaces] = useState([])
  const [tab, setTab] = useState('ips')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [modal, setModal] = useState(null)
  const [editForm, setEditForm] = useState({})
  const [ifaceForm, setIfaceForm] = useState(EMPTY_IFACE_FORM)
  const [assocForm, setAssocForm] = useState({ ip_id: '', interface_name: '', is_primary: false })
  const [ipSearch, setIpSearch] = useState('')
  const [ipSearchResults, setIpSearchResults] = useState([])
  const [ipSearching, setIpSearching] = useState(false)
  const [selectedIpLabel, setSelectedIpLabel] = useState('')
  const ipSearchTimer = useRef(null)
  const [deleteIfaceConfirm, setDeleteIfaceConfirm] = useState(null)
  const [deleteIpConfirm, setDeleteIpConfirm] = useState(null)
  const [snmpCreds, setSnmpCreds] = useState(null) // null = not loaded, false = not found
  const [snmpLoading, setSnmpLoading] = useState(false)
  const [snmpError, setSnmpError] = useState(null)
  const [snmpRevealed, setSnmpRevealed] = useState({}) // field key -> bool
  const [saving, setSaving] = useState(false)
  const [cfDefs, setCfDefs] = useState([])
  const [locations, setLocations] = useState([])
  const [racks, setRacks] = useState([])

  useEffect(() => {
    loadAll()
    loadCfDefs()
    loadLocationsList()
  }, [id])

  async function loadLocationsList() {
    try {
      const data = await getLocations()
      setLocations(Array.isArray(data) ? data : (data?.locations ?? []))
    } catch {}
  }

  async function loadRacksForLocation(locationId) {
    if (!locationId) { setRacks([]); return }
    try {
      const data = await getRacks(locationId)
      setRacks(Array.isArray(data) ? data : (data?.racks ?? []))
    } catch { setRacks([]) }
  }

  async function loadCfDefs() {
    try {
      const res = await getCustomFields('device')
      setCfDefs(Array.isArray(res.data) ? res.data : [])
    } catch {}
  }

  async function loadAll() {
    try {
      setLoading(true)
      const [devRes, typesRes] = await Promise.all([
        getDevice(id),
        getDeviceTypes(),
      ])
      setDevice(devRes.data)
      setDeviceTypes(Array.isArray(typesRes.data) ? typesRes.data : [])
      await Promise.all([loadIPs(), loadInterfaces()])
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to load device')
    } finally {
      setLoading(false)
    }
  }

  async function loadIPs() {
    try {
      const res = await getDeviceIPs(id)
      setIpAddresses(Array.isArray(res.data) ? res.data : [])
    } catch {}
  }

  async function loadInterfaces() {
    try {
      const res = await getDeviceInterfaces(id)
      setInterfaces(Array.isArray(res.data) ? res.data : [])
    } catch {}
  }

  function openEdit() {
    const locId = device.locationId ? String(device.locationId) : ''
    if (locId) loadRacksForLocation(locId)
    setEditForm({
      hostname: device.hostname || '',
      type_id: device.typeId ? String(device.typeId) : '',
      description: device.description || '',
      vendor: device.vendor || '',
      model: device.model || '',
      os_version: device.osVersion || '',
      location_id: locId,
      rack_id: device.rackId ? String(device.rackId) : '',
      rack_unit_start: device.rackUnitStart != null ? String(device.rackUnitStart) : '',
      rack_unit_size: device.rackUnitSize != null ? String(device.rackUnitSize) : '',
      custom_fields: device.customFields || {},
      snmp_version: 'v2c',
      snmp_community: '',
      snmp_v3_user: '',
      snmp_v3_auth_proto: 'SHA',
      snmp_v3_auth_pass: '',
      snmp_v3_priv_proto: 'AES',
      snmp_v3_priv_pass: '',
    })
    setModal('edit')
  }

  async function handleEditSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        hostname: editForm.hostname,
        type_id: editForm.type_id ? parseInt(editForm.type_id) : null,
        description: editForm.description || null,
        vendor: editForm.vendor || null,
        model: editForm.model || null,
        os_version: editForm.os_version || null,
        location_id: editForm.location_id ? parseInt(editForm.location_id) : null,
        rack_id: editForm.rack_id ? parseInt(editForm.rack_id) : null,
        rack_unit_start: editForm.rack_unit_start ? parseInt(editForm.rack_unit_start) : null,
        rack_unit_size: editForm.rack_unit_size ? parseInt(editForm.rack_unit_size) : null,
        custom_fields: editForm.custom_fields || {},
        snmp_version: editForm.snmp_version || 'v2c',
        snmp_community: editForm.snmp_community || null,
        snmp_v3_user: editForm.snmp_v3_user || null,
        snmp_v3_auth_proto: editForm.snmp_v3_auth_proto || null,
        snmp_v3_auth_pass: editForm.snmp_v3_auth_pass || null,
        snmp_v3_priv_proto: editForm.snmp_v3_priv_proto || null,
        snmp_v3_priv_pass: editForm.snmp_v3_priv_pass || null,
      }
      const res = await updateDevice(id, body)
      setDevice(res.data)
      setModal(null)
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to update device')
    } finally {
      setSaving(false)
    }
  }

  function openAssociateIP() {
    setAssocForm({ ip_id: '', interface_name: '', is_primary: false })
    setIpSearch('')
    setIpSearchResults([])
    setSelectedIpLabel('')
    setModal('assoc')
  }

  const handleIpSearchChange = useCallback((value) => {
    setIpSearch(value)
    setAssocForm(f => ({ ...f, ip_id: '' }))
    setSelectedIpLabel('')
    if (ipSearchTimer.current) clearTimeout(ipSearchTimer.current)
    if (!value.trim()) { setIpSearchResults([]); return }
    ipSearchTimer.current = setTimeout(async () => {
      setIpSearching(true)
      try {
        const res = await searchIPAddressesGlobal(value.trim())
        setIpSearchResults(Array.isArray(res.data) ? res.data : [])
      } catch { setIpSearchResults([]) }
      finally { setIpSearching(false) }
    }, 300)
  }, [])

  function selectIpResult(ip) {
    setAssocForm(f => ({ ...f, ip_id: ip.id }))
    setSelectedIpLabel(`${ip.address}${ip.hostname ? ` (${ip.hostname})` : ''}`)
    setIpSearch('')
    setIpSearchResults([])
  }

  async function handleAssocSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      await associateDeviceIP(id, assocForm.ip_id, {
        interface_name: assocForm.interface_name || null,
        is_primary: assocForm.is_primary,
      })
      setModal(null)
      await loadIPs()
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to associate IP')
    } finally {
      setSaving(false)
    }
  }

  async function handleDisassociateIP(ipId) {
    try {
      await disassociateDeviceIP(id, ipId)
      setDeleteIpConfirm(null)
      await loadIPs()
    } catch {
      setError('Failed to remove IP association')
    }
  }

  function openAddInterface() {
    setIfaceForm(EMPTY_IFACE_FORM)
    setModal('iface-add')
  }

  function openEditInterface(iface) {
    setIfaceForm({
      name: iface.name || '',
      description: iface.description || '',
      speed_mbps: iface.speedMbps ? String(iface.speedMbps) : '',
      media_type: iface.mediaType || '',
      vlan_id: iface.vlanId ? String(iface.vlanId) : '',
    })
    setModal({ ifaceEdit: iface })
  }

  async function handleIfaceSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        name: ifaceForm.name,
        description: ifaceForm.description || null,
        speed_mbps: ifaceForm.speed_mbps ? parseInt(ifaceForm.speed_mbps) : null,
        media_type: ifaceForm.media_type || null,
        vlan_id: ifaceForm.vlan_id ? parseInt(ifaceForm.vlan_id) : null,
      }
      if (modal === 'iface-add') {
        await createDeviceInterface(id, body)
      } else {
        await updateDeviceInterface(id, modal.ifaceEdit.id, body)
      }
      setModal(null)
      await loadInterfaces()
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to save interface')
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteInterface(ifId) {
    try {
      await deleteDeviceInterface(id, ifId)
      setDeleteIfaceConfirm(null)
      await loadInterfaces()
    } catch {
      setError('Failed to delete interface')
    }
  }

  if (loading) return <p className="text-gray-500">Loading device...</p>
  if (error && !device) return <p className="text-red-600">{error}</p>

  const isAdmin = getCachedUser()?.role === 'admin'
  const typeObj = deviceTypes.find(t => t.id === device?.typeId)
  const locationName = locations.find(l => l.id === device?.locationId)?.name
  const relationshipItems = [
    device?.locationId && {
      label: 'Location',
      value: locationName || `Location #${device.locationId}`,
      to: `/locations/${device.locationId}`,
      description: 'Physical assignment',
    },
    device?.rackId && {
      label: 'Rack',
      value: `Rack #${device.rackId}`,
      to: `/racks/${device.rackId}`,
      description: device.rackUnitStart != null ? `Mounted at U${device.rackUnitStart}-U${device.rackUnitStart + (device.rackUnitSize ?? 1) - 1}` : 'Rack assignment',
    },
    {
      label: 'IP Addresses',
      value: 'Associated IPs',
      count: ipAddresses.length,
      description: `${ipAddresses.length} address${ipAddresses.length === 1 ? '' : 'es'} linked to this device`,
    },
    {
      label: 'Interfaces',
      value: 'Device interfaces',
      count: interfaces.length,
      description: `${interfaces.length} interface${interfaces.length === 1 ? '' : 's'} defined`,
    },
  ]

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/devices" className="hover:text-blue-600">Devices</Link>
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-medium">{device?.hostname}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <span className="text-2xl">{typeObj?.icon || '🖥️'}</span>
          <div>
            <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{device?.hostname}</h1>
            {device?.description && (
              <p className="text-sm text-gray-500 dark:text-gray-400">{device.description}</p>
            )}
          </div>
        </div>
        <button onClick={openEdit} className="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-sm font-medium">
          Edit
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div>
            <dt className="text-gray-500 dark:text-gray-400">Type</dt>
            <dd className="text-gray-800 dark:text-gray-200 font-medium">{typeObj ? `${typeObj.icon} ${typeObj.name}` : '—'}</dd>
          </div>
          <div>
            <dt className="text-gray-500 dark:text-gray-400">Status</dt>
            <dd className="flex items-center gap-1.5">
              <span className={`w-2 h-2 rounded-full ${device?.isOnline ? 'bg-green-500' : 'bg-gray-400'}`}></span>
              <span className={`font-medium ${device?.isOnline ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}`}>
                {device?.isOnline ? 'Online' : 'Offline'}
              </span>
            </dd>
          </div>
          <div>
            <dt className="text-gray-500 dark:text-gray-400">Vendor</dt>
            <dd className="text-gray-800 dark:text-gray-200">{device?.vendor || '—'}</dd>
          </div>
          <div>
            <dt className="text-gray-500 dark:text-gray-400">Model</dt>
            <dd className="text-gray-800 dark:text-gray-200">{device?.model || '—'}</dd>
          </div>
          <div>
            <dt className="text-gray-500 dark:text-gray-400">OS Version</dt>
            <dd className="text-gray-800 dark:text-gray-200">{device?.osVersion || '—'}</dd>
          </div>
          <div>
            <dt className="text-gray-500 dark:text-gray-400">Last Ping</dt>
            <dd className="text-gray-800 dark:text-gray-200">
              {device?.lastPingAt ? new Date(device.lastPingAt).toLocaleString() : '—'}
            </dd>
          </div>
          {device?.locationId && (
            <div>
              <dt className="text-gray-500 dark:text-gray-400">Location</dt>
              <dd className="text-gray-800 dark:text-gray-200">
                <Link to={`/locations/${device.locationId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                  {locations.find(l => l.id === device.locationId)?.name || `#${device.locationId}`}
                </Link>
              </dd>
            </div>
          )}
          {device?.rackId && (
            <div>
              <dt className="text-gray-500 dark:text-gray-400">Rack</dt>
              <dd className="text-gray-800 dark:text-gray-200">
                <Link to={`/racks/${device.rackId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                  Rack #{device.rackId}
                  {device.rackUnitStart != null && ` (U${device.rackUnitStart}–U${device.rackUnitStart + (device.rackUnitSize ?? 1) - 1})`}
                </Link>
              </dd>
            </div>
          )}
        </dl>
        {cfDefs.length > 0 && device?.customFields && Object.keys(device.customFields).length > 0 && (
          <div className="mt-4 border-t dark:border-gray-700 pt-4">
            <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Custom Fields</p>
            <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
              {cfDefs.map(def => {
                const val = device.customFields[def.name]
                if (val == null) return null
                const today = new Date().toISOString().split('T')[0]
                const isPast = def.fieldType === 'date' && val && val < today
                return (
                  <div key={def.id}>
                    <dt className="text-gray-500 dark:text-gray-400">{def.label}</dt>
                    <dd className={`font-medium ${isPast ? 'text-red-600 dark:text-red-400' : 'text-gray-800 dark:text-gray-200'}`}>
                      {def.fieldType === 'url' && val ? (
                        <a href={val} target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:underline break-all">{val}</a>
                      ) : def.fieldType === 'checkbox' ? (
                        val === 'true' ? 'Yes' : 'No'
                      ) : val || '—'}
                    </dd>
                  </div>
                )
              })}
            </dl>
          </div>
        )}
      </div>


      {isAdmin && device && (
        <div className="mb-6">
          <FingerprintPanel deviceId={device.id} deviceIp={device.ipAddress || device.ip || ''} />
        </div>
      )}

      <ObjectRelationshipsPanel relationships={relationshipItems} />

      <div className="flex border-b dark:border-gray-700 mb-4">
        <button
          onClick={() => setTab('ips')}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition ${
            tab === 'ips'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-800 dark:hover:text-gray-200'
          }`}
        >
          IP Addresses
        </button>
        <button
          onClick={() => setTab('interfaces')}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition ${
            tab === 'interfaces'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-800 dark:hover:text-gray-200'
          }`}
        >
          Interfaces
        </button>
        <button
          onClick={() => setTab('credentials')}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition ${
            tab === 'credentials'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-800 dark:hover:text-gray-200'
          }`}
        >
          Credentials
        </button>
      </div>

      {tab === 'ips' && (
        <div>
          <div className="flex justify-end mb-3">
            <button onClick={openAssociateIP} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
              + Associate IP
            </button>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Address</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Interface</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Primary</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Subnet</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {ipAddresses.length === 0 && (
                  <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-400">No IP addresses associated</td></tr>
                )}
                {ipAddresses.map(ip => (
                  <tr key={ip.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{ip.address}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.interfaceName || '—'}</td>
                    <td className="px-4 py-3">
                      {ip.isPrimary && (
                        <span className="inline-block px-2 py-0.5 bg-blue-100 text-blue-700 text-xs font-medium rounded">Primary</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.subnetId ? `#${ip.subnetId}` : '—'}</td>
                    <td className="px-4 py-3 text-right">
                      {deleteIpConfirm === ip.id ? (
                        <span className="space-x-2">
                          <span className="text-red-600 text-xs">Remove?</span>
                          <button onClick={() => handleDisassociateIP(ip.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                          <button onClick={() => setDeleteIpConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                        </span>
                      ) : (
                        <button onClick={() => setDeleteIpConfirm(ip.id)} className="text-gray-400 hover:text-red-600 text-xs">Remove</button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        </div>
      )}

      {tab === 'interfaces' && (
        <div>
          <div className="flex justify-end mb-3">
            <button onClick={openAddInterface} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
              + Add Interface
            </button>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Speed</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Media</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">VLAN</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Connected To</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {interfaces.length === 0 && (
                  <tr><td colSpan={7} className="px-4 py-6 text-center text-gray-400">No interfaces defined</td></tr>
                )}
                {interfaces.map(iface => (
                  <tr key={iface.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{iface.name}</td>
                    <td className="px-4 py-3 text-gray-700 dark:text-gray-200">{iface.description || '—'}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {iface.speedMbps ? `${iface.speedMbps} Mbps` : '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{iface.mediaType || '—'}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{iface.vlanId || '—'}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                      {iface.connectedToDeviceId ? (
                        <Link to={`/devices/${iface.connectedToDeviceId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                          Device #{iface.connectedToDeviceId}
                        </Link>
                      ) : '—'}
                    </td>
                    <td className="px-4 py-3 text-right space-x-2">
                      <button onClick={() => openEditInterface(iface)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                      {deleteIfaceConfirm === iface.id ? (
                        <>
                          <span className="text-red-600 text-xs">Confirm?</span>
                          <button onClick={() => handleDeleteInterface(iface.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                          <button onClick={() => setDeleteIfaceConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                        </>
                      ) : (
                        <button onClick={() => setDeleteIfaceConfirm(iface.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        </div>
      )}

      {tab === 'credentials' && (
        <div className="max-w-lg space-y-4">
          <div>
            <h2 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">SNMP Credentials</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Stored credentials are revealed on demand. Each field must be individually shown.
            </p>
          </div>

          {snmpError && <p className="text-sm text-red-600">{snmpError}</p>}

          {snmpCreds === null && !snmpLoading && (
            <button
              onClick={async () => {
                setSnmpLoading(true)
                setSnmpError(null)
                try {
                  const res = await getDeviceSNMPCredentials(id)
                  setSnmpCreds(res.data)
                  setSnmpRevealed({})
                } catch (err) {
                  const status = err.response?.status
                  if (status === 403) setSnmpError('You do not have permission to view SNMP credentials.')
                  else if (status === 404) setSnmpCreds(false)
                  else setSnmpError('Failed to load credentials.')
                } finally {
                  setSnmpLoading(false)
                }
              }}
              className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition"
            >
              Load Credentials
            </button>
          )}

          {snmpLoading && <p className="text-sm text-gray-500">Loading…</p>}

          {snmpCreds === false && (
            <p className="text-sm text-gray-500">No SNMP credentials stored for this device.</p>
          )}

          {snmpCreds && snmpCreds !== false && (() => {
            const rows = [
              { key: 'snmpVersion', label: 'SNMP Version', value: snmpCreds.snmpVersion, sensitive: false },
              { key: 'snmpCommunity', label: 'Community String', value: snmpCreds.snmpCommunity, sensitive: true },
              { key: 'snmpV3User', label: 'SNMPv3 Username', value: snmpCreds.snmpV3User, sensitive: false },
              { key: 'snmpV3AuthProto', label: 'Auth Protocol', value: snmpCreds.snmpV3AuthProto, sensitive: false },
              { key: 'snmpV3AuthPass', label: 'Auth Password', value: snmpCreds.snmpV3AuthPass, sensitive: true },
              { key: 'snmpV3PrivProto', label: 'Priv Protocol', value: snmpCreds.snmpV3PrivProto, sensitive: false },
              { key: 'snmpV3PrivPass', label: 'Priv Password', value: snmpCreds.snmpV3PrivPass, sensitive: true },
            ].filter((r) => r.value != null && r.value !== '')

            return (
              <div className="border border-gray-200 dark:border-gray-700 rounded divide-y divide-gray-100 dark:divide-gray-700">
                {rows.map(({ key, label, value, sensitive }) => (
                  <div key={key} className="flex items-center justify-between px-4 py-3 gap-4">
                    <span className="text-sm text-gray-600 dark:text-gray-400 w-36 flex-shrink-0">{label}</span>
                    <div className="flex items-center gap-2 flex-1 min-w-0">
                      {!sensitive || snmpRevealed[key] ? (
                        <span className="text-sm font-mono text-gray-900 dark:text-gray-100 break-all">{value}</span>
                      ) : (
                        <span className="text-sm text-gray-400 font-mono">••••••••</span>
                      )}
                      {sensitive && (
                        <button
                          onClick={() => setSnmpRevealed((p) => ({ ...p, [key]: !p[key] }))}
                          className="flex-shrink-0 text-xs text-blue-600 hover:underline"
                        >
                          {snmpRevealed[key] ? 'Hide' : 'Reveal'}
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )
          })()}

          {snmpCreds && (
            <button
              onClick={() => { setSnmpCreds(null); setSnmpRevealed({}) }}
              className="text-xs text-gray-400 hover:text-gray-600 hover:underline"
            >
              Clear credentials from view
            </button>
          )}
        </div>
      )}

      {modal === 'edit' && (
        <Modal title="Edit Device" onClose={() => setModal(null)}>
          <form onSubmit={handleEditSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Hostname <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={editForm.hostname}
                onChange={e => setEditForm(f => ({ ...f, hostname: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={editForm.type_id}
                onChange={e => setEditForm(f => ({ ...f, type_id: e.target.value }))}
              >
                <option value="">No type</option>
                {deviceTypes.map(t => (
                  <option key={t.id} value={t.id}>{t.icon} {t.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={2}
                value={editForm.description}
                onChange={e => setEditForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Vendor</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={editForm.vendor}
                  onChange={e => setEditForm(f => ({ ...f, vendor: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Model</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={editForm.model}
                  onChange={e => setEditForm(f => ({ ...f, model: e.target.value }))}
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">OS Version</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={editForm.os_version}
                onChange={e => setEditForm(f => ({ ...f, os_version: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Location (optional)</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={editForm.location_id || ''}
                onChange={e => {
                  const locId = e.target.value
                  setEditForm(f => ({ ...f, location_id: locId, rack_id: '', rack_unit_start: '', rack_unit_size: '' }))
                  loadRacksForLocation(locId)
                }}
              >
                <option value="">No location</option>
                {locations.map(l => (
                  <option key={l.id} value={l.id}>{l.name}</option>
                ))}
              </select>
            </div>
            {editForm.location_id && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Rack (optional)</label>
                  <select
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    value={editForm.rack_id || ''}
                    onChange={e => setEditForm(f => ({ ...f, rack_id: e.target.value }))}
                  >
                    <option value="">No rack</option>
                    {racks.map(r => (
                      <option key={r.id} value={r.id}>{r.name}</option>
                    ))}
                  </select>
                </div>
                {editForm.rack_id && (
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Rack Unit Start</label>
                      <input
                        type="number"
                        min="1"
                        className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                        placeholder="1"
                        value={editForm.rack_unit_start || ''}
                        onChange={e => setEditForm(f => ({ ...f, rack_unit_start: e.target.value }))}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Rack Unit Size</label>
                      <input
                        type="number"
                        min="1"
                        className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                        placeholder="1"
                        value={editForm.rack_unit_size || ''}
                        onChange={e => setEditForm(f => ({ ...f, rack_unit_size: e.target.value }))}
                      />
                    </div>
                  </div>
                )}
              </>
            )}
            <div className="border-t dark:border-gray-600 pt-4">
              <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">SNMP</p>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">SNMP Version</label>
                  <select
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    value={editForm.snmp_version}
                    onChange={e => setEditForm(f => ({ ...f, snmp_version: e.target.value }))}
                  >
                    <option value="v1">v1</option>
                    <option value="v2c">v2c</option>
                    <option value="v3">v3</option>
                  </select>
                </div>
                {(editForm.snmp_version === 'v1' || editForm.snmp_version === 'v2c') && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Community String</label>
                    <input
                      type="text"
                      className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                      placeholder="Leave blank to keep existing"
                      value={editForm.snmp_community}
                      onChange={e => setEditForm(f => ({ ...f, snmp_community: e.target.value }))}
                    />
                  </div>
                )}
                {editForm.snmp_version === 'v3' && (
                  <>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username</label>
                      <input
                        type="text"
                        className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                        value={editForm.snmp_v3_user}
                        onChange={e => setEditForm(f => ({ ...f, snmp_v3_user: e.target.value }))}
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Auth Protocol</label>
                        <select
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          value={editForm.snmp_v3_auth_proto}
                          onChange={e => setEditForm(f => ({ ...f, snmp_v3_auth_proto: e.target.value }))}
                        >
                          <option value="SHA">SHA</option>
                          <option value="MD5">MD5</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Auth Password</label>
                        <input
                          type="password"
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          placeholder="Leave blank to keep existing"
                          value={editForm.snmp_v3_auth_pass}
                          onChange={e => setEditForm(f => ({ ...f, snmp_v3_auth_pass: e.target.value }))}
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Priv Protocol</label>
                        <select
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          value={editForm.snmp_v3_priv_proto}
                          onChange={e => setEditForm(f => ({ ...f, snmp_v3_priv_proto: e.target.value }))}
                        >
                          <option value="AES">AES</option>
                          <option value="DES">DES</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Priv Password</label>
                        <input
                          type="password"
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                          placeholder="Leave blank to keep existing"
                          value={editForm.snmp_v3_priv_pass}
                          onChange={e => setEditForm(f => ({ ...f, snmp_v3_priv_pass: e.target.value }))}
                        />
                      </div>
                    </div>
                  </>
                )}
              </div>
            </div>
            {cfDefs.length > 0 && (
              <div className="border-t dark:border-gray-600 pt-4">
                <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Custom Fields</p>
                <CustomFieldForm
                  definitions={cfDefs}
                  values={editForm.custom_fields}
                  onChange={(name, value) => setEditForm(f => ({ ...f, custom_fields: { ...f.custom_fields, [name]: value } }))}
                />
              </div>
            )}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal === 'assoc' && (
        <Modal title="Associate IP Address" onClose={() => setModal(null)}>
          <form onSubmit={handleAssocSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                IP Address <span className="text-red-500">*</span>
              </label>
              {assocForm.ip_id ? (
                <div className="flex items-center gap-2">
                  <span className="flex-1 px-3 py-2 border rounded text-sm font-mono bg-gray-50 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 border-gray-300">
                    {selectedIpLabel}
                  </span>
                  <button
                    type="button"
                    onClick={() => { setAssocForm(f => ({ ...f, ip_id: '' })); setSelectedIpLabel('') }}
                    className="px-2 py-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-sm"
                    title="Clear selection"
                  >✕</button>
                </div>
              ) : (
                <div>
                  <input
                    className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder="Type an IP address or hostname to search…"
                    value={ipSearch}
                    onChange={e => handleIpSearchChange(e.target.value)}
                    autoFocus
                  />
                  {(ipSearching || ipSearchResults.length > 0 || (ipSearch.trim() && !ipSearching)) && (
                    <div className="mt-1 border border-gray-200 dark:border-gray-700 rounded bg-white dark:bg-gray-800 max-h-48 overflow-y-auto">
                      {ipSearching && (
                        <div className="px-3 py-2 text-sm text-gray-400">Searching…</div>
                      )}
                      {!ipSearching && ipSearchResults.length === 0 && ipSearch.trim() && (
                        <div className="px-3 py-2 text-sm text-gray-400">No matching IPs found</div>
                      )}
                      {ipSearchResults.map(ip => (
                        <button
                          key={ip.id}
                          type="button"
                          onClick={() => selectIpResult(ip)}
                          className="w-full text-left px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 font-mono text-gray-900 dark:text-gray-100"
                        >
                          <span className="font-medium">{ip.address}</span>
                          {ip.hostname && <span className="ml-2 text-gray-400 text-xs">{ip.hostname}</span>}
                          {ip.status && <span className="ml-2 text-xs px-1 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">{ip.status}</span>}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Interface Name</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. eth0"
                value={assocForm.interface_name}
                onChange={e => setAssocForm(f => ({ ...f, interface_name: e.target.value }))}
              />
            </div>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={assocForm.is_primary}
                onChange={e => setAssocForm(f => ({ ...f, is_primary: e.target.checked }))}
                className="w-4 h-4 text-blue-600 rounded"
              />
              <span className="text-sm text-gray-700 dark:text-gray-300">Primary address</span>
            </label>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving || !assocForm.ip_id} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Associating...' : 'Associate'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {isAdmin && <ChangeHistory resourceType="device" resourceId={device?.id} />}

      {(modal === 'iface-add' || modal?.ifaceEdit) && (
        <Modal title={modal === 'iface-add' ? 'Add Interface' : 'Edit Interface'} onClose={() => setModal(null)}>
          <form onSubmit={handleIfaceSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="eth0"
                value={ifaceForm.name}
                onChange={e => setIfaceForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="Uplink to core switch"
                value={ifaceForm.description}
                onChange={e => setIfaceForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Speed (Mbps)</label>
                <input
                  type="number"
                  min="0"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="1000"
                  value={ifaceForm.speed_mbps}
                  onChange={e => setIfaceForm(f => ({ ...f, speed_mbps: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Media Type</label>
                <select
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={ifaceForm.media_type}
                  onChange={e => setIfaceForm(f => ({ ...f, media_type: e.target.value }))}
                >
                  <option value="">Select...</option>
                  {MEDIA_TYPES.map(m => <option key={m} value={m}>{m}</option>)}
                </select>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">VLAN ID</label>
              <input
                type="number"
                min="1"
                max="4094"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="100"
                value={ifaceForm.vlan_id}
                onChange={e => setIfaceForm(f => ({ ...f, vlan_id: e.target.value }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
