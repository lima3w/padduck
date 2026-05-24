import { useState, useRef } from 'react'
import { api } from '../api/client'
import { downloadFile } from '../utils/download'

// ── helpers ──────────────────────────────────────────────────────────────────

function Section({ title, description, children }) {
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

function DropZone({ onFile, accept = '.zip', label = 'Drop a file here, or click to browse' }) {
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
      <p className="text-sm font-medium text-gray-600 dark:text-gray-300">{label}</p>
      <p className="text-xs text-gray-400 mt-1">Accepts {accept} files</p>
    </div>
  )
}

// ── CSV import sub-tab state ──────────────────────────────────────────────────

const CSV_TABS = [
  { key: 'subnets', label: 'Subnets' },
  { key: 'ips', label: 'IP Addresses' },
  { key: 'phpipam', label: 'PHPIpam' },
]

function CSVImportTab({ tab, setError }) {
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
      setError(err.response?.data?.error || 'Import failed')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="space-y-3">
      <DropZone accept=".csv" label="Drop a CSV file here, or click to browse" onFile={setFile} />
      {file && (
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-600 dark:text-gray-300">{file.name}</span>
          <button
            onClick={handleImport}
            disabled={uploading}
            className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            {uploading ? 'Importing…' : 'Import'}
          </button>
          <button onClick={() => { setFile(null); setResult(null) }} className="text-xs text-gray-400 hover:text-gray-600">
            Clear
          </button>
        </div>
      )}
      {result && (
        <div className="text-sm space-y-1">
          <div className="flex gap-4">
            <span className="px-3 py-1 bg-gray-100 dark:bg-gray-700 rounded text-gray-700 dark:text-gray-200">
              Total: <strong>{result.total ?? result.imported ?? 0}</strong>
            </span>
            <span className="px-3 py-1 bg-green-100 dark:bg-green-900/30 rounded text-green-700 dark:text-green-300">
              Imported: <strong>{result.imported ?? 0}</strong>
            </span>
            {result.failed > 0 && (
              <span className="px-3 py-1 bg-red-100 dark:bg-red-900/30 rounded text-red-700 dark:text-red-300">
                Failed: <strong>{result.failed}</strong>
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
      showMsg('Backup downloaded successfully.')
    } catch (err) {
      setError(err.message || 'Download failed')
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
      showMsg(data.message || 'Backup restored successfully.')
      setRestoreFile(null)
      setRestoreConfirm(false)
    } catch (err) {
      setError(err.response?.data?.error || 'Restore failed')
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
        `ipam-export.${format}`
      )
    } catch (err) {
      setError(err.message || 'Export failed')
    } finally {
      setDownloading(null)
    }
  }

  async function handleMigrationBundle() {
    setDownloading('v2')
    setError('')
    try {
      await downloadFile('/api/v1/admin/export/v2-migration-bundle', 'ipam-v2-migration-bundle.zip')
    } catch (err) {
      setError(err.message || 'Export failed')
    } finally {
      setDownloading(null)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Backups</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Download a complete system backup or restore from a previous backup. You can also import and export data in CSV or JSON format.
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
      <Section
        title="Complete Backup"
        description="Downloads a ZIP archive containing the full database, all configuration settings, and any uploaded files. Use this to migrate to a new server or restore after data loss."
      >
        <button
          onClick={handleDownloadBackup}
          disabled={downloadingBackup}
          className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
        >
          {downloadingBackup ? (
            <><span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin inline-block" /> Generating&hellip;</>
          ) : (
            <><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>Download Complete Backup</>
          )}
        </button>
        <p className="text-xs text-gray-500 dark:text-gray-400">
          The backup archive includes: full PostgreSQL database dump, all admin settings, and files from the <code>./data/</code> directory.
        </p>
      </Section>

      {/* ── Restore ─────────────────────────────────────────────────────────── */}
      <Section
        title="Restore from Backup"
        description="Upload a backup ZIP archive to restore your data. This will overwrite all current data — use with caution."
      >
        {!restoreFile ? (
          <DropZone accept=".zip" label="Drop a backup .zip file here, or click to browse" onFile={f => { setRestoreFile(f); setRestoreConfirm(false) }} />
        ) : (
          <div className="space-y-3">
            <div className="flex items-center gap-3 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded">
              <svg className="w-5 h-5 text-yellow-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
              <div>
                <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">Ready to restore: <strong>{restoreFile.name}</strong></p>
                <p className="text-xs text-yellow-700 dark:text-yellow-300 mt-0.5">This will overwrite ALL current data and configuration. This action cannot be undone.</p>
              </div>
            </div>
            {!restoreConfirm ? (
              <div className="flex gap-3">
                <button
                  onClick={() => setRestoreConfirm(true)}
                  className="px-4 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700"
                >
                  Restore from This Backup
                </button>
                <button
                  onClick={() => setRestoreFile(null)}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded text-sm hover:bg-gray-200"
                >
                  Cancel
                </button>
              </div>
            ) : (
              <div className="space-y-2">
                <p className="text-sm font-semibold text-red-700 dark:text-red-400">Are you sure? This cannot be undone.</p>
                <div className="flex gap-3">
                  <button
                    onClick={handleRestore}
                    disabled={restoring}
                    className="px-4 py-2 bg-red-700 text-white rounded text-sm hover:bg-red-800 disabled:opacity-50"
                  >
                    {restoring ? 'Restoring…' : 'Yes, Restore Now'}
                  </button>
                  <button
                    onClick={() => { setRestoreFile(null); setRestoreConfirm(false) }}
                    className="px-4 py-2 bg-gray-100 text-gray-700 rounded text-sm hover:bg-gray-200"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </Section>

      {/* ── Data Export ─────────────────────────────────────────────────────── */}
      <Section
        title="Data Export"
        description="Export all sections, subnets, and IP addresses in a portable format."
      >
        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => handleDataExport('csv')}
            disabled={downloading !== null}
            className="px-4 py-2 bg-gray-700 dark:bg-gray-600 text-white rounded text-sm hover:bg-gray-800 disabled:opacity-50"
          >
            {downloading === 'csv' ? 'Generating…' : 'Export All Data (CSV)'}
          </button>
          <button
            onClick={() => handleDataExport('json')}
            disabled={downloading !== null}
            className="px-4 py-2 bg-gray-700 dark:bg-gray-600 text-white rounded text-sm hover:bg-gray-800 disabled:opacity-50"
          >
            {downloading === 'json' ? 'Generating…' : 'Export All Data (JSON)'}
          </button>
          <button
            onClick={handleMigrationBundle}
            disabled={downloading !== null}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 rounded text-sm hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
          >
            {downloading === 'v2' ? 'Generating…' : 'v2 Migration Bundle (ZIP)'}
          </button>
        </div>
      </Section>

      {/* ── Data Import ─────────────────────────────────────────────────────── */}
      <Section
        title="Data Import"
        description="Import subnets or IP addresses from a CSV file."
      >
        <div className="flex gap-2 mb-4">
          {CSV_TABS.map(t => (
            <button
              key={t.key}
              onClick={() => setImportTab(t.key)}
              className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                importTab === t.key
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
        <CSVImportTab key={importTab} tab={importTab} setError={setError} />
      </Section>
    </div>
  )
}
