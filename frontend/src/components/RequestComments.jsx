import { useState, useEffect, useCallback } from 'react'
import Avatar from './Avatar'
import { getRequestComments, addRequestComment } from '../api/requests'

function formatRelativeTime(isoString) {
  const now = Date.now()
  const then = new Date(isoString).getTime()
  const diff = Math.floor((now - then) / 1000)
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)} min ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)} hr ago`
  return `${Math.floor(diff / 86400)} days ago`
}

/**
 * RequestComments — renders a comment thread for a request.
 * Props:
 *   type  {string} — 'subnets' | 'ips'
 *   id    {number} — request ID
 */
export default function RequestComments({ type, id }) {
  const [comments, setComments] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [body, setBody] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState(null)

  const loadComments = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const res = await getRequestComments(type, id)
      setComments(Array.isArray(res.data) ? res.data : (res.data?.comments ?? []))
    } catch {
      setError('Failed to load comments')
    } finally {
      setLoading(false)
    }
  }, [type, id])

  useEffect(() => {
    if (id) loadComments()
  }, [id, loadComments])

  async function handleSubmit(e) {
    e.preventDefault()
    if (!body.trim()) return
    setSubmitting(true)
    setSubmitError(null)
    try {
      await addRequestComment(type, id, body.trim())
      setBody('')
      await loadComments()
    } catch {
      setSubmitError('Failed to post comment')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mt-4 border-t dark:border-gray-600 pt-4">
      <h4 className="text-sm font-semibold text-gray-600 dark:text-gray-300 mb-3">Comments</h4>

      {loading ? (
        <p className="text-sm text-gray-400">Loading comments...</p>
      ) : error ? (
        <p className="text-sm text-red-500">{error}</p>
      ) : comments.length === 0 ? (
        <p className="text-sm text-gray-400 mb-3">No comments yet.</p>
      ) : (
        <div className="space-y-3 mb-4">
          {comments.map((c) => (
            <div key={c.id} className="flex items-start gap-3">
              <Avatar
                email={c.email || c.userEmail}
                username={c.username}
                gravatarUrl={c.gravatarUrl}
                size={28}
              />
              <div className="flex-1 min-w-0">
                <div className="flex items-baseline gap-2">
                  <span className="text-sm font-medium text-gray-800 dark:text-gray-100">{c.username}</span>
                  <span className="text-xs text-gray-400">{formatRelativeTime(c.createdAt)}</span>
                </div>
                <p className="text-sm text-gray-600 dark:text-gray-300 mt-0.5 whitespace-pre-wrap">{c.body}</p>
              </div>
            </div>
          ))}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-2">
        {submitError && <p className="text-sm text-red-500">{submitError}</p>}
        <textarea
          className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
          rows={2}
          placeholder="Add a comment..."
          value={body}
          onChange={e => setBody(e.target.value)}
        />
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={submitting || !body.trim()}
            className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            {submitting ? 'Posting...' : 'Post Comment'}
          </button>
        </div>
      </form>
    </div>
  )
}
