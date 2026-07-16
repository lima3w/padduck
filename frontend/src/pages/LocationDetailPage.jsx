import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { getLocation } from '../api/locations'
import { getRacks, createRack, updateRack, deleteRack } from '../api/racks'
import { api } from '../api/client'
import Modal from '../components/Modal'
import ObjectRelationshipsPanel from '../components/ObjectRelationshipsPanel'

const RACK_EMPTY_FORM = { name: '', size_u: '42', description: '' }

export default function LocationDetailPage() {
  const { t } = useTranslation()
  const { id } = useParams()
  const [location, setLocation] = useState(null)
  const [breadcrumb, setBreadcrumb] = useState([])
  const [subnets, setSubnets] = useState([])
  const [devices, setDevices] = useState([])
  const [racks, setRacks] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [rackModal, setRackModal] = useState(null)
  const [rackForm, setRackForm] = useState(RACK_EMPTY_FORM)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [saving, setSaving] = useState(false)

  const loadSubnets = useCallback(async () => {
    try {
      const res = await api.get(`/locations/${id}/subnets`)
      setSubnets(res.data || [])
    } catch {}
  }, [id])

  const loadDevices = useCallback(async () => {
    try {
      const res = await api.get(`/locations/${id}/devices`)
      setDevices(res.data || [])
    } catch {}
  }, [id])

  const loadRacks = useCallback(async () => {
    try {
      const data = await getRacks(id)
      setRacks(Array.isArray(data) ? data : (data?.racks ?? []))
    } catch {}
  }, [id])

  const loadAll = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const loc = await getLocation(id)
      setLocation(loc)

      // Build breadcrumb by traversing parent chain
      const crumbs = []
      let current = loc
      while (current) {
        crumbs.unshift({ id: current.id, name: current.name })
        if (current.parentId) {
          try {
            current = await getLocation(current.parentId)
          } catch {
            break
          }
        } else {
          break
        }
      }
      setBreadcrumb(crumbs)

      await Promise.all([loadSubnets(), loadDevices(), loadRacks()])
    } catch (err) {
      setError(err.message || t('locationDetail.loadError'))
    } finally {
      setLoading(false)
    }
  }, [id, loadSubnets, loadDevices, loadRacks, t])

  useEffect(() => { loadAll() }, [loadAll])

  function openCreateRack() {
    setRackForm(RACK_EMPTY_FORM)
    setRackModal('create')
  }

  function openEditRack(rack) {
    setRackForm({
      name: rack.name || '',
      size_u: rack.sizeU ? String(rack.sizeU) : '42',
      description: rack.description || '',
    })
    setRackModal({ edit: rack })
  }

  async function handleRackSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        name: rackForm.name,
        location_id: parseInt(id),
        size_u: parseInt(rackForm.size_u) || 42,
        description: rackForm.description || null,
      }
      if (rackModal === 'create') {
        await createRack(body)
      } else {
        await updateRack(rackModal.edit.id, body)
      }
      setRackModal(null)
      loadRacks()
    } catch (err) {
      setError(err.message || t('locationDetail.saveRackFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteRack(rackId) {
    try {
      await deleteRack(rackId)
      setDeleteConfirm(null)
      loadRacks()
    } catch (err) {
      setError(err.message || t('locationDetail.deleteRackFailed'))
    }
  }

  if (loading) return <p className="text-gray-500">{t('locationDetail.loading')}</p>
  if (error && !location) return <p className="text-red-600">{error}</p>

  const parent = breadcrumb.length > 1 ? breadcrumb[breadcrumb.length - 2] : null
  const relationshipItems = [
    parent && {
      label: t('locationDetail.parentLocation'),
      value: parent.name,
      to: `/locations/${parent.id}`,
      description: t('locationDetail.parentDescription'),
    },
    {
      label: t('locationDetail.racksLabel'),
      value: t('locationDetail.assignedRacks'),
      count: racks.length,
      description: t('locationDetail.racksCount', { count: racks.length }),
    },
    {
      label: t('locationDetail.subnetsLabel'),
      value: t('locationDetail.assignedSubnets'),
      count: subnets.length,
      description: t('locationDetail.subnetsCount', { count: subnets.length }),
    },
    {
      label: t('locationDetail.devicesLabel'),
      value: t('locationDetail.assignedDevices'),
      count: devices.length,
      description: t('locationDetail.devicesCount', { count: devices.length }),
    },
  ]

  return (
    <div>
      {/* Breadcrumb */}
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1 flex-wrap">
        <Link to="/locations" className="hover:text-blue-600">{t('nav.locations')}</Link>
        {breadcrumb.map((crumb, i) => (
          <span key={crumb.id} className="flex items-center gap-1">
            <span>/</span>
            {i < breadcrumb.length - 1 ? (
              <Link to={`/locations/${crumb.id}`} className="hover:text-blue-600">{crumb.name}</Link>
            ) : (
              <span className="text-gray-800 dark:text-gray-200 font-medium">{crumb.name}</span>
            )}
          </span>
        ))}
      </nav>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{location?.name}</h1>
          {location?.description && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{location.description}</p>
          )}
        </div>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      <ObjectRelationshipsPanel relationships={relationshipItems} />

      {/* Location details */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div>
            <dt className="text-gray-500 dark:text-gray-400">{t('natRules.type')}</dt>
            <dd className="text-gray-800 dark:text-gray-200 font-medium capitalize">{location?.type}</dd>
          </div>
          {location?.address && (
            <div>
              <dt className="text-gray-500 dark:text-gray-400">{t('locations.address')}</dt>
              <dd className="text-gray-800 dark:text-gray-200">{location.address}</dd>
            </div>
          )}
        </dl>
      </div>

      {/* Racks */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">{t('locationDetail.racksLabel')}</h2>
          <button onClick={openCreateRack} className="px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
            {t('racks.addRack')}
          </button>
        </div>
        {racks.length === 0 ? (
          <p className="text-sm text-gray-400">{t('locationDetail.noRacks')}</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.name')}</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('racks.sizeU')}</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('locationDetail.utilization')}</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {racks.map(rack => {
                  const usedU = rack.usedU ?? 0
                  const sizeU = rack.sizeU ?? 42
                  const pct = sizeU > 0 ? Math.round((usedU / sizeU) * 100) : 0
                  return (
                    <tr key={rack.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                      <td className="px-4 py-3">
                        <Link to={`/racks/${rack.id}`} className="text-blue-600 dark:text-blue-400 hover:underline font-medium">
                          {rack.name}
                        </Link>
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{sizeU}U</td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <div className="flex-1 bg-gray-200 dark:bg-gray-600 rounded-full h-2 max-w-32">
                            <div
                              className={`h-2 rounded-full ${pct > 80 ? 'bg-red-500' : pct > 60 ? 'bg-yellow-500' : 'bg-green-500'}`}
                              style={{ width: `${pct}%` }}
                            ></div>
                          </div>
                          <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                            {usedU}/{sizeU}U ({pct}%)
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-right space-x-2">
                        <button onClick={() => openEditRack(rack)} className="text-gray-400 hover:text-blue-600 text-xs">{t('common.edit')}</button>
                        {deleteConfirm === rack.id ? (
                          <>
                            <span className="text-red-600 text-xs">{t('subnets.confirmDelete')}</span>
                            <button onClick={() => handleDeleteRack(rack.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">{t('common.yes')}</button>
                            <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">{t('common.no')}</button>
                          </>
                        ) : (
                          <button onClick={() => setDeleteConfirm(rack.id)} className="text-gray-400 hover:text-red-600 text-xs">{t('common.delete')}</button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
            </div>
          </div>
        )}
      </div>

      {/* Subnets */}
      <div className="mb-6">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">{t('locationDetail.subnetsLabel')}</h2>
        {subnets.length === 0 ? (
          <p className="text-sm text-gray-400">{t('locationDetail.noSubnets')}</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('locationDetail.network')}</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.description')}</th>
                </tr>
              </thead>
              <tbody>
                {subnets.map(s => (
                  <tr key={s.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-mono font-medium">
                      <Link
                        to={`/subnets/${s.id}/ip-addresses`}
                        className="text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        {s.networkAddress}/{s.prefixLength}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{s.description || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        )}
      </div>

      {/* Devices */}
      <div className="mb-6">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">{t('locationDetail.devicesLabel')}</h2>
        {devices.length === 0 ? (
          <p className="text-sm text-gray-400">{t('locationDetail.noDevices')}</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('dashboard.hostname')}</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('deviceInfo.type')}</th>
                  <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('delegations.status')}</th>
                </tr>
              </thead>
              <tbody>
                {devices.map(d => (
                  <tr key={d.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                    <td className="px-4 py-3 font-medium">
                      <Link to={`/devices/${d.id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                        {d.hostname}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{d.type?.name || '—'}</td>
                    <td className="px-4 py-3">
                      <span className="flex items-center gap-1.5 text-xs font-medium">
                        <span className={`w-2 h-2 rounded-full ${d.isOnline ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                        <span className={d.isOnline ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}>
                          {d.isOnline ? t('deviceInfo.online') : t('deviceInfo.offline')}
                        </span>
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        )}
      </div>

      {/* Rack modal */}
      {rackModal && (
        <Modal
          title={rackModal === 'create' ? t('locationDetail.addRackModalTitle') : t('locationDetail.editRackModalTitle')}
          onClose={() => setRackModal(null)}
        >
          <form onSubmit={handleRackSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {t('common.name')} <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder={t('locationDetail.rackNamePlaceholder')}
                value={rackForm.name}
                onChange={e => setRackForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('racks.sizeU')}</label>
              <input
                type="number"
                min="1"
                max="100"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={rackForm.size_u}
                onChange={e => setRackForm(f => ({ ...f, size_u: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.description')}</label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={2}
                value={rackForm.description}
                onChange={e => setRackForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setRackModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
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
