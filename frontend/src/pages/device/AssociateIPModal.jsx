import Modal from '../../components/Modal'

export default function AssociateIPModal({ assocForm, setAssocForm, ipSearch, ipSearchResults, ipSearching, ipCreating, selectedIpLabel, saving, onSearchChange, onSelectResult, onQuickCreate, onSubmit, onClose }) {
  return (
    <Modal title="Associate IP Address" onClose={onClose}>
      <form onSubmit={onSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            IP Address <span className="text-red-500">*</span>
          </label>
          {assocForm.ip_id ? (
            <div className="flex items-center gap-2">
              <span className="flex-1 px-3 py-2 border rounded text-sm font-mono bg-gray-50 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 border-gray-300">
                {selectedIpLabel}
              </span>
              <button
                type="button"
                onClick={() => { setAssocForm(f => ({ ...f, ip_id: '' })) }}
                className="px-2 py-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-sm"
                title="Clear selection"
              >✕</button>
            </div>
          ) : (
            <div>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="Type an IP address or hostname to search…"
                value={ipSearch}
                onChange={e => onSearchChange(e.target.value)}
                autoFocus
              />
              {(ipSearching || ipSearchResults.length > 0 || (ipSearch.trim() && !ipSearching)) && (
                <div className="mt-1 border border-gray-200 dark:border-gray-700 rounded bg-white dark:bg-gray-800 max-h-48 overflow-y-auto">
                  {ipSearching && (
                    <div className="px-3 py-2 text-sm text-gray-400">Searching…</div>
                  )}
                  {!ipSearching && ipSearchResults.length === 0 && ipSearch.trim() && (
                    <div>
                      <div className="px-3 py-2 text-sm text-gray-400">No matching IPs found</div>
                      {/^[\d.:a-fA-F]+$/.test(ipSearch.trim()) && (
                        <button
                          type="button"
                          onClick={() => onQuickCreate(ipSearch.trim())}
                          disabled={ipCreating}
                          className="w-full text-left px-3 py-2 text-sm text-blue-600 dark:text-blue-400 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50"
                        >
                          {ipCreating ? 'Creating…' : `+ Create ${ipSearch.trim()} and select`}
                        </button>
                      )}
                    </div>
                  )}
                  {ipSearchResults.map(ip => (
                    <button
                      key={ip.id}
                      type="button"
                      onClick={() => onSelectResult(ip)}
                      className="w-full text-left px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 font-mono text-gray-900 dark:text-gray-100"
                    >
                      <span className="font-medium">{ip.address}</span>
                      {ip.hostname && <span className="ml-2 text-gray-400 text-xs">{ip.hostname}</span>}
                      {ip.status && <span className="ml-2 text-xs px-1 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">{ip.status}</span>}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Interface Name</label>
          <input
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            placeholder="e.g. eth0"
            value={assocForm.interface_name}
            onChange={e => setAssocForm(f => ({ ...f, interface_name: e.target.value }))}
          />
        </div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={assocForm.is_primary}
            onChange={e => setAssocForm(f => ({ ...f, is_primary: e.target.checked }))}
            className="w-4 h-4 text-blue-600 rounded"
          />
          <span className="text-sm text-gray-700 dark:text-gray-300">Primary address</span>
        </label>
        <div className="flex justify-end gap-2 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
          <button type="submit" disabled={saving || !assocForm.ip_id} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
            {saving ? 'Associating...' : 'Associate'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
