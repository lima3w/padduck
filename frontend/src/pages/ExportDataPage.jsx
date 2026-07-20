import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { downloadFile } from '../utils/download'

export default function ExportDataPage() {
  const { t } = useTranslation()
  const [downloading, setDownloading] = useState(null) // 'csv' | 'json' | 'v2' | null
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
      setError(err.message || t('exportData.exportFailed'))
    } finally {
      setDownloading(null)
    }
  }

  async function handleMigrationBundleExport() {
    setError('')
    setDownloading('v2')
    try {
      await downloadFile('/api/v1/admin/export/v2-migration-bundle', 'ipam-v2-migration-bundle.zip')
    } catch (err) {
      setError(err.message || t('exportData.exportFailed'))
    } finally {
      setDownloading(null)
    }
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('exportData.title')}</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          {t('exportData.subtitle')}
        </p>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-700 dark:text-red-300 text-sm">
          {error}
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-base font-semibold text-gray-800 dark:text-gray-100 mb-1">{t('exportData.fullExportTitle')}</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
          {t('exportData.fullExportSubtitle')}
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
                {t('exportData.generating')}
              </>
            ) : (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                {t('exportData.exportCsv')}
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
                {t('exportData.generating')}
              </>
            ) : (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                {t('exportData.exportJson')}
              </>
            )}
          </button>
        </div>
      </div>

      <div className="mt-6 bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-base font-semibold text-gray-800 dark:text-gray-100 mb-1">{t('exportData.migrationBundleTitle')}</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
          {t('exportData.migrationBundleSubtitle')}
        </p>

        <button
          onClick={handleMigrationBundleExport}
          disabled={downloading !== null}
          className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
        >
          {downloading === 'v2' ? (
            <>
              <span className="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              {t('exportData.generating')}
            </>
          ) : (
            <>
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              {t('exportData.exportMigrationBundle')}
            </>
          )}
        </button>
      </div>
    </div>
  )
}
