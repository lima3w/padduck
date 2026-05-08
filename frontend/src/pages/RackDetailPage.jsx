import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getRack, getRackDevices } from '../api/racks'
import { getLocation } from '../api/locations'

const DEVICE_COLORS = [
  'bg-blue-200 dark:bg-blue-800 border-blue-400 dark:border-blue-600 text-blue-900 dark:text-blue-100',
  'bg-green-200 dark:bg-green-800 border-green-400 dark:border-green-600 text-green-900 dark:text-green-100',
  'bg-purple-200 dark:bg-purple-800 border-purple-400 dark:border-purple-600 text-purple-900 dark:text-purple-100',
  'bg-orange-200 dark:bg-orange-800 border-orange-400 dark:border-orange-600 text-orange-900 dark:text-orange-100',
  'bg-pink-200 dark:bg-pink-800 border-pink-400 dark:border-pink-600 text-pink-900 dark:text-pink-100',
  'bg-teal-200 dark:bg-teal-800 border-teal-400 dark:border-teal-600 text-teal-900 dark:text-teal-100',
]

export default function RackDetailPage() {
  const { id } = useParams()
  const [rack, setRack] = useState(null)
  const [rackDevices, setRackDevices] = useState([])
  const [locationBreadcrumb, setLocationBreadcrumb] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    loadAll()
  }, [id])

  async function loadAll() {
    try {
      setLoading(true)
      setError(null)
      const [rackData, devicesData] = await Promise.all([
        getRack(id),
        getRackDevices(id).catch(() => []),
      ])
      setRack(rackData)
      setRackDevices(Array.isArray(devicesData) ? devicesData : (devicesData?.devices ?? []))

      // Build location breadcrumb
      if (rackData.locationId) {
        const crumbs = []
        let locId = rackData.locationId
        while (locId) {
          try {
            const loc = await getLocation(locId)
            crumbs.unshift({ id: loc.id, name: loc.name })
            locId = loc.parentId || null
          } catch {
            break
          }
        }
        setLocationBreadcrumb(crumbs)
      }
    } catch (err) {
      setError(err.message || 'Failed to load rack')
    } finally {
      setLoading(false)
    }
  }

  if (loading) return <p className="text-gray-500">Loading rack...</p>
  if (error && !rack) return <p className="text-red-600">{error}</p>

  const sizeU = rack?.sizeU ?? 42

  // Build slot occupancy map: slotNumber -> device info
  const slotMap = {}
  rackDevices.forEach((d, idx) => {
    const start = d.rackUnitStart ?? 1
    const size = d.rackUnitSize ?? 1
    const colorClass = DEVICE_COLORS[idx % DEVICE_COLORS.length]
    for (let u = start; u < start + size; u++) {
      slotMap[u] = {
        device: d,
        isTop: u === start,
        size,
        color: colorClass,
      }
    }
  })

  const usedU = rackDevices.reduce((acc, d) => acc + (d.rackUnitSize ?? 1), 0)
  const freeU = sizeU - usedU

  return (
    <div>
      {/* Breadcrumb */}
      <nav className="text-sm text-gray-500 mb-4 flex items-center gap-1 flex-wrap">
        <Link to="/locations" className="hover:text-blue-600">Locations</Link>
        {locationBreadcrumb.map(crumb => (
          <span key={crumb.id} className="flex items-center gap-1">
            <span>/</span>
            <Link to={`/locations/${crumb.id}`} className="hover:text-blue-600">{crumb.name}</Link>
          </span>
        ))}
        <span>/</span>
        <span className="text-gray-800 dark:text-gray-200 font-medium">{rack?.name}</span>
      </nav>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{rack?.name}</h1>
          {rack?.description && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{rack.description}</p>
          )}
        </div>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-center">
          <p className="text-2xl font-bold text-gray-800 dark:text-gray-100">{sizeU}U</p>
          <p className="text-sm text-gray-500 dark:text-gray-400">Total</p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-center">
          <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">{usedU}U</p>
          <p className="text-sm text-gray-500 dark:text-gray-400">Used</p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-center">
          <p className="text-2xl font-bold text-green-600 dark:text-green-400">{freeU}U</p>
          <p className="text-sm text-gray-500 dark:text-gray-400">Free</p>
        </div>
      </div>

      <div className="flex gap-6 items-start">
        {/* Visual Rack Diagram */}
        <div className="flex-shrink-0">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">Rack Layout</h2>
          <div className="border-2 border-gray-400 dark:border-gray-500 rounded bg-gray-100 dark:bg-gray-900 overflow-hidden" style={{ width: 260 }}>
            {/* Rack header */}
            <div className="bg-gray-300 dark:bg-gray-700 text-center py-1 text-xs font-semibold text-gray-600 dark:text-gray-300 border-b border-gray-400 dark:border-gray-500">
              {rack?.name}
            </div>
            {Array.from({ length: sizeU }, (_, i) => {
              const uNum = i + 1
              const slot = slotMap[uNum]
              const isEmpty = !slot

              if (slot && !slot.isTop) return null // rendered as part of the top slot

              return (
                <div
                  key={uNum}
                  className={`flex items-center border-b border-gray-300 dark:border-gray-600 ${isEmpty ? '' : 'border-0'}`}
                  style={{ height: slot ? slot.size * 24 : 24 }}
                >
                  <span className="text-xs text-gray-400 dark:text-gray-500 w-7 text-right pr-1 flex-shrink-0 select-none font-mono">
                    {uNum}
                  </span>
                  {isEmpty ? (
                    <div className="flex-1 h-full bg-gray-50 dark:bg-gray-800 border border-dashed border-gray-300 dark:border-gray-600 mx-1 rounded-sm"></div>
                  ) : (
                    <div
                      className={`flex-1 mx-1 rounded-sm border ${slot.color} flex items-center px-2 overflow-hidden`}
                      style={{ height: '100%' }}
                    >
                      <div className="overflow-hidden">
                        <p className="text-xs font-medium truncate leading-tight">{slot.device.hostname}</p>
                        {slot.device.type?.name && (
                          <p className="text-xs opacity-70 truncate leading-tight">{slot.device.type.name}</p>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              )
            })}
            {/* Rack footer */}
            <div className="bg-gray-300 dark:bg-gray-700 text-center py-1 text-xs font-semibold text-gray-600 dark:text-gray-300 border-t border-gray-400 dark:border-gray-500">
              Bottom
            </div>
          </div>
        </div>

        {/* Device list */}
        <div className="flex-1">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">Devices in Rack</h2>
          {rackDevices.length === 0 ? (
            <p className="text-sm text-gray-400">No devices installed in this rack.</p>
          ) : (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                  <tr>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Hostname</th>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Position</th>
                    <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {rackDevices.map(d => (
                    <tr key={d.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                      <td className="px-4 py-3 font-medium">
                        <Link to={`/devices/${d.id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                          {d.hostname}
                        </Link>
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{d.type?.name || '—'}</td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 font-mono">
                        {d.rackUnitStart != null ? `U${d.rackUnitStart}–U${d.rackUnitStart + (d.rackUnitSize ?? 1) - 1}` : '—'}
                      </td>
                      <td className="px-4 py-3">
                        <span className="flex items-center gap-1.5 text-xs font-medium">
                          <span className={`w-2 h-2 rounded-full ${d.isOnline ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                          <span className={d.isOnline ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}>
                            {d.isOnline ? 'Online' : 'Offline'}
                          </span>
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
