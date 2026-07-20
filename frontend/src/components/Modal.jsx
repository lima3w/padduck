import { useEffect, useId, useRef } from 'react'
import { useTranslation } from 'react-i18next'

export default function Modal({ title, onClose, children }) {
  const { t } = useTranslation()
  const titleId = useId()
  const dialogRef = useRef(null)
  const previousFocusRef = useRef(null)

  // Focus the dialog once on mount and restore the previous element on unmount.
  // This runs only on mount (empty deps) so that button clicks inside the modal
  // don't re-trigger it and steal focus away from text inputs.
  useEffect(() => {
    previousFocusRef.current = document.activeElement
    dialogRef.current?.focus()
    return () => {
      previousFocusRef.current?.focus?.()
    }
  }, [])

  // Escape-to-close is a separate effect so it can track the latest onClose
  // without affecting the focus-on-mount behaviour above.
  useEffect(() => {
    function handleKeyDown(event) {
      if (event.key === 'Escape') {
        event.preventDefault()
        onClose()
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      onMouseDown={(event) => { if (event.target === event.currentTarget) onClose() }}
    >
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md max-h-[90vh] overflow-hidden focus:outline-none"
      >
        <div className="flex items-center justify-between gap-4 px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 id={titleId} className="text-lg font-semibold text-gray-800 dark:text-gray-100">{title}</h3>
          <button
            type="button"
            onClick={onClose}
            aria-label={t('modal.closeDialog')}
            className="rounded text-gray-500 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-gray-300 dark:hover:text-gray-100 text-xl leading-none"
          >
            &times;
          </button>
        </div>
        <div className="px-6 py-4 overflow-y-auto max-h-[calc(90vh-4.5rem)]">{children}</div>
      </div>
    </div>
  )
}
