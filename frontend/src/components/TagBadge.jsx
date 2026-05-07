export default function TagBadge({ tag }) {
  if (!tag) return <span className="text-gray-400 text-xs">—</span>

  // Determine readable text colour based on background brightness
  const hex = tag.colour.replace('#', '')
  const r = parseInt(hex.substring(0, 2), 16)
  const g = parseInt(hex.substring(2, 4), 16)
  const b = parseInt(hex.substring(4, 6), 16)
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255
  const textClass = luminance > 0.6 ? 'text-gray-800' : 'text-white'

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs font-medium border ${textClass}`}
      style={{ backgroundColor: tag.colour, borderColor: luminance > 0.9 ? '#D1D5DB' : tag.colour }}
    >
      <span
        className="inline-block w-2 h-2 rounded-full border border-white/30"
        style={{ backgroundColor: luminance > 0.6 ? '#374151' : 'rgba(255,255,255,0.7)' }}
      />
      {tag.name}
    </span>
  )
}
