import { useTranslation } from 'react-i18next'
import Modal from '../../components/Modal'
import CustomFieldForm from '../../components/CustomFieldForm'

export default function EditDeviceModal({ editForm, setEditForm, deviceTypes, locations, racks, cfDefs, saving, onSubmit, onClose, onLocationChange }) {
  const { t } = useTranslation()
  return (
    <Modal title={t('editDevice.title')} onClose={onClose}>
      <form onSubmit={onSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('editDevice.hostname')} <span className="text-red-500">*</span></label>
          <input
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={editForm.hostname}
            onChange={e => setEditForm(f => ({ ...f, hostname: e.target.value }))}
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('deviceInfo.type')}</label>
          <select
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={editForm.type_id}
            onChange={e => setEditForm(f => ({ ...f, type_id: e.target.value }))}
          >
            <option value="">{t('editDevice.noType')}</option>
            {deviceTypes.map(dt => (
              <option key={dt.id} value={dt.id}>{dt.icon} {dt.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.description')}</label>
          <textarea
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            rows={2}
            value={editForm.description}
            onChange={e => setEditForm(f => ({ ...f, description: e.target.value }))}
          />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('deviceInfo.vendor')}</label>
            <input
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              value={editForm.vendor}
              onChange={e => setEditForm(f => ({ ...f, vendor: e.target.value }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('deviceInfo.model')}</label>
            <input
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              value={editForm.model}
              onChange={e => setEditForm(f => ({ ...f, model: e.target.value }))}
            />
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('deviceInfo.osVersion')}</label>
          <input
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={editForm.os_version}
            onChange={e => setEditForm(f => ({ ...f, os_version: e.target.value }))}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('subnetForm.locationOptional')}</label>
          <select
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={editForm.location_id || ''}
            onChange={e => {
              const locId = e.target.value
              setEditForm(f => ({ ...f, location_id: locId, rack_id: '', rack_unit_start: '', rack_unit_size: '' }))
              onLocationChange(locId)
            }}
          >
            <option value="">{t('subnetForm.noLocation')}</option>
            {locations.map(l => (
              <option key={l.id} value={l.id}>{l.name}</option>
            ))}
          </select>
        </div>
        {editForm.location_id && (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('editDevice.rackOptional')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={editForm.rack_id || ''}
                onChange={e => setEditForm(f => ({ ...f, rack_id: e.target.value }))}
              >
                <option value="">{t('editDevice.noRack')}</option>
                {racks.map(r => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </div>
            {editForm.rack_id && (
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('editDevice.rackUnitStart')}</label>
                  <input
                    type="number"
                    min="1"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder="1"
                    value={editForm.rack_unit_start || ''}
                    onChange={e => setEditForm(f => ({ ...f, rack_unit_start: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('editDevice.rackUnitSize')}</label>
                  <input
                    type="number"
                    min="1"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder="1"
                    value={editForm.rack_unit_size || ''}
                    onChange={e => setEditForm(f => ({ ...f, rack_unit_size: e.target.value }))}
                  />
                </div>
              </div>
            )}
          </>
        )}
        <div className="border-t dark:border-gray-600 pt-4">
          <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">{t('editDevice.snmp')}</p>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('credentials.snmpVersion')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={editForm.snmp_version}
                onChange={e => setEditForm(f => ({ ...f, snmp_version: e.target.value }))}
              >
                <option value="v1">v1</option>
                <option value="v2c">v2c</option>
                <option value="v3">v3</option>
              </select>
            </div>
            {(editForm.snmp_version === 'v1' || editForm.snmp_version === 'v2c') && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('credentials.communityString')}</label>
                <input
                  type="text"
                  className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                  placeholder={t('editDevice.leaveBlankToKeep')}
                  value={editForm.snmp_community}
                  onChange={e => setEditForm(f => ({ ...f, snmp_community: e.target.value }))}
                />
              </div>
            )}
            {editForm.snmp_version === 'v3' && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('login.username')}</label>
                  <input
                    type="text"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    value={editForm.snmp_v3_user}
                    onChange={e => setEditForm(f => ({ ...f, snmp_v3_user: e.target.value }))}
                  />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('credentials.authProtocol')}</label>
                    <select
                      className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                      value={editForm.snmp_v3_auth_proto}
                      onChange={e => setEditForm(f => ({ ...f, snmp_v3_auth_proto: e.target.value }))}
                    >
                      <option value="SHA">SHA</option>
                      <option value="MD5">MD5</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('credentials.authPassword')}</label>
                    <input
                      type="password"
                      className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                      placeholder={t('editDevice.leaveBlankToKeep')}
                      value={editForm.snmp_v3_auth_pass}
                      onChange={e => setEditForm(f => ({ ...f, snmp_v3_auth_pass: e.target.value }))}
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('credentials.privProtocol')}</label>
                    <select
                      className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                      value={editForm.snmp_v3_priv_proto}
                      onChange={e => setEditForm(f => ({ ...f, snmp_v3_priv_proto: e.target.value }))}
                    >
                      <option value="AES">AES</option>
                      <option value="DES">DES</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('credentials.privPassword')}</label>
                    <input
                      type="password"
                      className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                      placeholder={t('editDevice.leaveBlankToKeep')}
                      value={editForm.snmp_v3_priv_pass}
                      onChange={e => setEditForm(f => ({ ...f, snmp_v3_priv_pass: e.target.value }))}
                    />
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
        {cfDefs.length > 0 && (
          <div className="border-t dark:border-gray-600 pt-4">
            <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">{t('subnetForm.customFields')}</p>
            <CustomFieldForm
              definitions={cfDefs}
              values={editForm.custom_fields}
              onChange={(name, value) => setEditForm(f => ({ ...f, custom_fields: { ...f.custom_fields, [name]: value } }))}
            />
          </div>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t('common.cancel')}</button>
          <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
            {saving ? t('common.saving') : t('common.save')}
          </button>
        </div>
      </form>
    </Modal>
  )
}
