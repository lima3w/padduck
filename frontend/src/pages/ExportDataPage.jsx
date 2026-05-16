import { useState } from 'react'
import { downloadFile } from '../utils/download'

export default function ExportDataPage() {
  const [downloading, setDownloading] = useState(null) // 'csv' | 'json' | null
  const [error, setError] = useState('')

  async function handleExport(format) {
    setError('')
    setDownloading(format)
    try {
      await downloadFile(
        `/api/v1/admin/export/full?format=${format}`,
        `ipam-export.${format}`
      )
    } catch (err) {
      setError(err.message || 'Export failed')
    } finally {
      setDownloading(null)
    }
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Data Export</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Download a full export of all IPAM data in your preferred format.
        </p>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-700 dark:text-red-300 text-sm">
          {error}
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-base font-semibold text-gray-800 dark:text-gray-100 mb-1">Full Data Export</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
          Exports all sections, subnets, and IP addresses. Large datasets may take a moment to generate.
        </p>

        <div className="flex flex-col sm:flex-row gap-3">
          <button
            onClick={() => handleExport('csv')}
            disabled={downloading !== null}
            className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
          >
            {downloading === 'csv' ? (
              <>
                <span className="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                Export All Data (CSV)
              </>
            )}
          </button>

          <button
            onClick={() => handleExport('json')}
            disabled={downloading !== null}
            className="flex items-center gap-2 px-5 py-2.5 bg-gray-700 dark:bg-gray-600 text-white rounded hover:bg-gray-800 dark:hover:bg-gray-500 disabled:opacity-50 text-sm font-medium"
          >
            {downloading === 'json' ? (
              <>
                <span className="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                Export All Data (JSON)
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
