import { useState } from 'react'
import i18n, { AVAILABLE_LOCALES } from '../i18n'
import { getStoredItem, setStoredItem, STORAGE_KEYS } from '../utils/storageKeys'
import * as client from '../api/auth'

export { AVAILABLE_LOCALES }

/** Manages the active UI locale: i18next language, localStorage, and best-effort server persistence. */
export function useLocale() {
  const [locale, setLocaleState] = useState(() => {
    try {
      return getStoredItem(STORAGE_KEYS.locale) || i18n.language || 'en'
    } catch {
      return i18n.language || 'en'
    }
  })

  async function setLocale(next) {
    setLocaleState(next)
    setStoredItem(STORAGE_KEYS.locale, next)
    await i18n.changeLanguage(next)
    try {
      // Best-effort: localStorage remains the source of truth even if this fails
      // (e.g. logged out, or offline).
      await client.updateMyLocale(next)
    } catch {}
  }

  return { locale, setLocale }
}
