import { useState } from 'react'
import TableActions from './TableActions'

function UtilBar({ pct }) {
  const colour =
    pct > 90 ? 'bg-red-500' :
    pct > 70 ? 'bg-yellow-500' :
    'bg-green-500'
  return (
    <div className="flex items-center gap-1 w-28">
      <div className="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
        <div
          className={`${colour} h-1.5 rounded-full`}
          style={{ width: `${Math.min(pct, 100)}%` }}
        />
      </div>
      <span className="text-xs text-gray-400 w-8 text-right">{pct.toFixed(0)}%</span>
    </div>
  )
}

function SubnetRow({ node, depth, onEdit, onDelete, deleteConfirm, setDeleteConfirm, navigate }) {
  const [expanded, setExpanded] = useState(depth === 0)
  const hasChildren = node.children && node.children.length > 0
  const indent = depth * 20

  return (
    <>
      <tr className="border-b last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700/30">
        <td className="px-4 py-2.5">
          <div className="flex items-center gap-1" style={{ paddingLeft: indent }}>
            {hasChildren ? (
              <button
                onClick={() => setExpanded(v => !v)}
                className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 w-4 h-4 flex items-center justify-center flex-shrink-0"
                aria-label={expanded ? 'Collapse' : 'Expand'}
              >
                {expanded ? '▼' : '▶'}
              </button>
            ) : (
              <span className="w-4 flex-shrink-0" />
            )}
            <span
              className="font-mono font-medium text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"
              onClick={() => navigate && navigate(`/subnets/${node.id}/ip-addresses`)}
            >
              {node.cidr}
            </span>
          </div>
        </td>
        <td className="px-4 py-2.5 text-sm text-gray-500 dark:text-gray-400">
          {node.description || '—'}
        </td>
        <td className="px-4 py-2.5 text-sm text-gray-500 dark:text-gray-400 whitespace-nowrap">
          {node.used}/{node.total}
        </td>
        <td className="px-4 py-2.5">
          <UtilBar pct={node.utilizationPct} />
        </td>
        <td className="px-4 py-2.5 text-right">
          <TableActions
            onEdit={onEdit ? () => onEdit(node) : undefined}
            onDelete={onDelete ? () => onDelete(node.id) : undefined}
            confirming={deleteConfirm === node.id}
            onRequestDelete={onDelete ? () => setDeleteConfirm(node.id) : undefined}
            onCancelDelete={() => setDeleteConfirm(null)}
          />
        </td>
      </tr>
      {hasChildren && expanded && node.children.map(child => (
        <SubnetRow
          key={child.id}
          node={child}
          depth={depth + 1}
          onEdit={onEdit}
          onDelete={onDelete}
          deleteConfirm={deleteConfirm}
          setDeleteConfirm={setDeleteConfirm}
          navigate={navigate}
        />
      ))}
    </>
  )
}

/**
 * SubnetTree renders a collapsible indented table of subnets.
 * Props:
 *   nodes       {Array}    — tree nodes from GET /networks/:id/subnets/tree
 *   onEdit      {Function} — called with subnet node to edit
 *   onDelete    {Function} — called with subnet id to delete
 *   navigate    {Function} — react-router navigate fn
 */
export default function SubnetTree({ nodes, onEdit, onDelete, navigate }) {
  const [deleteConfirm, setDeleteConfirm] = useState(null)

  if (!nodes || nodes.length === 0) {
    return (
      <tr>
        <td colSpan={5} className="px-4 py-6 text-center text-gray-400">
          No subnets yet
        </td>
      </tr>
    )
  }

  return (
    <>
      {nodes.map(node => (
        <SubnetRow
          key={node.id}
          node={node}
          depth={0}
          onEdit={onEdit}
          onDelete={onDelete}
          deleteConfirm={deleteConfirm}
          setDeleteConfirm={setDeleteConfirm}
          navigate={navigate}
        />
      ))}
    </>
  )
}
