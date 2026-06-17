export default function SortTh({ label, col, sortCol, sortDir, onSort, className = '' }) {
  const active = sortCol === col
  return (
    <th
      className={`text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium cursor-pointer select-none hover:text-blue-600 dark:hover:text-blue-400 ${className}`}
      onClick={() => onSort(col)}
    >
      {label}
      <span className="ml-1 text-xs">
        {active ? (sortDir === 'asc' ? '↑' : '↓') : <span className="opacity-30">↕</span>}
      </span>
    </th>
  )
}
