import { useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Car, StickyNote, X, Gauge, Disc, Droplet, Settings2, Share2, Copy, Check } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { VehicleForm } from '@/components/customer/VehicleForm'
import { CurrentCondition } from '@/components/vehicle/CurrentCondition'
import { ServiceHistory } from '@/components/vehicle/ServiceHistory'
import { formatDate } from '@/utils/date'
import { imageSrc } from '@/utils/imageUrl'
import { useUpdateVehicle } from '@/hooks/useCustomers'
import {
  useVehicleProfile, useCreateVehicleRecord,
  useCreateServiceEvent, useUpdateVehicleIntervals,
  useEnsureShareLink, useRevokeShareLink,
} from '@/hooks/useVehicleProfile'
import type { DueStatus, ServiceEventType, VehicleProfile, UpdateVehicleIntervalsRequest } from '@/types/vehicleProfile'

export function VehicleDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const vehicleId = parseInt(id || '0')

  const { data: vehicle, isLoading } = useVehicleProfile(vehicleId)
  const updateVehicle = useUpdateVehicle()
  const createRecord = useCreateVehicleRecord(vehicleId)
  const createServiceEvent = useCreateServiceEvent(vehicleId)
  const updateIntervals = useUpdateVehicleIntervals(vehicleId)

  const [showEdit, setShowEdit] = useState(false)
  const [showAddRecord, setShowAddRecord] = useState(false)
  const [lightboxUrl, setLightboxUrl] = useState<string | null>(null)
  const [logType, setLogType] = useState<ServiceEventType | null>(null)
  const [showIntervals, setShowIntervals] = useState(false)
  const [showShare, setShowShare] = useState(false)

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!vehicle) return <p className="text-sm text-destructive">Vehicle not found</p>

  // Each vehicle has its own unit — a miles-only import stays in miles.
  const unit = vehicle.distance_unit === 'mi' ? 'mi' : 'km'

  const specLine = [vehicle.make, vehicle.model].filter(Boolean).join(' ') + (vehicle.year ? ` (${vehicle.year})` : '')

  return (
    <div className="space-y-6">
      <PageHeader
        title={vehicle.plate_number}
        backTo={`/customers/${vehicle.customer_id}`}
        subtitle={specLine.trim() || undefined}
        actions={
          <>
            <Button variant="outline" size="sm" onClick={() => setShowShare(true)}>
              <Share2 className="mr-1.5 h-3.5 w-3.5" /> Share
            </Button>
            <Button variant="outline" size="sm" onClick={() => setShowAddRecord(true)}>
              <StickyNote className="mr-1.5 h-3.5 w-3.5" /> Add Note
            </Button>
            <Button variant="outline" size="sm" onClick={() => setShowEdit(true)}>Edit</Button>
          </>
        }
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
          <p className="text-sm font-semibold flex items-center gap-1.5"><Car className="h-4 w-4" /> Vehicle</p>
          {vehicle.vin && <p className="text-sm text-muted-foreground">VIN: {vehicle.vin}</p>}
          {vehicle.color && <p className="text-sm text-muted-foreground">Color: {vehicle.color}</p>}
          {vehicle.notes && <p className="text-sm text-muted-foreground">{vehicle.notes}</p>}
          {!vehicle.vin && !vehicle.color && !vehicle.notes && <p className="text-sm text-muted-foreground">No additional details</p>}
        </div>

        <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
          <p className="text-sm font-semibold">Owner</p>
          <p className="text-sm cursor-pointer text-primary hover:underline" onClick={() => navigate(`/customers/${vehicle.customer_id}`)}>
            {vehicle.customer_name}
          </p>
          {vehicle.customer_phone && <p className="text-sm text-muted-foreground">{vehicle.customer_phone}</p>}
        </div>

        <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
          <p className="text-sm font-semibold flex items-center gap-1.5"><Gauge className="h-4 w-4" /> Last known</p>
          <p className="text-sm text-muted-foreground">
            {vehicle.last_mileage != null ? `${vehicle.last_mileage.toLocaleString()} ${unit}` : 'No mileage recorded'}
          </p>
          <p className="text-sm text-muted-foreground">
            {vehicle.last_service_at ? formatDate(vehicle.last_service_at) : 'No service history'}
          </p>
        </div>
      </div>

      <div className="bg-card rounded-lg p-5 shadow-sm">
        <div className="mb-3 flex items-center justify-between">
          <p className="text-sm font-semibold">Service Reminders</p>
          <button onClick={() => setShowIntervals(true)} className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground">
            <Settings2 className="h-3.5 w-3.5" /> Intervals
          </button>
        </div>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {vehicle.due.filter((d) => d.event_type === 'oil' || d.event_type === 'tire').map((d) => (
            <DueCard key={d.key} due={d} unit={unit} onLog={() => setLogType(d.event_type as ServiceEventType)} />
          ))}
        </div>
      </div>

      <CurrentCondition vehicleId={vehicleId} bodyType={vehicle.body_type} due={vehicle.due} unit={unit} />

      <ServiceHistory vehicleId={vehicleId} bodyType={vehicle.body_type} unit={unit} onPhoto={setLightboxUrl} />

      {showEdit && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-md rounded-lg bg-card p-5 shadow-lg">
            <p className="mb-3 text-sm font-semibold">Edit Vehicle</p>
            <VehicleForm
              initial={vehicle}
              onSubmit={(data) => updateVehicle.mutate({ id: vehicleId, data }, { onSuccess: () => setShowEdit(false) })}
              onCancel={() => setShowEdit(false)}
              loading={updateVehicle.isPending}
            />
          </div>
        </div>
      )}

      {showAddRecord && (
        <AddRecordDialog
          unit={unit}
          onClose={() => setShowAddRecord(false)}
          onCreate={(data) => createRecord.mutate(data, { onSuccess: () => setShowAddRecord(false) })}
          loading={createRecord.isPending}
        />
      )}

      {logType && (
        <LogServiceEventDialog
          unit={unit}
          eventType={logType}
          onClose={() => setLogType(null)}
          onCreate={(data) => createServiceEvent.mutate(data, { onSuccess: () => setLogType(null) })}
          loading={createServiceEvent.isPending}
        />
      )}

      {showIntervals && (
        <IntervalsDialog
          vehicle={vehicle}
          unit={unit}
          onClose={() => setShowIntervals(false)}
          onSave={(data) => updateIntervals.mutate(data, { onSuccess: () => setShowIntervals(false) })}
          loading={updateIntervals.isPending}
        />
      )}

      {showShare && (
        <ShareDialog vehicle={vehicle} vehicleId={vehicleId} onClose={() => setShowShare(false)} />
      )}

      {lightboxUrl && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/85 p-4" onClick={() => setLightboxUrl(null)}>
          <button className="absolute right-4 top-4 text-white" onClick={() => setLightboxUrl(null)} aria-label="Close">
            <X className="h-6 w-6" />
          </button>
          <img src={imageSrc(lightboxUrl)} alt="Record photo" className="max-h-full max-w-full rounded object-contain" />
        </div>
      )}
    </div>
  )
}

