import { useState, useEffect } from 'react'
import { getAutomationPolicies, createAutomationPolicy, updateAutomationPolicy, deleteAutomationPolicy } from '../api/admin'

const EMPTY_FORM = { name: '', workflow: '*', action: '*', effect: 'allow', message: '', conditions: '', enabled: true }

function effectBadge(effect) {
  if (effect === 'deny') return 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
  if (effect === 'manual_review') return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
  return 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
}

export default function AutomationPoliciesPage() {
  const [policies, setPolicies] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showModal, setShowModal] = useState(false)
  const [editingID, setEditingID] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState(null)

  function load() {
    setLoading(true)
    getAutomationPolicies()
      .then(res => setPolicies(res.data || []))
      .catch(err => setError(err.response?.data?.error || 'Failed to load policies'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  function openCreate() {
    setForm(EMPTY_FORM)
    setEditingID(null)
    setFormError(null)
    setShowModal(true)
  }

  function openEdit(p) {
    const condStr = Object.entries(p.conditions || {}).map(([k, v]) => `${k}=${v}`).join(', ')
    setForm({ name: p.name, workflow: p.workflow, action: p.action, effect: p.effect, message: p.message || '', conditions: condStr, enabled: p.enabled })
    setEditingID(p.id)
    setFormError(null)
    setShowModal(true)
  }

  function parseConditions(str) {
    const result = {}
    str.split(',').forEach(pair => {
      const [k, v] = pair.split('=').map(s => s.trim())
      if (k && v) result[k] = v
    })
    return result
  }

  async function handleSave(e) {
    e.preventDefault()
    setSaving(true)
    setFormError(null)
    const payload = {
      name: form.name.trim(),
      workflow: form.workflow.trim() || '*',
      action: form.action.trim() || '*',
      effect: form.effect,
      message: form.message.trim(),
      conditions: parseConditions(form.conditions),
      enabled: form.enabled,
    }
    try {
      if (editingID) {
        await updateAutomationPolicy(editingID, payload)
      } else {
        await createAutomationPolicy(payload)
      }
      setShowModal(false)
      load()
    } catch (err) {
      setFormError(err.response?.data?.error || 'Failed to save policy')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    if (!window.confirm('Delete this policy?')) return
    try {
      await deleteAutomationPolicy(id)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to delete policy')
    }
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Automation Policies</h1>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Rules that allow, deny, or require manual review before automation actions are applied.
          </p>
        </div>
        <button
          onClick={openCreate}
          className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
          New Policy
        </button>
      </div>

      {loading && <p className="text-sm text-gray-500 dark:text-gray-400">Loading…</p>}
      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      {!loading && !error && (
        <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800">
              <tr>
                {['Name', 'Workflow', 'Action', 'Effect', 'Message', 'Enabled', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {policies.length === 0 ? (
                <tr><td colSpan={7} className="px-4 py-6 text-center text-gray-400">No policies configured.</td></tr>
              ) : policies.map(p => (
                <tr key={p.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{p.name}</td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-600 dark:text-gray-400">{p.workflow}</td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-600 dark:text-gray-400">{p.action}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${effectBadge(p.effect)}`}>{p.effect}</span>
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 max-w-xs truncate">{p.message || '—'}</td>
                  <td className="px-4 py-3">
                    {p.enabled
                      ? <span className="text-green-600 dark:text-green-400">Yes</span>
                      : <span className="text-gray-400">No</span>}
                  </td>
                  <td className="px-4 py-3 text-right space-x-2 whitespace-nowrap">
                    <button onClick={() => openEdit(p)} className="text-blue-600 hover:text-blue-800 dark:text-blue-400 text-xs">Edit</button>
                    <button onClick={() => handleDelete(p.id)} className="text-red-600 hover:text-red-800 dark:text-red-400 text-xs">Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-lg">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                {editingID ? 'Edit Policy' : 'New Policy'}
              </h2>
            </div>
            <form onSubmit={handleSave} className="px-6 py-4 space-y-4">
              {formError && <p className="text-sm text-red-600 dark:text-red-400">{formError}</p>}
              {[
                { label: 'Name', key: 'name', required: true },
                { label: 'Workflow (use * for all)', key: 'workflow' },
                { label: 'Action (use * for all)', key: 'action' },
                { label: 'Message (shown when denied/reviewed)', key: 'message' },
                { label: 'Conditions (e.g. network=Production, vrf=Global)', key: 'conditions' },
              ].map(({ label, key, required }) => (
                <div key={key}>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{label}</label>
                  <input
                    className="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    value={form[key]}
                    onChange={e => setForm(f => ({ ...f, [key]: e.target.value }))}
                    required={required}
                  />
                </div>
              ))}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Effect</label>
                <select
                  className="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={form.effect}
                  onChange={e => setForm(f => ({ ...f, effect: e.target.value }))}
                >
                  <option value="allow">allow</option>
                  <option value="deny">deny</option>
                  <option value="manual_review">manual_review</option>
                </select>
              </div>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={form.enabled} onChange={e => setForm(f => ({ ...f, enabled: e.target.checked }))} />
                <span className="text-sm text-gray-700 dark:text-gray-300">Enabled</span>
              </label>

              <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={() => setShowModal(false)} className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded">Cancel</button>
                <button type="submit" disabled={saving} className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors">
                  {saving ? 'Saving…' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
