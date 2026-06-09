import { useState, useEffect, useRef } from 'react'
import { api } from '../api/client'
import Modal from '../components/Modal'
import PageSpinner from '../components/PageSpinner'
import ErrorBanner from '../components/ErrorBanner'
import EmptyRow from '../components/EmptyRow'

const SCAN_TYPE_LABELS = { ping: 'Ping', snmp: 'SNMP', 'ping+snmp': 'Ping + SNMP' }

const CRON_PRESETS = [
  { label: 'Manual only', value: '' },
  { label: 'Every 15 minutes', value: '*/15 * * * *' },
  { label: 'Every hour', value: '0 * * * *' },
  { label: 'Every 6 hours', value: '0 */6 * * *' },
  { label: 'Daily at midnight', value: '0 0 * * *' },
  { label: 'Weekly (Sunday)', value: '0 0 * * 0' },
  { label: 'Custom', value: '__custom__' },
]

function formatDate(val) {
  if (!val) return '—'
  const d = new Date(val)
  return isNaN(d.getTime()) ? '—' : d.toLocaleString()
}

function cronPresetValue(cron) {
  if (!cron) return ''
  const match = CRON_PRESETS.find(p => p.value === cron && p.value !== '__custom__')
  return match ? cron : '__custom__'
}

const CHANGE_COLORS = {
  new: 'bg-green-50 text-green-800',
  gone: 'bg-red-50 text-red-800',
  changed: 'bg-yellow-50 text-yellow-800',
}

const EMPTY_CREATE_FORM = {
  name: '', subnet: '', schedule_cron: '', is_active: true,
  scan_type: 'ping', ping_concurrency: 20, notify_on_change: false,
  auto_add_ips: true, discover_dns: true, dns_overwrite: false,
}

