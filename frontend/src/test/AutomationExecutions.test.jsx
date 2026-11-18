import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AutomationExecutions, { canRollback } from '../components/AutomationExecutions'

vi.mock('../api/admin', () => ({
  listAutomationExecutions: vi.fn(),
  rollbackAutomationExecution: vi.fn(),
}))
let role = 'admin'
vi.mock('../utils/storageKeys', () => ({ getCachedUser: () => ({ role }) }))

import { listAutomationExecutions, rollbackAutomationExecution } from '../api/admin'

const tagRun = {
  id: 7, policyName: 'tag lab hosts', resourceType: 'ip_address', resourceName: '10.9.0.5', decision: 'allow',
  triggeredBy: 'api', executedAt: '2026-07-19T10:00:00Z',
  actionsTaken: [
    { type: 'tag', status: 'success' },
    { type: 'tag', status: 'failed', error: 'invalid tag_id' },
  ],
}
const denied = { id: 8, policyName: 'no guests', resourceType: 'ip_address', decision: 'deny', triggeredBy: 'api', executedAt: '2026-07-19T09:00:00Z', actionsTaken: [] }

describe('AutomationExecutions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    role = 'admin'
    listAutomationExecutions.mockResolvedValue({ data: { executions: [tagRun, denied] } })
  })

  it('queries by policy and lists each execution with its action outcomes', async () => {
    render(<AutomationExecutions policyId={3} />)
    expect(await screen.findByText('tag lab hosts')).toBeInTheDocument()
    expect(listAutomationExecutions).toHaveBeenCalledWith({ policy_id: 3 })
    expect(screen.getByText('failed')).toBeInTheDocument()
    expect(screen.getByText('(invalid tag_id)')).toBeInTheDocument()
    expect(screen.getByText('no guests')).toBeInTheDocument()
  })

  it('queries by resource', async () => {
    render(<AutomationExecutions resourceType="ip_address" resourceId={12} />)
    await screen.findByText('tag lab hosts')
    expect(listAutomationExecutions).toHaveBeenCalledWith({ resource_type: 'ip_address', resource_id: 12 })
  })

  it('offers rollback only for entries with a reversible, successful action, and asks to confirm', async () => {
    rollbackAutomationExecution.mockResolvedValue({})
    render(<AutomationExecutions policyId={3} />)
    await screen.findByText('tag lab hosts')
    expect(screen.getAllByRole('button', { name: 'Roll back' })).toHaveLength(1)

    await userEvent.click(screen.getByRole('button', { name: 'Roll back' }))
    expect(rollbackAutomationExecution).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Confirm roll back' }))
    expect(rollbackAutomationExecution).toHaveBeenCalledWith(7)
    await waitFor(() => expect(listAutomationExecutions).toHaveBeenCalledTimes(2))
  })

  it('shows the server message when a rollback is refused', async () => {
    rollbackAutomationExecution.mockRejectedValue({ response: { data: { error: { message: 'this execution has already been rolled back' } } } })
    render(<AutomationExecutions policyId={3} />)
    await userEvent.click(await screen.findByRole('button', { name: 'Roll back' }))
    await userEvent.click(screen.getByRole('button', { name: 'Confirm roll back' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('already been rolled back')
  })

  it('hides itself on resource pages when there is no history, and for non-admins', async () => {
    listAutomationExecutions.mockResolvedValue({ data: { executions: [] } })
    const { container } = render(<AutomationExecutions resourceType="device" resourceId={1} hideWhenEmpty />)
    await waitFor(() => expect(listAutomationExecutions).toHaveBeenCalled())
    expect(container).toBeEmptyDOMElement()

    role = 'user'
    listAutomationExecutions.mockClear()
    const { container: c2 } = render(<AutomationExecutions policyId={1} />)
    expect(c2).toBeEmptyDOMElement()
    expect(listAutomationExecutions).not.toHaveBeenCalled()
  })

  it('shows an empty message on the policy view', async () => {
    listAutomationExecutions.mockResolvedValue({ data: { executions: [] } })
    render(<AutomationExecutions policyId={3} />)
    expect(await screen.findByText('No automation policy has matched yet.')).toBeInTheDocument()
  })
})

describe('canRollback', () => {
  it('is false once rolled back, for failed actions and for irreversible types', () => {
    expect(canRollback({ actionsTaken: [{ type: 'tag', status: 'success' }] })).toBe(true)
    expect(canRollback({ actionsTaken: [{ type: 'scan', status: 'success' }] })).toBe(true)
    expect(canRollback({ rolledBackAt: 'x', actionsTaken: [{ type: 'tag', status: 'success' }] })).toBe(false)
    expect(canRollback({ actionsTaken: [{ type: 'tag', status: 'failed' }] })).toBe(false)
    expect(canRollback({ actionsTaken: [{ type: 'notify', status: 'success' }] })).toBe(false)
    expect(canRollback({ actionsTaken: [] })).toBe(false)
  })
})
