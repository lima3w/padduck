import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { createFirewallZone, createFirewallZoneMapping, deleteFirewallZone, deleteFirewallZoneMapping, getFirewallZoneMappings, getFirewallZones, updateFirewallZone, updateFirewallZoneMapping } from '../api/modules'
import Modal from '../components/Modal'

const ZONE_EMPTY = { name: '', description: '', color: '#2563eb', status: 'active' }
const MAPPING_EMPTY = { zone_id: '', object_type: 'cidr', object_id: '', cidr: '', direction: 'both', description: '', status: 'active' }
const OBJECT_TYPES = ['cidr', 'network', 'subnet', 'ip_address', 'device', 'rack', 'location', 'vlan', 'vrf', 'nat_rule', 'dhcp_server', 'dhcp_lease', 'physical_circuit', 'logical_circuit']

export default function FirewallZonesPage() {
  const { t } = useTranslation()
  const [zones, setZones] = useState([])
  const [mappings, setMappings] = useState([])
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => { load() }, [])

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [zonesRes, mappingsRes] = await Promise.all([getFirewallZones(), getFirewallZoneMappings()])
      setZones(zonesRes.data || [])
      setMappings(mappingsRes.data || [])
    } catch (err) {
      setError(err.response?.data?.error || t('firewallZones.loadError'))
    } finally {
      setLoading(false)
    }
  }

  async function saveZone(e) {
    e.preventDefault()
    try {
      if (modal?.item?.id) await updateFirewallZone(modal.item.id, form)
      else await createFirewallZone(form)
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('natRules.saveFailed'))
    }
  }

  async function saveMapping(e) {
    e.preventDefault()
    const body = {
      ...form,
      zone_id: Number(form.zone_id),
      object_id: form.object_id ? Number(form.object_id) : null,
      cidr: form.cidr || '',
    }
    try {
      if (modal?.item?.id) await updateFirewallZoneMapping(modal.item.id, body)
      else await createFirewallZoneMapping(body)
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('natRules.saveFailed'))
    }
  }

  function editZone(zone) {
    setForm({ name: zone.name, description: zone.description || '', color: zone.color || '#2563eb', status: zone.status || 'active' })
    setModal({ type: 'zone', item: zone })
  }

  function editMapping(mapping) {
    setForm({
      zone_id: mapping.zoneId,
      object_type: mapping.objectType || 'cidr',
      object_id: mapping.objectId || '',
      cidr: mapping.cidr || '',
      direction: mapping.direction || 'both',
      description: mapping.description || '',
      status: mapping.status || 'active',
    })
    setModal({ type: 'mapping', item: mapping })
  }

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('nav.firewallZones')}</h1>
        <div className="flex gap-2">
          <button onClick={() => { setForm(ZONE_EMPTY); setModal({ type: 'zone' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('firewallZones.addZone')}</button>
          <button onClick={() => { setForm({ ...MAPPING_EMPTY, zone_id: zones[0]?.id || '' }); setModal({ type: 'mapping' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('firewallZones.addMapping')}</button>
        </div>
      </div>
      {error && <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>}

      {loading ? <div className="text-sm text-gray-500">{t('common.loading')}</div> : (
        <>
          <network>
            <h2 className="text-lg font-semibold mb-3 text-gray-900 dark:text-gray-100">{t('firewallZones.zonesTitle')}</h2>
            <div className="overflow-x-auto rounded border border-gray-200">
              <table className="min-w-full divide-y divide-gray-200 text-sm">
                <thead className="bg-gray-50"><tr><th className="px-4 py-3 text-left">{t('common.name')}</th><th className="px-4 py-3 text-left">{t('common.description')}</th><th className="px-4 py-3 text-left">{t('userTabs.privacy.status')}</th><th /></tr></thead>
                <tbody className="divide-y divide-gray-100 bg-white">
                  {zones.length === 0 && <tr><td colSpan={4} className="px-4 py-6 text-center text-gray-500">{t('firewallZones.noZonesYet')}</td></tr>}
                  {zones.map(zone => (
                    <tr key={zone.id}>
                      <td className="px-4 py-3 font-medium"><span className="inline-block h-3 w-3 rounded-full mr-2 align-middle" style={{ backgroundColor: zone.color }} />{zone.name}</td>
                      <td className="px-4 py-3 text-gray-600">{zone.description || '-'}</td>
                      <td className="px-4 py-3 text-gray-600">{zone.status}</td>
                      <td className="px-4 py-3 text-right space-x-2"><button className="text-blue-600 text-xs" onClick={() => editZone(zone)}>{t('common.edit')}</button><button className="text-red-600 text-xs" onClick={async () => { await deleteFirewallZone(zone.id); load() }}>{t('common.delete')}</button></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </network>

          <network>
            <h2 className="text-lg font-semibold mb-3 text-gray-900 dark:text-gray-100">{t('firewallZones.mappingsTitle')}</h2>
            <div className="overflow-x-auto rounded border border-gray-200">
              <table className="min-w-full divide-y divide-gray-200 text-sm">
                <thead className="bg-gray-50"><tr><th className="px-4 py-3 text-left">{t('firewallZones.zone')}</th><th className="px-4 py-3 text-left">{t('firewallZones.target')}</th><th className="px-4 py-3 text-left">{t('firewallZones.direction')}</th><th className="px-4 py-3 text-left">{t('userTabs.privacy.status')}</th><th /></tr></thead>
                <tbody className="divide-y divide-gray-100 bg-white">
                  {mappings.length === 0 && <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-500">{t('firewallZones.noMappingsYet')}</td></tr>}
                  {mappings.map(mapping => (
                    <tr key={mapping.id}>
                      <td className="px-4 py-3 font-medium">{mapping.zoneName || mapping.zoneId}</td>
                      <td className="px-4 py-3 font-mono">{mapping.cidr || `${mapping.objectType}:${mapping.objectId}`}</td>
                      <td className="px-4 py-3 text-gray-600">{mapping.direction}</td>
                      <td className="px-4 py-3 text-gray-600">{mapping.status}</td>
                      <td className="px-4 py-3 text-right space-x-2"><button className="text-blue-600 text-xs" onClick={() => editMapping(mapping)}>{t('common.edit')}</button><button className="text-red-600 text-xs" onClick={async () => { await deleteFirewallZoneMapping(mapping.id); load() }}>{t('common.delete')}</button></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </network>
        </>
      )}

      {modal?.type === 'zone' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">{t('firewallZones.zoneModalTitle')}</h2><form onSubmit={saveZone} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <input required placeholder={t('natRules.namePlaceholder')} value={form.name || ''} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input type="color" value={form.color || '#2563eb'} onChange={e => setForm(f => ({ ...f, color: e.target.value }))} className="border rounded px-3 py-2 h-10" />
        <select value={form.status || 'active'} onChange={e => setForm(f => ({ ...f, status: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">{t('natRules.active')}</option><option value="planned">{t('natRules.planned')}</option><option value="disabled">{t('natRules.disabled')}</option><option value="retired">{t('natRules.retired')}</option></select>
        <input placeholder={t('natRules.descriptionPlaceholder')} value={form.description || ''} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setModal(null)} className="px-4 py-2 border rounded text-sm">{t('common.cancel')}</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('common.save')}</button></div>
      </form></Modal>}

      {modal?.type === 'mapping' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">{t('firewallZones.mappingModalTitle')}</h2><form onSubmit={saveMapping} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <select required value={form.zone_id || ''} onChange={e => setForm(f => ({ ...f, zone_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">{t('firewallZones.selectZone')}</option>{zones.map(zone => <option key={zone.id} value={zone.id}>{zone.name}</option>)}</select>
        <select value={form.object_type || 'cidr'} onChange={e => setForm(f => ({ ...f, object_type: e.target.value }))} className="border rounded px-3 py-2 text-sm">{OBJECT_TYPES.map(type => <option key={type} value={type}>{type}</option>)}</select>
        <input placeholder={t('firewallZones.objectIdPlaceholder')} value={form.object_id || ''} onChange={e => setForm(f => ({ ...f, object_id: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder={t('firewallZones.cidrPlaceholder')} value={form.cidr || ''} onChange={e => setForm(f => ({ ...f, cidr: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <select value={form.direction || 'both'} onChange={e => setForm(f => ({ ...f, direction: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="both">{t('firewallZones.both')}</option><option value="inbound">{t('firewallZones.inbound')}</option><option value="outbound">{t('firewallZones.outbound')}</option></select>
        <select value={form.status || 'active'} onChange={e => setForm(f => ({ ...f, status: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">{t('natRules.active')}</option><option value="planned">{t('natRules.planned')}</option><option value="disabled">{t('natRules.disabled')}</option><option value="retired">{t('natRules.retired')}</option></select>
        <input placeholder={t('natRules.descriptionPlaceholder')} value={form.description || ''} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className="md:col-span-2 border rounded px-3 py-2 text-sm" />
        <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setModal(null)} className="px-4 py-2 border rounded text-sm">{t('common.cancel')}</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('common.save')}</button></div>
      </form></Modal>}
    </div>
  )
}
