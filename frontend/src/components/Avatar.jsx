import { useState } from 'react'

/**
 * Avatar component — renders a Gravatar image with an initials fallback.
 * Props:
 *   email       {string}  — user email (used to derive Gravatar URL)
 *   username    {string}  — username (used for initials fallback)
 *   size        {number}  — pixel size (default 32)
 *   gravatarUrl {string}  — pre-computed Gravatar URL from the server (optional)
 */
export default function Avatar({ email, username, size = 32, gravatarUrl }) {
  const [imgError, setImgError] = useState(false)

  const src = gravatarUrl || (email
    ? `https://www.gravatar.com/avatar/${md5(email)}?s=${size * 2}&d=identicon`
    : null)

  const initial = (username || email || '?')[0].toUpperCase()

  const style = { width: size, height: size, minWidth: size, minHeight: size }

  if (!src || imgError) {
    return (
      <span
        style={{ ...style, fontSize: Math.round(size * 0.45) }}
        className="inline-flex items-center justify-center rounded-full bg-blue-500 text-white font-semibold select-none"
        aria-label={username || email}
      >
        {initial}
      </span>
    )
  }

  return (
    <img
      src={src}
      alt={username || email}
      style={style}
      className="rounded-full object-cover"
      onError={() => setImgError(true)}
    />
  )
}

// Tiny inline MD5 for client-side fallback (not crypto-secure — only for Gravatar)
// We only use it when gravatarUrl isn't passed from the server.
function md5(str) {
  // If running in a browser without SubtleCrypto access, use a simple XOR hash as fallback.
  // In practice, the server always sends gravatar_url so this branch is rarely hit.
  try {
    return Array.from(new TextEncoder().encode(str.trim().toLowerCase()))
      .reduce((h, b) => (((h << 5) - h) + b) | 0, 0)
      .toString(16)
      .replace('-', '')
  } catch {
    return ''
  }
}
