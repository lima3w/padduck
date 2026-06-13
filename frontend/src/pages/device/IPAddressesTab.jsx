export default function IPAddressesTab({ ipAddresses, deleteIpConfirm, setDeleteIpConfirm, onAssociateIP, onDisassociateIP }) {
  return (
    <div>
      <div className="flex justify-end mb-3">
        <button onClick={onAssociateIP} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + Associate IP
        </button>
      </div>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Address</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Interface</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Primary</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Subnet</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {ipAddresses.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-400">No IP addresses associated</td></tr>
            )}
            {ipAddresses.map(ip => (
              <tr key={ip.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{ip.address}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.interfaceName || '—'}</td>
                <td className="px-4 py-3">
                  {ip.isPrimary && (
                    <span className="inline-block px-2 py-0.5 bg-blue-100 text-blue-700 text-xs font-medium rounded">Primary</span>
                  )}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{ip.subnetId ? `#${ip.subnetId}` : '—'}</td>
                <td className="px-4 py-3 text-right">
                  {deleteIpConfirm === ip.id ? (
                    <span className="space-x-2">
                      <span className="text-red-600 text-xs">Remove?</span>
                      <button onClick={() => onDisassociateIP(ip.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteIpConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </span>
                  ) : (
                    <button onClick={() => setDeleteIpConfirm(ip.id)} className="text-gray-400 hover:text-red-600 text-xs">Remove</button>
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
