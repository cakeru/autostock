import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/button'
import { formatUSD } from '@/utils/currency'
import { formatDate } from '@/utils/date'
import { imageSrc } from '@/utils/imageUrl'
import { useSupplier, useSupplierPurchases, usePaySupplier } from '@/hooks/useSuppliers'
import type { Purchase } from '@/types/supplier'

// One selectable payment unit: an unpaid invoice, or a whole purchase that
// has no invoices recorded.
interface PayableItem {
  key: string
  label: string
  sub: string
  owed: number
  invoiceId?: number
  batchId?: number
}

function payableItems(purchases: Purchase[]): PayableItem[] {
  const items: PayableItem[] = []
  for (const p of purchases) {
    const unpaidInvoices = p.invoices.filter((inv) => inv.owed > 0)
    if (unpaidInvoices.length > 0) {
      for (const inv of unpaidInvoices) {
        items.push({
          key: `inv:${inv.id}`,
          label: inv.invoice_number ? `Invoice ${inv.invoice_number}` : 'Invoice',
          sub: `${p.product_name || 'Purchase'} · ${formatDate(p.received_at)}`,
          owed: inv.owed,
          invoiceId: inv.id,
        })
      }
    } else if (p.owed > 0) {
      items.push({
        key: `batch:${p.batch_id}`,
        label: p.product_name || 'Purchase',
        sub: `No invoice recorded · ${formatDate(p.received_at)}`,
        owed: p.owed,
        batchId: p.batch_id,
      })
    }
  }
  return items
}

