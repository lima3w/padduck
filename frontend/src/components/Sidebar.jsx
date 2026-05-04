import { NavLink } from 'react-router-dom'

export default function Sidebar() {
  return (
    <aside className="w-48 bg-gray-800 text-gray-200 min-h-full flex flex-col">
      <nav className="flex flex-col p-4 gap-1">
        <NavLink
          to="/sections"
          className={({ isActive }) =>
            `px-3 py-2 rounded text-sm font-medium transition-colors ${
              isActive ? 'bg-blue-600 text-white' : 'hover:bg-gray-700'
            }`
          }
        >
          Sections
        </NavLink>
      </nav>
    </aside>
  )
}
