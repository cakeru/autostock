import type { Batch } from '@/types/product'

// A physical read-out for bulk stock (oil, coolant) drawn from drums. The
// "open" drum is the oldest batch still holding liquid — the one with the tap
// on it — drawn as a filling steel drum; full drums behind it are shown as
// reserve. Fill *height* is how full that open drum is; fill *colour* is
// reorder urgency across all drums, so a near-empty open drum with spares
// behind it doesn't cry wolf.

const LEVEL = { ok: '#059669', low: '#d97706', out: '#dc2626' }

function fmt(n: number): string {
  // Trim trailing zeros: 208, 4.5, 12.25
  return Number(n.toFixed(2)).toLocaleString()
}

export function BarrelGauge({ batches, unit, minStockAlert }: { batches: Batch[]; unit?: string; minStockAlert?: number }) {
  const u = unit && unit !== 'piece' ? unit : 'L'
  const live = batches.filter((b) => b.quantity_remaining > 0)
  // Oldest non-empty drum is the one being drawn from (FIFO). API returns
  // batches newest-first, so the oldest live one is the last in that order.
  const open = live.length ? live[live.length - 1] : null
  const spares = Math.max(live.length - 1, 0)
  const spareLiters = live.slice(0, Math.max(live.length - 1, 0)).reduce((s, b) => s + b.quantity_remaining, 0)
  const total = live.reduce((s, b) => s + b.quantity_remaining, 0)

  const capacity = open && open.quantity_received > 0 ? open.quantity_received : 208
  const ratio = open ? Math.max(0, Math.min(1, open.quantity_remaining / capacity)) : 0

  // Reorder urgency across all drums. Fall back to a fraction of one drum when
  // no alert threshold is set for this product.
  const threshold = minStockAlert && minStockAlert > 0 ? minStockAlert : capacity * 0.15
  const level = total <= 0 ? LEVEL.out : total <= threshold ? LEVEL.out : total <= threshold * 2 ? LEVEL.low : LEVEL.ok

  // Drum geometry (viewBox 0 0 170 214).
  const topY = 30
  const botY = 186
  const yTop = botY - ratio * (botY - topY)

  return (
    <div className="flex items-stretch gap-4 rounded-lg border border-border bg-card p-4">
      <svg viewBox="0 0 170 214" className="h-44 w-auto flex-shrink-0" role="img" aria-label={`${fmt(total)} ${u} of bulk stock on hand`}>
        <defs>
          <clipPath id="drum-body">
            <path d="M30 30 L30 186 A50 13 0 0 0 130 186 L130 30 Z" />
          </clipPath>
        </defs>

        {/* Back rim */}
        <ellipse cx="80" cy={topY} rx="50" ry="13" fill="none" stroke="var(--color-primary)" strokeOpacity="0.35" strokeWidth="1.5" />

        {/* Liquid (clipped to the drum body) */}
        <g clipPath="url(#drum-body)">
          <rect x="30" y={yTop} width="100" height={botY + 13 - yTop} fill={level} fillOpacity="0.85" />
          {ratio > 0 && ratio < 1 && (
            <ellipse cx="80" cy={yTop} rx="50" ry="13" fill={level} />
          )}
        </g>

        {/* Drum body outline */}
        <path d="M30 30 L30 186 A50 13 0 0 0 130 186 L130 30" fill="none" stroke="var(--color-primary)" strokeWidth="2" />
        <ellipse cx="80" cy={topY} rx="50" ry="13" fill="none" stroke="var(--color-primary)" strokeWidth="2" />
        {/* Rolling hoops */}
        <ellipse cx="80" cy="82" rx="50" ry="13" fill="none" stroke="var(--color-primary)" strokeOpacity="0.4" strokeWidth="1.5" />
        <ellipse cx="80" cy="134" rx="50" ry="13" fill="none" stroke="var(--color-primary)" strokeOpacity="0.4" strokeWidth="1.5" />

        {/* Measurement scale */}
        {[0, 0.25, 0.5, 0.75, 1].map((f) => {
          const y = botY - f * (botY - topY)
          return (
            <g key={f}>
              <line x1="132" y1={y} x2="140" y2={y} stroke="var(--color-brand-gold)" strokeWidth="1.5" />
              <text x="144" y={y + 3} fontSize="8" fill="var(--color-brand-gold)" className="tabular-nums">
                {f === 0 ? '0' : f === 1 ? fmt(capacity) : `${f * 100}%`}
              </text>
            </g>
          )
        })}
      </svg>

      <div className="flex min-w-0 flex-1 flex-col justify-center">
        <div className="flex items-baseline gap-1.5">
          <span className="text-3xl font-semibold tabular-nums" style={{ color: level }}>{fmt(total)}</span>
          <span className="text-sm text-muted-foreground">{u} on hand</span>
        </div>

        {open ? (
          <p className="mt-1 text-sm text-muted-foreground">
            Open drum: <span className="font-medium tabular-nums text-foreground">{fmt(open.quantity_remaining)}</span>
            <span className="tabular-nums"> / {fmt(capacity)} {u}</span>
            <span className="ml-1.5 tabular-nums">({Math.round(ratio * 100)}%)</span>
          </p>
        ) : (
          <p className="mt-1 text-sm font-medium" style={{ color: LEVEL.out }}>No stock — every drum is empty</p>
        )}

        <p className="mt-0.5 text-xs text-muted-foreground">
          {spares > 0
            ? <>＋{spares} sealed drum{spares > 1 ? 's' : ''} in reserve <span className="tabular-nums">({fmt(spareLiters)} {u})</span></>
            : open ? 'Last drum open — no spares behind it' : ''}
        </p>

        {level !== LEVEL.ok && total > 0 && (
          <span className="mt-2 inline-flex w-fit items-center rounded-full px-2 py-0.5 text-xs font-medium" style={{ backgroundColor: `${level}1a`, color: level }}>
            {level === LEVEL.out ? 'Reorder now' : 'Running low'}
          </span>
        )}
      </div>
    </div>
  )
}
