import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { CornerData, PartColor, PartStatus } from '@/types/vehicleProfile'

// Semantic condition colors — used sparingly, on small marks, so they read as
// instrument indicators rather than decoration.
const COLOR: Record<PartColor, string> = {
  green: '#059669', yellow: '#d97706', red: '#dc2626', grey: '#94a3b8',
}
const STATUS_WORD: Record<PartColor, string> = {
  green: 'Good', yellow: 'Watch', red: 'Attention', grey: 'Not checked',
}

// Manually-assessed inspection parts. Tires + oil are auto-derived. Backend
// accepts any part_key, so extending this list is all it takes to add a part.
type DviPart = { key: string; label: string; motorcycle?: boolean; motoOnly?: boolean }
const DVI_PARTS: DviPart[] = [
  { key: 'brakes_front', label: 'Brakes (front)', motorcycle: true },
  { key: 'brakes_rear', label: 'Brakes (rear)', motorcycle: true },
  { key: 'battery', label: 'Battery', motorcycle: true },
  { key: 'lights', label: 'Lights', motorcycle: true },
  { key: 'wipers', label: 'Wipers' },
  { key: 'suspension', label: 'Suspension' },
  { key: 'air_filter', label: 'Air filter' },
  { key: 'coolant', label: 'Coolant' },
  { key: 'belts', label: 'Belts & hoses' },
  { key: 'chain_sprocket', label: 'Chain & sprocket', motorcycle: true, motoOnly: true },
]

export function normalizeType(bodyType?: string): 'sedan' | 'suv' | 'pickup' | 'motorcycle' {
  switch (bodyType) {
    case 'suv': case 'van': return 'suv'
    case 'pickup': case 'truck': return 'pickup'
    case 'motorcycle': case 'moto': return 'motorcycle'
    default: return 'sedan'
  }
}

function dotAgeYears(dot?: string): number | null {
  if (!dot) return null
  const m = dot.match(/(\d{2})(\d{2})$/)
  if (!m) return null
  const week = parseInt(m[1], 10)
  const year = 2000 + parseInt(m[2], 10)
  if (week < 1 || week > 53) return null
  const made = new Date(year, 0, 1 + (week - 1) * 7)
  const years = (Date.now() - made.getTime()) / (365.25 * 24 * 3600 * 1000)
  return years >= 0 && years < 40 ? years : null
}

const RANK: Record<PartColor, number> = { grey: 0, green: 1, yellow: 2, red: 3 }
const worst = (a: PartColor, b: PartColor): PartColor => (RANK[a] >= RANK[b] ? a : b)

export function tireColor(c?: CornerData): PartColor {
  if (!c) return 'grey'
  let color: PartColor = 'grey'
  if (c.tread_mm != null) color = worst(color, c.tread_mm < 3 ? 'red' : c.tread_mm < 5 ? 'yellow' : 'green')
  const age = dotAgeYears(c.tire_dot)
  if (age != null) color = worst(color, age > 6 ? 'red' : age >= 4 ? 'yellow' : 'green')
  return color
}

// ---------------------------------------------------------------------------
// Geometry — wheel anchors per body type in the 340 × 300 viewBox (front = top,
// car body centred on x=170, dimension callouts in the side margins).
// ---------------------------------------------------------------------------

export type WheelGeom = { cx: number; cy: number; w: number; h: number; side: 'left' | 'right'; label: string }

export function wheelGeometry(type: string): Record<string, WheelGeom> {
  if (type === 'motorcycle') {
    return {
      FL: { cx: 170, cy: 66, w: 13, h: 42, side: 'left', label: 'FRONT' },
      RL: { cx: 170, cy: 240, w: 15, h: 46, side: 'right', label: 'REAR' },
    }
  }
  const frontY = type === 'sedan' ? 102 : 104
  const rearY = type === 'suv' ? 220 : type === 'pickup' ? 224 : 216
  return {
    FL: { cx: 118, cy: frontY, w: 13, h: 40, side: 'left', label: 'FL' },
    FR: { cx: 222, cy: frontY, w: 13, h: 40, side: 'right', label: 'FR' },
    RL: { cx: 118, cy: rearY, w: 13, h: 40, side: 'left', label: 'RL' },
    RR: { cx: 222, cy: rearY, w: 13, h: 40, side: 'right', label: 'RR' },
  }
}

