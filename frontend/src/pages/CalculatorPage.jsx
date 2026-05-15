import { useState, useCallback } from 'react'

// Pure IPv4 calculation functions
function ipv4ToInt(ip) {
  const parts = ip.split('.')
  if (parts.length !== 4) return null
  let n = 0
  for (const p of parts) {
    const v = parseInt(p, 10)
    if (isNaN(v) || v < 0 || v > 255) return null
    n = (n << 8) | v
  }
  return n >>> 0
}

function intToIPv4(n) {
  return [24, 16, 8, 0].map(s => (n >>> s) & 0xff).join('.')
}

function maskFromPrefix(prefix) {
  if (prefix === 0) return 0
  return (0xffffffff << (32 - prefix)) >>> 0
}

function prefixFromMask(maskStr) {
  const n = ipv4ToInt(maskStr)
  if (n === null) return null
  const bin = n.toString(2).padStart(32, '0')
  if (!/^1*0*$/.test(bin)) return null
  return bin.split('').filter(c => c === '1').length
}

function calcIPv4(ip, prefix) {
  const ipInt = ipv4ToInt(ip)
  if (ipInt === null || prefix < 0 || prefix > 32) return null
  const mask = maskFromPrefix(prefix)
  const network = (ipInt & mask) >>> 0
  const broadcast = (network | (~mask >>> 0)) >>> 0
  const total = Math.pow(2, 32 - prefix)
  const usable = prefix >= 31 ? total : Math.max(0, total - 2)
  const first = prefix >= 31 ? network : network + 1
  const last = prefix >= 31 ? broadcast : broadcast - 1
  const wildcard = (~mask) >>> 0
  return {
    networkAddress: intToIPv4(network),
    broadcastAddress: intToIPv4(broadcast),
    firstHost: intToIPv4(first),
    lastHost: intToIPv4(last),
    totalHosts: total,
    usableHosts: usable,
    wildcardMask: intToIPv4(wildcard),
    subnetMask: intToIPv4(mask),
    prefixLength: prefix,
    ipInt,
    maskInt: mask,
    networkInt: network,
  }
}

function getSubSubnets(networkInt, prefix, newPrefix) {
  if (newPrefix <= prefix || newPrefix > 32) return []
  const count = Math.pow(2, newPrefix - prefix)
  const size = Math.pow(2, 32 - newPrefix)
  const results = []
  for (let i = 0; i < Math.min(count, 64); i++) {
    const net = (networkInt + i * size) >>> 0
    results.push(`${intToIPv4(net)}/${newPrefix}`)
  }
  return results
}

function getSupernet(networkInt, prefix) {
  if (prefix <= 0) return null
  const superPrefix = prefix - 1
  const superMask = maskFromPrefix(superPrefix)
  const superNet = (networkInt & superMask) >>> 0
  return `${intToIPv4(superNet)}/${superPrefix}`
}

function toBinaryGroups(ipInt, maskInt) {
  const ipBin = ipInt.toString(2).padStart(32, '0')
  const networkBits = maskInt.toString(2).split('').filter(c => c === '1').length
  const netPart = ipBin.slice(0, networkBits)
  const hostPart = ipBin.slice(networkBits)
  return { netPart, hostPart, networkBits }
}

// IPv6 functions using BigInt
function ipv6ToBigInt(ip) {
  try {
    const expanded = expandIPv6(ip)
    if (!expanded) return null
    return BigInt('0x' + expanded.replace(/:/g, ''))
  } catch {
    return null
  }
}

function expandIPv6(ip) {
  if (ip.includes('::')) {
    const parts = ip.split('::')
    const left = parts[0] ? parts[0].split(':') : []
    const right = parts[1] ? parts[1].split(':') : []
    const missing = 8 - left.length - right.length
    const middle = Array(missing).fill('0000')
    const all = [...left, ...middle, ...right]
    if (all.length !== 8) return null
    return all.map(g => g.padStart(4, '0')).join(':')
  }
  const groups = ip.split(':')
  if (groups.length !== 8) return null
  return groups.map(g => g.padStart(4, '0')).join(':')
}

function bigIntToIPv6(n) {
  const hex = n.toString(16).padStart(32, '0')
  return hex.match(/.{4}/g).join(':')
}

function calcIPv6(ip, prefix) {
  const ipInt = ipv6ToBigInt(ip)
  if (ipInt === null || prefix < 0 || prefix > 128) return null
  const bits = BigInt(128 - prefix)
  const mask = prefix === 0 ? 0n : ((1n << BigInt(prefix)) - 1n) << bits
  const network = ipInt & mask
  const broadcast = network | ((1n << bits) - 1n)
  const total = 1n << bits
  return {
    networkAddress: bigIntToIPv6(network),
    broadcastAddress: bigIntToIPv6(broadcast),
    firstHost: bigIntToIPv6(network + 1n),
    lastHost: bigIntToIPv6(broadcast - 1n),
    totalHosts: total > BigInt(Number.MAX_SAFE_INTEGER) ? total.toString() : Number(total),
    usableHosts: total > BigInt(Number.MAX_SAFE_INTEGER) ? (total - 2n).toString() : Number(total - 2n),
    prefixLength: prefix,
    isIPv6: true,
  }
}

