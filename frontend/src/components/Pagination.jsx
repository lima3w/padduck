import { useTranslation } from 'react-i18next'

/**
 * Pagination component.
 * Props:
 *   page      {number}   — current page (1-based)
 *   limit     {number}   — items per page
 *   total     {number}   — total item count
 *   onChange  {Function} — called with new page number
 */
export default function Pagination({ page, limit, total, onChange }) {
  const { t } = useTranslation()
  const totalPages = Math.ceil(total / limit) || 1
  const from = total === 0 ? 0 : (page - 1) * limit + 1
  const to = Math.min(page * limit, total)

  // Build page numbers to show: always show first, last, and up to 2 around current
  const pages = []
  for (let p = 1; p <= totalPages; p++) {
    if (
      p === 1 ||
      p === totalPages ||
      (p >= page - 2 && p <= page + 2)
    ) {
      pages.push(p)
    }
  }

  // Insert ellipses where there are gaps
  const withEllipsis = []
  for (let i = 0; i < pages.length; i++) {
    if (i > 0 && pages[i] - pages[i - 1] > 1) {
      withEllipsis.push('...')
    }
    withEllipsis.push(pages[i])
  }

  function btn(label, targetPage, disabled, active = false) {
    return (
      <button
        key={`${label}-${targetPage}`}
        onClick={() => !disabled && onChange(targetPage)}
        disabled={disabled}
        className={`px-3 py-1.5 text-sm rounded border transition
          ${active
            ? 'bg-blue-600 text-white border-blue-600'
            : disabled
              ? 'text-gray-300 dark:text-gray-600 border-gray-200 dark:border-gray-700 cursor-not-allowed'
              : 'text-gray-600 dark:text-gray-300 border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700'
          }`}
      >
        {label}
      </button>
    )
  }

  return (
    <div className="flex items-center justify-between mt-4">
      <span className="text-sm text-gray-500 dark:text-gray-400">
        {total === 0
          ? t('pagination.noResults')
          : t('pagination.showingRange', { from, to, total })}
      </span>
      <div className="flex items-center gap-1">
        {btn(t('pagination.prev'), page - 1, page <= 1)}
        {withEllipsis.map((p, i) =>
          p === '...'
            ? <span key={`ellipsis-${i}`} className="px-2 text-gray-400">…</span>
            : btn(p, p, false, p === page)
        )}
        {btn(t('pagination.next'), page + 1, page >= totalPages)}
      </div>
    </div>
  )
}
