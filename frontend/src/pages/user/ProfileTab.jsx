import { useState, useRef } from 'react'
import * as client from '../../api/auth'
import { gravatarUrl } from '../../utils/md5'

const AVATAR_ENDPOINT = '/api/v1/auth/me/avatar'
const MAX_AVATAR_PX = 256   // resize canvas target
const MAX_AVATAR_BYTES = 2 * 1024 * 1024  // 2 MiB raw before base64

/** Resize an image file to at most MAX_AVATAR_PX × MAX_AVATAR_PX and return a JPEG data URL. */
function resizeImage(file) {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(url)
      const scale = Math.min(1, MAX_AVATAR_PX / img.width, MAX_AVATAR_PX / img.height)
      const w = Math.round(img.width * scale)
      const h = Math.round(img.height * scale)
      const canvas = document.createElement('canvas')
      canvas.width = w
      canvas.height = h
      canvas.getContext('2d').drawImage(img, 0, 0, w, h)
      resolve(canvas.toDataURL('image/jpeg', 0.88))
    }
    img.onerror = reject
    img.src = url
  })
}

export default function ProfileTab({ user, onAvatarChange }) {
  const [source, setSource] = useState(user?.avatarSource || 'gravatar')
  const [preview, setPreview] = useState(null)   // data URL for the chosen custom image
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const fileRef = useRef(null)

  // Live avatar shown in the card; append updatedAt as a cache-buster for custom avatars
  const currentAvatarSrc =
    user?.avatarSource === 'custom'
      ? `${AVATAR_ENDPOINT}${user.updatedAt ? `?_v=${encodeURIComponent(user.updatedAt)}` : ''}`
      : (user?.email ? gravatarUrl(user.email, 80) : null)

  // Preview shown while the user has selected a new file
  const displaySrc = preview || currentAvatarSrc

  async function handleFileChange(e) {
    const file = e.target.files?.[0]
    if (!file) return
    setError('')
    if (file.size > MAX_AVATAR_BYTES) {
      setError('Image must be smaller than 2 MB')
      return
    }
    try {
      const dataUrl = await resizeImage(file)
      setPreview(dataUrl)
      setSource('custom')
    } catch {
      setError('Could not read image file')
    }
  }

  function handleSourceChange(val) {
    setSource(val)
    if (val === 'gravatar') setPreview(null)
  }

  async function handleSave() {
    if (source === 'custom' && !preview) {
      setError('Please select an image first')
      return
    }
    setSaving(true)
    setError('')
    setSuccess(false)
    try {
      const data = source === 'custom' ? preview : undefined
      await client.updateMyAvatar(source, data)
      // Refresh the user from the server so the header avatar updates immediately.
      const res = await client.getCurrentUser()
      if (res.data) {
        onAvatarChange?.(res.data)
      }
      setPreview(null)
      setSuccess(true)
      setTimeout(() => setSuccess(false), 3000)
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save avatar')
    } finally {
      setSaving(false)
    }
  }

  // Dirty when: new image selected (custom) OR switching back to Gravatar from custom
  const isDirty = (source === 'custom' && preview !== null) ||
    (source === 'gravatar' && user?.avatarSource === 'custom')

  return (
    <div className="max-w-lg space-y-6">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Profile</h2>

      {/* Current avatar preview */}
      <div className="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        {displaySrc ? (
          <img
            src={displaySrc}
            alt="Profile avatar"
            className="w-16 h-16 rounded-full border-2 border-gray-200 dark:border-gray-600 object-cover"
            onError={(e) => { e.currentTarget.style.display = 'none' }}
          />
        ) : (
          <div className="w-16 h-16 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-2xl font-bold text-blue-600 dark:text-blue-300 border-2 border-gray-200 dark:border-gray-600">
            {(user?.username || '?')[0].toUpperCase()}
          </div>
        )}
        <div>
          <p className="text-sm font-medium text-gray-700 dark:text-gray-200">{user?.username}</p>
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{user?.email}</p>
        </div>
      </div>

      {/* Avatar source selector */}
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 space-y-3">
        <p className="text-sm font-medium text-gray-700 dark:text-gray-200">Avatar source</p>

        <label className="flex items-start gap-3 cursor-pointer">
          <input
            type="radio"
            name="avatar_source"
            value="gravatar"
            checked={source === 'gravatar'}
            onChange={() => handleSourceChange('gravatar')}
            className="mt-0.5 accent-blue-600"
          />
          <span>
            <span className="text-sm font-medium text-gray-700 dark:text-gray-200">Gravatar</span>
            <span className="block text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Automatically use the avatar associated with your email at{' '}
              <a href="https://gravatar.com" target="_blank" rel="noreferrer" className="text-blue-500 hover:underline">
                gravatar.com
              </a>
            </span>
          </span>
        </label>

        <label className="flex items-start gap-3 cursor-pointer">
          <input
            type="radio"
            name="avatar_source"
            value="custom"
            checked={source === 'custom'}
            onChange={() => handleSourceChange('custom')}
            className="mt-0.5 accent-blue-600"
          />
          <span>
            <span className="text-sm font-medium text-gray-700 dark:text-gray-200">Custom image</span>
            <span className="block text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Upload your own photo (JPEG, PNG, WebP — max 2 MB)
            </span>
          </span>
        </label>

        {source === 'custom' && (
          <div className="ml-7 space-y-2">
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              className="text-sm px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200 transition"
            >
              {preview ? 'Choose a different image…' : 'Choose image…'}
            </button>
            <input
              ref={fileRef}
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              className="hidden"
              onChange={handleFileChange}
            />
            {preview && (
              <div className="flex items-center gap-3">
                <img src={preview} alt="Preview" className="w-12 h-12 rounded-full object-cover border border-gray-200 dark:border-gray-600" />
                <span className="text-xs text-gray-500 dark:text-gray-400">Preview</span>
              </div>
            )}
          </div>
        )}

        {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
        {success && <p className="text-sm text-green-600 dark:text-green-400">Avatar saved.</p>}

        <button
          type="button"
          onClick={handleSave}
          disabled={saving || !isDirty}
          className="mt-1 px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {saving ? 'Saving…' : 'Save avatar'}
        </button>
      </div>

      {/* Account info */}
      <dl className="space-y-3">
        <div>
          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Username</dt>
          <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">{user?.username}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Email</dt>
          <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">{user?.email}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Role</dt>
          <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100 capitalize">{user?.role}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Account State</dt>
          <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100 capitalize">{user?.state?.replace(/_/g, ' ')}</dd>
        </div>
      </dl>
    </div>
  )
}
