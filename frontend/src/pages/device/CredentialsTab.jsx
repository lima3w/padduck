export default function CredentialsTab({ snmpCreds, snmpLoading, snmpError, snmpRevealed, onLoadCredentials, onClearCredentials, onToggleReveal }) {
  return (
    <div className="max-w-lg space-y-4">
      <div>
        <h2 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">SNMP Credentials</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Stored credentials are revealed on demand. Each field must be individually shown.
        </p>
      </div>

      {snmpError && <p className="text-sm text-red-600">{snmpError}</p>}

      {snmpCreds === null && !snmpLoading && (
        <button
          onClick={onLoadCredentials}
          className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        >
          Load Credentials
        </button>
      )}

      {snmpLoading && <p className="text-sm text-gray-500">Loading…</p>}

      {snmpCreds === false && (
        <p className="text-sm text-gray-500">No SNMP credentials stored for this device.</p>
      )}

      {snmpCreds && snmpCreds !== false && (() => {
        const rows = [
          { key: 'snmpVersion', label: 'SNMP Version', value: snmpCreds.snmpVersion, sensitive: false },
          { key: 'snmpCommunity', label: 'Community String', value: snmpCreds.snmpCommunity, sensitive: true },
          { key: 'snmpV3User', label: 'SNMPv3 Username', value: snmpCreds.snmpV3User, sensitive: false },
          { key: 'snmpV3AuthProto', label: 'Auth Protocol', value: snmpCreds.snmpV3AuthProto, sensitive: false },
          { key: 'snmpV3AuthPass', label: 'Auth Password', value: snmpCreds.snmpV3AuthPass, sensitive: true },
          { key: 'snmpV3PrivProto', label: 'Priv Protocol', value: snmpCreds.snmpV3PrivProto, sensitive: false },
          { key: 'snmpV3PrivPass', label: 'Priv Password', value: snmpCreds.snmpV3PrivPass, sensitive: true },
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
                      {snmpRevealed[key] ? 'Hide' : 'Reveal'}
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
          Clear credentials from view
        </button>
      )}
    </div>
  )
}
