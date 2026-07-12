import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ShoppingBag } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { TableCard, Th } from '@/components/ui/table'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { formatUSD } from '@/utils/currency'
import { usePurchaseOrders, useCreatePurchaseOrder } from '@/hooks/usePurchaseOrders'
import { useSuppliers } from '@/hooks/useSuppliers'
import type { POStatus } from '@/types/purchaseorder'

const STATUS_STYLES: Record<POStatus, string> = {
  draft: 'bg-muted text-muted-foreground',
  ordered: 'bg-accent/10 text-accent',
  partial: 'bg-accent/10 text-accent',
  received: 'bg-primary/10 text-primary',
  cancelled: 'bg-muted text-muted-foreground line-through',
}

export function PurchaseOrders() {
  const navigate = useNavigate()
  const { data: pos, isLoading } = usePurchaseOrders()
  const createMutation = useCreatePurchaseOrder()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div className="space-y-6">
      <PageHeader
        title="Purchase Orders"
        subtitle="Reorder from suppliers, then receive against the order to update stock and payables"
        actions={<Button size="sm" onClick={() => setShowCreate(true)}>New Purchase Order</Button>}
      />

      {isLoading ? (
        <div className="bg-card rounded-lg shadow-sm p-4">
          <TableSkeleton rows={5} cols={5} />
        </div>
      ) : !pos || pos.length === 0 ? (
        <div className="rounded-lg bg-card p-10 text-center text-sm text-muted-foreground shadow-sm">
          <ShoppingBag className="mx-auto mb-2 h-8 w-8 text-muted-foreground/50" />
          No purchase orders yet
        </div>
      ) : (
        <TableCard>
          <table className="w-full text-sm">
            <thead>
              <tr className="sticky top-0 z-10 border-b bg-card">
                <Th>PO #</Th>
                <Th>Supplier</Th>
                <Th>Status</Th>
                <Th className="text-right">Items</Th>
                <Th className="text-right">Total</Th>
                <Th>Date</Th>
              </tr>
            </thead>
            <tbody>
              {pos.map((po) => (
                <tr
                  key={po.id}
                  onClick={() => navigate(`/purchase-orders/${po.id}`)}
                  className="cursor-pointer border-b last:border-0 transition-colors hover:bg-muted/50"
                >
                  <td className="px-4 py-2.5 font-mono text-xs">{po.po_number}</td>
                  <td className="px-4 py-2.5">{po.supplier_name}</td>
                  <td className="px-4 py-2.5">
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize ${STATUS_STYLES[po.status]}`}>
                      {po.status}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{po.item_count}</td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{formatUSD(po.total_cost)}</td>
                  <td className="px-4 py-2.5 text-muted-foreground">{new Date(po.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </TableCard>
      )}

      {showCreate && (
        <CreatePODialog
          onClose={() => setShowCreate(false)}
          onCreate={(data) => createMutation.mutate(data, {
            onSuccess: (po) => { setShowCreate(false); navigate(`/purchase-orders/${po.id}`) },
          })}
          loading={createMutation.isPending}
        />
      )}
    </div>
  )
}

function CreatePODialog({ onClose, onCreate, loading }: {
  onClose: () => void
  onCreate: (data: { supplier_id: number; notes?: string }) => void
  loading: boolean
}) {
  const { data: suppliers } = useSuppliers()
  const [supplierId, setSupplierId] = useState('')
  const [notes, setNotes] = useState('')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-3 text-sm font-semibold">New Purchase Order</p>
        <div className="space-y-3">
          <div className="space-y-1.5">
            <Label>Supplier</Label>
            <Select value={supplierId} onChange={(e) => setSupplierId(e.target.value)}>
              <option value="">Select...</option>
              {(suppliers || []).map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label>Notes</Label>
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="optional" />
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button
            type="button"
            onClick={() => onCreate({ supplier_id: parseInt(supplierId), notes: notes || undefined })}
            disabled={!supplierId || loading}
          >
            {loading ? 'Creating...' : 'Create'}
          </Button>
        </div>
      </div>
    </div>
  )
}
