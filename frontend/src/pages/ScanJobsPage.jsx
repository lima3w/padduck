import { useState, useEffect } from 'react'
import { api } from '../api/client'

export default function ScanJobsPage() {
  const [jobs, setJobs] = useState([])
  const [selectedJob, setSelectedJob] = useState(null)
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(true)
  const [running, setRunning] = useState(null)
  const [error, setError] = useState('')

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
    loadResults(job.id)
  }

  if (loading) return <div className="p-6 text-gray-500">Loading…</div>

  return (
    <div className="max-w-6xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Discovery Scan Jobs</h1>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>
      )}

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
                        selectedJob?.id === job.id ? 'bg-blue-50 border-l-2 border-blue-500' : ''
                      }`}
                    >
                      <p className="font-medium text-gray-900 text-sm truncate">{job.name}</p>
                      <p className="text-xs text-gray-500 mt-0.5">
                        {job.schedule_cron || 'Manual only'} &middot;{' '}
                        <span
                          className={`font-medium ${job.is_active ? 'text-green-600' : 'text-gray-400'}`}
                        >
                          {job.is_active ? 'Active' : 'Inactive'}
                        </span>
                      </p>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Results */}
        <div className="lg:col-span-2">
          {!selectedJob ? (
            <div className="bg-white border border-gray-200 rounded-lg p-8 text-center text-gray-400 text-sm">
              Select a job to view results
            </div>
          ) : (
            <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
              <div className="px-4 py-3 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                <h2 className="font-semibold text-gray-700 text-sm">{selectedJob.name} — Recent Results</h2>
                <button
                  onClick={() => runNow(selectedJob)}
                  disabled={running === selectedJob.id}
                  className="text-xs bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 disabled:bg-blue-400 transition"
                >
                  {running === selectedJob.id ? 'Starting…' : 'Run Now'}
                </button>
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
                      <tr>
                        <td colSpan={5} className="px-4 py-6 text-center text-gray-400">
                          No results yet
                        </td>
                      </tr>
                    ) : (
                      results.map((r) => (
                        <tr key={r.id} className="hover:bg-gray-50">
                          <td className="px-4 py-2 font-mono text-xs text-gray-900">{r.ip_address}</td>
                          <td className="px-4 py-2">
                            <span
                              className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                                r.is_alive
                                  ? 'bg-green-100 text-green-700'
                                  : 'bg-gray-100 text-gray-500'
                              }`}
                            >
                              {r.is_alive ? 'Alive' : 'Down'}
                            </span>
                          </td>
                          <td className="px-4 py-2 text-gray-600 text-xs">
                            {r.response_time_ms ?? '—'}
                          </td>
                          <td className="px-4 py-2 font-mono text-xs">
                            {r.ptr_record ? (
                              <span className={r.fwd_rev_mismatch ? 'text-amber-600' : 'text-gray-700'}>
                                {r.ptr_record}
                                {r.fwd_rev_mismatch && (
                                  <span
                                    className="ml-1 text-amber-500"
                                    title="Forward/reverse DNS mismatch"
                                  >
                                    ⚠
                                  </span>
                                )}
                              </span>
                            ) : (
                              <span className="text-gray-300">—</span>
                            )}
                          </td>
                          <td className="px-4 py-2 text-gray-500 text-xs">
                            {new Date(r.scanned_at).toLocaleString()}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
