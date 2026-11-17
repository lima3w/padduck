import { render, screen, waitFor, act } from '@testing-library/react'
import { FeaturesProvider, useFeatures, notifyFeaturesChanged } from '../context/FeaturesContext'

vi.mock('../api/app', () => ({ getFeatures: vi.fn() }))

import { getFeatures } from '../api/app'

function Probe() {
  const features = useFeatures()
  return <div>{features ? `vlans:${features.vlans}` : 'loading'}</div>
}

describe('FeaturesProvider', () => {
  it('refetches features when notified of a change', async () => {
    getFeatures
      .mockResolvedValueOnce({ data: { features: { vlans: true } } })
      .mockResolvedValueOnce({ data: { features: { vlans: false } } })

    render(<FeaturesProvider><Probe /></FeaturesProvider>)
    expect(await screen.findByText('vlans:true')).toBeInTheDocument()

    act(() => notifyFeaturesChanged())

    await waitFor(() => expect(screen.getByText('vlans:false')).toBeInTheDocument())
    expect(getFeatures).toHaveBeenCalledTimes(2)
  })
})
