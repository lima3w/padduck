import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import Modal from '../components/Modal'
import { getAdminRoles, createRole, updateRole, deleteRole, addPermissionToRole, removePermissionFromRole, listAvailablePermissions } from '../api/admin'

// Role model has no JSON tags — PascalCase: ID, Name, Description, IsSystem, Permissions, CreatedAt, UpdatedAt
// RolePermission: ID, RoleID, Permission, ResourceType (*string), ResourceID (*int64), CreatedAt

export default function AdminRolesPage() {
  const { t } = useTranslation()
  const [roles, setRoles] = useState([])
  const [allPerms, setAllPerms] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [message, setMessage] = useState(null)

  // Role CRUD modal
  const [roleModal, setRoleModal] = useState(null) // null | 'create' | { edit: role }
  const [roleForm, setRoleForm] = useState({ Name: '', Description: '' })
  const [roleSaving, setRoleSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  // Permission editor
  const [editPermsFor, setEditPermsFor] = useState(null) // role object
  const [permFilter, setPermFilter] = useState('')
  const [addingPerm, setAddingPerm] = useState(null) // permission string being added

  useEffect(() => {
    Promise.all([loadRoles(), loadPerms()])
  }, [])

  async function loadRoles() {
    try {
      setLoading(true)
      setError(null)
      const res = await getAdminRoles()
      setRoles(Array.isArray(res.data) ? res.data : [])
    } catch (err) {
      setError(err.response?.data?.error || t('adminRoles.loadRolesFailed'))
    } finally {
      setLoading(false)
    }
  }

  async function loadPerms() {
    try {
      const res = await listAvailablePermissions()
      setAllPerms(Array.isArray(res.data) ? res.data : [])
    } catch {
      // Non-critical
    }
  }

  const showMessage = (text, type = 'success') => {
    setMessage({ text, type })
    setTimeout(() => setMessage(null), 4000)
  }

  const openCreate = () => {
    setRoleForm({ Name: '', Description: '' })
    setRoleModal('create')
  }

  const openEdit = (role) => {
    setRoleForm({ Name: role.Name, Description: role.Description || '' })
    setRoleModal({ edit: role })
  }

  const closeRoleModal = () => { setRoleModal(null); setRoleForm({ Name: '', Description: '' }) }

  const handleSaveRole = async () => {
    if (!roleForm.Name.trim()) return
    setRoleSaving(true)
    try {
      const payload = { name: roleForm.Name.trim(), description: roleForm.Description.trim() }
      if (roleModal === 'create') {
        await createRole(payload)
        showMessage(t('adminRoles.created'))
      } else {
        await updateRole(roleModal.edit.ID, payload)
        showMessage(t('adminRoles.updated'))
        if (editPermsFor?.ID === roleModal.edit.ID) {
          setEditPermsFor((prev) => ({ ...prev, Name: roleForm.Name.trim(), Description: roleForm.Description.trim() }))
        }
      }
      closeRoleModal()
      await loadRoles()
    } catch (err) {
      showMessage(err.response?.data?.error || t('adminRoles.saveFailed'), 'error')
    } finally {
      setRoleSaving(false)
    }
  }

  const handleDeleteRole = async (role) => {
    try {
      await deleteRole(role.ID)
      setDeleteConfirm(null)
      if (editPermsFor?.ID === role.ID) setEditPermsFor(null)
      showMessage(t('adminRoles.deleted'))
      await loadRoles()
    } catch (err) {
      showMessage(err.response?.data?.error || t('adminRoles.deleteFailed'), 'error')
    }
  }

  const openPermEditor = (role) => {
    setEditPermsFor(role)
    setPermFilter('')
    setAddingPerm(null)
  }

  const handleAddPerm = async (permission) => {
    setAddingPerm(permission)
    try {
      const res = await addPermissionToRole(editPermsFor.ID, { permission })
      const newPerm = res.data
      setEditPermsFor((prev) => ({ ...prev, Permissions: [...(prev.Permissions || []), newPerm] }))
      setRoles((prev) =>
        prev.map((r) =>
          r.ID === editPermsFor.ID
            ? { ...r, Permissions: [...(r.Permissions || []), newPerm] }
            : r
        )
      )
    } catch (err) {
      showMessage(err.response?.data?.error || t('adminRoles.addPermFailed'), 'error')
    } finally {
      setAddingPerm(null)
    }
  }

  const handleRemovePerm = async (perm) => {
    try {
      await removePermissionFromRole(editPermsFor.ID, perm.ID)
      setEditPermsFor((prev) => ({
        ...prev,
        Permissions: (prev.Permissions || []).filter((p) => p.ID !== perm.ID),
      }))
      setRoles((prev) =>
        prev.map((r) =>
          r.ID === editPermsFor.ID
            ? { ...r, Permissions: (r.Permissions || []).filter((p) => p.ID !== perm.ID) }
            : r
        )
      )
    } catch (err) {
      showMessage(err.response?.data?.error || t('adminRoles.removePermFailed'), 'error')
    }
  }

  const assignedPerms = new Set((editPermsFor?.Permissions || []).map((p) => p.Permission))
  const filteredAvail = allPerms.filter(
    (p) => !assignedPerms.has(p) && p.toLowerCase().includes(permFilter.toLowerCase())
  )

  const inputClass = 'w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100'

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('adminRoles.title')}</h1>
        <button
          onClick={openCreate}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-blue-700 transition"
        >
          {t('adminRoles.createRole')}
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded text-sm ${message.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'}`}>
          {message.text}
        </div>
      )}

      {loading ? (
        <p className="text-sm text-gray-500">{t('common.loading')}</p>
      ) : error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
        <div className="space-y-3">
          {roles.map((role) => (
            <div key={role.ID} className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-semibold text-gray-900 dark:text-gray-100">{role.Name}</span>
                    {role.IsSystem && (
                      <span className="text-xs px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 border border-gray-200 dark:border-gray-600 rounded">
                        {t('adminRoles.systemBadge')}
                      </span>
                    )}
                    <span className="text-xs text-gray-500">
                      {t('adminRoles.permissionsCount', { count: (role.Permissions || []).length })}
                    </span>
                  </div>
                  {role.Description && (
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{role.Description}</p>
                  )}
                  {(role.Permissions || []).length > 0 && (
                    <div className="mt-2 flex flex-wrap gap-1">
                      {(role.Permissions || []).slice(0, 8).map((p) => (
                        <span key={p.ID} className="text-xs px-1.5 py-0.5 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 border border-blue-100 dark:border-blue-800 rounded font-mono">
                          {p.Permission}
                        </span>
                      ))}
                      {(role.Permissions || []).length > 8 && (
                        <span className="text-xs text-gray-400">{t('adminRoles.moreCount', { count: (role.Permissions || []).length - 8 })}</span>
                      )}
                    </div>
                  )}
                </div>
                <div className="flex gap-2 flex-shrink-0">
                  <button
                    onClick={() => openPermEditor(role)}
                    className="text-xs text-blue-600 border border-blue-200 px-2.5 py-1 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20 transition"
                  >
                    {t('adminRoles.permissionsButton')}
                  </button>
                  {!role.IsSystem && (
                    <>
                      <button
                        onClick={() => openEdit(role)}
                        className="text-xs text-gray-600 border border-gray-200 px-2.5 py-1 rounded hover:bg-gray-50 transition"
                      >
                        {t('common.edit')}
                      </button>
                      <button
                        onClick={() => setDeleteConfirm(role)}
                        className="text-xs text-red-600 border border-red-200 px-2.5 py-1 rounded hover:bg-red-50 transition"
                      >
                        {t('common.delete')}
                      </button>
                    </>
                  )}
                </div>
              </div>
            </div>
          ))}

          {roles.length === 0 && (
            <div className="text-center py-12 text-gray-500">{t('adminRoles.noRolesFound')}</div>
          )}
        </div>
      )}

      {/* Role create/edit modal */}
      {roleModal && (
        <Modal
          title={roleModal === 'create' ? t('adminRoles.createRoleModalTitle') : t('adminRoles.editRoleModalTitle')}
          onClose={closeRoleModal}
        >
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('common.name')} <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={roleForm.Name}
                onChange={(e) => setRoleForm((p) => ({ ...p, Name: e.target.value }))}
                className={inputClass}
                placeholder={t('adminRoles.namePlaceholder')}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('common.description')}
              </label>
              <input
                type="text"
                value={roleForm.Description}
                onChange={(e) => setRoleForm((p) => ({ ...p, Description: e.target.value }))}
                className={inputClass}
                placeholder={t('adminRoles.descriptionPlaceholder')}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={closeRoleModal} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition">{t('common.cancel')}</button>
              <button
                onClick={handleSaveRole}
                disabled={roleSaving || !roleForm.Name.trim()}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition"
              >
                {roleSaving ? t('common.saving') : roleModal === 'create' ? t('vrfs.create') : t('common.save')}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Delete confirmation */}
      {deleteConfirm && (
        <Modal title={t('adminRoles.deleteRoleModalTitle')} onClose={() => setDeleteConfirm(null)}>
          <p className="text-sm text-gray-700 dark:text-gray-300 mb-4">
            {t('adminRoles.confirmDeletePrefix')}<strong>{deleteConfirm.Name}</strong>{t('adminRoles.confirmDeleteSuffix')}
          </p>
          <div className="flex justify-end gap-2">
            <button onClick={() => setDeleteConfirm(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition">{t('common.cancel')}</button>
            <button onClick={() => handleDeleteRole(deleteConfirm)} className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700 transition">{t('common.delete')}</button>
          </div>
        </Modal>
      )}

      {/* Permission editor panel */}
      {editPermsFor && (
        <Modal
          title={`${t('adminRoles.permissionsModalTitlePrefix')}${editPermsFor.Name}`}
          onClose={() => setEditPermsFor(null)}
        >
          <div className="space-y-4 min-h-[300px]">
            {/* Current permissions */}
            <div>
              <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{t('adminRoles.assignedLabel')}</p>
              {(editPermsFor.Permissions || []).length === 0 ? (
                <p className="text-sm text-gray-400">{t('adminRoles.noPermissionsAssigned')}</p>
              ) : (
                <div className="flex flex-wrap gap-1.5 max-h-36 overflow-y-auto">
                  {(editPermsFor.Permissions || []).map((p) => (
                    <span key={p.ID} className="inline-flex items-center gap-1 text-xs px-2 py-0.5 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-700 rounded font-mono">
                      {p.Permission}
                      {!editPermsFor.IsSystem && (
                        <button
                          onClick={() => handleRemovePerm(p)}
                          className="text-blue-400 hover:text-red-500 ml-0.5 leading-none"
                          title={t('adminRoles.removeTitle')}
                        >
                          ×
                        </button>
                      )}
                    </span>
                  ))}
                </div>
              )}
            </div>

            {/* Add permissions */}
            {!editPermsFor.IsSystem && (
              <div>
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{t('adminRoles.addPermissionLabel')}</p>
                <input
                  type="text"
                  value={permFilter}
                  onChange={(e) => setPermFilter(e.target.value)}
                  className={inputClass}
                  placeholder={t('adminRoles.filterPermissionsPlaceholder')}
                />
                <div className="mt-2 max-h-48 overflow-y-auto border border-gray-200 dark:border-gray-600 rounded divide-y divide-gray-100 dark:divide-gray-700">
                  {filteredAvail.length === 0 ? (
                    <p className="text-xs text-gray-400 p-3">{permFilter ? t('adminRoles.noMatchesPeriod') : t('adminRoles.allPermissionsAssigned')}</p>
                  ) : (
                    filteredAvail.map((p) => (
                      <div key={p} className="flex items-center justify-between px-3 py-1.5 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                        <span className="text-xs font-mono text-gray-700 dark:text-gray-300">{p}</span>
                        <button
                          onClick={() => handleAddPerm(p)}
                          disabled={addingPerm === p}
                          className="text-xs text-blue-600 hover:underline disabled:opacity-50"
                        >
                          {addingPerm === p ? '…' : t('adminRoles.add')}
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}

            <div className="flex justify-end pt-2">
              <button onClick={() => setEditPermsFor(null)} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition">{t('myRequests.close')}</button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
