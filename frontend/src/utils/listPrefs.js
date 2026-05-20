import { getStoredItem, setStoredItem } from './storageKeys'

export function loadPrefs(key, defaults, legacyKey) {
  try {
    const saved = JSON.parse(getStoredItem(key, legacyKey))
    if (saved && typeof saved === 'object' && !Array.isArray(saved)) {
      return { ...defaults, ...saved }
    }
  } catch {}
  return { ...defaults }
}

export function savePrefs(key, value, legacyKey) {
  try { setStoredItem(key, JSON.stringify(value), legacyKey) } catch {}
}

export function loadColPrefs(key, defaults, legacyKey) {
  try {
    const saved = JSON.parse(getStoredItem(key, legacyKey))
    if (saved && typeof saved === 'object') return { ...defaults, ...saved }
  } catch {}
  return { ...defaults }
}

export function saveColPrefs(key, value, legacyKey) {
  try { setStoredItem(key, JSON.stringify(value), legacyKey) } catch {}
}
