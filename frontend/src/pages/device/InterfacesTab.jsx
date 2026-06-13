import { Link } from 'react-router-dom'

export default function InterfacesTab({ interfaces, vlanList, deleteIfaceConfirm, setDeleteIfaceConfirm, onAddInterface, onEditInterface, onDeleteInterface }) {
  return (
    <div>
      <div className="flex justify-end mb-3">
        <button onClick={onAddInterface} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + Add Interface
        </button>
      </div>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Description</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Speed</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Media</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">VLAN</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Connected To</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {interfaces.length === 0 && (
              <tr><td colSpan={7} className="px-4 py-6 text-center text-gray-400">No interfaces defined</td></tr>
            )}
            {interfaces.map(iface => (
              <tr key={iface.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{iface.name}</td>
                <td className="px-4 py-3 text-gray-700 dark:text-gray-200">{iface.description || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {iface.speedMbps ? `${iface.speedMbps} Mbps` : '—'}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{iface.mediaType || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {iface.vlanId ? (() => { const v = vlanList.find(x => x.id === iface.vlanId); return v ? `VLAN ${v.vlanId} — ${v.name}` : `VLAN #${iface.vlanId}` })() : '—'}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {iface.connectedToDeviceId ? (
                    <Link to={`/devices/${iface.connectedToDeviceId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                      Device #{iface.connectedToDeviceId}
                    </Link>
                  ) : '—'}
                </td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => onEditInterface(iface)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                  {deleteIfaceConfirm === iface.id ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => onDeleteInterface(iface.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteIfaceConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteIfaceConfirm(iface.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>
    </div>
  )
}
