import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { createDHCPLease, createDHCPServer, deleteDHCPLease, deleteDHCPServer, getDHCPLeases, getDHCPServers, updateDHCPLease, updateDHCPServer } from '../api/modules'
import { getLocations } from '../api/locations'
import Modal from '../components/Modal'

const SERVER_EMPTY = { name: '', address: '', vendor: '', version: '', location_id: '', description: '', status: 'active' }
const LEASE_EMPTY = { server_id: '', ip_address: '', mac_address: '', hostname: '', state: 'active' }

export default function DHCPPage() {
  const { t } = useTranslation()
  const [servers, setServers] = useState([])
  const [leases, setLeases] = useState([])
  const [locations, setLocations] = useState([])
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState({})
  const [error, setError] = useState('')

  useEffect(() => { load() }, [])

  async function load() {
    try {
      const [serversRes, leasesRes] = await Promise.all([getDHCPServers(), getDHCPLeases()])
      setServers(serversRes.data || [])
      setLeases(leasesRes.data || [])
    } catch (err) {
      setError(err.response?.data?.error || t('dhcp.loadError'))
    }
    try {
      const locationsRes = await getLocations()
      setLocations(Array.isArray(locationsRes) ? locationsRes : [])
    } catch {
      // Locations feature may be disabled; location dropdown will be empty
    }
  }

  async function saveServer(e) {
    e.preventDefault()
    const body = { ...form, location_id: form.location_id ? Number(form.location_id) : null }
    try {
      if (modal?.item?.id) await updateDHCPServer(modal.item.id, body)
      else await createDHCPServer(body)
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('natRules.saveFailed'))
    }
  }

  async function saveLease(e) {
    e.preventDefault()
    const body = { ...form, server_id: Number(form.server_id) }
    try {
      if (modal?.item?.id) await updateDHCPLease(modal.item.id, body)
      else await createDHCPLease(body)
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('natRules.saveFailed'))
    }
  }

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{t('nav.dhcp')}</h1>
        <div className="flex gap-2">
          <button onClick={() => { setForm(SERVER_EMPTY); setModal({ type: 'server' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('dhcp.addServer')}</button>
          <button onClick={() => { setForm(LEASE_EMPTY); setModal({ type: 'lease' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('dhcp.addLease')}</button>
        </div>
      </div>
      {error && <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>}

      <network>
        <h2 className="text-lg font-semibold mb-3">{t('dhcp.serversTitle')}</h2>
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50"><tr><th className="px-4 py-3 text-left">{t('common.name')}</th><th className="px-4 py-3 text-left">{t('dhcp.address')}</th><th className="px-4 py-3 text-left">{t('deviceInfo.vendor')}</th><th className="px-4 py-3 text-left">{t('subnets.location')}</th><th /></tr></thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {servers.length === 0 && <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-500">{t('dhcp.noServersYet')}</td></tr>}
              {servers.map(s => <tr key={s.id}><td className="px-4 py-3 font-medium">{s.name}</td><td className="px-4 py-3 font-mono">{s.address}</td><td className="px-4 py-3">{s.vendor || '-'}</td><td className="px-4 py-3">{s.locationName || '-'}</td><td className="px-4 py-3 text-right space-x-2"><button className="text-blue-600 text-xs" onClick={() => { setForm({ ...SERVER_EMPTY, ...s, location_id: s.locationId || '' }); setModal({ type: 'server', item: s }) }}>{t('common.edit')}</button><button className="text-red-600 text-xs" onClick={async () => { await deleteDHCPServer(s.id); load() }}>{t('common.delete')}</button></td></tr>)}
            </tbody>
          </table>
        </div>
      </network>

      <network>
        <h2 className="text-lg font-semibold mb-3">{t('dhcp.leasesTitle')}</h2>
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50"><tr><th className="px-4 py-3 text-left">{t('associateIp.ipAddress')}</th><th className="px-4 py-3 text-left">{t('dhcp.mac')}</th><th className="px-4 py-3 text-left">{t('dashboard.hostname')}</th><th className="px-4 py-3 text-left">{t('dhcp.server')}</th><th className="px-4 py-3 text-left">{t('dhcp.state')}</th><th /></tr></thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {leases.length === 0 && <tr><td colSpan={6} className="px-4 py-6 text-center text-gray-500">{t('dhcp.noLeasesYet')}</td></tr>}
              {leases.map(l => <tr key={l.id}><td className="px-4 py-3 font-mono">{l.ipAddress}</td><td className="px-4 py-3 font-mono">{l.macAddress}</td><td className="px-4 py-3">{l.hostname || '-'}</td><td className="px-4 py-3">{l.serverName || l.serverId}</td><td className="px-4 py-3">{l.state}</td><td className="px-4 py-3 text-right space-x-2"><button className="text-blue-600 text-xs" onClick={() => { setForm({ ...LEASE_EMPTY, server_id: l.serverId, ip_address: l.ipAddress, mac_address: l.macAddress, hostname: l.hostname, state: l.state }); setModal({ type: 'lease', item: l }) }}>{t('common.edit')}</button><button className="text-red-600 text-xs" onClick={async () => { await deleteDHCPLease(l.id); load() }}>{t('common.delete')}</button></td></tr>)}
            </tbody>
          </table>
        </div>
      </network>

      {modal?.type === 'server' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">{t('dhcp.serverModalTitle')}</h2><form onSubmit={saveServer} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <input required placeholder={t('natRules.namePlaceholder')} value={form.name || ''} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input required placeholder={t('dhcp.ipAddressPlaceholder')} value={form.address || ''} onChange={e => setForm(f => ({ ...f, address: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder={t('dhcp.vendorPlaceholder')} value={form.vendor || ''} onChange={e => setForm(f => ({ ...f, vendor: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder={t('dhcp.versionPlaceholder')} value={form.version || ''} onChange={e => setForm(f => ({ ...f, version: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <select value={form.location_id || ''} onChange={e => setForm(f => ({ ...f, location_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">{t('subnetForm.noLocation')}</option>{locations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}</select>
        <select value={form.status || 'active'} onChange={e => setForm(f => ({ ...f, status: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">{t('natRules.active')}</option><option value="disabled">{t('natRules.disabled')}</option><option value="planned">{t('natRules.planned')}</option><option value="retired">{t('natRules.retired')}</option></select>
        <input placeholder={t('natRules.descriptionPlaceholder')} value={form.description || ''} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className="md:col-span-2 border rounded px-3 py-2 text-sm" />
        <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setModal(null)} className="px-4 py-2 border rounded text-sm">{t('common.cancel')}</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('common.save')}</button></div>
      </form></Modal>}

      {modal?.type === 'lease' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">{t('dhcp.leaseModalTitle')}</h2><form onSubmit={saveLease} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <select required value={form.server_id || ''} onChange={e => setForm(f => ({ ...f, server_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">{t('dhcp.selectServer')}</option>{servers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}</select>
        <select value={form.state || 'active'} onChange={e => setForm(f => ({ ...f, state: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">{t('natRules.active')}</option><option value="expired">{t('dhcp.expired')}</option><option value="reserved">{t('dhcp.reserved')}</option><option value="declined">{t('dhcp.declined')}</option><option value="released">{t('dhcp.released')}</option></select>
        <input required placeholder={t('dhcp.ipAddressPlaceholder')} value={form.ip_address || ''} onChange={e => setForm(f => ({ ...f, ip_address: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input required placeholder={t('dhcp.macAddressPlaceholder')} value={form.mac_address || ''} onChange={e => setForm(f => ({ ...f, mac_address: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder={t('dhcp.hostnamePlaceholder')} value={form.hostname || ''} onChange={e => setForm(f => ({ ...f, hostname: e.target.value }))} className="md:col-span-2 border rounded px-3 py-2 text-sm" />
        <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setModal(null)} className="px-4 py-2 border rounded text-sm">{t('common.cancel')}</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">{t('common.save')}</button></div>
      </form></Modal>}
    </div>
  )
}
