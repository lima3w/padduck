/**
 * Trigger a file download from an authenticated API endpoint.
 * Uses a temporary <a> element with the `download` attribute so the browser
 * saves the response rather than navigating to it.
 */
export async function downloadFile(url, filename) {
  const token = localStorage.getItem('auth_token')
  const headers = {}
  if (token) headers['Authorization'] = `Bearer ${token}`

  const response = await fetch(url, { headers })
  if (!response.ok) {
    throw new Error(`Download failed: ${response.status} ${response.statusText}`)
  }

  const blob = await response.blob()

  // Try to derive filename from Content-Disposition header if not provided
  if (!filename) {
    const disposition = response.headers.get('Content-Disposition')
    if (disposition) {
      const match = disposition.match(/filename[^;=\n]*=(['"]?)([^'";\n]+)\1/)
      if (match) filename = match[2]
    }
    if (!filename) filename = 'download'
  }

  const objectUrl = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = objectUrl
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(objectUrl)
}
