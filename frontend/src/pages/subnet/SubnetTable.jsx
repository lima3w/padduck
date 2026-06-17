import { Link } from 'react-router-dom'
import SortTh from '../../components/SortTh'
import EmptyRow from '../../components/EmptyRow'
import Pagination from '../../components/Pagination'

export default function SubnetTable({
  subnets, total, isSearchActive, page, defaultLimit,
  sortCol, sortDir, onSort, sortedSubnets,
  searchableFields, locations, vlans, isAdmin,
  deleteConfirm, onDeleteConfirm, onDeleteCancel, onDelete,
  onEdit, onSplit, onMerge, onResize,
  onPageChange, onNavigate,
  addCfFilterFromValue,
}) {
  return (
    <>
      {!isSearchActive && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">
          {total} subnet{total !== 1 ? 's' : ''}
        </p>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
              <tr>
                <SortTh label="Network" col="network" sortCol={sortCol} sortDir={sortDir} onSort={onSort} />
                <SortTh label="Prefix" col="prefix" sortCol={sortCol} sortDir={sortDir} onSort={onSort} />
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">Location</th>
                <th className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">VLAN</th>
                <SortTh label="Description" col="description" sortCol={sortCol} sortDir={sortDir} onSort={onSort} />
                {searchableFields.map(d => (
                  <th key={d.name} className="text-left px-4 py-3 text-gray-600 dark:text-gray-300 font-medium">{d.label}</th>
                ))}
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {subnets.length === 0 && (
                <EmptyRow colSpan={6 + searchableFields.length} message="No subnets yet." />
              )}
              {sortedSubnets(subnets).map(s => (
                <tr key={s.id} className="border-b dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td
                    className="px-4 py-3 font-mono font-medium text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"
                    onClick={() => onNavigate(`/subnets/${s.id}/ip-addresses`)}
                  >
                    {s.networkAddress}
                  </td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">/{s.prefixLength}</td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    {s.locationId ? (
                      <Link to={`/locations/${s.locationId}`} className="text-blue-600 dark:text-blue-400 hover:underline text-xs">
                        {locations.find(l => l.id === s.locationId)?.name || `#${s.locationId}`}
                      </Link>
                    ) : '—'}
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">
                    {s.vlanId != null ? (
                      <Link to={`/vlans/${s.vlanId}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                        VLAN {vlans.find(v => v.id === s.vlanId)?.vlanId ?? `#${s.vlanId}`}
                      </Link>
                    ) : '—'}
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400">
                    {s.isContainer ? <span className="text-gray-400 italic text-xs">Container subnet</span> : s.description}
                  </td>
                  {searchableFields.map(d => {
                    const val = s.customFields?.[d.name]
                    return (
                      <td key={d.name} className="px-4 py-3 text-gray-500 dark:text-gray-400">
                        {val ? (
                          <button
                            className="hover:text-blue-600 dark:hover:text-blue-400 underline decoration-dotted text-left"
                            onClick={() => addCfFilterFromValue(d.name, val)}
                            title="Filter by this value"
                          >
                            {val}
                          </button>
                        ) : '—'}
                      </td>
                    )
                  })}
                  <td className="px-4 py-3 text-right space-x-2">
                    <button onClick={() => onEdit(s)} className="text-gray-400 hover:text-blue-600 text-xs">Edit</button>
                    {isAdmin && (
                      <>
                        <button onClick={() => onSplit(s)} className="text-gray-400 hover:text-purple-600 text-xs">Split</button>
                        <button onClick={() => onMerge(s)} className="text-gray-400 hover:text-indigo-600 text-xs">Merge</button>
                        <button onClick={() => onResize(s)} className="text-gray-400 hover:text-teal-600 text-xs">Resize</button>
                      </>
                    )}
                    {deleteConfirm === s.id ? (
                      <>
                        <span className="text-red-600 text-xs">Confirm?</span>
                        <button onClick={() => onDelete(s.id)} className="text-red-600 hover:text-red-800 text-xs font-medium">Yes</button>
                        <button onClick={onDeleteCancel} className="text-gray-400 hover:text-gray-600 text-xs">No</button>
                      </>
                    ) : (
                      <button onClick={() => onDeleteConfirm(s.id)} className="text-gray-400 hover:text-red-600 text-xs">Delete</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {!isSearchActive && total > defaultLimit && (
        <Pagination page={page} limit={defaultLimit} total={total} onChange={onPageChange} />
      )}
    </>
  )
}
