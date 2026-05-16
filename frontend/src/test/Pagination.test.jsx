import { render, screen, fireEvent } from '@testing-library/react'
import Pagination from '../components/Pagination'

describe('Pagination', () => {
  it('shows "No results" when total is 0', () => {
    render(<Pagination page={1} limit={20} total={0} onChange={() => {}} />)
    expect(screen.getByText('No results')).toBeInTheDocument()
  })

  it('shows correct range label', () => {
    render(<Pagination page={2} limit={10} total={35} onChange={() => {}} />)
    expect(screen.getByText('Showing 11–20 of 35')).toBeInTheDocument()
  })

  it('clamps "to" at total on last page', () => {
    render(<Pagination page={4} limit={10} total={35} onChange={() => {}} />)
    expect(screen.getByText('Showing 31–35 of 35')).toBeInTheDocument()
  })

  it('calls onChange with next page when Next is clicked', () => {
    const onChange = vi.fn()
    render(<Pagination page={1} limit={10} total={30} onChange={onChange} />)
    fireEvent.click(screen.getByText('Next'))
    expect(onChange).toHaveBeenCalledWith(2)
  })

  it('calls onChange with previous page when Prev is clicked', () => {
    const onChange = vi.fn()
    render(<Pagination page={3} limit={10} total={30} onChange={onChange} />)
    fireEvent.click(screen.getByText('Prev'))
    expect(onChange).toHaveBeenCalledWith(2)
  })

  it('disables Prev on first page', () => {
    render(<Pagination page={1} limit={10} total={30} onChange={() => {}} />)
    expect(screen.getByText('Prev')).toBeDisabled()
  })

  it('disables Next on last page', () => {
    render(<Pagination page={3} limit={10} total={30} onChange={() => {}} />)
    expect(screen.getByText('Next')).toBeDisabled()
  })

  it('renders page buttons for small page count', () => {
    render(<Pagination page={1} limit={10} total={30} onChange={() => {}} />)
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })
})
