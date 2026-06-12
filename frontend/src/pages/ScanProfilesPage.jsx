import { useState, useEffect } from 'react'
import { getScanProfiles, createScanProfile, updateScanProfile, deleteScanProfile } from '../api/admin'
import Modal from '../components/Modal'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const SCAN_TYPE_LABELS = { ping: 'Ping', snmp: 'SNMP', 'ping+snmp': 'Ping + SNMP' }

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
      setError('Failed to load scan profiles')
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
      setError(err.response?.data?.error || 'Failed to save scan profile')
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
      setError('Failed to delete scan profile')
    }
  }

  if (loading) return <PageSpinner message="Loading scan profiles..." />

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Scan Profiles</h1>
        <button
          onClick={openCreate}
          className="text-sm bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 transition"
        >
          + New Profile
        </button>
      </div>

      <ErrorBanner error={error} />

      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Name</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Scan Type</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Concurrency</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">TCP Ports</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">DNS Lookup</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">SNMP Community</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">SNMP Version</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {profiles.length === 0 ? (
                <EmptyRow colSpan={8} message="No scan profiles configured." />
              ) : (
                profiles.map((profile) => (
                  <tr key={profile.id} className="hover:bg-gray-50">
                    <td className="px-4 py-2 font-medium text-gray-900">
                      {profile.name}
                      {profile.description && (
                        <p className="text-xs text-gray-500 font-normal mt-0.5">{profile.description}</p>
                      )}
                    </td>
                    <td className="px-4 py-2 text-gray-700">{SCAN_TYPE_LABELS[profile.scanType] || profile.scanType}</td>
                    <td className="px-4 py-2 text-gray-700">{profile.pingConcurrency}</td>
                    <td className="px-4 py-2 font-mono text-xs text-gray-700">{profile.tcpPorts || <span className="text-gray-300">—</span>}</td>
                    <td className="px-4 py-2">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${profile.dnsLookup ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                        {profile.dnsLookup ? 'Yes' : 'No'}
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
                          Edit
                        </button>
                        {deleteConfirm === profile.id ? (
                          <span className="flex items-center gap-1 text-xs">
                            <span className="text-red-600">Delete?</span>
                            <button onClick={() => handleDelete(profile.id)} className="text-red-600 font-medium hover:text-red-800">Yes</button>
                            <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">No</button>
                          </span>
                        ) : (
                          <button
                            onClick={() => setDeleteConfirm(profile.id)}
                            className="text-xs text-red-500 hover:underline"
                          >
                            Delete
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
          title={modal === 'create' ? 'New Scan Profile' : `Edit: ${modal.edit?.name}`}
          onClose={() => setModal(null)}
        >
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Name <span className="text-red-500">*</span>
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
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Optional description"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Scan Type</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={form.scan_type}
                onChange={e => setForm(f => ({ ...f, scan_type: e.target.value }))}
              >
                <option value="ping">Ping</option>
                <option value="snmp">SNMP</option>
                <option value="ping+snmp">Ping + SNMP</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Ping Concurrency <span className="text-gray-400 font-normal">(1–100)</span>
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
                TCP Ports <span className="text-gray-400 font-normal">(comma-separated, e.g. 22,80,443)</span>
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
                <span className="text-sm text-gray-700">DNS Lookup</span>
              </label>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">SNMP Community</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="public"
                value={form.snmp_community}
                onChange={e => setForm(f => ({ ...f, snmp_community: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">SNMP Version</label>
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
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? 'Saving...' : modal === 'create' ? 'Create' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
