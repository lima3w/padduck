import { useState } from 'react'
import { searchSubnets } from '../../api/ipam'
import { loadPrefs, savePrefs } from '../../utils/listPrefs'
import { STORAGE_KEYS, LEGACY_STORAGE_KEYS } from '../../utils/storageKeys'

const FILTER_KEY = STORAGE_KEYS.subnetFilters
const LEGACY_FILTER_KEY = LEGACY_STORAGE_KEYS.subnetFilters
const SORT_KEY = STORAGE_KEYS.subnetSort

function ipToNum(addr) {
  return (addr || '').split('.').reduce((acc, p) => (acc * 256) + (Number(p) || 0), 0)
}

/**
 * Manages search, filter, and sort state for the subnet list.
 * @param {object} opts
 * @param {string|number} opts.networkID
 * @param {Array} opts.cfDefs - custom field definitions (for searchableFields)
 * @param {Function} opts.load - (page: number) => Promise
 * @param {Function} opts.setError - (msg: string|null) => void
 * @param {Function} opts.setPage - (page: number) => void
 * @param {Function} opts.setSubnets - state setter for subnet list
 * @param {Function} opts.setTotal - state setter for total count
 */
export function useSubnetSearch({ networkID, cfDefs, load, setError, setSubnets, setTotal }) {
  const [searchQuery, setSearchQuery] = useState('')
  const [searching, setSearching] = useState(false)
  const [isSearchActive, setIsSearchActive] = useState(false)
  const [cfFilterRows, setCfFilterRows] = useState([])
  const [filterLocationId, setFilterLocationId] = useState(
    () => loadPrefs(FILTER_KEY, { filterLocationId: '' }, LEGACY_FILTER_KEY).filterLocationId
  )
  const [sortCol, setSortCol] = useState(() => loadPrefs(SORT_KEY, { col: 'network', dir: 'asc' }).col)
  const [sortDir, setSortDir] = useState(() => loadPrefs(SORT_KEY, { col: 'network', dir: 'asc' }).dir)

  const searchableFields = cfDefs.filter(d => d.isSearchable)

  function handleSort(col) {
    if (col === sortCol) {
      const next = sortDir === 'asc' ? 'desc' : 'asc'
      setSortDir(next)
      savePrefs(SORT_KEY, { col, dir: next })
    } else {
      setSortCol(col)
      setSortDir('asc')
      savePrefs(SORT_KEY, { col, dir: 'asc' })
    }
  }

  function sortedSubnets(list) {
    return [...list].sort((a, b) => {
      let av, bv
      if (sortCol === 'network') {
        av = ipToNum(a.networkAddress)
        bv = ipToNum(b.networkAddress)
        return sortDir === 'asc' ? av - bv : bv - av
      }
      if (sortCol === 'prefix') {
        av = a.prefixLength ?? 0
        bv = b.prefixLength ?? 0
        return sortDir === 'asc' ? av - bv : bv - av
      }
      av = (a.description ?? '').toLowerCase()
      bv = (b.description ?? '').toLowerCase()
      return sortDir === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
    })
  }

  function addCfFilterRow() {
    if (searchableFields.length === 0) return
    setCfFilterRows(rows => [...rows, { field: searchableFields[0].name, op: 'is', value: '' }])
  }

  function updateCfFilterRow(idx, patch) {
    setCfFilterRows(rows => rows.map((r, i) => i === idx ? { ...r, ...patch } : r))
  }

  function removeCfFilterRow(idx) {
    setCfFilterRows(rows => rows.filter((_, i) => i !== idx))
  }

  function addCfFilterFromValue(fieldName, value) {
    setCfFilterRows(rows => {
      const existing = rows.findIndex(r => r.field === fieldName)
      if (existing >= 0) return rows.map((r, i) => i === existing ? { ...r, value } : r)
      return [...rows, { field: fieldName, op: 'is', value }]
    })
  }

  async function handleSearch(e) {
    e.preventDefault()
    const hasQuery = searchQuery.trim()
    const cfFilters = {}
    cfFilterRows.forEach(r => { if (r.value.trim()) cfFilters[r.field] = r.value.trim() })
    const hasCf = Object.keys(cfFilters).length > 0
    const hasLoc = Boolean(filterLocationId)
    if (!hasQuery && !hasCf && !hasLoc) {
      setIsSearchActive(false)
      load(1)
      return
    }
    try {
      setSearching(true)
      setIsSearchActive(true)
      const body = { query: searchQuery || '', limit: 100, offset: 0 }
      if (hasCf) body.custom_fields = cfFilters
      if (hasLoc) body.location_id = parseInt(filterLocationId)
      const res = await searchSubnets(networkID, body)
      const data = res.data
      const items = Array.isArray(data) ? data : (data.data ?? [])
      setSubnets(items)
      setTotal(items.length)
    } catch {
      setError('Failed to search subnets')
    } finally {
      setSearching(false)
    }
  }

  function handleClearSearch() {
    setSearchQuery('')
    setCfFilterRows([])
    setFilterLocationId('')
    setIsSearchActive(false)
    load(1)
  }

  return {
    searchQuery, setSearchQuery,
    searching, isSearchActive, setIsSearchActive,
    cfFilterRows,
    filterLocationId, setFilterLocationId,
    sortCol, sortDir,
    searchableFields,
    handleSort, sortedSubnets,
    addCfFilterRow, updateCfFilterRow, removeCfFilterRow, addCfFilterFromValue,
    handleSearch, handleClearSearch,
  }
}
