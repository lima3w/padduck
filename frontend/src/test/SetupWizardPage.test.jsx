import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router'
import SetupWizardPage from '../pages/SetupWizardPage'

vi.mock('../api/admin', () => ({
  getAdminConfig: vi.fn(),
  updateAdminConfig: vi.fn(),
}))

import { getAdminConfig, updateAdminConfig } from '../api/admin'

function renderWizard() {
  return render(
    <MemoryRouter initialEntries={['/setup']}>
      <Routes>
        <Route path="/setup" element={<SetupWizardPage />} />
        <Route path="/dashboard" element={<div>Dashboard page</div>} />
      </Routes>
    </MemoryRouter>
  )
}

describe('SetupWizardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getAdminConfig.mockResolvedValue({ data: { config: {} } })
    updateAdminConfig.mockResolvedValue({ data: {} })
  })

  it('starts on the telemetry step with Next disabled until a choice is made', async () => {
    renderWizard()
    expect(await screen.findByText('Help Make Padduck Better')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Back' })).toBeDisabled()
  })

  it('saves each step on Next and finishes with the completed flag', async () => {
    const user = userEvent.setup()
    renderWizard()

    await user.click(await screen.findByRole('radio', { name: 'Enable Telemetry' }))
    await user.click(screen.getByRole('button', { name: 'Next' }))

    await waitFor(() => expect(updateAdminConfig).toHaveBeenCalledWith({ telemetry_enabled: 'true' }))
    expect(await screen.findByText('Stay up to date')).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: 'No thanks' }))
    await user.click(screen.getByRole('button', { name: 'Finish' }))

    await waitFor(() => expect(updateAdminConfig).toHaveBeenLastCalledWith({
      update_check_enabled: 'false',
      setup_wizard_completed: 'true',
    }))
    expect(await screen.findByText('Dashboard page')).toBeInTheDocument()
  })

  it('Back returns to the previous step keeping the choice', async () => {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('radio', { name: 'No Thanks' }))
    await user.click(screen.getByRole('button', { name: 'Next' }))
    await screen.findByText('Stay up to date')

    await user.click(screen.getByRole('button', { name: 'Back' }))
    expect(await screen.findByText('Help Make Padduck Better')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'No Thanks' })).toHaveAttribute('aria-checked', 'true')
  })

  it('Skip this step advances without saving that step', async () => {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('button', { name: 'Skip this step' }))

    expect(await screen.findByText('Stay up to date')).toBeInTheDocument()
    expect(updateAdminConfig).not.toHaveBeenCalled()
  })

  it('skipping the last step finishes the wizard without saving a choice for it', async () => {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('button', { name: 'Skip this step' }))
    await screen.findByText('Stay up to date')
    await user.click(screen.getByRole('button', { name: 'Skip this step' }))

    await waitFor(() => expect(updateAdminConfig).toHaveBeenCalledWith({ setup_wizard_completed: 'true' }))
    expect(await screen.findByText('Dashboard page')).toBeInTheDocument()
  })

  it('"Skip setup for now" dismisses the wizard', async () => {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('button', { name: 'Skip setup for now' }))

    await waitFor(() => expect(updateAdminConfig).toHaveBeenCalledWith({ setup_wizard_completed: 'true' }))
    expect(await screen.findByText('Dashboard page')).toBeInTheDocument()
  })

  it('stays on the step and shows an error when saving fails', async () => {
    const user = userEvent.setup()
    updateAdminConfig.mockRejectedValueOnce(new Error('boom'))
    renderWizard()
    await user.click(await screen.findByRole('radio', { name: 'Enable Telemetry' }))
    await user.click(screen.getByRole('button', { name: 'Next' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Failed to save your choice')
    expect(screen.getByText('Help Make Padduck Better')).toBeInTheDocument()
  })

  it('prefills stored answers', async () => {
    getAdminConfig.mockResolvedValue({ data: { config: { telemetry_enabled: 'true' } } })
    renderWizard()
    await waitFor(() => expect(screen.getByRole('radio', { name: 'Enable Telemetry' })).toHaveAttribute('aria-checked', 'true'))
  })
})
