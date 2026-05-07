import { useState, useEffect } from 'react'

const STORAGE_KEY = 'ipam-color-scheme'

export function useDarkMode() {
  const [mode, setMode] = useState(() => {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) return stored // 'dark' | 'light' | 'system'
    return 'system'
  })

  useEffect(() => {
    const html = document.documentElement
    const mq = window.matchMedia('(prefers-color-scheme: dark)')

    function apply() {
      const effectiveDark = mode === 'dark' || (mode === 'system' && mq.matches)
      html.classList.toggle('dark', effectiveDark)
      html.classList.toggle('light', !effectiveDark)
    }

    apply()
    mq.addEventListener('change', apply)
    return () => mq.removeEventListener('change', apply)
  }, [mode])

  function setPreference(pref) {
    localStorage.setItem(STORAGE_KEY, pref)
    setMode(pref)
  }

  return { mode, setPreference }
}
