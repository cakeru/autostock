import { useState } from 'react'
import type { CornerData } from '@/types/vehicleProfile'
import { CarBody, normalizeType, wheelGeometry } from './CarDiagram'

const GREEN = '#059669'
type Phase = 'before' | 'after'
type Metric = 'camber' | 'caster' | 'toe'
const METRICS: { key: Metric; label: string }[] = [
  { key: 'camber', label: 'Camber' },
  { key: 'caster', label: 'Caster' },
  { key: 'toe', label: 'Toe' },
]

function val(c: CornerData, metric: Metric, phase: Phase): string {
  return (c[`${metric}_${phase}` as keyof CornerData] as string) || ''
}
function cornerHasAlignment(c?: CornerData): boolean {
  if (!c) return false
  return METRICS.some((m) => val(c, m.key, 'before') || val(c, m.key, 'after'))
}
// Loose equality so "0.40" vs "0.4" or "-1°01'" vs "-1° 01'" don't read as a change.
function normNum(s: string): string {
  return s.replace(/[\s°'"]/g, '').replace(/(\.\d*?)0+$/, '$1').replace(/\.$/, '').toLowerCase()
}

// True when the vehicle's latest wheel data carries any alignment reading —
// the parent uses this to decide whether to show the blueprint at all.
export function hasAlignmentData(corners: Record<string, CornerData>): boolean {
  return Object.values(corners).some(cornerHasAlignment)
}

// A top-down alignment drawing with a Before | After toggle. Each wheel's
// camber / caster / toe reads out in the margin; flipping to After highlights
// the angles that were actually adjusted in green.
export function AlignmentBlueprint({ bodyType, corners, className }: {
  bodyType?: string
  corners: Record<string, CornerData>
  className?: string
}) {
  const [phase, setPhase] = useState<Phase>('before')
  const type = normalizeType(bodyType)
  const wheels = wheelGeometry(type)

  return (
    <div className={className}>
      <div className="mb-3 flex items-center justify-between">
        <p className="text-sm font-semibold">Wheel Alignment</p>
        <div className="inline-flex overflow-hidden rounded-md border text-xs">
          {(['before', 'after'] as Phase[]).map((p) => (
            <button
              key={p}
              onClick={() => setPhase(p)}
              className={`px-3 py-1 font-medium capitalize transition-colors ${phase === p ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'}`}
            >
              {p}
            </button>
          ))}
        </div>
      </div>

      <div className="mx-auto w-full max-w-[360px]" style={{ color: 'var(--color-primary)' }}>
        <svg viewBox="0 0 340 300" role="img" aria-label={`Wheel alignment, ${phase}`}>
          <defs>
            <pattern id="align-grid" width="20" height="20" patternUnits="userSpaceOnUse">
              <path d="M20 0H0V20" fill="none" stroke="currentColor" strokeOpacity="0.07" strokeWidth="0.5" />
            </pattern>
          </defs>
          <rect width="340" height="300" fill="url(#align-grid)" />
          <g strokeWidth="1" style={{ stroke: 'var(--color-brand-gold)' }}>
            <path d="M6 14V6h8M326 6h8v8M6 286v8h8M334 286v8h-8" fill="none" />
          </g>
          <text x="170" y="13" textAnchor="middle" fontSize="7" letterSpacing="3" style={{ fill: 'var(--color-accent)' }}>FRONT</text>
          {/* faint phase watermark down the centre */}
          <text x="170" y="160" textAnchor="middle" fontSize="30" fontWeight="700" letterSpacing="6" fill="currentColor" fillOpacity="0.05">
            {phase.toUpperCase()}
          </text>

          <CarBody type={type} />

          {Object.entries(wheels).map(([pos, g]) => {
            const c = corners[pos]
            const isL = g.side === 'left'
            const edgeX = isL ? g.cx - g.w / 2 : g.cx + g.w / 2
            const wheel = (
              <rect x={g.cx - g.w / 2} y={g.cy - g.h / 2} width={g.w} height={g.h} rx={3}
                fill="currentColor" fillOpacity={0.08} stroke="currentColor" strokeWidth={1.4} />
            )
            if (!cornerHasAlignment(c) || !c) {
              return <g key={pos}>{wheel}</g>
            }
            const leaderEnd = isL ? 108 : 232
            const labelX = isL ? 40 : 300     // metric label (faint), outer column
            const valueX = isL ? 104 : 236    // value (bold), nearest the car
            return (
              <g key={pos}>
                {wheel}
                <line x1={edgeX} y1={g.cy} x2={leaderEnd} y2={g.cy} stroke="currentColor" strokeOpacity="0.35" strokeWidth="0.75" />
                <circle cx={edgeX + (isL ? -1 : 1)} cy={g.cy} r="1.4" fill="currentColor" fillOpacity="0.5" />
                <text x={valueX} y={g.cy - 22} textAnchor={isL ? 'end' : 'start'} fontSize="8" letterSpacing="1" fill="currentColor" fillOpacity="0.55">{g.label}</text>
                {METRICS.map((m, i) => {
                  const before = val(c, m.key, 'before')
                  const after = val(c, m.key, 'after')
                  const shown = phase === 'before' ? before : after
                  const adjusted = phase === 'after' && !!before && !!after && normNum(before) !== normNum(after)
                  const y = g.cy - 9 + i * 11
                  return (
                    <g key={m.key}>
                      <text x={labelX} y={y} textAnchor={isL ? 'start' : 'end'} fontSize="7" fill="currentColor" fillOpacity="0.5">{m.label}</text>
                      <text x={valueX} y={y} textAnchor={isL ? 'end' : 'start'} fontSize="9.5" fontWeight="600"
                        style={{ fontVariantNumeric: 'tabular-nums', fill: adjusted ? GREEN : 'currentColor' }}>
                        {shown || '—'}
                      </text>
                    </g>
                  )
                })}
              </g>
            )
          })}
        </svg>
      </div>

      <p className="mt-2 text-center text-[11px] text-muted-foreground">
        {phase === 'before'
          ? 'Readings when the car arrived — tap After to see the corrections.'
          : 'After adjustment. Angles we changed are shown in green.'}
      </p>
    </div>
  )
}