export function CarDiagram({ bodyType, corners, oilStatus, partStatuses, partDue, onSetPart, saving, readOnly }: {
  bodyType?: string
  corners: Record<string, CornerData>
  oilStatus: PartColor
  partStatuses: Record<string, PartStatus>
  partDue?: Record<string, PartColor>
  onSetPart?: (key: string, status: PartColor, note: string) => void
  saving?: boolean
  readOnly?: boolean
}) {
  const type = normalizeType(bodyType)
  const wheels = wheelGeometry(type)
  const parts = DVI_PARTS.filter((p) => (type === 'motorcycle' ? p.motorcycle : !p.motoOnly))
  const [editing, setEditing] = useState<{ key: string; label: string } | null>(null)

  // A tech's manual assessment wins; otherwise the part's reminder-due colour
  // lights it up (grey when neither is set).
  const statusOf = (key: string): PartColor => partStatuses[key]?.status ?? partDue?.[key] ?? 'grey'
  const attention = parts.filter((p) => statusOf(p.key) === 'red').length
  const watch = parts.filter((p) => statusOf(p.key) === 'yellow').length

  return (
    <div className="flex flex-col gap-6 lg:flex-row lg:items-start">
      {/* Technical drawing: the car with dimension-style tire callouts.
          Linework in the K&S brand navy; ticks + FRONT label in brand gold. */}
      <div className="mx-auto w-full max-w-[360px] flex-shrink-0" style={{ color: 'var(--color-primary)' }}>
        <svg viewBox="0 0 340 300" role="img" aria-label="Top-down vehicle condition drawing">
          <defs>
            <pattern id="dvi-grid" width="20" height="20" patternUnits="userSpaceOnUse">
              <path d="M20 0H0V20" fill="none" stroke="currentColor" strokeOpacity="0.07" strokeWidth="0.5" />
            </pattern>
          </defs>
          <rect width="340" height="300" fill="url(#dvi-grid)" />
          {/* registration ticks */}
          <g strokeWidth="1" style={{ stroke: 'var(--color-brand-gold)' }}>
            <path d="M6 14V6h8M326 6h8v8M6 286v8h8M334 286v8h-8" fill="none" />
          </g>
          <text x="170" y="13" textAnchor="middle" fontSize="7" letterSpacing="3"
            style={{ fill: 'var(--color-accent)' }}>FRONT</text>

          <CarBody type={type} />

          {Object.entries(wheels).map(([pos, g]) => (
            <Wheel key={pos} geom={g} corner={corners[pos]} />
          ))}

          <OilMarker type={type} status={oilStatus} />
        </svg>
      </div>

      {/* Inspection sheet */}
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline justify-between border-b pb-2">
          <p className="text-[11px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">Inspection</p>
          <p className="text-[11px] tabular-nums text-muted-foreground">
            {attention > 0 && <span className="font-semibold" style={{ color: COLOR.red }}>{attention} attention</span>}
            {attention > 0 && watch > 0 && ' · '}
            {watch > 0 && <span className="font-semibold" style={{ color: COLOR.yellow }}>{watch} watch</span>}
          </p>
        </div>

        {/* Oil — auto-derived, first row of the sheet */}
        <div className="flex items-center gap-3 border-b border-dashed px-1 py-2">
          <StatusKey status={oilStatus} />
          <span className="w-28 flex-shrink-0 text-xs font-medium sm:w-32">Oil / engine</span>
          <span className="min-w-0 flex-1 truncate text-xs text-muted-foreground">
            {oilStatus === 'red' ? 'Service overdue' : oilStatus === 'yellow' ? 'Due soon' : oilStatus === 'green' ? 'On track' : 'No service history'}
          </span>
          <span className="text-[10px] uppercase tracking-wide text-muted-foreground/60">auto</span>
        </div>

        <div className="divide-y divide-border/60">
          {parts.map((part) => {
            const status = statusOf(part.key)
            const note = partStatuses[part.key]?.note
            const inner = (
              <>
                <StatusKey status={status} />
                <span className="w-28 flex-shrink-0 text-xs font-medium sm:w-32">{part.label}</span>
                <span className="min-w-0 flex-1 truncate text-xs text-muted-foreground">{note || (status === 'grey' ? '—' : '')}</span>
                <span className="flex-shrink-0 text-[10px] font-medium uppercase tracking-wide"
                  style={{ color: status === 'grey' ? undefined : COLOR[status] }}>
                  {status === 'grey' ? <span className="text-muted-foreground/50">not checked</span> : STATUS_WORD[status]}
                </span>
              </>
            )
            return readOnly ? (
              <div key={part.key} className="flex w-full items-center gap-3 px-1 py-2">{inner}</div>
            ) : (
              <button
                key={part.key}
                onClick={() => setEditing({ key: part.key, label: part.label })}
                className="flex w-full items-center gap-3 px-1 py-2 text-left transition-colors hover:bg-muted/40"
              >
                {inner}
              </button>
            )
          })}
        </div>
        {!readOnly && (
          <p className="mt-2 text-[11px] text-muted-foreground">
            Tap a part to set its condition. An empty box means it hasn't been checked.
          </p>
        )}
      </div>

      {editing && !readOnly && (
        <SetStatusDialog
          label={editing.label}
          current={partStatuses[editing.key]}
          onClose={() => setEditing(null)}
          onSave={(status, note) => { onSetPart?.(editing.key, status, note); setEditing(null) }}
          saving={!!saving}
        />
      )}
    </div>
  )
}

