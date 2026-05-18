import { useEffect, useState } from 'react'
import {
  createWebhookEndpoint,
  deleteWebhookEndpoint,
  getWebhookDeliveries,
  getWebhookEndpoints,
  updateWebhookEndpoint,
} from '../api/client'
import Modal from '../components/Modal'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const EMPTY_FORM = { name: '', url: '', secret: '', events: '*', isActive: true }

function eventList(value) {
  return value
    .split(',')
    .map(v => v.trim())
    .filter(Boolean)
}

function deliveryStatusClass(status) {
  if (status === 'delivered') return 'bg-green-100 text-green-700'
  if (status === 'failed') return 'bg-red-100 text-red-700'
  if (status === 'retrying') return 'bg-amber-100 text-amber-700'
  return 'bg-gray-100 text-gray-600'
}

export default function AdminWebhooksPage() {
  const [endpoints, setEndpoints] = useState([])
  const [deliveries, setDeliveries] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      const [endpointRes, deliveryRes] = await Promise.all([
        getWebhookEndpoints(),
        getWebhookDeliveries(25),
      ])
      setEndpoints(endpointRes.data || [])
      setDeliveries(deliveryRes.data || [])
    } catch {
      setError('Failed to load webhooks')
    } finally {
      setLoading(false)
    }
  }

  function showMessage(text) {
    setMessage(text)
    setTimeout(() => setMessage(''), 3000)
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(endpoint) {
    setForm({
      name: endpoint.name || '',
      url: endpoint.url || '',
      secret: '',
      events: (endpoint.events || []).join(', ') || '*',
      isActive: endpoint.isActive !== false,
    })
    setModal({ edit: endpoint })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    setError('')
    const payload = {
      name: form.name,
      url: form.url,
      secret: form.secret,
      events: eventList(form.events),
      is_active: form.isActive,
    }
    try {
      if (modal === 'create') {
        await createWebhookEndpoint(payload)
        showMessage('Webhook endpoint created')
      } else {
        await updateWebhookEndpoint(modal.edit.id, payload)
        showMessage('Webhook endpoint updated')
      }
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save webhook endpoint')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteWebhookEndpoint(id)
      setDeleteConfirm(null)
      showMessage('Webhook endpoint deleted')
      await load()
    } catch {
      setError('Failed to delete webhook endpoint')
    }
  }

  if (loading) return <PageSpinner message="Loading webhooks..." />

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Webhooks</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Deliver IPAM audit events to external systems.</p>
        </div>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          New Webhook
        </button>
      </div>

      {message && <div className="mb-4 p-3 bg-green-50 text-green-700 border border-green-200 rounded text-sm">{message}</div>}
      <ErrorBanner error={error} onDismiss={() => setError('')} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden mb-8">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">URL</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Events</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y dark:divide-gray-700">
            {endpoints.length === 0 ? (
              <EmptyRow colSpan={5} message="No webhook endpoints configured." />
            ) : endpoints.map(endpoint => (
              <tr key={endpoint.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/40">
                <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{endpoint.name}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400 font-mono text-xs break-all">{endpoint.url}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{(endpoint.events || []).join(', ') || '*'}</td>
                <td className="px-4 py-3">
                  <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${endpoint.isActive ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                    {endpoint.isActive ? 'Active' : 'Disabled'}
                  </span>
                </td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(endpoint)} className="text-xs text-blue-600 hover:underline">Edit</button>
                  {deleteConfirm === endpoint.id ? (
                    <span className="inline-flex items-center gap-1 text-xs">
                      <span className="text-red-600">Delete?</span>
                      <button onClick={() => handleDelete(endpoint.id)} className="text-red-600 font-medium hover:text-red-800">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">No</button>
                    </span>
                  ) : (
                    <button onClick={() => setDeleteConfirm(endpoint.id)} className="text-xs text-red-600 hover:underline">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Recent Deliveries</h2>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Event</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Endpoint</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Response</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y dark:divide-gray-700">
            {deliveries.length === 0 ? (
              <EmptyRow colSpan={5} message="No deliveries yet." />
            ) : deliveries.map(delivery => (
              <tr key={delivery.id}>
                <td className="px-4 py-3 font-mono text-xs text-gray-700 dark:text-gray-300">{delivery.eventType}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{delivery.endpointId}</td>
                <td className="px-4 py-3">
                  <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${deliveryStatusClass(delivery.status)}`}>
                    {delivery.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{delivery.responseStatus || delivery.errorMsg || '-'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{delivery.createdAt ? new Date(delivery.createdAt).toLocaleString() : '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <Modal title={modal === 'create' ? 'New Webhook' : 'Edit Webhook'} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">URL</label>
              <input value={form.url} onChange={e => setForm({ ...form, url: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Signing Secret</label>
              <input value={form.secret} onChange={e => setForm({ ...form, secret: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" type="password" autoComplete="new-password" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Events</label>
              <input value={form.events} onChange={e => setForm({ ...form, events: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" />
            </div>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={form.isActive} onChange={e => setForm({ ...form, isActive: e.target.checked })} />
              Active
            </label>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-3 py-2 text-sm border rounded hover:bg-gray-50">Cancel</button>
              <button type="submit" disabled={saving} className="px-3 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
