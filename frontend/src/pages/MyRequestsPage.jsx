import { useState, useEffect } from 'react'
import {
  getMySubnetRequests,
  getMyIPRequests,
  cancelSubnetRequest,
  cancelIPRequest,
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

export default function MyRequestsPage() {
  const [subnetRequests, setSubnetRequests] = useState([])
  const [ipRequests, setIPRequests] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [cancelConfirm, setCancelConfirm] = useState(null) // { type, id }
  const [detailModal, setDetailModal] = useState(null) // { request, requestType }
  const [reRequestModal, setReRequestModal] = useState(null) // { requestType, prefilled }

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      setError(null)
      const [subRes, ipRes] = await Promise.all([
        getMySubnetRequests(),
        getMyIPRequests(),
      ])
      setSubnetRequests(Array.isArray(subRes.data) ? subRes.data : (subRes.data?.requests ?? []))
      setIPRequests(Array.isArray(ipRes.data) ? ipRes.data : (ipRes.data?.requests ?? []))
    } catch {
      setError('Failed to load requests')
    } finally {
      setLoading(false)
    }
  }

  async function handleCancel({ type, id }) {
    try {
      if (type === 'subnets') await cancelSubnetRequest(id)
      else await cancelIPRequest(id)
      setCancelConfirm(null)
      load()
    } catch {
      setError('Failed to cancel request')
    }
  }

  const allRequests = [
    ...subnetRequests.map(r => ({ ...r, _type: 'subnets' })),
    ...ipRequests.map(r => ({ ...r, _type: 'ips' })),
  ].sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">My Requests</h1>
        <button
          onClick={load}
          className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
        >
          Refresh
        </button>
      </div>

      {error && <p className="mb-4 text-red-600 text-sm">{error}</p>}

      {loading ? (
        <p className="text-gray-500">Loading requests...</p>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Type</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Target</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Purpose</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Submitted</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Status</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Reviewer</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {allRequests.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-6 text-center text-gray-400">
                    You haven&apos;t made any requests yet.
                  </td>
                </tr>
              )}
              {allRequests.map(r => (
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
                  <td className="px-4 py-3 font-mono text-gray-600 dark:text-gray-300">
                    {r._type === 'subnets'
                      ? (r.prefixLength ? `/${r.prefixLength}` : '—')
                      : (r.specificIp || 'auto-assign')}
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 max-w-xs truncate">{r.purpose || '—'}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs whitespace-nowrap">
                    {formatDate(r.createdAt)}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={r.status} />
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    <div>
                      <span>{r.reviewerUsername || r.reviewer || '—'}</span>
                      {r.reviewerNote && (
                        <p className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{r.reviewerNote}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-right space-x-2" onClick={e => e.stopPropagation()}>
                    {r.status === 'pending' && (
                      cancelConfirm?.type === r._type && cancelConfirm?.id === r.id ? (
                        <>
                          <span className="text-red-600 text-xs">Cancel?</span>
                          <button
                            onClick={() => handleCancel({ type: r._type, id: r.id })}
                            className="text-red-600 hover:text-red-800 text-xs font-medium"
                          >
                            Yes
                          </button>
                          <button
                            onClick={() => setCancelConfirm(null)}
                            className="text-gray-400 hover:text-gray-600 text-xs"
                          >
                            No
                          </button>
                        </>
                      ) : (
                        <button
                          onClick={() => setCancelConfirm({ type: r._type, id: r.id })}
                          className="text-gray-400 hover:text-red-600 text-xs"
                        >
                          Cancel
                        </button>
                      )
                    )}
                    {r.status === 'rejected' && (
                      <button
                        onClick={() => setReRequestModal({ requestType: r._type, prefilled: r })}
                        className="text-blue-600 hover:text-blue-800 text-xs font-medium"
                      >
                        Re-request
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
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
              {detailModal.requestType === 'subnets' ? (
                <>
                  <span className="font-medium text-gray-600 dark:text-gray-400">Prefix Length</span>
                  <span className="font-mono text-gray-800 dark:text-gray-200">/{detailModal.request.prefixLength || '—'}</span>
                </>
              ) : (
                <>
                  <span className="font-medium text-gray-600 dark:text-gray-400">Specific IP</span>
                  <span className="font-mono text-gray-800 dark:text-gray-200">{detailModal.request.specificIp || 'auto-assign'}</span>
                  {detailModal.request.dnsName && (
                    <>
                      <span className="font-medium text-gray-600 dark:text-gray-400">DNS Name</span>
                      <span className="text-gray-800 dark:text-gray-200">{detailModal.request.dnsName}</span>
                    </>
                  )}
                </>
              )}
              <span className="font-medium text-gray-600 dark:text-gray-400">Purpose</span>
              <span className="text-gray-800 dark:text-gray-200">{detailModal.request.purpose || '—'}</span>
              <span className="font-medium text-gray-600 dark:text-gray-400">Submitted</span>
              <span className="text-gray-500 dark:text-gray-400">{formatDate(detailModal.request.createdAt)}</span>
              {detailModal.request.reviewerUsername && (
                <>
                  <span className="font-medium text-gray-600 dark:text-gray-400">Reviewer</span>
                  <span className="text-gray-800 dark:text-gray-200">{detailModal.request.reviewerUsername}</span>
                </>
              )}
              {detailModal.request.reviewerNote && (
                <>
                  <span className="font-medium text-gray-600 dark:text-gray-400">Reviewer Note</span>
                  <span className="text-gray-800 dark:text-gray-200">{detailModal.request.reviewerNote}</span>
                </>
              )}
            </div>

            {detailModal.request.status === 'rejected' && (
              <button
                onClick={() => { setDetailModal(null); setReRequestModal({ requestType: detailModal.requestType, prefilled: detailModal.request }) }}
                className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                Re-request
              </button>
            )}

            <RequestComments type={detailModal.requestType} id={detailModal.request.id} />
          </div>
        </Modal>
      )}

      {/* Re-request modal — just shows info for now; actual submit goes through NetworksPage / IPAddressesPage forms */}
      {reRequestModal && (
        <Modal
          title={`Re-request ${reRequestModal.requestType === 'subnets' ? 'Subnet' : 'IP'}`}
          onClose={() => setReRequestModal(null)}
        >
          <div className="space-y-3 text-sm">
            <p className="text-gray-600 dark:text-gray-400">
              To re-request, please use the{' '}
              <strong>{reRequestModal.requestType === 'subnets' ? '"Request Subnet"' : '"Request IP"'}</strong>{' '}
              button on the {reRequestModal.requestType === 'subnets' ? 'Networks' : 'IP Addresses'} page.
            </p>
            <div className="bg-gray-50 dark:bg-gray-700 rounded p-3 space-y-1">
              <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Previous values</p>
              {reRequestModal.requestType === 'subnets' ? (
                <>
                  <p>Prefix length: <span className="font-mono">/{reRequestModal.prefilled.prefixLength || '—'}</span></p>
                </>
              ) : (
                <>
                  <p>Specific IP: <span className="font-mono">{reRequestModal.prefilled.specificIp || 'auto-assign'}</span></p>
                  {reRequestModal.prefilled.dnsName && (
                    <p>DNS Name: {reRequestModal.prefilled.dnsName}</p>
                  )}
                </>
              )}
              <p>Purpose: {reRequestModal.prefilled.purpose}</p>
            </div>
            <div className="flex justify-end">
              <button onClick={() => setReRequestModal(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">
                Close
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
