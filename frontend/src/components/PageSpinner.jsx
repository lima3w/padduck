export default function PageSpinner({ message = 'Loading...' }) {
  return (
    <div className="flex items-center justify-center py-16 text-gray-400 dark:text-gray-500">
      <svg className="animate-spin w-5 h-5 mr-2 text-blue-500" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
      </svg>
      <span className="text-sm">{message}</span>
    </div>
  )
}
