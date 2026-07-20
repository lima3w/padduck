import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getAutomationPolicies, createAutomationPolicy, updateAutomationPolicy, deleteAutomationPolicy, simulateAutomation } from '../api/admin'

const OPERATOR_KEYS = [
  { value: 'eq', labelKey: 'operatorEquals' },
  { value: 'neq', labelKey: 'operatorNotEquals' },
  { value: 'contains', labelKey: 'operatorContains' },
  { value: 'starts_with', labelKey: 'operatorStartsWith' },
  { value: 'ends_with', labelKey: 'operatorEndsWith' },
  { value: 'gt', labelKey: 'operatorGreaterThan' },
  { value: 'lt', labelKey: 'operatorLessThan' },
  { value: 'glob', labelKey: 'operatorGlob' },
]

const KNOWN_FIELDS = [
  'network_id', 'parent_subnet_id', 'prefix_len', 'hostname',
  'subnet_id', 'tag_id', 'location_id', 'device_type',
]

const ACTION_TYPE_KEYS = [
  { value: 'notify', labelKey: 'actionTypeNotify' },
  { value: 'webhook', labelKey: 'actionTypeWebhook' },
  { value: 'audit_annotation', labelKey: 'actionTypeAuditAnnotation' },
  { value: 'scan', labelKey: 'actionTypeScan' },
  { value: 'tag', labelKey: 'actionTypeTag' },
]

const EMPTY_CONDITION = { field: 'hostname', operator: 'eq', value: '' }
const EMPTY_ACTION = { type: 'notify', params: {} }
const EMPTY_FORM = { name: '', workflow: '*', action: '*', effect: 'allow', message: '', conditions: [], actions: [], enabled: true }
const EMPTY_SIM_ROW = { key: '', value: '' }
const EMPTY_SIM_FORM = { workflow: '', action: '', rows: [] }

function decisionColors(decision) {
  if (decision === 'deny') return 'bg-red-50 border-red-300 text-red-800 dark:bg-red-900/30 dark:border-red-700 dark:text-red-300'
  if (decision === 'manual_review') return 'bg-yellow-50 border-yellow-300 text-yellow-800 dark:bg-yellow-900/30 dark:border-yellow-700 dark:text-yellow-300'
  return 'bg-green-50 border-green-300 text-green-800 dark:bg-green-900/30 dark:border-green-700 dark:text-green-300'
}

function effectBadge(effect) {
  if (effect === 'deny') return 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
  if (effect === 'manual_review') return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
  return 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
}

function ConditionRow({ cond, index, onChange, onRemove }) {
  const { t } = useTranslation()
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
        {OPERATOR_KEYS.map(op => <option key={op.value} value={op.value}>{t(`automationPolicies.${op.labelKey}`)}</option>)}
      </select>
      <input
        className="flex-1 px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        placeholder={t('adminCustomFields.valuePlaceholder')}
        value={cond.value}
        onChange={e => onChange(index, 'value', e.target.value)}
      />
      <button
        type="button"
        onClick={() => onRemove(index)}
        className="text-gray-400 hover:text-red-600 dark:hover:text-red-400 text-lg leading-none px-1"
        title={t('automationPolicies.removeConditionTitle')}
      >×</button>
    </div>
  )
}

function ActionParamField({ label, paramKey, action, onChange }) {
  return (
    <input
      className="flex-1 px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      placeholder={label}
      value={action.params[paramKey] || ''}
      onChange={e => onChange({ ...action, params: { ...action.params, [paramKey]: e.target.value } })}
    />
  )
}