export function SupplierDetail() {
  const { id } = useParams<{ id: string }>()
  const supplierId = parseInt(id || '0')
  const { data: supplier, isLoading } = useSupplier(supplierId)
  const { data: purchases } = useSupplierPurchases(supplierId)
  const pay = usePaySupplier()
  const [selected, setSelected] = useState<Set<string>>(new Set())

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!supplier) return <p className="text-sm text-destructive">Supplier not found</p>

  const payable = payableItems(purchases || [])
  const toggle = (key: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }
  const selectedTotal = payable.filter((it) => selected.has(it.key)).reduce((sum, it) => sum + it.owed, 0)

  const doPay = (items: PayableItem[]) => {
    if (items.length === 0) return
    pay.mutate({
      id: supplierId,
      data: {
        invoice_ids: items.filter((it) => it.invoiceId).map((it) => it.invoiceId!),
        batch_ids: items.filter((it) => it.batchId).map((it) => it.batchId!),
      },
    }, { onSuccess: () => setSelected(new Set()) })
  }

  return (
    <div className="max-w-4xl space-y-6">
      <PageHeader title={supplier.name} backTo={-1} breadcrumb="Suppliers" />

      <div className="grid grid-cols-3 gap-3">
        <Kpi label="Total bought" value={formatUSD(supplier.total_purchased)} />
        <Kpi label="Owed" value={formatUSD(supplier.outstanding)} alert={supplier.outstanding > 0} />
        <Kpi label="Purchases" value={String(supplier.purchase_count)} />
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-lg bg-card p-5 shadow-sm space-y-1.5">
          <p className="text-sm font-semibold">Contact</p>
          {supplier.phone && <p className="text-sm"><a href={`tel:${supplier.phone}`} className="text-primary hover:underline">{supplier.phone}</a></p>}
          {supplier.email && <p className="text-sm"><a href={`mailto:${supplier.email}`} className="text-primary hover:underline">{supplier.email}</a></p>}
          {supplier.address && <p className="text-sm text-muted-foreground">{supplier.address}</p>}
          {supplier.notes && <p className="text-sm text-muted-foreground">{supplier.notes}</p>}
          {!supplier.phone && !supplier.email && !supplier.address && <p className="text-sm text-muted-foreground">No contact details</p>}
        </div>

        <div className="rounded-lg bg-card p-5 shadow-sm">
          <p className="mb-2 text-sm font-semibold">Record payment</p>
          {payable.length > 0 ? (
            <>
              <p className="mb-2 text-xs text-muted-foreground">Tick the invoices to pay — each is settled in full.</p>
              <div className="max-h-56 space-y-1 overflow-y-auto pr-1">
                {payable.map((it) => (
                  <label key={it.key} className="flex cursor-pointer items-center gap-2 rounded border border-border px-2 py-1.5 text-sm hover:bg-muted/50">
                    <input type="checkbox" checked={selected.has(it.key)} onChange={() => toggle(it.key)} className="accent-primary" />
                    <span className="min-w-0 flex-1 truncate">
                      <span className="font-medium">{it.label}</span>
                      <span className="ml-2 text-xs text-muted-foreground">{it.sub}</span>
                    </span>
                    <span className="tabular-nums text-destructive">{formatUSD(it.owed)}</span>
                  </label>
                ))}
              </div>
              <div className="mt-3 flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground">
                  {selected.size > 0 ? `Selected: ${formatUSD(selectedTotal)}` : 'Nothing selected'}
                </span>
                <div className="flex gap-2">
                  <Button size="sm" variant="ghost" onClick={() => doPay(payable)} disabled={pay.isPending}>
                    Pay all
                  </Button>
                  <Button size="sm" disabled={selected.size === 0 || pay.isPending} onClick={() => doPay(payable.filter((it) => selected.has(it.key)))}>
                    {pay.isPending ? 'Saving…' : `Pay selected${selected.size > 0 ? ` (${formatUSD(selectedTotal)})` : ''}`}
                  </Button>
                </div>
              </div>
            </>
          ) : (
            <p className="text-sm text-muted-foreground">Nothing owed — all purchases paid.</p>
          )}
        </div>
      </div>

      <div className="rounded-lg bg-card p-5 shadow-sm">
        <p className="mb-2 text-sm font-semibold">Purchase history</p>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[560px] text-sm">
            <thead>
              <tr className="border-b text-left text-xs text-muted-foreground">
                <th className="py-2 pr-3 font-medium">Date</th>
                <th className="py-2 pr-3 font-medium">Product</th>
                <th className="py-2 pr-3 font-medium">Invoices</th>
                <th className="py-2 pr-3 text-right font-medium">Qty</th>
                <th className="py-2 pr-3 text-right font-medium">Unit</th>
                <th className="py-2 pr-3 text-right font-medium">Total</th>
                <th className="py-2 text-right font-medium">Owed</th>
              </tr>
            </thead>
            <tbody>
              {!purchases || purchases.length === 0 ? (
                <tr><td colSpan={7} className="py-6 text-center text-muted-foreground">No purchases from this supplier yet</td></tr>
              ) : purchases.map((p) => (
                <tr key={p.batch_id} className="border-b last:border-0">
                  <td className="py-2 pr-3 tabular-nums text-muted-foreground">{formatDate(p.received_at)}</td>
                  <td className="py-2 pr-3">{p.product_name || '—'}</td>
                  <td className="py-2 pr-3">
                    {p.invoices.length > 0 ? (
                      <span className="flex flex-wrap items-center gap-1.5">
                        {p.invoices.map((inv) => (
                          <span key={inv.id} className="flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-xs">
                            {inv.invoice_number && <span className="text-muted-foreground">{inv.invoice_number}</span>}
                            <span className="tabular-nums">{formatUSD(inv.amount)}</span>
                            {inv.invoice_image && (
                              <a href={imageSrc(inv.invoice_image)} target="_blank" rel="noreferrer" title="View invoice photo">
                                <img src={imageSrc(inv.invoice_image)} alt={`Invoice ${inv.invoice_number || ''}`} className="h-4 w-4 rounded object-cover bg-muted" />
                              </a>
                            )}
                          </span>
                        ))}
                      </span>
                    ) : '—'}
                  </td>
                  <td className="py-2 pr-3 text-right tabular-nums">{p.quantity}</td>
                  <td className="py-2 pr-3 text-right tabular-nums">{formatUSD(p.unit_cost)}</td>
                  <td className="py-2 pr-3 text-right tabular-nums">{formatUSD(p.total_cost)}</td>
                  <td className={`py-2 text-right tabular-nums ${p.owed > 0 ? 'font-medium text-destructive' : 'text-muted-foreground'}`}>{p.owed > 0 ? formatUSD(p.owed) : 'paid'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

function Kpi({ label, value, alert }: { label: string; value: string; alert?: boolean }) {
  return (
    <div className="rounded-lg bg-card p-4 shadow-sm">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={`mt-1 text-xl font-bold tabular-nums ${alert ? 'text-destructive' : ''}`}>{value}</p>
    </div>
  )
}
