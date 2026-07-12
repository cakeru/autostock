import { lazy, Suspense, useState } from 'react'
import { ArrowDownLeft, ArrowUpRight, ChevronRight, Package, QrCode } from 'lucide-react'
import { SlideOver } from '@/components/ui/SlideOver'
import { Button } from '@/components/ui/button'
import { useProductMovements, useProductBatches, useBatchConsumers } from '@/hooks/useProducts'
import { useSettings } from '@/hooks/useSettings'
import { BarrelGauge } from './BarrelGauge'
// Lazy — the QR generator only loads when someone prints batch labels.
const BatchLabelSheet = lazy(() => import('./BatchLabelSheet').then(m => ({ default: m.BatchLabelSheet })))
import { formatDateTime, formatDate } from '@/utils/date'
import { formatUSD } from '@/utils/currency'
import type { Product, StockMovement, Batch } from '@/types/product'

function describe(m: StockMovement): { label: string; detail?: string } {
  switch (m.reason) {
    case 'received': return { label: 'Stock received' }
    case 'opening': return { label: 'Opening stock' }
    case 'invoice_issued': return { label: 'Sold', detail: m.invoice_number }
    case 'invoice_voided': return { label: 'Returned (voided)', detail: m.invoice_number }
    default: return { label: 'Adjustment', detail: m.reason }
  }
}

function MovementRow({ m }: { m: StockMovement }) {
  const isIn = m.quantity_change >= 0
  const { label, detail } = describe(m)
  return (
    <div className="flex items-center gap-3 py-3">
      <div className={`grid h-8 w-8 flex-shrink-0 place-items-center rounded-full ${isIn ? 'bg-primary/10 text-primary' : 'bg-destructive/10 text-destructive'}`}>
        {isIn ? <ArrowDownLeft className="h-4 w-4" /> : <ArrowUpRight className="h-4 w-4" />}
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium">
          {label}
          {detail && <span className="ml-1.5 font-mono text-xs font-normal text-muted-foreground">{detail}</span>}
        </p>
        <p className="text-xs text-muted-foreground">
          {formatDateTime(m.created_at)}
          {m.recorded_by_name && <> · by {m.recorded_by_name}</>}
          {m.batch_no && <> · <span className="font-mono">{m.batch_no}</span></>}
        </p>
      </div>
      <div className="flex-shrink-0 text-right">
        <p className={`text-sm font-semibold tabular-nums ${isIn ? 'text-primary' : 'text-destructive'}`}>
          {isIn ? '+' : ''}{m.quantity_change}
        </p>
        <p className="text-xs text-muted-foreground tabular-nums">bal {m.balance_after}</p>
      </div>
    </div>
  )
}

