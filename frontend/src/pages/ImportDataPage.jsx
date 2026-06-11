import { useState, useRef } from 'react'
import { api } from '../api/client'

const TABS = [
  { key: 'subnets', label: 'Subnets CSV' },
  { key: 'ips', label: 'IP Addresses CSV' },
  { key: 'phpipam', label: 'PHPIpam' },
]

const SUBNET_HEADERS = 'cidr,description,network,gateway,vlan,vrf,location'
const IP_HEADERS = 'address,hostname,status,subnet_cidr,assigned_to,mac_address'
const PHPIPAM_SUBNET_HEADERS = 'subnet,mask,description,sectionName'
const PHPIPAM_IP_HEADERS = 'ip,hostname,description,subnetIp,subnetMask,state'

function downloadTemplate(filename, content) {
  const blob = new Blob([content], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

function DropZone({ onFile, accept }) {
  const [dragging, setDragging] = useState(false)
  const inputRef = useRef(null)

  function handleDrop(e) {
    e.preventDefault()
    setDragging(false)
    const file = e.dataTransfer.files[0]
    if (file) onFile(file)
  }

  function handleChange(e) {
    const file = e.target.files[0]
    if (file) onFile(file)
  }

  return (
    <div
      className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
        dragging
          ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
          : 'border-gray-300 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500'
      }`}
      onClick={() => inputRef.current?.click()}
      onDragOver={e => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
    >
      <input
        ref={inputRef}
        type="file"
        accept={accept || '.csv'}
        className="hidden"
        onChange={handleChange}
        onClick={e => { e.stopPropagation() }}
      />
      <div className="text-gray-400 dark:text-gray-500 text-sm">
        <p className="font-medium text-gray-600 dark:text-gray-300">Drop a CSV file here, or click to browse</p>
        <p className="mt-1 text-xs">Accepts .csv files</p>
      </div>
    </div>
  )
}

function ResultPanel({ result }) {
  if (!result) return null
  const { total, imported, failed, errors } = result
  return (
    <div className="mt-4 space-y-3">
      <div className="flex gap-4 text-sm">
        <span className="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-700 dark:text-gray-200">
          Total: <strong>{total}</strong>
        </span>
        <span className="px-3 py-1.5 bg-green-100 dark:bg-green-900/30 rounded text-green-700 dark:text-green-300">
          Imported: <strong>{imported}</strong>
        </span>
        {failed > 0 && (
          <span className="px-3 py-1.5 bg-red-100 dark:bg-red-900/30 rounded text-red-700 dark:text-red-300">
            Failed: <strong>{failed}</strong>
          </span>
        )}
      </div>
      {Array.isArray(errors) && errors.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b dark:border-gray-700">
            <p className="text-sm font-medium text-red-700 dark:text-red-300">Import Errors</p>
          </div>
          <table className="w-full text-xs">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <th className="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">Row</th>
                <th className="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">Value</th>
                <th className="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">Error</th>
              </tr>
            </thead>
            <tbody>
              {errors.map((err, i) => (
                <tr key={i} className="border-b dark:border-gray-700 last:border-0">
                  <td className="px-4 py-2 text-gray-500 dark:text-gray-400">{err.row}</td>
                  <td className="px-4 py-2 font-mono text-gray-700 dark:text-gray-300">{err.value}</td>
                  <td className="px-4 py-2 text-red-600 dark:text-red-400">{err.message}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function DryRunPreviewPanel({ result }) {
  if (!result) return null
  const actionConfig = {
    create:  { label: 'Create',  cls: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' },
    skip:    { label: 'Skip',    cls: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300' },
    warning: { label: 'Warning', cls: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300' },
    error:   { label: 'Error',   cls: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300' },
  }
  return (
    <div className="mt-4 space-y-3">
      <div className="p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded text-sm text-blue-800 dark:text-blue-200 font-medium">
        Dry Run Preview — no changes were made
      </div>
      <div className="flex gap-3 text-sm flex-wrap">
        <span className="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-700 dark:text-gray-200">Total: <strong>{result.total}</strong></span>
        <span className="px-3 py-1.5 bg-green-100 dark:bg-green-900/30 rounded text-green-700 dark:text-green-300">Would Create: <strong>{result.creates}</strong></span>
        {result.skips > 0 && <span className="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-600 dark:text-gray-300">Skips: <strong>{result.skips}</strong></span>}
        {result.warnings > 0 && <span className="px-3 py-1.5 bg-yellow-100 dark:bg-yellow-900/30 rounded text-yellow-700 dark:text-yellow-300">Warnings: <strong>{result.warnings}</strong></span>}
        {result.errors > 0 && <span className="px-3 py-1.5 bg-red-100 dark:bg-red-900/30 rounded text-red-700 dark:text-red-300">Errors: <strong>{result.errors}</strong></span>}
      </div>
      {result.rows?.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <th className="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">Row</th>
                <th className="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">Action</th>
                <th className="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">Value</th>
                <th className="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">Reason</th>
              </tr>
            </thead>
            <tbody>
              {result.rows.map((r, i) => {
                const cfg = actionConfig[r.action] || actionConfig.error
                return (
                  <tr key={i} className="border-b dark:border-gray-700 last:border-0">
                    <td className="px-4 py-2 text-gray-500 dark:text-gray-400">{r.row}</td>
                    <td className="px-4 py-2">
                      <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${cfg.cls}`}>{cfg.label}</span>
                    </td>
                    <td className="px-4 py-2 font-mono text-gray-700 dark:text-gray-300">{r.value}</td>
                    <td className="px-4 py-2 text-gray-500 dark:text-gray-400">{r.reason || '—'}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          </div>
        </div>
      )}
    </div>
  )
}

