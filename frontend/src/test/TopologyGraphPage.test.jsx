import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import TopologyGraphPage, { resourcePath } from '../pages/TopologyGraphPage'

const handlers = {}
const cyStub = {
  on: vi.fn((evt, sel, fn) => { handlers[typeof sel === 'function' ? evt : `${evt}:${sel}`] = fn || sel }),
  destroy: vi.fn(),
  batch: (fn) => fn(),
  elements: () => ({ removeClass: vi.fn(), addClass: vi.fn(), not: () => ({ not: () => ({ addClass: vi.fn() }) }), filter: () => ({ removeClass: () => ({ addClass: vi.fn() }) }) }),
  nodes: () => ({ forEach: vi.fn(), filter: () => ({ length: 0 }) }),
  edges: () => ({ forEach: vi.fn() }),
  png: vi.fn(() => 'data:image/png;base64,'),
  animate: vi.fn(),
}
vi.mock('cytoscape', () => ({ default: vi.fn(() => cyStub) }))
vi.mock('../api/topology', () => ({
  getTopologyGraph: vi.fn(),
  getTopologyPath: vi.fn(),
  getTopologyEdges: vi.fn(),
  setTopologyEdgeStatus: vi.fn(),
  syncTopologyGraph: vi.fn(),
}))
let role = 'admin'
vi.mock('../utils/storageKeys', () => ({ getCachedUser: () => ({ role }) }))

import cytoscape from 'cytoscape'
import { getTopologyGraph, getTopologyPath, getTopologyEdges, setTopologyEdgeStatus, syncTopologyGraph } from '../api/topology'

const graph = {
  nodes: [
    { data: { id: 'device_1', resourceType: 'device', resourceId: 1, label: 'core-sw', detail: 'Ubiquiti USW' } },
    { data: { id: 'subnet_2', resourceType: 'subnet', resourceId: 2, label: '10.0.0.0/24' } },
    { data: { id: 'vlan_3', resourceType: 'vlan', resourceId: 3, label: 'VLAN 10' } },
  ],
  edges: [
    { data: { id: 'edge_1', source: 'device_1', target: 'subnet_2', relationship: 'member_of', status: 'confirmed' } },
    { data: { id: 'edge_2', source: 'subnet_2', target: 'vlan_3', relationship: 'member_of', status: 'suggested' } },
  ],
}
const suggestion = { id: 2, relationship: 'uplink_to', status: 'suggested', evidence: 'UniFi uplink via port 7', sourceLabel: 'ap-1', targetLabel: 'core-sw' }

function renderPage() {
  return render(<MemoryRouter><TopologyGraphPage /></MemoryRouter>)
}

describe('TopologyGraphPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    Object.keys(handlers).forEach(k => delete handlers[k])
    role = 'admin'
    getTopologyGraph.mockResolvedValue({ data: graph })
    getTopologyEdges.mockResolvedValue({ data: { edges: [suggestion] } })
  })

  it('builds the cytoscape graph from the API payload', async () => {
    renderPage()
    await waitFor(() => expect(cytoscape).toHaveBeenCalled())
    const { elements } = cytoscape.mock.calls[0][0]
    expect(elements.filter(e => e.data.source)).toHaveLength(2)
    expect(elements.find(e => e.data.id === 'device_1').data).toMatchObject({ label: 'core-sw', type: 'device' })
  })

  it('lists suggested relationships and lets an admin confirm one', async () => {
    setTopologyEdgeStatus.mockResolvedValue({})
    renderPage()
    expect(await screen.findByText('UniFi uplink via port 7', { exact: false })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Confirm' }))
    expect(setTopologyEdgeStatus).toHaveBeenCalledWith(2, 'confirmed')
    await waitFor(() => expect(getTopologyGraph).toHaveBeenCalledTimes(2))
  })

  it('dismisses a suggestion', async () => {
    setTopologyEdgeStatus.mockResolvedValue({})
    renderPage()
    await userEvent.click(await screen.findByRole('button', { name: 'Dismiss' }))
    expect(setTopologyEdgeStatus).toHaveBeenCalledWith(2, 'dismissed')
  })

  it('hides review and rebuild controls from non-admins', async () => {
    role = 'user'
    renderPage()
    await screen.findByTestId('suggested-edges')
    expect(screen.queryByRole('button', { name: 'Confirm' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Rebuild graph' })).not.toBeInTheDocument()
  })

  it('rebuilds the graph on request', async () => {
    syncTopologyGraph.mockResolvedValue({})
    renderPage()
    await userEvent.click(await screen.findByRole('button', { name: 'Rebuild graph' }))
    expect(syncTopologyGraph).toHaveBeenCalled()
    await waitFor(() => expect(getTopologyGraph).toHaveBeenCalledTimes(2))
  })

  it('shows details, relationships and a link for a clicked node', async () => {
    renderPage()
    await waitFor(() => expect(handlers['tap:node']).toBeDefined())
    act(() => handlers['tap:node']({ target: { id: () => 'device_1' } }))
    const panel = await screen.findByRole('complementary')
    expect(panel).toHaveTextContent('core-sw')
    expect(panel).toHaveTextContent('Ubiquiti USW')
    expect(panel).toHaveTextContent('member of 10.0.0.0/24')
    expect(screen.getByRole('link', { name: 'Open resource' })).toHaveAttribute('href', '/devices/1')
  })

  it('finds a path between the selected node and another', async () => {
    getTopologyPath.mockResolvedValue({ data: { nodes: [{ data: { id: 'device_1' } }, { data: { id: 'subnet_2' } }, { data: { id: 'vlan_3' } }], edges: [] } })
    renderPage()
    await waitFor(() => expect(handlers['tap:node']).toBeDefined())
    act(() => handlers['tap:node']({ target: { id: () => 'device_1' } }))
    await userEvent.selectOptions(await screen.findByLabelText('Path to...'), 'vlan_3')
    await userEvent.click(screen.getByRole('button', { name: 'Find path' }))
    expect(getTopologyPath).toHaveBeenCalledWith({ type: 'device', id: 1 }, { type: 'vlan', id: 3 })
    expect(await screen.findByRole('status')).toHaveTextContent('Path found: 2 hops')
  })

  it('reports a load failure', async () => {
    getTopologyGraph.mockRejectedValue(new Error('boom'))
    renderPage()
    expect(await screen.findByText('Failed to load the topology graph')).toBeInTheDocument()
  })
})

describe('resourcePath', () => {
  it('maps each resource type to its page', () => {
    expect(resourcePath('subnet', 5)).toBe('/subnets/5/ip-addresses')
    expect(resourcePath('device', 5)).toBe('/devices/5')
    expect(resourcePath('vlan', 5)).toBe('/vlans/5')
    expect(resourcePath('location', 5)).toBe('/locations/5')
    expect(resourcePath('other', 5)).toBeNull()
  })
})
