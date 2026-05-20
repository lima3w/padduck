import { render, screen, fireEvent } from '@testing-library/react'
import TableActions from '../components/TableActions'

describe('TableActions', () => {
  it('shows edit and delete actions before confirmation', () => {
    render(
      <TableActions
        onEdit={() => {}}
        onDelete={() => {}}
        onRequestDelete={() => {}}
      />
    )

    expect(screen.getByRole('button', { name: 'Edit' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
    expect(screen.queryByText('Confirm?')).not.toBeInTheDocument()
  })

  it('asks for confirmation before calling delete', () => {
    const onDelete = vi.fn()
    const onCancelDelete = vi.fn()

    render(
      <TableActions
        onDelete={onDelete}
        confirming
        onCancelDelete={onCancelDelete}
      />
    )

    expect(screen.getByText('Confirm?')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Yes' }))
    expect(onDelete).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('button', { name: 'No' }))
    expect(onCancelDelete).toHaveBeenCalledTimes(1)
  })
})
