import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from './locales/en.json'
import fr from './locales/fr.json'
import { getStoredItem, STORAGE_KEYS } from './utils/storageKeys'

// Locales bundled directly (no lazy HTTP loading) so init is synchronous —
// this keeps rendering, and tests, deterministic on first paint.
export const AVAILABLE_LOCALES = [
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'Français' },
]

// Not all test environments provide localStorage (e.g. vitest's default
// "node" environment) — fall back to 'en' rather than throwing on import.
let storedLocale = null
try {
  storedLocale = getStoredItem(STORAGE_KEYS.locale)
} catch {}

i18n
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      fr: { translation: fr },
    },
    lng: storedLocale || 'en',
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // React already escapes values
    },
  })

export default i18n
