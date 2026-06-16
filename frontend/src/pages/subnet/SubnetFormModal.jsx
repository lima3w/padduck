import Modal from '../../components/Modal'
import CustomFieldForm from '../../components/CustomFieldForm'

export default function SubnetFormModal({ modal, form, setForm, overlapError, saving, locations, nameservers, vlans, cfDefs, onSubmit, onClose }) {
  const isCreate = modal === 'create'
  return (
    <Modal title={isCreate ? 'New Subnet' : 'Edit Subnet'} onClose={onClose}>
      <form onSubmit={onSubmit} className="space-y-4">
        {overlapError && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{overlapError}</div>
        )}
        {isCreate && (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Network Address</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="192.168.0.0"
                value={form.network_address}
                onChange={e => setForm(f => ({ ...f, network_address: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Prefix Length</label>
              <input
                type="number" min="0" max="32"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                placeholder="24"
                value={form.prefix_length}
                onChange={e => setForm(f => ({ ...f, prefix_length: e.target.value }))}
                required
              />
            </div>
          </>
        )}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
          <input
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={form.description}
            onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Gateway (optional)</label>
          <input
            className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={form.gateway}
            onChange={e => setForm(f => ({ ...f, gateway: e.target.value }))}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Location (optional)</label>
          <select
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={form.location_id}
            onChange={e => setForm(f => ({ ...f, location_id: e.target.value }))}
          >
            <option value="">No location</option>
            {locations.map(l => (
              <option key={l.id} value={l.id}>{l.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Nameserver (optional)</label>
          <select
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={form.nameserver_id}
            onChange={e => setForm(f => ({ ...f, nameserver_id: e.target.value }))}
          >
            <option value="">No nameserver</option>
            {nameservers.map(ns => (
              <option key={ns.id} value={ns.id}>{ns.name} ({ns.server1})</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">VLAN (optional)</label>
          <select
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={form.vlan_id}
            onChange={e => setForm(f => ({ ...f, vlan_id: e.target.value }))}
          >
            <option value="">No VLAN</option>
            {vlans.map(vlan => (
              <option key={vlan.id} value={vlan.id}>VLAN {vlan.vlanId} — {vlan.name}</option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={form.auto_reserve_first}
              onChange={e => setForm(f => ({ ...f, auto_reserve_first: e.target.checked }))}
              className="w-4 h-4 text-blue-600 rounded"
            />
            <span className="text-sm text-gray-700">Auto-reserve first IP (network address)</span>
          </label>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={form.auto_reserve_last}
              onChange={e => setForm(f => ({ ...f, auto_reserve_last: e.target.checked }))}
              className="w-4 h-4 text-blue-600 rounded"
            />
            <span className="text-sm text-gray-700">Auto-reserve last IP (broadcast address)</span>
          </label>
        </div>
        <div className="border-t dark:border-gray-600 pt-4 space-y-4">
          <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Technitium DHCP</p>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Scope name (optional)</label>
            <input
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              placeholder="e.g. Office LAN"
              value={form.technitium_scope_name || ''}
              onChange={e => setForm(f => ({ ...f, technitium_scope_name: e.target.value }))}
            />
            <p className="text-xs text-gray-400 mt-1">Links this subnet to a Technitium DHCP scope for lease sync and reservation push.</p>
          </div>
        </div>
        <div className="border-t dark:border-gray-600 pt-4 space-y-4">
          <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Alert Settings</p>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Alert Threshold % (optional)</label>
            <input
              type="number" min="1" max="100"
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              placeholder="e.g. 80"
              value={form.alert_threshold_pct}
              onChange={e => setForm(f => ({ ...f, alert_threshold_pct: e.target.value }))}
            />
            <p className="text-xs text-gray-400 mt-1">Send alert when utilization exceeds this percentage</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Alert Email Override (optional)</label>
            <input
              type="email"
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              placeholder="alerts@example.com"
              value={form.alert_email_override}
              onChange={e => setForm(f => ({ ...f, alert_email_override: e.target.value }))}
            />
            <p className="text-xs text-gray-400 mt-1">Override the default alert recipient for this subnet</p>
          </div>
        </div>
        {cfDefs.length > 0 && (
          <div className="border-t dark:border-gray-600 pt-4">
            <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Custom Fields</p>
            <CustomFieldForm
              definitions={cfDefs}
              values={form.custom_fields}
              onChange={(name, value) => setForm(f => ({ ...f, custom_fields: { ...f.custom_fields, [name]: value } }))}
            />
          </div>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
          <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
