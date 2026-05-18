import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import Modal from '../components/Modal'
import {
  getVlans,
  createVlan,
  updateVlan,
  deleteVlan,
  getVlanDomains,
  getVlanGroups,
} from '../api/client'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const EMPTY_FORM = { vlanId: '', name: '', description: '', domainId: '', groupId: '' }

// VLAN model has no JSON tags — Go field names come through as-is (PascalCase):
//   ID, VlanID, DomainID, GroupID, VRFID, Name, Description, CreatedAt, UpdatedAt
// VLANDomain has json tags → snake_case → camelCase interceptor: id, name, description, createdAt
// VLANGroup has json tags → snake_case → camelCase interceptor: id, name, colour, description, createdAt

function GroupBadge({ group }) {
  if (!group) return <span className="text-gray-400 text-xs">—</span>
  const hex = (group.colour || '#6B7280').replace('#', '')
  const r = parseInt(hex.substring(0, 2), 16)
  const g = parseInt(hex.substring(2, 4), 16)
  const b = parseInt(hex.substring(4, 6), 16)
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255
  const textClass = luminance > 0.6 ? 'text-gray-800' : 'text-white'
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${textClass}`}
      style={{ backgroundColor: group.colour || '#6B7280' }}
    >
      {group.name}
    </span>
  )
}

export default function VlansPage() {
  const [vlans, setVlans] = useState([])
  const [domains, setDomains] = useState([])
  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: vlan }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const [vlansRes, domainsRes, groupsRes] = await Promise.allSettled([
        getVlans(),
        getVlanDomains(),
        getVlanGroups(),
      ])
      if (vlansRes.status === 'fulfilled') {
        const d = vlansRes.value.data
        setVlans(Array.isArray(d) ? d : (d?.vlans ?? []))
      } else {
        setError('Failed to load VLANs')
      }
      if (domainsRes.status === 'fulfilled') {
        const d = domainsRes.value.data
        setDomains(Array.isArray(d) ? d : (d?.domains ?? []))
      }
      if (groupsRes.status === 'fulfilled') {
        const d = groupsRes.value.data
        setGroups(Array.isArray(d) ? d : (d?.groups ?? []))
      }
    } finally {
      setLoading(false)
    }
  }

  function showMsg(text, type = 'success') {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 3000)
  }

  // VLANDomain uses json tags so interceptor gives lowercase fields
  function getDomain(domainId) { return domains.find(d => d.id === domainId) }
  // VLANGroup uses json tags so interceptor gives lowercase fields
  function getGroup(groupId) { return groups.find(g => g.id === groupId) }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(vlan) {
    setForm({
      vlanId: vlan.vlanId != null ? String(vlan.vlanId) : '',
      name: vlan.name || '',
      description: vlan.description || '',
      domainId: vlan.domainId != null ? String(vlan.domainId) : '',
      groupId: vlan.groupId != null ? String(vlan.groupId) : '',
    })
    setModal({ edit: vlan })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const payload = {
        vlan_id: form.vlanId ? parseInt(form.vlanId) : undefined,
        name: form.name,
        description: form.description || null,
        domain_id: form.domainId ? parseInt(form.domainId) : null,
        group_id: form.groupId ? parseInt(form.groupId) : null,
      }
      if (modal === 'create') {
        await createVlan(payload)
        showMsg('VLAN created')
      } else {
        await updateVlan(modal.edit.id, payload)
        showMsg('VLAN updated')
      }
      setModal(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save VLAN')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteVlan(id)
      setDeleteConfirm(null)
      showMsg('VLAN deleted')
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to delete VLAN')
      setDeleteConfirm(null)
    }
  }

  if (loading) return <PageSpinner message="Loading VLANs..." />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">VLANs</h1>
        <button
          onClick={openCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
        >
          + New VLAN
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      <ErrorBanner error={error} onDismiss={() => setError(null)} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">ID</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Domain</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Group</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {vlans.length === 0 && <EmptyRow colSpan={6} message="No VLANs yet. Add your first VLAN to get started." />}
            {vlans.map(vlan => {
              const domain = getDomain(vlan.domainId)
              const group = getGroup(vlan.groupId)
              return (
                <tr key={vlan.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">
                    <Link to={`/vlans/${vlan.id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                      {vlan.vlanId}
                    </Link>
                  </td>
                  <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{vlan.name}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    {domain ? (
                      <span className="text-xs text-gray-700 dark:text-gray-300">{domain.name}</span>
                    ) : '—'}
                  </td>
                  <td className="px-4 py-3">
                    <GroupBadge group={group} />
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{vlan.description || '—'}</td>
                  <td className="px-4 py-3 text-right space-x-2">
                    <button
                      onClick={() => openEdit(vlan)}
                      className="text-gray-400 hover:text-blue-600 text-xs"
                    >
                      Edit
                    </button>
                    {deleteConfirm === vlan.id ? (
                      <>
                        <span className="text-red-600 text-xs">Confirm?</span>
                        <button
                          onClick={() => handleDelete(vlan.id)}
                          className="text-red-600 hover:text-red-800 text-xs font-medium"
                        >
                          Yes
                        </button>
                        <button
                          onClick={() => setDeleteConfirm(null)}
                          className="text-gray-400 hover:text-gray-600 text-xs"
                        >
                          No
                        </button>
                      </>
                    ) : (
                      <button
                        onClick={() => setDeleteConfirm(vlan.id)}
                        className="text-gray-400 hover:text-red-600 text-xs"
                      >
                        Delete
                      </button>
                    )}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {modal && (
        <Modal
          title={modal === 'create' ? 'New VLAN' : 'Edit VLAN'}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                VLAN ID <span className="text-red-500">*</span>
              </label>
              <input
                type="number"
                min="1"
                max="4094"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. 100"
                value={form.vlanId}
                onChange={e => setForm(f => ({ ...f, vlanId: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. Management"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={2}
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            {domains.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Domain</label>
                <select
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.domainId}
                  onChange={e => setForm(f => ({ ...f, domainId: e.target.value }))}
                >
                  <option value="">No domain</option>
                  {domains.map(d => (
                    <option key={d.id} value={d.id}>{d.name}</option>
                  ))}
                </select>
              </div>
            )}
            {groups.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Group</label>
                <select
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.groupId}
                  onChange={e => setForm(f => ({ ...f, groupId: e.target.value }))}
                >
                  <option value="">No group</option>
                  {groups.map(g => (
                    <option key={g.id} value={g.id}>{g.name}</option>
                  ))}
                </select>
              </div>
            )}
            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setModal(null)}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
