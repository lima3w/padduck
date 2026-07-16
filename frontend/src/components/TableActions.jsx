import { useTranslation } from 'react-i18next'

export default function TableActions({
  onEdit,
  onDelete,
  confirming = false,
  onRequestDelete,
  onCancelDelete,
  editLabel,
  deleteLabel,
  confirmLabel,
  className = '',
}) {
  const { t } = useTranslation()
  const resolvedEditLabel = editLabel ?? t('common.edit')
  const resolvedDeleteLabel = deleteLabel ?? t('common.delete')
  const resolvedConfirmLabel = confirmLabel ?? t('subnets.confirmDelete')
  return (
    <div className={`inline-flex items-center justify-end gap-2 text-xs whitespace-nowrap ${className}`}>
      {onEdit && (
        <button
          type="button"
          onClick={onEdit}
          className="text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded-sm"
        >
          {resolvedEditLabel}
        </button>
      )}
      {onDelete && confirming ? (
        <>
          <span className="text-red-600 dark:text-red-400">{resolvedConfirmLabel}</span>
          <button
            type="button"
            onClick={onDelete}
            className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 font-medium focus:outline-none focus:ring-2 focus:ring-red-500 rounded-sm"
          >
            {t('common.yes')}
          </button>
          <button
            type="button"
            onClick={onCancelDelete}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-400 rounded-sm"
          >
            {t('common.no')}
          </button>
        </>
      ) : (
        onRequestDelete && (
          <button
            type="button"
            onClick={onRequestDelete}
            className="text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400 focus:outline-none focus:ring-2 focus:ring-red-500 rounded-sm"
          >
            {resolvedDeleteLabel}
          </button>
        )
      )}
    </div>
  )
}
