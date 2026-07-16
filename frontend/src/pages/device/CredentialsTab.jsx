import { useTranslation } from 'react-i18next'

export default function CredentialsTab({ snmpCreds, snmpLoading, snmpError, snmpRevealed, onLoadCredentials, onClearCredentials, onToggleReveal }) {
  const { t } = useTranslation()
  return (
    <div className="max-w-lg space-y-4">
      <div>
        <h2 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">{t('credentials.title')}</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {t('credentials.subtitle')}
        </p>
      </div>

      {snmpError && <p className="text-sm text-red-600">{snmpError}</p>}

      {snmpCreds === null && !snmpLoading && (
        <button
          onClick={onLoadCredentials}
          className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        >
          {t('credentials.loadCredentials')}
        </button>
      )}

      {snmpLoading && <p className="text-sm text-gray-500">{t('common.loading')}</p>}

      {snmpCreds === false && (
        <p className="text-sm text-gray-500">{t('credentials.noCredsStored')}</p>
      )}

      {snmpCreds && snmpCreds !== false && (() => {
        const rows = [
          { key: 'snmpVersion', label: t('credentials.snmpVersion'), value: snmpCreds.snmpVersion, sensitive: false },
          { key: 'snmpCommunity', label: t('credentials.communityString'), value: snmpCreds.snmpCommunity, sensitive: true },
          { key: 'snmpV3User', label: t('credentials.snmpV3Username'), value: snmpCreds.snmpV3User, sensitive: false },
          { key: 'snmpV3AuthProto', label: t('credentials.authProtocol'), value: snmpCreds.snmpV3AuthProto, sensitive: false },
          { key: 'snmpV3AuthPass', label: t('credentials.authPassword'), value: snmpCreds.snmpV3AuthPass, sensitive: true },
          { key: 'snmpV3PrivProto', label: t('credentials.privProtocol'), value: snmpCreds.snmpV3PrivProto, sensitive: false },
          { key: 'snmpV3PrivPass', label: t('credentials.privPassword'), value: snmpCreds.snmpV3PrivPass, sensitive: true },
        ].filter((r) => r.value != null && r.value !== '')

        return (
          <div className="border border-gray-200 dark:border-gray-700 rounded divide-y divide-gray-100 dark:divide-gray-700">
            {rows.map(({ key, label, value, sensitive }) => (
              <div key={key} className="flex items-center justify-between px-4 py-3 gap-4">
                <span className="text-sm text-gray-600 dark:text-gray-400 w-36 flex-shrink-0">{label}</span>
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  {!sensitive || snmpRevealed[key] ? (
                    <span className="text-sm font-mono text-gray-900 dark:text-gray-100 break-all">{value}</span>
                  ) : (
                    <span className="text-sm text-gray-400 font-mono">••••••••</span>
                  )}
                  {sensitive && (
                    <button
                      onClick={() => onToggleReveal(key)}
                      className="flex-shrink-0 text-xs text-blue-600 hover:underline"
                    >
                      {snmpRevealed[key] ? t('credentials.hide') : t('credentials.reveal')}
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )
      })()}

      {snmpCreds && (
        <button
          onClick={onClearCredentials}
          className="text-xs text-gray-400 hover:text-gray-600 hover:underline"
        >
          {t('credentials.clearFromView')}
        </button>
      )}
    </div>
  )
}
