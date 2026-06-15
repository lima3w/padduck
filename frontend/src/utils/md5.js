// Minimal MD5 implementation for Gravatar URL generation.
// Based on the RSA Data Security, Inc. MD5 Message-Digest Algorithm.

function safeAdd(x, y) {
  const lsw = (x & 0xffff) + (y & 0xffff)
  const msw = (x >> 16) + (y >> 16) + (lsw >> 16)
  return (msw << 16) | (lsw & 0xffff)
}
function bitRotateLeft(num, cnt) {
  return (num << cnt) | (num >>> (32 - cnt))
}
function md5cmn(q, a, b, x, s, t) {
  return safeAdd(bitRotateLeft(safeAdd(safeAdd(a, q), safeAdd(x, t)), s), b)
}
function md5ff(a, b, c, d, x, s, t) { return md5cmn((b & c) | (~b & d), a, b, x, s, t) }
function md5gg(a, b, c, d, x, s, t) { return md5cmn((b & d) | (c & ~d), a, b, x, s, t) }
function md5hh(a, b, c, d, x, s, t) { return md5cmn(b ^ c ^ d, a, b, x, s, t) }
function md5ii(a, b, c, d, x, s, t) { return md5cmn(c ^ (b | ~d), a, b, x, s, t) }

function md5cycle(x, k) {
  let [a, b, c, d] = x
  a = md5ff(a, b, c, d, k[0], 7, -680876936); d = md5ff(d, a, b, c, k[1], 12, -389564586)
  c = md5ff(c, d, a, b, k[2], 17, 606105819); b = md5ff(b, c, d, a, k[3], 22, -1044525330)
  a = md5ff(a, b, c, d, k[4], 7, -176418897); d = md5ff(d, a, b, c, k[5], 12, 1200080426)
  c = md5ff(c, d, a, b, k[6], 17, -1473231341); b = md5ff(b, c, d, a, k[7], 22, -45705983)
  a = md5ff(a, b, c, d, k[8], 7, 1770035416); d = md5ff(d, a, b, c, k[9], 12, -1958414417)
  c = md5ff(c, d, a, b, k[10], 17, -42063); b = md5ff(b, c, d, a, k[11], 22, -1990404162)
  a = md5ff(a, b, c, d, k[12], 7, 1804603682); d = md5ff(d, a, b, c, k[13], 12, -40341101)
  c = md5ff(c, d, a, b, k[14], 17, -1502002290); b = md5ff(b, c, d, a, k[15], 22, 1236535329)
  a = md5gg(a, b, c, d, k[1], 5, -165796510); d = md5gg(d, a, b, c, k[6], 9, -1069501632)
  c = md5gg(c, d, a, b, k[11], 14, 643717713); b = md5gg(b, c, d, a, k[0], 20, -373897302)
  a = md5gg(a, b, c, d, k[5], 5, -701558691); d = md5gg(d, a, b, c, k[10], 9, 38016083)
  c = md5gg(c, d, a, b, k[15], 14, -660478335); b = md5gg(b, c, d, a, k[4], 20, -405537848)
  a = md5gg(a, b, c, d, k[9], 5, 568446438); d = md5gg(d, a, b, c, k[14], 9, -1019803690)
  c = md5gg(c, d, a, b, k[3], 14, -187363961); b = md5gg(b, c, d, a, k[8], 20, 1163531501)
  a = md5gg(a, b, c, d, k[13], 5, -1444681467); d = md5gg(d, a, b, c, k[2], 9, -51403784)
  c = md5gg(c, d, a, b, k[7], 14, 1735328473); b = md5gg(b, c, d, a, k[12], 20, -1926607734)
  a = md5hh(a, b, c, d, k[5], 4, -378558); d = md5hh(d, a, b, c, k[8], 11, -2022574463)
  c = md5hh(c, d, a, b, k[11], 16, 1839030562); b = md5hh(b, c, d, a, k[14], 23, -35309556)
  a = md5hh(a, b, c, d, k[1], 4, -1530992060); d = md5hh(d, a, b, c, k[4], 11, 1272893353)
  c = md5hh(c, d, a, b, k[7], 16, -155497632); b = md5hh(b, c, d, a, k[10], 23, -1094730640)
  a = md5hh(a, b, c, d, k[13], 4, 681279174); d = md5hh(d, a, b, c, k[0], 11, -358537222)
  c = md5hh(c, d, a, b, k[3], 16, -722521979); b = md5hh(b, c, d, a, k[6], 23, 76029189)
  a = md5hh(a, b, c, d, k[9], 4, -640364487); d = md5hh(d, a, b, c, k[12], 11, -421815835)
  c = md5hh(c, d, a, b, k[15], 16, 530742520); b = md5hh(b, c, d, a, k[2], 23, -995338651)
  a = md5ii(a, b, c, d, k[0], 6, -198630844); d = md5ii(d, a, b, c, k[7], 10, 1126891415)
  c = md5ii(c, d, a, b, k[14], 15, -1416354905); b = md5ii(b, c, d, a, k[5], 21, -57434055)
  a = md5ii(a, b, c, d, k[12], 6, 1700485571); d = md5ii(d, a, b, c, k[3], 10, -1894986606)
  c = md5ii(c, d, a, b, k[10], 15, -1051523); b = md5ii(b, c, d, a, k[1], 21, -2054922799)
  a = md5ii(a, b, c, d, k[8], 6, 1873313359); d = md5ii(d, a, b, c, k[15], 10, -30611744)
  c = md5ii(c, d, a, b, k[6], 15, -1560198380); b = md5ii(b, c, d, a, k[13], 21, 1309151649)
  a = md5ii(a, b, c, d, k[4], 6, -145523070); d = md5ii(d, a, b, c, k[11], 10, -1120210379)
  c = md5ii(c, d, a, b, k[2], 15, 718787259); b = md5ii(b, c, d, a, k[9], 21, -343485551)
  x[0] = safeAdd(a, x[0]); x[1] = safeAdd(b, x[1])
  x[2] = safeAdd(c, x[2]); x[3] = safeAdd(d, x[3])
  return x
}

