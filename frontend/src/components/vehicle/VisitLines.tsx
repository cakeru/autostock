import type { ComponentType, ReactNode } from 'react'
import { Disc, Droplet, Wrench, Crosshair, StickyNote, QrCode, X } from 'lucide-react'
import { AlignmentBlueprint, hasAlignmentData } from './AlignmentBlueprint'
import type { Visit, CornerData } from '@/types/vehicleProfile'

const POS_LABEL: Record<string, string> = {
  FL: 'Front-left', FR: 'Front-right', RL: 'Rear-left', RR: 'Rear-right', SPARE: 'Spare',
}

// The service facts of one visit — tires, alignment, oil, parts, notes. Pure and
// presentational so the internal timeline and the customer report render an
// identical body; photos, prices and controls live in the respective wrappers.
// `showNotes` is off for the customer report (notes can be internal shorthand).
// `onRemoveEvent`, when passed (internal only), adds a remove control to the
// auto-logged oil/tire lines so a mis-logged event can be corrected here.
export function VisitLines({ visit: v, bodyType, showNotes = true, onRemoveEvent }: {
  visit: Visit
  bodyType?: string
  showNotes?: boolean
  onRemoveEvent?: (eventId: number, kind: 'oil' | 'tire' | 'service') => void
}) {
  const svc = v.wheel_service
  const cornerMap: Record<string, CornerData> = {}
  if (svc) for (const c of svc.corners) cornerMap[c.position] = c
  const replaced = svc ? svc.corners.filter((c) => c.tread_before_mm != null) : []
  const showAlignment = !!svc && hasAlignmentData(cornerMap)

  return (
    <div className="space-y-2.5">
      {replaced.length > 0 && (
        <Row icon={Disc} title="Tires replaced">
          <div className="mt-1 grid grid-cols-1 gap-x-8 gap-y-1 sm:grid-cols-2">
            {replaced.map((c) => (
              <div key={c.position} className="flex items-center justify-between text-xs">
                <span className="text-muted-foreground">{POS_LABEL[c.position] || c.position}</span>
                <span className="tabular-nums">
                  <span className="text-amber-600 dark:text-amber-400">{c.tread_before_mm} mm</span>
                  <span className="mx-1.5 text-muted-foreground">→</span>
                  <span className="font-semibold text-emerald-600 dark:text-emerald-400">
                    {c.tread_mm != null ? `${c.tread_mm} mm` : 'new'}
                  </span>
                </span>
              </div>
            ))}
          </div>
        </Row>
      )}

      {/* A tire install logged at sale with no per-corner detail — don't repeat
          when the wheel snapshot already itemised the replaced tires. */}
      {v.tire_change && replaced.length === 0 && (
        <Row icon={Disc} title="Tires" detail={v.tire_note || undefined}
          action={onRemoveEvent && v.tire_event_id ? <RemoveButton onClick={() => onRemoveEvent(v.tire_event_id!, 'tire')} /> : undefined} />
      )}

      {v.oil_change && (
        <Row icon={Droplet} title="Oil change" detail={v.oil_note || undefined}
          action={onRemoveEvent && v.oil_event_id ? <RemoveButton onClick={() => onRemoveEvent(v.oil_event_id!, 'oil')} /> : undefined} />
      )}

      {(v.services || []).map((s) => (
        <Row key={s.id} icon={Wrench} title={s.name}
          action={onRemoveEvent ? <RemoveButton onClick={() => onRemoveEvent(s.id, 'service')} /> : undefined} />
      ))}

      {(v.parts || []).map((p) => (
        <Row key={p.id} icon={Wrench} title={`${p.part_name} replaced`} detail={p.position || undefined} />
      ))}

      {showAlignment && (
        <Row icon={Crosshair} title="Wheel alignment">
          <div className="mt-2"><AlignmentBlueprint bodyType={bodyType} corners={cornerMap} /></div>
        </Row>
      )}

      {/* Scanned batch traceability — internal only (never present on the public report). */}
      {(v.installs || []).map((ins, i) => (
        <Row
          key={`in-${i}`}
          icon={QrCode}
          title={`Fitted: ${ins.product_name}`}
          detail={[
            ins.position,
            `batch ${ins.batch_no}`,
            ins.dot_code ? `DOT ${ins.dot_code}` : '',
            ins.mechanic_name,
          ].filter(Boolean).join(' · ')}
        />
      ))}

      {showNotes && (v.notes || []).map((n, i) => (
        <Row key={i} icon={StickyNote} title={n} muted />
      ))}
    </div>
  )
}

export function Row({ icon: Icon, title, detail, muted, action, children }: {
  icon: ComponentType<{ className?: string }>
  title: string
  detail?: string
  muted?: boolean
  action?: ReactNode
  children?: React.ReactNode
}) {
  return (
    <div className="group/row">
      <div className="flex items-center gap-2">
        <Icon className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
        <span className={`text-sm ${muted ? 'text-muted-foreground' : 'font-medium'}`}>{title}</span>
        {detail && <span className="text-xs text-muted-foreground">· {detail}</span>}
        {action && <span className="ml-auto">{action}</span>}
      </div>
      {children}
    </div>
  )
}

function RemoveButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      aria-label="Remove this record"
      title="Remove this record (recomputes the reminder)"
      className="text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover/row:opacity-100"
    >
      <X className="h-3.5 w-3.5" />
    </button>
  )
}
