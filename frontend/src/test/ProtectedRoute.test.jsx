import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import ProtectedRoute from '../components/ProtectedRoute'

vi.mock('../hooks/useAuth', () => ({
  useAuth: vi.fn(),
}))

import { useAuth } from '../hooks/useAuth'

function renderProtected() {
  return render(
    <MemoryRouter initialEntries={['/secret']}>
      <Routes>
        <Route path="/login" element={<div>Login Screen</div>} />
        <Route
          path="/secret"
          element={
            <ProtectedRoute>
              <div>Secret Content</div>
            </ProtectedRoute>
          }
        />
      </Routes>
    </MemoryRouter>
  )
}

describe('ProtectedRoute', () => {
  it('shows a loading state while the session check is in flight', () => {
    useAuth.mockReturnValue({ isAuthenticated: false, loading: true })
    renderProtected()
    expect(screen.getByText('Loading...')).toBeInTheDocument()
    expect(screen.queryByText('Secret Content')).not.toBeInTheDocument()
    expect(screen.queryByText('Login Screen')).not.toBeInTheDocument()
  })

  it('redirects unauthenticated users to the login page', () => {
    useAuth.mockReturnValue({ isAuthenticated: false, loading: false })
    renderProtected()
    expect(screen.getByText('Login Screen')).toBeInTheDocument()
    expect(screen.queryByText('Secret Content')).not.toBeInTheDocument()
  })

  it('renders the protected content for authenticated users', () => {
    useAuth.mockReturnValue({ isAuthenticated: true, loading: false })
    renderProtected()
    expect(screen.getByText('Secret Content')).toBeInTheDocument()
    expect(screen.queryByText('Login Screen')).not.toBeInTheDocument()
  })
})
