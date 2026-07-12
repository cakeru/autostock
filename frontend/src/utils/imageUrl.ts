// Product images stored by the local driver come back as a relative path
// (/uploads/…) served by the backend, while the R2 driver returns an absolute
// URL. Resolve relative paths against the API origin; pass absolute ones through.
const apiBase = import.meta.env.VITE_API_URL || '/api/v1'
let apiOrigin = ''
try {
  apiOrigin = new URL(apiBase, window.location.origin).origin
} catch {
  apiOrigin = ''
}

export function imageSrc(url?: string): string {
  if (!url) return ''
  if (/^https?:\/\//.test(url) || url.startsWith('data:')) return url
  return apiOrigin + url
}
