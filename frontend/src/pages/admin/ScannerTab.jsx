import { useState } from 'react'

export default function ScannerTab({ config, handleConfigChange, handleSaveConfig, saving }) {
  const [showSnmpCommunity, setShowSnmpCommunity] = useState(false)

  return (
        <div className="space-y-4">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">Discovery Scanner</h2>
            <div className="space-y-4">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.scanner_resolve_hostnames !== 'false'}
                  onChange={(e) =>
                    handleConfigChange('scanner_resolve_hostnames', e.target.checked ? 'true' : 'false')
                  }
                  className="w-4 h-4 text-blue-600"
                />
                <div>
                  <span className="font-medium text-gray-900">Resolve hostnames via reverse DNS</span>
                  <p className="text-xs text-gray-500 mt-0.5">
                    After a successful ping, the scanner performs a PTR lookup and stores the result in the
                    IP address record. A forward lookup is also done to flag mismatches. Adds up to 2 s per
                    alive host. Default: enabled.
                  </p>
                </div>
              </label>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-base font-semibold mb-4">SNMP</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Global community string</label>
                <div className="relative">
                  <input
                    type={showSnmpCommunity ? 'text' : 'password'}
                    value={config.scanner_snmp_community ?? ''}
                    onChange={(e) => handleConfigChange('scanner_snmp_community', e.target.value)}
                    placeholder="public"
                    className="w-full px-3 py-2 pr-10 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => setShowSnmpCommunity(v => !v)}
                    className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-400 hover:text-gray-600"
                    aria-label={showSnmpCommunity ? 'Hide community string' : 'Show community string'}
                  >
                    {showSnmpCommunity ? (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 4.411m0 0L21 21" /></svg>
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                    )}
                  </button>
                </div>
                <p className="text-xs text-gray-500 mt-1">Used when no per-device community is configured. Default: public.</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">SNMP version</label>
                <select
                  value={config.scanner_snmp_version ?? '2c'}
                  onChange={(e) => handleConfigChange('scanner_snmp_version', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                >
                  <option value="2c">v2c</option>
                  <option value="3">v3</option>
                </select>
                <p className="text-xs text-gray-500 mt-1">Global default version. Per-device credentials override this.</p>
              </div>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-base font-semibold mb-4">Port Scanning</h3>
            <div className="space-y-4">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.scanner_port_scan_enabled === 'true'}
                  onChange={(e) =>
                    handleConfigChange('scanner_port_scan_enabled', e.target.checked ? 'true' : 'false')
                  }
                  className="w-4 h-4 text-blue-600"
                />
                <div>
                  <span className="font-medium text-gray-900">Enable TCP port scanning</span>
                  <p className="text-xs text-gray-500 mt-0.5">
                    After a successful ping, probe the ports listed below on each alive host. Default: disabled.
                  </p>
                </div>
              </label>
              {config.scanner_port_scan_enabled === 'true' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Port list</label>
                  <input
                    type="text"
                    value={config.scanner_port_list ?? ''}
                    onChange={(e) => handleConfigChange('scanner_port_list', e.target.value)}
                    placeholder="22,80,443,3306,5432,8080,8443"
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 text-sm"
                  />
                  <p className="text-xs text-gray-500 mt-1">Comma-separated port numbers. Default: 22,80,443,3306,5432,8080,8443.</p>
                </div>
              )}
            </div>
          </div>

          <div className="flex gap-3 items-center">
            <button
              onClick={handleSaveConfig}
              disabled={saving}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition font-medium"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
  )
}
