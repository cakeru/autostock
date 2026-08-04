// Downsizes/compresses a photo in the browser before upload, so phone camera
// shots (often 5-15 MB, HEIC or JPEG) leave the device as a small standard
// JPEG. The backend re-encodes to 800px anyway, so the raw size was pure
// bandwidth waste. Falls back to the original file when the browser can't
// decode it (e.g. HEIC on Android) — the server then reports its own error.
export async function compressImage(file: File, maxDim = 1280, quality = 0.82): Promise<File> {
  if (file.size <= 300 * 1024) return file // already small — keep as-is

  let bitmap: ImageBitmap | null = null
  try {
    bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' })
  } catch {
    return file
  }

  try {
    const scale = Math.min(1, maxDim / Math.max(bitmap.width, bitmap.height))
    const w = Math.max(1, Math.round(bitmap.width * scale))
    const h = Math.max(1, Math.round(bitmap.height * scale))
    const canvas = document.createElement('canvas')
    canvas.width = w
    canvas.height = h
    const ctx = canvas.getContext('2d')
    if (!ctx) return file

    ctx.drawImage(bitmap, 0, 0, w, h)
    const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/jpeg', quality))
    if (!blob || blob.size >= file.size) return file // compression didn't help

    const name = file.name.replace(/\.[^.]+$/, '') + '.jpg'
    return new File([blob], name, { type: 'image/jpeg' })
  } finally {
    bitmap.close()
  }
}
