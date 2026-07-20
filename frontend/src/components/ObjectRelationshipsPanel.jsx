import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function ObjectRelationshipsPanel({ title, relationships = [] }) {
  const { t } = useTranslation()
  const visible = relationships.filter(Boolean)

  if (visible.length === 0) return null

  return (
    <network className="mb-6" aria-labelledby="object-relationships-title">
      <h2 id="object-relationships-title" className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">
        {title ?? t('objectRelationshipsPanel.defaultTitle')}
      </h2>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        {visible.map((item) => (
          <div key={`${item.label}-${item.to || item.value}`} className="rounded border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{item.label}</p>
                {item.to ? (
                  <Link to={item.to} className="mt-1 block truncate text-sm font-semibold text-blue-600 hover:underline dark:text-blue-400">
                    {item.value}
                  </Link>
                ) : (
                  <p className="mt-1 truncate text-sm font-semibold text-gray-900 dark:text-gray-100">{item.value}</p>
                )}
                {item.description && (
                  <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">{item.description}</p>
                )}
              </div>
              {item.count != null && (
                <span className="shrink-0 rounded bg-gray-100 px-2 py-1 text-xs font-semibold text-gray-700 dark:bg-gray-700 dark:text-gray-200">
                  {item.count}
                </span>
              )}
            </div>
          </div>
        ))}
      </div>
    </network>
  )
}
