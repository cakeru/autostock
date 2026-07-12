import { useEffect, useMemo, useState } from 'react'
import QRCode from 'qrcode'
import { X, Printer } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { Batch, Product } from '@/types/product'

// A print sheet of identical batch-QR labels. Every unit from one intake batch
// carries the same code (KSB:<batchId>) — the batch is the traceable unit, so a
// mechanic scanning any label records "a unit from this batch went on the car".
// All labels are identical, so we render one QR and tile it.
export function BatchLabelSheet({ product, batch, onClose }: {
  product: Product
  batch: Batch
  onClose: () => void
}) {
  const [qty, setQty] = useState(String(Math.max(1, Math.ceil(batch.quantity_remaining))))
  const [qr, setQr] = useState<string>('')

  const code = `KSB:${batch.id}`
  useEffect(() => {
    QRCode.toDataURL(code, { margin: 0, width: 220, errorCorrectionLevel: 'M' }).then(setQr).catch(() => setQr(''))
  }, [code])

  const count = useMemo(() => Math.max(0, Math.min(200, parseInt(qty) || 0)), [qty])
  const dot = batch.dot_code

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4 print:static print:bg-white print:p-0">
      <div className="my-6 w-full max-w-3xl rounded-lg bg-card p-5 shadow-lg print:my-0 print:max-w-full print:rounded-none print:shadow-none">
        {/* Controls (hidden when printing) */}
        <div className="mb-4 flex items-center justify-between print:hidden">
          <div>
            <p className="text-sm font-semibold">Batch labels — {batch.batch_no}</p>
            <p className="text-xs text-muted-foreground">{product.name}</p>
          </div>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground" aria-label="Close"><X className="h-5 w-5" /></button>
        </div>

        <div className="mb-4 flex flex-wrap items-end gap-3 print:hidden">
          <div className="space-y-1">
            <label className="text-xs font-medium">How many labels</label>
            <Input value={qty} onChange={(e) => setQty(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" className="w-28" />
            <p className="text-[11px] text-muted-foreground">{batch.quantity_remaining} on the shelf now</p>
          </div>
          <Button onClick={() => window.print()} disabled={!qr || count < 1}>
            <Printer className="mr-1.5 h-3.5 w-3.5" /> Print {count} label{count === 1 ? '' : 's'}
          </Button>
          <p className="max-w-xs text-[11px] text-muted-foreground">
            Stick one on each item. A mechanic scans it before fitting so the exact batch is recorded against the car.
          </p>
        </div>

        {/* Label grid — this is what prints */}
        <div className="grid grid-cols-3 gap-2 sm:grid-cols-4 print:grid-cols-4 print:gap-1.5">
          {Array.from({ length: count }).map((_, i) => (
            <div key={i} className="flex items-center gap-2 rounded border border-black/70 p-2 print:break-inside-avoid">
              {qr && <img src={qr} alt="" className="h-14 w-14 flex-shrink-0" />}
              <div className="min-w-0 leading-tight">
                <p className="truncate text-[10px] font-semibold">{product.tire_brand || product.name}</p>
                {product.tire_size && <p className="text-[9px] text-black/70">{product.tire_size}</p>}
                <p className="font-mono text-[9px] text-black/70">{batch.batch_no}</p>
                {dot && <p className="text-[9px] text-black/70">DOT {dot}</p>}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
