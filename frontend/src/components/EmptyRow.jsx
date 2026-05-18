export default function EmptyRow({ colSpan = 1, message }) {
  return (
    <tr>
      <td colSpan={colSpan} className="px-4 py-8 text-center text-gray-400 dark:text-gray-500 text-sm">
        {message}
      </td>
    </tr>
  )
}
