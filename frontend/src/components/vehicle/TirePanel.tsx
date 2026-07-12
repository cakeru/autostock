import { formatDate } from '@/utils/date'
import type { CornerData } from '@/types/vehicleProfile'

const ORDER = ['FL', 'FR', 'RL', 'RR', 'SPARE']
const POS_LABEL: Record<string, string> = {
  FL: 'Front left', FR: 'Front right', RL: 'Rear left', RR: 'Rear right', SPARE: 'Spare',
}

// DOT = WWYY (week + year). Returns a compact age like "8 mo" or "3.2 yr", or
// null when the code isn't a plausible date.
function tireAge(dot?: string): string | null {
  if (!dot) return null
  const m = dot.match(/(\d{2})(\d{2})$/)
  if (!m) return null
  const week = parseInt(m[1], 10)
  const year = 2000 + parseInt(m[2], 10)
  if (week < 1 || week > 53) return null
  const made = new Date(year, 0, 1 + (week - 1) * 7)
  const years = (Date.now() - made.getTime()) / (365.25 * 24 * 3600 * 1000)
  if (years < 0 || years > 40) return null
  return years < 1 ? `${Math.round(years * 12)} mo` : `${years.toFixed(1)} yr`
}

function treadClass(mm?: number): string {
  if (mm == null) return 'text-muted-foreground'
  if (mm < 3) return 'text-destructive font-semibold'
  if (mm < 5) return 'text-amber-600 dark:text-amber-400 font-semibold'
  return ''
}

// The spec of what's actually mounted on each corner right now — brand, size,
// DOT/age, tread, pressure, wear — as a dense inspection grid beneath the
// condition diagram. Complements the diagram's at-a-glance colour with the
// "give me the numbers" detail. Alignment lives on its own blueprint, not here.
export function TirePanel({ entries }: { entries: Record<string, { corner: CornerData; date: string }> }) {
  const rows = ORDER.filter((p) => entries[p]).map((p) => ({ pos: p, ...entries[p] }))
  if (rows.length === 0) return null

  // Show a per-corner "as of" date only if the corners weren't all logged together.
  const dates = new Set(rows.map((r) => r.date.slice(0, 10)))
  const showDates = dates.size > 1

  return (
    <div className="mt-4 border-t pt-3">
      <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tires on the car</p>
      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-left text-[10px] uppercase tracking-wide text-muted-foreground">
              <th className="pb-1.5 pr-3 font-medium">Corner</th>
              <th className="pb-1.5 pr-3 font-medium">Tire</th>
              <th className="pb-1.5 pr-3 font-medium">DOT</th>
              <th className="pb-1.5 pr-3 text-right font-medium">Tread</th>
              <th className="pb-1.5 pr-3 text-right font-medium">Pressure</th>
              {showDates && <th className="pb-1.5 text-right font-medium">Logged</th>}
            </tr>
          </thead>
          <tbody className="tabular-nums">
            {rows.map(({ pos, corner: c, date }) => {
              const tire = [c.tire_brand, c.tire_size].filter(Boolean).join(' ')
              const age = tireAge(c.tire_dot)
              return (
                <tr key={pos} className="border-t border-border/60 align-top">
                  <td className="py-1.5 pr-3 font-medium">{POS_LABEL[pos] || pos}</td>
                  <td className="py-1.5 pr-3">
                    {tire || <span className="text-muted-foreground">—</span>}
                    {c.wear_note && <span className="block text-[11px] italic text-muted-foreground">{c.wear_note}</span>}
                  </td>
                  <td className="py-1.5 pr-3 text-muted-foreground">
                    {c.tire_dot ? <>{c.tire_dot}{age && <span className="text-muted-foreground/70"> · {age}</span>}</> : '—'}
                  </td>
                  <td className={`py-1.5 pr-3 text-right ${treadClass(c.tread_mm)}`}>
                    {c.tread_mm != null ? `${c.tread_mm} mm` : '—'}
                  </td>
                  <td className="py-1.5 pr-3 text-right text-muted-foreground">
                    {c.pressure != null ? `${c.pressure} psi` : '—'}
                  </td>
                  {showDates && <td className="py-1.5 text-right text-muted-foreground">{formatDate(date)}</td>}
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
