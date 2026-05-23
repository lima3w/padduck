import { useState } from 'react'
import { gravatarUrl as computeGravatarUrl } from '../utils/md5'

/**
 * Avatar component — renders the user's avatar with an initials fallback.
 * Props:
 *   email        {string}  — user email (used to derive Gravatar URL)
 *   username     {string}  — username (used for initials fallback)
 *   size         {number}  — pixel size (default 32)
 *   gravatarUrl  {string}  — pre-computed Gravatar URL from the server (optional)
 *   avatarSource {string}  — 'gravatar' | 'custom' (default 'gravatar')
 *   avatarBust   {string}  — cache-buster string (e.g. updated_at) for custom avatars
 */
export default function Avatar({ email, username, size = 32, gravatarUrl, avatarSource, avatarBust }) {
  const [imgError, setImgError] = useState(false)

  let src
  if (avatarSource === 'custom') {
    // Serve the user's custom avatar through our own endpoint.
    // Append a cache-buster so the browser re-fetches after an upload.
    src = `/api/v1/auth/me/avatar${avatarBust ? `?_v=${encodeURIComponent(avatarBust)}` : ''}`
  } else {
    // Fall back to the server-provided Gravatar URL, or compute it client-side.
    src = gravatarUrl || (email ? computeGravatarUrl(email, size * 2) : null)
  }

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
