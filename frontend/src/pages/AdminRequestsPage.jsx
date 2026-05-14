import { useState, useEffect, useCallback } from 'react'
import {
  adminGetSubnetRequests,
  adminGetIPRequests,
  adminApproveSubnetRequest,
  adminRejectSubnetRequest,
  adminApproveIPRequest,
  adminRejectIPRequest,
} from '../api/requests'
import Modal from '../components/Modal'
import RequestComments from '../components/RequestComments'

const STATUS_COLORS = {
  pending: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
  approved: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  rejected: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  cancelled: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400',
}

function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

function StatusBadge({ status }) {
  return (
    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[status] || STATUS_COLORS.pending}`}>
      {status}
    </span>
  )
}

export default function AdminRequestsPage() {
  const [tab, setTab] = useState('all') // 'all' | 'subnets' | 'ips'
  const [subnetRequests, setSubnetRequests] = useState([])
  const [ipRequests, setIPRequests] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [statusFilter, setStatusFilter] = useState('pending')
  const [searchRequester, setSearchRequester] = useState('')
  const [actionModal, setActionModal] = useState(null) // { type: 'approve'|'reject', requestType: 'subnets'|'ips', id }
  const [reviewerNote, setReviewerNote] = useState('')
  const [actionSaving, setActionSaving] = useState(false)
  const [actionError, setActionError] = useState(null)
  const [detailModal, setDetailModal] = useState(null) // { request, requestType }

  const load = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const params = {}
      if (statusFilter) params.status = statusFilter
      if (searchRequester) params.requester = searchRequester
      const [subRes, ipRes] = await Promise.all([
        adminGetSubnetRequests(params),
        adminGetIPRequests(params),
      ])
      setSubnetRequests(Array.isArray(subRes.data) ? subRes.data : (subRes.data?.requests ?? []))
      setIPRequests(Array.isArray(ipRes.data) ? ipRes.data : (ipRes.data?.requests ?? []))
    } catch {
      setError('Failed to load requests')
    } finally {
      setLoading(false)
    }
  }, [statusFilter, searchRequester])

  useEffect(() => { load() }, [load])

  function openAction(type, requestType, id) {
    setReviewerNote('')
    setActionError(null)
    setActionModal({ type, requestType, id })
  }

  async function handleAction(e) {
    e.preventDefault()
    setActionSaving(true)
    setActionError(null)
    try {
      const { type, requestType, id } = actionModal
      if (requestType === 'subnets') {
        if (type === 'approve') await adminApproveSubnetRequest(id, reviewerNote)
        else await adminRejectSubnetRequest(id, reviewerNote)
      } else {
        if (type === 'approve') await adminApproveIPRequest(id, reviewerNote)
        else await adminRejectIPRequest(id, reviewerNote)
      }
      setActionModal(null)
      load()
    } catch (err) {
      setActionError(err.response?.data?.error || 'Action failed')
    } finally {
      setActionSaving(false)
    }
  }

  const allRequests = [
    ...subnetRequests.map(r => ({ ...r, _type: 'subnets' })),
    ...ipRequests.map(r => ({ ...r, _type: 'ips' })),
  ].sort((a, b) => new Date(b.createdAt || b.created_at) - new Date(a.createdAt || a.created_at))

  function getDisplayRequests() {
    if (tab === 'subnets') return subnetRequests.map(r => ({ ...r, _type: 'subnets' }))
    if (tab === 'ips') return ipRequests.map(r => ({ ...r, _type: 'ips' }))
    return allRequests
  }

  const displayRequests = getDisplayRequests()

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Request Management</h1>
        <button
          onClick={load}
          className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
        >
          Refresh
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      {/* Filter bar */}
      <div className="mb-4 flex flex-wrap gap-3 items-center">
        <select
          value={statusFilter}
          onChange={e => setStatusFilter(e.target.value)}
          className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
        >
          <option value="">All Statuses</option>
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <input
          type="text"
          placeholder="Search by requester..."
          value={searchRequester}
          onChange={e => setSearchRequester(e.target.value)}
          className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
        />
      </div>

      {/* Tabs */}
      <div className="flex border-b dark:border-gray-700 mb-4">
        {[
          { key: 'all', label: `All (${allRequests.length})` },
          { key: 'subnets', label: `Subnet Requests (${subnetRequests.length})` },
          { key: 'ips', label: `IP Requests (${ipRequests.length})` },
        ].map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              tab === t.key
                ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading ? (
        <p className="text-gray-500">Loading requests...</p>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Requester</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Target</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Purpose</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Submitted</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {displayRequests.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-6 text-center text-gray-400">
                    No requests found
                  </td>
                </tr>
              )}
              {displayRequests.map(r => (
                <tr
                  key={`${r._type}-${r.id}`}
                  className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30 cursor-pointer"
                  onClick={() => setDetailModal({ request: r, requestType: r._type })}
                >
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                      r._type === 'subnets'
                        ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                        : 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300'
                    }`}>
                      {r._type === 'subnets' ? 'Subnet' : 'IP'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-800 dark:text-gray-200">{r.requesterUsername || r.username || '—'}</td>
                  <td className="px-4 py-3 font-mono text-gray-600 dark:text-gray-300">
                    {r._type === 'subnets'
                      ? (r.prefixLength ? `/${r.prefixLength}` : '—')
                      : (r.specificIp || r.specific_ip || 'auto-assign')}
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 max-w-xs truncate">{r.purpose || '—'}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs whitespace-nowrap">
                    {formatDate(r.createdAt || r.created_at)}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={r.status} />
                  </td>
                  <td className="px-4 py-3 text-right space-x-2" onClick={e => e.stopPropagation()}>
                    {r.status === 'pending' && (
                      <>
                        <button
                          onClick={() => openAction('approve', r._type, r.id)}
                          className="px-2 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700"
                        >
                          Approve
                        </button>
                        <button
                          onClick={() => openAction('reject', r._type, r.id)}
                          className="px-2 py-1 text-xs bg-red-600 text-white rounded hover:bg-red-700"
                        >
                          Reject
                        </button>
                      </>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Approve/Reject modal */}
      {actionModal && (
        <Modal
          title={`${actionModal.type === 'approve' ? 'Approve' : 'Reject'} Request`}
          onClose={() => setActionModal(null)}
        >
          <form onSubmit={handleAction} className="space-y-4">
            {actionError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{actionError}</div>
            )}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Reviewer Note <span className="text-gray-400 font-normal">(optional)</span>
              </label>
              <textarea
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
                rows={3}
                placeholder="Optional note for the requester..."
                value={reviewerNote}
                onChange={e => setReviewerNote(e.target.value)}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setActionModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">
                Cancel
              </button>
              <button
                type="submit"
                disabled={actionSaving}
                className={`px-4 py-2 text-white rounded text-sm disabled:opacity-50 ${
                  actionModal.type === 'approve' ? 'bg-green-600 hover:bg-green-700' : 'bg-red-600 hover:bg-red-700'
                }`}
              >
                {actionSaving ? 'Saving...' : (actionModal.type === 'approve' ? 'Approve' : 'Reject')}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* Detail modal */}
      {detailModal && (
        <Modal
          title={`${detailModal.requestType === 'subnets' ? 'Subnet' : 'IP'} Request Details`}
          onClose={() => setDetailModal(null)}
        >
          <div className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-2">
              <span className="font-medium text-gray-600 dark:text-gray-400">Status</span>
              <StatusBadge status={detailModal.request.status} />
              <span className="font-medium text-gray-600 dark:text-gray-400">Requester</span>
              <span className="text-gray-800 dark:text-gray-200">{detailModal.request.requesterUsername || detailModal.request.username || '—'}</span>
              {detailModal.requestType === 'subnets' ? (
                <>
                  <span className="font-medium text-gray-600 dark:text-gray-400">Prefix Length</span>
                  <span className="font-mono text-gray-800 dark:text-gray-200">/{detailModal.request.prefixLength || '—'}</span>
                </>
              ) : (
                <>
                  <span className="font-medium text-gray-600 dark:text-gray-400">Specific IP</span>
                  <span className="font-mono text-gray-800 dark:text-gray-200">{detailModal.request.specificIp || detailModal.request.specific_ip || 'auto-assign'}</span>
                  {(detailModal.request.dnsName || detailModal.request.dns_name) && (
                    <>
                      <span className="font-medium text-gray-600 dark:text-gray-400">DNS Name</span>
                      <span className="text-gray-800 dark:text-gray-200">{detailModal.request.dnsName || detailModal.request.dns_name}</span>
                    </>
                  )}
                </>
              )}
              <span className="font-medium text-gray-600 dark:text-gray-400">Purpose</span>
              <span className="text-gray-800 dark:text-gray-200">{detailModal.request.purpose || '—'}</span>
              <span className="font-medium text-gray-600 dark:text-gray-400">Submitted</span>
              <span className="text-gray-500 dark:text-gray-400">{formatDate(detailModal.request.createdAt || detailModal.request.created_at)}</span>
              {detailModal.request.reviewerNote && (
                <>
                  <span className="font-medium text-gray-600 dark:text-gray-400">Reviewer Note</span>
                  <span className="text-gray-800 dark:text-gray-200">{detailModal.request.reviewerNote}</span>
                </>
              )}
            </div>

            {detailModal.request.status === 'pending' && (
              <div className="flex gap-2 pt-2">
                <button
                  onClick={() => { setDetailModal(null); openAction('approve', detailModal.requestType, detailModal.request.id) }}
                  className="px-3 py-1.5 text-xs bg-green-600 text-white rounded hover:bg-green-700"
                >
                  Approve
                </button>
                <button
                  onClick={() => { setDetailModal(null); openAction('reject', detailModal.requestType, detailModal.request.id) }}
                  className="px-3 py-1.5 text-xs bg-red-600 text-white rounded hover:bg-red-700"
                >
                  Reject
                </button>
              </div>
            )}

            <RequestComments type={detailModal.requestType} id={detailModal.request.id} />
          </div>
        </Modal>
      )}
    </div>
  )
}
