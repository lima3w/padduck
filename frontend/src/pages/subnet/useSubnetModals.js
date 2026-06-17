import { useState } from 'react'
import { api } from '../../api/client'
import { getSubnet, createSubnet, updateSubnet, deleteSubnet } from '../../api/ipam'

const EMPTY_FORM = {
  network_address: '', prefix_length: '24', description: '', gateway: '',
  auto_reserve_first: false, auto_reserve_last: false, location_id: '',
  nameserver_id: '', vlan_id: '', custom_fields: {},
  alert_threshold_pct: '', alert_email_override: '', technitium_scope_name: '',
}

/**
 * Manages all subnet CRUD and tool modals (create/edit/delete, split, merge, resize).
 * @param {object} opts
 * @param {string|number} opts.networkID
 * @param {Function} opts.load - (page: number) => Promise  (reload the list)
 * @param {Function} opts.loadTree - () => Promise  (reload tree if in tree mode)
 * @param {number} opts.page - current list page
 * @param {string} opts.viewMode - 'list' | 'tree'
 * @param {Function} opts.showToast - (msg: string) => void
 * @param {Function} opts.setError - (msg: string|null) => void
 */
export function useSubnetModals({ networkID, load, loadTree, page, viewMode, showToast, setError }) {
  // Create / edit modal
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [overlapError, setOverlapError] = useState(null)
  const [saving, setSaving] = useState(false)
  // Inline delete confirm
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  // Split
  const [splitModal, setSplitModal] = useState(null)
  const [splitPrefix, setSplitPrefix] = useState('')
  const [splitting, setSplitting] = useState(false)
  const [splitError, setSplitError] = useState('')
  const [splitBlockingIPs, setSplitBlockingIPs] = useState([])
  // Merge
  const [mergeModal, setMergeModal] = useState(null)
  const [mergeSelected, setMergeSelected] = useState([])
  const [merging, setMerging] = useState(false)
  const [mergeError, setMergeError] = useState('')
  // Resize
  const [resizeModal, setResizeModal] = useState(null)
  const [resizePrefix, setResizePrefix] = useState('')
  const [resizing, setResizing] = useState(false)
  const [resizeError, setResizeError] = useState(null)
  const [resizeConfirmText, setResizeConfirmText] = useState('')

  function openCreate() {
    setForm(EMPTY_FORM)
    setOverlapError(null)
    setModal('create')
  }

  async function openEdit(subnet) {
    try {
      const res = await getSubnet(subnet.id)
      const full = res.data
      setForm({
        network_address: full.networkAddress || '',
        prefix_length: full.prefixLength != null ? String(full.prefixLength) : '24',
        description: full.description || '',
        gateway: full.gateway || '',
        auto_reserve_first: full.autoReserveFirst || false,
        auto_reserve_last: full.autoReserveLast || false,
        location_id: full.locationId ? String(full.locationId) : '',
        nameserver_id: full.nameserverId ? String(full.nameserverId) : '',
        vlan_id: full.vlanId != null ? String(full.vlanId) : '',
        custom_fields: full.customFields || {},
        alert_threshold_pct: full.alertThresholdPct != null ? String(full.alertThresholdPct) : '',
        alert_email_override: full.alertEmailOverride || '',
        technitium_scope_name: full.technitiumScopeName || '',
      })
      setOverlapError(null)
      setModal({ edit: full })
    } catch {
      setError('Failed to load subnet details')
    }
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    setOverlapError(null)
    try {
      if (modal === 'create') {
        await createSubnet(networkID, {
          network_address: form.network_address,
          prefix_length: form.prefix_length !== '' ? parseInt(form.prefix_length) : 24,
          description: form.description,
          gateway: form.gateway || null,
          auto_reserve_first: form.auto_reserve_first,
          auto_reserve_last: form.auto_reserve_last,
          location_id: form.location_id ? parseInt(form.location_id) : null,
          nameserver_id: form.nameserver_id ? parseInt(form.nameserver_id) : null,
          vlan_id: form.vlan_id !== '' ? parseInt(form.vlan_id) : null,
          custom_fields: form.custom_fields || {},
        })
      } else {
        const id = modal.edit.id
        await updateSubnet(id, {
          description: form.description,
          gateway: form.gateway || null,
          auto_reserve_first: form.auto_reserve_first,
          auto_reserve_last: form.auto_reserve_last,
          location_id: form.location_id ? parseInt(form.location_id) : null,
          nameserver_id: form.nameserver_id ? parseInt(form.nameserver_id) : null,
          vlan_id: form.vlan_id !== '' ? parseInt(form.vlan_id) : null,
          custom_fields: form.custom_fields || {},
          alert_threshold_pct: form.alert_threshold_pct ? parseInt(form.alert_threshold_pct) : null,
          alert_email_override: form.alert_email_override || null,
          technitium_scope_name: form.technitium_scope_name || '',
        })
      }
      setModal(null)
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      if (err.response?.status === 409) {
        const conflicting = err.response.data.conflictingCidr
        setOverlapError(`Subnet overlaps with existing subnet${conflicting ? ': ' + conflicting : ''}`)
      } else {
        setError(err.response?.data?.error || 'Failed to save subnet')
      }
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteSubnet(id)
      setDeleteConfirm(null)
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch {
      setError('Failed to delete subnet')
    }
  }

  function openSplit(subnet) {
    setSplitModal({ subnet })
    setSplitPrefix(String(subnet.prefixLength + 1))
    setSplitError('')
    setSplitBlockingIPs([])
  }

  async function handleSplit() {
    if (!splitModal) return
    setSplitting(true)
    setSplitError('')
    setSplitBlockingIPs([])
    try {
      await api.post(`/admin/subnets/${splitModal.subnet.id}/split`, { new_prefix_len: parseInt(splitPrefix) })
      setSplitModal(null)
      showToast('Subnet split successfully')
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      const data = err.response?.data
      if (data?.blockingIps?.length) {
        setSplitBlockingIPs(data.blockingIps)
        setSplitError('Split blocked: the following IPs fall on network or broadcast addresses and must be removed first.')
      } else {
        setSplitError(data?.error || 'Failed to split subnet')
      }
    } finally {
      setSplitting(false)
    }
  }

  async function openMerge(subnet) {
    try {
      const res = await api.get(`/networks/${networkID}/subnets`)
      const all = res.data?.data ?? res.data ?? []
      const siblings = all.filter(s => s.id !== subnet.id && s.prefixLength === subnet.prefixLength)
      setMergeModal({ subnet, siblings })
      setMergeSelected([])
      setMergeError('')
    } catch {
      setError('Failed to load siblings for merge')
    }
  }

  async function handleMerge() {
    if (!mergeModal) return
    setMerging(true)
    setMergeError('')
    try {
      const ids = [mergeModal.subnet.id, ...mergeSelected]
      await api.post('/admin/subnets/merge', { subnet_ids: ids })
      setMergeModal(null)
      showToast('Subnets merged successfully')
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      setMergeError(err.response?.data?.error || 'Failed to merge subnets')
    } finally {
      setMerging(false)
    }
  }

  function openResize(subnet) {
    setResizeModal({ subnet })
    setResizePrefix(`${subnet.networkAddress}/${subnet.prefixLength}`)
    setResizeError(null)
    setResizeConfirmText('')
  }

  async function handleResize() {
    if (!resizeModal) return
    setResizing(true)
    setResizeError(null)
    try {
      await api.post(`/admin/subnets/${resizeModal.subnet.id}/resize`, { new_prefix: resizePrefix })
      setResizeModal(null)
      showToast('Subnet resized successfully')
      load(page)
      if (viewMode === 'tree') loadTree()
    } catch (err) {
      if (err.response?.status === 409) {
        const d = err.response.data
        setResizeError({
          message: d.error || 'Resize conflicts with existing data',
          conflictingIps: d.conflictingIps || [],
          conflictingSubnets: d.conflictingSubnets || [],
        })
      } else {
        setResizeError({ message: err.response?.data?.error || 'Failed to resize subnet' })
      }
    } finally {
      setResizing(false)
    }
  }

  return {
    // create/edit
    modal, setModal, form, setForm, overlapError, saving,
    openCreate, openEdit, handleSubmit,
    // delete
    deleteConfirm, setDeleteConfirm, handleDelete,
    // split
    splitModal, setSplitModal, splitPrefix, setSplitPrefix,
    splitting, splitError, splitBlockingIPs,
    openSplit, handleSplit,
    // merge
    mergeModal, setMergeModal, mergeSelected, setMergeSelected,
    merging, mergeError,
    openMerge, handleMerge,
    // resize
    resizeModal, setResizeModal, resizePrefix, setResizePrefix,
    resizing, resizeError, setResizeError, resizeConfirmText, setResizeConfirmText,
    openResize, handleResize,
  }
}
