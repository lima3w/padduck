// Shows a compact data completeness indicator.
// fields: array of { label: string, ok: boolean }
export default function DataQualityBadge({ fields, _entityLabel = 'record' }) {
  const filled = fields.filter(f => f.ok).length
  const total = fields.length
  const pct = total === 0 ? 0 : Math.round((filled / total) * 100)
  const colour = pct >= 80 ? 'text-green-600 dark:text-green-400' : pct >= 50 ? 'text-yellow-600 dark:text-yellow-400' : 'text-red-600 dark:text-red-400'
  const barColour = pct >= 80 ? 'bg-green-500' : pct >= 50 ? 'bg-yellow-500' : 'bg-red-500'
  const missing = fields.filter(f => !f.ok).map(f => f.label)

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">Data Quality</h3>
        <span className={`text-sm font-bold ${colour}`}>{pct}%</span>
      </div>
      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 mb-2">
        <div className={`${barColour} h-2 rounded-full transition-all`} style={{ width: `${pct}%` }} />
      </div>
      <p className="text-xs text-gray-500 dark:text-gray-400">{filled}/{total} fields complete</p>
      {missing.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {missing.map(m => (
            <span key={m} className="inline-block px-1.5 py-0.5 rounded text-xs bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">
              {m}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
