import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate, useLocation, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { getNetwork, getSubnetsPaginated, getSubnetTree } from '../api/ipam'
import { getNameservers } from '../api/dns'
import { getVlans } from '../api/vlans'
import { getCustomFields } from '../api/admin'
import SubnetTree from '../components/SubnetTree'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import { getLocations } from '../api/locations'
import { downloadFile } from '../utils/download'
import { getCachedUser } from '../utils/storageKeys'
import { useToast } from '../context/ToastContext'
import SplitSubnetModal from './subnet/SplitSubnetModal'
import MergeSubnetModal from './subnet/MergeSubnetModal'
import ResizeSubnetModal from './subnet/ResizeSubnetModal'
import SubnetFormModal from './subnet/SubnetFormModal'
import SubnetTable from './subnet/SubnetTable'
import { useSubnetModals } from './subnet/useSubnetModals'
import { useSubnetSearch } from './subnet/useSubnetSearch'

const DEFAULT_LIMIT = 25

export default function SubnetsPage() {
  const { t } = useTranslation()
  const { networkID } = useParams()
  const navigate = useNavigate()
  const location = useLocation()
  const showToast = useToast()

  const user = getCachedUser()
  const isAdmin = user?.role === 'admin'

  // Core list state
  const [network, setSection] = useState(null)
  const [subnets, setSubnets] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [cfDefs, setCfDefs] = useState([])
  const [locations, setLocations] = useState([])
  const [nameservers, setNameservers] = useState([])
  const [vlans, setVlans] = useState([])
  const [viewMode, setViewMode] = useState('list')
  const [treeData, setTreeData] = useState([])
  const [treeLoading, setTreeLoading] = useState(false)
  const [downloading, setDownloading] = useState(false)

  const loadLocations = useCallback(async () => {
    try {
      const data = await getLocations()
      setLocations(Array.isArray(data) ? data : (data?.locations ?? []))
    } catch {}
  }, [])

  const loadNameservers = useCallback(async () => {
    try {
      const res = await getNameservers()
      const data = res.data
      setNameservers(Array.isArray(data) ? data : (data?.nameservers ?? []))
    } catch {}
  }, [])

  const loadVlans = useCallback(async () => {
    try {
      const res = await getVlans()
      const data = res.data
      setVlans(Array.isArray(data) ? data : (data?.vlans ?? []))
    } catch {}
  }, [])

  const loadCfDefs = useCallback(async () => {
    try {
      const res = await getCustomFields('subnet')
      setCfDefs(Array.isArray(res.data) ? res.data : [])
    } catch {}
  }, [])

  const load = useCallback(async (p) => {
    try {
      setLoading(true)
      const [secRes, subRes] = await Promise.all([
        getNetwork(networkID),
        getSubnetsPaginated(networkID, p, DEFAULT_LIMIT),
      ])
      setSection(secRes.data)
      const data = subRes.data
      setSubnets(data.data ?? data)
      setTotal(data.total ?? (Array.isArray(data) ? data.length : 0))
    } catch {
      setError(t('subnets.loadError'))
    } finally {
      setLoading(false)
    }
  }, [networkID]) // eslint-disable-line react-hooks/exhaustive-deps

  const loadTree = useCallback(async () => {
    try {
      setTreeLoading(true)
      const res = await getSubnetTree(networkID)
      setTreeData(res.data)
    } catch {
      setError(t('subnets.loadTreeError'))
    } finally {
      setTreeLoading(false)
    }
  }, [networkID])

  const modals = useSubnetModals({
    networkID, load, loadTree, page, viewMode, showToast, setError,
  })

  const search = useSubnetSearch({
    networkID, cfDefs, load, setError, setSubnets, setTotal,
  })

  useEffect(() => {
    setPage(1)
    setSubnets([])
    search.setIsSearchActive(false)
    search.setSearchQuery('')
    load(1)
    loadCfDefs()
    loadLocations()
    loadNameservers()
    loadVlans()
  }, [networkID, location.key, load, loadCfDefs, loadLocations, loadNameservers, loadVlans]) // eslint-disable-line react-hooks/exhaustive-deps

  function handleViewMode(mode) {
    setViewMode(mode)
    if (mode === 'tree' && treeData.length === 0) loadTree()
  }

  function handlePageChange(newPage) {
    setPage(newPage)
    load(newPage)
  }

  async function handleExportSubnets() {
    setDownloading(true)
    try {
      await downloadFile(`/api/v1/admin/reports/export/subnets?format=csv`, `subnets-network-${networkID}.csv`)
    } catch {
      setError(t('subnets.exportError'))
    } finally {
      setDownloading(false)
    }
  }

  if (loading) return <PageSpinner message={t('subnets.loadingSubnets')} />

  return (
    <div>
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1">
        <Link to="/networks" className="hover:text-blue-600">{t('nav.networks')}</Link>
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-medium">{network?.name}</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('dashboard.subnets')}</h1>
        <div className="flex items-center gap-2">
          <div className="flex rounded overflow-hidden border border-gray-300 dark:border-gray-600">
            <button
              onClick={() => handleViewMode('list')}
              className={`px-3 py-1.5 text-sm font-medium transition ${viewMode === 'list' ? 'bg-blue-600 text-white' : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'}`}
            >{t('subnets.list')}</button>
            <button
              onClick={() => handleViewMode('tree')}
              className={`px-3 py-1.5 text-sm font-medium transition border-l border-gray-300 dark:border-gray-600 ${viewMode === 'tree' ? 'bg-blue-600 text-white' : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'}`}
            >{t('subnets.tree')}</button>
          </div>
          <button onClick={handleExportSubnets} disabled={downloading} className="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-sm disabled:opacity-50">
            {downloading ? t('networks.exporting') : t('networks.exportCsv')}
          </button>
          <button onClick={modals.openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
            {t('subnets.newSubnet')}
          </button>
        </div>
      </div>

      <ErrorBanner error={error} />

      {viewMode === 'list' && (
        <>
          <div className="mb-4 space-y-2">
            <form onSubmit={search.handleSearch} className="flex gap-2 flex-wrap">
              <input
                type="text"
                placeholder={t('subnets.searchPlaceholder')}
                value={search.searchQuery}
                onChange={e => search.setSearchQuery(e.target.value)}
                className="flex-1 min-w-40 border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
              />
              {locations.length > 0 && (
                <select
                  value={search.filterLocationId}
                  onChange={e => search.setFilterLocationId(e.target.value)}
                  className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
                >
                  <option value="">{t('subnets.allLocations')}</option>
                  {locations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
                </select>
              )}
              {search.searchableFields.length > 0 && (
                <button type="button" onClick={search.addCfFilterRow} className="px-3 py-2 text-sm border rounded hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-300">
                  {t('subnets.addFilter')}
                </button>
              )}
              <button type="submit" disabled={search.searching} className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm font-medium disabled:opacity-50">
                {search.searching ? t('networks.searching') : t('header.search')}
              </button>
              {(search.isSearchActive || search.cfFilterRows.length > 0 || search.filterLocationId) && (
                <button type="button" onClick={search.handleClearSearch} className="px-4 py-2 bg-gray-400 text-white rounded hover:bg-gray-500 text-sm font-medium">
                  {t('common.clear')}
                </button>
              )}
            </form>
            {search.cfFilterRows.map((row, idx) => (
              <div key={idx} className="flex gap-2 items-center">
                <select value={row.field} onChange={e => search.updateCfFilterRow(idx, { field: e.target.value })} className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100">
                  {search.searchableFields.map(d => <option key={d.name} value={d.name}>{d.label}</option>)}
                </select>
                <select value={row.op} onChange={e => search.updateCfFilterRow(idx, { op: e.target.value })} className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100">
                  <option value="is">{t('subnets.is')}</option>
                  <option value="contains">{t('subnets.contains')}</option>
                  <option value="is not">{t('subnets.isNot')}</option>
                </select>
                <input type="text" value={row.value} onChange={e => search.updateCfFilterRow(idx, { value: e.target.value })} placeholder={t('subnets.valuePlaceholder')} className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100" />
                <button type="button" onClick={() => search.removeCfFilterRow(idx)} className="text-gray-400 hover:text-red-600 text-sm px-1">&times;</button>
              </div>
            ))}
          </div>

          <SubnetTable
            subnets={subnets}
            total={total}
            isSearchActive={search.isSearchActive}
            page={page}
            defaultLimit={DEFAULT_LIMIT}
            sortCol={search.sortCol}
            sortDir={search.sortDir}
            onSort={search.handleSort}
            sortedSubnets={search.sortedSubnets}
            searchableFields={search.searchableFields}
            locations={locations}
            vlans={vlans}
            isAdmin={isAdmin}
            deleteConfirm={modals.deleteConfirm}
            onDeleteConfirm={modals.setDeleteConfirm}
            onDeleteCancel={() => modals.setDeleteConfirm(null)}
            onDelete={modals.handleDelete}
            onEdit={modals.openEdit}
            onSplit={modals.openSplit}
            onMerge={modals.openMerge}
            onResize={modals.openResize}
            onPageChange={handlePageChange}
            onNavigate={navigate}
            addCfFilterFromValue={search.addCfFilterFromValue}
          />
        </>
      )}

      {viewMode === 'tree' && (
        <>
          {treeLoading ? (
            <p className="text-gray-500">{t('subnets.loadingTree')}</p>
          ) : (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                    <tr>
                      <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('subnets.network')}</th>
                      <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.description')}</th>
                      <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('subnets.usedTotal')}</th>
                      <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('dashboard.utilisation')}</th>
                      <th className="px-4 py-3"></th>
                    </tr>
                  </thead>
                  <tbody>
                    <SubnetTree nodes={treeData} onEdit={modals.openEdit} onDelete={modals.handleDelete} navigate={navigate} />
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      )}

      {modals.splitModal && (
        <SplitSubnetModal
          splitModal={modals.splitModal}
          splitPrefix={modals.splitPrefix}
          setSplitPrefix={modals.setSplitPrefix}
          splitting={modals.splitting}
          splitError={modals.splitError}
          splitBlockingIPs={modals.splitBlockingIPs}
          onSplit={modals.handleSplit}
          onClose={() => modals.setSplitModal(null)}
        />
      )}

      {modals.mergeModal && (
        <MergeSubnetModal
          mergeModal={modals.mergeModal}
          mergeSelected={modals.mergeSelected}
          setMergeSelected={modals.setMergeSelected}
          merging={modals.merging}
          mergeError={modals.mergeError}
          onMerge={modals.handleMerge}
          onClose={() => modals.setMergeModal(null)}
        />
      )}

      {modals.resizeModal && (
        <ResizeSubnetModal
          resizeModal={modals.resizeModal}
          resizePrefix={modals.resizePrefix}
          setResizePrefix={modals.setResizePrefix}
          resizing={modals.resizing}
          resizeError={modals.resizeError}
          setResizeError={modals.setResizeError}
          resizeConfirmText={modals.resizeConfirmText}
          setResizeConfirmText={modals.setResizeConfirmText}
          onResize={modals.handleResize}
          onClose={() => modals.setResizeModal(null)}
        />
      )}

      {modals.modal && (
        <SubnetFormModal
          modal={modals.modal}
          form={modals.form}
          setForm={modals.setForm}
          overlapError={modals.overlapError}
          saving={modals.saving}
          locations={locations}
          nameservers={nameservers}
          vlans={vlans}
          cfDefs={cfDefs}
          onSubmit={modals.handleSubmit}
          onClose={() => modals.setModal(null)}
        />
      )}
    </div>
  )
}
