import { useEffect, useState } from 'react'
import { createCustomer, createCustomerAssociation, deleteCustomer, deleteCustomerAssociation, getCustomerAssociations, getCustomers, updateCustomer } from '../api/modules'
import Modal from '../components/Modal'

const EMPTY_FORM = { name: '', description: '', email: '', phone: '', notes: '' }
const EMPTY_ASSOC = { customer_id: '', object_type: 'subnet', object_id: '', relationship: 'owner', notes: '' }

export default function CustomersPage() {
  const [customers, setCustomers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [modal, setModal] = useState(null) // 'create' | 'edit'
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
      const res = await getCustomers()
      const assocRes = await getCustomerAssociations()
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

  function closeModal() {
    setModal(null)
    setEditing(null)
    setForm(EMPTY_FORM)
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
    try {
      await createCustomerAssociation({
        ...assocForm,
        customer_id: Number(assocForm.customer_id),
        object_id: Number(assocForm.object_id),
      })
      setAssocForm(EMPTY_ASSOC)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save association')
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
      {error && (
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
                    <button
                      onClick={() => openEdit(c)}
                      className="text-blue-600 hover:underline text-xs"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => setDeleteConfirm(c)}
                      className="text-red-600 hover:underline text-xs"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-900 mb-3">Customer Associations</h2>
        <form onSubmit={handleAssociationSave} className="grid grid-cols-1 md:grid-cols-5 gap-3 mb-4">
          <select required value={assocForm.customer_id} onChange={e => setAssocForm(f => ({ ...f, customer_id: e.target.value }))} className="border border-gray-300 rounded px-3 py-2 text-sm">
            <option value="">Customer</option>
            {customers.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
          <select value={assocForm.object_type} onChange={e => setAssocForm(f => ({ ...f, object_type: e.target.value }))} className="border border-gray-300 rounded px-3 py-2 text-sm">
            {['network', 'subnet', 'ip_address', 'device', 'rack', 'location', 'vlan', 'vrf', 'nat_rule', 'dhcp_server', 'dhcp_lease', 'physical_circuit', 'logical_circuit'].map(t => <option key={t} value={t}>{t}</option>)}
          </select>
          <input required type="number" min="1" placeholder="Object ID" value={assocForm.object_id} onChange={e => setAssocForm(f => ({ ...f, object_id: e.target.value }))} className="border border-gray-300 rounded px-3 py-2 text-sm" />
          <select value={assocForm.relationship} onChange={e => setAssocForm(f => ({ ...f, relationship: e.target.value }))} className="border border-gray-300 rounded px-3 py-2 text-sm">
            {['owner', 'consumer', 'billing', 'technical', 'stakeholder'].map(r => <option key={r} value={r}>{r}</option>)}
          </select>
          <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">Add Association</button>
        </form>
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50"><tr><th className="px-4 py-3 text-left font-medium text-gray-600">Customer</th><th className="px-4 py-3 text-left font-medium text-gray-600">Object</th><th className="px-4 py-3 text-left font-medium text-gray-600">Relationship</th><th /></tr></thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {associations.length === 0 && <tr><td colSpan={4} className="px-4 py-6 text-center text-gray-500">No associations yet.</td></tr>}
              {associations.map(a => <tr key={a.id}><td className="px-4 py-3">{a.customerName || a.customerId}</td><td className="px-4 py-3 font-mono">{a.objectType} #{a.objectId}</td><td className="px-4 py-3">{a.relationship}</td><td className="px-4 py-3 text-right"><button onClick={async () => { await deleteCustomerAssociation(a.id); load() }} className="text-red-600 text-xs">Delete</button></td></tr>)}
            </tbody>
          </table>
        </div>
      </div>

      {modal && (
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
