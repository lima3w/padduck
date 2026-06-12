import { useState, useEffect, useCallback } from 'react'
import Modal from '../components/Modal'
import { getLocations } from '../api/locations'
import { getAdminUsers, getAdminRoles, getUserRoles, assignUserRole, removeUserRole, createUser, adminUnlockUser, suspendUser, unsuspendUser, impersonateUser, sendPasswordResetEmail, updateUserEmail, gdprDeleteUser, bulkSuspendUsers, bulkActivateUsers, bulkDeleteUsers, getBreakGlassStatus, activateBreakGlass, endBreakGlass } from '../api/admin'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'
import PermissionDenied from '../components/PermissionDenied'

// ── Break-Glass helpers ──────────────────────────────────────────────────────

function bgFormatDatetime(str) {
  if (!str) return '—'
  return new Date(str).toLocaleString(undefined, {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

function bgFormatDuration(startStr, endStr) {
  if (!endStr) return 'Active'
  const ms = new Date(endStr) - new Date(startStr)
  if (ms < 0) return '—'
  const totalSecs = Math.floor(ms / 1000)
  const h = Math.floor(totalSecs / 3600)
  const m = Math.floor((totalSecs % 3600) / 60)
  const s = totalSecs % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

function BgSessionStatus({ session }) {
  if (session.isActive) {
    return (
      <span className="px-2 py-0.5 rounded text-xs font-semibold bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300">
        Active
      </span>
    )
  }
  if (session.endedAt) {
    return (
      <span className="px-2 py-0.5 rounded text-xs font-semibold bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400">
        Ended
      </span>
    )
  }
  return (
    <span className="px-2 py-0.5 rounded text-xs font-semibold bg-yellow-100 text-yellow-700 dark:bg-yellow-900/50 dark:text-yellow-300">
      Expired
    </span>
  )
}

function BgTimeRemaining({ expiresAt }) {
  const [remaining, setRemaining] = useState('')

  useEffect(() => {
    function calc() {
      const diff = new Date(expiresAt) - Date.now()
      if (diff <= 0) { setRemaining('Expired'); return }
      const totalSecs = Math.floor(diff / 1000)
      const h = Math.floor(totalSecs / 3600)
      const m = Math.floor((totalSecs % 3600) / 60)
      const s = totalSecs % 60
      if (h > 0) setRemaining(`${h}h ${m}m ${s}s`)
      else if (m > 0) setRemaining(`${m}m ${s}s`)
      else setRemaining(`${s}s`)
    }
    calc()
    const id = setInterval(calc, 1000)
    return () => clearInterval(id)
  }, [expiresAt])

  return <span className="font-mono font-semibold">{remaining}</span>
}

function BreakGlassTab() {
  const [active, setActive] = useState(null)
  const [history, setHistory] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [justification, setJustification] = useState('')
  const [activating, setActivating] = useState(false)
  const [ending, setEnding] = useState(false)
  const [actionError, setActionError] = useState(null)

  const fetchStatus = useCallback(() => {
    setLoading(true)
    setError(null)
    getBreakGlassStatus()
      .then(res => {
        setActive(res.data?.active ?? null)
        setHistory(res.data?.history ?? [])
      })
      .catch(() => setError('Failed to load break-glass status'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    fetchStatus()
  }, [fetchStatus])

  async function handleActivate(e) {
    e.preventDefault()
    setActionError(null)
    if (justification.trim().length < 10) {
      setActionError('Justification must be at least 10 characters')
      return
    }
    if (!window.confirm('This action is fully audited. A break-glass session will be recorded and visible to all administrators. Continue?')) {
      return
    }
    setActivating(true)
    try {
      await activateBreakGlass(justification.trim())
      setJustification('')
      fetchStatus()
    } catch (err) {
      setActionError(err.response?.data?.error || 'Failed to activate break-glass session')
    } finally {
      setActivating(false)
    }
  }

  async function handleEnd() {
    setActionError(null)
    if (!window.confirm('End the current break-glass session?')) return
    setEnding(true)
    try {
      await endBreakGlass()
      fetchStatus()
    } catch (err) {
      setActionError(err.response?.data?.error || 'Failed to end break-glass session')
    } finally {
      setEnding(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
        Loading...
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-6 text-center text-sm text-red-700 dark:text-red-300">
        {error}
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Break-Glass Emergency Access</h2>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Enable a temporary emergency access session. All actions are fully audited.
        </p>
      </div>

      {actionError && (
        <div className="rounded bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-red-700 dark:text-red-300">
          {actionError}
        </div>
      )}

      {/* Current Status */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Current Status</h3>
        {active ? (
          <div className="rounded-lg border-2 border-red-500 bg-red-50 dark:bg-red-900/20 p-4 space-y-3">
            <div className="flex items-center gap-3">
              <span className="text-lg font-bold text-red-700 dark:text-red-300 uppercase tracking-wide">
                BREAK-GLASS ACTIVE
              </span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">Activated by user ID:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{active.initiatedByUserId}</span>
              </div>
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">Started:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{bgFormatDatetime(active.createdAt)}</span>
              </div>
              <div className="sm:col-span-2">
                <span className="font-medium text-gray-700 dark:text-gray-300">Justification:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{active.justification}</span>
              </div>
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">Time remaining:</span>{' '}
                <BgTimeRemaining expiresAt={active.expiresAt} />
              </div>
            </div>
            <div className="pt-2">
              <button
                onClick={handleEnd}
                disabled={ending}
                className="px-4 py-2 rounded text-sm font-medium text-white bg-red-600 hover:bg-red-700 disabled:opacity-50 transition-colors"
              >
                {ending ? 'Ending...' : 'End Session'}
              </button>
            </div>
          </div>
        ) : (
          <div className="rounded-lg border border-green-300 dark:border-green-700 bg-green-50 dark:bg-green-900/20 px-4 py-3 text-sm text-green-800 dark:text-green-300 font-medium">
            No active break-glass session.
          </div>
        )}
      </div>

      {/* Activate Form (only shown when no active session) */}
      {!active && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">Activate Break-Glass</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            The session will last 1 hour. Provide a justification explaining why emergency access is required.
          </p>
          <form onSubmit={handleActivate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Justification <span className="text-red-500">*</span>
                <span className="ml-1 text-gray-400 dark:text-gray-500 font-normal">(minimum 10 characters)</span>
              </label>
              <textarea
                value={justification}
                onChange={e => setJustification(e.target.value)}
                rows={4}
                placeholder="Describe the emergency requiring break-glass access..."
                className="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-red-500"
              />
            </div>
            <button
              type="submit"
              disabled={activating}
              className="px-4 py-2 rounded text-sm font-medium text-white bg-red-600 hover:bg-red-700 disabled:opacity-50 transition-colors"
            >
              {activating ? 'Activating...' : 'Activate Break-Glass'}
            </button>
          </form>
        </div>
      )}

      {/* Session History */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Session History</h3>
        {history.length === 0 ? (
          <div className="rounded bg-white dark:bg-gray-800 shadow px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
            No break-glass sessions recorded.
          </div>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Date</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Initiated By</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Justification</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Duration</th>
                    <th className="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                  {history.map(s => (
                    <tr key={s.id} className="hover:bg-gray-50 dark:hover:bg-gray-750">
                      <td className="px-4 py-3 text-gray-700 dark:text-gray-300 whitespace-nowrap">
                        {bgFormatDatetime(s.createdAt)}
                      </td>
                      <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                        User {s.initiatedByUserId}
                      </td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-xs truncate" title={s.justification}>
                        {s.justification}
                      </td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                        {bgFormatDuration(s.createdAt, s.endedAt || (s.isActive ? null : s.expiresAt))}
                      </td>
                      <td className="px-4 py-3">
                        <BgSessionStatus session={s} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              </div>
            </div>
        )}
      </div>
    </div>
  )
}

const ASSIGN_EMPTY_FORM = { role_id: '', location_id: '' }
const CREATE_EMPTY_FORM = { username: '', email: '', password: '', role: 'user' }

function normalizeRole(role) {
  if (!role || typeof role !== 'object') return role
  return {
    ...role,
    id: role.id ?? role.ID,
    name: role.name ?? role.Name,
    description: role.description ?? role.Description,
  }
}

function normalizeRoles(data) {
  return Array.isArray(data) ? data.map(normalizeRole) : []
}

export default function AdminUsersPage() {
  const [activeTab, setActiveTab] = useState('users')
  const [users, setUsers] = useState([])
  const [roles, setRoles] = useState([])
  const [locations, setLocations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [permissionDenied, setPermissionDenied] = useState(false)
  const [expandedUser, setExpandedUser] = useState(null)
  const [userRoles, setUserRoles] = useState({}) // userId -> roles[]
  const [assignModal, setAssignModal] = useState(null) // userId
  const [assignForm, setAssignForm] = useState(ASSIGN_EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [removeConfirm, setRemoveConfirm] = useState(null) // { userId, roleId }
  const [createModal, setCreateModal] = useState(false)
  const [createForm, setCreateForm] = useState(CREATE_EMPTY_FORM)
  const [createError, setCreateError] = useState('')
  const [message, setMessage] = useState(null)

  // Lifecycle actions
  const [suspendModal, setSuspendModal] = useState(null) // user object
  const [suspendReason, setSuspendReason] = useState('')
  const [emailModal, setEmailModal] = useState(null) // user object
  const [emailValue, setEmailValue] = useState('')
  const [gdprConfirm, setGdprConfirm] = useState(null) // user object
  const [actionLoading, setActionLoading] = useState(null) // userId + action key

  // Bulk actions
  const [selected, setSelected] = useState(new Set())
  const [bulkAction, setBulkAction] = useState('')
  const [bulkSuspendReason, setBulkSuspendReason] = useState('')

  useEffect(() => {
    loadAll()
  }, [])

  async function loadAll() {
    try {
      setLoading(true)
      setError(null)
      const [usersRes, rolesRes] = await Promise.all([
        getAdminUsers(),
        getAdminRoles(),
      ])
      const usersData = usersRes.data
      setUsers(Array.isArray(usersData) ? usersData : (usersData?.users ?? []))
      const rolesData = rolesRes.data
      setRoles(normalizeRoles(rolesData))
      const locsData = await getLocations().catch(() => [])
      setLocations(Array.isArray(locsData) ? locsData : (locsData?.locations ?? []))
    } catch (err) {
      if (err.response?.status === 403) {
        setPermissionDenied(true)
      } else {
        setError(err.response?.data?.error || err.message || 'Failed to load data')
      }
    } finally {
      setLoading(false)
    }
  }

  async function loadUserRoles(userId) {
    try {
      const res = await getUserRoles(userId)
      setUserRoles(prev => ({ ...prev, [userId]: normalizeRoles(res.data) }))
    } catch {}
  }

  async function toggleExpand(userId) {
    if (expandedUser === userId) {
      setExpandedUser(null)
      return
    }
    setExpandedUser(userId)
    if (!userRoles[userId]) {
      await loadUserRoles(userId)
    }
  }

  function openAssign(userId) {
    setAssignForm(ASSIGN_EMPTY_FORM)
    setAssignModal(userId)
  }

  async function handleAssignSubmit(e) {
    e.preventDefault()
    if (!assignForm.role_id) return
    setSaving(true)
    try {
      const body = { role_id: parseInt(assignForm.role_id) }
      if (assignForm.location_id) body.location_id = parseInt(assignForm.location_id)
      await assignUserRole(assignModal, body)
      setAssignModal(null)
      await loadUserRoles(assignModal)
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to assign role')
    } finally {
      setSaving(false)
    }
  }

  async function handleRemoveRole(userId, roleId) {
    try {
      await removeUserRole(userId, roleId)
      setRemoveConfirm(null)
      await loadUserRoles(userId)
    } catch {
      setError('Failed to remove role')
    }
  }

  async function handleCreateSubmit(e) {
    e.preventDefault()
    setSaving(true)
    setCreateError('')
    try {
      await createUser(createForm)
      setCreateModal(false)
      setCreateForm(CREATE_EMPTY_FORM)
      await loadAll()
    } catch (err) {
      setCreateError(err.response?.data?.error || err.message || 'Failed to create user')
    } finally {
      setSaving(false)
    }
  }

  const showMsg = (text, type = 'success') => {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 4000)
  }

  const runAction = async (userId, key, fn) => {
    setActionLoading(`${userId}-${key}`)
    try {
      await fn()
      showMsg(`Done.`)
      await loadAll()
    } catch (err) {
      showMsg(err.response?.data?.error || 'Action failed.', 'error')
    } finally {
      setActionLoading(null)
    }
  }

  const handleSuspend = async () => {
    if (!suspendModal) return
    setActionLoading(`${suspendModal.id}-suspend`)
    try {
      await suspendUser(suspendModal.id, suspendReason)
      setSuspendModal(null)
      setSuspendReason('')
      showMsg(`${suspendModal.username} suspended.`)
      await loadAll()
    } catch (err) {
      showMsg(err.response?.data?.error || 'Failed to suspend user.', 'error')
    } finally {
      setActionLoading(null)
    }
  }

  const handleUpdateEmail = async () => {
    if (!emailModal || !emailValue.trim()) return
    setActionLoading(`${emailModal.id}-email`)
    try {
      await updateUserEmail(emailModal.id, emailValue.trim())
      setEmailModal(null)
      setEmailValue('')
      showMsg('Email updated.')
      await loadAll()
    } catch (err) {
      showMsg(err.response?.data?.error || 'Failed to update email.', 'error')
    } finally {
      setActionLoading(null)
    }
  }

  const handleGdprDelete = async () => {
    if (!gdprConfirm) return
    try {
      await gdprDeleteUser(gdprConfirm.id)
      setGdprConfirm(null)
      showMsg(`${gdprConfirm.username} deleted (GDPR).`)
      await loadAll()
    } catch (err) {
      showMsg(err.response?.data?.error || 'GDPR delete failed.', 'error')
    }
  }

  const handleBulkAction = async () => {
    const ids = [...selected]
    if (!ids.length || !bulkAction) return
    try {
      if (bulkAction === 'suspend') await bulkSuspendUsers(ids, bulkSuspendReason)
      else if (bulkAction === 'activate') await bulkActivateUsers(ids)
      else if (bulkAction === 'delete') {
        if (!confirm(`Delete ${ids.length} user(s)? This cannot be undone.`)) return
        await bulkDeleteUsers(ids)
      }
      setSelected(new Set())
      setBulkAction('')
      setBulkSuspendReason('')
      showMsg(`Bulk ${bulkAction} applied to ${ids.length} user(s).`)
      await loadAll()
    } catch (err) {
      showMsg(err.response?.data?.error || 'Bulk action failed.', 'error')
    }
  }

  const toggleSelect = (id) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleSelectAll = () => {
    if (selected.size === users.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(users.map((u) => u.id)))
    }
  }

  if (loading) return <PageSpinner message="Loading users..." />
  if (permissionDenied) return <PermissionDenied />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Users &amp; Roles</h1>
        {activeTab === 'users' && (
          <button
            onClick={() => { setCreateForm(CREATE_EMPTY_FORM); setCreateError(''); setCreateModal(true) }}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
          >
            + Create User
          </button>
        )}
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 mb-6 border-b border-gray-200 dark:border-gray-700">
        <button
          onClick={() => setActiveTab('users')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'users'
              ? 'border-blue-600 text-blue-600 dark:text-blue-400 dark:border-blue-400'
              : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
          }`}
        >
          Users
        </button>
        <button
          onClick={() => setActiveTab('break-glass')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'break-glass'
              ? 'border-blue-600 text-blue-600 dark:text-blue-400 dark:border-blue-400'
              : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
          }`}
        >
          Break-Glass
        </button>
      </div>

      {activeTab === 'break-glass' && <BreakGlassTab />}

      {activeTab === 'users' && (<>
      <ErrorBanner error={error} />
      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'}`}>
          {message.text}
        </div>
      )}

      {/* Bulk action bar */}
      {selected.size > 0 && (
        <div className="mb-4 flex items-center gap-3 p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded">
          <span className="text-sm text-blue-700 dark:text-blue-300 font-medium">{selected.size} selected</span>
          <select
            value={bulkAction}
            onChange={(e) => setBulkAction(e.target.value)}
            className="text-sm border border-gray-300 rounded px-2 py-1 bg-white dark:bg-gray-700 dark:border-gray-600"
          >
            <option value="">Choose action…</option>
            <option value="suspend">Suspend</option>
            <option value="activate">Activate</option>
            <option value="delete">Delete</option>
          </select>
          {bulkAction === 'suspend' && (
            <input
              type="text"
              value={bulkSuspendReason}
              onChange={(e) => setBulkSuspendReason(e.target.value)}
              placeholder="Reason (optional)"
              className="text-sm border border-gray-300 rounded px-2 py-1 bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            />
          )}
          <button
            onClick={handleBulkAction}
            disabled={!bulkAction}
            className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 disabled:opacity-50"
          >
            Apply
          </button>
          <button onClick={() => setSelected(new Set())} className="text-sm text-gray-500 hover:text-gray-700">Clear</button>
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="px-3 py-3 w-6">
                <input type="checkbox" checked={selected.size === users.length && users.length > 0} onChange={toggleSelectAll} className="rounded" />
              </th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium w-6"></th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Username</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Email</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">System Role</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {users.length === 0 && (
              <EmptyRow colSpan={7} message="No users found." />
            )}
            {users.map(user => (
              <>
                <tr
                  key={user.id}
                  className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/30 cursor-pointer"
                  onClick={() => toggleExpand(user.id)}
                >
                  <td className="px-3 py-3" onClick={e => e.stopPropagation()}>
                    <input type="checkbox" checked={selected.has(user.id)} onChange={() => toggleSelect(user.id)} className="rounded" />
                  </td>
                  <td className="px-4 py-3 text-gray-400 text-xs">
                    {expandedUser === user.id ? '▼' : '▶'}
                  </td>
                  <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-100">{user.username}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{user.email}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-0.5 text-xs font-medium rounded ${
                      user.role === 'admin'
                        ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300'
                        : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                    }`}>
                      {user.role}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    {user.state === 'suspended' ? (
                      <span className="text-xs text-red-600 dark:text-red-400 font-medium">Suspended</span>
                    ) : user.state === 'active' ? (
                      <span className="text-xs text-green-600 dark:text-green-400 font-medium">Active</span>
                    ) : (
                      <span className="text-xs text-gray-500 dark:text-gray-400">{user.state || 'Inactive'}</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right" onClick={e => e.stopPropagation()}>
                    <div className="flex items-center justify-end gap-1.5 flex-wrap">
                      <button
                        onClick={() => openAssign(user.id)}
                        className="px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 font-medium"
                      >
                        + Role
                      </button>
                      {user.state === 'suspended' ? (
                        <button
                          onClick={() => runAction(user.id, 'unsuspend', () => unsuspendUser(user.id))}
                          disabled={actionLoading === `${user.id}-unsuspend`}
                          className="px-2 py-1 text-xs border border-green-300 text-green-700 rounded hover:bg-green-50 disabled:opacity-50"
                        >
                          Unsuspend
                        </button>
                      ) : (
                        <button
                          onClick={() => { setSuspendModal(user); setSuspendReason('') }}
                          className="px-2 py-1 text-xs border border-orange-300 text-orange-700 rounded hover:bg-orange-50"
                        >
                          Suspend
                        </button>
                      )}
                      {user.locked && (
                        <button
                          onClick={() => runAction(user.id, 'unlock', () => adminUnlockUser(user.id))}
                          disabled={actionLoading === `${user.id}-unlock`}
                          className="px-2 py-1 text-xs border border-yellow-300 text-yellow-700 rounded hover:bg-yellow-50 disabled:opacity-50"
                        >
                          Unlock
                        </button>
                      )}
                      <button
                        onClick={() => runAction(user.id, 'pwreset', () => sendPasswordResetEmail(user.id))}
                        disabled={actionLoading === `${user.id}-pwreset`}
                        className="px-2 py-1 text-xs border border-gray-300 text-gray-600 rounded hover:bg-gray-50 disabled:opacity-50"
                      >
                        Reset PW
                      </button>
                      <button
                        onClick={() => { setEmailModal(user); setEmailValue(user.email || '') }}
                        className="px-2 py-1 text-xs border border-gray-300 text-gray-600 rounded hover:bg-gray-50"
                      >
                        Email
                      </button>
                      <button
                        onClick={() => runAction(user.id, 'impersonate', async () => {
                          await impersonateUser(user.id)
                          window.location.href = '/dashboard'
                        })}
                        disabled={actionLoading === `${user.id}-impersonate`}
                        className="px-2 py-1 text-xs border border-purple-300 text-purple-700 rounded hover:bg-purple-50 disabled:opacity-50"
                      >
                        Impersonate
                      </button>
                      <button
                        onClick={() => setGdprConfirm(user)}
                        className="px-2 py-1 text-xs border border-red-300 text-red-600 rounded hover:bg-red-50"
                      >
                        GDPR Delete
                      </button>
                    </div>
                  </td>
                </tr>
                {expandedUser === user.id && (
                  <tr key={`${user.id}-roles`} className="border-b dark:border-gray-700 bg-gray-50/50 dark:bg-gray-700/20">
                    <td colSpan={7} className="px-8 py-3">
                      <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                        Assigned Custom Roles
                      </div>
                      {!userRoles[user.id] ? (
                        <p className="text-sm text-gray-400">Loading...</p>
                      ) : userRoles[user.id].length === 0 ? (
                        <p className="text-sm text-gray-400">No custom roles assigned.</p>
                      ) : (
                        <div className="space-y-1">
                          {userRoles[user.id].map(ur => (
                            <div key={ur.id} className="flex items-center justify-between gap-3 py-1.5 px-3 bg-white dark:bg-gray-800 rounded border dark:border-gray-700">
                              <div>
                                <span className="text-sm font-medium text-gray-800 dark:text-gray-100">{ur.name || `Role #${ur.id}`}</span>
                                {ur.description && (
                                  <span className="ml-2 text-xs text-gray-400">{ur.description}</span>
                                )}
                              </div>
                                <div>
                                  {removeConfirm?.userId === user.id && removeConfirm?.roleId === ur.id ? (
                                    <span className="space-x-2">
                                      <span className="text-red-600 text-xs">Remove?</span>
                                      <button
                                        onClick={() => handleRemoveRole(user.id, ur.id)}
                                        className="text-red-600 hover:text-red-800 text-xs font-medium"
                                      >Yes</button>
                                      <button
                                        onClick={() => setRemoveConfirm(null)}
                                        className="text-gray-400 hover:text-gray-600 text-xs"
                                      >No</button>
                                    </span>
                                  ) : (
                                    <button
                                      onClick={() => setRemoveConfirm({ userId: user.id, roleId: ur.id })}
                                      className="text-gray-400 hover:text-red-600 text-xs"
                                    >Remove</button>
                                  )}
                                </div>
                              </div>
                            ))}
                        </div>
                      )}
                    </td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {createModal && (
        <Modal title="Create User" onClose={() => setCreateModal(false)}>
          <form onSubmit={handleCreateSubmit} className="space-y-4">
            {createError && <p className="text-red-600 text-sm">{createError}</p>}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={createForm.username}
                onChange={e => setCreateForm(f => ({ ...f, username: e.target.value }))}
                required
                autoComplete="off"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email <span className="text-red-500">*</span></label>
              <input
                type="email"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={createForm.email}
                onChange={e => setCreateForm(f => ({ ...f, email: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Password <span className="text-red-500">*</span></label>
              <input
                type="password"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={createForm.password}
                onChange={e => setCreateForm(f => ({ ...f, password: e.target.value }))}
                required
                autoComplete="new-password"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Role</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={createForm.role}
                onChange={e => setCreateForm(f => ({ ...f, role: e.target.value }))}
              >
                <option value="user">User</option>
                <option value="viewer">Viewer</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setCreateModal(false)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Creating...' : 'Create User'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* Suspend modal */}
      {suspendModal && (
        <Modal title={`Suspend ${suspendModal.username}`} onClose={() => setSuspendModal(null)}>
          <div className="space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">The user will be unable to log in while suspended.</p>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Reason (optional)</label>
              <input
                type="text"
                value={suspendReason}
                onChange={(e) => setSuspendReason(e.target.value)}
                className="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="e.g. Policy violation"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button onClick={() => setSuspendModal(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">Cancel</button>
              <button
                onClick={handleSuspend}
                disabled={actionLoading === `${suspendModal.id}-suspend`}
                className="px-4 py-2 text-sm bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50"
              >
                {actionLoading === `${suspendModal.id}-suspend` ? 'Suspending…' : 'Suspend User'}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Update email modal */}
      {emailModal && (
        <Modal title={`Update Email — ${emailModal.username}`} onClose={() => setEmailModal(null)}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">New email address</label>
              <input
                type="email"
                value={emailValue}
                onChange={(e) => setEmailValue(e.target.value)}
                className="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button onClick={() => setEmailModal(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">Cancel</button>
              <button
                onClick={handleUpdateEmail}
                disabled={!emailValue.trim() || actionLoading === `${emailModal.id}-email`}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
              >
                {actionLoading === `${emailModal.id}-email` ? 'Saving…' : 'Update Email'}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* GDPR delete confirmation */}
      {gdprConfirm && (
        <Modal title="GDPR Delete User" onClose={() => setGdprConfirm(null)}>
          <p className="text-sm text-gray-700 dark:text-gray-300 mb-4">
            Permanently anonymise and delete all personal data for <strong>{gdprConfirm.username}</strong>?
            This cannot be undone.
          </p>
          <div className="flex justify-end gap-2">
            <button onClick={() => setGdprConfirm(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50">Cancel</button>
            <button onClick={handleGdprDelete} className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700">
              GDPR Delete
            </button>
          </div>
        </Modal>
      )}

      {assignModal && (
        <Modal title="Assign Role to User" onClose={() => setAssignModal(null)}>
          <form onSubmit={handleAssignSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Role <span className="text-red-500">*</span>
              </label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={assignForm.role_id}
                onChange={e => setAssignForm(f => ({ ...f, role_id: e.target.value }))}
                required
              >
                <option value="">Select a role...</option>
                {roles.map(r => (
                  <option key={r.id} value={r.id}>{r.name}{r.description ? ` — ${r.description}` : ''}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Scope to Location <span className="text-gray-400 font-normal">(optional)</span>
              </label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={assignForm.location_id}
                onChange={e => setAssignForm(f => ({ ...f, location_id: e.target.value }))}
              >
                <option value="">All locations (global)</option>
                {locations.map(l => (
                  <option key={l.id} value={l.id}>{l.name}</option>
                ))}
              </select>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Restrict this role assignment to a specific location.
              </p>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setAssignModal(null)}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving || !assignForm.role_id}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? 'Assigning...' : 'Assign Role'}
              </button>
            </div>
          </form>
        </Modal>
      )}
      </>)}
    </div>
  )
}
