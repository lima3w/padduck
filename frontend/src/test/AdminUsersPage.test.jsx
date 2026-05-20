import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import AdminUsersPage from '../pages/AdminUsersPage'

vi.mock('../api/client', () => ({
  getAdminUsers: vi.fn(),
  getAdminRoles: vi.fn(),
  getUserRoles: vi.fn(),
  assignUserRole: vi.fn(),
  removeUserRole: vi.fn(),
  createUser: vi.fn(),
  adminUnlockUser: vi.fn(),
  suspendUser: vi.fn(),
  unsuspendUser: vi.fn(),
  impersonateUser: vi.fn(),
  sendPasswordResetEmail: vi.fn(),
  updateUserEmail: vi.fn(),
  gdprDeleteUser: vi.fn(),
  bulkSuspendUsers: vi.fn(),
  bulkActivateUsers: vi.fn(),
  bulkDeleteUsers: vi.fn(),
}))

vi.mock('../api/locations', () => ({
  getLocations: vi.fn(),
}))

import { getAdminUsers, getAdminRoles, getUserRoles } from '../api/client'
import { getLocations } from '../api/locations'

describe('AdminUsersPage', () => {
  it('renders assigned roles returned with Go PascalCase fields', async () => {
    getAdminUsers.mockResolvedValue({
      data: [
        {
          id: 7,
          username: 'alice',
          email: 'alice@example.com',
          role: 'user',
          state: 'active',
        },
      ],
    })
    getAdminRoles.mockResolvedValue({
      data: [
        {
          ID: 3,
          Name: 'Network Operator',
          Description: 'Can manage network records',
        },
      ],
    })
    getUserRoles.mockResolvedValue({
      data: [
        {
          ID: 3,
          Name: 'Network Operator',
          Description: 'Can manage network records',
        },
      ],
    })
    getLocations.mockResolvedValue([])

    render(<AdminUsersPage />)

    await screen.findByText('alice')
    fireEvent.click(screen.getByText('alice'))

    await waitFor(() => {
      expect(screen.getByText('Network Operator')).toBeInTheDocument()
    })
    expect(screen.queryByText(/Role #undefined/)).not.toBeInTheDocument()
  })
})
