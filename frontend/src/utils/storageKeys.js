export const STORAGE_KEYS = {
  currentUser: 'padduck_current_user',
  colorScheme: 'padduck-color-scheme',
  ipColumns: 'padduck_ip_columns',
  deviceFilters: 'padduck_filters_devices',
  deviceColumns: 'padduck_cols_devices',
  subnetFilters: 'padduck_filters_subnets',
}

export const LEGACY_STORAGE_KEYS = {
  currentUser: 'current_user',
  colorScheme: 'ipam-color-scheme',
  ipColumns: 'ipam_ip_columns',
  deviceFilters: 'ipam_filters_devices',
  deviceColumns: 'ipam_cols_devices',
  subnetFilters: 'ipam_filters_subnets',
}

export function getStoredItem(key, legacyKey) {
  const value = localStorage.getItem(key)
  if (value !== null || !legacyKey) return value

  const legacyValue = localStorage.getItem(legacyKey)
  if (legacyValue !== null) {
    localStorage.setItem(key, legacyValue)
    localStorage.removeItem(legacyKey)
  }
  return legacyValue
}

export function setStoredItem(key, value, legacyKey) {
  localStorage.setItem(key, value)
  if (legacyKey) localStorage.removeItem(legacyKey)
}

export function removeStoredItem(key, legacyKey) {
  localStorage.removeItem(key)
  if (legacyKey) localStorage.removeItem(legacyKey)
}

export function getCachedUser() {
  const cached = getStoredItem(STORAGE_KEYS.currentUser, LEGACY_STORAGE_KEYS.currentUser)
  if (!cached) return null
  try {
    return JSON.parse(cached)
  } catch {
    return null
  }
}

export function setCachedUser(user) {
  setStoredItem(
    STORAGE_KEYS.currentUser,
    JSON.stringify(user),
    LEGACY_STORAGE_KEYS.currentUser,
  )
}

export function clearCachedUser() {
  removeStoredItem(STORAGE_KEYS.currentUser, LEGACY_STORAGE_KEYS.currentUser)
}
