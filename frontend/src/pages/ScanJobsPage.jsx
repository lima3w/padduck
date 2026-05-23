import { useState, useEffect } from 'react'
import { api } from '../api/client'
import Modal from '../components/Modal'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const SCAN_TYPE_LABELS = { ping: 'Ping', snmp: 'SNMP', 'ping+snmp': 'Ping + SNMP' }

function formatDate(val) {
  if (!val) return '—'
  const d = new Date(val)
  return isNaN(d.getTime()) ? '—' : d.toLocaleString()
}

const CHANGE_COLORS = {
  new: 'bg-green-50 text-green-800',
  gone: 'bg-red-50 text-red-800',
  changed: 'bg-yellow-50 text-yellow-800',
}

const EMPTY_JOB_FORM = { name: '', subnet: '', schedule_cron: '', is_active: true, scan_type: 'ping', notify_on_change: false }

export default function ScanJobsPage() {
  const [jobs, setJobs] = useState([])
  const [selectedJob, setSelectedJob] = useState(null)
  const [results, setResults] = useState([])
  const [history, setHistory] = useState([])
  const [selectedRun, setSelectedRun] = useState(null)
  const [runDetail, setRunDetail] = useState(null)
  const [activeTab, setActiveTab] = useState('results')
  const [loading, setLoading] = useState(true)
  const [running, setRunning] = useState(null)
  const [error, setError] = useState('')
  const [jobModal, setJobModal] = useState(null)
  const [jobForm, setJobForm] = useState(EMPTY_JOB_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => {
    loadJobs()
  }, [])

  async function loadJobs() {
    try {
      const { data } = await api.get('/admin/scan-jobs')
      setJobs(data || [])
    } catch {
      setError('Failed to load scan jobs')
    } finally {
      setLoading(false)
    }
  }

  async function loadResults(jobId) {
    try {
      const { data } = await api.get(`/admin/scan-jobs/${jobId}/results?limit=200`)
      setResults(data || [])
    } catch {
      setResults([])
    }
  }

  async function loadHistory(jobId) {
    try {
      const { data } = await api.get(`/admin/scan-jobs/${jobId}/history`)
      setHistory(data || [])
    } catch {
      setHistory([])
    }
  }

  async function loadRunDetail(jobId, runId) {
    try {
      const { data } = await api.get(`/admin/scan-jobs/${jobId}/history/${runId}`)
      // Flatten run + changes into a single object so templates can access fields directly
      setRunDetail({ ...data.run, changes: data.changes })
    } catch {
      setRunDetail(null)
    }
  }

  async function runNow(job) {
    setRunning(job.id)
    try {
      await api.post(`/admin/scan-jobs/${job.id}/run`)
    } catch {
      /* ignore */
    } finally {
      setRunning(null)
    }
  }

  function selectJob(job) {
    setSelectedJob(job)
    setResults([])
    setHistory([])
    setSelectedRun(null)
    setRunDetail(null)
    setActiveTab('results')
    loadResults(job.id)
  }

  function handleTabChange(tab) {
    setActiveTab(tab)
    if (tab === 'history' && selectedJob) {
      loadHistory(selectedJob.id)
    }
  }

  function selectRun(run) {
    setSelectedRun(run.id)
    loadRunDetail(selectedJob.id, run.id)
  }

  function openCreate() {
    setJobForm(EMPTY_JOB_FORM)
    setJobModal('create')
  }

  function openEdit(job) {
    setJobForm({
      name: job.name || '',
      subnet: job.subnet || '',
      schedule_cron: job.schedule_cron || '',
      is_active: job.is_active !== false,
      scan_type: job.scan_type || 'ping',
      notify_on_change: job.notify_on_change || false,
    })
    setJobModal({ edit: job })
  }

  async function handleJobSubmit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {
        name: jobForm.name,
        subnet: jobForm.subnet,
        schedule_cron: jobForm.schedule_cron || null,
        is_active: jobForm.is_active,
        scan_type: jobForm.scan_type,
        notify_on_change: jobForm.notify_on_change,
      }
      if (jobModal === 'create') {
        await api.post('/admin/scan-jobs', body)
      } else {
        await api.put(`/admin/scan-jobs/${jobModal.edit.id}`, body)
      }
      setJobModal(null)
      await loadJobs()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save scan job')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(jobId) {
    try {
      await api.delete(`/admin/scan-jobs/${jobId}`)
      setDeleteConfirm(null)
      if (selectedJob?.id === jobId) setSelectedJob(null)
      await loadJobs()
    } catch {
      setError('Failed to delete scan job')
    }
  }

  if (loading) return <PageSpinner message="Loading scan jobs..." />

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Discovery Scan Jobs</h1>
        <button
          onClick={openCreate}
          className="text-sm bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 transition"
        >
          + New Job
        </button>
      </div>

      <ErrorBanner error={error} />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Job list */}
        <div className="lg:col-span-1">
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <div className="px-4 py-3 border-b border-gray-100 bg-gray-50">
              <h2 className="font-semibold text-gray-700 text-sm">Jobs</h2>
            </div>
            {jobs.length === 0 ? (
              <div className="p-4 text-sm text-gray-500">No scan jobs configured.</div>
            ) : (
              <ul className="divide-y divide-gray-100 dark:divide-gray-700">
                {jobs.map((job) => (
                  <li key={job.id}>
                    <button
                      onClick={() => selectJob(job)}
                      className={`w-full text-left px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition ${
                        selectedJob?.id === job.id
                          ? 'bg-blue-50 dark:bg-blue-900/30 border-l-2 border-blue-500 dark:border-blue-400'
                          : ''
                      }`}
                    >
                      <p className="font-medium text-gray-900 dark:text-gray-100 text-sm truncate">{job.name}</p>
                      <p className="text-xs text-gray-500 mt-0.5">
                        {job.schedule_cron || 'Manual only'} &middot;{' '}
                        <span className={`font-medium ${job.is_active ? 'text-green-600' : 'text-gray-400'}`}>
                          {job.is_active ? 'Active' : 'Inactive'}
                        </span>
                        {job.scan_type && job.scan_type !== 'ping' && (
                          <span className="ml-1 text-blue-500">&middot; {SCAN_TYPE_LABELS[job.scan_type]}</span>
                        )}
                      </p>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Results / History panel */}
        <div className="lg:col-span-2">
          {!selectedJob ? (
            <div className="bg-white border border-gray-200 rounded-lg p-8 text-center text-gray-400 text-sm">
              Select a job to view results
            </div>
          ) : (
            <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
              <div className="px-4 py-3 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <h2 className="font-semibold text-gray-700 text-sm">{selectedJob.name}</h2>
                  <div className="flex border border-gray-200 rounded overflow-hidden text-xs">
                    <button
                      onClick={() => handleTabChange('results')}
                      className={`px-3 py-1 transition ${activeTab === 'results' ? 'bg-blue-600 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
                    >
                      Results
                    </button>
                    <button
                      onClick={() => handleTabChange('history')}
                      className={`px-3 py-1 transition ${activeTab === 'history' ? 'bg-blue-600 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
                    >
                      History
                    </button>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => openEdit(selectedJob)}
                    className="text-xs bg-gray-100 text-gray-700 px-3 py-1.5 rounded hover:bg-gray-200 transition"
                  >
                    Edit
                  </button>
                  {deleteConfirm === selectedJob.id ? (
                    <span className="flex items-center gap-1 text-xs">
                      <span className="text-red-600">Delete?</span>
                      <button onClick={() => handleDelete(selectedJob.id)} className="text-red-600 font-medium hover:text-red-800">Yes</button>
                      <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">No</button>
                    </span>
                  ) : (
                    <button
                      onClick={() => setDeleteConfirm(selectedJob.id)}
                      className="text-xs bg-gray-100 text-gray-700 px-3 py-1.5 rounded hover:bg-red-50 hover:text-red-700 transition"
                    >
                      Delete
                    </button>
                  )}
                  <button
                    onClick={() => runNow(selectedJob)}
                    disabled={running === selectedJob.id}
                    className="text-xs bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 disabled:bg-blue-400 transition"
                  >
                    {running === selectedJob.id ? 'Starting…' : 'Run Now'}
                  </button>
                </div>
              </div>

              {activeTab === 'results' && (
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead className="bg-gray-50 border-b border-gray-100">
                      <tr>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">IP Address</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Status</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">RTT (ms)</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">PTR Record</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Scanned At</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50">
                      {results.length === 0 ? (
                        <EmptyRow colSpan={5} message="No results yet." />
                      ) : (
                        results.map((r) => (
                          <tr key={r.id} className="hover:bg-gray-50">
                            <td className="px-4 py-2 font-mono text-xs text-gray-900">{r.ip_address}</td>
                            <td className="px-4 py-2">
                              <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${r.is_alive ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                                {r.is_alive ? 'Alive' : 'Down'}
                              </span>
                            </td>
                            <td className="px-4 py-2 text-gray-600 text-xs">{r.response_time_ms ?? '—'}</td>
                            <td className="px-4 py-2 font-mono text-xs">
                              {r.ptr_record ? (
                                <span className={r.fwd_rev_mismatch ? 'text-amber-600' : 'text-gray-700'}>
                                  {r.ptr_record}
                                  {r.fwd_rev_mismatch && (
                                    <span className="ml-1 text-amber-500" title="Forward/reverse DNS mismatch">⚠</span>
                                  )}
                                </span>
                              ) : (
                                <span className="text-gray-300">—</span>
                              )}
                            </td>
                            <td className="px-4 py-2 text-gray-500 text-xs">{formatDate(r.scanned_at)}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              )}

              {activeTab === 'history' && (
                <div>
                  {!runDetail ? (
                    <div className="overflow-x-auto">
                      <table className="min-w-full text-sm">
                        <thead className="bg-gray-50 border-b border-gray-100">
                          <tr>
                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Started</th>
                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Finished</th>
                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">New</th>
                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Gone</th>
                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Changed</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-50">
                          {history.length === 0 ? (
                            <EmptyRow colSpan={5} message="No scan history yet." />
                          ) : (
                            history.map((run) => (
                              <tr
                                key={run.id}
                                onClick={() => selectRun(run)}
                                className="hover:bg-blue-50 cursor-pointer"
                              >
                                <td className="px-4 py-2 text-xs text-gray-700">{formatDate(run.started_at)}</td>
                                <td className="px-4 py-2 text-xs text-gray-700">
                                  {run.finished_at ? formatDate(run.finished_at) : <span className="text-gray-400">—</span>}
                                </td>
                                <td className="px-4 py-2 text-xs">
                                  <span className="text-green-700 font-medium">+{run.new_count}</span>
                                </td>
                                <td className="px-4 py-2 text-xs">
                                  <span className="text-red-700 font-medium">-{run.gone_count}</span>
                                </td>
                                <td className="px-4 py-2 text-xs">
                                  <span className="text-yellow-700 font-medium">~{run.changed_count}</span>
                                </td>
                              </tr>
                            ))
                          )}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div>
                      <div className="px-4 py-2 border-b border-gray-100 bg-gray-50 flex items-center gap-3">
                        <button
                          onClick={() => { setRunDetail(null); setSelectedRun(null) }}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          ← Back to history
                        </button>
                        <span className="text-xs text-gray-500">
                          Run {formatDate(runDetail.started_at)}
                        </span>
                        <span className="text-xs text-green-700 font-medium">+{runDetail.new_count} new</span>
                        <span className="text-xs text-red-700 font-medium">-{runDetail.gone_count} gone</span>
                        <span className="text-xs text-yellow-700 font-medium">~{runDetail.changed_count} changed</span>
                      </div>
                      <div className="overflow-x-auto">
                        <table className="min-w-full text-sm">
                          <thead className="bg-gray-50 border-b border-gray-100">
                            <tr>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">IP Address</th>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Change</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-gray-50">
                            {(!runDetail.changes || runDetail.changes.length === 0) ? (
                              <tr>
                                <td colSpan={2} className="px-4 py-6 text-center text-gray-400">No changes in this run</td>
                              </tr>
                            ) : (
                              runDetail.changes.map((c, i) => (
                                <tr key={i} className={CHANGE_COLORS[c.change_type] || ''}>
                                  <td className="px-4 py-2 font-mono text-xs">{c.ip_address}</td>
                                  <td className="px-4 py-2 text-xs capitalize font-medium">{c.change_type}</td>
                                </tr>
                              ))
                            )}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {jobModal && (
        <Modal
          title={jobModal === 'create' ? 'New Scan Job' : `Edit: ${jobModal.edit?.name}`}
          onClose={() => setJobModal(null)}
        >
          <form onSubmit={handleJobSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Office subnet scan"
                value={jobForm.name}
                onChange={e => setJobForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Subnet / Target <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="192.168.1.0/24"
                value={jobForm.subnet}
                onChange={e => setJobForm(f => ({ ...f, subnet: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Scan Type</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={jobForm.scan_type}
                onChange={e => setJobForm(f => ({ ...f, scan_type: e.target.value }))}
              >
                <option value="ping">Ping</option>
                <option value="snmp">SNMP</option>
                <option value="ping+snmp">Ping + SNMP</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Schedule (cron)</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="0 * * * * (hourly)"
                value={jobForm.schedule_cron}
                onChange={e => setJobForm(f => ({ ...f, schedule_cron: e.target.value }))}
              />
            </div>
            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={jobForm.is_active}
                  onChange={e => setJobForm(f => ({ ...f, is_active: e.target.checked }))}
                  className="w-4 h-4"
                />
                <span className="text-sm text-gray-700">Active</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={jobForm.notify_on_change}
                  onChange={e => setJobForm(f => ({ ...f, notify_on_change: e.target.checked }))}
                  className="w-4 h-4"
                />
                <span className="text-sm text-gray-700">Notify on change</span>
              </label>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setJobModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Saving…' : jobModal === 'create' ? 'Create' : 'Save'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
