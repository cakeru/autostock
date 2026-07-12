import { cn } from '@/lib/utils'

interface StockBadgeProps {
  quantity: number
  minAlert: number
  reserved?: number
}

export function StockBadge({ quantity, minAlert, reserved = 0 }: StockBadgeProps) {
  const available = quantity - reserved
  const status = available <= 0 ? 'out' : available < minAlert ? 'low' : 'ok'

  return (
    <span className="inline-flex items-center gap-1">
      <span
        className={cn(
          'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
          status === 'out' && 'bg-destructive/10 text-destructive',
          status === 'low' && 'bg-accent/10 text-accent',
          status === 'ok' && 'bg-primary/10 text-primary'
        )}
        title={reserved > 0 ? `${quantity} on hand · ${reserved} reserved for jobs · ${available} available to sell` : undefined}
      >
        {quantity}
      </span>
      {reserved > 0 && (
        <span className="text-[10px] text-muted-foreground" title={`${reserved} reserved for scheduled jobs`}>
          −{reserved} held
        </span>
      )}
    </span>
  )
}
