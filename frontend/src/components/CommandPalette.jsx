import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { globalSearch } from '../api/ipam'

function buildItems(data, t) {
  const items = []
  for (const s of data.networks || []) {
    items.push({
      type: 'network',
      label: s.name,
      sub: s.description || '',
      url: `/networks/${s.id}/subnets`,
    })
  }
  for (const s of data.subnets || []) {
    items.push({
      type: 'subnet',
      label: `${s.networkAddress}/${s.prefixLength}`,
      sub: s.description || t('commandPalette.networkFallback', { id: s.sectionId }),
      url: `/networks/${s.sectionId}/subnets`,
    })
  }
  for (const d of data.devices || []) {
    items.push({
      type: 'device',
      label: d.hostname,
      sub: d.description || (d.type?.name ?? ''),
      url: `/devices/${d.id}`,
    })
  }
  return items
}

const TYPE_LABEL_KEYS = { network: 'subnets.network', subnet: 'dashboard.subnet', device: 'ipAddressesPage.columnDevice' }
const TYPE_COLOR = {
  network: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  subnet:  'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
  device:  'bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-300',
}

export default function CommandPalette({ open, onClose }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [cursor, setCursor] = useState(0)
  const inputRef = useRef(null)
  const timerRef = useRef(null)

  useEffect(() => {
    if (open) {
      setQuery('')
      setItems([])
      setCursor(0)
      setTimeout(() => inputRef.current?.focus(), 10)
    }
  }, [open])

  const search = useCallback((q) => {
    if (!q.trim()) { setItems([]); setLoading(false); return }
    setLoading(true)
    globalSearch(q)
      .then((res) => {
        setItems(buildItems(res.data, t))
        setCursor(0)
      })
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }, [t])

  function handleInput(e) {
    const q = e.target.value
    setQuery(q)
    clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => search(q), 250)
  }

  function select(item) {
    onClose()
    navigate(item.url)
  }

  function handleKey(e) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setCursor((c) => Math.min(c + 1, items.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setCursor((c) => Math.max(c - 1, 0))
    } else if (e.key === 'Enter' && items[cursor]) {
      select(items[cursor])
    } else if (e.key === 'Escape') {
      onClose()
    }
  }

  if (!open) return null

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t('commandPalette.dialogLabel')}
      className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] bg-black/40 dark:bg-black/60"
      onMouseDown={(e) => { if (e.target === e.currentTarget) onClose() }}
    >
      <div className="w-full max-w-lg bg-white dark:bg-gray-800 rounded-xl shadow-2xl overflow-hidden border border-gray-200 dark:border-gray-700">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <svg className="w-4 h-4 text-gray-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8" strokeWidth="2" />
            <path d="m21 21-4.35-4.35" strokeWidth="2" strokeLinecap="round" />
          </svg>
          <input
            ref={inputRef}
            value={query}
            onChange={handleInput}
            onKeyDown={handleKey}
            aria-label={t('commandPalette.searchAriaLabel')}
            aria-controls="command-palette-results"
            aria-activedescendant={items[cursor] ? `command-palette-item-${cursor}` : undefined}
            role="combobox"
            aria-expanded={items.length > 0}
            aria-autocomplete="list"
            placeholder={t('commandPalette.searchPlaceholder')}
            className="flex-1 bg-transparent text-sm outline-none text-gray-900 dark:text-gray-100 placeholder-gray-400"
          />
          {loading && (
            <svg className="animate-spin w-4 h-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
            </svg>
          )}
          <kbd className="hidden sm:inline-flex items-center px-1.5 py-0.5 text-xs font-medium text-gray-400 border border-gray-300 dark:border-gray-600 rounded">Esc</kbd>
        </div>

        {items.length > 0 && (
          <ul id="command-palette-results" role="listbox" className="max-h-72 overflow-y-auto py-1">
            {items.map((item, i) => (
              <li key={i} id={`command-palette-item-${i}`} role="option" aria-selected={i === cursor}>
                <button
                  type="button"
                  onMouseEnter={() => setCursor(i)}
                  onMouseDown={() => select(item)}
                  className={`w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500 ${
                    i === cursor ? 'bg-blue-50 dark:bg-blue-900/30' : 'hover:bg-gray-50 dark:hover:bg-gray-700/50'
                  }`}
                >
                  <span className={`shrink-0 text-xs font-medium px-1.5 py-0.5 rounded ${TYPE_COLOR[item.type]}`}>
                    {t(TYPE_LABEL_KEYS[item.type])}
                  </span>
                  <span className="flex-1 min-w-0">
                    <span className="block text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{item.label}</span>
                    {item.sub && (
                      <span className="block text-xs text-gray-500 dark:text-gray-400 truncate">{item.sub}</span>
                    )}
                  </span>
                  <svg className="w-3.5 h-3.5 text-gray-300 dark:text-gray-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              </li>
            ))}
          </ul>
        )}

        {query && !loading && items.length === 0 && (
          <p className="px-4 py-6 text-sm text-center text-gray-400">{t('commandPalette.noResultsFor', { query })}</p>
        )}

        {!query && (
          <p className="px-4 py-4 text-xs text-center text-gray-400">
            {t('commandPalette.typeToSearch')}
          </p>
        )}

        <div className="flex items-center gap-4 px-4 py-2 border-t border-gray-100 dark:border-gray-700 text-xs text-gray-400">
          <span><kbd className="font-mono">↑↓</kbd> {t('commandPalette.navigateHint')}</span>
          <span><kbd className="font-mono">↵</kbd> {t('commandPalette.openHint')}</span>
          <span><kbd className="font-mono">Esc</kbd> {t('commandPalette.closeHint')}</span>
        </div>
      </div>
    </div>
  )
}
