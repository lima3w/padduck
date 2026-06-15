import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import ChangeHistory from '../components/ChangeHistory'
import FingerprintPanel from '../components/FingerprintPanel'
import ObjectRelationshipsPanel from '../components/ObjectRelationshipsPanel'
import { getLocations } from '../api/locations'
import { getRacks } from '../api/racks'
import { searchIPAddressesGlobal, quickCreateIPAddress } from '../api/ipam'
import { getVlans } from '../api/vlans'
import { getDevice, updateDevice, getDeviceTypes, getDeviceIPs, associateDeviceIP, disassociateDeviceIP, getDeviceInterfaces, createDeviceInterface, updateDeviceInterface, deleteDeviceInterface, getDeviceSNMPCredentials } from '../api/devices'
import { getCustomFields } from '../api/admin'
import { getCachedUser } from '../utils/storageKeys'
import DeviceInfoPanel from './device/DeviceInfoPanel'
import IPAddressesTab from './device/IPAddressesTab'
import InterfacesTab from './device/InterfacesTab'
import CredentialsTab from './device/CredentialsTab'
import EditDeviceModal from './device/EditDeviceModal'
import AssociateIPModal from './device/AssociateIPModal'
import InterfaceModal from './device/InterfaceModal'

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
  const [ipCreating, setIpCreating] = useState(false)
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
  const [vlanList, setVlanList] = useState([])

  useEffect(() => {
    loadAll()
    loadCfDefs()
    loadLocationsList()
    loadVlanList()
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  async function loadVlanList() {
    try {
      const res = await getVlans()
      setVlanList(Array.isArray(res.data) ? res.data : [])
    } catch {}
  }

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

  async function handleQuickCreate(address) {
    setIpCreating(true)
    try {
      const res = await quickCreateIPAddress(address)
      selectIpResult(res.data)
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create IP address')
    } finally {
      setIpCreating(false)
    }
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

  async function handleCreateInterfaceFromAssoc(name) {
    const res = await createDeviceInterface(id, { name })
    setInterfaces(prev => [...prev, res.data].sort((a, b) => a.name.localeCompare(b.name)))
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

      <DeviceInfoPanel device={device} typeObj={typeObj} locations={locations} cfDefs={cfDefs} />

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
        <IPAddressesTab
          ipAddresses={ipAddresses}
          deleteIpConfirm={deleteIpConfirm}
          setDeleteIpConfirm={setDeleteIpConfirm}
          onAssociateIP={openAssociateIP}
          onDisassociateIP={handleDisassociateIP}
        />
      )}

      {tab === 'interfaces' && (
        <InterfacesTab
          interfaces={interfaces}
          vlanList={vlanList}
          deleteIfaceConfirm={deleteIfaceConfirm}
          setDeleteIfaceConfirm={setDeleteIfaceConfirm}
          onAddInterface={openAddInterface}
          onEditInterface={openEditInterface}
          onDeleteInterface={handleDeleteInterface}
        />
      )}

      {tab === 'credentials' && (
        <CredentialsTab
          snmpCreds={snmpCreds}
          snmpLoading={snmpLoading}
          snmpError={snmpError}
          snmpRevealed={snmpRevealed}
          onLoadCredentials={async () => {
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
          onClearCredentials={() => { setSnmpCreds(null); setSnmpRevealed({}) }}
          onToggleReveal={(key) => setSnmpRevealed((p) => ({ ...p, [key]: !p[key] }))}
        />
      )}

      {modal === 'edit' && (
        <EditDeviceModal
          editForm={editForm}
          setEditForm={setEditForm}
          deviceTypes={deviceTypes}
          locations={locations}
          racks={racks}
          cfDefs={cfDefs}
          saving={saving}
          onSubmit={handleEditSubmit}
          onClose={() => setModal(null)}
          onLocationChange={loadRacksForLocation}
        />
      )}

      {modal === 'assoc' && (
        <AssociateIPModal
          assocForm={assocForm}
          setAssocForm={setAssocForm}
          ipSearch={ipSearch}
          ipSearchResults={ipSearchResults}
          ipSearching={ipSearching}
          ipCreating={ipCreating}
          selectedIpLabel={selectedIpLabel}
          saving={saving}
          interfaces={interfaces}
          onCreateInterface={handleCreateInterfaceFromAssoc}
          onSearchChange={handleIpSearchChange}
          onSelectResult={selectIpResult}
          onQuickCreate={handleQuickCreate}
          onSubmit={handleAssocSubmit}
          onClose={() => setModal(null)}
        />
      )}

      {isAdmin && <ChangeHistory resourceType="device" resourceId={device?.id} />}

      {(modal === 'iface-add' || modal?.ifaceEdit) && (
        <InterfaceModal
          modal={modal}
          ifaceForm={ifaceForm}
          setIfaceForm={setIfaceForm}
          vlanList={vlanList}
          saving={saving}
          onSubmit={handleIfaceSubmit}
          onClose={() => setModal(null)}
        />
      )}
    </div>
  )
}
