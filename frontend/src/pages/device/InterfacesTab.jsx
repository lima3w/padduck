import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function InterfacesTab({ interfaces, vlanList, deleteIfaceConfirm, setDeleteIfaceConfirm, onAddInterface, onEditInterface, onDeleteInterface }) {
  const { t } = useTranslation()
  return (
    <div>
      <div className="flex justify-end mb-3">
        <button onClick={onAddInterface} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          {t('interfaces.addInterface')}
        </button>
      </div>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.name')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('common.description')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('interfaces.speed')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('interfaces.media')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('subnets.vlan')}</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{t('interfaces.connectedTo')}</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {interfaces.length === 0 && (
              <tr><td colSpan={7} className="px-4 py-6 text-center text-gray-400">{t('interfaces.noInterfacesDefined')}</td></tr>
            )}
            {interfaces.map(iface => (
              <tr key={iface.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">{iface.name}</td>
                <td className="px-4 py-3 text-gray-700 dark:text-gray-200">{iface.description || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {iface.speedMbps ? t('interfaces.speedMbps', { speed: iface.speedMbps }) : '—'}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{iface.mediaType || '—'}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {iface.vlanId ? (() => { const v = vlanList.find(x => x.id === iface.vlanId); return v ? `${t('subnets.vlan')} ${v.vlanId} — ${v.name}` : t('interfaces.vlanHash', { id: iface.vlanId }) })() : '—'}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                  {iface.connectedToDeviceId ? (
                    <Link to={`/devices/${iface.connectedToDeviceId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                      {t('interfaces.deviceHash', { id: iface.connectedToDeviceId })}
                    </Link>
                  ) : '—'}
                </td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => onEditInterface(iface)} className="text-gray-400 hover:text-blue-600 text-xs">{t('common.edit')}</button>
                  {deleteIfaceConfirm === iface.id ? (
                    <>
                      <span className="text-red-600 text-xs">{t('interfaces.confirmDelete')}</span>
                      <button onClick={() => onDeleteInterface(iface.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">{t('common.yes')}</button>
                      <button onClick={() => setDeleteIfaceConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">{t('common.no')}</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteIfaceConfirm(iface.id)} className="text-gray-400 hover:text-red-600 text-xs">{t('common.delete')}</button>
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
