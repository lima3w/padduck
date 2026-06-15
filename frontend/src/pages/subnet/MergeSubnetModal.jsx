import Modal from '../../components/Modal'

export default function MergeSubnetModal({ mergeModal, mergeSelected, setMergeSelected, merging, mergeError, onMerge, onClose }) {
  return (
    <Modal title={`Merge with ${mergeModal.subnet.networkAddress}/${mergeModal.subnet.prefixLength}`} onClose={onClose}>
      <div className="space-y-4">
        {mergeError && <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{mergeError}</div>}
        {mergeModal.siblings.length === 0 ? (
          <p className="text-sm text-gray-500">No sibling subnets with the same prefix length found in this network.</p>
        ) : (
          <>
            <p className="text-sm text-gray-600 dark:text-gray-400">Select subnets to merge with <strong className="font-mono">{mergeModal.subnet.networkAddress}/{mergeModal.subnet.prefixLength}</strong>:</p>
            <div className="space-y-1 max-h-48 overflow-y-auto">
              {mergeModal.siblings.map(s => (
                <label key={s.id} className="flex items-center gap-2 cursor-pointer p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-700">
                  <input
                    type="checkbox"
                    className="w-4 h-4"
                    checked={mergeSelected.includes(s.id)}
                    onChange={e => setMergeSelected(prev => e.target.checked ? [...prev, s.id] : prev.filter(id => id !== s.id))}
                  />
                  <span className="font-mono text-sm">{s.networkAddress}/{s.prefixLength}</span>
                  {s.description && <span className="text-xs text-gray-400">{s.description}</span>}
                </label>
              ))}
            </div>
            {mergeSelected.length > 0 && (
              <div className="p-3 bg-blue-50 dark:bg-blue-900/20 rounded text-sm text-blue-800 dark:text-blue-300">
                Merging {1 + mergeSelected.length} subnets with /{mergeModal.subnet.prefixLength - 1} prefix
              </div>
            )}
          </>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
          {mergeModal.siblings.length > 0 && (
            <button
              onClick={onMerge}
              disabled={merging || mergeSelected.length === 0}
              className="px-4 py-2 bg-indigo-600 text-white rounded text-sm hover:bg-indigo-700 disabled:opacity-50"
            >
              {merging ? 'Merging...' : 'Merge'}
            </button>
          )}
        </div>
      </div>
    </Modal>
  )
}