function ActionRow({ action, index, onChange, onRemove }) {
  const { t } = useTranslation()
  function handleTypeChange(type) {
    onChange(index, { type, params: {} })
  }
  function handleChange(updated) {
    onChange(index, updated)
  }
  return (
    <div className="flex items-start gap-2">
      <select
        className="px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        value={action.type}
        onChange={e => handleTypeChange(e.target.value)}
      >
        {ACTION_TYPE_KEYS.map(at => <option key={at.value} value={at.value}>{t(`automationPolicies.${at.labelKey}`)}</option>)}
      </select>
      <div className="flex-1 flex flex-wrap gap-2">
        {action.type === 'notify' && (
          <>
            <ActionParamField label={t('automationPolicies.paramUserIdOrBlank')} paramKey="user_id" action={action} onChange={handleChange} />
            <ActionParamField label={t('automationPolicies.paramRole')} paramKey="role" action={action} onChange={handleChange} />
          </>
        )}
        {action.type === 'webhook' && (
          <ActionParamField label={t('automationPolicies.paramWebhookId')} paramKey="webhook_id" action={action} onChange={handleChange} />
        )}
        {action.type === 'audit_annotation' && (
          <ActionParamField label={t('automationPolicies.paramMessage')} paramKey="message" action={action} onChange={handleChange} />
        )}
        {action.type === 'scan' && (
          <ActionParamField label={t('automationPolicies.paramProfileIdOptional')} paramKey="profile_id" action={action} onChange={handleChange} />
        )}
        {action.type === 'tag' && (
          <ActionParamField label={t('automationPolicies.paramTagId')} paramKey="tag_id" action={action} onChange={handleChange} />
        )}
      </div>
      <button
        type="button"
        onClick={() => onRemove(index)}
        className="text-gray-400 hover:text-red-600 dark:hover:text-red-400 text-lg leading-none px-1 mt-0.5"
        title={t('automationPolicies.removeActionTitle')}
      >×</button>
    </div>
  )
}

