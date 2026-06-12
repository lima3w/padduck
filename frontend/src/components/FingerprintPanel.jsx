import { useState, useEffect } from 'react'
import { getDeviceFingerprint, buildDeviceFingerprint } from '../api/devices'

export default function FingerprintPanel({ deviceId, deviceIp }) {
  const [fp, setFp] = useState(null)
  const [loading, setLoading] = useState(true)
  const [building, setBuilding] = useState(false)

  useEffect(() => {
    if (!deviceId) return
    getDeviceFingerprint(deviceId)
      .then(res => setFp(res.data.fingerprint))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [deviceId])

  async function handleBuild() {
    setBuilding(true)
    try {
      const res = await buildDeviceFingerprint(deviceId, { device_ip: deviceIp, is_alive: true })
      setFp(res.data.fingerprint)
    } catch {}
    finally { setBuilding(false) }
  }

  if (loading) return null

  const pct = fp ? Math.round(fp.confidenceScore * 100) : 0
  const barCls = pct >= 70 ? 'bg-green-500' : pct >= 40 ? 'bg-yellow-500' : 'bg-red-500'

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">
          Device Fingerprint
        </h3>
        <button
          onClick={handleBuild}
          disabled={building}
          className="text-xs px-2 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {building ? 'Building...' : 'Rebuild'}
        </button>
      </div>

      {!fp ? (
        <p className="text-sm text-gray-400 dark:text-gray-500">No fingerprint yet. Click Rebuild to generate one.</p>
      ) : (
        <div className="space-y-2 text-sm">
          <div>
            <div className="flex justify-between text-xs text-gray-500 mb-1">
              <span>Confidence</span>
              <span>{pct}%</span>
            </div>
            <div className="h-1.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
              <div className={`h-full ${barCls} rounded-full`} style={{ width: `${pct}%` }} />
            </div>
          </div>
          {fp.osGuess && <div className="text-gray-600 dark:text-gray-400">OS: <span className="font-medium text-gray-800 dark:text-gray-200">{fp.osGuess}</span></div>}
          {fp.vendorGuess && <div className="text-gray-600 dark:text-gray-400">Vendor: <span className="font-medium text-gray-800 dark:text-gray-200">{fp.vendorGuess}</span></div>}
          {fp.openPorts?.length > 0 && (
            <div className="text-gray-600 dark:text-gray-400">
              Open ports: <span className="font-mono text-xs text-gray-800 dark:text-gray-200">{fp.openPorts.join(', ')}</span>
            </div>
          )}
          {fp.evidence?.length > 0 && (
            <ul className="text-xs text-gray-500 dark:text-gray-400 list-disc list-inside">
              {fp.evidence.map((e, i) => <li key={i}>{e}</li>)}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
