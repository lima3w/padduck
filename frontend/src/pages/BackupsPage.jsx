import { useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'
import { downloadFile } from '../utils/download'

// ── helpers ──────────────────────────────────────────────────────────────────

function Network({ title, description, children }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 space-y-4">
      <div>
        <h2 className="text-base font-semibold text-gray-800 dark:text-gray-100">{title}</h2>
        {description && (
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{description}</p>
        )}
      </div>
      {children}
    </div>
  )
}

function DropZone({ onFile, accept = '.zip', label }) {
  const { t } = useTranslation()
  const [dragging, setDragging] = useState(false)
  const inputRef = useRef(null)

  function handleDrop(e) {
    e.preventDefault()
    setDragging(false)
    const file = e.dataTransfer.files[0]
    if (file) onFile(file)
  }

  return (
    <div
      className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
        dragging ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-gray-300 dark:border-gray-600 hover:border-blue-400'
      }`}
      onClick={() => inputRef.current?.click()}
      onDragOver={e => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
    >
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={e => { const f = e.target.files[0]; if (f) onFile(f) }}
        onClick={e => e.stopPropagation()}
      />
      <p className="text-sm font-medium text-gray-600 dark:text-gray-300">{label ?? t('backups.dropZoneDefaultLabel')}</p>
      <p className="text-xs text-gray-400 mt-1">{t('backups.acceptsFiles', { accept })}</p>
    </div>
  )
}

// ── CSV import sub-tab state ──────────────────────────────────────────────────

const CSV_TAB_KEYS = [
  { key: 'subnets', labelKey: 'csvTabSubnets' },
  { key: 'ips', labelKey: 'csvTabIps' },
  { key: 'phpipam', labelKey: 'csvTabPhpipam' },
]

function CSVImportTab({ tab, setError }) {
  const { t } = useTranslation()
  const [file, setFile] = useState(null)
  const [result, setResult] = useState(null)
  const [uploading, setUploading] = useState(false)

  async function handleImport() {
    if (!file) return
    setUploading(true)
    setResult(null)
    try {
      const form = new FormData()
      form.append('file', file)
      let url = '/admin/import/'
      if (tab === 'subnets') url += 'subnets'
      else if (tab === 'ips') url += 'ips'
      else url += 'phpipam'
      const { data } = await api.post(url, form, { headers: { 'Content-Type': 'multipart/form-data' } })
      setResult(data)
    } catch (err) {
      setError(err.response?.data?.error || t('backups.importFailed'))
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="space-y-3">
      <DropZone accept=".csv" label={t('backups.csvDropLabel')} onFile={setFile} />
      {file && (
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-600 dark:text-gray-300">{file.name}</span>
          <button
            onClick={handleImport}
            disabled={uploading}
            className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            {uploading ? t('backups.importing') : t('backups.import')}
          </button>
          <button onClick={() => { setFile(null); setResult(null) }} className="text-xs text-gray-400 hover:text-gray-600">
            {t('backups.clear')}
          </button>
        </div>
      )}
      {result && (
        <div className="text-sm space-y-1">
          <div className="flex gap-4">
            <span className="px-3 py-1 bg-gray-100 dark:bg-gray-700 rounded text-gray-700 dark:text-gray-200">
              {t('backups.totalLabel')} <strong>{result.total ?? result.imported ?? 0}</strong>
            </span>
            <span className="px-3 py-1 bg-green-100 dark:bg-green-900/30 rounded text-green-700 dark:text-green-300">
              {t('backups.importedLabel')} <strong>{result.imported ?? 0}</strong>
            </span>
            {result.failed > 0 && (
              <span className="px-3 py-1 bg-red-100 dark:bg-red-900/30 rounded text-red-700 dark:text-red-300">
                {t('backups.failedLabel')} <strong>{result.failed}</strong>
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

// ── Main component ────────────────────────────────────────────────────────────

export default function BackupsPage() {
  const { t } = useTranslation()
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  // Complete backup
  const [downloadingBackup, setDownloadingBackup] = useState(false)

  // Restore
  const [restoreFile, setRestoreFile] = useState(null)
  const [restoring, setRestoring] = useState(false)
  const [restoreConfirm, setRestoreConfirm] = useState(false)

  // Data export
  const [downloading, setDownloading] = useState(null)

  // Data import
  const [importTab, setImportTab] = useState('subnets')

  function showMsg(text) {
    setMessage(text)
    setTimeout(() => setMessage(''), 5000)
  }

  async function handleDownloadBackup() {
    setDownloadingBackup(true)
    setError('')
    try {
      await downloadFile('/api/v1/admin/backups/download', `padduck-backup-${new Date().toISOString().slice(0, 10)}.zip`)
      showMsg(t('backups.backupDownloaded'))
    } catch (err) {
      setError(err.message || t('backups.downloadFailed'))
    } finally {
      setDownloadingBackup(false)
    }
  }

  async function handleRestore() {
    if (!restoreFile) return
    setRestoring(true)
    setError('')
    try {
      const form = new FormData()
      form.append('file', restoreFile)
      const { data } = await api.post('/admin/backups/restore', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      showMsg(data.message || t('backups.restoredSuccessfully'))
      setRestoreFile(null)
      setRestoreConfirm(false)
    } catch (err) {
      setError(err.response?.data?.error || t('backups.restoreFailed'))
    } finally {
      setRestoring(false)
    }
  }

  async function handleDataExport(format) {
    setDownloading(format)
    setError('')
    try {
      await downloadFile(
        `/api/v1/admin/export/full?format=${format}`,
        `padduck_export.${format}`
      )
    } catch (err) {
      setError(err.message || t('exportData.exportFailed'))
    } finally {
      setDownloading(null)
    }
  }

  async function handleMigrationBundle() {
    setDownloading('v2')
    setError('')
    try {
      await downloadFile('/api/v1/admin/export/v2-migration-bundle', 'padduck_v2_migration_bundle.zip')
    } catch (err) {
      setError(err.message || t('exportData.exportFailed'))
    } finally {
      setDownloading(null)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t('backups.title')}</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          {t('backups.subtitle')}
        </p>
      </div>

      {error && (
        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-700 dark:text-red-300 text-sm">
          {error}
        </div>
      )}
      {message && (
        <div className="p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded text-green-700 dark:text-green-300 text-sm">
          {message}
        </div>
      )}

      {/* ── Complete Backup ─────────────────────────────────────────────────── */}
      <Network
        title={t('backups.completeBackupTitle')}
        description={t('backups.completeBackupDescription')}
      >
        <button
          onClick={handleDownloadBackup}
          disabled={downloadingBackup}
          className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
        >
          {downloadingBackup ? (
            <><span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin inline-block" /> {t('backups.generating')}</>
          ) : (
            <><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>{t('backups.downloadCompleteBackup')}</>
          )}
        </button>
        <p className="text-xs text-gray-500 dark:text-gray-400">
          {t('backups.backupArchiveIncludesPrefix')}<code>./data/</code>{t('backups.backupArchiveIncludesSuffix')}
        </p>
      </Network>

      {/* ── Restore ─────────────────────────────────────────────────────────── */}
      <Network
        title={t('backups.restoreTitle')}
        description={t('backups.restoreDescription')}
      >
        {!restoreFile ? (
          <DropZone accept=".zip" label={t('backups.restoreDropLabel')} onFile={f => { setRestoreFile(f); setRestoreConfirm(false) }} />
        ) : (
          <div className="space-y-3">
            <div className="flex items-center gap-3 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded">
              <svg className="w-5 h-5 text-yellow-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
              <div>
                <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">{t('backups.readyToRestorePrefix')}<strong>{restoreFile.name}</strong></p>
                <p className="text-xs text-yellow-700 dark:text-yellow-300 mt-0.5">{t('backups.restoreWarning')}</p>
              </div>
            </div>
            {!restoreConfirm ? (
              <div className="flex gap-3">
                <button
                  onClick={() => setRestoreConfirm(true)}
                  className="px-4 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700"
                >
                  {t('backups.restoreFromThisBackup')}
                </button>
                <button
                  onClick={() => setRestoreFile(null)}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded text-sm hover:bg-gray-200"
                >
                  {t('common.cancel')}
                </button>
              </div>
            ) : (
              <div className="space-y-2">
                <p className="text-sm font-semibold text-red-700 dark:text-red-400">{t('backups.areYouSure')}</p>
                <div className="flex gap-3">
                  <button
                    onClick={handleRestore}
                    disabled={restoring}
                    className="px-4 py-2 bg-red-700 text-white rounded text-sm hover:bg-red-800 disabled:opacity-50"
                  >
                    {restoring ? t('backups.restoring') : t('backups.yesRestoreNow')}
                  </button>
                  <button
                    onClick={() => { setRestoreFile(null); setRestoreConfirm(false) }}
                    className="px-4 py-2 bg-gray-100 text-gray-700 rounded text-sm hover:bg-gray-200"
                  >
                    {t('common.cancel')}
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </Network>

      {/* ── Data Export ─────────────────────────────────────────────────────── */}
      <Network
        title={t('exportData.title')}
        description={t('backups.dataExportDescription')}
      >
        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => handleDataExport('csv')}
            disabled={downloading !== null}
            className="px-4 py-2 bg-gray-700 dark:bg-gray-600 text-white rounded text-sm hover:bg-gray-800 disabled:opacity-50"
          >
            {downloading === 'csv' ? t('backups.generating') : t('exportData.exportCsv')}
          </button>
          <button
            onClick={() => handleDataExport('json')}
            disabled={downloading !== null}
            className="px-4 py-2 bg-gray-700 dark:bg-gray-600 text-white rounded text-sm hover:bg-gray-800 disabled:opacity-50"
          >
            {downloading === 'json' ? t('backups.generating') : t('exportData.exportJson')}
          </button>
          <button
            onClick={handleMigrationBundle}
            disabled={downloading !== null}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 rounded text-sm hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
          >
            {downloading === 'v2' ? t('backups.generating') : t('backups.v2MigrationBundleZip')}
          </button>
        </div>
      </Network>

      {/* ── Data Import ─────────────────────────────────────────────────────── */}
      <Network
        title={t('backups.dataImportTitle')}
        description={t('backups.dataImportDescription')}
      >
        <div className="flex gap-2 mb-4">
          {CSV_TAB_KEYS.map(tabInfo => (
            <button
              key={tabInfo.key}
              onClick={() => setImportTab(tabInfo.key)}
              className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                importTab === tabInfo.key
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600'
              }`}
            >
              {t(`backups.${tabInfo.labelKey}`)}
            </button>
          ))}
        </div>
        <CSVImportTab key={importTab} tab={importTab} setError={setError} />
      </Network>
    </div>
  )
}
