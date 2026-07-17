import { useTranslation } from 'react-i18next'

export default function ApprovalsTab({ approvals, handleApprove, handleReject }) {
  const { t } = useTranslation()
  return (
        <div>
          {approvals.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {t('approvals.noPendingApprovals')}
            </div>
          ) : (
            <div className="space-y-3">
              {approvals.map((approval) => (
                <div
                  key={approval.id}
                  className="bg-white border border-gray-200 rounded-lg p-4 flex items-center justify-between"
                >
                  <div>
                    <p className="font-medium text-gray-900">{approval.username}</p>
                    <p className="text-sm text-gray-500">{approval.email}</p>
                    <p className="text-xs text-gray-400">
                      {t('approvals.registeredOn', { date: new Date(approval.createdAt).toLocaleDateString() })}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleApprove(approval.id)}
                      className="bg-green-600 text-white px-4 py-1.5 rounded text-sm hover:bg-green-700 transition font-medium"
                    >
                      {t('approvals.approve')}
                    </button>
                    <button
                      onClick={() => handleReject(approval.id)}
                      className="bg-red-600 text-white px-4 py-1.5 rounded text-sm hover:bg-red-700 transition font-medium"
                    >
                      {t('approvals.reject')}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
  )
}
