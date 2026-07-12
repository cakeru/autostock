import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { TableCard, Th } from '@/components/ui/table'
import { formatUSD } from '@/utils/currency'
import {
  usePurchaseOrder, useAddPOItem, useRemovePOItem,
  usePlacePO, useCancelPO, useReceivePO,
} from '@/hooks/usePurchaseOrders'
import { useProducts } from '@/hooks/useProducts'
import type { POStatus } from '@/types/purchaseorder'

const STATUS_STYLES: Record<POStatus, string> = {
  draft: 'bg-muted text-muted-foreground',
  ordered: 'bg-accent/10 text-accent',
  partial: 'bg-accent/10 text-accent',
  received: 'bg-primary/10 text-primary',
  cancelled: 'bg-muted text-muted-foreground line-through',
}

export function PurchaseOrderDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const poId = parseInt(id || '0')

  const { data: po, isLoading } = usePurchaseOrder(poId)
  const { data: productsData } = useProducts({ per_page: 100 })
  const addItemMutation = useAddPOItem(poId)
  const removeItemMutation = useRemovePOItem(poId)
  const placeMutation = usePlacePO()
  const cancelMutation = useCancelPO()
  const receiveMutation = useReceivePO()

  const [productFilter, setProductFilter] = useState('')
  const [selectedProductId, setSelectedProductId] = useState('')
  const [qtyOrdered, setQtyOrdered] = useState('1')
  const [unitCost, setUnitCost] = useState('')
  const [receiveDrafts, setReceiveDrafts] = useState<Record<number, string>>({})
  const [markPaid, setMarkPaid] = useState(false)
  const [confirmCancel, setConfirmCancel] = useState(false)

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!po) return <p className="text-sm text-destructive">Purchase order not found</p>

  const isDraft = po.status === 'draft'
  const canReceive = po.status === 'ordered' || po.status === 'partial'
  const canCancel = po.status === 'draft' || po.status === 'ordered'

  const products = (productsData?.data || []).filter((p) =>
    !productFilter || p.name.toLowerCase().includes(productFilter.toLowerCase()) || p.sku.toLowerCase().includes(productFilter.toLowerCase())
  )

  const handleAddItem = () => {
    const pid = parseInt(selectedProductId)
    const qty = parseInt(qtyOrdered)
    const cost = parseFloat(unitCost)
    if (!pid || !qty || qty < 1 || isNaN(cost) || cost < 0) return
    addItemMutation.mutate({ product_id: pid, quantity_ordered: qty, unit_cost: cost }, {
      onSuccess: () => { setSelectedProductId(''); setProductFilter(''); setQtyOrdered('1'); setUnitCost('') },
    })
  }

  const handleReceive = () => {
    const items = po.items
      .map((it) => {
        const remaining = it.quantity_ordered - it.quantity_received
        const raw = receiveDrafts[it.id]
        const qty = raw === undefined ? remaining : parseInt(raw) || 0
        return { item_id: it.id, quantity: qty }
      })
      .filter((l) => l.quantity > 0)
    if (items.length === 0) return
    receiveMutation.mutate({ id: poId, data: { items, paid: markPaid } }, {
      onSuccess: () => setReceiveDrafts({}),
    })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={po.po_number}
        backTo="/purchase-orders"
        subtitle={`${po.supplier_name}${po.notes ? ' · ' + po.notes : ''}`}
        badges={<span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize ${STATUS_STYLES[po.status]}`}>{po.status}</span>}
        actions={
          <>
            {isDraft && <Button size="sm" onClick={() => placeMutation.mutate(poId)} disabled={po.item_count === 0 || placeMutation.isPending}>Place Order</Button>}
            {canCancel && <Button variant="outline" size="sm" onClick={() => setConfirmCancel(true)}>Cancel</Button>}
          </>
        }
      />

      {isDraft && (
        <div className="rounded-lg bg-card p-4 shadow-sm">
          <p className="mb-2 text-sm font-semibold">Add line item</p>
          <div className="flex flex-wrap items-end gap-2">
            <div className="flex-1 min-w-[160px] space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Search</label>
              <Input value={productFilter} onChange={(e) => setProductFilter(e.target.value)} placeholder="Name or SKU..." />
            </div>
            <div className="flex-1 min-w-[200px] space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Product</label>
              <Select value={selectedProductId} onChange={(e) => {
                setSelectedProductId(e.target.value)
                const p = products.find((x) => x.id === parseInt(e.target.value))
                if (p && !unitCost) setUnitCost(p.buy_price.toString())
              }}>
                <option value="">Select...</option>
                {products.map((p) => <option key={p.id} value={p.id}>{p.name} ({p.sku})</option>)}
              </Select>
            </div>
            <div className="w-24 space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Qty</label>
              <Input type="number" min="1" value={qtyOrdered} onChange={(e) => setQtyOrdered(e.target.value)} />
            </div>
            <div className="w-28 space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Unit Cost</label>
              <Input type="number" min="0" step="0.01" value={unitCost} onChange={(e) => setUnitCost(e.target.value)} />
            </div>
            <Button onClick={handleAddItem} disabled={!selectedProductId || addItemMutation.isPending}>Add</Button>
          </div>
        </div>
      )}

      <TableCard>
        <table className="w-full text-sm">
          <thead>
            <tr className="sticky top-0 z-10 border-b bg-card">
              <Th>Product</Th>
              <Th className="text-right">Ordered</Th>
              <Th className="text-right">Received</Th>
              <Th className="text-right">Unit Cost</Th>
              <Th className="text-right">Total</Th>
              {canReceive && <Th className="text-right">Receive now</Th>}
              {isDraft && <Th className="text-right">Actions</Th>}
            </tr>
          </thead>
          <tbody>
            {po.items.length === 0 ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">No line items yet</td></tr>
            ) : po.items.map((item) => {
              const remaining = item.quantity_ordered - item.quantity_received
              return (
                <tr key={item.id} className="border-b last:border-0">
                  <td className="px-4 py-2.5">
                    <p className="font-medium">{item.product_name}</p>
                    <p className="font-mono text-xs text-muted-foreground">{item.sku}</p>
                  </td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{item.quantity_ordered}</td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{item.quantity_received}</td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{formatUSD(item.unit_cost)}</td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{formatUSD(item.total_cost)}</td>
                  {canReceive && (
                    <td className="px-4 py-2.5 text-right">
                      {remaining > 0 ? (
                        <Input
                          type="number"
                          min="0"
                          max={remaining}
                          className="ml-auto w-20 text-right"
                          value={receiveDrafts[item.id] ?? remaining.toString()}
                          onChange={(e) => setReceiveDrafts((d) => ({ ...d, [item.id]: e.target.value }))}
                        />
                      ) : (
                        <span className="text-xs text-muted-foreground">complete</span>
                      )}
                    </td>
                  )}
                  {isDraft && (
                    <td className="px-4 py-2.5 text-right">
                      <Button variant="ghost" size="icon" onClick={() => removeItemMutation.mutate(item.id)} aria-label="Remove" title="Remove line">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </td>
                  )}
                </tr>
              )
            })}
          </tbody>
        </table>
      </TableCard>

      {canReceive && po.items.length > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg bg-card p-4 shadow-sm">
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={markPaid} onChange={(e) => setMarkPaid(e.target.checked)} className="accent-primary" />
            Paid on delivery <span className="text-xs text-muted-foreground">(uncheck to record as owed to the supplier)</span>
          </label>
          <Button onClick={handleReceive} disabled={receiveMutation.isPending}>
            {receiveMutation.isPending ? 'Receiving...' : 'Receive'}
          </Button>
        </div>
      )}

      <ConfirmDialog
        open={confirmCancel}
        onClose={() => setConfirmCancel(false)}
        onConfirm={() => cancelMutation.mutate(poId, {
          onSuccess: () => { setConfirmCancel(false); navigate('/purchase-orders') },
        })}
        title="Cancel Purchase Order"
        message="Cancel this order? This cannot be undone."
        destructive
        loading={cancelMutation.isPending}
      />
    </div>
  )
}
