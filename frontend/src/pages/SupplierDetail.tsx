import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { formatUSD } from '@/utils/currency'
import { formatDate } from '@/utils/date'
import { useSupplier, useSupplierPurchases, usePaySupplier } from '@/hooks/useSuppliers'

export function SupplierDetail() {
  const { id } = useParams<{ id: string }>()
  const supplierId = parseInt(id || '0')
  const { data: supplier, isLoading } = useSupplier(supplierId)
  const { data: purchases } = useSupplierPurchases(supplierId)
  const pay = usePaySupplier()
  const [payAmount, setPayAmount] = useState('')

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!supplier) return <p className="text-sm text-destructive">Supplier not found</p>

  const doPay = () => {
    const amt = parseFloat(payAmount)
    if (isNaN(amt) || amt <= 0) return
    pay.mutate({ id: supplierId, amount: amt }, { onSuccess: () => setPayAmount('') })
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
          {supplier.outstanding > 0 ? (
            <>
              <p className="mb-2 text-xs text-muted-foreground">Owed {formatUSD(supplier.outstanding)}. Payment applies to oldest unpaid purchases first.</p>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">$</span>
                <Input value={payAmount} onChange={(e) => setPayAmount(e.target.value)} type="number" step="0.01" min="0" className="w-32" />
                <Button size="sm" onClick={() => setPayAmount(String(supplier.outstanding))} variant="ghost">Pay all</Button>
                <Button size="sm" disabled={!payAmount || pay.isPending} onClick={doPay}>{pay.isPending ? 'Saving…' : 'Pay'}</Button>
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
          <table className="w-full min-w-[520px] text-sm">
            <thead>
              <tr className="border-b text-left text-xs text-muted-foreground">
                <th className="py-2 pr-3 font-medium">Date</th>
                <th className="py-2 pr-3 font-medium">Product</th>
                <th className="py-2 pr-3 text-right font-medium">Qty</th>
                <th className="py-2 pr-3 text-right font-medium">Unit</th>
                <th className="py-2 pr-3 text-right font-medium">Total</th>
                <th className="py-2 text-right font-medium">Owed</th>
              </tr>
            </thead>
            <tbody>
              {!purchases || purchases.length === 0 ? (
                <tr><td colSpan={6} className="py-6 text-center text-muted-foreground">No purchases from this supplier yet</td></tr>
              ) : purchases.map((p) => (
                <tr key={p.batch_id} className="border-b last:border-0">
                  <td className="py-2 pr-3 tabular-nums text-muted-foreground">{formatDate(p.received_at)}</td>
                  <td className="py-2 pr-3">{p.product_name || '—'}</td>
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
