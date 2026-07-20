import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import ChangeHistory from '../components/ChangeHistory'
import ObjectRelationshipsPanel from '../components/ObjectRelationshipsPanel'
import { getNetworks, getSubnetsPaginated } from '../api/ipam'
import { assignSubnetToVlan, getVlan, getVlanDomains, getVlanGroups, getVlanSubnets, removeSubnetFromVlan, updateVlan } from '../api/vlans'
import Modal from '../components/Modal'
import { getCachedUser } from '../utils/storageKeys'

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
  const { t } = useTranslation()
  const { id } = useParams()
  const [vlan, setVlan] = useState(null)
  const [subnets, setSubnets] = useState([])
  const [networks, setSections] = useState([])
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

  const load = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const [vlanRes, subnetsRes, domainsRes, groupsRes, sectionsRes] = await Promise.allSettled([
        getVlan(id),
        getVlanSubnets(id),
        getVlanDomains(),
        getVlanGroups(),
        getNetworks(),
      ])
      if (vlanRes.status === 'fulfilled') {
        setVlan(vlanRes.value.data)
      } else {
        setError(t('vlanDetail.loadFailed'))
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
        setSections(Array.isArray(d) ? d : (d?.networks ?? []))
      }
    } finally {
      setLoading(false)
    }
  }, [id, t])

  useEffect(() => { load() }, [load])

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
      showMsg(t('vlans.updated'))
      setEditModal(false)
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('vlanDetail.updateFailed'))
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
      setError(err.response?.data?.error || t('vlanDetail.loadSubnetsFailed'))
    } finally {
      setLoadingAssignSubnets(false)
    }
  }

  async function openAssignSubnet() {
    const firstSectionId = networks[0]?.id ?? ''
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
      showMsg(t('vlanDetail.subnetAssigned'))
      setAssignModal(false)
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('vlanDetail.assignFailed'))
    } finally {
      setAssigning(false)
    }
  }

  async function handleRemoveSubnet(subnetId) {
    try {
      await removeSubnetFromVlan(id, subnetId)
      showMsg(t('vlanDetail.subnetRemoved'))
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('vlanDetail.removeFailed'))
    }
  }

  if (loading) return <p className="text-gray-500">{t('vlanDetail.loadingVlan')}</p>
  if (error && !vlan) return <p className="text-red-600">{error}</p>
  if (!vlan) return <p className="text-gray-500">{t('vlanDetail.vlanNotFound')}</p>

  const isAdmin = getCachedUser()?.role === 'admin'
  const domain = getDomain(vlan.domainId)
  const group = getGroup(vlan.groupId)
  const sectionIds = new Set(subnets.map(s => s.networkId).filter(Boolean))
  const relationshipItems = [
    domain && {
      label: t('vlans.domain'),
      value: domain.name,
      description: t('vlanDetail.domainDescription'),
    },
    group && {
      label: t('vlans.group'),
      value: group.name,
      description: t('vlanDetail.groupDescription'),
    },
    {
      label: t('dashboard.subnets'),
      value: t('vlanDetail.assignedSubnetsValue'),
      count: subnets.length,
      description: t('vlanDetail.subnetsAssignedToVlan', { count: subnets.length }),
    },
    {
      label: t('nav.networks'),
      value: t('vlanDetail.relatedNetworksValue'),
      count: sectionIds.size,
      description: t('vlanDetail.networksRepresented', { count: sectionIds.size }),
    },
  ]

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-4 text-sm text-gray-500 dark:text-gray-400">
        <Link to="/vlans" className="hover:text-blue-600 dark:hover:text-blue-400">{t('nav.vlans')}</Link>
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-medium">{t('vlanDetail.vlanHeading', { vlanId: vlan.vlanId, name: vlan.name })}</span>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {message.text}
        </div>
      )}
      {error && (
        <div className="mb-4 p-3 rounded text-sm bg-red-50 text-red-700 border border-red-200">
          {error}
          <button onClick={() => setError(null)} className="ml-2 underline">{t('common.dismiss')}</button>
        </div>
      )}

      {/* VLAN Details Card */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow mb-6">
        <div className="flex items-center justify-between px-6 py-4 border-b dark:border-gray-700">
          <h1 className="text-xl font-bold text-gray-800 dark:text-gray-100">
            {t('vlanDetail.vlanHeading', { vlanId: vlan.vlanId, name: vlan.name })}
          </h1>
          <button
            onClick={openEdit}
            className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded"
          >
            {t('common.edit')}
          </button>
        </div>
        <div className="px-6 py-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">{t('vlans.vlanIdLabel')}</p>
            <p className="font-mono font-semibold text-gray-800 dark:text-gray-200">{vlan.vlanId}</p>
          </div>
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">{t('common.name')}</p>
            <p className="font-medium text-gray-800 dark:text-gray-200">{vlan.name}</p>
          </div>
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">{t('vlans.domain')}</p>
            {domain ? (
              <span className="text-gray-700 dark:text-gray-300">{domain.name}</span>
            ) : (
              <span className="text-gray-400">—</span>
            )}
          </div>
          <div>
            <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">{t('vlans.group')}</p>
            <GroupBadge group={group} />
          </div>
          {vlan.description && (
            <div className="col-span-2">
              <p className="text-gray-500 dark:text-gray-400 text-xs font-medium uppercase tracking-wider mb-1">{t('common.description')}</p>
              <p className="text-gray-700 dark:text-gray-300">{vlan.description}</p>
            </div>
          )}
        </div>
      </div>

      <ObjectRelationshipsPanel relationships={relationshipItems} />

      {/* Subnets in this VLAN */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div className="px-6 py-4 border-b dark:border-gray-700 flex items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">{t('vlanDetail.subnetsInVlanTitle')}</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
              {t('vlanDetail.subnetsAssignedCountSubtitle', { count: subnets.length, vlanId: vlan.vlanId })}
            </p>
          </div>
          <button
            onClick={openAssignSubnet}
            className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded"
          >
            {t('vlanDetail.addSubnet')}
          </button>
        </div>
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('topology.cidr')}</th>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.description')}</th>
              <th className="text-left px-6 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('subnets.network')}</th>
              <th className="px-6 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {subnets.length === 0 && (
              <tr>
                <td colSpan={4} className="px-6 py-6 text-center text-gray-400">
                  {t('vlanDetail.noSubnetsAssigned')}
                </td>
              </tr>
            )}
            {subnets.map(subnet => {
              const cidr = subnet.networkAddress || ''
              const prefix = subnet.prefixLength ?? ''
              const desc = subnet.description || ''
              const secId = subnet.networkId
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
                        to={`/networks/${secId}/subnets`}
                        className="text-blue-600 dark:text-blue-400 hover:underline text-xs"
                      >
                        {t('vlanDetail.networkHash', { id: secId })}
                      </Link>
                    ) : '—'}
                  </td>
                  <td className="px-6 py-3 text-right">
                    <button
                      onClick={() => handleRemoveSubnet(subnetId)}
                      className="text-xs text-gray-400 hover:text-red-600"
                    >
                      {t('deviceIp.remove')}
                    </button>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {isAdmin && <ChangeHistory resourceType="vlan" resourceId={vlan?.id} />}

      {editModal && (
        <Modal title={t('vlans.editVlanModalTitle')} onClose={() => setEditModal(false)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('vlans.vlanIdLabel')} <span className="text-red-500">*</span>
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
                {t('common.name')} <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
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
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            {domains.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('vlans.domain')}</label>
                <select
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.domainId}
                  onChange={e => setForm(f => ({ ...f, domainId: e.target.value }))}
                >
                  <option value="">{t('vlans.noDomain')}</option>
                  {domains.map(d => (
                    <option key={d.id} value={d.id}>{d.name}</option>
                  ))}
                </select>
              </div>
            )}
            {groups.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('vlans.group')}</label>
                <select
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={form.groupId}
                  onChange={e => setForm(f => ({ ...f, groupId: e.target.value }))}
                >
                  <option value="">{t('vlans.noGroup')}</option>
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

      {assignModal && (
        <Modal title={t('vlanDetail.addSubnetModalTitle')} onClose={() => setAssignModal(false)}>
          <form onSubmit={handleAssignSubnet} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('subnets.network')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={assignSectionId}
                onChange={e => {
                  setAssignSectionId(e.target.value)
                  loadAssignableSubnets(e.target.value)
                }}
              >
                {networks.length === 0 && <option value="">{t('vlanDetail.noNetworksAvailable')}</option>}
                {networks.map(network => (
                  <option key={network.id} value={network.id}>
                    {network.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('dashboard.subnet')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={assignSubnetId}
                onChange={e => setAssignSubnetId(e.target.value)}
                disabled={loadingAssignSubnets || assignSubnets.length === 0}
              >
                {loadingAssignSubnets && <option value="">{t('vlanDetail.loadingSubnets')}</option>}
                {!loadingAssignSubnets && assignSubnets.length === 0 && <option value="">{t('vlanDetail.noAvailableSubnets')}</option>}
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
                {t('common.cancel')}
              </button>
              <button
                type="submit"
                disabled={assigning || !assignSubnetId}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {assigning ? t('adminLdap.adding') : t('vlanDetail.addSubnet')}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
