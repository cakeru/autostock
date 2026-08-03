import { useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Disc, Plus, Camera, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { vehicleProfileApi } from '@/services/vehicleProfile'
import { CarDiagram } from './CarDiagram'
import { TirePanel } from './TirePanel'
import {
  useWheelServices, useCreateWheelService, useTireOptions,
  useCreatePart, usePartStatuses, useSetPartStatus,
} from '@/hooks/useVehicleProfile'
import type { CornerData, WheelService, WheelPosition, TireOption, PartColor, PartStatus, DueStatus } from '@/types/vehicleProfile'

const CORNERS: WheelPosition[] = ['FL', 'FR', 'RL', 'RR']
const POSITION_LABEL: Record<WheelPosition, string> = {
  FL: 'Front left', FR: 'Front right', RL: 'Rear left', RR: 'Rear right', SPARE: 'Spare',
}

// Current state of the car = for each corner, the most recent snapshot that
// actually touched that corner (fronts and rears may have been done on
// different visits), so the diagram shows what's on the car right now.
export function latestByPosition(services: WheelService[]): Record<string, { corner: CornerData; date: string }> {
  const out: Record<string, { corner: CornerData; date: string }> = {}
  for (const svc of services) { // already newest-first
    for (const corner of svc.corners) {
      if (!out[corner.position]) out[corner.position] = { corner, date: svc.performed_at }
    }
  }
  return out
}

const DUE_COLOR: Record<string, PartColor> = {
  overdue: 'red', due_soon: 'yellow', ok: 'green', unknown: 'grey',
}

// The car as it stands today: the DVI diagram of current tire/part condition,
// plus the entry points to log a new wheel service or part. Past visits live in
// the service-history timeline, not here.
export function CurrentCondition({ vehicleId, bodyType, due, unit }: {
  vehicleId: number
  bodyType?: string
  due?: DueStatus[]
  unit: string
}) {
  const qc = useQueryClient()
  const { data: services } = useWheelServices(vehicleId)
  const { data: tireOptions } = useTireOptions(vehicleId)
  const { data: partStatuses } = usePartStatuses(vehicleId)
  const createWheel = useCreateWheelService(vehicleId)
  const createPart = useCreatePart(vehicleId)
  const setPartStatus = useSetPartStatus(vehicleId)

  const [showDialog, setShowDialog] = useState(false)
  const [showParts, setShowParts] = useState(false)

  const current = useMemo(() => latestByPosition(services || []), [services])
  const cornerData = useMemo(() => {
    const out: Record<string, CornerData> = {}
    for (const [pos, entry] of Object.entries(current)) out[pos] = entry.corner
    return out
  }, [current])
  const statusMap = useMemo(() => {
    const out: Record<string, PartStatus> = {}
    for (const s of partStatuses || []) out[s.part_key] = s
    return out
  }, [partStatuses])
  const oilStatus: PartColor = DUE_COLOR[(due || []).find((d) => d.event_type === 'oil')?.status ?? 'unknown'] ?? 'grey'
  // Per-part reminder colour, keyed by part_key (skip "ok"/"unknown" — only a
  // due part should light the diagram when a tech hasn't set a colour).
  const partDue = useMemo(() => {
    const out: Record<string, PartColor> = {}
    for (const d of due || []) {
      if (d.event_type === 'part' && (d.status === 'overdue' || d.status === 'due_soon')) {
        out[d.key] = DUE_COLOR[d.status]
      }
    }
    return out
  }, [due])

  const handleCreate = (payload: Parameters<typeof createWheel.mutate>[0], file: File | null) => {
    createWheel.mutate(payload, {
      onSuccess: async (svc) => {
        if (file) {
          try {
            await vehicleProfileApi.addWheelServicePhoto(svc.id, file)
            qc.invalidateQueries({ queryKey: ['wheel-services', vehicleId] })
            qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
          } catch { /* photo is best-effort; the record itself saved */ }
        }
        setShowDialog(false)
      },
    })
  }

  return (
    <div className="bg-card rounded-lg p-5 shadow-sm">
      <div className="mb-3 flex items-center justify-between">
        <p className="flex items-center gap-1.5 text-sm font-semibold"><Disc className="h-4 w-4" /> Current condition</p>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setShowParts(true)}>
            <Plus className="mr-1 h-3.5 w-3.5" /> Log part
          </Button>
          <Button variant="outline" size="sm" onClick={() => setShowDialog(true)}>
            <Plus className="mr-1 h-3.5 w-3.5" /> Log wheel service
          </Button>
        </div>
      </div>

      <CarDiagram
        bodyType={bodyType}
        corners={cornerData}
        oilStatus={oilStatus}
        partStatuses={statusMap}
        partDue={partDue}
        onSetPart={(key, status, note) => setPartStatus.mutate({ part_key: key, status, note })}
        saving={setPartStatus.isPending}
      />
      <TirePanel entries={current} />

      {(!services || services.length === 0) && (
        <p className="mt-2 text-xs text-muted-foreground">
          No wheel records yet — tires show grey. Log a wheel service to fill in tread, age and alignment per corner. Tap a part on the diagram to set its condition.
        </p>
      )}

      {showDialog && (
        <WheelServiceDialog
          unit={unit}
          tireOptions={tireOptions || []}
          onClose={() => setShowDialog(false)}
          onCreate={handleCreate}
          loading={createWheel.isPending}
        />
      )}

      {showParts && (
        <AddPartDialog
          unit={unit}
          onClose={() => setShowParts(false)}
          onCreate={(data) => createPart.mutate(data, { onSuccess: () => setShowParts(false) })}
          loading={createPart.isPending}
        />
      )}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Wheel-service entry dialog
// ---------------------------------------------------------------------------

type CornerForm = {
  tireSel: string
  tire_brand: string; tire_size: string; tire_dot: string
  tread_before_mm: string; tread_mm: string; pressure: string
  camber_before: string; camber_after: string
  caster_before: string; caster_after: string
  toe_before: string; toe_after: string
  wear_note: string
}
const emptyCorner = (): CornerForm => ({
  tireSel: '', tire_brand: '', tire_size: '', tire_dot: '', tread_before_mm: '', tread_mm: '', pressure: '',
  camber_before: '', camber_after: '', caster_before: '', caster_after: '', toe_before: '', toe_after: '', wear_note: '',
})

function cornerHasData(f: CornerForm): boolean {
  return Object.values(f).some((v) => v !== '')
}

function WheelServiceDialog({ unit, tireOptions, onClose, onCreate, loading }: {
  unit: string
  tireOptions: TireOption[]
  onClose: () => void
  onCreate: (payload: { performed_at?: string; mileage?: number; notes?: string; corners: CornerData[] }, file: File | null) => void
  loading: boolean
}) {
  const [performedAt, setPerformedAt] = useState(new Date().toISOString().slice(0, 10))
  const [mileage, setMileage] = useState('')
  const [notes, setNotes] = useState('')
  const [includeAlignment, setIncludeAlignment] = useState(false)
  const [includeSpare, setIncludeSpare] = useState(false)
  const [file, setFile] = useState<File | null>(null)

  const positions = includeSpare ? [...CORNERS, 'SPARE' as WheelPosition] : CORNERS
  const [corners, setCorners] = useState<Record<string, CornerForm>>(
    () => Object.fromEntries(CORNERS.map((p) => [p, emptyCorner()])),
  )

  const update = (pos: string, patch: Partial<CornerForm>) =>
    setCorners((prev) => ({ ...prev, [pos]: { ...(prev[pos] ?? emptyCorner()), ...patch } }))

  const pickTire = (pos: string, sel: string) => {
    const opt = tireOptions.find((o) => String(o.product_id) === sel)
    if (opt) update(pos, { tireSel: sel, tire_brand: opt.name, tire_size: opt.size || '' })
    else update(pos, { tireSel: '' })
  }

  const submit = () => {
    const out: CornerData[] = []
    for (const pos of positions) {
      const f = corners[pos] ?? emptyCorner()
      if (!cornerHasData(f)) continue
      out.push({
        position: pos,
        tire_product_id: f.tireSel ? parseInt(f.tireSel, 10) : undefined,
        tire_brand: f.tire_brand || undefined,
        tire_size: f.tire_size || undefined,
        tire_dot: f.tire_dot || undefined,
        tread_mm: f.tread_mm ? parseFloat(f.tread_mm) : undefined,
        tread_before_mm: f.tread_before_mm ? parseFloat(f.tread_before_mm) : undefined,
        pressure: f.pressure ? parseFloat(f.pressure) : undefined,
        camber_before: f.camber_before || undefined,
        camber_after: f.camber_after || undefined,
        caster_before: f.caster_before || undefined,
        caster_after: f.caster_after || undefined,
        toe_before: f.toe_before || undefined,
        toe_after: f.toe_after || undefined,
        wear_note: f.wear_note || undefined,
      })
    }
    if (out.length === 0) return
    onCreate({
      performed_at: performedAt,
      mileage: mileage ? parseInt(mileage, 10) : undefined,
      notes: notes || undefined,
      corners: out,
    }, file)
  }

  const anyData = positions.some((p) => cornerHasData(corners[p] ?? emptyCorner()))

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4">
      <div className="my-6 w-full max-w-2xl rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Log Wheel Service</p>
        <p className="mb-3 text-xs text-muted-foreground">
          Fill only the corners you touched. Toggle alignment on for an alignment job.
        </p>

        <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
          <div className="space-y-1">
            <Label>Date</Label>
            <Input type="date" value={performedAt} onChange={(e) => setPerformedAt(e.target.value)} />
          </div>
          <div className="space-y-1">
            <Label>Odometer ({unit})</Label>
            <Input value={mileage} onChange={(e) => setMileage(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="82,000" />
          </div>
          <div className="col-span-2 space-y-1">
            <Label>Notes</Label>
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="e.g. 4-wheel alignment + front tires" />
          </div>
        </div>

        <div className="mt-3 flex flex-wrap gap-4">
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input type="checkbox" checked={includeAlignment} onChange={(e) => setIncludeAlignment(e.target.checked)} className="accent-primary" />
            Alignment job (show camber / caster / toe)
          </label>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input type="checkbox" checked={includeSpare} onChange={(e) => setIncludeSpare(e.target.checked)} className="accent-primary" />
            Include spare
          </label>
        </div>

        <div className="mt-3 space-y-3">
          {positions.map((pos) => {
            const f = corners[pos] ?? emptyCorner()
            return (
              <div key={pos} className="rounded-md border p-3">
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide">{POSITION_LABEL[pos]}</p>
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                  <div className="col-span-2 space-y-1 sm:col-span-1">
                    <Label className="text-[11px]">Tire (recent)</Label>
                    <Select value={f.tireSel} onChange={(e) => pickTire(pos, e.target.value)}>
                      <option value="">— manual —</option>
                      {tireOptions.map((o) => (
                        <option key={o.product_id} value={o.product_id}>{o.name}</option>
                      ))}
                    </Select>
                  </div>
                  <div className="space-y-1">
                    <Label className="text-[11px]">Brand</Label>
                    <Input value={f.tire_brand} onChange={(e) => update(pos, { tire_brand: e.target.value })} placeholder="Michelin" />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-[11px]">Size</Label>
                    <Input value={f.tire_size} onChange={(e) => update(pos, { tire_size: e.target.value })} placeholder="225/65R17" />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-[11px]">DOT (age)</Label>
                    <Input value={f.tire_dot} onChange={(e) => update(pos, { tire_dot: e.target.value })} placeholder="3823" />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-[11px]">Old tread (mm)</Label>
                    <Input value={f.tread_before_mm} onChange={(e) => update(pos, { tread_before_mm: e.target.value.replace(/[^\d.]/g, '') })} inputMode="decimal" placeholder="if replaced" />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-[11px]">Tread (mm)</Label>
                    <Input value={f.tread_mm} onChange={(e) => update(pos, { tread_mm: e.target.value.replace(/[^\d.]/g, '') })} inputMode="decimal" placeholder="7.5" />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-[11px]">Pressure (psi)</Label>
                    <Input value={f.pressure} onChange={(e) => update(pos, { pressure: e.target.value.replace(/[^\d.]/g, '') })} inputMode="decimal" placeholder="33" />
                  </div>
                  <div className="col-span-2 space-y-1">
                    <Label className="text-[11px]">Wear note</Label>
                    <Input value={f.wear_note} onChange={(e) => update(pos, { wear_note: e.target.value })} placeholder="inner edge wear" />
                  </div>
                </div>
                {includeAlignment && (
                  <div className="mt-2 grid grid-cols-3 gap-2 border-t pt-2">
                    <div className="text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
                      <p className="h-6" />
                      <p className="flex h-8 items-center">Camber</p>
                      <p className="flex h-8 items-center">Caster</p>
                      <p className="flex h-8 items-center">Toe</p>
                    </div>
                    <div>
                      <p className="mb-0.5 text-center text-[10px] text-muted-foreground">Before</p>
                      <Input className="mb-1 h-7" value={f.camber_before} onChange={(e) => update(pos, { camber_before: e.target.value })} placeholder="-1°01'" />
                      <Input className="mb-1 h-7" value={f.caster_before} onChange={(e) => update(pos, { caster_before: e.target.value })} placeholder="4°29'" />
                      <Input className="h-7" value={f.toe_before} onChange={(e) => update(pos, { toe_before: e.target.value })} placeholder="0.40" />
                    </div>
                    <div>
                      <p className="mb-0.5 text-center text-[10px] text-muted-foreground">After</p>
                      <Input className="mb-1 h-7" value={f.camber_after} onChange={(e) => update(pos, { camber_after: e.target.value })} placeholder="0°08'" />
                      <Input className="mb-1 h-7" value={f.caster_after} onChange={(e) => update(pos, { caster_after: e.target.value })} placeholder="4°12'" />
                      <Input className="h-7" value={f.toe_after} onChange={(e) => update(pos, { toe_after: e.target.value })} placeholder="0.05" />
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>

        <div className="mt-3">
          <label className="flex w-fit cursor-pointer items-center gap-2 rounded-md border px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted/50">
            <Camera className="h-3.5 w-3.5" />
            {file ? file.name : 'Attach alignment printout (optional)'}
            <input type="file" accept="image/*" className="hidden" onChange={(e) => setFile(e.target.files?.[0] || null)} />
          </label>
          {file && (
            <button onClick={() => setFile(null)} className="ml-2 text-xs text-muted-foreground hover:text-destructive">
              <Trash2 className="inline h-3 w-3" /> remove
            </button>
          )}
        </div>

        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} disabled={loading || !anyData}>
            {loading ? 'Saving...' : 'Save wheel record'}
          </Button>
        </div>
      </div>
    </div>
  )
}

// Inspection parts that carry a reminder — picking one ties the log entry to
// its reminder rule. "Other" leaves part_key blank (no reminder).
const PART_TYPES: { key: string; label: string }[] = [
  { key: 'brakes_front', label: 'Brakes (front)' },
  { key: 'brakes_rear', label: 'Brakes (rear)' },
  { key: 'battery', label: 'Battery' },
  { key: 'lights', label: 'Lights' },
  { key: 'wipers', label: 'Wipers' },
  { key: 'suspension', label: 'Suspension' },
  { key: 'air_filter', label: 'Air filter' },
  { key: 'coolant', label: 'Coolant' },
  { key: 'belts', label: 'Belts & hoses' },
  { key: 'chain_sprocket', label: 'Chain & sprocket' },
]

function AddPartDialog({ unit, onClose, onCreate, loading }: {
  unit: string
  onClose: () => void
  onCreate: (data: { part_name: string; part_key?: string; position?: string; replaced_at?: string; mileage?: number }) => void
  loading: boolean
}) {
  const [partKey, setPartKey] = useState('')
  const [partName, setPartName] = useState('')
  const [position, setPosition] = useState('')
  const [replacedAt, setReplacedAt] = useState(new Date().toISOString().slice(0, 10))
  const [mileage, setMileage] = useState('')

  // Selecting a known part type pre-fills the display name (still editable).
  const pickType = (key: string) => {
    setPartKey(key)
    const t = PART_TYPES.find((p) => p.key === key)
    if (t && (!partName || PART_TYPES.some((p) => p.label === partName))) setPartName(t.label)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-3 text-sm font-semibold">Log Part Replaced</p>
        <div className="space-y-2">
          <div className="space-y-1">
            <Label>Part type</Label>
            <Select value={partKey} onChange={(e) => pickType(e.target.value)}>
              <option value="">Other (no reminder)</option>
              {PART_TYPES.map((t) => (
                <option key={t.key} value={t.key}>{t.label}</option>
              ))}
            </Select>
            <p className="text-[11px] text-muted-foreground">Choosing a type schedules its next-due reminder.</p>
          </div>
          <div className="space-y-1">
            <Label>Name *</Label>
            <Input value={partName} onChange={(e) => setPartName(e.target.value)} placeholder="e.g. Brake pads (front)" />
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <Label>Position</Label>
              <Input value={position} onChange={(e) => setPosition(e.target.value)} placeholder="front / rear-left" />
            </div>
            <div className="space-y-1">
              <Label>Odometer ({unit})</Label>
              <Input value={mileage} onChange={(e) => setMileage(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="82,000" />
            </div>
          </div>
          <div className="space-y-1">
            <Label>Date</Label>
            <Input type="date" value={replacedAt} onChange={(e) => setReplacedAt(e.target.value)} />
          </div>
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button
            onClick={() => onCreate({ part_name: partName, part_key: partKey || undefined, position: position || undefined, replaced_at: replacedAt, mileage: mileage ? parseInt(mileage, 10) : undefined })}
            disabled={loading || !partName.trim()}
          >
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>
    </div>
  )
}