function SubnetsTab() {
  const [file, setFile] = useState(null)
  const [uploading, setUploading] = useState(false)
  const [previewing, setPreviewing] = useState(false)
  const [result, setResult] = useState(null)
  const [dryRunResult, setDryRunResult] = useState(null)
  const [error, setError] = useState('')

  function handleFile(f) {
    setFile(f)
    setResult(null)
    setDryRunResult(null)
    setError('')
  }

  async function handlePreview() {
    if (!file) return
    setPreviewing(true)
    setResult(null)
    setDryRunResult(null)
    setError('')
    try {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await api.post('/admin/import/subnets?dry_run=true', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setDryRunResult(data)
    } catch (err) {
      setError(err.response?.data?.error || 'Preview failed')
    } finally {
      setPreviewing(false)
    }
  }

  async function handleUpload() {
    if (!file) return
    setUploading(true)
    setResult(null)
    setDryRunResult(null)
    setError('')
    try {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await api.post('/admin/import/subnets', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setResult(data)
    } catch (err) {
      setError(err.response?.data?.error || 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">Import Subnets from CSV</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            Expected headers: <code className="font-mono text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">{SUBNET_HEADERS}</code>
          </p>
        </div>
        <button
          onClick={() => downloadTemplate('subnets-template.csv', SUBNET_HEADERS + '\n')}
          className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300"
        >
          Download Template
        </button>
      </div>

      <DropZone onFile={handleFile} />

      {file && (
        <div className="flex items-center gap-3 flex-wrap">
          <span className="text-sm text-gray-600 dark:text-gray-400">
            Selected: <strong>{file.name}</strong> ({(file.size / 1024).toFixed(1)} KB)
          </span>
          <button
            onClick={handlePreview}
            disabled={previewing || uploading}
            className="px-4 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50"
          >
            {previewing ? 'Previewing...' : 'Preview (Dry Run)'}
          </button>
          <button
            onClick={handleUpload}
            disabled={uploading || previewing}
            className="px-4 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            {uploading ? 'Uploading...' : 'Upload & Import'}
          </button>
          <button
            onClick={() => { setFile(null); setResult(null); setDryRunResult(null); setError('') }}
            className="text-sm text-gray-400 hover:text-gray-600"
          >
            Clear
          </button>
        </div>
      )}

      {error && <p className="text-red-600 dark:text-red-400 text-sm">{error}</p>}
      <DryRunPreviewPanel result={dryRunResult} />
      <ResultPanel result={result} />
    </div>
  )
}

function IPsTab() {
  const [file, setFile] = useState(null)
  const [uploading, setUploading] = useState(false)
  const [previewing, setPreviewing] = useState(false)
  const [result, setResult] = useState(null)
  const [dryRunResult, setDryRunResult] = useState(null)
  const [error, setError] = useState('')

  function handleFile(f) {
    setFile(f)
    setResult(null)
    setDryRunResult(null)
    setError('')
  }

  async function handlePreview() {
    if (!file) return
    setPreviewing(true)
    setResult(null)
    setDryRunResult(null)
    setError('')
    try {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await api.post('/admin/import/ips?dry_run=true', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setDryRunResult(data)
    } catch (err) {
      setError(err.response?.data?.error || 'Preview failed')
    } finally {
      setPreviewing(false)
    }
  }

  async function handleUpload() {
    if (!file) return
    setUploading(true)
    setResult(null)
    setDryRunResult(null)
    setError('')
    try {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await api.post('/admin/import/ips', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setResult(data)
    } catch (err) {
      setError(err.response?.data?.error || 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">Import IP Addresses from CSV</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            Expected headers: <code className="font-mono text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">{IP_HEADERS}</code>
          </p>
        </div>
        <button
          onClick={() => downloadTemplate('ip-addresses-template.csv', IP_HEADERS + '\n')}
          className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300"
        >
          Download Template
        </button>
      </div>

      <DropZone onFile={handleFile} />

      {file && (
        <div className="flex items-center gap-3 flex-wrap">
          <span className="text-sm text-gray-600 dark:text-gray-400">
            Selected: <strong>{file.name}</strong> ({(file.size / 1024).toFixed(1)} KB)
          </span>
          <button
            onClick={handlePreview}
            disabled={previewing || uploading}
            className="px-4 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50"
          >
            {previewing ? 'Previewing...' : 'Preview (Dry Run)'}
          </button>
          <button
            onClick={handleUpload}
            disabled={uploading || previewing}
            className="px-4 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            {uploading ? 'Uploading...' : 'Upload & Import'}
          </button>
          <button
            onClick={() => { setFile(null); setResult(null); setDryRunResult(null); setError('') }}
            className="text-sm text-gray-400 hover:text-gray-600"
          >
            Clear
          </button>
        </div>
      )}

      {error && <p className="text-red-600 dark:text-red-400 text-sm">{error}</p>}
      <DryRunPreviewPanel result={dryRunResult} />
      <ResultPanel result={result} />
    </div>
  )
}

function PHPIpamTab() {
  const [kind, setKind] = useState('subnets')
  const [file, setFile] = useState(null)
  const [uploading, setUploading] = useState(false)
  const [previewing, setPreviewing] = useState(false)
  const [result, setResult] = useState(null)
  const [dryRunResult, setDryRunResult] = useState(null)
  const [error, setError] = useState('')

  function handleFile(f) {
    setFile(f)
    setResult(null)
    setDryRunResult(null)
    setError('')
  }

  function handleKindChange(newKind) {
    setKind(newKind)
    setFile(null)
    setResult(null)
    setDryRunResult(null)
    setError('')
  }

  async function handlePreview() {
    if (!file) return
    setPreviewing(true)
    setResult(null)
    setDryRunResult(null)
    setError('')
    try {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await api.post(`/admin/import/phpipam?kind=${kind}&dry_run=true`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setDryRunResult(data)
    } catch (err) {
      setError(err.response?.data?.error || 'Preview failed')
    } finally {
      setPreviewing(false)
    }
  }

  async function handleUpload() {
    if (!file) return
    setUploading(true)
    setResult(null)
    setDryRunResult(null)
    setError('')
    try {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await api.post(`/admin/import/phpipam?kind=${kind}`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setResult(data)
    } catch (err) {
      setError(err.response?.data?.error || 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  const templateHeaders = kind === 'subnets' ? PHPIPAM_SUBNET_HEADERS : PHPIPAM_IP_HEADERS
  const templateFilename = kind === 'subnets' ? 'phpipam-subnets-template.csv' : 'phpipam-ips-template.csv'

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">Import from PHPIpam</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            Import a CSV export from a PHPIpam instance
          </p>
        </div>
        <button
          onClick={() => downloadTemplate(templateFilename, templateHeaders + '\n')}
          className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300"
        >
          Download Template
        </button>
      </div>

      <div className="flex gap-4">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="radio"
            name="phpipam-kind"
            value="subnets"
            checked={kind === 'subnets'}
            onChange={() => handleKindChange('subnets')}
            className="accent-blue-600"
          />
          <span className="text-sm text-gray-700 dark:text-gray-300">Subnets</span>
        </label>
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="radio"
            name="phpipam-kind"
            value="ips"
            checked={kind === 'ips'}
            onChange={() => handleKindChange('ips')}
            className="accent-blue-600"
          />
          <span className="text-sm text-gray-700 dark:text-gray-300">IP Addresses</span>
        </label>
      </div>

      <div>
        <p className="text-xs text-gray-500 dark:text-gray-400 mb-2">
          Expected headers: <code className="font-mono bg-gray-100 dark:bg-gray-700 px-1 rounded">{templateHeaders}</code>
        </p>
        <DropZone onFile={handleFile} />
      </div>

      {file && (
        <div className="flex items-center gap-3 flex-wrap">
          <span className="text-sm text-gray-600 dark:text-gray-400">
            Selected: <strong>{file.name}</strong> ({(file.size / 1024).toFixed(1)} KB)
          </span>
          <button
            onClick={handlePreview}
            disabled={previewing || uploading}
            className="px-4 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50"
          >
            {previewing ? 'Previewing...' : 'Preview (Dry Run)'}
          </button>
          <button
            onClick={handleUpload}
            disabled={uploading || previewing}
            className="px-4 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            {uploading ? 'Uploading...' : 'Upload & Import'}
          </button>
          <button
            onClick={() => { setFile(null); setResult(null); setDryRunResult(null); setError('') }}
            className="text-sm text-gray-400 hover:text-gray-600"
          >
            Clear
          </button>
        </div>
      )}

      {error && <p className="text-red-600 dark:text-red-400 text-sm">{error}</p>}
      <DryRunPreviewPanel result={dryRunResult} />
      <ResultPanel result={result} />
    </div>
  )
}

export default function ImportDataPage() {
  const [activeTab, setActiveTab] = useState('subnets')

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Data Import</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Import subnets and IP addresses from CSV files or PHPIpam exports.
        </p>
      </div>

      <div className="flex border-b dark:border-gray-700 mb-6">
        {TABS.map(tab => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`px-5 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px ${
              activeTab === tab.key
                ? 'border-blue-600 text-blue-600 dark:text-blue-400 dark:border-blue-400'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        {activeTab === 'subnets' && <SubnetsTab />}
        {activeTab === 'ips' && <IPsTab />}
        {activeTab === 'phpipam' && <PHPIpamTab />}
      </div>
    </div>
  )
}
