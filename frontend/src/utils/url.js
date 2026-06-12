// Only http(s) URLs are safe to render as anchor hrefs; anything else
// (javascript:, data:, vbscript:, protocol-relative tricks) is rejected.
export function isSafeHttpUrl(value) {
  return typeof value === 'string' && /^https?:\/\//i.test(value.trim())
}
