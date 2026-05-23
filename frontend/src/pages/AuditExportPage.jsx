import { useState } from 'react'
import { exportAuditLog } from '../api/client'
import ErrorBanner from '../components/ErrorBanner'

export default function AuditExportPage() {
  const [exporting, setExporting] = useState(false)
  const [error, setError] = useState(null)
  const [exportForm, setExportForm] = useState({ since: '', until: '', format: 'json' })

  async function handleExport(e) {
    e.preventDefault()
    setExporting(true)
    setError(null)
    try {
      const params = { format: exportForm.format }
      if (exportForm.since) params.since = new Date(exportForm.since).toISOString()
      if (exportForm.until) {
        const d = new Date(exportForm.until)
        d.setHours(23, 59, 59, 999)
        params.until = d.toISOString()
      }
      const res = await exportAuditLog(params)
      const url = URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = exportForm.format === 'csv' ? 'audit-export.csv' : 'audit-export.json'
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch {
      setError('Export failed. Please try again.')
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Export Audit Log</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Download audit log entries filtered by date range and format.
        </p>
      </div>

      <ErrorBanner error={error} onDismiss={() => setError(null)} />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 max-w-lg">
        <form onSubmit={handleExport} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Since (optional)
              </label>
              <input
                type="date"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={exportForm.since}
                onChange={e => setExportForm(f => ({ ...f, since: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Until (optional)
              </label>
              <input
                type="date"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                value={exportForm.until}
                onChange={e => setExportForm(f => ({ ...f, until: e.target.value }))}
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Format
            </label>
            <select
              className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
              value={exportForm.format}
              onChange={e => setExportForm(f => ({ ...f, format: e.target.value }))}
            >
              <option value="json">JSON</option>
              <option value="csv">CSV</option>
            </select>
          </div>
          <div className="pt-2">
            <button
              type="submit"
              disabled={exporting}
              className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
            >
              {exporting ? 'Exporting…' : 'Export'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
