import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import {
  getSecrets, createSecret, rotateSecret, deleteSecret, getSecretVersions, deleteSecretVersion,
} from '../api/admin'
import Modal from '../components/Modal'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const EMPTY_FORM = { name: '', description: '', value: '' }

function formatDate(v) {
  return v ? new Date(v).toLocaleString() : '—'
}

export default function AdminSecretsPage() {
  const { t } = useTranslation()
  const [secrets, setSecrets] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [modal, setModal] = useState(null) // null | 'create' | { rotate: secret } | { versions: secret }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [versions, setVersions] = useState([])
  const [versionDeleteConfirm, setVersionDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      const { data } = await getSecrets()
      setSecrets(data || [])
    } catch {
      setError(t('adminSecrets.loadFailed'))
    } finally {
      setLoading(false)
    }
  }

  function flash(text) {
    setMessage(text)
    setTimeout(() => setMessage(''), 3000)
  }

  function closeModal() {
    setModal(null)
    setForm(EMPTY_FORM)
    setVersions([])
    setVersionDeleteConfirm(null)
  }

  async function handleCreate(e) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      await createSecret({ name: form.name, description: form.description, value: form.value })
      closeModal()
      flash(t('adminSecrets.created'))
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('adminSecrets.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleRotate(e) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      await rotateSecret(modal.rotate.id, form.value)
      closeModal()
      flash(t('adminSecrets.rotated'))
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('adminSecrets.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function openVersions(secret) {
    setModal({ versions: secret })
    try {
      const { data } = await getSecretVersions(secret.id)
      setVersions(data || [])
    } catch {
      setError(t('adminSecrets.versionsFailed'))
    }
  }

  async function handleDeleteVersion(secret, version) {
    try {
      await deleteSecretVersion(secret.id, version)
      setVersionDeleteConfirm(null)
      flash(t('adminSecrets.versionDeleted'))
      const { data } = await getSecretVersions(secret.id)
      setVersions(data || [])
    } catch (err) {
      setError(err.response?.data?.error || t('adminSecrets.deleteFailed'))
    }
  }

  async function handleDelete(id) {
    try {
      await deleteSecret(id)
      setDeleteConfirm(null)
      flash(t('adminSecrets.deleted'))
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('adminSecrets.deleteFailed'))
    }
  }

  async function copyReference(name) {
    try {
      await navigator.clipboard.writeText(`secret://${name}`)
      flash(t('adminSecrets.copied'))
    } catch {
      /* clipboard unavailable; the reference is visible in the table */
    }
  }

  if (loading) return <PageSpinner message={t('adminSecrets.loading')} />

  const inputClass = 'w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500'

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="flex items-center justify-between mb-2">
        <h1 className="text-2xl font-bold text-gray-900">{t('adminSecrets.title')}</h1>
        <button
          onClick={() => { setForm(EMPTY_FORM); setModal('create') }}
          className="text-sm bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 transition"
        >
          {t('adminSecrets.newSecret')}
        </button>
      </div>
      <p className="text-sm text-gray-600 mb-6">
        {t('adminSecrets.intro')} <code className="font-mono text-xs bg-gray-100 px-1 py-0.5 rounded">secret://name</code>.
      </p>

      <ErrorBanner error={error} />
      {message && (
        <div role="status" className="mb-4 p-3 bg-green-50 border border-green-200 text-green-800 rounded text-sm">{message}</div>
      )}

      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('adminSecrets.nameColumn')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('adminSecrets.reference')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('adminSecrets.versionColumn')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('adminSecrets.rotatedColumn')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('vrfs.actions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {secrets.length === 0 ? (
                <EmptyRow colSpan={5} message={t('adminSecrets.noSecrets')} />
              ) : (
                secrets.map((secret) => (
                  <tr key={secret.id} className="hover:bg-gray-50">
                    <td className="px-4 py-2 font-medium text-gray-900">
                      {secret.name}
                      {secret.description && (
                        <p className="text-xs text-gray-500 font-normal mt-0.5">{secret.description}</p>
                      )}
                    </td>
                    <td className="px-4 py-2">
                      <button
                        onClick={() => copyReference(secret.name)}
                        title={t('adminSecrets.copyReference')}
                        aria-label={`${t('adminSecrets.copyReference')}: secret://${secret.name}`}
                        className="font-mono text-xs text-gray-700 hover:text-blue-600"
                      >
                        secret://{secret.name}
                      </button>
                    </td>
                    <td className="px-4 py-2 text-gray-700">v{secret.version}</td>
                    <td className="px-4 py-2 text-gray-700">{formatDate(secret.rotatedAt || secret.createdAt)}</td>
                    <td className="px-4 py-2">
                      <div className="flex items-center gap-3">
                        <button
                          onClick={() => { setForm(EMPTY_FORM); setModal({ rotate: secret }) }}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          {t('adminSecrets.rotate')}
                        </button>
                        <button onClick={() => openVersions(secret)} className="text-xs text-blue-600 hover:underline">
                          {t('adminSecrets.versions')}
                        </button>
                        {deleteConfirm === secret.id ? (
                          <span className="flex items-center gap-1 text-xs">
                            <span className="text-red-600">{t('adminSecrets.deleteConfirm')}</span>
                            <button onClick={() => handleDelete(secret.id)} className="text-red-600 font-medium hover:text-red-800">{t('common.yes')}</button>
                            <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">{t('common.no')}</button>
                          </span>
                        ) : (
                          <button onClick={() => setDeleteConfirm(secret.id)} className="text-xs text-red-500 hover:underline">
                            {t('common.delete')}
                          </button>
                        )}
                      </div>
                      {deleteConfirm === secret.id && (
                        <p className="text-xs text-red-600 mt-1">{t('adminSecrets.deleteWarning')}</p>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {modal === 'create' && (
        <Modal title={t('adminSecrets.createTitle')} onClose={closeModal}>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label htmlFor="secret-name" className="block text-sm font-medium text-gray-700 mb-1">
                {t('adminSecrets.nameLabel')} <span className="text-red-500">*</span>
              </label>
              <input
                id="secret-name" required className={inputClass} value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                pattern="[A-Za-z0-9][A-Za-z0-9._\-]{0,127}" maxLength={128} autoComplete="off"
              />
              <p className="text-xs text-gray-500 mt-1">{t('adminSecrets.nameHelp')}</p>
            </div>
            <div>
              <label htmlFor="secret-description" className="block text-sm font-medium text-gray-700 mb-1">{t('adminSecrets.descriptionLabel')}</label>
              <input
                id="secret-description" className={inputClass} value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div>
              <label htmlFor="secret-value" className="block text-sm font-medium text-gray-700 mb-1">
                {t('adminSecrets.valueLabel')} <span className="text-red-500">*</span>
              </label>
              <input
                id="secret-value" type="password" required className={inputClass} value={form.value}
                onChange={e => setForm(f => ({ ...f, value: e.target.value }))} autoComplete="new-password"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button type="button" onClick={closeModal} className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {t('adminSecrets.create')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal?.rotate && (
        <Modal title={`${t('adminSecrets.rotateTitle')}: ${modal.rotate.name}`} onClose={closeModal}>
          <form onSubmit={handleRotate} className="space-y-4">
            <div>
              <label htmlFor="secret-new-value" className="block text-sm font-medium text-gray-700 mb-1">
                {t('adminSecrets.newValueLabel')} <span className="text-red-500">*</span>
              </label>
              <input
                id="secret-new-value" type="password" required className={inputClass} value={form.value}
                onChange={e => setForm(f => ({ ...f, value: e.target.value }))} autoComplete="new-password"
              />
              <p className="text-xs text-gray-500 mt-1">{t('adminSecrets.rotateHelp')}</p>
            </div>
            <div className="flex justify-end gap-2">
              <button type="button" onClick={closeModal} className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {t('adminSecrets.rotate')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal?.versions && (
        <Modal title={`${t('adminSecrets.versionsTitle')}: ${modal.versions.name}`} onClose={closeModal}>
          <table className="min-w-full text-sm">
            <tbody className="divide-y divide-gray-100">
              {versions.map((v) => (
                <tr key={v.version}>
                  <td className="py-2 pr-4 font-medium text-gray-900">v{v.version}</td>
                  <td className="py-2 pr-4 text-gray-600">{formatDate(v.createdAt)}</td>
                  <td className="py-2 pr-4">
                    {v.current && (
                      <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-700">{t('adminSecrets.current')}</span>
                    )}
                  </td>
                  <td className="py-2 text-right">
                    {!v.current && (versionDeleteConfirm === v.version ? (
                      <span className="flex items-center justify-end gap-1 text-xs">
                        <span className="text-red-600">{t('adminSecrets.deleteConfirm')}</span>
                        <button onClick={() => handleDeleteVersion(modal.versions, v.version)} className="text-red-600 font-medium hover:text-red-800">{t('common.yes')}</button>
                        <button onClick={() => setVersionDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">{t('common.no')}</button>
                      </span>
                    ) : (
                      <button
                        onClick={() => setVersionDeleteConfirm(v.version)}
                        aria-label={`${t('adminSecrets.deleteVersion')} v${v.version}`}
                        className="text-xs text-red-500 hover:underline"
                      >
                        {t('adminSecrets.deleteVersion')}
                      </button>
                    ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Modal>
      )}
    </div>
  )
}
