/**
 * Padduck wordmark as an inline SVG component.
 * Rendered inline so it shares the page's loaded Inter font.
 * Use logo.svg for non-React contexts (README, etc.).
 */
export default function PadduckLogo({ className = 'h-16 w-auto' }) {
  return (
    <svg
      viewBox="0 0 260 80"
      className={className}
      role="img"
      aria-label="Padduck IPAM"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect width="260" height="80" rx="12" fill="#07162b"/>
      {/* Duck mascot (favicon geometry, no background square) */}
      <svg x="8" y="8" width="64" height="64" viewBox="0 0 64 64">
        <circle cx="32" cy="32" r="21" fill="#F5B800"/>
        <path
          d="M42 29c7 0 12 3 12 7s-5 7-12 7"
          fill="#ff8a00"
          stroke="#07162B"
          strokeWidth="4"
          strokeLinecap="round"
        />
        <circle cx="37" cy="25" r="4" fill="#07162B"/>
        <path d="M19 38h22" stroke="#07162B" strokeWidth="5" strokeLinecap="round"/>
      </svg>
      {/* Wordmark: "PAD" white + "DUCK" gold */}
      <text
        x="82"
        y="46"
        fontFamily="'Inter', 'Helvetica Neue', Arial, system-ui, sans-serif"
        fontWeight="900"
        fontSize="36"
      >
        <tspan fill="white">PAD</tspan>
        <tspan fill="#F5B800">DUCK</tspan>
      </text>
      {/* Subtitle */}
      <text
        x="84"
        y="64"
        fontFamily="'Inter', 'Helvetica Neue', Arial, system-ui, sans-serif"
        fontWeight="500"
        fontSize="11"
        letterSpacing="5"
        fill="#a8b8cb"
      >
        IPAM
      </text>
    </svg>
  )
}
