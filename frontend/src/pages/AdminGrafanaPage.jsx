import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { generateTokenForMe } from '../api/auth'

const METRIC_NAMES = ['subnet_utilization', 'ip_by_status', 'section_summary']

export default function AdminGrafanaPage() {
  const { t } = useTranslation()
  const METRICS = METRIC_NAMES.map(name => ({ name, desc: t(`adminGrafana.metrics.${name}`) }))
  const [token, setToken] = useState('')
  const [tokenName, setTokenName] = useState('grafana-datasource')
  const [generating, setGenerating] = useState(false)
  const [error, setError] = useState('')

  const baseUrl = window.location.origin + '/api/grafana'

  async function handleGenerate(e) {
    e.preventDefault()
    setGenerating(true)
    setError('')
    try {
      const res = await generateTokenForMe(tokenName)
      setToken(res.data?.token || res.data?.rawToken || '')
    } catch {
      setError(t('adminGrafana.generateFailed'))
    } finally {
      setGenerating(false)
    }
  }

  return (
    <div className="p-6 max-w-3xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-1">{t('adminGrafana.title')}</h1>
        <p className="text-sm text-gray-500">
          {t('adminGrafana.subtitle')}
        </p>
      </div>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">{t('adminGrafana.step1Title')}</h2>
        <div className="bg-gray-50 border border-gray-200 rounded px-4 py-3 font-mono text-sm text-gray-800 select-all">
          {baseUrl}
        </div>
        <p className="text-xs text-gray-500">
          {t('adminGrafana.step1HintPrefix')}<strong>{t('adminGrafana.jsonApiLabel')}</strong>{t('adminGrafana.step1HintMiddle')}<strong>{t('adminGrafana.simpleJsonLabel')}</strong>{t('adminGrafana.step1HintSuffix')}
        </p>
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">{t('adminGrafana.step2Title')}</h2>
        <form onSubmit={handleGenerate} className="flex gap-2 items-end">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-600 mb-1">{t('adminGrafana.tokenNameLabel')}</label>
            <input
              type="text"
              value={tokenName}
              onChange={e => setTokenName(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <button
            type="submit"
            disabled={generating}
            className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {generating ? t('adminGrafana.generating') : t('adminGrafana.generate')}
          </button>
        </form>
        {error && <p className="text-red-600 text-sm">{error}</p>}
        {token && (
          <div>
            <p className="text-xs text-amber-600 font-medium mb-1">{t('adminGrafana.copyTokenNotice')}</p>
            <div className="bg-yellow-50 border border-yellow-200 rounded px-4 py-3 font-mono text-sm break-all select-all">
              {token}
            </div>
            <p className="text-xs text-gray-500 mt-1">
              {t('adminGrafana.bearerHintPrefix')}<strong>{t('adminGrafana.bearerLabel')}</strong>{t('adminGrafana.bearerHintSuffix')}{' '}
              <code className="bg-gray-100 px-1 rounded">Authorization: Bearer {'<token>'}</code>
            </p>
          </div>
        )}
      </network>

      <network className="space-y-3">
        <h2 className="text-base font-semibold text-gray-800">{t('adminGrafana.step3Title')}</h2>
        <div className="rounded border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('adminGrafana.metricColumn')}</th>
                <th className="px-4 py-3 text-left font-medium text-gray-600">{t('common.description')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {METRICS.map(m => (
                <tr key={m.name}>
                  <td className="px-4 py-3 font-mono text-gray-900">{m.name}</td>
                  <td className="px-4 py-3 text-gray-600">{m.desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </network>

      <network className="space-y-2">
        <h2 className="text-base font-semibold text-gray-800">{t('adminGrafana.step4Title')}</h2>
        <ol className="list-decimal list-inside space-y-1 text-sm text-gray-700">
          <li>{t('adminGrafana.step4Item1Prefix')}<strong>{t('adminGrafana.marcussonPlugin')}</strong>{t('adminGrafana.step4Item1Middle')}<strong>{t('adminGrafana.simpleJsonLabel')}</strong>{t('adminGrafana.step4Item1Suffix')}</li>
          <li>{t('adminGrafana.step4Item2')}</li>
          <li>{t('adminGrafana.step4Item3Prefix')}<em>{t('adminGrafana.customHttpHeaders')}</em>{t('adminGrafana.step4Item3Middle')}<code className="bg-gray-100 px-1 rounded">Authorization</code>{t('adminGrafana.step4Item3ThenValue')}<code className="bg-gray-100 px-1 rounded">Bearer &lt;your-token&gt;</code>{t('adminGrafana.step4Item3End')}</li>
          <li>{t('adminGrafana.step4Item4Prefix')}<code className="bg-gray-100 px-1 rounded">ok</code>{t('adminGrafana.step4Item4Suffix')}</li>
          <li>{t('adminGrafana.step4Item5')}</li>
        </ol>
      </network>
    </div>
  )
}
