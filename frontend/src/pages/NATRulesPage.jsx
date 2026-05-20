import { useEffect, useState } from 'react'
import { createNATRule, deleteNATRule, getCustomers, getNATRules, updateNATRule } from '../api/client'
import Modal from '../components/Modal'

const EMPTY = { name: '', type: 'static', internal_cidr: '', external_cidr: '', protocol: 'any', internal_port: '', external_port: '', customer_id: '', description: '', status: 'active' }

export default function NATRulesPage() {
  const [items, setItems] = useState([])
  const [customers, setCustomers] = useState([])
  const [form, setForm] = useState(EMPTY)
  const [editing, setEditing] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => { load() }, [])

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [rulesRes, customersRes] = await Promise.all([getNATRules(), getCustomers()])
      setItems(rulesRes.data || [])
      setCustomers(customersRes.data || [])
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load NAT rules')
    } finally {
      setLoading(false)
    }
  }

  function openEdit(item) {
    setEditing(item)
    setForm({
      name: item.name || '',
      type: item.type || 'static',
      internal_cidr: item.internalCidr || '',
      external_cidr: item.externalCidr || '',
      protocol: item.protocol || 'any',
      internal_port: item.internalPort || '',
      external_port: item.externalPort || '',
      customer_id: item.customerId || '',
      description: item.description || '',
      status: item.status || 'active',
    })
  }

  async function save(e) {
    e.preventDefault()
    const body = {
      ...form,
      internal_port: form.internal_port ? Number(form.internal_port) : null,
      external_port: form.external_port ? Number(form.external_port) : null,
      customer_id: form.customer_id ? Number(form.customer_id) : null,
    }
    try {
      if (editing?.id) await updateNATRule(editing.id, body)
      else await createNATRule(body)
      setEditing(null)
      setForm(EMPTY)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Save failed')
    }
  }

  async function remove(id) {
    await deleteNATRule(id)
    await load()
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">NAT Rules</h1>
        <button onClick={() => { setEditing({}); setForm(EMPTY) }} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">+ New NAT Rule</button>
      </div>
      {error && <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>}
      {loading ? <div className="text-gray-500 text-sm">Loading...</div> : (
        <div className="overflow-x-auto rounded border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50"><tr>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Name</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Type</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Internal</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">External</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Protocol</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Customer</th>
              <th className="px-4 py-3" />
            </tr></thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {items.length === 0 && <tr><td colSpan={7} className="px-4 py-6 text-center text-gray-500">No NAT rules yet.</td></tr>}
              {items.map(item => (
                <tr key={item.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{item.name}</td>
                  <td className="px-4 py-3 text-gray-600">{item.type}</td>
                  <td className="px-4 py-3 font-mono text-gray-700">{item.internalCidr}{item.internalPort ? `:${item.internalPort}` : ''}</td>
                  <td className="px-4 py-3 font-mono text-gray-700">{item.externalCidr}{item.externalPort ? `:${item.externalPort}` : ''}</td>
                  <td className="px-4 py-3 text-gray-600">{item.protocol}</td>
                  <td className="px-4 py-3 text-gray-600">{item.customerName || '-'}</td>
                  <td className="px-4 py-3 text-right space-x-2"><button onClick={() => openEdit(item)} className="text-blue-600 text-xs hover:underline">Edit</button><button onClick={() => remove(item.id)} className="text-red-600 text-xs hover:underline">Delete</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {editing && (
        <Modal onClose={() => setEditing(null)}>
          <h2 className="text-lg font-semibold mb-4">{editing.id ? 'Edit NAT Rule' : 'New NAT Rule'}</h2>
          <form onSubmit={save} className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <input required placeholder="Name" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
            <select value={form.type} onChange={e => setForm(f => ({ ...f, type: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="static">Static</option><option value="dynamic">Dynamic</option><option value="pat">PAT</option></select>
            <input required placeholder="Internal IP or CIDR" value={form.internal_cidr} onChange={e => setForm(f => ({ ...f, internal_cidr: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
            <input required placeholder="External IP or CIDR" value={form.external_cidr} onChange={e => setForm(f => ({ ...f, external_cidr: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
            <select value={form.protocol} onChange={e => setForm(f => ({ ...f, protocol: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="any">Any</option><option value="tcp">TCP</option><option value="udp">UDP</option><option value="icmp">ICMP</option></select>
            <select value={form.status} onChange={e => setForm(f => ({ ...f, status: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="active">Active</option><option value="disabled">Disabled</option><option value="planned">Planned</option><option value="retired">Retired</option></select>
            <input type="number" min="1" max="65535" placeholder="Internal port" value={form.internal_port} onChange={e => setForm(f => ({ ...f, internal_port: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
            <input type="number" min="1" max="65535" placeholder="External port" value={form.external_port} onChange={e => setForm(f => ({ ...f, external_port: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
            <select value={form.customer_id} onChange={e => setForm(f => ({ ...f, customer_id: e.target.value }))} className="border rounded px-3 py-2 text-sm"><option value="">No customer</option>{customers.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}</select>
            <input placeholder="Description" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className="border rounded px-3 py-2 text-sm" />
            <div className="md:col-span-2 flex justify-end gap-2"><button type="button" onClick={() => setEditing(null)} className="px-4 py-2 border rounded text-sm">Cancel</button><button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded text-sm">Save</button></div>
          </form>
        </Modal>
      )}
    </div>
  )
}
