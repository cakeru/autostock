import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Droplet, Disc, Wrench, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { DueForServiceItem } from '@/types/vehicleProfile'

const DAY = 86_400_000
const iso = (d: Date) => d.toISOString().slice(0, 10)
const startOfMonth = (d: Date) => new Date(d.getFullYear(), d.getMonth(), 1)
const addMonths = (d: Date, n: number) => new Date(d.getFullYear(), d.getMonth() + n, 1)
const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December']
const WEEKDAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

function TypeIcon({ t, className }: { t: string; className?: string }) {
  const C = t === 'oil' ? Droplet : t === 'tire' ? Disc : Wrench
  return <C className={className} />
}

// A month calendar of upcoming service dues, placed on their due date. The
// certain/estimated distinction is honoured visually: date-basis dues (a fixed
// day interval — a real calendar date) get a solid marker; mileage-basis dues
// (a drifting km projection) get a soft dashed one with a "~". Overdue work,
// which lives in the past, is surfaced in a strip above the grid instead of
// being buried in last month.
export function DueCalendar({ items }: { items: DueForServiceItem[] }) {
  const navigate = useNavigate()
  const [month, setMonth] = useState(() => startOfMonth(new Date()))

  const overdue = useMemo(() => items.filter((i) => i.status === 'overdue'), [items])

  // Group the non-overdue, dated items by their due date (YYYY-MM-DD).
  const byDate = useMemo(() => {
    const m = new Map<string, DueForServiceItem[]>()
    for (const i of items) {
      if (i.status === 'overdue' || !i.due_date) continue
      const key = i.due_date.slice(0, 10)
      const arr = m.get(key)
      if (arr) arr.push(i)
      else m.set(key, [i])
    }
    return m
  }, [items])

  // Build the visible 6-week grid (Mon-start) covering this month.
  const cells = useMemo(() => {
    const first = startOfMonth(month)
    const offset = (first.getDay() + 6) % 7 // Mon=0
    const gridStart = new Date(first.getTime() - offset * DAY)
    return Array.from({ length: 42 }, (_, i) => new Date(gridStart.getTime() + i * DAY))
  }, [month])

  const todayIso = iso(new Date())

  return (
    <div className="space-y-4">
      {overdue.length > 0 && (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-3">
          <p className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-destructive">
            <AlertTriangle className="h-3.5 w-3.5" /> {overdue.length} overdue — call these first
          </p>
          <div className="flex flex-wrap gap-1.5">
            {overdue.map((i) => (
              <button
                key={`${i.vehicle_id}-${i.key}`}
                onClick={() => navigate(`/vehicles/${i.vehicle_id}`)}
                className="inline-flex items-center gap-1.5 rounded-md border border-destructive/30 bg-card px-2 py-1 text-xs hover:bg-destructive/10"
                title={`${i.customer_name}${i.customer_phone ? ' · ' + i.customer_phone : ''}`}
              >
                <TypeIcon t={i.event_type} className="h-3 w-3 text-destructive" />
                <span className="font-medium">{i.plate_number}</span>
                <span className="text-muted-foreground">{i.label}</span>
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="rounded-lg border bg-card">
        {/* Month header */}
        <div className="flex items-center justify-between border-b px-3 py-2">
          <p className="text-sm font-semibold">{MONTHS[month.getMonth()]} {month.getFullYear()}</p>
          <div className="flex items-center gap-1">
            <Button variant="outline" size="sm" onClick={() => setMonth(startOfMonth(new Date()))}>Today</Button>
            <Button variant="ghost" size="icon" onClick={() => setMonth(addMonths(month, -1))} aria-label="Previous month"><ChevronLeft className="h-4 w-4" /></Button>
            <Button variant="ghost" size="icon" onClick={() => setMonth(addMonths(month, 1))} aria-label="Next month"><ChevronRight className="h-4 w-4" /></Button>
          </div>
        </div>

        {/* Weekday row */}
        <div className="grid grid-cols-7 border-b text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
          {WEEKDAYS.map((d) => <div key={d} className="px-2 py-1.5 text-center">{d}</div>)}
        </div>

        {/* Day grid */}
        <div className="grid grid-cols-7">
          {cells.map((d, idx) => {
            const inMonth = d.getMonth() === month.getMonth()
            const key = iso(d)
            const dues = byDate.get(key) || []
            const isToday = key === todayIso
            return (
              <div key={idx} className={`min-h-[76px] border-b border-r p-1 ${idx % 7 === 6 ? 'border-r-0' : ''} ${inMonth ? '' : 'bg-muted/30'}`}>
                <div className={`mb-1 text-right text-[11px] tabular-nums ${isToday ? 'font-bold text-primary' : inMonth ? 'text-muted-foreground' : 'text-muted-foreground/50'}`}>
                  {isToday ? <span className="inline-grid h-5 w-5 place-items-center rounded-full bg-primary text-primary-foreground">{d.getDate()}</span> : d.getDate()}
                </div>
                <div className="space-y-0.5">
                  {dues.slice(0, 3).map((i) => <DueChip key={`${i.vehicle_id}-${i.key}`} item={i} onClick={() => navigate(`/vehicles/${i.vehicle_id}`)} />)}
                  {dues.length > 3 && <p className="px-1 text-[10px] text-muted-foreground">+{dues.length - 3} more</p>}
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-muted-foreground">
        <span className="flex items-center gap-1.5"><span className="h-2.5 w-2.5 rounded-sm bg-amber-500" /> certain date (day interval)</span>
        <span className="flex items-center gap-1.5"><span className="h-2.5 w-2.5 rounded-sm border border-dashed border-muted-foreground" /> ~ estimated (from mileage)</span>
      </div>
    </div>
  )
}

function DueChip({ item, onClick }: { item: DueForServiceItem; onClick: () => void }) {
  const certain = item.due_basis === 'date'
  const soon = item.status === 'due_soon'
  return (
    <button
      onClick={onClick}
      title={`${item.plate_number} · ${item.label} · ${certain ? 'firm date' : 'estimated from mileage'}${item.customer_name ? ' · ' + item.customer_name : ''}`}
      className={`flex w-full items-center gap-1 truncate rounded px-1 py-0.5 text-left text-[10px] leading-tight ${
        certain
          ? (soon ? 'bg-amber-500/20 text-amber-700 dark:text-amber-300' : 'bg-amber-500/10 text-amber-700 dark:text-amber-300')
          : 'border border-dashed border-muted-foreground/40 text-muted-foreground'
      }`}
    >
      <TypeIcon t={item.event_type} className="h-2.5 w-2.5 flex-shrink-0" />
      <span className="truncate font-medium">{item.plate_number}</span>
      {!certain && <span className="flex-shrink-0">~</span>}
    </button>
  )
}