function md5blks(s) {
  const md5blksize = 64
  const length = s.length
  const orig = length + 8
  const count = ((orig - (orig % md5blksize)) / md5blksize + 1) * 16
  const blks = new Array(count).fill(0)
  let i = 0
  for (; i < length; i++) blks[i >> 2] |= s.charCodeAt(i) << ((i % 4) * 8)
  blks[i >> 2] |= 0x80 << ((i % 4) * 8)
  blks[count - 2] = length * 8
  return blks
}

function hex(x) {
  const hexChars = '0123456789abcdef'
  let out = ''
  for (let i = 0; i < 4; i++) {
    out +=
      hexChars[(x >> (i * 8 + 4)) & 0xf] +
      hexChars[(x >> (i * 8)) & 0xf]
  }
  return out
}

/**
 * Returns the MD5 hex digest of a string.
 * Input must be an ASCII/Latin1 string; for UTF-8 encode with encodeURIComponent first.
 */
export function md5(str) {
  let state = [1732584193, -271733879, -1732584194, 271733878]
  const blks = md5blks(str)
  for (let i = 0; i < blks.length; i += 16) {
    state = md5cycle(state, blks.slice(i, i + 16))
  }
  return hex(state[0]) + hex(state[1]) + hex(state[2]) + hex(state[3])
}

/**
 * Returns a Gravatar URL for the given email address.
 * Falls back to the "identicon" auto-generated avatar if no Gravatar is set.
 */
export function gravatarUrl(email, size = 80) {
  const normalized = (email || '').trim().toLowerCase()
  // Encode non-ASCII characters for MD5 input consistency
  const encoded = unescape(encodeURIComponent(normalized))
  const hash = md5(encoded)
  return `https://www.gravatar.com/avatar/${hash}?d=identicon&s=${size}`
}
