import { createContext, useContext, useState, useCallback, useEffect, useRef } from 'react'

const ToastContext = createContext(null)

let _globalShowToast = null

/** Called by the Axios interceptor (outside React) to surface API errors. */
export function showToastGlobal(message, type = 'error') {
  _globalShowToast?.(message, type)
}

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])
  const counter = useRef(0)

  const showToast = useCallback((message, type = 'success') => {
    const id = ++counter.current
    setToasts(prev => [...prev, { id, message, type }])
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), 4000)
  }, [])

  useEffect(() => {
    _globalShowToast = showToast
    return () => { _globalShowToast = null }
  }, [showToast])

  return (
    <ToastContext.Provider value={showToast}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
        {toasts.map(t => (
          <div
            key={t.id}
            className={`px-4 py-2 rounded shadow-lg text-sm font-medium pointer-events-auto transition-all
              ${t.type === 'error'
                ? 'bg-red-600 text-white'
                : 'bg-green-600 text-white'
              }`}
          >
            {t.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  return useContext(ToastContext)
}
