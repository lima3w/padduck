import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Link } from 'react-router'
import { useTranslation } from 'react-i18next'
import cytoscape from 'cytoscape'
import {
  getTopologyGraph, getTopologyPath, getTopologyEdges, setTopologyEdgeStatus, syncTopologyGraph,
} from '../api/topology'
import { getCachedUser } from '../utils/storageKeys'

const TYPES = ['subnet', 'device', 'vlan', 'location']
const TYPE_COLORS = { subnet: '#3b82f6', device: '#10b981', vlan: '#8b5cf6', location: '#f59e0b' }
const TYPE_SHAPES = { subnet: 'round-rectangle', device: 'ellipse', vlan: 'hexagon', location: 'diamond' }

export function resourcePath(type, id) {
  switch (type) {
    case 'subnet': return `/subnets/${id}/ip-addresses`
    case 'device': return `/devices/${id}`
    case 'vlan': return `/vlans/${id}`
    case 'location': return `/locations/${id}`
    default: return null
  }
}

// The API client camel-cases response keys, so accept either spelling.
const field = (d, camel, snake) => d?.[camel] ?? d?.[snake]

export default function TopologyGraphPage() {
  const { t } = useTranslation()
  const isAdmin = getCachedUser()?.role === 'admin'
  const containerRef = useRef(null)
  const cyRef = useRef(null)
  const [graph, setGraph] = useState(null)
  const [suggested, setSuggested] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [selectedId, setSelectedId] = useState(null)
  const [visibleTypes, setVisibleTypes] = useState(new Set(TYPES))
  const [search, setSearch] = useState('')
  const [pathTarget, setPathTarget] = useState('')
  const [pathMsg, setPathMsg] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [g, s] = await Promise.all([getTopologyGraph(), getTopologyEdges('suggested')])
      setGraph(g.data)
      setSuggested(s.data?.edges || [])
    } catch {
      setError(t('topologyGraph.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => { load() }, [load])

  const nodes = useMemo(() => (graph?.nodes || []).map(n => n.data), [graph])
  const edges = useMemo(() => (graph?.edges || []).map(e => e.data), [graph])
  const nodeById = useMemo(() => Object.fromEntries(nodes.map(n => [n.id, n])), [nodes])

  // Build the Cytoscape instance whenever the graph changes.
  useEffect(() => {
    if (!graph || !containerRef.current) return undefined
    const cy = cytoscape({
      container: containerRef.current,
      elements: [
        ...nodes.map(n => ({
          data: { id: n.id, label: n.label || n.id, type: field(n, 'resourceType', 'resource_type') },
        })),
        ...edges.map(e => ({
          data: { id: e.id, source: e.source, target: e.target, relationship: e.relationship, status: e.status },
        })),
      ],
      style: [
        {
          selector: 'node',
          style: {
            label: 'data(label)', 'font-size': 10, 'text-valign': 'bottom', 'text-margin-y': 4,
            color: '#374151', 'background-color': (ele) => TYPE_COLORS[ele.data('type')] || '#6b7280',
            shape: (ele) => TYPE_SHAPES[ele.data('type')] || 'ellipse', width: 28, height: 28,
          },
        },
        { selector: 'node:selected', style: { 'border-width': 3, 'border-color': '#111827' } },
        {
          selector: 'edge',
          style: {
            width: 1.5, 'line-color': '#9ca3af', 'target-arrow-color': '#9ca3af', 'target-arrow-shape': 'triangle',
            'curve-style': 'bezier', label: 'data(relationship)', 'font-size': 8, color: '#6b7280',
            'text-rotation': 'autorotate',
          },
        },
        { selector: 'edge[relationship = "uplink_to"]', style: { width: 2.5, 'line-color': '#ef4444', 'target-arrow-color': '#ef4444' } },
        { selector: 'edge[status = "suggested"]', style: { 'line-style': 'dashed', 'line-color': '#f59e0b', 'target-arrow-color': '#f59e0b' } },
        { selector: '.dim', style: { opacity: 0.12 } },
        { selector: 'edge.path', style: { width: 4, 'line-color': '#111827', 'target-arrow-color': '#111827' } },
        { selector: 'node.hidden-type, edge.hidden-type', style: { display: 'none' } },
      ],
      layout: { name: 'cose', animate: false, padding: 30, nodeRepulsion: () => 9000 },
      wheelSensitivity: 0.3,
    })
    cy.on('tap', 'node', (evt) => { setSelectedId(evt.target.id()); setPathMsg('') })
    cy.on('tap', (evt) => { if (evt.target === cy) { setSelectedId(null); setPathMsg('') } })
    cyRef.current = cy
    return () => { cy.destroy(); cyRef.current = null }
  }, [graph, nodes, edges])

  // Type filter and search dimming.
  useEffect(() => {
    const cy = cyRef.current
    if (!cy) return
    cy.batch(() => {
      cy.elements().removeClass('hidden-type dim')
      cy.nodes().forEach(n => { if (!visibleTypes.has(n.data('type'))) n.addClass('hidden-type') })
      cy.edges().forEach(e => {
        if (e.source().hasClass('hidden-type') || e.target().hasClass('hidden-type')) e.addClass('hidden-type')
      })
      const q = search.trim().toLowerCase()
      if (q) {
        const hits = cy.nodes().filter(n => !n.hasClass('hidden-type') && (n.data('label') || '').toLowerCase().includes(q))
        cy.elements().not(hits).not(hits.connectedEdges()).addClass('dim')
        if (hits.length) cy.animate({ fit: { eles: hits, padding: 80 }, duration: 200 })
      }
    })
  }, [visibleTypes, search, graph])

  const selected = selectedId ? nodeById[selectedId] : null
  const selectedType = selected && field(selected, 'resourceType', 'resource_type')
  const selectedRid = selected && field(selected, 'resourceId', 'resource_id')
  const connected = useMemo(() => edges.filter(e => e.source === selectedId || e.target === selectedId), [edges, selectedId])

  async function decide(edgeId, status) {
    setError('')
    try {
      await setTopologyEdgeStatus(edgeId, status)
      await load()
    } catch {
      setError(t('topologyGraph.updateError'))
    }
  }

  async function findPath() {
    const cy = cyRef.current
    const target = nodeById[pathTarget]
    if (!cy || !selected || !target) return
    cy.elements().removeClass('path dim')
    try {
      const { data } = await getTopologyPath(
        { type: selectedType, id: selectedRid },
        { type: field(target, 'resourceType', 'resource_type'), id: field(target, 'resourceId', 'resource_id') },
      )
      const ids = new Set((data.nodes || []).map(n => n.data.id))
      const eids = new Set((data.edges || []).map(e => e.data.id))
      cy.elements().addClass('dim')
      cy.elements().filter(el => ids.has(el.id()) || eids.has(el.id())).removeClass('dim').addClass('path')
      setPathMsg(t('topologyGraph.pathHops', { count: Math.max(ids.size - 1, 0) }))
    } catch {
      setPathMsg(t('topologyGraph.noPath'))
    }
  }

  async function rebuild() {
    setError('')
    try {
      await syncTopologyGraph()
      await load()
    } catch {
      setError(t('topologyGraph.syncError'))
    }
  }

  function exportPng() {
    if (!cyRef.current) return
    const a = document.createElement('a')
    a.href = cyRef.current.png({ full: true, scale: 2, bg: '#ffffff' })
    a.download = 'topology-graph.png'
    a.click()
  }

  function toggleType(type) {
    setVisibleTypes(prev => {
      const next = new Set(prev)
      if (next.has(type)) next.delete(type)
      else next.add(type)
      return next
    })
  }

  const labelOf = (id) => nodeById[id]?.label || id

  return (
    <div className="flex flex-col h-full">
      <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
        <h1 className="text-xl font-semibold text-gray-900 dark:text-gray-100">{t('topologyGraph.title')}</h1>
        <div className="flex flex-wrap items-center gap-2">
          <input
            type="search"
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder={t('topologyGraph.searchPlaceholder')}
            aria-label={t('topologyGraph.searchPlaceholder')}
            className="text-sm border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100 rounded px-2 py-1.5"
          />
          {TYPES.map(type => (
            <label key={type} className="flex items-center gap-1 text-xs text-gray-700 dark:text-gray-300">
              <input type="checkbox" checked={visibleTypes.has(type)} onChange={() => toggleType(type)} />
              <span className="inline-block w-2.5 h-2.5 rounded-sm" style={{ backgroundColor: TYPE_COLORS[type] }} />
              {t(`topologyGraph.types.${type}`)}
            </label>
          ))}
          {isAdmin && (
            <button onClick={rebuild} className="text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-3 py-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
              {t('topologyGraph.rebuild')}
            </button>
          )}
          <button onClick={exportPng} className="text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-3 py-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
            {t('topologyGraph.exportPng')}
          </button>
        </div>
      </div>

      {error && <div className="mb-3 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{error}</div>}

      {suggested.length > 0 && (
        <div className="mb-3 p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded" data-testid="suggested-edges">
          <p className="text-sm font-medium text-amber-900 dark:text-amber-200 mb-2">
            {t('topologyGraph.suggestedHeading', { count: suggested.length })}
          </p>
          <ul className="space-y-1">
            {suggested.map(e => (
              <li key={e.id} className="flex flex-wrap items-center gap-2 text-sm text-gray-800 dark:text-gray-200">
                <span>
                  <strong>{field(e, 'sourceLabel', 'source_label')}</strong>{' '}
                  {e.relationship.replace(/_/g, ' ')}{' '}
                  <strong>{field(e, 'targetLabel', 'target_label')}</strong>
                </span>
                {e.evidence && <span className="text-xs text-gray-500 dark:text-gray-400">({e.evidence})</span>}
                {isAdmin && (
                  <span className="ml-auto flex gap-2">
                    <button onClick={() => decide(e.id, 'confirmed')} className="text-xs bg-green-600 text-white px-2 py-1 rounded hover:bg-green-700">{t('topologyGraph.confirm')}</button>
                    <button onClick={() => decide(e.id, 'dismissed')} className="text-xs bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-200 px-2 py-1 rounded hover:bg-gray-300 dark:hover:bg-gray-600">{t('topologyGraph.dismiss')}</button>
                  </span>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}

      {loading && <div className="text-sm text-gray-500 dark:text-gray-400">{t('topologyGraph.loading')}</div>}
      {!loading && graph && nodes.length === 0 && (
        <div className="text-sm text-gray-500 dark:text-gray-400">{t('topologyGraph.empty')}</div>
      )}

      <div className="flex gap-4 flex-1 min-h-0" style={{ height: 'calc(100vh - 260px)' }}>
        <div
          ref={containerRef}
          data-testid="topology-canvas"
          className="flex-1 bg-white dark:bg-gray-100 rounded-lg shadow border border-gray-200 dark:border-gray-700"
          style={{ minHeight: 400 }}
        />
        {selected && (
          <aside className="w-64 bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 p-4 flex-shrink-0 overflow-y-auto" aria-label={t('topologyGraph.details')}>
            <div className="flex items-center justify-between mb-2">
              <h2 className="font-semibold text-sm text-gray-800 dark:text-gray-200">{t(`topologyGraph.types.${selectedType}`, selectedType)}</h2>
              <button onClick={() => setSelectedId(null)} aria-label={t('common.close', 'Close')} className="text-gray-400 hover:text-gray-600 text-lg leading-none">&times;</button>
            </div>
            <p className="font-medium text-gray-900 dark:text-gray-100 break-words">{selected.label}</p>
            {selected.detail && <p className="text-xs text-gray-500 dark:text-gray-400 break-words">{selected.detail}</p>}
            {resourcePath(selectedType, selectedRid) && (
              <Link to={resourcePath(selectedType, selectedRid)} className="inline-block mt-2 text-sm text-blue-600 dark:text-blue-400 hover:underline">
                {t('topologyGraph.openResource')}
              </Link>
            )}

            <h3 className="mt-4 mb-1 text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">{t('topologyGraph.relationships')}</h3>
            {connected.length === 0 && <p className="text-xs text-gray-500 dark:text-gray-400">{t('topologyGraph.noRelationships')}</p>}
            <ul className="space-y-1 text-sm text-gray-800 dark:text-gray-200">
              {connected.map(e => {
                const outgoing = e.source === selectedId
                return (
                  <li key={e.id}>
                    {outgoing ? '→' : '←'} {e.relationship.replace(/_/g, ' ')} {labelOf(outgoing ? e.target : e.source)}
                    {e.status === 'suggested' && <span className="ml-1 text-xs text-amber-600">({t('topologyGraph.suggested')})</span>}
                  </li>
                )
              })}
            </ul>

            <h3 className="mt-4 mb-1 text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">{t('topologyGraph.findPath')}</h3>
            <select
              value={pathTarget}
              onChange={e => setPathTarget(e.target.value)}
              aria-label={t('topologyGraph.pathTarget')}
              className="w-full text-sm border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 rounded px-2 py-1"
            >
              <option value="">{t('topologyGraph.pathTarget')}</option>
              {nodes.filter(n => n.id !== selectedId).map(n => <option key={n.id} value={n.id}>{n.label || n.id}</option>)}
            </select>
            <button
              onClick={findPath}
              disabled={!pathTarget}
              className="mt-2 text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {t('topologyGraph.findPath')}
            </button>
            {pathMsg && <p className="mt-2 text-xs text-gray-600 dark:text-gray-300" role="status">{pathMsg}</p>}
          </aside>
        )}
      </div>

      <div className="mt-3 flex flex-wrap gap-4 text-xs text-gray-500 dark:text-gray-400">
        <span>{t('topologyGraph.legendSolid')}</span>
        <span>{t('topologyGraph.legendDashed')}</span>
        <span>{t('topologyGraph.legendUplink')}</span>
      </div>
    </div>
  )
}
