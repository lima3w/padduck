import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import Modal from '../components/Modal'
import Layout from '../components/Layout'

vi.mock('../components/Header', () => ({
  default: () => <header>Header</header>,
}))

vi.mock('../components/Sidebar', () => ({
  default: () => <aside>Sidebar</aside>,
}))

vi.mock('../components/CommandPalette', () => ({
  default: () => null,
}))

vi.mock('../hooks/useDarkMode', () => ({
  useDarkMode: () => ({ mode: 'system', setPreference: vi.fn() }),
}))

describe('shared accessibility surfaces', () => {
  it('labels modals as dialogs and closes them with Escape', () => {
    const onClose = vi.fn()
    render(
      <Modal title="Delete subnet" onClose={onClose}>
        <button type="button">Confirm</button>
      </Modal>
    )

    expect(screen.getByRole('dialog', { name: 'Delete subnet' })).toHaveAttribute('aria-modal', 'true')
    expect(screen.getByRole('button', { name: 'Close dialog' })).toBeInTheDocument()

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('provides a skip link to the main content region', () => {
    render(
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
    )

    expect(screen.getByRole('link', { name: 'Skip to main content' })).toHaveAttribute('href', '#main-content')
    expect(document.querySelector('#main-content')).toBeInTheDocument()
  })
})
