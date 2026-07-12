import { useEffect, useRef, useState } from 'react'
import { Html5Qrcode } from 'html5-qrcode'

// Thin wrapper around html5-qrcode that owns the camera lifecycle: it starts the
// rear camera, fires onResult once on the first decode, and always releases the
// camera on unmount. Falls back to a message if no camera is available.
export function QrScanner({ onResult }: { onResult: (text: string) => void }) {
  const ref = useRef<HTMLDivElement>(null)
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const doneRef = useRef(false)
  const [error, setError] = useState<string>('')

  useEffect(() => {
    if (!ref.current) return
    const id = 'qr-reader-' + Math.random().toString(36).slice(2)
    ref.current.id = id
    const scanner = new Html5Qrcode(id)
    scannerRef.current = scanner

    scanner
      .start(
        { facingMode: 'environment' },
        { fps: 10, qrbox: { width: 230, height: 230 } },
        (decoded) => {
          if (doneRef.current) return
          doneRef.current = true
          onResult(decoded)
        },
        () => { /* per-frame decode misses are normal; ignore */ },
      )
      .catch((e) => setError(String(e?.message || e)))

    return () => {
      const s = scannerRef.current
      if (s) {
        s.stop().then(() => s.clear()).catch(() => { /* already stopped */ })
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (error) {
    return (
      <div className="rounded-lg border border-dashed p-4 text-center text-sm text-muted-foreground">
        Couldn't open the camera ({error}). Use the manual entry below.
      </div>
    )
  }

  return <div ref={ref} className="mx-auto w-full max-w-xs overflow-hidden rounded-lg border" />
}