// Square condition key — filled when assessed, outlined when not checked.
function StatusKey({ status }: { status: PartColor }) {
  return status === 'grey'
    ? <span className="h-2.5 w-2.5 flex-shrink-0 rounded-[1px] border border-current opacity-30" />
    : <span className="h-2.5 w-2.5 flex-shrink-0 rounded-[1px]" style={{ background: COLOR[status] }} />
}

// ---------------------------------------------------------------------------
// SVG pieces
// ---------------------------------------------------------------------------

function Wheel({ geom, corner }: { geom: WheelGeom; corner?: CornerData }) {
  const color = tireColor(corner)
  const c = COLOR[color]
  const { cx, cy, w, h, side, label } = geom
  const isLeft = side === 'left'
  const edgeX = isLeft ? cx - w / 2 : cx + w / 2
  const endX = isLeft ? 68 : 272
  const textX = isLeft ? 64 : 276
  const anchor = isLeft ? 'end' : 'start'
  const age = dotAgeYears(corner?.tire_dot)
  const warn = color === 'red' || color === 'yellow'

  return (
    <g>
      <title>{label}: {corner?.tread_mm != null ? `${corner.tread_mm} mm tread` : 'no tread recorded'}</title>
      <rect x={cx - w / 2} y={cy - h / 2} width={w} height={h} rx={3}
        fill={c} fillOpacity={color === 'grey' ? 0.06 : 0.12}
        stroke={color === 'grey' ? 'currentColor' : c} strokeOpacity={color === 'grey' ? 0.45 : 1}
        strokeWidth={color === 'grey' ? 1 : 1.8} />
      {/* dimension leader */}
      <line x1={edgeX} y1={cy} x2={endX} y2={cy} stroke="currentColor" strokeOpacity="0.35" strokeWidth="0.75" />
      <circle cx={edgeX + (isLeft ? -1 : 1)} cy={cy} r="1.4" fill="currentColor" fillOpacity="0.5" />

      <text x={textX} y={cy - 7} textAnchor={anchor} fontSize="7.5" letterSpacing="1"
        fill="currentColor" fillOpacity="0.55">{label}</text>
      {corner?.tread_mm != null ? (
        <>
          <text x={textX} y={cy + 4} textAnchor={anchor} fontSize="11" fontWeight="600"
            style={{ fontVariantNumeric: 'tabular-nums' }}
            fill={warn ? c : 'currentColor'}>{corner.tread_mm} mm</text>
          {age != null && (
            <text x={textX} y={cy + 14} textAnchor={anchor} fontSize="7.5"
              style={{ fontVariantNumeric: 'tabular-nums' }}
              fill="currentColor" fillOpacity="0.55">{age.toFixed(1)} yr old</text>
          )}
        </>
      ) : (
        <text x={textX} y={cy + 4} textAnchor={anchor} fontSize="8" fill="currentColor" fillOpacity="0.4">not logged</text>
      )}
    </g>
  )
}

