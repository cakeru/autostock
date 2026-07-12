import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Trash2, CheckCircle2, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { TableCard, Th } from '@/components/ui/table'
import {
  useStocktake, useAddStocktakeItem, useSetStocktakeCount,
  useRemoveStocktakeItem, useCancelStocktake, useFinalizeStocktake,
} from '@/hooks/useStocktakes'
import { useProducts } from '@/hooks/useProducts'
import type { FinalizeResult } from '@/types/stocktake'

const STATUS_STYLES = {
  draft: 'bg-accent/10 text-accent',
  completed: 'bg-primary/10 text-primary',
  cancelled: 'bg-muted text-muted-foreground',
}

export function StocktakeDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const stocktakeId = parseInt(id || '0')

  const { data: st, isLoading } = useStocktake(stocktakeId)
  const { data: productsData } = useProducts({ per_page: 100 })
  const addItemMutation = useAddStocktakeItem(stocktakeId)
  const setCountMutation = useSetStocktakeCount(stocktakeId)
  const removeItemMutation = useRemoveStocktakeItem(stocktakeId)
  const cancelMutation = useCancelStocktake()
  const finalizeMutation = useFinalizeStocktake()

  const [productFilter, setProductFilter] = useState('')
  const [selectedProductId, setSelectedProductId] = useState('')
  const [countDrafts, setCountDrafts] = useState<Record<number, string>>({})
  const [confirmCancel, setConfirmCancel] = useState(false)
  const [confirmFinalize, setConfirmFinalize] = useState(false)
  const [finalizeResult, setFinalizeResult] = useState<FinalizeResult | null>(null)

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!st) return <p className="text-sm text-destructive">Stocktake not found</p>

  const isDraft = st.status === 'draft'
  const products = (productsData?.data || []).filter((p) =>
    !productFilter || p.name.toLowerCase().includes(productFilter.toLowerCase()) || p.sku.toLowerCase().includes(productFilter.toLowerCase())
  )
  const availableProducts = products.filter((p) => !st.items.some((i) => i.product_id === p.id))

  const handleAddItem = () => {
    const pid = parseInt(selectedProductId)
    if (!pid) return
    addItemMutation.mutate(pid, {
      onSuccess: () => { setSelectedProductId(''); setProductFilter('') },
    })
  }

  const commitCount = (itemId: number) => {
    const raw = countDrafts[itemId]
    if (raw === undefined || raw === '') return
    const qty = parseInt(raw)
    if (isNaN(qty) || qty < 0) return
    setCountMutation.mutate({ itemId, countedQty: qty })
  }

  const handleFinalize = () => {
    finalizeMutation.mutate(stocktakeId, {
      onSuccess: (result) => { setFinalizeResult(result); setConfirmFinalize(false) },
    })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Stocktake #${st.id}`}
        backTo="/stocktakes"
        subtitle={st.notes}
        badges={<span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize ${STATUS_STYLES[st.status]}`}>{st.status}</span>}
        actions={isDraft ? (
          <>
            <Button variant="outline" size="sm" onClick={() => setConfirmCancel(true)}>Cancel Stocktake</Button>
            <Button size="sm" onClick={() => setConfirmFinalize(true)} disabled={st.counted_count === 0}>Finalize</Button>
          </>
        ) : undefined}
      />

      {finalizeResult && (
        <div className="rounded-lg bg-card p-4 shadow-sm">
          <p className="mb-2 text-sm font-semibold">Finalize result</p>
          <div className="flex flex-wrap items-center gap-4 text-sm">
            <span className="flex items-center gap-1.5 text-primary"><CheckCircle2 className="h-4 w-4" /> {finalizeResult.adjusted} adjusted</span>
            <span className="text-muted-foreground">{finalizeResult.unchanged} unchanged (matched)</span>
            {finalizeResult.skipped > 0 && (
              <span className="flex items-center gap-1.5 text-destructive"><XCircle className="h-4 w-4" /> {finalizeResult.skipped} skipped</span>
            )}
          </div>
          {finalizeResult.errors && finalizeResult.errors.length > 0 && (
            <div className="mt-2 space-y-1 text-xs">
              {finalizeResult.errors.map((e, i) => (
                <p key={i} className="text-destructive">{e.sku}: {e.message}</p>
              ))}
            </div>
          )}
        </div>
      )}

      {isDraft && (
        <div className="rounded-lg bg-card p-4 shadow-sm">
          <p className="mb-2 text-sm font-semibold">Add product to count sheet</p>
          <div className="flex flex-wrap items-end gap-2">
            <div className="flex-1 min-w-[180px] space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Search</label>
              <Input value={productFilter} onChange={(e) => setProductFilter(e.target.value)} placeholder="Name or SKU..." />
            </div>
            <div className="flex-1 min-w-[220px] space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Product</label>
              <Select value={selectedProductId} onChange={(e) => setSelectedProductId(e.target.value)}>
                <option value="">Select...</option>
                {availableProducts.map((p) => (
                  <option key={p.id} value={p.id}>{p.name} ({p.sku}) — on hand: {p.stock_quantity}</option>
                ))}
              </Select>
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
              <Th className="text-right">Expected</Th>
              <Th className="text-right">Counted</Th>
              <Th className="text-right">Variance</Th>
              {isDraft && <Th className="text-right">Actions</Th>}
            </tr>
          </thead>
          <tbody>
            {st.items.length === 0 ? (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">No products on this sheet yet</td></tr>
            ) : st.items.map((item) => (
              <tr key={item.id} className="border-b last:border-0">
                <td className="px-4 py-2.5">
                  <p className="font-medium">{item.product_name}</p>
                  <p className="font-mono text-xs text-muted-foreground">{item.sku}</p>
                </td>
                <td className="px-4 py-2.5 text-right tabular-nums text-muted-foreground">{item.expected_qty}</td>
                <td className="px-4 py-2.5 text-right">
                  {isDraft ? (
                    <Input
                      type="number"
                      min="0"
                      className="ml-auto w-24 text-right"
                      value={countDrafts[item.id] ?? item.counted_qty?.toString() ?? ''}
                      onChange={(e) => setCountDrafts((d) => ({ ...d, [item.id]: e.target.value }))}
                      onBlur={() => commitCount(item.id)}
                      onKeyDown={(e) => { if (e.key === 'Enter') (e.target as HTMLInputElement).blur() }}
                      placeholder="—"
                    />
                  ) : (
                    <span className="tabular-nums">{item.counted_qty ?? '—'}</span>
                  )}
                </td>
                <td className="px-4 py-2.5 text-right tabular-nums">
                  {item.variance === undefined || item.variance === null ? (
                    <span className="text-muted-foreground">—</span>
                  ) : item.variance === 0 ? (
                    <span className="text-muted-foreground">0</span>
                  ) : (
                    <span className={item.variance > 0 ? 'text-primary' : 'text-destructive'}>
                      {item.variance > 0 ? `+${item.variance}` : item.variance}
                    </span>
                  )}
                </td>
                {isDraft && (
                  <td className="px-4 py-2.5 text-right">
                    <Button variant="ghost" size="icon" onClick={() => removeItemMutation.mutate(item.id)} aria-label="Remove" title="Remove from sheet">
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </TableCard>

      <ConfirmDialog
        open={confirmCancel}
        onClose={() => setConfirmCancel(false)}
        onConfirm={() => cancelMutation.mutate(stocktakeId, {
          onSuccess: () => { setConfirmCancel(false); navigate('/stocktakes') },
        })}
        title="Cancel Stocktake"
        message="Discard this count sheet without applying any adjustments? This cannot be undone."
        destructive
        loading={cancelMutation.isPending}
      />

      <ConfirmDialog
        open={confirmFinalize}
        onClose={() => setConfirmFinalize(false)}
        onConfirm={handleFinalize}
        title="Finalize Stocktake"
        message={`Apply counted quantities as real stock adjustments for ${st.counted_count} counted product(s)? Uncounted lines are left untouched. This locks the sheet and cannot be undone.`}
        loading={finalizeMutation.isPending}
      />
    </div>
  )
}
