import { Pencil, Trash2, PackagePlus, Scale, History } from 'lucide-react'
import { StockBadge } from '@/components/inventory/StockBadge'
import { ProductThumb, productSpec } from '@/components/inventory/ProductThumb'
import { formatUSD } from '@/utils/currency'
import type { Product } from '@/types/product'

interface ProductCardProps {
  product: Product
  onEdit: () => void
  onReceive: () => void
  onAdjust: () => void
  onHistory: () => void
  onDelete: () => void
}

export function ProductCard({ product: p, onEdit, onReceive, onAdjust, onHistory, onDelete }: ProductCardProps) {
  const spec = productSpec(p)

  return (
    <div className="group flex flex-col overflow-hidden rounded-lg bg-card shadow-sm transition-shadow hover:shadow-md">
      <div className="relative aspect-[4/3] bg-muted">
        <ProductThumb product={p} className="absolute inset-0 h-full w-full rounded-none" />
        <div className="absolute left-2 top-2">
          <StockBadge quantity={p.stock_quantity} minAlert={p.min_stock_alert} reserved={p.reserved_quantity} />
        </div>
        {p.location && (
          <span className="absolute right-2 top-2 rounded bg-black/55 px-1.5 py-0.5 font-mono text-[10px] text-white">
            {p.location}
          </span>
        )}
      </div>

      <div className="flex flex-1 flex-col p-3">
        <p className="truncate text-sm font-medium" title={p.name}>{p.name}</p>
        <p className="font-mono text-xs text-muted-foreground">{p.sku}</p>
        <p className="mt-0.5 truncate text-xs text-muted-foreground">
          {spec || <span className="capitalize">{p.type}</span>}
        </p>
        <div className="mt-auto pt-2 text-lg font-semibold tabular-nums">{formatUSD(p.sell_price)}</div>
      </div>

      <div className="flex divide-x border-t text-muted-foreground">
        <button onClick={onEdit} title="Edit" aria-label="Edit" className="flex flex-1 items-center justify-center py-2 transition-colors hover:bg-muted hover:text-foreground">
          <Pencil className="h-4 w-4" />
        </button>
        <button onClick={onReceive} title="Receive stock" aria-label="Receive stock" className="flex flex-1 items-center justify-center py-2 transition-colors hover:bg-muted hover:text-foreground">
          <PackagePlus className="h-4 w-4" />
        </button>
        <button onClick={onAdjust} title="Adjust stock" aria-label="Adjust stock" className="flex flex-1 items-center justify-center py-2 transition-colors hover:bg-muted hover:text-foreground">
          <Scale className="h-4 w-4" />
        </button>
        <button onClick={onHistory} title="Stock history" aria-label="Stock history" className="flex flex-1 items-center justify-center py-2 transition-colors hover:bg-muted hover:text-foreground">
          <History className="h-4 w-4" />
        </button>
        <button onClick={onDelete} title="Delete product" aria-label="Delete product" className="flex flex-1 items-center justify-center py-2 transition-colors hover:bg-muted hover:text-destructive">
          <Trash2 className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
