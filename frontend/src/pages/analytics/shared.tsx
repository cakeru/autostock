import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import type { ReactNode } from 'react'

export function monthStart() {
  const d = new Date()
  return new Date(d.getFullYear(), d.getMonth(), 1).toISOString().slice(0, 10)
}
export const today = () => new Date().toISOString().slice(0, 10)

export function StatCard({ label, value, sub, accent }: { label: string; value: ReactNode; sub?: ReactNode; accent?: 'up' | 'down' | 'warn' }) {
  const accentClass = accent === 'up' ? 'text-emerald-600' : accent === 'down' ? 'text-destructive' : accent === 'warn' ? 'text-amber-600' : ''
  return (
    <div className="rounded-lg bg-card p-5 shadow-sm">
      <p className="text-xs font-medium text-muted-foreground">{label}</p>
      <p className={`mt-1 text-2xl font-bold tabular-nums ${accentClass}`}>{value}</p>
      {sub != null && <p className="mt-0.5 text-xs text-muted-foreground">{sub}</p>}
    </div>
  )
}

export function Panel({ title, action, children }: { title: string; action?: ReactNode; children: ReactNode }) {
  return (
    <div className="rounded-lg bg-card p-5 shadow-sm">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold">{title}</h2>
        {action}
      </div>
      {children}
    </div>
  )
}

export function RangeBar({ from, to, setFrom, setTo, right }: {
  from: string; to: string; setFrom: (v: string) => void; setTo: (v: string) => void; right?: ReactNode
}) {
  const d = new Date()
  const lastMonthStart = new Date(d.getFullYear(), d.getMonth() - 1, 1).toISOString().slice(0, 10)
  const lastMonthEnd = new Date(d.getFullYear(), d.getMonth(), 0).toISOString().slice(0, 10)
  const yearStart = new Date(d.getFullYear(), 0, 1).toISOString().slice(0, 10)
  const preset = (label: string, f: string, t: string) => (
    <button onClick={() => { setFrom(f); setTo(t) }}
      className={`rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${from === f && to === t ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}>
      {label}
    </button>
  )
  return (
    <div className="flex flex-wrap items-center gap-2">
      {preset('This month', monthStart(), today())}
      {preset('Last month', lastMonthStart, lastMonthEnd)}
      {preset('This year', yearStart, today())}
      <div className="ml-auto flex items-center gap-2">
        <Input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="w-40" />
        <span className="text-sm text-muted-foreground">to</span>
        <Input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="w-40" />
        {right}
      </div>
    </div>
  )
}

export function DaysSelect({ days, setDays }: { days: number; setDays: (n: number) => void }) {
  return (
    <div className="flex items-center gap-2">
      {[30, 90, 180, 365].map((d) => (
        <button key={d} onClick={() => setDays(d)}
          className={`rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${days === d ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}>
          {d === 365 ? '1 year' : `${d} days`}
        </button>
      ))}
    </div>
  )
}

export function pct(n: number) {
  const sign = n > 0 ? '+' : ''
  return `${sign}${n.toFixed(1)}%`
}

export function EmptyRow({ cols, text }: { cols: number; text: string }) {
  return <tr><td colSpan={cols} className="py-6 text-center text-sm text-muted-foreground">{text}</td></tr>
}

export interface Column<T> {
  header: string
  cell: (row: T) => ReactNode
  align?: 'right'
  primary?: boolean // used as the card title on mobile
  tdClass?: string
}

/**
 * Responsive table: a real <table> at sm and up, stacked label/value cards on
 * phones so dense numeric columns never collide. `action` renders a trailing
 * control (e.g. delete) — last column on desktop, top-right on the mobile card.
 */
export function DataTable<T>({ columns, rows, rowKey, empty, action }: {
  columns: Column<T>[]
  rows: T[]
  rowKey: (row: T, i: number) => string | number
  empty: string
  action?: (row: T) => ReactNode
}) {
  if (rows.length === 0) return <p className="py-6 text-center text-sm text-muted-foreground">{empty}</p>
  const primary = columns.find((c) => c.primary)
  const rest = columns.filter((c) => !c.primary)

  return (
    <>
      {/* Desktop */}
      <table className="hidden w-full text-sm sm:table">
        <thead>
          <tr className="border-b text-left text-xs text-muted-foreground">
            {columns.map((c, i) => (
              <th key={i} className={cn('py-2 pr-3 font-medium last:pr-0', c.align === 'right' && 'text-right')}>{c.header}</th>
            ))}
            {action && <th className="py-2" />}
          </tr>
        </thead>
        <tbody>
          {rows.map((r, ri) => (
            <tr key={rowKey(r, ri)} className="border-b last:border-0">
              {columns.map((c, ci) => (
                <td key={ci} className={cn('py-2 pr-3 last:pr-0', c.align === 'right' && 'text-right tabular-nums', c.tdClass)}>{c.cell(r)}</td>
              ))}
              {action && <td className="py-2 text-right">{action(r)}</td>}
            </tr>
          ))}
        </tbody>
      </table>

      {/* Mobile */}
      <div className="space-y-2 sm:hidden">
        {rows.map((r, ri) => (
          <div key={rowKey(r, ri)} className="rounded-lg border p-3">
            {(primary || action) && (
              <div className="mb-2 flex items-start justify-between gap-2">
                {primary && <p className="text-sm font-medium">{primary.cell(r)}</p>}
                {action && <div className="flex-shrink-0">{action(r)}</div>}
              </div>
            )}
            <dl className="grid grid-cols-2 gap-x-4 gap-y-1.5 text-xs">
              {rest.map((c, ci) => (
                <div key={ci} className="flex items-center justify-between gap-2">
                  <dt className="text-muted-foreground">{c.header}</dt>
                  <dd className="tabular-nums font-medium">{c.cell(r)}</dd>
                </div>
              ))}
            </dl>
          </div>
        ))}
      </div>
    </>
  )
}

// Brand-consistent chart colors.
export const CHART = {
  navy: 'hsl(231, 44%, 30%)',
  gold: 'hsl(38, 56%, 55%)',
  grid: 'hsl(0, 0%, 0%, 0.08)',
  axis: 'hsl(0, 0%, 45%)',
}
