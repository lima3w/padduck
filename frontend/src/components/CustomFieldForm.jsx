export default function CustomFieldForm({ definitions, values, onChange, readOnly }) {
  if (!definitions || definitions.length === 0) return null

  const today = new Date().toISOString().split('T')[0]

  return (
    <div className="space-y-4">
      {definitions.map(def => {
        const val = values?.[def.name] ?? def.default_value ?? ''
        const isDatePast = def.field_type === 'date' && readOnly && val && val < today

        return (
          <div key={def.id}>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {def.label}
              {def.is_required && <span className="text-red-500 ml-1">*</span>}
            </label>

            {readOnly ? (
              <div className={`text-sm ${isDatePast ? 'text-red-600 dark:text-red-400 font-medium' : 'text-gray-700 dark:text-gray-300'}`}>
                {def.field_type === 'url' && val ? (
                  <a href={val} target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:underline break-all">
                    {val}
                  </a>
                ) : def.field_type === 'checkbox' ? (
                  val === 'true' || val === true ? 'Yes' : 'No'
                ) : (
                  val || <span className="text-gray-400">—</span>
                )}
              </div>
            ) : (
              <>
                {def.field_type === 'text' && (
                  <input
                    type="text"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder={def.placeholder || ''}
                    value={val}
                    required={def.is_required}
                    onChange={e => onChange(def.name, e.target.value)}
                  />
                )}
                {def.field_type === 'number' && (
                  <input
                    type="number"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder={def.placeholder || ''}
                    value={val}
                    required={def.is_required}
                    onChange={e => onChange(def.name, e.target.value)}
                  />
                )}
                {def.field_type === 'textarea' && (
                  <textarea
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder={def.placeholder || ''}
                    rows={3}
                    value={val}
                    required={def.is_required}
                    onChange={e => onChange(def.name, e.target.value)}
                  />
                )}
                {def.field_type === 'dropdown' && (
                  <select
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    value={val}
                    required={def.is_required}
                    onChange={e => onChange(def.name, e.target.value)}
                  >
                    <option value="">Select...</option>
                    {(def.options || []).map(opt => (
                      <option key={opt.value} value={opt.value}>{opt.label || opt.value}</option>
                    ))}
                  </select>
                )}
                {def.field_type === 'checkbox' && (
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={val === 'true' || val === true}
                      className="w-4 h-4 text-blue-600 rounded"
                      onChange={e => onChange(def.name, e.target.checked ? 'true' : 'false')}
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">{def.placeholder || def.label}</span>
                  </label>
                )}
                {def.field_type === 'date' && (
                  <input
                    type="date"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    value={val}
                    required={def.is_required}
                    onChange={e => onChange(def.name, e.target.value)}
                  />
                )}
                {def.field_type === 'url' && (
                  <input
                    type="url"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder={def.placeholder || 'https://'}
                    value={val}
                    required={def.is_required}
                    onChange={e => onChange(def.name, e.target.value)}
                  />
                )}
                {def.field_type === 'email' && (
                  <input
                    type="email"
                    className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                    placeholder={def.placeholder || ''}
                    value={val}
                    required={def.is_required}
                    onChange={e => onChange(def.name, e.target.value)}
                  />
                )}
              </>
            )}
          </div>
        )
      })}
    </div>
  )
}
