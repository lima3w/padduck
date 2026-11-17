import { renderHook, waitFor } from '@testing-library/react'
import useUniFiLinks from '../hooks/useUniFiLinks'

vi.mock('../api/unifi', () => ({ getUniFiLinks: vi.fn() }))
import { getUniFiLinks } from '../api/unifi'

describe('useUniFiLinks', () => {
  it('indexes VLAN and device links by Padduck id', async () => {
    getUniFiLinks.mockResolvedValue({ data: [
      { id: 1, vlanId: 7, unifiName: 'Users' },
      { id: 2, deviceId: 9, unifiName: 'sw-01', details: { model: 'USW-24' } },
    ] })
    const { result } = renderHook(() => useUniFiLinks())
    await waitFor(() => expect(result.current.vlans[7]?.unifiName).toBe('Users'))
    expect(result.current.devices[9].details.model).toBe('USW-24')
  })

  it('stays empty when the lookup fails', async () => {
    getUniFiLinks.mockRejectedValue(new Error('403'))
    const { result } = renderHook(() => useUniFiLinks())
    await waitFor(() => expect(getUniFiLinks).toHaveBeenCalled())
    expect(result.current).toEqual({ vlans: {}, devices: {} })
  })
})
