import { useEffect, useRef, useState } from 'react'
import { getDevices } from '../api/devices'
import { getNetworks } from '../api/ipam'
import { searchIPAddressesGlobal, globalSearch } from '../api/ipam'
import { getLocations } from '../api/locations'
import { createCustomer, createCustomerAssociation, deleteCustomer, deleteCustomerAssociation, getCustomerAssociations, getCustomers, updateCustomer } from '../api/modules'
import { getDHCPServers, getPhysicalCircuits, getLogicalCircuits, getNATRules } from '../api/modules'
import { getRacks } from '../api/racks'
import { getVlans, getVrfs } from '../api/vlans'
import Modal from '../components/Modal'

const EMPTY_FORM = { name: '', description: '', email: '', phone: '', notes: '' }
const EMPTY_ASSOC = { customer_id: '', object_type: 'subnet', object_id: '', object_name: '', relationship: 'owner', notes: '' }

const OBJECT_TYPES = [
  { value: 'network',           label: 'Network' },
  { value: 'subnet',            label: 'Subnet' },
  { value: 'ip_address',        label: 'IP Address' },
  { value: 'device',            label: 'Device' },
  { value: 'rack',              label: 'Rack' },
  { value: 'location',          label: 'Location' },
  { value: 'vlan',              label: 'VLAN' },
  { value: 'vrf',               label: 'VRF' },
  { value: 'nat_rule',          label: 'NAT Rule' },
  { value: 'dhcp_server',       label: 'DHCP Server' },
  { value: 'physical_circuit',  label: 'Physical Circuit' },
  { value: 'logical_circuit',   label: 'Logical Circuit' },
]

const OBJECT_TYPE_LABEL = Object.fromEntries(OBJECT_TYPES.map(t => [t.value, t.label]))

const RELATIONSHIPS = ['owner', 'consumer', 'billing', 'technical', 'stakeholder']

function labelFor(item, type) {
  if (!item) return ''
  switch (type) {
    case 'network':          return item.name || `#${item.id}`
    case 'subnet': {
      const cidr = item.networkAddress && item.prefixLength != null
        ? `${item.networkAddress}/${item.prefixLength}`
        : null
      return cidr ? `${cidr}${item.description ? ' — ' + item.description : ''}` : `#${item.id}`
    }
    case 'ip_address':       return item.address || `#${item.id}`
    case 'device':           return item.hostname || item.name || `#${item.id}`
    case 'rack':             return item.name || `#${item.id}`
    case 'location':         return item.name || `#${item.id}`
    case 'vlan':             return item.name ? `${item.name} (VID ${item.vid ?? item.vlanId ?? ''})` : `VID ${item.vid ?? item.vlanId ?? item.id}`
    case 'vrf':              return item.name || `#${item.id}`
    case 'nat_rule':         return item.name || `${item.sourceAddress || ''} → ${item.translatedAddress || ''}` || `#${item.id}`
    case 'dhcp_server':      return item.name || `#${item.id}`
    case 'physical_circuit': return item.name || item.circuitId || `#${item.id}`
    case 'logical_circuit':  return item.name || item.serviceId || `#${item.id}`
    default:                 return `#${item.id}`
  }
}

async function fetchObjectList(type) {
  switch (type) {
    case 'network':          return (await getNetworks()).data || []
    case 'device':           return (await getDevices()).data || []
    case 'rack':             return await getRacks() || []
    case 'location':         return await getLocations() || []
    case 'vlan':             return (await getVlans()).data || []
    case 'vrf':              return (await getVrfs()).data || []
    case 'nat_rule':         return (await getNATRules()).data || []
    case 'dhcp_server':      return (await getDHCPServers()).data || []
    case 'physical_circuit': return (await getPhysicalCircuits()).data || []
    case 'logical_circuit':  return (await getLogicalCircuits()).data || []
    default:                 return null
  }
}

