import { useState, useEffect } from 'react'
import { getIntegrationTemplates } from '../api/client'

export default function IntegrationTemplatesPage() {
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [expanded, setExpanded] = useState(null)

  useEffect(() => {
    getIntegrationTemplates()
      .then(res => setTemplates(res.data || []))
      .catch(err => setError(err.response?.data?.error || 'Failed to load templates'))
      .finally(() => setLoading(false))
  }, [])

  function toggle(id) {
    setExpanded(prev => prev === id ? null : id)
  }

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Integration Templates</h1>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Ready-to-use templates for connecting IPAM to common automation platforms and network tools.
      </p>

      {loading && <p className="text-sm text-gray-500 dark:text-gray-400">Loading…</p>}
      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      {!loading && !error && (
        <div className="space-y-3">
          {templates.length === 0 && (
            <p className="text-sm text-gray-400">No templates available.</p>
          )}
          {templates.map(t => (
            <div key={t.id} className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 overflow-hidden">
              <button
                onClick={() => toggle(t.id)}
                className="w-full flex items-center justify-between px-5 py-4 text-left hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <span className="font-semibold text-gray-900 dark:text-gray-100">{t.name}</span>
                  <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
                    {t.category}
                  </span>
                </div>
                <svg className={`w-4 h-4 text-gray-400 transition-transform ${expanded === t.id ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>

              {expanded === t.id && (
                <div className="border-t border-gray-100 dark:border-gray-800 px-5 py-4 space-y-4">
                  <p className="text-sm text-gray-600 dark:text-gray-400">{t.description}</p>

                  <div>
                    <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Setup Steps</h3>
                    <ol className="list-decimal list-inside space-y-1">
                      {(t.steps || []).map((step, i) => (
                        <li key={i} className="text-sm text-gray-700 dark:text-gray-300">{step}</li>
                      ))}
                    </ol>
                  </div>

                  {t.endpoints?.length > 0 && (
                    <div>
                      <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Relevant Endpoints</h3>
                      <ul className="space-y-1">
                        {t.endpoints.map((ep, i) => (
                          <li key={i} className="font-mono text-xs text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-gray-800 px-3 py-1 rounded">{ep}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
