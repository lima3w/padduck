import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import Modal from '../components/Modal'
import { getCustomFields, createCustomField, updateCustomField, deleteCustomField, reorderCustomFields } from '../api/admin'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const ENTITY_TYPES = ['subnet', 'ip_address', 'device']
const ENTITY_LABEL_KEYS = { subnet: 'entityLabelSubnets', ip_address: 'entityLabelIpAddresses', device: 'entityLabelDevices' }
const FIELD_TYPES = ['text', 'number', 'textarea', 'dropdown', 'checkbox', 'date', 'url', 'email']

const EMPTY_FORM = {
  entity_type: 'subnet',
  name: '',
  label: '',
  field_type: 'text',
  options: [],
  is_required: false,
  default_value: '',
  placeholder: '',
  is_searchable: false,
}

export default function AdminCustomFieldsPage() {
  const { t } = useTranslation()
  const ENTITY_LABELS = Object.fromEntries(ENTITY_TYPES.map(et => [et, t(`adminCustomFields.${ENTITY_LABEL_KEYS[et]}`)]))
  const [fields, setFields] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [activeTab, setActiveTab] = useState('subnet')
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)
  const [newOption, setNewOption] = useState({ value: '', label: '' })

  useEffect(() => {
    load()
  }, [])

  async function load() {
    try {
      setLoading(true)
      const res = await getCustomFields()
      setFields(Array.isArray(res.data) ? res.data : [])
    } catch {
      setError(t('adminCustomFields.loadFailed'))
    } finally {
      setLoading(false)
    }
  }

  function fieldsForTab(entityType) {
    return fields.filter(f => f.entityType === entityType).sort((a, b) => a.sortOrder - b.sortOrder)
  }

  function openCreate() {
    setForm({ ...EMPTY_FORM, entity_type: activeTab })
    setNewOption({ value: '', label: '' })
    setModal('create')
  }

  function openEdit(field) {
    setForm({
      entity_type: field.entityType,
      name: field.name || '',
      label: field.label || '',
      field_type: field.fieldType || 'text',
      options: field.options ? JSON.parse(JSON.stringify(field.options)) : [],
      is_required: field.isRequired || false,
      default_value: field.defaultValue || '',
      placeholder: field.placeholder || '',
      is_searchable: field.isSearchable || false,
    })
    setNewOption({ value: '', label: '' })
    setModal({ edit: field })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        entity_type: form.entity_type,
        name: form.name,
        label: form.label,
        field_type: form.field_type,
        options: form.field_type === 'dropdown' ? form.options : [],
        is_required: form.is_required,
        default_value: form.default_value || null,
        placeholder: form.placeholder || null,
        is_searchable: form.is_searchable,
      }
      if (modal === 'create') {
        await createCustomField(body)
      } else {
        await updateCustomField(modal.edit.id, body)
      }
      setModal(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || err.message || t('adminCustomFields.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteCustomField(id)
      setDeleteConfirm(null)
      load()
    } catch {
      setError(t('adminCustomFields.deleteFailed'))
    }
  }

  async function handleReorder(entityType, ids) {
    try {
      await reorderCustomFields(ids)
      load()
    } catch {
      setError(t('adminCustomFields.reorderFailed'))
    }
  }

  function moveField(entityType, index, direction) {
    const list = fieldsForTab(entityType)
    const newIndex = index + direction
    if (newIndex < 0 || newIndex >= list.length) return
    const newList = [...list]
    ;[newList[index], newList[newIndex]] = [newList[newIndex], newList[index]]
    handleReorder(entityType, newList.map(f => f.id))
  }

  function addOption() {
    if (!newOption.value.trim()) return
    setForm(f => ({
      ...f,
      options: [...(f.options || []), { value: newOption.value.trim(), label: newOption.label.trim() || newOption.value.trim() }],
    }))
    setNewOption({ value: '', label: '' })
  }

  function removeOption(idx) {
    setForm(f => ({ ...f, options: f.options.filter((_, i) => i !== idx) }))
  }

  if (loading) return <PageSpinner message={t('adminCustomFields.loadingFields')} />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('adminCustomFields.title')}</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          {t('adminCustomFields.newField')}
        </button>
      </div>

      <ErrorBanner error={error} />

      <div className="flex gap-1 mb-4 border-b border-gray-200 dark:border-gray-700">
        {ENTITY_TYPES.map(et => (
          <button
            key={et}
            onClick={() => setActiveTab(et)}
            className={`px-4 py-2 text-sm font-medium rounded-t transition ${
              activeTab === et
                ? 'bg-white dark:bg-gray-800 border border-b-white dark:border-gray-600 dark:border-b-gray-800 text-blue-600 -mb-px'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
            }`}
          >
            {ENTITY_LABELS[et]}
          </button>
        ))}
      </div>

      {ENTITY_TYPES.map(et => et === activeTab && (
        <div key={et} className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium w-10">{t('adminCustomFields.orderColumn')}</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminCustomFields.labelColumn')}</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminCustomFields.typeColumn')}</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminCustomFields.requiredColumn')}</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('adminCustomFields.searchableColumn')}</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {fieldsForTab(et).length === 0 && (
                <EmptyRow colSpan={6} message={t('adminCustomFields.noFieldsFor', { entityLabel: ENTITY_LABELS[et] })} />
              )}
              {fieldsForTab(et).map((field, idx, list) => (
                <tr key={field.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td className="px-4 py-3">
                    <div className="flex flex-col gap-0.5">
                      <button
                        onClick={() => moveField(et, idx, -1)}
                        disabled={idx === 0}
                        className="text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 disabled:opacity-25 text-xs leading-none"
                        title={t('adminCustomFields.moveUp')}
                      >▲</button>
                      <button
                        onClick={() => moveField(et, idx, 1)}
                        disabled={idx === list.length - 1}
                        className="text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 disabled:opacity-25 text-xs leading-none"
                        title={t('adminCustomFields.moveDown')}
                      >▼</button>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-gray-700 dark:text-gray-300">{field.label}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{field.fieldType}</td>
                  <td className="px-4 py-3">
                    {field.isRequired ? (
                      <span className="inline-block px-2 py-0.5 bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 text-xs rounded">{t('common.yes')}</span>
                    ) : (
                      <span className="text-gray-400 text-xs">{t('common.no')}</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    {field.isSearchable ? (
                      <span className="inline-block px-2 py-0.5 bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 text-xs rounded">{t('common.yes')}</span>
                    ) : (
                      <span className="text-gray-400 text-xs">{t('common.no')}</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right space-x-2">
                    <button onClick={() => openEdit(field)} className="text-gray-400 hover:text-blue-600 text-xs">{t('common.edit')}</button>
                    {deleteConfirm === field.id ? (
                      <>
                        <span className="text-red-600 text-xs">{t('subnets.confirmDelete')}</span>
                        <button onClick={() => handleDelete(field.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">{t('common.yes')}</button>
                        <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">{t('common.no')}</button>
                      </>
                    ) : (
                      <button onClick={() => setDeleteConfirm(field.id)} className="text-gray-400 hover:text-red-600 text-xs">{t('common.delete')}</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          </div>
        </div>
      ))}

      {modal && (
        <Modal
          title={modal === 'create' ? t('adminCustomFields.newFieldModalTitle') : `${t('adminCustomFields.editFieldModalTitlePrefix')}${modal.edit.label}`}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4 max-h-[70vh] overflow-y-auto pr-1">
            {modal === 'create' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminCustomFields.entityType')} <span className="text-red-500">*</span></label>
                <select
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.entity_type}
                  onChange={e => setForm(f => ({ ...f, entity_type: e.target.value }))}
                >
                  {ENTITY_TYPES.map(et => <option key={et} value={et}>{ENTITY_LABELS[et]}</option>)}
                </select>
              </div>
            )}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.name')} <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="circuit_id"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
                disabled={modal !== 'create'}
              />
              {modal === 'create' && <p className="text-xs text-gray-500 mt-1">{t('adminCustomFields.nameHint')}</p>}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminCustomFields.label')} <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="Circuit ID"
                value={form.label}
                onChange={e => setForm(f => ({ ...f, label: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminCustomFields.fieldType')} <span className="text-red-500">*</span></label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={form.field_type}
                onChange={e => setForm(f => ({ ...f, field_type: e.target.value }))}
              >
                {FIELD_TYPES.map(ft => <option key={ft} value={ft}>{ft}</option>)}
              </select>
            </div>
            {form.field_type === 'dropdown' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('adminCustomFields.options')}</label>
                <div className="space-y-1 mb-2">
                  {(form.options || []).map((opt, idx) => (
                    <div key={idx} className="flex items-center gap-2 text-sm">
                      <span className="font-mono text-gray-600 dark:text-gray-400 flex-1">{opt.value}</span>
                      <span className="text-gray-500 dark:text-gray-400 flex-1">{opt.label}</span>
                      <button type="button" onClick={() => removeOption(idx)} className="text-gray-400 hover:text-red-600 text-xs">{t('adminCustomFields.remove')}</button>
                    </div>
                  ))}
                </div>
                <div className="flex gap-2">
                  <input
                    className="flex-1 border rounded px-2 py-1.5 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder={t('adminCustomFields.valuePlaceholder')}
                    value={newOption.value}
                    onChange={e => setNewOption(o => ({ ...o, value: e.target.value }))}
                  />
                  <input
                    className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder={t('adminCustomFields.labelPlaceholder')}
                    value={newOption.label}
                    onChange={e => setNewOption(o => ({ ...o, label: e.target.value }))}
                  />
                  <button type="button" onClick={addOption} className="px-3 py-1.5 bg-gray-100 dark:bg-gray-600 text-gray-700 dark:text-gray-200 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-500">{t('adminCustomFields.add')}</button>
                </div>
              </div>
            )}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminCustomFields.defaultValue')}</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.default_value}
                  onChange={e => setForm(f => ({ ...f, default_value: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('adminCustomFields.placeholder')}</label>
                <input
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.placeholder}
                  onChange={e => setForm(f => ({ ...f, placeholder: e.target.value }))}
                />
              </div>
            </div>
            <div className="space-y-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.is_required}
                  onChange={e => setForm(f => ({ ...f, is_required: e.target.checked }))}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">{t('adminCustomFields.required')}</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.is_searchable}
                  onChange={e => setForm(f => ({ ...f, is_searchable: e.target.checked }))}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">{t('adminCustomFields.searchable')}</span>
              </label>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('common.save')}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
