import { useEffect, useState } from 'react'
import { getUniFiLinks } from '../api/unifi'

// Which VLANs and devices were synced from UniFi, keyed by Padduck id. Empty
// maps until loaded and whenever the lookup fails or UniFi is not in use, so
// callers can render unconditionally.
export default function useUniFiLinks() {
  const [links, setLinks] = useState({ vlans: {}, devices: {} })

  useEffect(() => {
    let cancelled = false
    getUniFiLinks()
      .then(res => {
        if (cancelled) return
        const next = { vlans: {}, devices: {} }
        for (const l of Array.isArray(res.data) ? res.data : []) {
          if (l.vlanId != null) next.vlans[l.vlanId] = l
          if (l.deviceId != null) next.devices[l.deviceId] = l
        }
        setLinks(next)
      })
      .catch(() => {})
    return () => { cancelled = true }
  }, [])

  return links
}
