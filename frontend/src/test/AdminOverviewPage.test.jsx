import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AdminOverviewPage, { ADMIN_SURFACE_SECTIONS } from '../pages/AdminOverviewPage'

describe('AdminOverviewPage', () => {
  it('groups core admin surfaces in one overview', () => {
    render(
      <MemoryRouter>
        <AdminOverviewPage />
      </MemoryRouter>
    )

    expect(screen.getByRole('heading', { name: 'Admin' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Identity & Access' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Configuration' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Integrations & Automation' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Discovery & Operations' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Audit, Reports & Data' })).toBeInTheDocument()
  })

  it('links every configured admin surface', () => {
    render(
      <MemoryRouter>
        <AdminOverviewPage />
      </MemoryRouter>
    )

    for (const network of ADMIN_SURFACE_SECTIONS) {
      for (const link of network.links) {
        expect(screen.getByRole('link', { name: new RegExp(link.title) })).toHaveAttribute('href', link.to)
      }
    }
  })
})
