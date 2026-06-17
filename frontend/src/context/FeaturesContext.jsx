import { createContext, useContext, useState, useEffect } from 'react'
import { getFeatures } from '../api/app'
import { normalizeFeatures } from '../utils/features'

const FeaturesContext = createContext(null)

export function FeaturesProvider({ children }) {
  const [features, setFeatures] = useState(null)

  useEffect(() => {
    let cancelled = false
    getFeatures()
      .then(res => { if (!cancelled) setFeatures(normalizeFeatures(res.data)) })
      .catch(() => { if (!cancelled) setFeatures(normalizeFeatures()) })
    return () => { cancelled = true }
  }, [])

  return <FeaturesContext.Provider value={features}>{children}</FeaturesContext.Provider>
}

/** Returns the features map, or null while loading. */
export function useFeatures() {
  return useContext(FeaturesContext)
}
