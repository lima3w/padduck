import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { getNetworksPaginated, createNetwork, updateNetwork, deleteNetwork, searchNetworks } from '../api/ipam'
import { submitSubnetRequest } from '../api/requests'
import Modal from '../components/Modal'
import Pagination from '../components/Pagination'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'
import { downloadFile } from '../utils/download'
import { getCachedUser, STORAGE_KEYS } from '../utils/storageKeys'
import { loadPrefs, savePrefs } from '../utils/listPrefs'
import SortTh from '../components/SortTh'

const DEFAULT_LIMIT = 25
const SORT_KEY = STORAGE_KEYS.networkSort

const SUBNET_REQUEST_EMPTY = { network_id: '', prefix_length: '24', purpose: '', parent_subnet_id: '' }

export default function NetworksPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const user = getCachedUser()
  const canCreateSubnet = user?.role === 'admin'

  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [sortCol, setSortCol] = useState(() => loadPrefs(SORT_KEY, { col: 'name', dir: 'asc' }).col)
  const [sortDir, setSortDir] = useState(() => loadPrefs(SORT_KEY, { col: 'name', dir: 'asc' }).dir)
  const [searchQuery, setSearchQuery] = useState('')
  const [searching, setSearching] = useState(false)
  const [searchResults, setSearchResults] = useState(null) // null = not searching
  const [actionError, setActionError] = useState(null)
  const [modal, setModal] = useState(null) // null | 'create' | { edit: network } | { requestSubnet: network|null }
  const [form, setForm] = useState({ name: '', description: '' })
  const [subnetReqForm, setSubnetReqForm] = useState(SUBNET_REQUEST_EMPTY)
  const [subnetReqError, setSubnetReqError] = useState(null)
  const [subnetReqSuccess, setSubnetReqSuccess] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [requestSubmitting, setRequestSubmitting] = useState(false)
  const [downloading, setDownloading] = useState(false)

  const listQuery = useQuery({
    queryKey: ['networks', page],
    queryFn: () => getNetworksPaginated(page, DEFAULT_LIMIT).then(r => r.data),
    placeholderData: keepPreviousData,
  })

  useEffect(() => { savePrefs(SORT_KEY, { col: sortCol, dir: sortDir }) }, [sortCol, sortDir])

  function handleSort(col) {
    if (col === sortCol) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortCol(col)
      setSortDir('asc')
    }
  }

  const isSearchActive = searchResults !== null
  const listData = listQuery.data
  const rawNetworks = isSearchActive
    ? searchResults.items
    : (listData?.data ?? (Array.isArray(listData) ? listData : []))
  const networks = [...rawNetworks].sort((a, b) => {
    const av = (a[sortCol] ?? '').toLowerCase()
    const bv = (b[sortCol] ?? '').toLowerCase()
    return sortDir === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
  })
  const total = isSearchActive
    ? searchResults.total
    : (listData?.total ?? (Array.isArray(listData) ? listData.length : 0))
  const loading = listQuery.isLoading
  const error = actionError ?? (listQuery.isError ? t('networks.loadError') : null)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['networks'] })

  const saveMutation = useMutation({
    mutationFn: ({ id, body }) => (id ? updateNetwork(id, body) : createNetwork(body)),
    onSuccess: () => {
      setModal(null)
      invalidate()
    },
    onError: () => setActionError(t('networks.saveError')),
  })

  const deleteMutation = useMutation({
    mutationFn: (id) => deleteNetwork(id),
    onSuccess: () => {
      setDeleteConfirm(null)
      invalidate()
    },
    onError: () => setActionError(t('networks.deleteError')),
  })

  const saving = saveMutation.isPending || requestSubmitting

  async function handleExport() {
    setDownloading(true)
    try { await downloadFile('/api/v1/admin/reports/export/networks', 'networks.csv') }
    catch { setActionError(t('networks.exportError')) }
    finally { setDownloading(false) }
  }

  function handlePageChange(newPage) {
    setPage(newPage)
  }

  async function handleSearch(e) {
    e.preventDefault()
    if (!searchQuery.trim()) {
      setSearchResults(null)
      return
    }
    try {
      setSearching(true)
      const res = await searchNetworks(searchQuery)
      const data = res.data
      const items = Array.isArray(data) ? data : (data.data ?? [])
      setSearchResults({ items, total: Array.isArray(data) ? data.length : (data.total ?? items.length) })
      setPage(1)
    } catch {
      setActionError(t('networks.searchError'))
    } finally {
      setSearching(false)
    }
  }

  function handleClearSearch() {
    setSearchQuery('')
    setSearchResults(null)
  }

  function openCreate() {
    setForm({ name: '', description: '' })
    setModal('create')
  }

  function openEdit(network) {
    setForm({ name: network.name, description: network.description })
    setModal({ edit: network })
  }

  function handleSubmit(e) {
    e.preventDefault()
    if (modal === 'create') {
      saveMutation.mutate({ body: { name: form.name, description: form.description, created_by: 1 } })
    } else {
      saveMutation.mutate({ id: modal.edit.id, body: { name: form.name, description: form.description } })
    }
  }

  function handleDelete(id) {
    deleteMutation.mutate(id)
  }

  function openSubnetRequest(network) {
    setSubnetReqForm({ ...SUBNET_REQUEST_EMPTY, network_id: network ? String(network.id) : '' })
    setSubnetReqError(null)
    setSubnetReqSuccess(false)
    setModal({ requestSubnet: network })
  }

  async function handleSubnetRequestSubmit(e) {
    e.preventDefault()
    setSubnetReqError(null)
    setRequestSubmitting(true)
    try {
      await submitSubnetRequest({
        network_id: subnetReqForm.network_id ? parseInt(subnetReqForm.network_id) : null,
        prefix_length: parseInt(subnetReqForm.prefix_length),
        purpose: subnetReqForm.purpose,
        parent_subnet_id: subnetReqForm.parent_subnet_id ? parseInt(subnetReqForm.parent_subnet_id) : null,
      })
      setSubnetReqSuccess(true)
      setTimeout(() => setModal(null), 1500)
    } catch (err) {
      setSubnetReqError(err.response?.data?.error || t('networks.requestError'))
    } finally {
      setRequestSubmitting(false)
    }
  }

  if (loading) return <PageSpinner message={t('networks.loadingNetworks')} />

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800">{t('nav.networks')}</h1>
        <div className="flex items-center gap-2">
          {!canCreateSubnet && (
            <button
              onClick={() => openSubnetRequest(null)}
              className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 text-sm font-medium"
            >
              {t('networks.requestSubnet')}
            </button>
          )}
          {canCreateSubnet && (
            <>
              <button onClick={handleExport} disabled={downloading} className="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-sm disabled:opacity-50">
                {downloading ? t('networks.exporting') : t('networks.exportCsv')}
              </button>
              <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
                {t('networks.newNetwork')}
              </button>
            </>
          )}
        </div>
      </div>

      <ErrorBanner error={error} />

      <div className="mb-4">
        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            type="text"
            placeholder={t('networks.searchPlaceholder')}
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="flex-1 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            disabled={searching}
            className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm font-medium disabled:opacity-50"
          >
            {searching ? t('networks.searching') : t('header.search')}
          </button>
          {isSearchActive && (
            <button
              type="button"
              onClick={handleClearSearch}
              className="px-4 py-2 bg-gray-400 text-white rounded hover:bg-gray-500 text-sm font-medium"
            >
              {t('common.clear')}
            </button>
          )}
        </form>
      </div>

      {!isSearchActive && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">
          {t('networks.count', { count: total })}
        </p>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <SortTh label={t('common.name')} col="name" sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />
              <SortTh label={t('common.description')} col="description" sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {networks.length === 0 && (
              <EmptyRow colSpan={3} message={t('networks.noNetworksYet')} />
            )}
            {networks.map(s => (
              <tr key={s.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td
                  className="px-4 py-3 font-medium text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"
                  onClick={() => navigate(`/networks/${s.id}/subnets`)}
                >
                  {s.name}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{s.description}</td>
                <td className="px-4 py-3 text-right space-x-2">
                  <Link to={`/networks/${s.id}/topology`} className="text-gray-400 hover:text-blue-600 text-xs">{t('networks.topology')}</Link>
                  {!canCreateSubnet && (
                    <button onClick={() => openSubnetRequest(s)} className="text-green-600 hover:text-green-800 text-xs font-medium">{t('networks.requestSubnet')}</button>
                  )}
                  {canCreateSubnet && (
                    <>
                      <button onClick={() => openEdit(s)} className="text-gray-400 hover:text-blue-600 text-xs">{t('common.edit')}</button>
                      {deleteConfirm === s.id ? (
                        <>
                          <span className="text-red-600 text-xs">{t('networks.confirmDelete')}</span>
                          <button onClick={() => handleDelete(s.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">{t('common.yes')}</button>
                          <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">{t('common.no')}</button>
                        </>
                      ) : (
                        <button onClick={() => setDeleteConfirm(s.id)} className="text-gray-400 hover:text-red-600 text-xs">{t('common.delete')}</button>
                      )}
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {!isSearchActive && total > DEFAULT_LIMIT && (
        <Pagination
          page={page}
          limit={DEFAULT_LIMIT}
          total={total}
          onChange={handlePageChange}
        />
      )}

      {(modal === 'create' || modal?.edit) && (
        <Modal title={modal === 'create' ? t('networks.newNetworkModalTitle') : t('networks.editNetworkModalTitle')} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="network-name" className="block text-sm font-medium text-gray-700 mb-1">{t('common.name')}</label>
              <input
                id="network-name"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label htmlFor="network-description" className="block text-sm font-medium text-gray-700 mb-1">{t('common.description')}</label>
              <input
                id="network-description"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? t('common.saving') : t('common.save')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {modal?.requestSubnet !== undefined && (
        <Modal title={t('networks.requestSubnetModalTitle')} onClose={() => setModal(null)}>
          {subnetReqSuccess ? (
            <div className="py-4 text-center text-green-600 font-medium">{t('networks.requestSubmittedSuccess')}</div>
          ) : (
            <form onSubmit={handleSubnetRequestSubmit} className="space-y-4">
              {subnetReqError && (
                <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{subnetReqError}</div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {t('networks.networkRequired')} <span className="text-red-500">*</span>
                </label>
                <select
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  value={subnetReqForm.network_id}
                  onChange={e => setSubnetReqForm(f => ({ ...f, network_id: e.target.value }))}
                  required
                >
                  <option value="">{t('networks.selectNetworkPlaceholder')}</option>
                  {networks.map(s => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {t('networks.prefixLength')} <span className="text-red-500">*</span>
                </label>
                <input
                  type="number"
                  min="8"
                  max="30"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder="24"
                  value={subnetReqForm.prefix_length}
                  onChange={e => setSubnetReqForm(f => ({ ...f, prefix_length: e.target.value }))}
                  required
                />
                <p className="text-xs text-gray-500 mt-1">{t('networks.between8and30')}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {t('networks.parentSubnetId')} <span className="text-gray-400 font-normal">{t('networks.optional')}</span>
                </label>
                <input
                  type="number"
                  min="1"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder={t('networks.parentSubnetPlaceholder')}
                  value={subnetReqForm.parent_subnet_id}
                  onChange={e => setSubnetReqForm(f => ({ ...f, parent_subnet_id: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {t('networks.purpose')} <span className="text-red-500">*</span>
                </label>
                <textarea
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  rows={3}
                  placeholder={t('networks.purposePlaceholder')}
                  value={subnetReqForm.purpose}
                  onChange={e => setSubnetReqForm(f => ({ ...f, purpose: e.target.value }))}
                  required
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
                <button type="submit" disabled={saving} className="px-4 py-2 bg-green-600 text-white rounded text-sm hover:bg-green-700 disabled:opacity-50">
                  {saving ? t('networks.submitting') : t('networks.submitRequest')}
                </button>
              </div>
            </form>
          )}
        </Modal>
      )}
    </div>
  )
}
