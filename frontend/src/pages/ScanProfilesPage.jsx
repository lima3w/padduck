import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getScanProfiles, createScanProfile, updateScanProfile, deleteScanProfile } from '../api/admin'
import Modal from '../components/Modal'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const SCAN_TYPE_LABEL_KEYS = { ping: 'scanTypePing', snmp: 'scanTypeSnmp', 'ping+snmp': 'scanTypePingSnmp' }

const EMPTY_FORM = {
  name: '',
  description: '',
  scan_type: 'ping',
  ping_concurrency: 20,
  tcp_ports: '',
  dns_lookup: false,
  snmp_community: '',
  snmp_version: 'v2c',
}

export default function ScanProfilesPage() {
  const { t } = useTranslation()
  const [profiles, setProfiles] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [modal, setModal] = useState(null) // null | 'create' | { edit: profile }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => {
    loadProfiles()
  }, [])

  async function loadProfiles() {
    try {
      const { data } = await getScanProfiles()
      setProfiles(data || [])
    } catch {
      setError(t('scanProfilesPage.loadFailed'))
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setModal('create')
  }

  function openEdit(profile) {
    setForm({
      name: profile.name || '',
      description: profile.description || '',
      scan_type: profile.scanType || 'ping',
      ping_concurrency: profile.pingConcurrency || 20,
      tcp_ports: profile.tcpPorts || '',
      dns_lookup: profile.dnsLookup || false,
      snmp_community: profile.snmpCommunity || '',
      snmp_version: profile.snmpVersion || 'v2c',
    })
    setModal({ edit: profile })
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      const body = {
        name: form.name,
        description: form.description || null,
        scan_type: form.scan_type,
        ping_concurrency: Number(form.ping_concurrency) || 20,
        tcp_ports: form.tcp_ports || null,
        dns_lookup: form.dns_lookup,
        snmp_community: form.snmp_community || null,
        snmp_version: form.snmp_version,
      }
      if (modal === 'create') {
        await createScanProfile(body)
      } else {
        await updateScanProfile(modal.edit.id, body)
      }
      setModal(null)
      await loadProfiles()
    } catch (err) {
      setError(err.response?.data?.error || t('scanProfilesPage.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await deleteScanProfile(id)
      setDeleteConfirm(null)
      await loadProfiles()
    } catch {
      setError(t('scanProfilesPage.deleteFailed'))
    }
  }

  if (loading) return <PageSpinner message={t('scanProfilesPage.loadingScanProfiles')} />

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">{t('scanProfilesPage.title')}</h1>
        <button
          onClick={openCreate}
          className="text-sm bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 transition"
        >
          {t('scanProfilesPage.newProfile')}
        </button>
      </div>

      <ErrorBanner error={error} />

      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('common.name')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('scanJobs.scanTypeLabel')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('scanJobs.concurrencyLabel')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('scanProfilesPage.tcpPortsColumn')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('scanProfilesPage.dnsLookupColumn')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('scanProfilesPage.snmpCommunityColumn')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('credentials.snmpVersion')}</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">{t('vrfs.actions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {profiles.length === 0 ? (
                <EmptyRow colSpan={8} message={t('scanProfilesPage.noScanProfilesConfigured')} />
              ) : (
                profiles.map((profile) => (
                  <tr key={profile.id} className="hover:bg-gray-50">
                    <td className="px-4 py-2 font-medium text-gray-900">
                      {profile.name}
                      {profile.description && (
                        <p className="text-xs text-gray-500 font-normal mt-0.5">{profile.description}</p>
                      )}
                    </td>
                    <td className="px-4 py-2 text-gray-700">{profile.scanType && SCAN_TYPE_LABEL_KEYS[profile.scanType] ? t(`scanJobs.${SCAN_TYPE_LABEL_KEYS[profile.scanType]}`) : profile.scanType}</td>
                    <td className="px-4 py-2 text-gray-700">{profile.pingConcurrency}</td>
                    <td className="px-4 py-2 font-mono text-xs text-gray-700">{profile.tcpPorts || <span className="text-gray-300">—</span>}</td>
                    <td className="px-4 py-2">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${profile.dnsLookup ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                        {profile.dnsLookup ? t('common.yes') : t('common.no')}
                      </span>
                    </td>
                    <td className="px-4 py-2 text-gray-700">{profile.snmpCommunity || <span className="text-gray-300">—</span>}</td>
                    <td className="px-4 py-2 text-gray-700">{profile.snmpVersion}</td>
                    <td className="px-4 py-2">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => openEdit(profile)}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          {t('common.edit')}
                        </button>
                        {deleteConfirm === profile.id ? (
                          <span className="flex items-center gap-1 text-xs">
                            <span className="text-red-600">{t('adminAgents.deleteConfirm')}</span>
                            <button onClick={() => handleDelete(profile.id)} className="text-red-600 font-medium hover:text-red-800">{t('common.yes')}</button>
                            <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">{t('common.no')}</button>
                          </span>
                        ) : (
                          <button
                            onClick={() => setDeleteConfirm(profile.id)}
                            className="text-xs text-red-500 hover:underline"
                          >
                            {t('common.delete')}
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {modal && (
        <Modal
          title={modal === 'create' ? t('scanProfilesPage.newProfileModalTitle') : `${t('scanProfilesPage.editProfileModalTitlePrefix')}${modal.edit?.name}`}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('common.name')} <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Office ICMP scan"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('common.description')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder={t('adminRoles.descriptionPlaceholder')}
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('scanJobs.scanTypeLabel')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.scan_type}
                onChange={e => setForm(f => ({ ...f, scan_type: e.target.value }))}
              >
                <option value="ping">{t('scanJobs.scanTypePing')}</option>
                <option value="snmp">{t('scanJobs.scanTypeSnmp')}</option>
                <option value="ping+snmp">{t('scanJobs.scanTypePingSnmp')}</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('scanProfilesPage.pingConcurrencyLabel')} <span className="text-gray-400 font-normal">{t('scanProfilesPage.concurrencyRangeHint')}</span>
              </label>
              <input
                type="number"
                min="1"
                max="100"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.ping_concurrency}
                onChange={e => setForm(f => ({ ...f, ping_concurrency: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('scanProfilesPage.tcpPortsLabel')} <span className="text-gray-400 font-normal">{t('scanProfilesPage.tcpPortsHint')}</span>
              </label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="22,80,443"
                value={form.tcp_ports}
                onChange={e => setForm(f => ({ ...f, tcp_ports: e.target.value }))}
              />
            </div>
            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.dns_lookup}
                  onChange={e => setForm(f => ({ ...f, dns_lookup: e.target.checked }))}
                  className="w-4 h-4"
                />
                <span className="text-sm text-gray-700">{t('scanProfilesPage.dnsLookupLabel')}</span>
              </label>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('scanProfilesPage.snmpCommunityLabel')}</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="public"
                value={form.snmp_community}
                onChange={e => setForm(f => ({ ...f, snmp_community: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('credentials.snmpVersion')}</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.snmp_version}
                onChange={e => setForm(f => ({ ...f, snmp_version: e.target.value }))}
              >
                <option value="v2c">v2c</option>
                <option value="v3">v3</option>
              </select>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setModal(null)}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                {t('common.cancel')}
              </button>
              <button
                type="submit"
                disabled={saving}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? t('common.saving') : modal === 'create' ? t('vrfs.create') : t('common.save')}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
