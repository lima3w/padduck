import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import Modal from '../components/Modal'
import { getVlanDomains, createVlanDomain, updateVlanDomain, deleteVlanDomain } from '../api/vlans'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import TableActions from '../components/TableActions'
import { getCachedUser } from '../utils/storageKeys'

const EMPTY_FORM = { name: '', description: '' }

export default function VlanDomainsPage() {
  const { t } = useTranslation()
  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  const [domains, setDomains] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: domain }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const res = await getVlanDomains()
      const data = res.data
      setDomains(Array.isArray(data) ? data : (data?.domains ?? []))
    } catch (err) {
      setError(err.response?.data?.error || t('vlanDomains.loadError'))
    } finally {
      setLoading(false)
    }
  }

  function showMsg(text, type = 'success') {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 3000)
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(domain) {
    setForm({ name: domain.name || '', description: domain.description || '' })
    setModal({ edit: domain })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const payload = {
        name: form.name,
        description: form.description || null,
      }
      if (modal === 'create') {
        await createVlanDomain(payload)
        showMsg(t('vlanDomains.created'))
      } else {
        await updateVlanDomain(modal.edit.id, payload)
        showMsg(t('vlanDomains.updated'))
      }
      setModal(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('vlanDomains.saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteVlanDomain(id)
      setDeleteConfirm(null)
      showMsg(t('vlanDomains.deleted'))
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('vlanDomains.deleteError'))
      setDeleteConfirm(null)
    }
  }

  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-red-600 font-semibold text-lg mb-2">{t('vlanDomains.accessDenied')}</p>
          <p className="text-gray-500 text-sm">{t('vlanDomains.adminRequired')}</p>
        </div>
      </div>
    )
  }

  if (loading) return <PageSpinner message={t('vlanDomains.loading')} />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('vlanDomains.title')}</h1>
        <button
          onClick={openCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
        >
          {t('vlanDomains.newDomain')}
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      <ErrorBanner error={error} onDismiss={() => setError(null)} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.name')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.description')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('nav.vlans')}</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {domains.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-gray-400">
                  {t('vlanDomains.noDomainsYet')}
                </td>
              </tr>
            )}
            {domains.map(domain => (
              <tr key={domain.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{domain.name}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{domain.description || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
                    {domain.vlanCount ?? 0}
                  </span>
                </td>
                <td className="px-4 py-3 text-right">
                  <TableActions
                    onEdit={() => openEdit(domain)}
                    onDelete={() => handleDelete(domain.id)}
                    confirming={deleteConfirm === domain.id}
                    onRequestDelete={() => setDeleteConfirm(domain.id)}
                    onCancelDelete={() => setDeleteConfirm(null)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {modal && (
        <Modal
          title={modal === 'create' ? t('vlanDomains.newDomainModalTitle') : t('vlanDomains.editDomainModalTitle')}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('common.name')} <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. Campus LAN"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.description')}</label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={2}
                placeholder="Optional description"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setModal(null)}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                {t('common.cancel')}
              </button>
              <button
                type="submit"
                disabled={saving}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? t('common.saving') : t('common.save')}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
