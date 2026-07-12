import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { formatUSD } from '@/utils/currency'
import { useCreateReturn } from '@/hooks/useReturns'
import type { InvoiceDetail } from '@/types/invoice'

export function ReturnDialog({ invoice, returnedByItem, onClose }: {
  invoice: InvoiceDetail
  returnedByItem: Record<string, number>
  onClose: () => void
}) {
  const create = useCreateReturn()
  const [qtys, setQtys] = useState<Record<number, string>>({})
  const [method, setMethod] = useState<'cash' | 'store_credit'>('cash')
  const [reason, setReason] = useState('')

  const remaining = (item: InvoiceDetail['items'][number]) => item.quantity - (returnedByItem[String(item.id)] || 0)
  const setQty = (id: number, v: string) => setQtys((p) => ({ ...p, [id]: v }))

  const refundTotal = invoice.items.reduce((s, it) => s + (parseFloat(qtys[it.id] || '0') || 0) * it.unit_price_usd, 0)
  const hasSomething = invoice.items.some((it) => (parseFloat(qtys[it.id] || '0') || 0) > 0)
  const overReturn = invoice.items.some((it) => (parseFloat(qtys[it.id] || '0') || 0) > remaining(it) + 0.001)
  const storeCreditDisabled = !invoice.customer_id

  const submit = () => {
    const items = invoice.items
      .map((it) => ({ invoice_item_id: it.id, quantity: parseFloat(qtys[it.id] || '0') || 0 }))
      .filter((x) => x.quantity > 0)
    if (items.length === 0) return
    create.mutate({ invoice_id: invoice.id, refund_method: method, reason: reason || undefined, items }, { onSuccess: onClose })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-auto bg-black/50 p-4">
      <div className="w-full max-w-md rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Return / refund</p>
        <p className="mb-3 text-xs text-muted-foreground">{invoice.invoice_number} — enter quantities to return</p>

        <div className="space-y-2">
          {invoice.items.map((it) => {
            const rem = remaining(it)
            return (
              <div key={it.id} className="flex items-center justify-between gap-2 border-b pb-2 last:border-0">
                <div className="min-w-0">
                  <p className="truncate text-sm">{it.description}</p>
                  <p className="text-xs text-muted-foreground">{rem > 0 ? `${rem} of ${it.quantity} left · ${formatUSD(it.unit_price_usd)} ea` : 'fully returned'}</p>
                </div>
                <Input
                  type="number" min="0" step="1" max={rem}
                  disabled={rem <= 0}
                  value={qtys[it.id] || ''}
                  onChange={(e) => setQty(it.id, e.target.value)}
                  className="w-16 flex-shrink-0"
                  placeholder="0"
                />
              </div>
            )
          })}
        </div>

        <div className="mt-3 space-y-2">
          <div className="flex gap-2">
            {(['cash', 'store_credit'] as const).map((m) => (
              <button key={m} onClick={() => setMethod(m)} disabled={m === 'store_credit' && storeCreditDisabled}
                className={`flex-1 rounded-md border py-2 text-sm font-medium transition-colors disabled:opacity-40 ${method === m ? 'border-primary bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}>
                {m === 'cash' ? 'Refund cash' : 'Store credit'}
              </button>
            ))}
          </div>
          {method === 'store_credit' && storeCreditDisabled && <p className="text-xs text-muted-foreground">Store credit needs a customer on the invoice.</p>}
          <Input value={reason} onChange={(e) => setReason(e.target.value)} placeholder="Reason (optional)" />
        </div>

        <div className="mt-3 flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Refund</span>
          <span className="text-lg font-bold tabular-nums">{formatUSD(refundTotal)}</span>
        </div>
        {overReturn && <p className="mt-1 text-xs text-destructive">Cannot return more than remains.</p>}

        <div className="mt-3 flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button variant="destructive" onClick={submit} disabled={!hasSomething || overReturn || create.isPending}>
            {create.isPending ? 'Processing…' : `Process return ${formatUSD(refundTotal)}`}
          </Button>
        </div>
      </div>
    </div>
  )
}
