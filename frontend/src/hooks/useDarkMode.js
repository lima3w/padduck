import { useState, useEffect } from 'react'
import { getStoredItem, LEGACY_STORAGE_KEYS, setStoredItem, STORAGE_KEYS } from '../utils/storageKeys'

const STORAGE_KEY = STORAGE_KEYS.colorScheme
const LEGACY_STORAGE_KEY = LEGACY_STORAGE_KEYS.colorScheme

export function useDarkMode() {
  const [mode, setMode] = useState(() => {
    const stored = getStoredItem(STORAGE_KEY, LEGACY_STORAGE_KEY)
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
    setStoredItem(STORAGE_KEY, pref, LEGACY_STORAGE_KEY)
    setMode(pref)
  }

  return { mode, setPreference }
}