function detectIPVersion(ip) {
  if (ip.includes(':')) return 6
  if (ip.includes('.')) return 4
  return null
}

function validateIP(ip) {
  const v = detectIPVersion(ip)
  if (v === 4) return ipv4ToInt(ip) !== null
  if (v === 6) return ipv6ToBigInt(ip) !== null
  return false
}

export default function CalculatorPage() {
  const [ipInput, setIpInput] = useState('192.168.1.100')
  const [prefixInput, setPrefixInput] = useState('24')
  const [maskInput, setMaskInput] = useState('255.255.255.0')
  const [inputMode, setInputMode] = useState('prefix') // 'prefix' | 'mask'
  const [subSplitPrefix, setSubSplitPrefix] = useState('')
  const [error, setError] = useState('')

  const handlePrefixChange = useCallback((val) => {
    setPrefixInput(val)
    const n = parseInt(val, 10)
    if (!isNaN(n) && n >= 0 && n <= 32) {
      setMaskInput(intToIPv4(maskFromPrefix(n)))
    }
    setError('')
  }, [])

  const handleMaskChange = useCallback((val) => {
    setMaskInput(val)
    const p = prefixFromMask(val)
    if (p !== null) {
      setPrefixInput(String(p))
    }
    setError('')
  }, [])

  const handleIpChange = useCallback((val) => {
    setIpInput(val)
    setError('')
  }, [])

  const ipVersion = detectIPVersion(ipInput)
  const prefix = parseInt(prefixInput, 10)

  let result = null
  let calcError = null
  if (ipInput.trim()) {
    if (!validateIP(ipInput.trim())) {
      calcError = 'Invalid IP address'
    } else if (isNaN(prefix)) {
      calcError = 'Invalid prefix length'
    } else {
      if (ipVersion === 4) {
        result = calcIPv4(ipInput.trim(), prefix)
        if (!result) calcError = 'Invalid calculation parameters'
      } else if (ipVersion === 6) {
        result = calcIPv6(ipInput.trim(), prefix)
        if (!result) calcError = 'Invalid calculation parameters'
      }
    }
  }

  const subSubnets = result && !result.isIPv6 && subSplitPrefix
    ? getSubSubnets(result.networkInt, result.prefixLength, parseInt(subSplitPrefix, 10))
    : []

  const supernet = result && !result.isIPv6
    ? getSupernet(result.networkInt, result.prefixLength)
    : null

  const binaryDisplay = result && !result.isIPv6
    ? toBinaryGroups(result.ipInt, result.maskInt)
    : null

  const inputClass = "w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
  const labelClass = "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
  const rowClass = "flex justify-between py-2 border-b border-gray-100 dark:border-gray-700 last:border-0"
  const keyClass = "text-sm text-gray-500 dark:text-gray-400"
  const valClass = "text-sm font-mono font-medium text-gray-900 dark:text-gray-100"

  return (
    <div className="max-w-3xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">IP / Subnet Calculator</h1>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className={labelClass}>IP Address</label>
            <input
              className={inputClass}
              placeholder="192.168.1.0 or 2001:db8::1"
              value={ipInput}
              onChange={e => handleIpChange(e.target.value)}
            />
          </div>
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className={labelClass.replace('mb-1', '')}>
                {inputMode === 'prefix' ? 'Prefix Length' : 'Subnet Mask'}
              </label>
              {ipVersion !== 6 && (
                <button
                  onClick={() => setInputMode(m => m === 'prefix' ? 'mask' : 'prefix')}
                  className="text-xs text-blue-600 hover:underline"
                >
                  Switch to {inputMode === 'prefix' ? 'mask' : 'prefix'}
                </button>
              )}
            </div>
            {inputMode === 'prefix' || ipVersion === 6 ? (
              <input
                className={inputClass}
                placeholder={ipVersion === 6 ? '64' : '24'}
                value={prefixInput}
                onChange={e => handlePrefixChange(e.target.value)}
                type="number"
                min={0}
                max={ipVersion === 6 ? 128 : 32}
              />
            ) : (
              <input
                className={inputClass}
                placeholder="255.255.255.0"
                value={maskInput}
                onChange={e => handleMaskChange(e.target.value)}
              />
            )}
          </div>
        </div>
      </div>

      {calcError && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm dark:bg-red-900/20 dark:border-red-700 dark:text-red-400">
          {calcError}
        </div>
      )}

      {result && (
        <>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
            <h2 className="text-base font-semibold text-gray-700 dark:text-gray-300 mb-4">Results</h2>
            <div>
              <div className={rowClass}>
                <span className={keyClass}>Network address</span>
                <span className={valClass}>{result.networkAddress}</span>
              </div>
              {!result.isIPv6 && (
                <div className={rowClass}>
                  <span className={keyClass}>Broadcast address</span>
                  <span className={valClass}>{result.broadcastAddress}</span>
                </div>
              )}
              <div className={rowClass}>
                <span className={keyClass}>First usable host</span>
                <span className={valClass}>{result.firstHost}</span>
              </div>
              <div className={rowClass}>
                <span className={keyClass}>Last usable host</span>
                <span className={valClass}>{result.lastHost}</span>
              </div>
              <div className={rowClass}>
                <span className={keyClass}>Total hosts</span>
                <span className={valClass}>{typeof result.totalHosts === 'bigint' ? result.totalHosts.toString() : result.totalHosts.toLocaleString()}</span>
              </div>
              <div className={rowClass}>
                <span className={keyClass}>Usable hosts</span>
                <span className={valClass}>{typeof result.usableHosts === 'bigint' ? result.usableHosts.toString() : result.usableHosts.toLocaleString()}</span>
              </div>
              {!result.isIPv6 && (
                <>
                  <div className={rowClass}>
                    <span className={keyClass}>Subnet mask</span>
                    <span className={valClass}>{result.subnetMask}</span>
                  </div>
                  <div className={rowClass}>
                    <span className={keyClass}>Wildcard mask</span>
                    <span className={valClass}>{result.wildcardMask}</span>
                  </div>
                </>
              )}
              <div className={rowClass}>
                <span className={keyClass}>Prefix length</span>
                <span className={valClass}>/{result.prefixLength}</span>
              </div>
            </div>
          </div>

          {binaryDisplay && (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
              <h2 className="text-base font-semibold text-gray-700 dark:text-gray-300 mb-3">Binary Representation</h2>
              <div className="font-mono text-xs overflow-x-auto">
                <div className="flex flex-wrap gap-px">
                  {binaryDisplay.netPart.split('').map((bit, i) => (
                    <span
                      key={i}
                      className="w-4 h-6 flex items-center justify-center bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-sm"
                    >
                      {bit}
                    </span>
                  ))}
                  {binaryDisplay.hostPart.split('').map((bit, i) => (
                    <span
                      key={i + binaryDisplay.networkBits}
                      className="w-4 h-6 flex items-center justify-center bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded-sm"
                    >
                      {bit}
                    </span>
                  ))}
                </div>
                <div className="mt-2 flex gap-4 text-xs text-gray-500 dark:text-gray-400">
                  <span className="flex items-center gap-1">
                    <span className="w-3 h-3 bg-blue-100 dark:bg-blue-900 rounded-sm inline-block"></span>
                    Network ({binaryDisplay.networkBits} bits)
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="w-3 h-3 bg-green-100 dark:bg-green-900 rounded-sm inline-block"></span>
                    Host ({32 - binaryDisplay.networkBits} bits)
                  </span>
                </div>
              </div>
            </div>
          )}

          {!result.isIPv6 && (
            <>
              {supernet && (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
                  <h2 className="text-base font-semibold text-gray-700 dark:text-gray-300 mb-2">Supernet</h2>
                  <p className="font-mono text-sm text-gray-900 dark:text-gray-100">
                    Smallest containing supernet: <strong>{supernet}</strong>
                  </p>
                </div>
              )}

              <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
                <h2 className="text-base font-semibold text-gray-700 dark:text-gray-300 mb-3">Sub-subnet Split</h2>
                <div className="flex gap-2 mb-4">
                  <input
                    type="number"
                    min={result.prefixLength + 1}
                    max={32}
                    className="border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 w-32"
                    placeholder={`/${result.prefixLength + 1}`}
                    value={subSplitPrefix}
                    onChange={e => setSubSplitPrefix(e.target.value)}
                  />
                  <span className="text-sm text-gray-500 dark:text-gray-400 self-center">new prefix length</span>
                </div>
                {subSubnets.length > 0 && (
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                    {subSubnets.map((s, i) => (
                      <span key={i} className="font-mono text-xs bg-gray-50 dark:bg-gray-700 rounded px-2 py-1 text-gray-800 dark:text-gray-200">
                        {s}
                      </span>
                    ))}
                    {Math.pow(2, parseInt(subSplitPrefix, 10) - result.prefixLength) > 64 && (
                      <span className="text-xs text-gray-400 col-span-full mt-1">Showing first 64 of {Math.pow(2, parseInt(subSplitPrefix, 10) - result.prefixLength)}</span>
                    )}
                  </div>
                )}
              </div>
            </>
          )}
        </>
      )}
    </div>
  )
}
