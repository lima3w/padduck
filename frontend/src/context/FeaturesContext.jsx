import { createContext, useContext, useState, useEffect } from 'react'
import { getFeatures } from '../api/app'
import { normalizeFeatures } from '../utils/features'

const FeaturesContext = createContext(null)

const FEATURES_CHANGED_EVENT = 'padduck:features-changed'

/** Tells the provider to refetch features (call after saving feature flags). */
export function notifyFeaturesChanged() {
  window.dispatchEvent(new Event(FEATURES_CHANGED_EVENT))
}

export function FeaturesProvider({ children }) {
  const [features, setFeatures] = useState(null)

  useEffect(() => {
    let cancelled = false
    function load() {
      getFeatures()
        .then(res => { if (!cancelled) setFeatures(normalizeFeatures(res.data)) })
        .catch(() => { if (!cancelled) setFeatures(normalizeFeatures()) })
    }
    load()
    window.addEventListener(FEATURES_CHANGED_EVENT, load)
    return () => {
      cancelled = true
      window.removeEventListener(FEATURES_CHANGED_EVENT, load)
    }
  }, [])

  return <FeaturesContext.Provider value={features}>{children}</FeaturesContext.Provider>
}

/** Returns the features map, or null while loading. */
export function useFeatures() {
  return useContext(FeaturesContext)
}
