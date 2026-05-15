import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getVlan, getVlanSubnets, getVlanDomains, getVlanGroups, updateVlan } from '../api/client'
import Modal from '../components/Modal'

// VLAN model has no JSON tags — Go field names come through as PascalCase:
//   ID, VlanID, DomainID, GroupID, VRFID, Name, Description, CreatedAt, UpdatedAt
// VLANDomain/VLANGroup have json tags → snake_case → camelCase via interceptor

function GroupBadge({ group }) {
  if (!group) return <span className="text-gray-400">—</span>
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

export default function VlanDetailPage() {
  const { id } = useParams()
  const [vlan, setVlan] = useState(null)
  const [subnets, setSubnets] = useState([])
  const [domains, setDomains] = useState([])
  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [editModal, setEditModal] = useState(false)
  const [form, setForm] = useState({ vlanId: '', name: '', description: '', domainId: '', groupId: '' })
  const [saving, setSaving] = useState(false)

  useEffect(() => { load() }, [id])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const [vlanRes, subnetsRes, domainsRes, groupsRes] = await Promise.allSettled([
        getVlan(id),
        getVlanSubnets(id),
        getVlanDomains(),
        getVlanGroups(),
      ])
      if (vlanRes.status === 'fulfilled') {
        setVlan(vlanRes.value.data)
      } else {
        setError('Failed to load VLAN')
      }
      if (subnetsRes.status === 'fulfilled') {
        const d = subnetsRes.value.data
        setSubnets(Array.isArray(d) ? d : (d?.subnets ?? []))
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

  function getDomain(domainId) { return domains.find(d => d.id === domainId) }
  function getGroup(groupId) { return groups.find(g => g.id === groupId) }

  function openEdit() {
    if (!vlan) return
    setForm({
      vlanId: vlan.VlanID != null ? String(vlan.VlanID) : '',
      name: vlan.Name || '',
      description: vlan.Description || '',
      domainId: vlan.DomainID != null ? String(vlan.DomainID) : '',
      groupId: vlan.GroupID != null ? String(vlan.GroupID) : '',
    })
    setEditModal(true)
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
      await updateVlan(id, payload)
      showMsg('VLAN updated')
      setEditModal(false)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update VLAN')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <p className="text-gray-500">Loading VLAN...</p>
  if (error && !vlan) return <p className="text-red-600">{error}</p>
  if (!vlan) return <p className="text-gray-500">VLAN not found.</p>

  // VLAN fields are PascalCase; domain/group use camelCase (have json tags)
  const domain = getDomain(vlan.DomainID)
  const group = getGroup(vlan.GroupID)

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-4 text-sm text-gray-500 dark:text-gray-400">
        <Link to="/vlans" className="hover:text-blue-600 dark:hover:text-blue-400">VLANs</Link>
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-medium">VLAN {vlan.VlanID} — {vlan.Name}</span>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      {error && (
        <div className="mb-4 p-3 rounded text-sm bg-red-50 text-red-700 border border-red-200">
          {error}
          <button onClick={() => setError(null)} className="ml-2 underline">Dismiss</button>
        </div>
      )}

      {/* VLAN Details Card */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow mb-6">
        <div className="flex items-center justify-between px-6 py-4 border-b dark:border-gray-700">
          <h1 className="text-xl font-bold text-gray-800 dark:text-gray-100">
            VLAN {vlan.VlanID} — {vlan.Name}
          </h1>
          <button
            onClick={openEdit}
            className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded"
          >
            Edit
          </button>
        </div>
        <div className="px-6 py-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">VLAN ID</p>
            <p className="font-mono font-semibold text-gray-800 dark:text-gray-200">{vlan.VlanID}</p>
          </div>
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">Name</p>
            <p className="font-medium text-gray-800 dark:text-gray-200">{vlan.Name}</p>
          </div>
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">Domain</p>
            {domain ? (
              <span className="text-gray-700 dark:text-gray-300">{domain.name}</span>
            ) : (
              <span className="text-gray-400">—</span>
            )}
          </div>
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">Group</p>
            <GroupBadge group={group} />
          </div>
          {vlan.Description && (
            <div className="col-span-2">
              <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">Description</p>
              <p className="text-gray-700 dark:text-gray-300">{vlan.Description}</p>
            </div>
          )}
        </div>
      </div>

      {/* Subnets in this VLAN */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div className="px-6 py-4 border-b dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">Subnets in this VLAN</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            {subnets.length} subnet{subnets.length !== 1 ? 's' : ''} assigned to VLAN {vlan.VlanID}
          </p>
        </div>
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">CIDR</th>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">Section</th>
            </tr>
          </thead>
          <tbody>
            {subnets.length === 0 && (
              <tr>
                <td colSpan={3} className="px-6 py-6 text-center text-gray-400">
                  No subnets are assigned to this VLAN.
                </td>
              </tr>
            )}
            {subnets.map(subnet => {
              // Subnet model may or may not have json tags — handle both
              const cidr = subnet.networkAddress || subnet.network_address || subnet.NetworkAddress || ''
              const prefix = subnet.prefixLength ?? subnet.prefix_length ?? subnet.PrefixLength ?? ''
              const desc = subnet.description || subnet.Description || ''
              const secId = subnet.sectionId ?? subnet.section_id ?? subnet.SectionID
              const subnetId = subnet.id ?? subnet.ID
              return (
                <tr key={subnetId} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td className="px-6 py-3">
                    <Link
                      to={`/subnets/${subnetId}/ip-addresses`}
                      className="font-mono text-blue-600 dark:text-blue-400 hover:underline font-medium"
                    >
                      {cidr}/{prefix}
                    </Link>
                  </td>
                  <td className="px-6 py-3 text-gray-500 dark:text-gray-400">
                    {desc || '—'}
                  </td>
                  <td className="px-6 py-3 text-gray-500 dark:text-gray-400">
                    {secId ? (
                      <Link
                        to={`/sections/${secId}/subnets`}
                        className="text-blue-600 dark:text-blue-400 hover:underline text-xs"
                      >
                        Section #{secId}
                      </Link>
                    ) : '—'}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {editModal && (
        <Modal title="Edit VLAN" onClose={() => setEditModal(false)}>
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
                onClick={() => setEditModal(false)}
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