function BatchRow({ batch, onPrintLabels }: { batch: Batch; onPrintLabels?: () => void }) {
  const [open, setOpen] = useState(false)
  const { data: consumers, isLoading } = useBatchConsumers(open ? batch.id : null)
  const depleted = batch.quantity_remaining <= 0
  return (
    <div className="py-3">
      <button onClick={() => setOpen(!open)} className="flex w-full items-center gap-3 text-left">
        <div className="grid h-8 w-8 flex-shrink-0 place-items-center rounded-full bg-muted text-muted-foreground">
          <Package className="h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium">
            <span className="font-mono">{batch.batch_no}</span>
            {batch.supplier && <span className="ml-1.5 font-normal text-muted-foreground">{batch.supplier}</span>}
          </p>
          <p className="text-xs text-muted-foreground">
            {formatDate(batch.received_at)}
            {batch.dot_code && <> · DOT <span className="font-mono">{batch.dot_code}</span></>}
            {batch.unit_cost > 0 && <> · {formatUSD(batch.unit_cost)}/unit</>}
            {batch.received_by_name && <> · by {batch.received_by_name}</>}
          </p>
        </div>
        <div className="flex flex-shrink-0 items-center gap-2 text-right">
          {onPrintLabels && (
            <span
              role="button"
              tabIndex={0}
              onClick={(e) => { e.stopPropagation(); onPrintLabels() }}
              onKeyDown={(e) => { if (e.key === 'Enter') { e.stopPropagation(); onPrintLabels() } }}
              title="Print scan labels for this batch"
              className="grid h-7 w-7 place-items-center rounded-md border text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <QrCode className="h-3.5 w-3.5" />
            </span>
          )}
          <div>
            <p className={`text-sm font-semibold tabular-nums ${depleted ? 'text-muted-foreground' : ''}`}>
              {batch.quantity_remaining}<span className="text-xs font-normal text-muted-foreground">/{batch.quantity_received}</span>
            </p>
            <p className="text-[10px] uppercase tracking-wide text-muted-foreground">left</p>
          </div>
          <ChevronRight className={`h-4 w-4 text-muted-foreground transition-transform ${open ? 'rotate-90' : ''}`} />
        </div>
      </button>

      {open && (
        <div className="ml-11 mt-2 rounded-md bg-muted/40 p-2">
          <p className="mb-1 text-xs font-medium text-muted-foreground">Sold to</p>
          {isLoading ? (
            <p className="py-1 text-xs text-muted-foreground">Loading…</p>
          ) : !consumers || consumers.length === 0 ? (
            <p className="py-1 text-xs text-muted-foreground">No units from this batch have been sold yet</p>
          ) : (
            <div className="divide-y divide-border/60">
              {consumers.map((c) => (
                <div key={c.invoice_id} className="flex items-center justify-between py-1.5 text-sm">
                  <span className="min-w-0 truncate">
                    <span className="font-mono text-xs text-muted-foreground">{c.invoice_number}</span>
                    <span className="ml-2">{c.customer_name}</span>
                  </span>
                  <span className="flex-shrink-0 tabular-nums text-muted-foreground">×{c.quantity}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export function StockHistory({ product, onClose }: { product: Product | null; onClose: () => void }) {
  const [tab, setTab] = useState<'movements' | 'batches'>('movements')
  const [page, setPage] = useState(1)
  const [labelBatch, setLabelBatch] = useState<Batch | null>(null)
  const { data: settings } = useSettings()
  const { data: mData, isLoading: mLoading } = useProductMovements(product?.id ?? 0, page)
  const { data: batches, isLoading: bLoading } = useProductBatches(product?.id ?? 0, tab === 'batches' || !!product?.is_bulk)
  const movements = mData?.data || []
  const meta = mData?.meta
  const scanOn = !!settings?.feature_batch_scan

  return (
    <SlideOver open={!!product} onClose={onClose} title="Stock history & batches">
      {product && (
        <div className="mb-3">
          <p className="text-sm font-medium">{product.name}</p>
          <p className="font-mono text-xs text-muted-foreground">{product.sku}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            On hand: <span className="font-medium text-foreground tabular-nums">{product.stock_quantity}</span>
            {product.unit && product.unit !== 'piece' && <span className="text-muted-foreground"> {product.unit}</span>}
          </p>
          {product.is_bulk && batches && (
            <div className="mt-3">
              <BarrelGauge batches={batches} unit={product.unit} minStockAlert={product.min_stock_alert} />
            </div>
          )}
        </div>
      )}

      <div className="mb-1 flex gap-1 border-b">
        {(['movements', 'batches'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`-mb-px border-b-2 px-3 py-2 text-sm font-medium capitalize transition-colors ${
              tab === t ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            {t === 'movements' ? 'Movements' : 'Intake batches'}
          </button>
        ))}
      </div>

      {tab === 'movements' ? (
        mLoading ? (
          <p className="py-8 text-center text-sm text-muted-foreground">Loading…</p>
        ) : movements.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">No stock movements recorded yet</p>
        ) : (
          <>
            <div className="divide-y divide-border/70">
              {movements.map((m) => <MovementRow key={m.id} m={m} />)}
            </div>
            {meta && meta.total_pages > 1 && (
              <div className="mt-3 flex items-center justify-between border-t pt-3 text-sm">
                <span className="tabular-nums text-muted-foreground">page {meta.page} of {meta.total_pages}</span>
                <div className="flex gap-1">
                  <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
                  <Button variant="outline" size="sm" disabled={page >= meta.total_pages} onClick={() => setPage(page + 1)}>Next</Button>
                </div>
              </div>
            )}
          </>
        )
      ) : bLoading ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Loading…</p>
      ) : !batches || batches.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">No intake batches yet</p>
      ) : (
        <div className="divide-y divide-border/70">
          {batches.map((b) => <BatchRow key={b.id} batch={b} onPrintLabels={scanOn ? () => setLabelBatch(b) : undefined} />)}
        </div>
      )}

      {labelBatch && product && (
        <Suspense fallback={null}>
          <BatchLabelSheet product={product} batch={labelBatch} onClose={() => setLabelBatch(null)} />
        </Suspense>
      )}
    </SlideOver>
  )
}
