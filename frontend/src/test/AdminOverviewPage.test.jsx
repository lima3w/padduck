import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AdminOverviewPage, { buildAdminSurfaceSections } from '../pages/AdminOverviewPage'
import i18n from '../i18n'

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

    const sections = buildAdminSurfaceSections(i18n.t.bind(i18n))
    for (const network of sections) {
      for (const link of network.links) {
        expect(screen.getByRole('link', { name: new RegExp(link.title) })).toHaveAttribute('href', link.to)
      }
    }
  })
})