function ObjectPicker({ type, value, label, onChange }) {
  const [items, setItems] = useState(null)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState(label || '')
  const [open, setOpen] = useState(false)
  const [searchResults, setSearchResults] = useState([])
  const searchTimer = useRef(null)
  const inputRef = useRef(null)

  const isSearchType = type === 'subnet' || type === 'ip_address'

  useEffect(() => {
    setQuery(label || '')
    setItems(null)
    setSearchResults([])
  }, [type])

  useEffect(() => {
    setQuery(label || '')
  }, [label])

  useEffect(() => {
    if (isSearchType || items !== null) return
    setLoading(true)
    fetchObjectList(type)
      .then(list => setItems(list || []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }, [type, isSearchType, items])

  function handleInput(e) {
    const q = e.target.value
    setQuery(q)
    setOpen(true)
    onChange(null, '')

    if (isSearchType) {
      clearTimeout(searchTimer.current)
      if (q.trim().length < 2) { setSearchResults([]); return }
      searchTimer.current = setTimeout(async () => {
        try {
          if (type === 'ip_address') {
            const res = await searchIPAddressesGlobal(q)
            setSearchResults((res.data || []).map(ip => ({ id: ip.id, _label: ip.address || `#${ip.id}` })))
          } else {
            const res = await globalSearch(q)
            const subnets = res.data?.subnets || []
            setSearchResults(subnets.map(s => ({ id: s.id, _label: labelFor(s, 'subnet') })))
          }
        } catch {
          setSearchResults([])
        }
      }, 250)
    }
  }

  function select(item) {
    const lbl = item._label || labelFor(item, type)
    setQuery(lbl)
    setOpen(false)
    onChange(item.id, lbl)
  }

  const filtered = isSearchType
    ? searchResults
    : (items || []).filter(item => {
        const lbl = labelFor(item, type)
        return lbl.toLowerCase().includes(query.toLowerCase())
      })

  return (
    <div className="relative">
      <input
        ref={inputRef}
        type="text"
        value={query}
        onChange={handleInput}
        onFocus={() => setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
        placeholder={loading ? 'Loading…' : isSearchType ? 'Type to search…' : 'Search or select…'}
        className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      {open && (
        <div className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded shadow-lg max-h-52 overflow-y-auto">
          {loading && <div className="px-3 py-2 text-sm text-gray-400">Loading…</div>}
          {!loading && filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-400">
              {isSearchType && query.trim().length < 2 ? 'Type at least 2 characters' : 'No matches'}
            </div>
          )}
          {filtered.map(item => {
            const lbl = item._label || labelFor(item, type)
            return (
              <button
                key={item.id}
                type="button"
                onMouseDown={() => select(item)}
                className="w-full text-left px-3 py-2 text-sm hover:bg-blue-50 focus:bg-blue-50"
              >
                {lbl}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default function CustomersPage() {
  const [customers, setCustomers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [modal, setModal] = useState(null) // 'create' | 'edit' | 'assoc'
  const [editing, setEditing] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [associations, setAssociations] = useState([])
  const [assocForm, setAssocForm] = useState(EMPTY_ASSOC)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      const [res, assocRes] = await Promise.all([getCustomers(), getCustomerAssociations()])
      setCustomers(res.data || [])
      setAssociations(assocRes.data || [])
    } catch {
      setError('Failed to load customers')
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setEditing(null)
    setModal('create')
  }

  function openEdit(c) {
    setForm({ name: c.name, description: c.description, email: c.email, phone: c.phone, notes: c.notes })
    setEditing(c)
    setModal('edit')
  }

  function openAssoc() {
    setAssocForm(EMPTY_ASSOC)
    setError('')
    setModal('assoc')
  }

  function closeModal() {
    setModal(null)
    setEditing(null)
    setForm(EMPTY_FORM)
    setError('')
  }

  async function handleSave(e) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      if (modal === 'edit') {
        await updateCustomer(editing.id, form)
        setMessage('Customer updated')
      } else {
        await createCustomer(form)
        setMessage('Customer created')
      }
      closeModal()
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteCustomer(id)
      setDeleteConfirm(null)
      setMessage('Customer deleted')
      await load()
    } catch {
      setError('Delete failed')
    }
  }

  async function handleAssociationSave(e) {
    e.preventDefault()
    if (!assocForm.customer_id) { setError('Select a customer'); return }
    if (!assocForm.object_id) { setError('Select an object'); return }
    setError('')
    setSaving(true)
    try {
      await createCustomerAssociation({
        customer_id: Number(assocForm.customer_id),
        object_type: assocForm.object_type,
        object_id: Number(assocForm.object_id),
        object_name: assocForm.object_name,
        relationship: assocForm.relationship,
        notes: assocForm.notes,
      })
      closeModal()
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save association')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Customers</h1>
        <button
          onClick={openCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
        >
          + New Customer
        </button>
      </div>

      {message && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded text-sm">
          {message}
        </div>
      )}
      {error && !modal && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-gray-500 text-sm">Loading…</div>
      ) : customers.length === 0 ? (
        <div className="text-gray-500 text-sm">No customers yet.</div>
      ) : (
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Name</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Email</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Phone</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Description</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {customers.map(c => (
                <tr key={c.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{c.name}</td>
                  <td className="px-4 py-3 text-gray-600">{c.email || '—'}</td>
                  <td className="px-4 py-3 text-gray-600">{c.phone || '—'}</td>
                  <td className="px-4 py-3 text-gray-500 max-w-xs truncate">{c.description || '—'}</td>
                  <td className="px-4 py-3 text-right space-x-2 whitespace-nowrap">
                    <button onClick={() => openEdit(c)} className="text-blue-600 hover:underline text-xs">Edit</button>
                    <button onClick={() => setDeleteConfirm(c)} className="text-red-600 hover:underline text-xs">Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="mt-8">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-gray-900">Associations</h2>
          <button
            onClick={openAssoc}
            className="px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
          >
            + Add Association
          </button>
        </div>
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Customer</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Type</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Object</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">Relationship</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {associations.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-500">No associations yet.</td></tr>
              )}
              {associations.map(a => (
                <tr key={a.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{a.customerName || `#${a.customerId}`}</td>
                  <td className="px-4 py-3 text-gray-600">{OBJECT_TYPE_LABEL[a.objectType] || a.objectType}</td>
                  <td className="px-4 py-3 text-gray-700">{a.objectName || `#${a.objectId}`}</td>
                  <td className="px-4 py-3 text-gray-600 capitalize">{a.relationship}</td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={async () => { await deleteCustomerAssociation(a.id); load() }}
                      className="text-red-600 hover:underline text-xs"
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {modal === 'assoc' && (
        <Modal onClose={closeModal}>
          <h2 className="text-lg font-semibold mb-5">Add Association</h2>
          <form onSubmit={handleAssociationSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Customer</label>
              <select
                required
                value={assocForm.customer_id}
                onChange={e => setAssocForm(f => ({ ...f, customer_id: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select customer…</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Object Type</label>
              <select
                value={assocForm.object_type}
                onChange={e => setAssocForm(f => ({ ...f, object_type: e.target.value, object_id: '', object_name: '' }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {OBJECT_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {OBJECT_TYPE_LABEL[assocForm.object_type] || 'Object'}
              </label>
              <ObjectPicker
                key={assocForm.object_type}
                type={assocForm.object_type}
                value={assocForm.object_id}
                label={assocForm.object_name}
                onChange={(id, name) => setAssocForm(f => ({ ...f, object_id: id ?? '', object_name: name }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Relationship</label>
              <select
                value={assocForm.relationship}
                onChange={e => setAssocForm(f => ({ ...f, relationship: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {RELATIONSHIPS.map(r => <option key={r} value={r}>{r.charAt(0).toUpperCase() + r.slice(1)}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
              <input
                type="text"
                value={assocForm.notes}
                onChange={e => setAssocForm(f => ({ ...f, notes: e.target.value }))}
                placeholder="Optional"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            {error && <p className="text-red-600 text-sm">{error}</p>}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={closeModal} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">
                Cancel
              </button>
              <button type="submit" disabled={saving} className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving…' : 'Add'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {(modal === 'create' || modal === 'edit') && (
        <Modal onClose={closeModal}>
          <h2 className="text-lg font-semibold mb-4">{modal === 'edit' ? 'Edit Customer' : 'New Customer'}</h2>
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
              <input
                type="text"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
              <input
                type="email"
                value={form.email}
                onChange={e => setForm(f => ({ ...f, email: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
              <input
                type="text"
                value={form.phone}
                onChange={e => setForm(f => ({ ...f, phone: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                type="text"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
              <textarea
                value={form.notes}
                onChange={e => setForm(f => ({ ...f, notes: e.target.value }))}
                rows={3}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            {error && <p className="text-red-600 text-sm">{error}</p>}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={closeModal} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">
                Cancel
              </button>
              <button type="submit" disabled={saving} className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving…' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {deleteConfirm && (
        <Modal onClose={() => setDeleteConfirm(null)}>
          <h2 className="text-lg font-semibold mb-2">Delete Customer</h2>
          <p className="text-sm text-gray-600 mb-4">
            Are you sure you want to delete <strong>{deleteConfirm.name}</strong>? This cannot be undone.
          </p>
          <div className="flex justify-end gap-2">
            <button onClick={() => setDeleteConfirm(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">
              Cancel
            </button>
            <button onClick={() => handleDelete(deleteConfirm.id)} className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700">
              Delete
            </button>
          </div>
        </Modal>
      )}
    </div>
  )
}
