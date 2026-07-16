import { useTranslation } from 'react-i18next'
import Modal from '../../components/Modal'

const MEDIA_TYPES = ['copper', 'fiber', 'SFP', 'SFP+', 'QSFP', 'other']

export default function InterfaceModal({ modal, ifaceForm, setIfaceForm, vlanList, saving, onSubmit, onClose }) {
  const { t } = useTranslation()
  const isAdd = modal === 'iface-add'
  return (
    <Modal title={isAdd ? t('interfaceModal.addTitle') : t('interfaceModal.editTitle')} onClose={onClose}>
      <form onSubmit={onSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.name')} <span className="text-red-500">*</span></label>
          <input
            className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            placeholder="eth0"
            value={ifaceForm.name}
            onChange={e => setIfaceForm(f => ({ ...f, name: e.target.value }))}
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.description')}</label>
          <input
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            placeholder="Uplink to core switch"
            value={ifaceForm.description}
            onChange={e => setIfaceForm(f => ({ ...f, description: e.target.value }))}
          />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('interfaceModal.speedMbpsLabel')}</label>
            <input
              type="number"
              min="0"
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              placeholder="1000"
              value={ifaceForm.speed_mbps}
              onChange={e => setIfaceForm(f => ({ ...f, speed_mbps: e.target.value }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('interfaceModal.mediaType')}</label>
            <select
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              value={ifaceForm.media_type}
              onChange={e => setIfaceForm(f => ({ ...f, media_type: e.target.value }))}
            >
              <option value="">{t('interfaceModal.selectPlaceholder')}</option>
              {MEDIA_TYPES.map(m => <option key={m} value={m}>{m}</option>)}
            </select>
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('subnets.vlan')}</label>
          <select
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
            value={ifaceForm.vlan_id}
            onChange={e => setIfaceForm(f => ({ ...f, vlan_id: e.target.value }))}
          >
            <option value="">{t('interfaceModal.none')}</option>
            {vlanList.map(v => (
              <option key={v.id} value={v.id}>{t('subnets.vlan')} {v.vlanId} — {v.name}</option>
            ))}
          </select>
          {vlanList.length === 0 && (
            <p className="text-xs text-gray-400 mt-1">{t('interfaceModal.noVlansPrefix')}<a href="/vlans" className="text-blue-500 hover:underline">{t('nav.vlans')}</a>{t('interfaceModal.noVlansSuffix')}</p>
          )}
        </div>
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
