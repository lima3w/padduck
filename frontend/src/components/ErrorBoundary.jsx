import React from 'react'

// Full-page fallback rendered by the top-level boundary.
function AppErrorFallback({ error, componentStack }) {
  const isDev = import.meta.env.DEV

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-100 dark:bg-gray-900 p-6">
      <div className="w-full max-w-lg rounded-lg border border-red-200 dark:border-red-800 bg-white dark:bg-gray-800 shadow-md p-6">
        <div className="flex items-center gap-3 mb-4">
          <svg
            className="w-8 h-8 shrink-0 text-red-500 dark:text-red-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
            />
          </svg>
          <h1 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Something went wrong
          </h1>
        </div>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          An unexpected error occurred. Try{' '}
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="text-blue-600 dark:text-blue-400 underline hover:no-underline"
          >
            reloading the page
          </button>
          . If the problem persists, contact your administrator.
        </p>
        {error?.message && (
          <div className="rounded bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 mb-4">
            <p className="text-xs font-medium text-red-700 dark:text-red-400 break-words">
              {error.message}
            </p>
          </div>
        )}
        {isDev && componentStack && (
          <details className="mt-2">
            <summary className="cursor-pointer text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 select-none">
              Component stack (dev only)
            </summary>
            <pre className="mt-2 overflow-auto rounded bg-gray-100 dark:bg-gray-900 p-3 text-xs text-gray-700 dark:text-gray-300 whitespace-pre-wrap break-words">
              {componentStack}
            </pre>
          </details>
        )}
      </div>
    </div>
  )
}

// Inline fallback rendered by the per-route boundary inside the layout.
function RouteErrorFallback({ error, componentStack, onReset }) {
  const isDev = import.meta.env.DEV

  return (
    <div className="p-6 max-w-2xl">
      <div className="rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-5">
        <div className="flex items-start gap-3 mb-3">
          <svg
            className="w-5 h-5 shrink-0 mt-0.5 text-red-500 dark:text-red-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
            />
          </svg>
          <div>
            <h2 className="text-sm font-semibold text-red-800 dark:text-red-300 mb-1">
              This page failed to load
            </h2>
            <p className="text-sm text-red-700 dark:text-red-400">
              {error?.message || 'An unexpected error occurred.'}
            </p>
          </div>
        </div>
        <div className="flex gap-2 mt-4">
          <button
            type="button"
            onClick={onReset}
            className="rounded px-3 py-1.5 text-xs font-medium bg-white dark:bg-gray-800 border border-red-300 dark:border-red-700 text-red-700 dark:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/30 focus:outline-none focus:ring-2 focus:ring-red-400"
          >
            Try again
          </button>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="rounded px-3 py-1.5 text-xs font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 focus:outline-none"
          >
            Reload page
          </button>
        </div>
        {isDev && componentStack && (
          <details className="mt-3">
            <summary className="cursor-pointer text-xs text-red-600 dark:text-red-500 hover:text-red-800 dark:hover:text-red-300 select-none">
              Component stack (dev only)
            </summary>
            <pre className="mt-2 overflow-auto rounded bg-white dark:bg-gray-900 p-3 text-xs text-gray-700 dark:text-gray-300 whitespace-pre-wrap break-words">
              {componentStack}
            </pre>
          </details>
        )}
      </div>
    </div>
  )
}

/**
 * Top-level ErrorBoundary.
 *
 * Wraps the entire app so that any uncaught render error shows a full-page
 * fallback rather than a blank screen.
 */
export class AppErrorBoundary extends React.Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false, error: null, componentStack: null }
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error }
  }

  componentDidCatch(error, info) {
    if (import.meta.env.DEV) {
      console.error('[ErrorBoundary] Uncaught render error:', error, info.componentStack)
    }
    this.setState({ componentStack: info.componentStack ?? null })
  }

  render() {
    if (this.state.hasError) {
      return (
        <AppErrorFallback
          error={this.state.error}
          componentStack={this.state.componentStack}
        />
      )
    }
    return this.props.children
  }
}

/**
 * Per-route ErrorBoundary.
 *
 * Wraps routed content (the <Outlet />) so that a broken page component only
 * breaks the content area — the header, sidebar, and navigation remain
 * functional.
 *
 * The boundary resets automatically when the `resetKey` prop changes (e.g.
 * when the user navigates to a different route).
 */
export class RouteErrorBoundary extends React.Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false, error: null, componentStack: null }
    this.handleReset = this.handleReset.bind(this)
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error }
  }

  componentDidCatch(error, info) {
    if (import.meta.env.DEV) {
      console.error('[RouteErrorBoundary] Caught render error:', error, info.componentStack)
    }
    this.setState({ componentStack: info.componentStack ?? null })
  }

  componentDidUpdate(prevProps) {
    // Reset when the user navigates to a different route.
    if (this.state.hasError && prevProps.resetKey !== this.props.resetKey) {
      this.setState({ hasError: false, error: null, componentStack: null })
    }
  }

  handleReset() {
    this.setState({ hasError: false, error: null, componentStack: null })
  }

  render() {
    if (this.state.hasError) {
      return (
        <RouteErrorFallback
          error={this.state.error}
          componentStack={this.state.componentStack}
          onReset={this.handleReset}
        />
      )
    }
    return this.props.children
  }
}