function jobToSettingsForm(job) {
  return {
    name: job.name || '',
    schedule_cron: job.scheduleCron || '',
    is_active: job.isActive !== false,
    scan_type: job.scanType || 'ping',
    ping_concurrency: job.pingConcurrency || 20,
    notify_on_change: job.notifyOnChange || false,
    auto_add_ips: job.autoAddIps !== false,
    discover_dns: job.discoverDns !== false,
    dns_overwrite: job.dnsOverwrite || false,
  }
}

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
  // Settings panel state
  const [settingsForm, setSettingsForm] = useState(null)
  const [settingsDirty, setSettingsDirty] = useState(false)
  const [settingsSaving, setSettingsSaving] = useState(false)
  const [settingsError, setSettingsError] = useState('')
  // Create modal state
  const [createModal, setCreateModal] = useState(false)
  const [createForm, setCreateForm] = useState(EMPTY_CREATE_FORM)
  const [creating, setCreating] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [hideDown, setHideDown] = useState(false)
  const [goneIPs, setGoneIPs] = useState(new Set())
  const pollRef = useRef(null)

  useEffect(() => {
    loadJobs()
    return () => { if (pollRef.current) clearTimeout(pollRef.current) }
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
      const [resultsResp, historyResp] = await Promise.all([
        api.get(`/admin/scan-jobs/${jobId}/results?limit=200`),
        api.get(`/admin/scan-jobs/${jobId}/history`),
      ])
      setResults(resultsResp.data || [])
      const runs = historyResp.data || []
      if (runs.length > 0) {
        try {
          const { data: detail } = await api.get(`/admin/scan-jobs/${jobId}/history/${runs[0].id}`)
          const gone = new Set(
            (detail.changes || []).filter(c => c.changeType === 'gone').map(c => c.ipAddress)
          )
          setGoneIPs(gone)
        } catch {
          setGoneIPs(new Set())
        }
      } else {
        setGoneIPs(new Set())
      }
    } catch {
      setResults([])
      setGoneIPs(new Set())
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
      setRunDetail({ ...data.run, changes: data.changes })
    } catch {
      setRunDetail(null)
    }
  }

  function stopPolling() {
    if (pollRef.current) {
      clearTimeout(pollRef.current)
      pollRef.current = null
    }
  }

  async function pollJobStatus(jobId, activeTabSnapshot) {
    try {
      const { data } = await api.get(`/admin/scan-jobs/${jobId}/status`)
      if (data.running) {
        pollRef.current = setTimeout(() => pollJobStatus(jobId, activeTabSnapshot), 2000)
      } else {
        setRunning(null)
        if (activeTabSnapshot === 'results') loadResults(jobId)
        else if (activeTabSnapshot === 'history') loadHistory(jobId)
      }
    } catch {
      setRunning(null)
    }
  }

  async function runNow(job) {
    stopPolling()
    setRunning(job.id)
    try {
      await api.post(`/admin/scan-jobs/${job.id}/run`)
      pollJobStatus(job.id, activeTab)
    } catch {
      setRunning(null)
    }
  }

  function selectJob(job) {
    stopPolling()
    setRunning(null)
    setSelectedJob(job)
    setResults([])
    setHistory([])
    setSelectedRun(null)
    setRunDetail(null)
    setSettingsForm(jobToSettingsForm(job))
    setSettingsDirty(false)
    setSettingsError('')
    setGoneIPs(new Set())
    setActiveTab('results')
    loadResults(job.id)
  }

  function handleTabChange(tab) {
    setActiveTab(tab)
    if (tab === 'history' && selectedJob) loadHistory(selectedJob.id)
  }

  function selectRun(run) {
    setSelectedRun(run.id)
    loadRunDetail(selectedJob.id, run.id)
  }

  function updateSettings(patch) {
    setSettingsForm(f => ({ ...f, ...patch }))
    setSettingsDirty(true)
  }

  async function saveSettings(e) {
    e.preventDefault()
    setSettingsSaving(true)
    setSettingsError('')
    try {
      const { data } = await api.put(`/admin/scan-jobs/${selectedJob.id}`, {
        name: settingsForm.name,
        subnet_ids: selectedJob.subnetIds,
        schedule_cron: settingsForm.schedule_cron || null,
        is_active: settingsForm.is_active,
        scan_type: settingsForm.scan_type,
        ping_concurrency: settingsForm.ping_concurrency,
        notify_on_change: settingsForm.notify_on_change,
        auto_add_ips: settingsForm.auto_add_ips,
        discover_dns: settingsForm.discover_dns,
        dns_overwrite: settingsForm.dns_overwrite,
      })
      setSelectedJob(data)
      setSettingsForm(jobToSettingsForm(data))
      setSettingsDirty(false)
      await loadJobs()
    } catch (err) {
      setSettingsError(err.response?.data?.error || 'Failed to save settings')
    } finally {
      setSettingsSaving(false)
    }
  }

  async function handleCreate(e) {
    e.preventDefault()
    setCreating(true)
    try {
      await api.post('/admin/scan-jobs', {
        name: createForm.name,
        subnet: createForm.subnet,
        schedule_cron: createForm.schedule_cron || null,
        is_active: createForm.is_active,
        scan_type: createForm.scan_type,
        ping_concurrency: createForm.ping_concurrency,
        notify_on_change: createForm.notify_on_change,
        auto_add_ips: createForm.auto_add_ips,
        discover_dns: createForm.discover_dns,
        dns_overwrite: createForm.dns_overwrite,
      })
      setCreateModal(false)
      setCreateForm(EMPTY_CREATE_FORM)
      await loadJobs()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create scan job')
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(jobId) {
    try {
      await api.delete(`/admin/scan-jobs/${jobId}`)
      setDeleteConfirm(null)
      if (selectedJob?.id === jobId) {
        setSelectedJob(null)
        setSettingsForm(null)
      }
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
          onClick={() => { setCreateForm(EMPTY_CREATE_FORM); setCreateModal(true) }}
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
              <ul className="divide-y divide-gray-100">
                {jobs.map((job) => (
                  <li key={job.id}>
                    <button
                      onClick={() => selectJob(job)}
                      className={`w-full text-left px-4 py-3 hover:bg-gray-50 transition ${
                        selectedJob?.id === job.id
                          ? 'bg-blue-50 dark:bg-blue-900/20 border-l-2 border-blue-500'
                          : ''
                      }`}
                    >
                      <p className="font-medium text-gray-900 text-sm truncate">{job.name}</p>
                      <p className="text-xs text-gray-500 mt-0.5">
                        <span className={`font-medium ${job.isActive ? 'text-green-600' : 'text-gray-400'}`}>
                          {job.isActive ? 'Active' : 'Inactive'}
                        </span>
                        {job.scheduleCron
                          ? <span className="ml-1">&middot; {job.scheduleCron}</span>
                          : <span className="ml-1 text-gray-400">&middot; Manual only</span>}
                        {job.nextRunAt && (
                          <span className="ml-1 text-blue-500">&middot; next {formatDate(job.nextRunAt)}</span>
                        )}
                      </p>
                      <p className="text-xs text-gray-400 mt-0.5">
                        {SCAN_TYPE_LABELS[job.scanType] || 'Ping'}
                        {job.pingConcurrency && job.pingConcurrency !== 20 && (
                          <span> &middot; {job.pingConcurrency} workers</span>
                        )}
                      </p>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Detail panel */}
        <div className="lg:col-span-2">
          {!selectedJob ? (
            <div className="bg-white border border-gray-200 rounded-lg p-8 text-center text-gray-400 text-sm">
              Select a job to view details
            </div>
          ) : (
            <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
              {/* Header */}
              <div className="px-4 py-3 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <h2 className="font-semibold text-gray-700 text-sm">{selectedJob.name}</h2>
                  <div className="flex border border-gray-200 rounded overflow-hidden text-xs">
                    {['results', 'history', 'settings'].map(tab => (
                      <button
                        key={tab}
                        onClick={() => handleTabChange(tab)}
                        className={`px-3 py-1 capitalize transition ${activeTab === tab ? 'bg-blue-600 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
                      >
                        {tab}
                      </button>
                    ))}
                  </div>
                </div>
                <div className="flex gap-2">
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
                    {running === selectedJob.id ? 'Running…' : 'Run Now'}
                  </button>
                </div>
              </div>

              {/* Results tab */}
              {activeTab === 'results' && (
                <div>
                  <div className="px-4 py-2 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                    <span className="text-xs text-gray-500">
                      {hideDown
                        ? `${results.filter(r => r.isAlive).length} alive`
                        : `${results.length} total · ${results.filter(r => r.isAlive).length} alive · ${results.filter(r => !r.isAlive).length} down`}
                    </span>
                    <label className="flex items-center gap-2 cursor-pointer select-none">
                      <input
                        type="checkbox"
                        checked={hideDown}
                        onChange={e => setHideDown(e.target.checked)}
                        className="w-4 h-4"
                      />
                      <span className="text-xs text-gray-600">Hide down</span>
                    </label>
                  </div>
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
                        results.filter(r => !hideDown || r.isAlive).map((r) => (
                          <tr key={r.id} className="hover:bg-gray-50">
                            <td className="px-4 py-2 font-mono text-xs text-gray-900">{r.ipAddress}</td>
                            <td className="px-4 py-2">
                              {r.isAlive ? (
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-700">Alive</span>
                              ) : goneIPs.has(r.ipAddress) ? (
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-700" title="Was alive in the previous scan">Gone</span>
                              ) : (
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-500">Down</span>
                              )}
                            </td>
                            <td className="px-4 py-2 text-gray-600 text-xs">{r.responseTimeMs ?? '—'}</td>
                            <td className="px-4 py-2 font-mono text-xs">
                              {r.ptrRecord ? (
                                <span className={r.fwdRevMismatch ? 'text-amber-600' : 'text-gray-700'}>
                                  {r.ptrRecord}
                                  {r.fwdRevMismatch && (
                                    <span className="ml-1 text-amber-500" title="Forward/reverse DNS mismatch">⚠</span>
                                  )}
                                </span>
                              ) : (
                                <span className="text-gray-300">—</span>
                              )}
                            </td>
                            <td className="px-4 py-2 text-gray-500 text-xs">{formatDate(r.scannedAt)}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                  </div>
                </div>
              )}

              {/* History tab */}
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
                              <tr key={run.id} onClick={() => selectRun(run)} className="hover:bg-blue-50 dark:hover:bg-blue-900/20 cursor-pointer">
                                <td className="px-4 py-2 text-xs text-gray-700">{formatDate(run.startedAt)}</td>
                                <td className="px-4 py-2 text-xs text-gray-700">
                                  {run.finishedAt ? formatDate(run.finishedAt) : <span className="text-gray-400">—</span>}
                                </td>
                                <td className="px-4 py-2 text-xs"><span className="text-green-700 font-medium">+{run.newCount}</span></td>
                                <td className="px-4 py-2 text-xs"><span className="text-red-700 font-medium">-{run.goneCount}</span></td>
                                <td className="px-4 py-2 text-xs"><span className="text-yellow-700 font-medium">~{run.changedCount}</span></td>
                              </tr>
                            ))
                          )}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div>
                      <div className="px-4 py-2 border-b border-gray-100 bg-gray-50 flex items-center gap-3">
                        <button onClick={() => { setRunDetail(null); setSelectedRun(null) }} className="text-xs text-blue-600 hover:underline">
                          ← Back to history
                        </button>
                        <span className="text-xs text-gray-500">Run {formatDate(runDetail.startedAt)}</span>
                        <span className="text-xs text-green-700 font-medium">+{runDetail.newCount} new</span>
                        <span className="text-xs text-red-700 font-medium">-{runDetail.goneCount} gone</span>
                        <span className="text-xs text-yellow-700 font-medium">~{runDetail.changedCount} changed</span>
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
                              <tr><td colSpan={2} className="px-4 py-6 text-center text-gray-400">No changes in this run</td></tr>
                            ) : (
                              runDetail.changes.map((c, i) => (
                                <tr key={i} className={CHANGE_COLORS[c.changeType] || ''}>
                                  <td className="px-4 py-2 font-mono text-xs">{c.ipAddress}</td>
                                  <td className="px-4 py-2 text-xs capitalize font-medium">{c.changeType}</td>
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

              {/* Settings tab */}
              {activeTab === 'settings' && settingsForm && (
                <form onSubmit={saveSettings} className="p-6 space-y-6">
                  {settingsError && (
                    <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{settingsError}</div>
                  )}

                  {/* Basic */}
                  <div>
                    <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">General</h3>
                    <div className="space-y-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                        <input
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                          value={settingsForm.name}
                          onChange={e => updateSettings({ name: e.target.value })}
                          required
                        />
                      </div>
                      <div className="grid grid-cols-2 gap-y-2 gap-x-6">
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={settingsForm.is_active}
                            onChange={e => updateSettings({ is_active: e.target.checked })}
                            className="w-4 h-4"
                          />
                          <span className="text-sm text-gray-700">Active</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={settingsForm.notify_on_change}
                            onChange={e => updateSettings({ notify_on_change: e.target.checked })}
                            className="w-4 h-4"
                          />
                          <span className="text-sm text-gray-700">Notify on change</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={settingsForm.auto_add_ips !== false}
                            onChange={e => updateSettings({ auto_add_ips: e.target.checked })}
                            className="w-4 h-4"
                          />
                          <span className="text-sm text-gray-700">Auto-add active IPs to subnet</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={settingsForm.discover_dns !== false}
                            onChange={e => updateSettings({ discover_dns: e.target.checked })}
                            className="w-4 h-4"
                          />
                          <span className="text-sm text-gray-700">Discover reverse DNS name</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer col-span-2">
                          <input
                            type="checkbox"
                            checked={settingsForm.dns_overwrite || false}
                            onChange={e => updateSettings({ dns_overwrite: e.target.checked })}
                            className="w-4 h-4"
                            disabled={!settingsForm.discover_dns}
                          />
                          <span className={`text-sm ${settingsForm.discover_dns ? 'text-gray-700' : 'text-gray-400'}`}>
                            Overwrite existing DNS name when resolved
                          </span>
                        </label>
                      </div>
                    </div>
                  </div>

                  {/* Schedule */}
                  <div>
                    <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">Schedule</h3>
                    <div className="space-y-3">
                      <div className="flex flex-wrap gap-2">
                        {CRON_PRESETS.map(p => {
                          const isCustom = p.value === '__custom__'
                          const active = isCustom
                            ? cronPresetValue(settingsForm.schedule_cron) === '__custom__'
                            : settingsForm.schedule_cron === p.value
                          return (
                            <button
                              key={p.value}
                              type="button"
                              onClick={() => {
                                if (isCustom) {
                                  if (cronPresetValue(settingsForm.schedule_cron) !== '__custom__') {
                                    updateSettings({ schedule_cron: '' })
                                  }
                                } else {
                                  updateSettings({ schedule_cron: p.value })
                                }
                              }}
                              className={`px-3 py-1.5 text-xs rounded border transition ${
                                active
                                  ? 'bg-blue-600 text-white border-blue-600'
                                  : 'bg-white text-gray-600 border-gray-200 hover:border-blue-400'
                              }`}
                            >
                              {p.label}
                            </button>
                          )
                        })}
                      </div>
                      {cronPresetValue(settingsForm.schedule_cron) === '__custom__' && (
                        <div>
                          <input
                            className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                            placeholder="* * * * * (minute hour day month weekday)"
                            value={settingsForm.schedule_cron}
                            onChange={e => updateSettings({ schedule_cron: e.target.value })}
                          />
                          <p className="text-xs text-gray-400 mt-1">Standard 5-field cron expression</p>
                        </div>
                      )}
                      {selectedJob.nextRunAt && (
                        <p className="text-xs text-gray-500">Next scheduled run: <span className="font-medium">{formatDate(selectedJob.nextRunAt)}</span></p>
                      )}
                      {selectedJob.lastRunAt && (
                        <p className="text-xs text-gray-500">Last run: <span className="font-medium">{formatDate(selectedJob.lastRunAt)}</span></p>
                      )}
                    </div>
                  </div>

                  {/* Scan settings */}
                  <div>
                    <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">Scan Settings</h3>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Scan Type</label>
                        <select
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                          value={settingsForm.scan_type}
                          onChange={e => updateSettings({ scan_type: e.target.value })}
                        >
                          <option value="ping">Ping</option>
                          <option value="snmp">SNMP</option>
                          <option value="ping+snmp">Ping + SNMP</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Concurrency</label>
                        <input
                          type="number"
                          min={1}
                          max={100}
                          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                          value={settingsForm.ping_concurrency}
                          onChange={e => updateSettings({ ping_concurrency: parseInt(e.target.value, 10) || 20 })}
                        />
                        <p className="text-xs text-gray-400 mt-1">Parallel ping workers (1–100)</p>
                      </div>
                    </div>
                  </div>

                  <div className="flex justify-end gap-2 pt-2 border-t border-gray-100">
                    {settingsDirty && (
                      <button
                        type="button"
                        onClick={() => { setSettingsForm(jobToSettingsForm(selectedJob)); setSettingsDirty(false); setSettingsError('') }}
                        className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
                      >
                        Discard
                      </button>
                    )}
                    <button
                      type="submit"
                      disabled={settingsSaving || !settingsDirty}
                      className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
                    >
                      {settingsSaving ? 'Saving…' : 'Save Changes'}
                    </button>
                  </div>
                </form>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Create modal */}
      {createModal && (
        <Modal title="New Scan Job" onClose={() => setCreateModal(false)}>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Office subnet scan"
                value={createForm.name}
                onChange={e => setCreateForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Subnet / Target <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="192.168.1.0/24"
                value={createForm.subnet}
                onChange={e => setCreateForm(f => ({ ...f, subnet: e.target.value }))}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Scan Type</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={createForm.scan_type}
                onChange={e => setCreateForm(f => ({ ...f, scan_type: e.target.value }))}
              >
                <option value="ping">Ping</option>
                <option value="snmp">SNMP</option>
                <option value="ping+snmp">Ping + SNMP</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Schedule (cron)</label>
              <div className="flex flex-wrap gap-2 mb-2">
                {CRON_PRESETS.filter(p => p.value !== '__custom__').map(p => (
                  <button
                    key={p.value}
                    type="button"
                    onClick={() => setCreateForm(f => ({ ...f, schedule_cron: p.value }))}
                    className={`px-2 py-1 text-xs rounded border transition ${
                      createForm.schedule_cron === p.value
                        ? 'bg-blue-600 text-white border-blue-600'
                        : 'bg-white text-gray-600 border-gray-200 hover:border-blue-400'
                    }`}
                  >
                    {p.label}
                  </button>
                ))}
              </div>
              <input
                className="w-full border rounded px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Custom cron expression"
                value={createForm.schedule_cron}
                onChange={e => setCreateForm(f => ({ ...f, schedule_cron: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-y-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={createForm.is_active} onChange={e => setCreateForm(f => ({ ...f, is_active: e.target.checked }))} className="w-4 h-4" />
                <span className="text-sm text-gray-700">Active</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={createForm.notify_on_change} onChange={e => setCreateForm(f => ({ ...f, notify_on_change: e.target.checked }))} className="w-4 h-4" />
                <span className="text-sm text-gray-700">Notify on change</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={createForm.auto_add_ips !== false} onChange={e => setCreateForm(f => ({ ...f, auto_add_ips: e.target.checked }))} className="w-4 h-4" />
                <span className="text-sm text-gray-700">Auto-add active IPs</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={createForm.discover_dns !== false} onChange={e => setCreateForm(f => ({ ...f, discover_dns: e.target.checked }))} className="w-4 h-4" />
                <span className="text-sm text-gray-700">Discover reverse DNS</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer col-span-2">
                <input type="checkbox" checked={createForm.dns_overwrite || false} onChange={e => setCreateForm(f => ({ ...f, dns_overwrite: e.target.checked }))} className="w-4 h-4" disabled={!createForm.discover_dns} />
                <span className={`text-sm ${createForm.discover_dns ? 'text-gray-700' : 'text-gray-400'}`}>Overwrite existing DNS name when resolved</span>
              </label>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setCreateModal(false)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={creating} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {creating ? 'Creating…' : 'Create'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
