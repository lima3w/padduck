import { useAuth } from '../hooks/useAuth'
import { useNavigate } from 'react-router-dom'

export default function Header() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <header className="bg-blue-700 text-white px-6 py-3 flex items-center justify-between shadow">
      <div className="flex items-center">
        <span className="text-xl font-bold tracking-tight">IPAM Next</span>
        <span className="ml-3 text-blue-300 text-sm">IP Address Management</span>
      </div>
      <div className="flex items-center gap-4">
        {user && (
          <>
            <span className="text-sm text-blue-100">{user.username}</span>
            <button
              onClick={handleLogout}
              className="text-sm bg-blue-600 hover:bg-blue-800 px-3 py-1 rounded transition"
            >
              Logout
            </button>
          </>
        )}
      </div>
    </header>
  )
}
