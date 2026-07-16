import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import SafeUrlLink from '../../components/SafeUrlLink'

export default function DeviceInfoPanel({ device, typeObj, locations, cfDefs }) {
  const { t } = useTranslation()
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
      <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
        <div>
          <dt className="text-gray-500 dark:text-gray-400">{t('deviceInfo.type')}</dt>
          <dd className="text-gray-800 dark:text-gray-200 font-medium">{typeObj ? `${typeObj.icon} ${typeObj.name}` : '—'}</dd>
        </div>
        <div>
          <dt className="text-gray-500 dark:text-gray-400">{t('userTabs.privacy.status')}</dt>
          <dd className="flex items-center gap-1.5">
            <span className={`w-2 h-2 rounded-full ${device?.isOnline ? 'bg-green-500' : 'bg-gray-400'}`}></span>
            <span className={`font-medium ${device?.isOnline ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}`}>
              {device?.isOnline ? t('deviceInfo.online') : t('deviceInfo.offline')}
            </span>
          </dd>
        </div>
        <div>
          <dt className="text-gray-500 dark:text-gray-400">{t('deviceInfo.vendor')}</dt>
          <dd className="text-gray-800 dark:text-gray-200">{device?.vendor || '—'}</dd>
        </div>
        <div>
          <dt className="text-gray-500 dark:text-gray-400">{t('deviceInfo.model')}</dt>
          <dd className="text-gray-800 dark:text-gray-200">{device?.model || '—'}</dd>
        </div>
        <div>
          <dt className="text-gray-500 dark:text-gray-400">{t('deviceInfo.osVersion')}</dt>
          <dd className="text-gray-800 dark:text-gray-200">{device?.osVersion || '—'}</dd>
        </div>
        <div>
          <dt className="text-gray-500 dark:text-gray-400">{t('deviceInfo.lastPing')}</dt>
          <dd className="text-gray-800 dark:text-gray-200">
            {device?.lastPingAt ? new Date(device.lastPingAt).toLocaleString() : '—'}
          </dd>
        </div>
        {device?.locationId && (
          <div>
            <dt className="text-gray-500 dark:text-gray-400">{t('subnets.location')}</dt>
            <dd className="text-gray-800 dark:text-gray-200">
              <Link to={`/locations/${device.locationId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                {locations.find(l => l.id === device.locationId)?.name || `#${device.locationId}`}
              </Link>
            </dd>
          </div>
        )}
        {device?.rackId && (
          <div>
            <dt className="text-gray-500 dark:text-gray-400">{t('deviceInfo.rack')}</dt>
            <dd className="text-gray-800 dark:text-gray-200">
              <Link to={`/racks/${device.rackId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                {t('deviceInfo.rackHash', { id: device.rackId })}
                {device.rackUnitStart != null && ` ${t('deviceInfo.rackUnitRange', { start: device.rackUnitStart, end: device.rackUnitStart + (device.rackUnitSize ?? 1) - 1 })}`}
              </Link>
            </dd>
          </div>
        )}
      </dl>
      {cfDefs.length > 0 && device?.customFields && Object.keys(device.customFields).length > 0 && (
        <div className="mt-4 border-t dark:border-gray-700 pt-4">
          <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">{t('subnetForm.customFields')}</p>
          <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
            {cfDefs.map(def => {
              const val = device.customFields[def.name]
              if (val == null) return null
              const today = new Date().toISOString().split('T')[0]
              const isPast = def.fieldType === 'date' && val && val < today
              return (
                <div key={def.id}>
                  <dt className="text-gray-500 dark:text-gray-400">{def.label}</dt>
                  <dd className={`font-medium ${isPast ? 'text-red-600 dark:text-red-400' : 'text-gray-800 dark:text-gray-200'}`}>
                    {def.fieldType === 'url' && val ? (
                      <SafeUrlLink value={val} />
                    ) : def.fieldType === 'checkbox' ? (
                      val === 'true' ? t('common.yes') : t('common.no')
                    ) : val || '—'}
                  </dd>
                </div>
              )
            })}
          </dl>
        </div>
      )}
    </div>
  )
}
