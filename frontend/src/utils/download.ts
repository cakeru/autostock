import api from '@/services/api'

// Downloads a backend-generated file (CSV, gzipped backup, …) as a browser
// download — the user picks where to save via the OS' normal dialog.
export async function downloadFile(url: string, filename: string) {
  const res = await api.get(url, { responseType: 'blob' })
  const objectUrl = URL.createObjectURL(res.data)
  const a = document.createElement('a')
  a.href = objectUrl
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(objectUrl)
}
