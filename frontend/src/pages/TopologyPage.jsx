import { useState, useEffect, useRef } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import cytoscape from 'cytoscape'
import { api } from '../api/client'

function utilizationColor(u) {
  if (u >= 0.8) return '#ef4444'
  if (u >= 0.5) return '#f59e0b'
  return '#22c55e'
}

export default function TopologyPage() {
  const { t } = useTranslation()
  const { id: sectionId } = useParams()
  const cyRef = useRef(null)
  const containerRef = useRef(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState(null)

  useEffect(() => {
    let cy = null
    async function load() {
      try {
        const { data } = await api.get(`/networks/${sectionId}/topology`)
        if (!containerRef.current) return

        const elements = [
          ...(data.nodes || []).map(n => ({
            data: {
              id: String(n.id),
              label: n.label || n.cidr,
              cidr: n.cidr,
              prefixLen: n.prefixLen,
              isContainer: n.isContainer,
              utilization: n.utilization ?? 0,
              vlanId: n.vlanId,
            },
          })),
          ...(data.edges || []).map(e => ({
            data: {
              id: `e-${e.source}-${e.target}`,
              source: String(e.source),
              target: String(e.target),
            },
          })),
        ]

        cy = cytoscape({
          container: containerRef.current,
          elements,
          style: [
            {
              selector: 'node',
              style: {
                shape: 'roundrectangle',
                label: 'data(label)',
                'font-size': 11,
                'text-valign': 'center',
                'text-halign': 'center',
                'background-color': (ele) => utilizationColor(ele.data('utilization')),
                'border-width': 1,
                'border-color': '#6b7280',
                color: '#fff',
                width: 120,
                height: 40,
                'text-wrap': 'wrap',
                'text-max-width': 110,
              },
            },
            {
              selector: 'node[?isContainer]',
              style: {
                'border-style': 'dashed',
                'border-width': 2,
              },
            },
            {
              selector: 'node:selected',
              style: {
                'border-color': '#3b82f6',
                'border-width': 3,
              },
            },
            {
              selector: 'edge',
              style: {
                'line-color': '#9ca3af',
                'target-arrow-color': '#9ca3af',
                'target-arrow-shape': 'triangle',
                'curve-style': 'bezier',
                width: 1.5,
              },
            },
          ],
          layout: {
            name: 'breadthfirst',
            directed: true,
            padding: 20,
            spacingFactor: 1.4,
          },
          userZoomingEnabled: true,
          userPanningEnabled: true,
        })

        cy.on('tap', 'node', (evt) => {
          const node = evt.target
          setSelected({
            cidr: node.data('cidr'),
            label: node.data('label'),
            prefixLen: node.data('prefixLen'),
            isContainer: node.data('isContainer'),
            utilization: node.data('utilization'),
            vlanId: node.data('vlanId'),
          })
        })

        cy.on('tap', (evt) => {
          if (evt.target === cy) setSelected(null)
        })

        cyRef.current = cy
      } catch {
        setError(t('topology.loadError'))
      } finally {
        setLoading(false)
      }
    }
    load()
    return () => { cy?.destroy() }
  }, [sectionId, t])

  function exportPng() {
    if (!cyRef.current) return
    const png = cyRef.current.png({ full: true, scale: 2 })
    const a = document.createElement('a')
    a.href = png
    a.download = `topology-network-${sectionId}.png`
    a.click()
  }

  const pct = selected ? Math.round((selected.utilization ?? 0) * 100) : 0

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Link to="/networks" className="text-sm text-blue-600 hover:underline dark:text-blue-400">{t('nav.networks')}</Link>
          <span className="text-gray-400">/</span>
          <Link to={`/networks/${sectionId}/subnets`} className="text-sm text-blue-600 hover:underline dark:text-blue-400">{t('dashboard.subnets')}</Link>
          <span className="text-gray-400">/</span>
          <span className="text-sm text-gray-700 dark:text-gray-300 font-medium">{t('networks.topology')}</span>
        </div>
        <button
          onClick={exportPng}
          className="text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-3 py-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition"
        >
          {t('topology.exportPng')}
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{error}</div>
      )}

      {loading && (
        <div className="text-gray-500 dark:text-gray-400 text-sm">{t('topology.loading')}</div>
      )}

      <div className="flex gap-4 flex-1 min-h-0" style={{ height: 'calc(100vh - 200px)' }}>
        <div
          ref={containerRef}
          className="flex-1 bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700"
          style={{ minHeight: 400 }}
        />

        {selected && (
          <div className="w-56 bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 p-4 flex-shrink-0">
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-semibold text-gray-800 dark:text-gray-200 text-sm">{t('topology.subnetDetail')}</h3>
              <button onClick={() => setSelected(null)} className="text-gray-400 hover:text-gray-600 text-lg leading-none">&times;</button>
            </div>
            <div className="space-y-2 text-sm">
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">{t('topology.cidr')}</p>
                <p className="font-mono font-medium text-gray-900 dark:text-gray-100">{selected.cidr}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">{t('networks.prefixLength')}</p>
                <p className="font-mono text-gray-900 dark:text-gray-100">/{selected.prefixLen}</p>
              </div>
              {selected.isContainer && (
                <div>
                  <span className="text-xs bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 px-2 py-0.5 rounded">{t('topology.container')}</span>
                </div>
              )}
              {selected.vlanId && (
                <div>
                  <p className="text-xs text-gray-500 dark:text-gray-400">{t('topology.vlanId')}</p>
                  <p className="text-gray-900 dark:text-gray-100">{selected.vlanId}</p>
                </div>
              )}
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">{t('topology.utilisation')}</p>
                <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                  <div
                    className="h-2 rounded-full transition-all"
                    style={{ width: `${pct}%`, backgroundColor: utilizationColor(selected.utilization ?? 0) }}
                  />
                </div>
                <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">{pct}%</p>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="mt-3 flex gap-4 text-xs text-gray-500 dark:text-gray-400">
        <span className="flex items-center gap-1"><span className="w-3 h-3 bg-green-500 rounded-sm inline-block"></span> {t('topology.under50Used')}</span>
        <span className="flex items-center gap-1"><span className="w-3 h-3 bg-amber-500 rounded-sm inline-block"></span> {t('topology.between50And80Used')}</span>
        <span className="flex items-center gap-1"><span className="w-3 h-3 bg-red-500 rounded-sm inline-block"></span> {t('topology.over80Used')}</span>
      </div>
    </div>
  )
}