function AddRecordDialog({ unit, onClose, onCreate, loading }: {
  unit: string
  onClose: () => void
  onCreate: (data: { note?: string; mileage?: number }) => void
  loading: boolean
}) {
  const [note, setNote] = useState('')
  const [mileage, setMileage] = useState('')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Add Note</p>
        <p className="mb-3 text-xs text-muted-foreground">
          Jot down anything worth remembering about this car — alignment readings, damage, a customer request. It appears on the service timeline. (Add photos from the timeline entry itself.)
        </p>
        <div className="space-y-2">
          <div className="space-y-1">
            <Label>Odometer ({unit})</Label>
            <Input value={mileage} onChange={(e) => setMileage(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="e.g. 87450" />
          </div>
          <div className="space-y-1">
            <Label>Note</Label>
            <Textarea value={note} onChange={(e) => setNote(e.target.value)} rows={4} placeholder="e.g. Toe 0.12°, camber -0.5°, recommend tie-rod next visit" />
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button
            onClick={() => onCreate({ note: note || undefined, mileage: mileage ? parseInt(mileage) : undefined })}
            disabled={loading || (!note.trim() && !mileage)}
          >
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>
    </div>
  )
}

const STATUS_STYLE: Record<string, string> = {
  overdue: 'bg-destructive/10 text-destructive',
  due_soon: 'bg-amber-500/10 text-amber-600',
  ok: 'bg-primary/10 text-primary',
  unknown: 'bg-muted text-muted-foreground',
}
const STATUS_LABEL: Record<string, string> = {
  overdue: 'Overdue', due_soon: 'Due soon', ok: 'On track', unknown: 'No history',
}

function DueCard({ due, unit, onLog }: { due: DueStatus; unit: string; onLog: () => void }) {
  const Icon = due.event_type === 'oil' ? Droplet : Disc
  const label = due.event_type === 'oil' ? 'Oil Change' : 'Tires'

  return (
    <div className="rounded-md border p-3">
      <div className="flex items-center justify-between">
        <p className="flex items-center gap-1.5 text-sm font-medium"><Icon className="h-4 w-4" /> {label}</p>
        <span className={`rounded-full px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide ${STATUS_STYLE[due.status]}`}>
          {STATUS_LABEL[due.status]}
        </span>
      </div>
      {due.status === 'unknown' ? (
        <p className="mt-1.5 text-xs text-muted-foreground">No {due.event_type} sale or record logged yet for this car.</p>
      ) : (
        <div className="mt-1.5 space-y-0.5 text-xs text-muted-foreground">
          <p>Last: {due.last_service_at ? formatDate(due.last_service_at) : '—'}{due.last_mileage != null ? ` · ${due.last_mileage.toLocaleString()} ${unit}` : ''}</p>
          <p>
            Next due: {due.due_date ? formatDate(due.due_date) : '—'}
            {due.due_mileage != null ? ` · by ${due.due_mileage.toLocaleString()} ${unit}` : ''}
          </p>
        </div>
      )}
      <Button variant="outline" size="sm" className="mt-2 w-full" onClick={onLog}>
        Log {due.event_type === 'oil' ? 'oil change' : 'tire install'}
      </Button>
    </div>
  )
}

function LogServiceEventDialog({ unit, eventType, onClose, onCreate, loading }: {
  unit: string
  eventType: ServiceEventType
  onClose: () => void
  onCreate: (data: { event_type: ServiceEventType; mileage?: number; occurred_at?: string; product_name?: string; life_km?: number }) => void
  loading: boolean
}) {
  const [mileage, setMileage] = useState('')
  const [occurredAt, setOccurredAt] = useState(new Date().toISOString().slice(0, 10))
  const [productName, setProductName] = useState('')
  const [lifeKm, setLifeKm] = useState('')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Log {eventType === 'oil' ? 'Oil Change' : 'Tire Install'}</p>
        <p className="mb-3 text-xs text-muted-foreground">
          For a service done outside a normal sale (or to backfill history from before this feature existed).
        </p>
        <div className="space-y-2">
          <div className="space-y-1">
            <Label>Date</Label>
            <Input value={occurredAt} onChange={(e) => setOccurredAt(e.target.value)} type="date" />
          </div>
          <div className="space-y-1">
            <Label>Odometer ({unit})</Label>
            <Input value={mileage} onChange={(e) => setMileage(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="e.g. 87450" />
          </div>
          <div className="space-y-1">
            <Label>Product / notes (optional)</Label>
            <Input value={productName} onChange={(e) => setProductName(e.target.value)} placeholder={eventType === 'oil' ? 'e.g. 5W-30 full synthetic' : 'e.g. 4x Michelin 205/55R16'} />
          </div>
          {eventType === 'tire' && (
            <div className="space-y-1">
              <Label>Tire life ({unit})</Label>
              <Input value={lifeKm} onChange={(e) => setLifeKm(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="Shop default if blank" />
              <p className="text-[11px] text-muted-foreground">How many {unit} these tires should last before the next change.</p>
            </div>
          )}
        </div>
        <div className="flex justify-end gap-2 pt-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button
            onClick={() => onCreate({ event_type: eventType, mileage: mileage ? parseInt(mileage) : undefined, occurred_at: occurredAt, product_name: productName || undefined, life_km: eventType === 'tire' && lifeKm ? parseInt(lifeKm) : undefined })}
            disabled={loading}
          >
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>
    </div>
  )
}

function IntervalsDialog({ vehicle, unit, onClose, onSave, loading }: {
  vehicle: { oil_interval_km?: number; oil_interval_days?: number; tire_interval_km?: number; tire_interval_days?: number }
  unit: string
  onClose: () => void
  onSave: (data: UpdateVehicleIntervalsRequest) => void
  loading: boolean
}) {
  const [oilKm, setOilKm] = useState(vehicle.oil_interval_km?.toString() || '')
  const [oilDays, setOilDays] = useState(vehicle.oil_interval_days?.toString() || '')
  const [tireKm, setTireKm] = useState(vehicle.tire_interval_km?.toString() || '')
  const [tireDays, setTireDays] = useState(vehicle.tire_interval_days?.toString() || '')
  const num = (s: string) => s.replace(/[^\d]/g, '')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Reminder Overrides for This Vehicle</p>
        <p className="mb-3 text-xs text-muted-foreground">
          Leave blank to use the shop defaults. Set a <span className="font-medium text-foreground">days</span> interval for a firm calendar due date; a <span className="font-medium text-foreground">{unit}</span> interval is projected from how far the car is driven.
        </p>
        <div className="space-y-4">
          <div>
            <Label className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Oil change</Label>
            <div className="mt-1 grid grid-cols-2 gap-2">
              <Input value={oilKm} onChange={(e) => setOilKm(num(e.target.value))} inputMode="numeric" placeholder={`${unit} (default)`} />
              <Input value={oilDays} onChange={(e) => setOilDays(num(e.target.value))} inputMode="numeric" placeholder="days (e.g. 90)" />
            </div>
          </div>
          <div>
            <Label className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tires</Label>
            <div className="mt-1 grid grid-cols-2 gap-2">
              <Input value={tireKm} onChange={(e) => setTireKm(num(e.target.value))} inputMode="numeric" placeholder={`${unit} life (default)`} />
              <Input value={tireDays} onChange={(e) => setTireDays(num(e.target.value))} inputMode="numeric" placeholder="days (optional)" />
            </div>
            <p className="mt-1 text-[11px] text-muted-foreground">{unit} life is used when a sold tire has no rated life of its own. Whichever limit comes first wins.</p>
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-4">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button
            onClick={() => onSave({
              oil_interval_km: oilKm ? parseInt(oilKm) : null,
              oil_interval_days: oilDays ? parseInt(oilDays) : null,
              tire_interval_km: tireKm ? parseInt(tireKm) : null,
              tire_interval_days: tireDays ? parseInt(tireDays) : null,
            })}
            disabled={loading}
          >
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>
    </div>
  )
}

function ShareDialog({ vehicle, vehicleId, onClose }: {
  vehicle: VehicleProfile
  vehicleId: number
  onClose: () => void
}) {
  const ensure = useEnsureShareLink(vehicleId)
  const revoke = useRevokeShareLink(vehicleId)
  const [copied, setCopied] = useState(false)

  const url = vehicle.share_token ? `${window.location.origin}/report/${vehicle.share_token}` : ''

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* clipboard unavailable (http) — the input stays selectable */ }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-md rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Customer Report Link</p>
        <p className="mb-4 text-xs text-muted-foreground">
          A read-only condition report for this car — the customer can open it on their phone,
          keep it, or print it. No login needed; the link is its own key.
        </p>

        {vehicle.share_token ? (
          <>
            <div className="flex items-center gap-2">
              <Input readOnly value={url} onFocus={(e) => e.target.select()} className="font-mono text-xs" />
              <Button variant="outline" size="sm" onClick={copy} className="flex-shrink-0">
                {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
              </Button>
            </div>
            <div className="mt-4 flex items-center justify-between">
              <button
                onClick={() => revoke.mutate(undefined, { onSuccess: onClose })}
                disabled={revoke.isPending}
                className="text-xs text-destructive hover:underline"
              >
                {revoke.isPending ? 'Revoking…' : 'Revoke link'}
              </button>
              <Button variant="ghost" size="sm" onClick={onClose}>Close</Button>
            </div>
            <p className="mt-2 text-[11px] text-muted-foreground">
              Revoking kills the old link immediately; sharing again makes a fresh one.
            </p>
          </>
        ) : (
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>Cancel</Button>
            <Button onClick={() => ensure.mutate()} disabled={ensure.isPending}>
              {ensure.isPending ? 'Creating…' : 'Create link'}
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
