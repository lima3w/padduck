export default function SortTh({ col, label, sortCol, sortDir, onSort }) {
  const active = sortCol === col
  return (
    <th
      className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium cursor-pointer select-none hover:text-gray-900 dark:hover:text-white whitespace-nowrap"
      onClick={() => onSort(col)}
    >
      {label}
      <span className="ml-1 inline-block w-3 text-center opacity-60">
        {active ? (sortDir === 'asc' ? '↑' : '↓') : '↕'}
      </span>
    </th>
  )
}