export default function AutomationPoliciesPage() {
  const { t } = useTranslation()
  const [policies, setPolicies] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showModal, setShowModal] = useState(false)
  const [editingID, setEditingID] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState(null)
  const [showPreview, setShowPreview] = useState(false)
  const [showSimulate, setShowSimulate] = useState(false)
  const [simForm, setSimForm] = useState(EMPTY_SIM_FORM)
  const [simRunning, setSimRunning] = useState(false)
  const [simResult, setSimResult] = useState(null)
  const [simError, setSimError] = useState(null)

  function load() {
    setLoading(true)
    getAutomationPolicies()
      .then(res => setPolicies(res.data || []))
      .catch(err => setError(err.response?.data?.error || t('automationPolicies.loadFailed')))
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
      actions: Array.isArray(p.actions) ? p.actions.map(a => ({ ...a, params: { ...a.params } })) : [],
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

  function addAction() {
    setForm(f => ({ ...f, actions: [...(f.actions || []), { ...EMPTY_ACTION, params: {} }] }))
  }

  function updateAction(index, updated) {
    setForm(f => ({ ...f, actions: (f.actions || []).map((a, i) => i === index ? updated : a) }))
  }

  function removeAction(index) {
    setForm(f => ({ ...f, actions: (f.actions || []).filter((_, i) => i !== index) }))
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
      actions: (form.actions || []).filter(a => a.type),
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
      setFormError(err.response?.data?.error || t('automationPolicies.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id) {
    if (!window.confirm(t('automationPolicies.deleteConfirm'))) return
    try {
      await deleteAutomationPolicy(id)
      load()
    } catch (err) {
      setError(err.response?.data?.error || t('automationPolicies.deleteFailed'))
    }
  }

  function openSimulate() {
    setSimForm(EMPTY_SIM_FORM)
    setSimResult(null)
    setSimError(null)
    setShowSimulate(true)
  }

  async function handleSimulate(e) {
    e.preventDefault()
    setSimRunning(true)
    setSimResult(null)
    setSimError(null)
    const context = Object.fromEntries(
      simForm.rows.filter(r => r.key.trim()).map(r => [r.key.trim(), r.value])
    )
    try {
      const res = await simulateAutomation({ workflow: simForm.workflow.trim(), action: simForm.action.trim(), context })
      setSimResult(res.data)
    } catch (err) {
      setSimError(err.response?.data?.error || t('automationPolicies.simulationFailed'))
    } finally {
      setSimRunning(false)
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
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('automationPolicies.title')}</h1>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            {t('automationPolicies.subtitle')}
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={openSimulate}
            className="px-4 py-2 text-sm font-medium border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
          >
            {t('automationPolicies.simulate')}
          </button>
          <button
            onClick={openCreate}
            className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            {t('automationPolicies.newPolicy')}
          </button>
        </div>
      </div>

      {loading && <p className="text-sm text-gray-500 dark:text-gray-400">{t('common.loading')}</p>}
      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      {!loading && !error && (
        <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800">
              <tr>
                {[t('common.name'), t('automationPolicies.workflow'), t('auditLog.action'), t('automationPolicies.effect'), t('automationPolicies.conditions'), t('dnsTab.enabled'), ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {policies.length === 0 ? (
                <tr><td colSpan={7} className="px-4 py-6 text-center text-gray-400">{t('automationPolicies.noPoliciesConfigured')}</td></tr>
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
                      ? <span className="text-green-600 dark:text-green-400">{t('common.yes')}</span>
                      : <span className="text-gray-400">{t('common.no')}</span>}
                  </td>
                  <td className="px-4 py-3 text-right space-x-2 whitespace-nowrap">
                    <button onClick={() => openEdit(p)} className="text-blue-600 hover:text-blue-800 dark:text-blue-400 text-xs">{t('common.edit')}</button>
                    <button onClick={() => handleDelete(p.id)} className="text-red-600 hover:text-red-800 dark:text-red-400 text-xs">{t('common.delete')}</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showSimulate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{t('automationPolicies.simulatePolicyTitle')}</h2>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{t('automationPolicies.simulateReadOnlyHint')}</p>
              </div>
              <button onClick={() => setShowSimulate(false)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-xl leading-none">×</button>
            </div>
            <form onSubmit={handleSimulate} className="px-6 py-4 space-y-4">
              {simError && <p className="text-sm text-red-600 dark:text-red-400">{simError}</p>}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('automationPolicies.workflow')}</label>
                  <input
                    className="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder={t('automationPolicies.workflowPlaceholder')}
                    value={simForm.workflow}
                    onChange={e => setSimForm(f => ({ ...f, workflow: e.target.value }))}
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('auditLog.action')}</label>
                  <input
                    className="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder={t('automationPolicies.actionPlaceholder')}
                    value={simForm.action}
                    onChange={e => setSimForm(f => ({ ...f, action: e.target.value }))}
                    required
                  />
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">{t('automationPolicies.contextValuesLabel')}</label>
                  <button
                    type="button"
                    onClick={() => setSimForm(f => ({ ...f, rows: [...f.rows, { ...EMPTY_SIM_ROW }] }))}
                    className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400 font-medium"
                  >
                    {t('automationPolicies.addValue')}
                  </button>
                </div>
                {simForm.rows.length === 0 ? (
                  <p className="text-xs text-gray-400 dark:text-gray-500 italic">{t('automationPolicies.noContextValues')}</p>
                ) : (
                  <div className="space-y-2">
                    {simForm.rows.map((row, i) => (
                      <div key={i} className="flex items-center gap-2">
                        <input
                          className="flex-1 px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder={t('automationPolicies.keyPlaceholder')}
                          value={row.key}
                          onChange={e => setSimForm(f => ({ ...f, rows: f.rows.map((r, j) => j === i ? { ...r, key: e.target.value } : r) }))}
                        />
                        <input
                          className="flex-1 px-2 py-1.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder={t('adminCustomFields.valuePlaceholder')}
                          value={row.value}
                          onChange={e => setSimForm(f => ({ ...f, rows: f.rows.map((r, j) => j === i ? { ...r, value: e.target.value } : r) }))}
                        />
                        <button
                          type="button"
                          onClick={() => setSimForm(f => ({ ...f, rows: f.rows.filter((_, j) => j !== i) }))}
                          className="text-gray-400 hover:text-red-600 dark:hover:text-red-400 text-lg leading-none px-1"
                        >×</button>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {simResult && (
                <div className={`rounded-lg border p-4 space-y-3 ${decisionColors(simResult.effectiveDecision)}`}>
                  <div className="flex items-center gap-3">
                    <span className="text-base font-semibold capitalize">{simResult.effectiveDecision.replace('_', ' ')}</span>
                    {simResult.reviewNeeded && <span className="text-xs font-medium px-2 py-0.5 rounded bg-yellow-200 dark:bg-yellow-800 text-yellow-900 dark:text-yellow-100">{t('automationPolicies.reviewRequired')}</span>}
                    {simResult.allowed && !simResult.reviewNeeded && <span className="text-xs font-medium px-2 py-0.5 rounded bg-green-200 dark:bg-green-800 text-green-900 dark:text-green-100">{t('automationPolicies.allowedBadge')}</span>}
                    {!simResult.allowed && <span className="text-xs font-medium px-2 py-0.5 rounded bg-red-200 dark:bg-red-800 text-red-900 dark:text-red-100">{t('automationPolicies.blockedBadge')}</span>}
                  </div>
                  {simResult.matchedPolicies && simResult.matchedPolicies.length > 0 ? (
                    <div className="space-y-2">
                      <p className="text-xs font-medium opacity-70">{t('automationPolicies.policiesMatched', { count: simResult.matchedPolicies.length })}</p>
                      {simResult.matchedPolicies.map((mp, i) => (
                        <div key={i} className="rounded bg-white/50 dark:bg-black/20 px-3 py-2 text-xs space-y-1">
                          <div className="flex items-center gap-2">
                            <span className="font-semibold">{mp.name}</span>
                            <span className={`px-1.5 py-0.5 rounded font-medium ${effectBadge(mp.effect)}`}>{mp.effect}</span>
                          </div>
                          {mp.actionsWouldRun && mp.actionsWouldRun.length > 0 && (
                            <ul className="list-disc list-inside opacity-80 space-y-0.5">
                              {mp.actionsWouldRun.map((a, j) => <li key={j}>{a}</li>)}
                            </ul>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-xs opacity-70">{t('automationPolicies.noPoliciesMatched')}</p>
                  )}
                </div>
              )}

              <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={() => setShowSimulate(false)} className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded">{t('myRequests.close')}</button>
                <button type="submit" disabled={simRunning} className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors">
                  {simRunning ? t('automationPolicies.running') : t('automationPolicies.runSimulation')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                {editingID ? t('automationPolicies.editPolicyTitle') : t('automationPolicies.newPolicy')}
              </h2>
            </div>
            <form onSubmit={handleSave} className="px-6 py-4 space-y-4">
              {formError && <p className="text-sm text-red-600 dark:text-red-400">{formError}</p>}

              {[
                { label: t('common.name'), key: 'name', required: true },
                { label: t('automationPolicies.workflowFieldLabel'), key: 'workflow' },
                { label: t('automationPolicies.actionFieldLabel'), key: 'action' },
                { label: t('automationPolicies.messageFieldLabel'), key: 'message' },
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
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('automationPolicies.effect')}</label>
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
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">{t('automationPolicies.conditions')}</label>
                  <button
                    type="button"
                    onClick={addCondition}
                    className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400 font-medium"
                  >
                    {t('automationPolicies.addCondition')}
                  </button>
                </div>
                {form.conditions.length === 0 ? (
                  <p className="text-xs text-gray-400 dark:text-gray-500 italic">{t('automationPolicies.noConditionsHint')}</p>
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

              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">{t('automationPolicies.actionsLabel')} <span className="font-normal text-gray-400">{t('automationPolicies.actionsHintSuffix')}</span></label>
                  <button
                    type="button"
                    onClick={addAction}
                    className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400 font-medium"
                  >
                    {t('automationPolicies.addAction')}
                  </button>
                </div>
                {(form.actions || []).length === 0 ? (
                  <p className="text-xs text-gray-400 dark:text-gray-500 italic">{t('automationPolicies.noActionsHint')}</p>
                ) : (
                  <div className="space-y-2">
                    {(form.actions || []).map((action, i) => (
                      <ActionRow
                        key={i}
                        action={action}
                        index={i}
                        onChange={updateAction}
                        onRemove={removeAction}
                      />
                    ))}
                  </div>
                )}
              </div>

              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={form.enabled} onChange={e => setForm(f => ({ ...f, enabled: e.target.checked }))} />
                <span className="text-sm text-gray-700 dark:text-gray-300">{t('dnsTab.enabled')}</span>
              </label>

              <div>
                <button
                  type="button"
                  onClick={() => setShowPreview(p => !p)}
                  className="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 underline"
                >
                  {showPreview ? t('automationPolicies.hideJson') : t('automationPolicies.previewJson')}{t('automationPolicies.jsonSuffix')}
                </button>
                {showPreview && (
                  <pre className="mt-2 p-3 rounded bg-gray-50 dark:bg-gray-800 text-xs text-gray-700 dark:text-gray-300 overflow-x-auto border border-gray-200 dark:border-gray-700">
                    {JSON.stringify(previewPayload, null, 2)}
                  </pre>
                )}
              </div>

              <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={() => setShowModal(false)} className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded">{t('common.cancel')}</button>
                <button type="submit" disabled={saving} className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors">
                  {saving ? t('common.saving') : t('common.save')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