function OilMarker({ type, status }: { type: string; status: PartColor }) {
  const pos = type === 'motorcycle' ? { x: 202, y: 128 } : { x: 170, y: 58 }
  const c = status === 'grey' ? 'currentColor' : COLOR[status]
  return (
    <g>
      <title>Oil / engine</title>
      {type === 'motorcycle' && (
        <line x1="184" y1="128" x2="195" y2="128" stroke="currentColor" strokeOpacity="0.35" strokeWidth="0.75" />
      )}
      <circle cx={pos.x} cy={pos.y} r="5.5" fill="none" stroke={c}
        strokeOpacity={status === 'grey' ? 0.35 : 1} strokeWidth="1.8" />
      <circle cx={pos.x} cy={pos.y} r="1.6" fill={c} fillOpacity={status === 'grey' ? 0.35 : 1} />
      <text x={pos.x} y={pos.y + 15} textAnchor="middle" fontSize="6.5" letterSpacing="1.5"
        fill="currentColor" fillOpacity="0.5">OIL</text>
    </g>
  )
}

// ---------------------------------------------------------------------------
// Body linework per type — drawn top-down silhouettes, front = top, centred on
// x=170. Thin strokes, glass panels, mirrors; blueprint weight hierarchy.
// ---------------------------------------------------------------------------

export function CarBody({ type }: { type: string }) {
  const body = { fill: 'currentColor', fillOpacity: 0.03, stroke: 'currentColor', strokeWidth: 1.4 }
  const glass = { fill: 'currentColor', fillOpacity: 0.08, stroke: 'currentColor', strokeOpacity: 0.55, strokeWidth: 0.75 }
  const detail = { fill: 'none', stroke: 'currentColor', strokeOpacity: 0.3, strokeWidth: 0.75 }

  if (type === 'motorcycle') {
    return (
      <g>
        {/* forks, handlebar */}
        <path d="M164 87 L166 106 M176 87 L174 106" stroke="currentColor" strokeWidth="1.4" />
        <path d="M138 97 C152 91 188 91 202 97" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
        {/* tank */}
        <path d="M161 110 C157 121 157 135 162 147 C167 150 173 150 178 147 C183 135 183 121 179 110 C173 106 167 106 161 110 Z" {...body} />
        {/* seat */}
        <path d="M162 152 C158 166 158 186 164 202 L176 202 C182 186 182 166 178 152 Z" {...body} />
        {/* swingarm */}
        <path d="M165 206 L167 217 M175 206 L173 217" stroke="currentColor" strokeWidth="1.4" />
      </g>
    )
  }

  if (type === 'pickup') {
    return (
      <g>
        <path d="M170 30 C194 30 206 35 210 47 C213 58 215 77 215 102 L215 256 C215 265 212 269 204 270 L136 270 C128 269 125 265 125 256 L125 102 C125 77 127 58 130 47 C134 35 146 30 170 30 Z" {...body} />
        <path d="M138 80 C159 76 181 76 202 80 L195 102 C178 99 162 99 145 102 Z" {...glass} />
        <path d="M147 140 C163 143 177 143 193 140 L197 152 C179 155 161 155 143 152 Z" {...glass} />
        {/* bed */}
        <path d="M137 164 H203 V256 H137 Z" {...detail} strokeOpacity={0.5} />
        <path d="M137 195 H203 M137 225 H203" {...detail} strokeOpacity={0.15} />
        {/* hood creases */}
        <path d="M152 40 C148 56 146 68 145 78 M188 40 C192 56 194 68 195 78" {...detail} />
        <Mirrors y={100} />
      </g>
    )
  }

  if (type === 'suv') {
    return (
      <g>
        <path d="M170 26 C194 26 207 31 211 44 C214 56 215 76 215 102 L215 208 C215 234 214 252 211 261 C207 272 194 276 170 276 C146 276 133 272 129 261 C126 252 125 234 125 208 L125 102 C125 76 126 56 129 44 C133 31 146 26 170 26 Z" {...body} />
        <path d="M137 78 C158 74 182 74 203 78 L196 101 C179 98 161 98 144 101 Z" {...glass} />
        <path d="M144 234 C161 237 179 237 196 234 L203 254 C182 258 158 258 137 254 Z" {...glass} />
        {/* roof rails */}
        <path d="M141 108 L141 228 M199 108 L199 228" {...detail} />
        <path d="M151 36 C147 52 145 64 144 76 M189 36 C193 52 195 64 196 76" {...detail} />
        <Mirrors y={103} />
      </g>
    )
  }

  // sedan / hatchback
  return (
    <g>
      <path d="M170 28 C192 28 204 33 208 45 C211 55 213 73 214 95 C215 118 215 150 215 165 C215 198 214 224 210 243 C207 261 194 270 170 271 C146 270 133 261 130 243 C126 224 125 198 125 165 C125 150 125 118 126 95 C127 73 129 55 132 45 C136 33 148 28 170 28 Z" {...body} />
      <path d="M139 90 C159 86 181 86 201 90 L194 113 C178 110 162 110 146 113 Z" {...glass} />
      <path d="M146 186 C162 189 178 189 194 186 L201 210 C181 214 159 214 139 210 Z" {...glass} />
      {/* roof edges */}
      <path d="M146 118 L145 182 M194 118 L195 182" {...detail} />
      {/* hood creases + trunk line */}
      <path d="M152 40 C148 57 146 73 145 87 M188 40 C192 57 194 73 195 87" {...detail} />
      <path d="M141 248 C160 252 180 252 199 248" {...detail} />
      <Mirrors y={97} />
    </g>
  )
}

