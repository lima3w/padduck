import { useState, useEffect } from 'react'
import { api } from '../api/client'
import Modal from '../components/Modal'

const REPORT_TYPES = [
  { value: 'utilisation_summary', label: 'Utilisation Summary' },
  { value: 'inactive_ips', label: 'Inactive IPs' },
  { value: 'new_allocations', label: 'New Allocations' },
  { value: 'full_audit', label: 'Full Audit' },
  { value: 'subnet_gaps', label: 'Subnet Gaps' },
  { value: 'vlan_assignment', label: 'VLAN Assignment' },
  { value: 'ip_age', label: 'IP Age' },
  { value: 'dns_audit', label: 'DNS Audit' },
  { value: 'stale_leases', label: 'Stale Leases' },
  { value: 'inactive_devices', label: 'Inactive Devices' },
  { value: 'failed_scans', label: 'Failed Scans' },
]

const FORMAT_OPTIONS = [
  { value: 'csv', label: 'CSV' },
  { value: 'pdf', label: 'PDF' },
]

const DEFAULT_FILTERS = {
  utilisation_summary: '{}',
  inactive_ips: '{"days": 90}',
  new_allocations: '{"days": 30}',
  full_audit: '{}',
  subnet_gaps: '{}',
  vlan_assignment: '{}',
  ip_age: '{}',
  dns_audit: '{}',
  stale_leases: '{"days": 30}',
  inactive_devices: '{"days": 30}',
  failed_scans: '{"days": 7}',
}

const EMPTY_FORM = {
  name: '',
  reportType: 'utilisation_summary',
  scheduleCron: '0 8 * * 1',
  recipientEmails: '',
  format: 'csv',
  filters: DEFAULT_FILTERS.utilisation_summary,
}

