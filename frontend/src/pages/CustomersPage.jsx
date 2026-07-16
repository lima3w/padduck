import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
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

const OBJECT_TYPE_VALUES = [
  'network', 'subnet', 'ip_address', 'device', 'rack', 'location',
  'vlan', 'vrf', 'nat_rule', 'dhcp_server', 'physical_circuit', 'logical_circuit',
]

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

function ObjectPicker({ type, _value, label, onChange }) {
  const { t } = useTranslation()
  const [items, setItems] = useState(null)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState(label || '')
  const [open, setOpen] = useState(false)
  const [searchResults, setSearchResults] = useState([])
  const searchTimer = useRef(null)
  const inputRef = useRef(null)

  const isSearchType = type === 'subnet' || type === 'ip_address'

  useEffect(() => {
    setQuery('')
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
        placeholder={loading ? t('customers.picker.loading') : isSearchType ? t('customers.picker.typeToSearch') : t('customers.picker.searchOrSelect')}
        className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      {open && (
        <div className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded shadow-lg max-h-52 overflow-y-auto">
          {loading && <div className="px-3 py-2 text-sm text-gray-400">{t('customers.picker.loading')}</div>}
          {!loading && filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-400">
              {isSearchType && query.trim().length < 2 ? t('customers.picker.typeAtLeastTwoChars') : t('customers.picker.noMatches')}
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
  const { t } = useTranslation()
  const OBJECT_TYPE_LABEL = Object.fromEntries(OBJECT_TYPE_VALUES.map(v => [v, t(`customers.objectTypes.${v}`)]))
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
  const [activeTab, setActiveTab] = useState('customers')

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      const [res, assocRes] = await Promise.all([getCustomers(), getCustomerAssociations()])
      setCustomers(res.data || [])
      setAssociations(assocRes.data || [])
    } catch {
      setError(t('customers.loadError'))
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
        setMessage(t('customers.updated'))
      } else {
        await createCustomer(form)
        setMessage(t('customers.created'))
      }
      closeModal()
      await load()
    } catch (err) {
      setError(err.response?.data?.error || t('customers.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteCustomer(id)
      setDeleteConfirm(null)
      setMessage(t('customers.deleted'))
      await load()
    } catch {
      setError(t('customers.deleteFailed'))
    }
  }

  async function handleAssociationSave(e) {
    e.preventDefault()
    if (!assocForm.customer_id) { setError(t('customers.selectCustomer')); return }
    if (!assocForm.object_id) { setError(t('customers.selectObject')); return }
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
      setError(err.response?.data?.error || t('customers.saveAssociationFailed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('nav.customers')}</h1>
        {activeTab === 'customers' && (
          <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
            {t('customers.newCustomer')}
          </button>
        )}
        {activeTab === 'associations' && (
          <button onClick={openAssoc} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
            {t('customers.addAssociation')}
          </button>
        )}
      </div>

      <div className="flex border-b border-gray-200 dark:border-gray-700 mb-5">
        {['customers', 'associations'].map(tab => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              activeTab === tab
                ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
            }`}
          >
            {tab === 'customers' ? t('nav.customers') : t('customers.associationsTab')}
          </button>
        ))}
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

      {activeTab === 'customers' && (
        loading ? (
          <div className="text-gray-500 text-sm">{t('common.loading')}</div>
        ) : customers.length === 0 ? (
          <div className="text-gray-500 text-sm">{t('customers.noCustomersYet')}</div>
        ) : (
          <div className="overflow-x-auto rounded border border-gray-200 dark:border-gray-700">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('common.name')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('customers.email')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('customers.phone')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('common.description')}</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-700 bg-white dark:bg-gray-800">
                {customers.map(c => (
                  <tr key={c.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{c.name}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{c.email || '—'}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{c.phone || '—'}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 max-w-xs truncate">{c.description || '—'}</td>
                    <td className="px-4 py-3 text-right space-x-2 whitespace-nowrap">
                      <button onClick={() => openEdit(c)} className="text-blue-600 hover:underline text-xs">{t('common.edit')}</button>
                      <button onClick={() => setDeleteConfirm(c)} className="text-red-600 hover:underline text-xs">{t('common.delete')}</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}

      {activeTab === 'associations' && (
        <div className="overflow-x-auto rounded border border-gray-200 dark:border-gray-700">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('customers.customer')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('natRules.type')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('customers.object')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">{t('customers.relationship')}</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700 bg-white dark:bg-gray-800">
              {associations.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-500 dark:text-gray-400">{t('customers.noAssociationsYet')}</td></tr>
              )}
              {associations.map(a => (
                <tr key={a.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{a.customerName || `#${a.customerId}`}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{OBJECT_TYPE_LABEL[a.objectType] || a.objectType}</td>
                  <td className="px-4 py-3 text-gray-700 dark:text-gray-300">{a.objectName || `#${a.objectId}`}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400 capitalize">{a.relationship}</td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={async () => { await deleteCustomerAssociation(a.id); load() }}
                      className="text-red-600 hover:underline text-xs"
                    >
                      {t('customers.remove')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {modal === 'assoc' && (
        <Modal onClose={closeModal}>
          <h2 className="text-lg font-semibold mb-5">{t('customers.addAssociationModalTitle')}</h2>
          <form onSubmit={handleAssociationSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.customer')}</label>
              <select
                required
                value={assocForm.customer_id}
                onChange={e => setAssocForm(f => ({ ...f, customer_id: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">{t('customers.selectCustomerPlaceholder')}</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.objectType')}</label>
              <select
                value={assocForm.object_type}
                onChange={e => setAssocForm(f => ({ ...f, object_type: e.target.value, object_id: '', object_name: '' }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {OBJECT_TYPE_VALUES.map(v => <option key={v} value={v}>{OBJECT_TYPE_LABEL[v]}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {OBJECT_TYPE_LABEL[assocForm.object_type] || t('customers.objectGeneric')}
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
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.relationship')}</label>
              <select
                value={assocForm.relationship}
                onChange={e => setAssocForm(f => ({ ...f, relationship: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {RELATIONSHIPS.map(r => <option key={r} value={r}>{t(`customers.relationships.${r}`)}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.notes')}</label>
              <input
                type="text"
                value={assocForm.notes}
                onChange={e => setAssocForm(f => ({ ...f, notes: e.target.value }))}
                placeholder={t('common.optional')}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            {error && <p className="text-red-600 text-sm">{error}</p>}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={closeModal} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">
                {t('common.cancel')}
              </button>
              <button type="submit" disabled={saving} className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('customers.add')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {(modal === 'create' || modal === 'edit') && (
        <Modal onClose={closeModal}>
          <h2 className="text-lg font-semibold mb-4">{modal === 'edit' ? t('customers.editModalTitle') : t('customers.newModalTitle')}</h2>
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.nameRequired')}</label>
              <input
                type="text"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.email')}</label>
              <input
                type="email"
                value={form.email}
                onChange={e => setForm(f => ({ ...f, email: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.phone')}</label>
              <input
                type="text"
                value={form.phone}
                onChange={e => setForm(f => ({ ...f, phone: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('common.description')}</label>
              <input
                type="text"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('customers.notes')}</label>
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
                {t('common.cancel')}
              </button>
              <button type="submit" disabled={saving} className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('common.save')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {deleteConfirm && (
        <Modal onClose={() => setDeleteConfirm(null)}>
          <h2 className="text-lg font-semibold mb-2">{t('customers.deleteModalTitle')}</h2>
          <p className="text-sm text-gray-600 mb-4">
            {t('customers.confirmDeletePrefix')}<strong>{deleteConfirm.name}</strong>{t('customers.confirmDeleteSuffix')}
          </p>
          <div className="flex justify-end gap-2">
            <button onClick={() => setDeleteConfirm(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">
              {t('common.cancel')}
            </button>
            <button onClick={() => handleDelete(deleteConfirm.id)} className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700">
              {t('common.delete')}
            </button>
          </div>
        </Modal>
      )}
    </div>
  )
}
