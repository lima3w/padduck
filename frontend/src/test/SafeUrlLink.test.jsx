import { render, screen } from '@testing-library/react'
import SafeUrlLink from '../components/SafeUrlLink'
import { isSafeHttpUrl } from '../utils/url'

describe('isSafeHttpUrl', () => {
  it('accepts http and https URLs', () => {
    expect(isSafeHttpUrl('https://example.com')).toBe(true)
    expect(isSafeHttpUrl('http://example.com/path?q=1')).toBe(true)
    expect(isSafeHttpUrl('HTTPS://EXAMPLE.COM')).toBe(true)
    expect(isSafeHttpUrl('  https://example.com  ')).toBe(true)
  })

  it('rejects dangerous and malformed values', () => {
    expect(isSafeHttpUrl('javascript:alert(1)')).toBe(false)
    expect(isSafeHttpUrl('data:text/html;base64,PHNjcmlwdD4=')).toBe(false)
    expect(isSafeHttpUrl('vbscript:msgbox(1)')).toBe(false)
    expect(isSafeHttpUrl('//evil.example.com')).toBe(false)
    expect(isSafeHttpUrl('example.com')).toBe(false)
    expect(isSafeHttpUrl('')).toBe(false)
    expect(isSafeHttpUrl(null)).toBe(false)
    expect(isSafeHttpUrl(undefined)).toBe(false)
  })
})

describe('SafeUrlLink', () => {
  it('renders an anchor for https URLs', () => {
    render(<SafeUrlLink value="https://example.com/docs" />)
    const link = screen.getByRole('link', { name: 'https://example.com/docs' })
    expect(link).toHaveAttribute('href', 'https://example.com/docs')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('renders javascript: values as plain text, not a link', () => {
    render(<SafeUrlLink value="javascript:alert(1)" />)
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
    expect(screen.getByText('javascript:alert(1)')).toBeInTheDocument()
  })

  it('renders data: values as plain text, not a link', () => {
    render(<SafeUrlLink value="data:text/html,<script>alert(1)</script>" />)
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })
})