export default function ScheduledReportsPage() {
  const [reports, setReports] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [modal, setModal] = useState(null) // null | 'create' | { edit: report }
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [running, setRunning] = useState(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      const { data } = await api.get('/admin/reports/scheduled')
      setReports(Array.isArray(data) ? data : [])
    } catch {
      setError('Failed to load scheduled reports')
    } finally {
      setLoading(false)
    }
  }

  function openCreate() {
    setForm(EMPTY_FORM)
    setFormError('')
    setModal('create')
  }

  function openEdit(report) {
    setForm({
      name: report.name || '',
      reportType: report.reportType || 'utilisation_summary',
      scheduleCron: report.scheduleCron || '0 8 * * 1',
      recipientEmails: Array.isArray(report.recipientEmails)
        ? report.recipientEmails.join(', ')
        : (report.recipientEmails || ''),
      format: report.format || 'csv',
      filters: report.filters ? JSON.stringify(report.filters, null, 2) : '{}',
    })
    setFormError('')
    setModal({ edit: report })
  }

  function handleTypeChange(type) {
    setForm(f => ({ ...f, reportType: type, filters: DEFAULT_FILTERS[type] || '{}' }))
  }

  function validateFilters(str) {
    try { JSON.parse(str); return null } catch { return 'Filters must be valid JSON' }
  }

  async function handleSubmit(e) {
    e.preventDefault()
    const filterErr = validateFilters(form.filters)
    if (filterErr) { setFormError(filterErr); return }

    setSaving(true)
    setFormError('')
    try {
      const body = {
        name: form.name,
        report_type: form.reportType,
        schedule_cron: form.scheduleCron,
        recipient_emails: form.recipientEmails.split(',').map(s => s.trim()).filter(Boolean),
        format: form.format,
        filters: JSON.parse(form.filters),
      }
      if (modal === 'create') {
        await api.post('/admin/reports/scheduled', body)
      } else {
        await api.put(`/admin/reports/scheduled/${modal.edit.id}`, body)
      }
      setModal(null)
      load()
    } catch (err) {
      setFormError(err.response?.data?.error || 'Failed to save report')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    try {
      await api.delete(`/admin/reports/scheduled/${id}`)
      setDeleteConfirm(null)
      load()
    } catch {
      setError('Failed to delete report')
    }
  }

  async function handleRunNow(id) {
    setRunning(id)
    try {
      await api.post(`/admin/reports/scheduled/${id}/run`)
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to run report')
    } finally {
      setRunning(null)
    }
  }

  const inputClass = "w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
  const labelClass = "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"

  if (loading) return <p className="text-gray-500">Loading scheduled reports...</p>

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Scheduled Reports</h1>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium">
          + New Report
        </button>
      </div>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
            <tr>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Name</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Schedule</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Recipients</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Format</th>
              <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Last Run</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {reports.length === 0 && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-gray-400">No scheduled reports yet</td>
              </tr>
            )}
            {reports.map(r => (
              <tr key={r.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{r.name}</td>
                <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                  {REPORT_TYPES.find(t => t.value === r.reportType)?.label || r.reportType}
                </td>
                <td className="px-4 py-3 font-mono text-gray-500 dark:text-gray-400 text-xs">{r.scheduleCron}</td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                  {Array.isArray(r.recipientEmails) ? r.recipientEmails.join(', ') : (r.recipientEmails || '—')}
                </td>
                <td className="px-4 py-3 text-gray-500 dark:text-gray-400 uppercase text-xs">{r.format}</td>
                <td className="px-4 py-3 text-gray-400 text-xs">
                  {r.lastRunAt ? new Date(r.lastRunAt).toLocaleString() : 'Never'}
                </td>
                <td className="px-4 py-3 text-right space-x-2 whitespace-nowrap">
                  <button
                    onClick={() => handleRunNow(r.id)}
                    disabled={running === r.id}
                    className="text-gray-400 hover:text-green-600 text-xs disabled:opacity-50"
                  >
                    {running === r.id ? 'Running...' : 'Run Now'}
                  </button>
                  <button onClick={() => openEdit(r)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                  {deleteConfirm === r.id ? (
                    <>
                      <span className="text-red-600 text-xs">Confirm?</span>
                      <button onClick={() => handleDelete(r.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                    </>
                  ) : (
                    <button onClick={() => setDeleteConfirm(r.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      </div>

      {modal && (
        <Modal title={modal === 'create' ? 'New Scheduled Report' : 'Edit Report'} onClose={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="space-y-4">
            {formError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{formError}</div>
            )}
            <div>
              <label className={labelClass}>Name</label>
              <input className={inputClass} value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} required placeholder="Weekly Utilisation Report" />
            </div>
            <div>
              <label className={labelClass}>Report Type</label>
              <select className={inputClass} value={form.reportType} onChange={e => handleTypeChange(e.target.value)}>
                {REPORT_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </div>
            <div>
              <label className={labelClass}>Schedule (cron expression)</label>
              <input className={inputClass} value={form.scheduleCron} onChange={e => setForm(f => ({ ...f, scheduleCron: e.target.value }))} required placeholder="0 8 * * 1" />
              <p className="text-xs text-gray-400 mt-1">e.g. &apos;0 8 * * 1&apos; for weekly Monday 8am</p>
            </div>
            <div>
              <label className={labelClass}>Recipients (comma-separated emails)</label>
              <input className={inputClass} value={form.recipientEmails} onChange={e => setForm(f => ({ ...f, recipientEmails: e.target.value }))} placeholder="admin@example.com, ops@example.com" />
            </div>
            <div>
              <label className={labelClass}>Format</label>
              <select className={inputClass} value={form.format} onChange={e => setForm(f => ({ ...f, format: e.target.value }))}>
                {FORMAT_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
            <div>
              <label className={labelClass}>Filters (JSON)</label>
              <textarea
                className={`${inputClass} font-mono`}
                rows={4}
                value={form.filters}
                onChange={e => setForm(f => ({ ...f, filters: e.target.value }))}
                placeholder="{}"
              />
              <p className="text-xs text-gray-400 mt-1">Advanced: JSON object of filter parameters</p>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
