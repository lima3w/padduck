export const DEFAULT_FEATURES = {
  customers: true,
  vlans: true,
  vrfs: true,
  racks: true,
  locations: true,
  bgp: true,
  devices: true,
}

export function normalizeFeatures(data) {
  return {
    ...DEFAULT_FEATURES,
    ...(data?.features || data || {}),
  }
}
