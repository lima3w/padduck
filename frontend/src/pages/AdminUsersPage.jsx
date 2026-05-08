import { useState, useEffect } from 'react'
import Modal from '../components/Modal'
import { getLocations } from '../api/locations'

const ASSIGN_EMPTY_FORM = { role_id: '', location_id: '' }

export default function AdminUsersPage() {
  const [users, setUsers] = useState([])
  const [roles, setRoles] = useState([])
  const [locations, setLocations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [expandedUser, setExpandedUser] = useState(null)
  const [userRoles, setUserRoles] = useState({}) // userId -> roles[]
  const [assignModal, setAssignModal] = useState(null) // userId
  const [assignForm, setAssignForm] = useState(ASSIGN_EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [removeConfirm, setRemoveConfirm] = useState(null) // { userId, roleId }

  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
  }

  useEffect(() => {
    loadAll()
  }, [])

  async function loadAll() {
    try {
      setLoading(true)
      setError(null)
      const [usersRes, rolesRes] = await Promise.all([
        fetch('/api/v1/admin/users', { headers }),
        fetch('/api/v1/admin/roles', { headers }),
      ])
      if (!usersRes.ok) throw new Error('Failed to load users')
      const usersData = await usersRes.json()
      setUsers(Array.isArray(usersData) ? usersData : (usersData?.users ?? []))
      if (rolesRes.ok) {
        const rolesData = await rolesRes.json()
        setRoles(Array.isArray(rolesData) ? rolesData : [])
      }
      const locsData = await getLocations().catch(() => [])
      setLocations(Array.isArray(locsData) ? locsData : (locsData?.locations ?? []))
    } catch (err) {
      setError(err.message || 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }

  async function loadUserRoles(userId) {
    try {
      const res = await fetch(`/api/v1/admin/users/${userId}/roles`, { headers })
      if (res.ok) {
        const data = await res.json()
        setUserRoles(prev => ({ ...prev, [userId]: Array.isArray(data) ? data : [] }))
      }
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
      const res = await fetch(`/api/v1/admin/users/${assignModal}/roles`, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
      })
      if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.error || 'Failed') }
      setAssignModal(null)
      // Reload roles for this user
      await loadUserRoles(assignModal)
    } catch (err) {
      setError(err.message || 'Failed to assign role')
    } finally {
      setSaving(false)
    }
  }

  async function handleRemoveRole(userId, roleId) {
    try {
      const res = await fetch(`/api/v1/admin/users/${userId}/roles/${roleId}`, {
        method: 'DELETE',
        headers,
      })
      if (!res.ok) throw new Error()
      setRemoveConfirm(null)
      await loadUserRoles(userId)
    } catch {
      setError('Failed to remove role')
    }
  }

  if (loading) return <p className="text-gray-500">Loading users...</p>

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Users &amp; Roles</h1>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
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
              <tr>
                <td colSpan={6} className="px-4 py-6 text-center text-gray-400">No users found</td>
              </tr>
            )}
            {users.map(user => (
              <>
                <tr
                  key={user.id}
                  className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/30 cursor-pointer"
                  onClick={() => toggleExpand(user.id)}
                >
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
                    {user.is_suspended ? (
                      <span className="text-xs text-red-600 dark:text-red-400 font-medium">Suspended</span>
                    ) : user.is_active !== false ? (
                      <span className="text-xs text-green-600 dark:text-green-400 font-medium">Active</span>
                    ) : (
                      <span className="text-xs text-gray-500 dark:text-gray-400">Inactive</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right" onClick={e => e.stopPropagation()}>
                    <button
                      onClick={() => openAssign(user.id)}
                      className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 font-medium"
                    >
                      + Assign Role
                    </button>
                  </td>
                </tr>
                {expandedUser === user.id && (
                  <tr key={`${user.id}-roles`} className="border-b dark:border-gray-700 bg-gray-50/50 dark:bg-gray-700/20">
                    <td colSpan={6} className="px-8 py-3">
                      <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                        Assigned Custom Roles
                      </div>
                      {!userRoles[user.id] ? (
                        <p className="text-sm text-gray-400">Loading...</p>
                      ) : userRoles[user.id].length === 0 ? (
                        <p className="text-sm text-gray-400">No custom roles assigned.</p>
                      ) : (
                        <div className="space-y-1">
                          {userRoles[user.id].map(ur => {
                            const locName = ur.location_id
                              ? (locations.find(l => l.id === ur.location_id)?.name || `Location #${ur.location_id}`)
                              : null
                            return (
                              <div key={ur.id} className="flex items-center justify-between gap-3 py-1.5 px-3 bg-white dark:bg-gray-800 rounded border dark:border-gray-700">
                                <div>
                                  <span className="text-sm font-medium text-gray-800 dark:text-gray-100">{ur.role?.name || `Role #${ur.role_id}`}</span>
                                  {locName && (
                                    <span className="ml-2 text-xs text-blue-600 dark:text-blue-400">
                                      scoped to: {locName}
                                    </span>
                                  )}
                                  {ur.role?.description && (
                                    <span className="ml-2 text-xs text-gray-400">{ur.role.description}</span>
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
                            )
                          })}
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
    </div>
  )
}
