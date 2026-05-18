export function loadPrefs(key, defaults) {
  try {
    const saved = JSON.parse(localStorage.getItem(key))
    if (saved && typeof saved === 'object' && !Array.isArray(saved)) {
      return { ...defaults, ...saved }
    }
  } catch {}
  return { ...defaults }
}

export function savePrefs(key, value) {
  try { localStorage.setItem(key, JSON.stringify(value)) } catch {}
}

export function loadColPrefs(key, defaults) {
  try {
    const saved = JSON.parse(localStorage.getItem(key))
    if (saved && typeof saved === 'object') return { ...defaults, ...saved }
  } catch {}
  return { ...defaults }
}

export function saveColPrefs(key, value) {
  try { localStorage.setItem(key, JSON.stringify(value)) } catch {}
}
