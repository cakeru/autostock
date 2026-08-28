import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/button'
import { formatUSD } from '@/utils/currency'
import { formatDate } from '@/utils/date'
import { imageSrc } from '@/utils/imageUrl'
import { useSupplier, useSupplierPurchases, usePaySupplier } from '@/hooks/useSuppliers'

export function SupplierDetail() {
  const { id } = useParams<{ id: string }>()
  const supplierId = parseInt(id || '0')
  const { data: supplier, isLoading } = useSupplier(supplierId)
  const { data: purchases } = useSupplierPurchases(supplierId)
  const pay = usePaySupplier()
  const [selected, setSelected] = useState<Set<number>>(new Set())

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!supplier) return <p className="text-sm text-destructive">Supplier not found</p>

  const unpaid = (purchases || []).filter((p) => p.owed > 0)
  const toggle = (batchId: number) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(batchId)) next.delete(batchId)
      else next.add(batchId)
      return next
    })
  }
  const selectedTotal = unpaid.filter((p) => selected.has(p.batch_id)).reduce((sum, p) => sum + p.owed, 0)

  const doPay = (batchIds: number[]) => {
    if (batchIds.length === 0) return
    pay.mutate({ id: supplierId, batchIds }, { onSuccess: () => setSelected(new Set()) })
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
          {unpaid.length > 0 ? (
            <>
              <p className="mb-2 text-xs text-muted-foreground">Tick the invoices to pay — each is settled in full.</p>
              <div className="max-h-56 space-y-1 overflow-y-auto pr-1">
                {unpaid.map((p) => (
                  <label key={p.batch_id} className="flex cursor-pointer items-center gap-2 rounded border border-border px-2 py-1.5 text-sm hover:bg-muted/50">
                    <input type="checkbox" checked={selected.has(p.batch_id)} onChange={() => toggle(p.batch_id)} className="accent-primary" />
                    <span className="min-w-0 flex-1 truncate">
                      {p.invoice_number ? <span className="font-medium">Invoice {p.invoice_number}</span> : p.product_name || 'Purchase'}
                      <span className="ml-2 text-xs text-muted-foreground">{formatDate(p.received_at)}</span>
                    </span>
                    <span className="tabular-nums text-destructive">{formatUSD(p.owed)}</span>
                  </label>
                ))}
              </div>
              <div className="mt-3 flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground">
                  {selected.size > 0 ? `Selected: ${formatUSD(selectedTotal)}` : 'Nothing selected'}
                </span>
                <div className="flex gap-2">
                  <Button size="sm" variant="ghost" onClick={() => doPay(unpaid.map((p) => p.batch_id))} disabled={pay.isPending}>
                    Pay all
                  </Button>
                  <Button size="sm" disabled={selected.size === 0 || pay.isPending} onClick={() => doPay([...selected])}>
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
                <th className="py-2 pr-3 font-medium">Invoice</th>
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
                    {p.invoice_number || p.invoice_image ? (
                      <span className="flex items-center gap-1.5">
                        {p.invoice_number && <span className="text-muted-foreground">{p.invoice_number}</span>}
                        {p.invoice_image && (
                          <a href={imageSrc(p.invoice_image)} target="_blank" rel="noreferrer" title="View invoice photo">
                            <img src={imageSrc(p.invoice_image)} alt={`Invoice ${p.invoice_number || ''}`} className="h-6 w-6 rounded object-cover bg-muted" />
                          </a>
                        )}
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
