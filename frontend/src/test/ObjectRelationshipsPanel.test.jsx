import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import ObjectRelationshipsPanel from '../components/ObjectRelationshipsPanel'

describe('ObjectRelationshipsPanel', () => {
  it('renders relationship links, counts, and descriptions', () => {
    render(
      <MemoryRouter>
        <ObjectRelationshipsPanel
          relationships={[
            { label: 'Location', value: 'Datacenter A', to: '/locations/1', description: 'Physical assignment' },
            { label: 'Devices', value: 'Assigned devices', count: 3, description: '3 devices mapped' },
          ]}
        />
      </MemoryRouter>
    )

    expect(screen.getByRole('heading', { name: 'Relationships' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Datacenter A' })).toHaveAttribute('href', '/locations/1')
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('3 devices mapped')).toBeInTheDocument()
  })

  it('does not render when there are no relationships', () => {
    const { container } = render(<ObjectRelationshipsPanel relationships={[]} />)
    expect(container).toBeEmptyDOMElement()
  })
})
