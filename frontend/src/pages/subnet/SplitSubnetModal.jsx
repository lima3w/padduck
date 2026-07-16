import { useTranslation } from 'react-i18next'
import Modal from '../../components/Modal'

function splitCidrPreview(networkAddress, currentPrefix, newPrefix) {
  if (!networkAddress || isNaN(newPrefix) || newPrefix <= currentPrefix || newPrefix > 32) return []
  const parts = networkAddress.split('.').map(Number)
  if (parts.length !== 4 || parts.some(isNaN)) return []
  let base = 0
  for (const p of parts) base = (base << 8) | p
  base = base >>> 0
  const count = Math.pow(2, newPrefix - currentPrefix)
  const size = Math.pow(2, 32 - newPrefix)
  const results = []
  for (let i = 0; i < Math.min(count, 64); i++) {
    const net = (base + i * size) >>> 0
    const octets = [24, 16, 8, 0].map(s => (net >>> s) & 0xff)
    results.push(`${octets.join('.')}/${newPrefix}`)
  }
  return results
}

export default function SplitSubnetModal({ splitModal, splitPrefix, setSplitPrefix, splitting, splitError, splitBlockingIPs, onSplit, onClose }) {
  const { t } = useTranslation()
  return (
    <Modal title={t('subnetSplit.title', { cidr: `${splitModal.subnet.networkAddress}/${splitModal.subnet.prefixLength}` })} onClose={onClose}>
      <div className="space-y-4">
        {splitError && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
            <p>{splitError}</p>
            {splitBlockingIPs.length > 0 && (
              <ul className="mt-2 space-y-0.5 font-mono text-xs">
                {splitBlockingIPs.map(ip => <li key={ip} className="text-red-600">{ip}</li>)}
              </ul>
            )}
          </div>
        )}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('subnetSplit.newPrefixLength')}</label>
          <input
            type="number"
            min={splitModal.subnet.prefixLength + 1}
            max={32}
            className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={splitPrefix}
            onChange={e => setSplitPrefix(e.target.value)}
          />
        </div>
        {splitPrefix && !isNaN(parseInt(splitPrefix)) && parseInt(splitPrefix) > splitModal.subnet.prefixLength && (
          <div>
            <p className="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">{t('subnetSplit.previewLabel')}</p>
            <div className="grid grid-cols-2 gap-1 max-h-48 overflow-y-auto">
              {splitCidrPreview(splitModal.subnet.networkAddress, splitModal.subnet.prefixLength, parseInt(splitPrefix)).map((c, i) => (
                <span key={i} className="font-mono text-xs bg-gray-50 dark:bg-gray-700 rounded px-2 py-1">{c}</span>
              ))}
            </div>
          </div>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
          <button
            onClick={onSplit}
            disabled={splitting || !splitPrefix || isNaN(parseInt(splitPrefix)) || parseInt(splitPrefix) <= splitModal.subnet.prefixLength}
            className="px-4 py-2 bg-purple-600 text-white rounded text-sm hover:bg-purple-700 disabled:opacity-50"
          >
            {splitting ? t('subnetSplit.splitting') : t('subnets.split')}
          </button>
        </div>
      </div>
    </Modal>
  )
}
