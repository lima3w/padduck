import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { createWebhookEndpoint, deleteWebhookEndpoint, getWebhookDeliveries, getWebhookEndpoints, getWebhookFailureGroups, replayWebhookDelivery, updateWebhookEndpoint } from '../api/admin'
import Modal from '../components/Modal'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const EMPTY_FORM = {
  name: '',
  url: '',
  secret: '',
  events: '*',
  objectTypes: '',
  tagFilters: '',
  filterConditions: '',
  isActive: true,
}

function eventList(value) {
  return value
    .split(',')
    .map(v => v.trim())
    .filter(Boolean)
}

function conditionMap(value) {
  return Object.fromEntries(
    value
      .split(',')
      .map(part => part.trim())
      .filter(Boolean)
      .map(part => {
        const [key, ...rest] = part.split('=')
        return [key.trim(), rest.join('=').trim()]
      })
      .filter(([key, val]) => key && val)
  )
}

function deliveryStatusClass(status) {
  if (status === 'delivered') return 'bg-green-100 text-green-700'
  if (status === 'failed') return 'bg-red-100 text-red-700'
  if (status === 'retrying') return 'bg-amber-100 text-amber-700'
  return 'bg-gray-100 text-gray-600'
}

export default function AdminWebhooksPage() {
  const { t } = useTranslation()
  const [endpoints, setEndpoints] = useState([])
  const [deliveries, setDeliveries] = useState([])
  const [failureGroups, setFailureGroups] = useState([])
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
      const [endpointRes, deliveryRes, failureRes] = await Promise.all([
        getWebhookEndpoints(),
        getWebhookDeliveries(25),
        getWebhookFailureGroups(25),
      ])
      setEndpoints(endpointRes.data || [])
      setDeliveries(deliveryRes.data || [])
      setFailureGroups(failureRes.data || [])
    } catch {
      setError(t('adminWebhooks.loadFailed'))
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
      objectTypes: (endpoint.objectTypes || []).join(', '),
      tagFilters: (endpoint.tagFilters || []).join(', '),
      filterConditions: Object.entries(endpoint.filterConditions || {}).map(([k, v]) => `${k}=${v}`).join(', '),
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
      object_types: eventList(form.objectTypes),
      tag_filters: eventList(form.tagFilters),
      filter_conditions: conditionMap(form.filterConditions),
      is_active: form.isActive,
    }
    try {
      if (modal === 'create') {
        await createWebhookEndpoint(payload)
        showMessage(t('adminWebhooks.endpointCreated'))
      } else {
        await updateWebhookEndpoint(modal.edit.id, payload)
        showMessage(t('adminWebhooks.endpointUpdated'))
      }
      setModal(null)
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('adminWebhooks.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteWebhookEndpoint(id)
      setDeleteConfirm(null)
      showMessage(t('adminWebhooks.endpointDeleted'))
      await load()
    } catch {
      setError(t('adminWebhooks.deleteFailed'))
    }
  }

  async function handleReplay(id) {
    try {
      await replayWebhookDelivery(id)
      showMessage(t('adminWebhooks.replayQueued'))
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('adminWebhooks.replayFailed'))
    }
  }

  if (loading) return <PageSpinner message={t('adminWebhooks.loadingWebhooks')} />

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('adminWebhooks.title')}</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t('adminWebhooks.subtitle')}</p>
        </div>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          {t('adminWebhooks.newWebhook')}
        </button>
      </div>

      {message && <div className="mb-4 p-3 bg-green-50 text-green-700 border border-green-200 rounded text-sm">{message}</div>}
      <ErrorBanner error={error} onDismiss={() => setError('')} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden mb-8">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.name')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.urlColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.eventsColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.filtersColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.statusColumn')}</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y dark:divide-gray-700">
            {endpoints.length === 0 ? (
              <EmptyRow colSpan={6} message={t('adminWebhooks.noEndpointsConfigured')} />
            ) : endpoints.map(endpoint => (
              <tr key={endpoint.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/40">
                <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{endpoint.name}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400 font-mono text-xs break-all">{endpoint.url}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{(endpoint.events || []).join(', ') || '*'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                  <div>{(endpoint.objectTypes || []).join(', ') || t('adminWebhooks.allObjects')}</div>
                  <div>{(endpoint.tagFilters || []).join(', ') || t('adminWebhooks.allTags')}</div>
                </td>
                <td className="px-4 py-3">
                  <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${endpoint.isActive ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                    {endpoint.isActive ? t('natRules.active') : t('natRules.disabled')}
                  </span>
                </td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(endpoint)} className="text-xs text-blue-600 hover:underline">{t('common.edit')}</button>
                  {deleteConfirm === endpoint.id ? (
                    <span className="inline-flex items-center gap-1 text-xs">
                      <span className="text-red-600">{t('adminAgents.deleteConfirm')}</span>
                      <button onClick={() => handleDelete(endpoint.id)} className="text-red-600 font-medium hover:text-red-800">{t('common.yes')}</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">{t('common.no')}</button>
                    </span>
                  ) : (
                    <button onClick={() => setDeleteConfirm(endpoint.id)} className="text-xs text-red-600 hover:underline">{t('common.delete')}</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">{t('adminWebhooks.recentDeliveriesTitle')}</h2>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.eventColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.endpointColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.statusColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.responseColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.createdColumn')}</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y dark:divide-gray-700">
            {deliveries.length === 0 ? (
              <EmptyRow colSpan={6} message={t('adminWebhooks.noDeliveriesYet')} />
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
                <td className="px-4 py-3 text-right">
                  <button onClick={() => handleReplay(delivery.id)} className="text-xs text-blue-600 hover:underline">{t('adminWebhooks.replay')}</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mt-8 mb-3">{t('adminWebhooks.failureGroupsTitle')}</h2>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.eventColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.statusColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.countColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.errorColumn')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminWebhooks.lastSeenColumn')}</th>
            </tr>
          </thead>
          <tbody className="divide-y dark:divide-gray-700">
            {failureGroups.length === 0 ? (
              <EmptyRow colSpan={5} message={t('adminWebhooks.noFailedDeliveries')} />
            ) : failureGroups.map(group => (
              <tr key={`${group.endpointId}-${group.eventType}-${group.status}-${group.lastDeliveryId}`}>
                <td className="px-4 py-3 font-mono text-xs text-gray-700 dark:text-gray-300">{group.eventType}</td>
                <td className="px-4 py-3">
                  <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${deliveryStatusClass(group.status)}`}>
                    {group.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{group.count}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{group.errorMsg || '-'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{group.lastOccurredAt ? new Date(group.lastOccurredAt).toLocaleString() : '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {modal && (
        <Modal title={modal === 'create' ? t('adminWebhooks.newWebhookModalTitle') : t('adminWebhooks.editWebhookModalTitle')} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('common.name')}</label>
              <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminWebhooks.url')}</label>
              <input value={form.url} onChange={e => setForm({ ...form, url: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminWebhooks.signingSecret')}</label>
              <input value={form.secret} onChange={e => setForm({ ...form, secret: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" type="password" autoComplete="new-password" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminWebhooks.events')}</label>
              <input value={form.events} onChange={e => setForm({ ...form, events: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminWebhooks.objectTypes')}</label>
              <input value={form.objectTypes} onChange={e => setForm({ ...form, objectTypes: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" placeholder="ip_address, subnet, device" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminWebhooks.tagFilters')}</label>
              <input value={form.tagFilters} onChange={e => setForm({ ...form, tagFilters: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" placeholder="production, 12" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('adminWebhooks.conditions')}</label>
              <input value={form.filterConditions} onChange={e => setForm({ ...form, filterConditions: e.target.value })} className="w-full border rounded px-3 py-2 text-sm" placeholder="status=assigned, resource_name=web-*" />
            </div>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={form.isActive} onChange={e => setForm({ ...form, isActive: e.target.checked })} />
              {t('adminWebhooks.active')}
            </label>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-3 py-2 text-sm border rounded hover:bg-gray-50">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-3 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('common.save')}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
