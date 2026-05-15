import { useState, useEffect } from 'react'
import { api } from '../api/client'
import Modal from '../components/Modal'

function isOnline(lastSeen) {
  if (!lastSeen) return false
  return Date.now() - new Date(lastSeen).getTime() < 15 * 60 * 1000
}

export default function AdminAgentsPage() {
  const [agents, setAgents] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [createModal, setCreateModal] = useState(false)
  const [newName, setNewName] = useState('')
  const [saving, setSaving] = useState(false)
  const [newToken, setNewToken] = useState(null)
  const [newTokenAgentName, setNewTokenAgentName] = useState('')
  const [rotatingId, setRotatingId] = useState(null)
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  useEffect(() => {
    loadAgents()
  }, [])

  async function loadAgents() {
    try {
      const { data } = await api.get('/admin/scan-agents')
      setAgents(data || [])
    } catch {
      setError('Failed to load scan agents')
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const { data } = await api.post('/admin/scan-agents', { name: newName })
      setNewToken(data.token)
      setNewTokenAgentName(newName)
      setCreateModal(false)
      setNewName('')
      await loadAgents()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create agent')
    } finally {
      setSaving(false)
    }
  }

  async function handleRotate(agent) {
    setRotatingId(agent.id)
    try {
      const { data } = await api.post(`/admin/scan-agents/${agent.id}/rotate-token`)
      setNewToken(data.token)
      setNewTokenAgentName(agent.name)
    } catch {
      setError('Failed to rotate token')
    } finally {
      setRotatingId(null)
    }
  }

  async function handleDelete(id) {
    try {
      await api.delete(`/admin/scan-agents/${id}`)
      setDeleteConfirm(null)
      await loadAgents()
    } catch {
      setError('Failed to delete agent')
    }
  }

  function copyToken() {
    if (newToken) navigator.clipboard.writeText(newToken).catch(() => {})
  }

  if (loading) return <div className="p-6 text-gray-500">Loading…</div>

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Scan Agents</h1>
        <button
          onClick={() => { setNewName(''); setCreateModal(true) }}
          className="text-sm bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 transition"
        >
          + New Agent
        </button>
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>
      )}

      {newToken && (
        <div className="mb-6 p-4 bg-amber-50 border border-amber-300 rounded-lg">
          <p className="text-sm font-semibold text-amber-800 mb-1">
            Token for <span className="font-mono">{newTokenAgentName}</span> — copy it now, it will not be shown again.
          </p>
          <div className="flex items-center gap-2 mt-2">
            <code className="flex-1 bg-white border border-amber-200 rounded px-3 py-2 text-xs font-mono text-gray-800 break-all select-all">
              {newToken}
            </code>
            <button
              onClick={copyToken}
              className="shrink-0 px-3 py-2 bg-amber-600 text-white text-xs rounded hover:bg-amber-700 transition"
            >
              Copy
            </button>
          </div>
          <button
            onClick={() => setNewToken(null)}
            className="mt-2 text-xs text-amber-600 hover:underline"
          >
            Dismiss
          </button>
        </div>
      )}

      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <table className="min-w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-100">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Last Seen</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {agents.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-gray-400 text-sm">No scan agents configured.</td>
              </tr>
            ) : (
              agents.map((agent) => (
                <tr key={agent.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900 text-sm">{agent.name}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                      isOnline(agent.last_seen) ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                    }`}>
                      {isOnline(agent.last_seen) ? 'Online' : 'Offline'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-500 text-xs">
                    {agent.last_seen ? new Date(agent.last_seen).toLocaleString() : '—'}
                  </td>
                  <td className="px-4 py-3 text-right space-x-2">
                    <button
                      onClick={() => handleRotate(agent)}
                      disabled={rotatingId === agent.id}
                      className="text-xs text-gray-500 hover:text-blue-600 disabled:opacity-50"
                    >
                      {rotatingId === agent.id ? 'Rotating…' : 'Rotate Token'}
                    </button>
                    {deleteConfirm === agent.id ? (
                      <span className="inline-flex items-center gap-1 text-xs">
                        <span className="text-red-600">Delete?</span>
                        <button onClick={() => handleDelete(agent.id)} className="text-red-600 font-medium hover:text-red-800">Yes</button>
                        <button onClick={() => setDeleteConfirm(null)} className="text-gray-400 hover:text-gray-600">No</button>
                      </span>
                    ) : (
                      <button
                        onClick={() => setDeleteConfirm(agent.id)}
                        className="text-xs text-gray-500 hover:text-red-600"
                      >
                        Delete
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {createModal && (
        <Modal title="New Scan Agent" onClose={() => setCreateModal(false)}>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Agent Name <span className="text-red-500">*</span></label>
              <input
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="branch-office"
                value={newName}
                onChange={e => setNewName(e.target.value)}
                autoFocus
                required
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setCreateModal(false)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
              <button type="submit" disabled={saving || !newName.trim()} className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50">
                {saving ? 'Creating…' : 'Create'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
