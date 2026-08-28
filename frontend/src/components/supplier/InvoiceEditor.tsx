import { useRef } from 'react'
import { ImagePlus, Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { compressImage } from '@/utils/compressImage'
import { formatUSD } from '@/utils/currency'

// One supplier invoice being entered on a receive. A single receive can carry
// several invoices (e.g. a $100 purchase split into four $25 invoices) that
// are paid off one at a time.
export interface DraftInvoice {
  invoice_number: string
  amount: string
  file: File | null
  preview: string
  image_url: string
}

export function emptyInvoice(): DraftInvoice {
  return { invoice_number: '', amount: '', file: null, preview: '', image_url: '' }
}

export function InvoiceEditor({ invoices, onChange, total }: {
  invoices: DraftInvoice[]
  onChange: (invoices: DraftInvoice[]) => void
  total?: number
}) {
  const fileRefs = useRef<Record<number, HTMLInputElement | null>>({})
  const sum = invoices.reduce((s, inv) => s + (parseFloat(inv.amount) || 0), 0)
  const mismatch = total !== undefined && total > 0 && Math.abs(sum - total) > 0.005

  const pickFile = async (index: number, file: File | undefined) => {
    if (!file) return
    const prepared = await compressImage(file)
    onChange(invoices.map((inv, i) => (i === index ? { ...inv, file: prepared, preview: URL.createObjectURL(prepared) } : inv)))
  }

  const update = (index: number, patch: Partial<DraftInvoice>) =>
    onChange(invoices.map((inv, i) => (i === index ? { ...inv, ...patch } : inv)))

  const remove = (index: number) => onChange(invoices.filter((_, i) => i !== index))

  return (
    <div className="space-y-2">
      {invoices.map((inv, i) => (
        <div key={i} className="flex items-center gap-2">
          <Input
            value={inv.invoice_number}
            onChange={(e) => update(i, { invoice_number: e.target.value })}
            placeholder="Invoice #"
            className="w-32"
          />
          <Input
            value={inv.amount}
            onChange={(e) => update(i, { amount: e.target.value })}
            type="number"
            step="0.01"
            min="0"
            placeholder="Amount"
            className="w-24"
          />
          <div className="flex items-center gap-1.5">
            {inv.preview ? (
              <>
                <img src={inv.preview} alt="Invoice" className="h-8 w-8 rounded object-cover bg-muted" />
                <button
                  type="button"
                  onClick={() => update(i, { file: null, preview: '' })}
                  aria-label="Remove invoice photo"
                  className="grid h-5 w-5 place-items-center rounded-full bg-destructive text-destructive-foreground shadow"
                >
                  <X className="h-3 w-3" />
                </button>
              </>
            ) : (
              <Button type="button" variant="outline" size="sm" onClick={() => fileRefs.current[i]?.click()}>
                <ImagePlus className="h-3.5 w-3.5" /> Photo
              </Button>
            )}
            <input
              ref={(el) => { fileRefs.current[i] = el }}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={(e) => pickFile(i, e.target.files?.[0])}
            />
          </div>
          <button type="button" onClick={() => remove(i)} aria-label="Remove invoice" className="text-muted-foreground hover:text-destructive">
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
      <div className="flex items-center justify-between">
        <Button type="button" variant="ghost" size="sm" onClick={() => onChange([...invoices, emptyInvoice()])}>
          <Plus className="h-3.5 w-3.5" /> Add invoice
        </Button>
        {total !== undefined && total > 0 && (
          <span className={`text-xs ${mismatch ? 'text-amber-600' : 'text-muted-foreground'}`}>
            Invoices total: {formatUSD(sum)} of {formatUSD(total)}
            {mismatch && ' — check amounts'}
          </span>
        )}
      </div>
    </div>
  )
}