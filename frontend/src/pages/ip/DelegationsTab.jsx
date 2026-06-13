import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import Modal from '../../components/Modal'
import PageSpinner from '../../components/PageSpinner'
import ErrorBanner from '../../components/ErrorBanner'
import EmptyRow from '../../components/EmptyRow'

export default function DelegationsTab({ subnetId }) {
  const [delegations, setDelegations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState({ delegatedPrefix: '', delegatedToDescription: '', delegatedToDeviceId: '', validLifetimeSec: '', preferredLifetimeSec: '' })
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { loadDelegations() }, [subnetId])

  async function loadDelegations() {
    try {
      setLoading(true)
      const { data } = await api.get(`/subnets/${subnetId}/delegations`)
      setDelegations(Array.isArray(data) ? data : (data?.delegations ?? []))
    } catch {
      setError('Failed to load delegations')
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm({ delegatedPrefix: '', delegatedToDescription: '', delegatedToDeviceId: '', validLifetimeSec: '', preferredLifetimeSec: '' })
    setModal('create')
  }

  function openEdit(d) {
    setForm({
      delegatedPrefix: d.delegatedPrefix || '',
      delegatedToDescription: d.delegatedToDescription || '',
      delegatedToDeviceId: d.delegatedToDeviceId ? String(d.delegatedToDeviceId) : '',
      validLifetimeSec: d.validLifetimeSec ? String(d.validLifetimeSec) : '',
      preferredLifetimeSec: d.preferredLifetimeSec ? String(d.preferredLifetimeSec) : '',
    })
    setModal({ edit: d })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        delegated_prefix: form.delegatedPrefix,
        delegated_to_description: form.delegatedToDescription || null,
        delegated_to_device_id: form.delegatedToDeviceId ? parseInt(form.delegatedToDeviceId) : null,
        valid_lifetime_sec: form.validLifetimeSec ? parseInt(form.validLifetimeSec) : null,
        preferred_lifetime_sec: form.preferredLifetimeSec ? parseInt(form.preferredLifetimeSec) : null,
      }
      if (modal === 'create') {
        await api.post(`/subnets/${subnetId}/delegations`, body)
      } else {
        await api.put(`/delegations/${modal.edit.id}`, body)
      }
      setModal(null)
      loadDelegations()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save delegation')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await api.delete(`/delegations/${id}`)
      setDeleteConfirm(null)
      loadDelegations()
    } catch {
      setError('Failed to delete delegation')
    }
  }

  const inputClass = "w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
  const labelClass = "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"

  if (loading) return <PageSpinner message="Loading delegations..." />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">IPv6 Prefix Delegations</h2>
        <button onClick={openCreate} className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">+ Add Delegation</button>
      </div>
      <ErrorBanner error={error} />
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Delegated Prefix</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description / Device</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Expires At</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {delegations.length === 0 && (
              <EmptyRow colSpan={5} message="No delegations yet." />
            )}
            {delegations.map(d => (
              <tr key={d.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-mono text-blue-600 dark:text-blue-400">{d.delegatedPrefix}</td>
                <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{d.delegatedToDescription || (d.delegatedToDeviceId ? `Device #${d.delegatedToDeviceId}` : '—')}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{d.expiresAt ? new Date(d.expiresAt).toLocaleString() : '—'}</td>
                <td className="px-4 py-3">
                  <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${d.isExpired ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'}`}>
                    {d.isExpired ? 'Expired' : 'Active'}
                  </span>
                </td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(d)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                  {deleteConfirm === d.id ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => handleDelete(d.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteConfirm(d.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {modal && (
        <Modal title={modal === 'create' ? 'New Delegation' : 'Edit Delegation'} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className={labelClass}>Delegated Prefix</label>
              <input className={inputClass} placeholder="2001:db8::/48" value={form.delegatedPrefix} onChange={e => setForm(f => ({ ...f, delegatedPrefix: e.target.value }))} required />
            </div>
            <div>
              <label className={labelClass}>Description</label>
              <input className={inputClass} placeholder="Customer router" value={form.delegatedToDescription} onChange={e => setForm(f => ({ ...f, delegatedToDescription: e.target.value }))} />
            </div>
            <div>
              <label className={labelClass}>Device ID (optional)</label>
              <input type="number" className={inputClass} placeholder="Device ID" value={form.delegatedToDeviceId} onChange={e => setForm(f => ({ ...f, delegatedToDeviceId: e.target.value }))} />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className={labelClass}>Valid Lifetime (sec)</label>
                <input type="number" className={inputClass} placeholder="3600" value={form.validLifetimeSec} onChange={e => setForm(f => ({ ...f, validLifetimeSec: e.target.value }))} />
              </div>
              <div>
                <label className={labelClass}>Preferred Lifetime (sec)</label>
                <input type="number" className={inputClass} placeholder="1800" value={form.preferredLifetimeSec} onChange={e => setForm(f => ({ ...f, preferredLifetimeSec: e.target.value }))} />
              </div>
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
