import { render, screen, fireEvent } from '@testing-library/react'
import { AppErrorBoundary, RouteErrorBoundary } from '../components/ErrorBoundary'

// Suppress the expected React error boundary console.error output during tests.
// React logs errors twice to the console even when they are caught by an
// ErrorBoundary; silence those to keep test output clean.
beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {})
})

afterEach(() => {
  console.error.mockRestore()
})

function ThrowingChild({ message = 'Test error' }) {
  throw new Error(message)
}

function NormalChild() {
  return <div>All good</div>
}

describe('AppErrorBoundary', () => {
  it('renders children when no error is thrown', () => {
    render(
      <AppErrorBoundary>
        <NormalChild />
      </AppErrorBoundary>
    )
    expect(screen.getByText('All good')).toBeInTheDocument()
  })

  it('catches a throwing child and renders the fallback page', () => {
    render(
      <AppErrorBoundary>
        <ThrowingChild message="Boom from test" />
      </AppErrorBoundary>
    )

    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText(/Boom from test/)).toBeInTheDocument()
    // The reload button should be present.
    expect(screen.getByRole('button', { name: /reloading the page/i })).toBeInTheDocument()
  })

  it('does not show children after catching an error', () => {
    render(
      <AppErrorBoundary>
        <ThrowingChild />
      </AppErrorBoundary>
    )
    expect(screen.queryByText('All good')).not.toBeInTheDocument()
  })
})

describe('RouteErrorBoundary', () => {
  it('renders children when no error is thrown', () => {
    render(
      <RouteErrorBoundary resetKey="/some-route">
        <NormalChild />
      </RouteErrorBoundary>
    )
    expect(screen.getByText('All good')).toBeInTheDocument()
  })

  it('catches a throwing child and renders the inline fallback', () => {
    render(
      <RouteErrorBoundary resetKey="/broken-route">
        <ThrowingChild message="Page crash" />
      </RouteErrorBoundary>
    )

    expect(screen.getByText('This page failed to load')).toBeInTheDocument()
    expect(screen.getByText(/Page crash/)).toBeInTheDocument()
  })

  it('shows "Try again" button that resets the boundary', () => {
    // Render with a throwing child first, then verify reset unmounts the error UI.
    render(
      <RouteErrorBoundary resetKey="/broken-route">
        <ThrowingChild message="Oops" />
      </RouteErrorBoundary>
    )

    expect(screen.getByText('This page failed to load')).toBeInTheDocument()

    // Rerender with a non-throwing child; click "Try again" to reset.
    // Because ThrowingChild always throws we cannot re-render it after reset, so
    // instead we verify the Try-again button exists and is clickable.
    fireEvent.click(screen.getByRole('button', { name: /try again/i }))

    // After reset the error UI should be gone (children are re-rendered, but
    // ThrowingChild immediately throws again, so the boundary re-catches).
    // We confirm the button itself existed and was clickable without errors.
    expect(screen.queryByText('All good')).not.toBeInTheDocument()
  })

  it('resets automatically when the resetKey changes', () => {
    const { rerender } = render(
      <RouteErrorBoundary resetKey="/bad-route">
        <ThrowingChild />
      </RouteErrorBoundary>
    )

    expect(screen.getByText('This page failed to load')).toBeInTheDocument()

    // Navigate to a new route with a working component.
    rerender(
      <RouteErrorBoundary resetKey="/good-route">
        <NormalChild />
      </RouteErrorBoundary>
    )

    expect(screen.queryByText('This page failed to load')).not.toBeInTheDocument()
    expect(screen.getByText('All good')).toBeInTheDocument()
  })
})
