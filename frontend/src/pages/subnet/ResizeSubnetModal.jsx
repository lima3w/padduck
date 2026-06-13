import Modal from '../../components/Modal'

export default function ResizeSubnetModal({ resizeModal, resizePrefix, setResizePrefix, resizing, resizeError, setResizeError, resizeConfirmText, setResizeConfirmText, onResize, onClose }) {
  return (
    <Modal title={`Resize ${resizeModal.subnet.networkAddress}/${resizeModal.subnet.prefixLength}`} onClose={onClose}>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">New CIDR</label>
          <input
            className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            placeholder="192.168.0.0/23"
            value={resizePrefix}
            onChange={e => { setResizePrefix(e.target.value); setResizeError(null); setResizeConfirmText('') }}
          />
        </div>
        {resizeError && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm dark:bg-red-900/20 dark:border-red-700 dark:text-red-400">
            <p className="font-medium mb-1">{resizeError.message}</p>
            {(resizeError.conflictingIps?.length > 0 || resizeError.conflictingSubnets?.length > 0) && (
              <>
                {resizeError.conflictingIps?.length > 0 && (
                  <div className="mt-2">
                    <p className="text-xs font-semibold">Conflicting IPs:</p>
                    <p className="font-mono text-xs">{resizeError.conflictingIps.join(', ')}</p>
                  </div>
                )}
                {resizeError.conflictingSubnets?.length > 0 && (
                  <div className="mt-2">
                    <p className="text-xs font-semibold">Conflicting Subnets:</p>
                    <p className="font-mono text-xs">{resizeError.conflictingSubnets.join(', ')}</p>
                  </div>
                )}
                <div className="mt-3">
                  <label className="block text-xs font-medium mb-1">Type CONFIRM to proceed anyway:</label>
                  <input
                    className="w-full border rounded px-2 py-1 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-red-400 dark:bg-gray-700 dark:border-gray-600"
                    placeholder="CONFIRM"
                    value={resizeConfirmText}
                    onChange={e => setResizeConfirmText(e.target.value)}
                  />
                </div>
              </>
            )}
          </div>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
          <button
            onClick={onResize}
            disabled={resizing || !resizePrefix || (resizeError?.conflictingIps?.length > 0 && resizeConfirmText !== 'CONFIRM')}
            className="px-4 py-2 bg-teal-600 text-white rounded text-sm hover:bg-teal-700 disabled:opacity-50"
          >
            {resizing ? 'Resizing...' : 'Resize'}
          </button>
        </div>
      </div>
    </Modal>
  )
}
