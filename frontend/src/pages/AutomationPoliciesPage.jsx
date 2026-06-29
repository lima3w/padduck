import { useState, useEffect } from 'react'
import { getAutomationPolicies, createAutomationPolicy, updateAutomationPolicy, deleteAutomationPolicy } from '../api/admin'

const OPERATORS = [
  { value: 'eq', label: 'equals' },
  { value: 'neq', label: 'not equals' },
  { value: 'contains', label: 'contains' },
  { value: 'starts_with', label: 'starts with' },
  { value: 'ends_with', label: 'ends with' },
  { value: 'gt', label: 'greater than' },
  { value: 'lt', label: 'less than' },
  { value: 'glob', label: 'glob (wildcard *)' },
]

const KNOWN_FIELDS = [
  'network_id', 'parent_subnet_id', 'prefix_len', 'hostname',
  'subnet_id', 'tag_id', 'location_id', 'device_type',
]

const EMPTY_CONDITION = { field: 'hostname', operator: 'eq', value: '' }
const EMPTY_FORM = { name: '', workflow: '*', action: '*', effect: 'allow', message: '', conditions: [], enabled: true }

function effectBadge(effect) {
  if (effect === 'deny') return 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
  if (effect === 'manual_review') return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
  return 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
}

function ConditionRow({ cond, index, onChange, onRemove }) {
  return (
    <div className="flex items-center gap-2">
      <select
        className="flex-1 px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        value={cond.field}
        onChange={e => onChange(index, 'field', e.target.value)}
      >
        {KNOWN_FIELDS.map(f => <option key={f} value={f}>{f}</option>)}
        {!KNOWN_FIELDS.includes(cond.field) && <option value={cond.field}>{cond.field}</option>}
      </select>
      <select
        className="px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        value={cond.operator}
        onChange={e => onChange(index, 'operator', e.target.value)}
      >
        {OPERATORS.map(op => <option key={op.value} value={op.value}>{op.label}</option>)}
      </select>
      <input
        className="flex-1 px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        placeholder="value"
        value={cond.value}
        onChange={e => onChange(index, 'value', e.target.value)}
      />
      <button
        type="button"
        onClick={() => onRemove(index)}
        className="text-gray-400 hover:text-red-600 dark:hover:text-red-400 text-lg leading-none px-1"
        title="Remove condition"
      >×</button>
    </div>
  )
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
  const [showPreview, setShowPreview] = useState(false)

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
    setShowPreview(false)
    setShowModal(true)
  }

  function openEdit(p) {
    setForm({
      name: p.name,
      workflow: p.workflow,
      action: p.action,
      effect: p.effect,
      message: p.message || '',
      conditions: Array.isArray(p.conditions) ? p.conditions.map(c => ({ ...c })) : [],
      enabled: p.enabled,
    })
    setEditingID(p.id)
    setFormError(null)
    setShowPreview(false)
    setShowModal(true)
  }

  function addCondition() {
    setForm(f => ({ ...f, conditions: [...f.conditions, { ...EMPTY_CONDITION }] }))
  }

  function updateCondition(index, key, value) {
    setForm(f => {
      const conds = f.conditions.map((c, i) => i === index ? { ...c, [key]: value } : c)
      return { ...f, conditions: conds }
    })
  }

  function removeCondition(index) {
    setForm(f => ({ ...f, conditions: f.conditions.filter((_, i) => i !== index) }))
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
      conditions: form.conditions.filter(c => c.field && c.value !== ''),
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

  const previewPayload = {
    name: form.name || '(name)',
    workflow: form.workflow || '*',
    action: form.action || '*',
    effect: form.effect,
    message: form.message || undefined,
    conditions: form.conditions.filter(c => c.field && c.value !== ''),
    enabled: form.enabled,
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
                {['Name', 'Workflow', 'Action', 'Effect', 'Conditions', 'Enabled', ''].map(h => (
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
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                    {Array.isArray(p.conditions) && p.conditions.length > 0
                      ? p.conditions.map((c, i) => (
                          <span key={i} className="inline-block mr-1 mb-1 px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 font-mono">
                            {c.field} {c.operator} {c.value}
                          </span>
                        ))
                      : <span className="text-gray-300 dark:text-gray-600">—</span>}
                  </td>
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
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
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
                { label: 'Message (shown when denied or reviewed)', key: 'message' },
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

              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Conditions</label>
                  <button
                    type="button"
                    onClick={addCondition}
                    className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400 font-medium"
                  >
                    + Add Condition
                  </button>
                </div>
                {form.conditions.length === 0 ? (
                  <p className="text-xs text-gray-400 dark:text-gray-500 italic">No conditions — this policy matches all requests for the given workflow/action.</p>
                ) : (
                  <div className="space-y-2">
                    {form.conditions.map((cond, i) => (
                      <ConditionRow
                        key={i}
                        cond={cond}
                        index={i}
                        onChange={updateCondition}
                        onRemove={removeCondition}
                      />
                    ))}
                  </div>
                )}
              </div>

              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={form.enabled} onChange={e => setForm(f => ({ ...f, enabled: e.target.checked }))} />
                <span className="text-sm text-gray-700 dark:text-gray-300">Enabled</span>
              </label>

              <div>
                <button
                  type="button"
                  onClick={() => setShowPreview(p => !p)}
                  className="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 underline"
                >
                  {showPreview ? 'Hide' : 'Preview'} JSON
                </button>
                {showPreview && (
                  <pre className="mt-2 p-3 rounded bg-gray-50 dark:bg-gray-800 text-xs text-gray-700 dark:text-gray-300 overflow-x-auto border border-gray-200 dark:border-gray-700">
                    {JSON.stringify(previewPayload, null, 2)}
                  </pre>
                )}
              </div>

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
