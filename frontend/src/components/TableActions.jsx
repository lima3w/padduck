export default function TableActions({
  onEdit,
  onDelete,
  confirming = false,
  onRequestDelete,
  onCancelDelete,
  editLabel = 'Edit',
  deleteLabel = 'Delete',
  confirmLabel = 'Confirm?',
  className = '',
}) {
  return (
    <div className={`inline-flex items-center justify-end gap-2 text-xs whitespace-nowrap ${className}`}>
      {onEdit && (
        <button
          type="button"
          onClick={onEdit}
          className="text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded-sm"
        >
          {editLabel}
        </button>
      )}
      {onDelete && confirming ? (
        <>
          <span className="text-red-600 dark:text-red-400">{confirmLabel}</span>
          <button
            type="button"
            onClick={onDelete}
            className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 font-medium focus:outline-none focus:ring-2 focus:ring-red-500 rounded-sm"
          >
            Yes
          </button>
          <button
            type="button"
            onClick={onCancelDelete}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-400 rounded-sm"
          >
            No
          </button>
        </>
      ) : (
        onRequestDelete && (
          <button
            type="button"
            onClick={onRequestDelete}
            className="text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400 focus:outline-none focus:ring-2 focus:ring-red-500 rounded-sm"
          >
            {deleteLabel}
          </button>
        )
      )}
    </div>
  )
}