function Mirrors({ y }: { y: number }) {
  const mirror = { fill: 'currentColor', fillOpacity: 0.12, stroke: 'currentColor', strokeOpacity: 0.7, strokeWidth: 0.9 }
  return (
    <g>
      <path d={`M126 ${y} C117 ${y - 4} 112 ${y - 4} 111 ${y} C111 ${y + 4} 118 ${y + 5} 127 ${y + 6} Z`} {...mirror} />
      <path d={`M214 ${y} C223 ${y - 4} 228 ${y - 4} 229 ${y} C229 ${y + 4} 222 ${y + 5} 213 ${y + 6} Z`} {...mirror} />
    </g>
  )
}

// ---------------------------------------------------------------------------
// Condition dialog
// ---------------------------------------------------------------------------

function SetStatusDialog({ label, current, onClose, onSave, saving }: {
  label: string
  current?: PartStatus
  onClose: () => void
  onSave: (status: PartColor, note: string) => void
  saving: boolean
}) {
  const [note, setNote] = useState(current?.note || '')
  const options: PartColor[] = ['green', 'yellow', 'red', 'grey']
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-xs rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">{label}</p>
        <p className="mb-3 text-xs text-muted-foreground">Set this part's condition.</p>
        <div className="mb-3">
          <Input value={note} onChange={(e) => setNote(e.target.value)} placeholder="Note (optional) — e.g. pads at 3mm" />
        </div>
        <div className="grid grid-cols-2 gap-1.5">
          {options.map((opt) => (
            <button
              key={opt}
              disabled={saving}
              onClick={() => onSave(opt, note)}
              className={`flex items-center gap-2 rounded-sm border px-2.5 py-2 text-xs font-medium transition-colors hover:bg-muted/40 ${current?.status === opt ? 'border-current' : 'border-border'}`}
              style={current?.status === opt ? { color: COLOR[opt] } : undefined}
            >
              {opt === 'grey'
                ? <span className="h-2.5 w-2.5 rounded-[1px] border border-current opacity-40" />
                : <span className="h-2.5 w-2.5 rounded-[1px]" style={{ background: COLOR[opt] }} />}
              {STATUS_WORD[opt]}
            </button>
          ))}
        </div>
        <div className="mt-3 flex justify-end">
          <Button variant="ghost" size="sm" onClick={onClose}>Close</Button>
        </div>
      </div>
    </div>
  )
}
