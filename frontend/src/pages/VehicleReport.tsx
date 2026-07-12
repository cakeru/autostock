import { useMemo, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { Printer, Droplet, Disc, Wrench, Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { CarDiagram } from '@/components/vehicle/CarDiagram'
import { VisitLines } from '@/components/vehicle/VisitLines'
import { latestByPosition } from '@/components/vehicle/CurrentCondition'
import { usePublicReport } from '@/hooks/useVehicleProfile'
import { useSendToTelegram } from '@/hooks/useSendToTelegram'
import { useAuth } from '@/contexts/AuthContext'
import { imageSrc } from '@/utils/imageUrl'
import { formatDate } from '@/utils/date'
import type { CornerData, DueStatus, PartColor, PartStatus, Visit } from '@/types/vehicleProfile'

const OIL_STATUS: Record<string, PartColor> = {
  overdue: 'red', due_soon: 'yellow', ok: 'green', unknown: 'grey',
}
const DUE_COLOR: Record<string, string> = {
  red: '#dc2626', yellow: '#d97706', green: '#059669',
}

function dueSentence(d: DueStatus): { text: string; tone: 'red' | 'yellow' | 'green' | 'grey' } {
  const what = d.label || (d.event_type === 'oil' ? 'Oil change' : 'Tires')
  if (d.status === 'unknown') return { text: `${what}: no history on file yet.`, tone: 'grey' }
  const when = d.due_date ? formatDate(d.due_date) : ''
  if (d.status === 'overdue') return { text: `${what}: overdue since ${when} — please book a visit.`, tone: 'red' }
  if (d.status === 'due_soon') return { text: `${what}: due soon (${when}).`, tone: 'yellow' }
  return { text: `${what}: on track — next due around ${when}.`, tone: 'green' }
}

// Public, token-authorized condition report — what the customer keeps. Also
// the shop's print-friendly inspection sheet (the Print button hides itself).
export function VehicleReport() {
  const { token } = useParams<{ token: string }>()
  const { data: report, isLoading, isError } = usePublicReport(token || '')
  const { user } = useAuth()
  const reportRef = useRef<HTMLDivElement>(null)
  const { send: sendTelegram, sending } = useSendToTelegram()

  const handleSendTelegram = () => {
    if (!report || !reportRef.current) return
    const cap = `🚗 Vehicle report — ${report.plate_number}${report.customer_name ? ` · ${report.customer_name}` : ''}`
    sendTelegram(reportRef.current, `report-${report.plate_number}.pdf`, cap)
  }

  const corners = useMemo(() => {
    const out: Record<string, CornerData> = {}
    if (report) {
      for (const [pos, entry] of Object.entries(latestByPosition(report.wheel_services))) out[pos] = entry.corner
    }
    return out
  }, [report])

  const statusMap = useMemo(() => {
    const out: Record<string, PartStatus> = {}
    for (const s of report?.part_statuses || []) out[s.part_key] = s
    return out
  }, [report])

  const partDue = useMemo(() => {
    const out: Record<string, PartColor> = {}
    for (const d of report?.due || []) {
      if (d.event_type === 'part' && (d.status === 'overdue' || d.status === 'due_soon')) {
        out[d.key] = OIL_STATUS[d.status]
      }
    }
    return out
  }, [report])

  if (isLoading) return <p className="p-10 text-center text-sm text-muted-foreground">Loading report…</p>
  if (isError || !report) {
    return (
      <div className="mx-auto max-w-md p-10 text-center">
        <p className="text-sm font-semibold">This report link is no longer active.</p>
        <p className="mt-1 text-sm text-muted-foreground">The shop may have revoked or replaced it — give them a call for a fresh link.</p>
      </div>
    )
  }

  const oilDue = report.due.find((d) => d.event_type === 'oil')
  const shopName = report.shop_name || 'K&S Wheel-Tyre'
  const spec = [report.make, report.model].filter(Boolean).join(' ') + (report.year ? ` (${report.year})` : '')
  const tireRows = (['FL', 'FR', 'RL', 'RR', 'SPARE'] as const).filter((p) => corners[p])

  return (
    <div ref={reportRef} className="mx-auto max-w-2xl px-5 py-8 print:max-w-full print:px-0 print:py-0">
      {/* Report header */}
      <div className="flex items-start justify-between gap-4 border-b-2 pb-4" style={{ borderColor: 'var(--color-primary)' }}>
        <div>
          <p className="text-[11px] font-semibold uppercase tracking-[0.2em]" style={{ color: 'var(--color-accent)' }}>{shopName}</p>
          <h1 className="mt-1 text-xl font-semibold tracking-tight">Vehicle Condition Report</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            <span className="font-medium text-foreground">{report.plate_number}</span>
            {spec.trim() ? ` · ${spec.trim()}` : ''}
            {report.customer_name ? ` · ${report.customer_name}` : ''}
          </p>
        </div>
        <div className="flex flex-col items-end gap-2">
          <p className="text-xs tabular-nums text-muted-foreground">{formatDate(report.generated_at)}</p>
          {/* Actions are excluded from the PDF capture and hidden when printing. */}
          <div className="flex gap-2 print:hidden" data-html2canvas-ignore>
            {user && (
              <Button variant="outline" size="sm" onClick={handleSendTelegram} disabled={sending}>
                <Send className="mr-1.5 h-3.5 w-3.5" /> {sending ? 'Sending…' : 'Send to Telegram'}
              </Button>
            )}
            <Button variant="outline" size="sm" onClick={() => window.print()}>
              <Printer className="mr-1.5 h-3.5 w-3.5" /> Print
            </Button>
          </div>
        </div>
      </div>

      {/* Service due summary — oil + tires always, plus any part that needs
          attention. On-track parts stay quiet to keep this actionable. */}
      <div className="mt-5 space-y-1.5">
        {report.due
          .filter((d) => d.event_type !== 'part' || d.status === 'overdue' || d.status === 'due_soon')
          .map((d) => {
            const s = dueSentence(d)
            const Icon = d.event_type === 'oil' ? Droplet : d.event_type === 'tire' ? Disc : Wrench
            return (
              <p key={d.key} className="flex items-center gap-2 text-sm">
                <Icon className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
                <span style={s.tone !== 'grey' ? { color: DUE_COLOR[s.tone] } : undefined}
                  className={s.tone === 'grey' ? 'text-muted-foreground' : 'font-medium'}>
                  {s.text}
                </span>
              </p>
            )
          })}
      </div>

      {/* Current condition drawing + inspection sheet */}
      <div className="mt-7 border-t pt-5">
        <p className="mb-3 text-[11px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">Current condition</p>
        <CarDiagram
          bodyType={report.body_type}
          corners={corners}
          oilStatus={OIL_STATUS[oilDue?.status ?? 'unknown'] ?? 'grey'}
          partStatuses={statusMap}
          partDue={partDue}
          readOnly
        />
      </div>

      {/* Tire details */}
      {tireRows.length > 0 && (
        <div className="mt-6">
          <p className="border-b pb-2 text-[11px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">Tires on the car</p>
          <table className="w-full text-xs">
            <thead>
              <tr className="text-left text-[10px] uppercase tracking-wide text-muted-foreground">
                <th className="py-1.5 font-medium">Position</th>
                <th className="py-1.5 font-medium">Tire</th>
                <th className="py-1.5 font-medium">DOT</th>
                <th className="py-1.5 text-right font-medium">Tread</th>
              </tr>
            </thead>
            <tbody className="tabular-nums">
              {tireRows.map((pos) => {
                const c = corners[pos]
                return (
                  <tr key={pos} className="border-t border-border/60">
                    <td className="py-1.5 font-medium">{pos === 'SPARE' ? 'Spare' : pos}</td>
                    <td className="py-1.5">{[c.tire_brand, c.tire_size].filter(Boolean).join(' ') || '—'}</td>
                    <td className="py-1.5">{c.tire_dot || '—'}</td>
                    <td className="py-1.5 text-right">{c.tread_mm != null ? `${c.tread_mm} mm` : '—'}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Service history — the track record of work done, newest first */}
      {report.visits.length > 0 && (
        <div className="mt-7 border-t pt-5">
          <p className="mb-4 text-[11px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">Service history</p>
          <ol className="relative space-y-6 border-l border-border pl-6">
            {report.visits.map((v, i) => <PublicVisit key={`${v.date}-${i}`} visit={v} bodyType={report.body_type} />)}
          </ol>
        </div>
      )}

      {/* Footer */}
      <div className="mt-8 border-t pt-4 text-xs text-muted-foreground">
        <p className="font-medium text-foreground">{shopName}</p>
        {report.shop_address && <p>{report.shop_address}</p>}
        {report.shop_phone && <p>Questions or booking: {report.shop_phone}</p>}
        <p className="mt-2 text-[11px]">
          Tire and oil indicators are estimated from your service history, tread measurements and tire age.
        </p>
      </div>
    </div>
  )
}

// One visit on the customer-facing history timeline: the work done that day,
// plus any photos the shop chose to share. No prices or internal notes — the
// backend already stripped those from report.visits.
function PublicVisit({ visit: v, bodyType }: { visit: Visit; bodyType?: string }) {
  return (
    <li className="relative break-inside-avoid">
      <span className="absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 bg-white" style={{ borderColor: 'var(--color-primary)' }} />
      <div className="flex items-baseline justify-between gap-2">
        <p className="text-sm font-semibold">{formatDate(v.date)}</p>
        {v.mileage != null && <p className="text-xs tabular-nums text-muted-foreground">{v.mileage.toLocaleString()} km</p>}
      </div>
      <div className="mt-2">
        <VisitLines visit={v} bodyType={bodyType} showNotes={false} />
        {v.photos && v.photos.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-2">
            {v.photos.map((ph, i) => (
              <a key={i} href={imageSrc(ph.url)} target="_blank" rel="noreferrer" className="block">
                <img src={imageSrc(ph.url)} alt={ph.caption || 'Service photo'} className="h-24 w-24 rounded border object-cover" />
                {ph.caption && <span className="mt-0.5 block max-w-24 truncate text-[10px] text-muted-foreground">{ph.caption}</span>}
              </a>
            ))}
          </div>
        )}
      </div>
    </li>
  )
}
