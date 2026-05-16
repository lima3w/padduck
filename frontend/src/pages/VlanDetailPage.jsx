import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import {
  assignSubnetToVlan,
  getSections,
  getSubnetsPaginated,
  getVlan,
  getVlanDomains,
  getVlanGroups,
  getVlanSubnets,
  removeSubnetFromVlan,
  updateVlan,
} from '../api/client'
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
  const [sections, setSections] = useState([])
  const [domains, setDomains] = useState([])
  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)
  const [editModal, setEditModal] = useState(false)
  const [form, setForm] = useState({ vlanId: '', name: '', description: '', domainId: '', groupId: '' })
  const [saving, setSaving] = useState(false)
  const [assignModal, setAssignModal] = useState(false)
  const [assignSectionId, setAssignSectionId] = useState('')
  const [assignSubnets, setAssignSubnets] = useState([])
  const [assignSubnetId, setAssignSubnetId] = useState('')
  const [assigning, setAssigning] = useState(false)
  const [loadingAssignSubnets, setLoadingAssignSubnets] = useState(false)

  useEffect(() => { load() }, [id])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const [vlanRes, subnetsRes, domainsRes, groupsRes, sectionsRes] = await Promise.allSettled([
        getVlan(id),
        getVlanSubnets(id),
        getVlanDomains(),
        getVlanGroups(),
        getSections(),
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
      if (sectionsRes.status === 'fulfilled') {
        const d = sectionsRes.value.data
        setSections(Array.isArray(d) ? d : (d?.sections ?? []))
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
      vlanId: vlan.vlanId != null ? String(vlan.vlanId) : '',
      name: vlan.name || '',
      description: vlan.description || '',
      domainId: vlan.domainId != null ? String(vlan.domainId) : '',
      groupId: vlan.groupId != null ? String(vlan.groupId) : '',
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

  async function loadAssignableSubnets(sectionId) {
    if (!sectionId) {
      setAssignSubnets([])
      setAssignSubnetId('')
      return
    }
    setLoadingAssignSubnets(true)
    try {
      const res = await getSubnetsPaginated(sectionId, 1, 1000)
      const data = res.data
      const rows = data.data ?? data
      const assignedIds = new Set(subnets.map(s => s.id))
      const candidates = (Array.isArray(rows) ? rows : [])
        .filter(s => !assignedIds.has(s.id))
      setAssignSubnets(candidates)
      setAssignSubnetId(candidates[0] ? String(candidates[0].id) : '')
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load section subnets')
    } finally {
      setLoadingAssignSubnets(false)
    }
  }

  async function openAssignSubnet() {
    const firstSectionId = sections[0]?.id ?? ''
    setAssignSectionId(firstSectionId ? String(firstSectionId) : '')
    setAssignSubnetId('')
    setAssignSubnets([])
    setAssignModal(true)
    if (firstSectionId) await loadAssignableSubnets(String(firstSectionId))
  }

  async function handleAssignSubnet(e) {
    e.preventDefault()
    if (!assignSubnetId) return
    setAssigning(true)
    try {
      await assignSubnetToVlan(id, parseInt(assignSubnetId))
      showMsg('Subnet assigned to VLAN')
      setAssignModal(false)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to assign subnet')
    } finally {
      setAssigning(false)
    }
  }

  async function handleRemoveSubnet(subnetId) {
    try {
      await removeSubnetFromVlan(id, subnetId)
      showMsg('Subnet removed from VLAN')
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to remove subnet')
    }
  }

  if (loading) return <p className="text-gray-500">Loading VLAN...</p>
  if (error && !vlan) return <p className="text-red-600">{error}</p>
  if (!vlan) return <p className="text-gray-500">VLAN not found.</p>

  const domain = getDomain(vlan.domainId)
  const group = getGroup(vlan.groupId)

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-4 text-sm text-gray-500 dark:text-gray-400">
        <Link to="/vlans" className="hover:text-blue-600 dark:hover:text-blue-400">VLANs</Link>
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-medium">VLAN {vlan.vlanId} — {vlan.name}</span>
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
            VLAN {vlan.vlanId} — {vlan.name}
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
            <p className="font-mono font-semibold text-gray-800 dark:text-gray-200">{vlan.vlanId}</p>
          </div>
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">Name</p>
            <p className="font-medium text-gray-800 dark:text-gray-200">{vlan.name}</p>
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
          {vlan.description && (
            <div className="col-span-2">
              <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">Description</p>
              <p className="text-gray-700 dark:text-gray-300">{vlan.description}</p>
            </div>
          )}
        </div>
      </div>

      {/* Subnets in this VLAN */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div className="px-6 py-4 border-b dark:border-gray-700 flex items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">Subnets in this VLAN</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
              {subnets.length} subnet{subnets.length !== 1 ? 's' : ''} assigned to VLAN {vlan.vlanId}
            </p>
          </div>
          <button
            onClick={openAssignSubnet}
            className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded"
          >
            Add Subnet
          </button>
        </div>
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">CIDR</th>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">Section</th>
              <th className="px-6 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {subnets.length === 0 && (
              <tr>
                <td colSpan={4} className="px-6 py-6 text-center text-gray-400">
                  No subnets are assigned to this VLAN.
                </td>
              </tr>
            )}
            {subnets.map(subnet => {
              const cidr = subnet.networkAddress || ''
              const prefix = subnet.prefixLength ?? ''
              const desc = subnet.description || ''
              const secId = subnet.sectionId
              const subnetId = subnet.id
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
                  <td className="px-6 py-3 text-right">
                    <button
                      onClick={() => handleRemoveSubnet(subnetId)}
                      className="text-xs text-gray-400 hover:text-red-600"
                    >
                      Remove
                    </button>
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

      {assignModal && (
        <Modal title="Add Subnet to VLAN" onClose={() => setAssignModal(false)}>
          <form onSubmit={handleAssignSubnet} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Section</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={assignSectionId}
                onChange={e => {
                  setAssignSectionId(e.target.value)
                  loadAssignableSubnets(e.target.value)
                }}
              >
                {sections.length === 0 && <option value="">No sections available</option>}
                {sections.map(section => (
                  <option key={section.id} value={section.id}>
                    {section.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subnet</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={assignSubnetId}
                onChange={e => setAssignSubnetId(e.target.value)}
                disabled={loadingAssignSubnets || assignSubnets.length === 0}
              >
                {loadingAssignSubnets && <option value="">Loading subnets...</option>}
                {!loadingAssignSubnets && assignSubnets.length === 0 && <option value="">No available subnets</option>}
                {!loadingAssignSubnets && assignSubnets.map(subnet => {
                  const subnetId = subnet.id
                  const cidr = subnet.networkAddress || ''
                  const prefix = subnet.prefixLength ?? ''
                  const desc = subnet.description || ''
                  return (
                    <option key={subnetId} value={subnetId}>
                      {cidr}/{prefix}{desc ? ` — ${desc}` : ''}
                    </option>
                  )
                })}
              </select>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setAssignModal(false)}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={assigning || !assignSubnetId}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {assigning ? 'Adding...' : 'Add Subnet'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
