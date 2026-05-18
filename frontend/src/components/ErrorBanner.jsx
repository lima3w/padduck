export default function ErrorBanner({ error, onDismiss, className = 'mb-4' }) {
  if (!error) return null
  return (
    <div className={`flex items-start gap-2 rounded-md bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 text-sm text-red-700 dark:text-red-400 ${className}`}>
      <span className="shrink-0 mt-0.5">✕</span>
      <span className="flex-1">{error}</span>
      {onDismiss && (
        <button onClick={onDismiss} className="shrink-0 ml-1 text-red-400 hover:text-red-600">✕</button>
      )}
    </div>
  )
}
