import { useEffect, useState } from 'react'
import { createCircuitProvider, createLogicalCircuit, createPhysicalCircuit, deleteCircuitProvider, deleteLogicalCircuit, deletePhysicalCircuit, getCircuitProviders, getCustomers, getLogicalCircuits, getPhysicalCircuits, updateCircuitProvider, updateLogicalCircuit, updatePhysicalCircuit } from '../api/modules'
import { getLocations } from '../api/locations'
import Modal from '../components/Modal'

const PROVIDER_EMPTY = { name: '', account_no: '', support_email: '', support_phone: '', portal_url: '', notes: '' }
const PHYSICAL_EMPTY = { provider_id: '', circuit_id: '', name: '', type: 'ethernet', status: 'active', bandwidth_mbps: '', location_a_id: '', location_b_id: '', customer_id: '', install_date: '', notes: '' }
const LOGICAL_EMPTY = { physical_circuit_id: '', name: '', service_id: '', type: 'l2vpn', status: 'active', customer_id: '', bandwidth_mbps: '', notes: '' }

export default function CircuitsPage() {
  const [providers, setProviders] = useState([])
  const [physical, setPhysical] = useState([])
  const [logical, setLogical] = useState([])
  const [locations, setLocations] = useState([])
  const [customers, setCustomers] = useState([])
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState({})
  const [error, setError] = useState('')

  useEffect(() => { load() }, [])

  async function load() {
    try {
      const [p, pc, lc] = await Promise.all([getCircuitProviders(), getPhysicalCircuits(), getLogicalCircuits()])
      setProviders(p.data || [])
      setPhysical(pc.data || [])
      setLogical(lc.data || [])
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load circuits')
    }
    const [locs, cust] = await Promise.allSettled([getLocations(), getCustomers()])
    if (locs.status === 'fulfilled') setLocations(Array.isArray(locs.value) ? locs.value : [])
    if (cust.status === 'fulfilled') setCustomers(cust.value.data || [])
  }

  async function saveProvider(e) {
    e.preventDefault()
    if (modal.item?.id) await updateCircuitProvider(modal.item.id, form)
    else await createCircuitProvider(form)
    setModal(null)
    load()
  }

  async function savePhysical(e) {
    e.preventDefault()
    const body = {
      ...form,
      provider_id: Number(form.provider_id),
      bandwidth_mbps: form.bandwidth_mbps ? Number(form.bandwidth_mbps) : null,
      location_a_id: form.location_a_id ? Number(form.location_a_id) : null,
      location_b_id: form.location_b_id ? Number(form.location_b_id) : null,
      customer_id: form.customer_id ? Number(form.customer_id) : null,
      install_date: form.install_date || null,
    }
    if (modal.item?.id) await updatePhysicalCircuit(modal.item.id, body)
    else await createPhysicalCircuit(body)
    setModal(null)
    load()
  }

  async function saveLogical(e) {
    e.preventDefault()
    const body = {
      ...form,
      physical_circuit_id: form.physical_circuit_id ? Number(form.physical_circuit_id) : null,
      customer_id: form.customer_id ? Number(form.customer_id) : null,
      bandwidth_mbps: form.bandwidth_mbps ? Number(form.bandwidth_mbps) : null,
    }
    if (modal.item?.id) await updateLogicalCircuit(modal.item.id, body)
    else await createLogicalCircuit(body)
    setModal(null)
    load()
  }

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Circuits</h1>
        <div className="flex gap-2">
          <button onClick={() => { setForm(PROVIDER_EMPTY); setModal({ type: 'provider' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">+ Provider</button>
          <button onClick={() => { setForm(PHYSICAL_EMPTY); setModal({ type: 'physical' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">+ Physical</button>
          <button onClick={() => { setForm(LOGICAL_EMPTY); setModal({ type: 'logical' }) }} className="px-4 py-2 bg-blue-600 text-white rounded text-sm">+ Logical</button>
        </div>
      </div>
      {error && <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>}

      <Table title="Providers" cols={['Name', 'Account', 'Support', 'Portal']} empty="No providers yet." rows={providers.map(p => [p.name, p.accountNo || '-', p.supportEmail || p.supportPhone || '-', p.portalUrl || '-', <Actions key={p.id} onEdit={() => { setForm({ ...PROVIDER_EMPTY, ...p, account_no: p.accountNo, support_email: p.supportEmail, support_phone: p.supportPhone, portal_url: p.portalUrl }); setModal({ type: 'provider', item: p }) }} onDelete={async () => { await deleteCircuitProvider(p.id); load() }} />])} />
      <Table title="Physical Circuits" cols={['Name', 'Circuit ID', 'Provider', 'Status', 'Customer']} empty="No physical circuits yet." rows={physical.map(p => [p.name, p.circuitId, p.providerName || p.providerId, p.status, p.customerName || '-', <Actions key={p.id} onEdit={() => { setForm({ ...PHYSICAL_EMPTY, provider_id: p.providerId, circuit_id: p.circuitId, name: p.name, type: p.type, status: p.status, bandwidth_mbps: p.bandwidthMbps || '', location_a_id: p.locationAId || '', location_b_id: p.locationBId || '', customer_id: p.customerId || '', notes: p.notes || '' }); setModal({ type: 'physical', item: p }) }} onDelete={async () => { await deletePhysicalCircuit(p.id); load() }} />])} />
      <Table title="Logical Circuits" cols={['Name', 'Service ID', 'Type', 'Status', 'Customer']} empty="No logical circuits yet." rows={logical.map(l => [l.name, l.serviceId || '-', l.type, l.status, l.customerName || '-', <Actions key={l.id} onEdit={() => { setForm({ ...LOGICAL_EMPTY, physical_circuit_id: l.physicalCircuitId || '', name: l.name, service_id: l.serviceId || '', type: l.type, status: l.status, customer_id: l.customerId || '', bandwidth_mbps: l.bandwidthMbps || '', notes: l.notes || '' }); setModal({ type: 'logical', item: l }) }} onDelete={async () => { await deleteLogicalCircuit(l.id); load() }} />])} />

      {modal?.type === 'provider' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">Circuit Provider</h2><form onSubmit={saveProvider} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {['name', 'account_no', 'support_email', 'support_phone', 'portal_url', 'notes'].map(field => <input key={field} required={field === 'name'} placeholder={label(field)} value={form[field] || ''} onChange={e => setForm(f => ({ ...f, [field]: e.target.value }))} className="border rounded px-3 py-2 text-sm" />)}
        <Submit />
      </form></Modal>}

      {modal?.type === 'physical' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">Physical Circuit</h2><form onSubmit={savePhysical} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <select required value={form.provider_id || ''} onChange={e => setForm(f => ({ ...f, provider_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">Provider</option>{providers.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}</select>
        <input required placeholder="Circuit ID" value={form.circuit_id || ''} onChange={e => setForm(f => ({ ...f, circuit_id: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input required placeholder="Name" value={form.name || ''} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder="Type" value={form.type || ''} onChange={e => setForm(f => ({ ...f, type: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <StatusSelect />
        <input type="number" placeholder="Bandwidth Mbps" value={form.bandwidth_mbps || ''} onChange={e => setForm(f => ({ ...f, bandwidth_mbps: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <LocationSelect field="location_a_id" locations={locations} form={form} setForm={setForm} labelText="Location A" />
        <LocationSelect field="location_b_id" locations={locations} form={form} setForm={setForm} labelText="Location B" />
        <CustomerSelect customers={customers} form={form} setForm={setForm} />
        <input type="date" value={form.install_date || ''} onChange={e => setForm(f => ({ ...f, install_date: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder="Notes" value={form.notes || ''} onChange={e => setForm(f => ({ ...f, notes: e.target.value }))} className="md:col-span-2 border rounded px-3 py-2 text-sm" />
        <Submit />
      </form></Modal>}

      {modal?.type === 'logical' && <Modal onClose={() => setModal(null)}><h2 className="text-lg font-semibold mb-4">Logical Circuit</h2><form onSubmit={saveLogical} className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <select value={form.physical_circuit_id || ''} onChange={e => setForm(f => ({ ...f, physical_circuit_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">No physical circuit</option>{physical.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}</select>
        <input required placeholder="Name" value={form.name || ''} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder="Service ID" value={form.service_id || ''} onChange={e => setForm(f => ({ ...f, service_id: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <input placeholder="Type" value={form.type || ''} onChange={e => setForm(f => ({ ...f, type: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <StatusSelect />
        <input type="number" placeholder="Bandwidth Mbps" value={form.bandwidth_mbps || ''} onChange={e => setForm(f => ({ ...f, bandwidth_mbps: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <CustomerSelect customers={customers} form={form} setForm={setForm} />
        <input placeholder="Notes" value={form.notes || ''} onChange={e => setForm(f => ({ ...f, notes: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
        <Submit />
      </form></Modal>}
    </div>
  )

  function StatusSelect() {
    return <select value={form.status || 'active'} onChange={e => setForm(f => ({ ...f, status: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">Active</option><option value="planned">Planned</option><option value="down">Down</option><option value="retired">Retired</option></select>
  }

  function Submit() {
    return <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setModal(null)} className="px-4 py-2 border rounded text-sm">Cancel</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">Save</button></div>
  }
}

function Table({ title, cols, rows, empty }) {
  return <network><h2 className="text-lg font-semibold mb-3">{title}</h2><div className="overflow-x-auto rounded border border-gray-200"><table className="min-w-full divide-y divide-gray-200 text-sm"><thead className="bg-gray-50"><tr>{cols.map(c => <th key={c} className="px-4 py-3 text-left">{c}</th>)}<th /></tr></thead><tbody className="divide-y divide-gray-100 bg-white">{rows.length === 0 && <tr><td colSpan={cols.length + 1} className="px-4 py-6 text-center text-gray-500">{empty}</td></tr>}{rows.map((row, i) => <tr key={i}>{row.map((cell, j) => <td key={j} className="px-4 py-3">{cell}</td>)}</tr>)}</tbody></table></div></network>
}

function Actions({ onEdit, onDelete }) {
  return <div className="text-right space-x-2"><button onClick={onEdit} className="text-blue-600 text-xs">Edit</button><button onClick={onDelete} className="text-red-600 text-xs">Delete</button></div>
}

function LocationSelect({ field, locations, form, setForm, labelText }) {
  return <select value={form[field] || ''} onChange={e => setForm(f => ({ ...f, [field]: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">{labelText}</option>{locations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}</select>
}

function CustomerSelect({ customers, form, setForm }) {
  return <select value={form.customer_id || ''} onChange={e => setForm(f => ({ ...f, customer_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">No customer</option>{customers.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}</select>
}

function label(field) {
  return field.split('_').map(part => part.charAt(0).toUpperCase() + part.slice(1)).join(' ')
}
