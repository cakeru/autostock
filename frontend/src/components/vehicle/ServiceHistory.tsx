import { useRef, useState } from 'react'
import { Camera, Receipt, FileText, Eye, EyeOff } from 'lucide-react'
import { formatDate } from '@/utils/date'
import { formatUSD } from '@/utils/currency'
import { imageSrc } from '@/utils/imageUrl'
import { compressImage } from '@/utils/compressImage'
import { useNavigate } from 'react-router-dom'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { VisitLines } from './VisitLines'
import { useVehicleTimeline, useAddVisitPhoto, useSetPhotoVisibility, useDeleteServiceEvent } from '@/hooks/useVehicleProfile'
import type { Visit, VisitPhoto } from '@/types/vehicleProfile'

// The vehicle's whole service history as a vertical timeline — one card per
// visit, newest first, each showing everything done that day (tires, alignment
// blueprint, oil, parts, notes), what it cost, and photo evidence. Replaces the
// old split of wheel-history + activity list.
export function ServiceHistory({ vehicleId, bodyType, unit, onPhoto }: {
  vehicleId: number
  bodyType?: string
  unit: string
  onPhoto: (url: string) => void
}) {
  const { data: visits, isLoading } = useVehicleTimeline(vehicleId)
  const deleteEvent = useDeleteServiceEvent(vehicleId)
  const [pendingDelete, setPendingDelete] = useState<{ id: number; kind: 'oil' | 'tire' | 'service' } | null>(null)

  return (
    <div className="bg-card rounded-lg p-5 shadow-sm">
      <p className="mb-4 flex items-center gap-1.5 text-sm font-semibold"><Receipt className="h-4 w-4" /> Service history</p>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : !visits || visits.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No history yet. Sales, jobs, wheel services, oil changes and parts logged for this car will build a visit-by-visit timeline here.
        </p>
      ) : (
        <ol className="relative space-y-6 border-l border-border pl-6">
          {visits.map((v, i) => (
            <VisitCard key={`${v.date}-${i}`} visit={v} vehicleId={vehicleId} bodyType={bodyType} unit={unit} onPhoto={onPhoto}
              onRemoveEvent={(id, kind) => setPendingDelete({ id, kind })} />
          ))}
        </ol>
      )}

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        onConfirm={() => pendingDelete && deleteEvent.mutate(pendingDelete.id, { onSuccess: () => setPendingDelete(null) })}
        title={`Remove ${pendingDelete?.kind === 'oil' ? 'oil change' : pendingDelete?.kind === 'service' ? 'service' : 'tire'} record`}
        message="Remove this auto-logged record? The due-for-service estimate will recompute from what's left. The sale/invoice itself is not affected."
        destructive
        loading={deleteEvent.isPending}
      />
    </div>
  )
}

function VisitCard({ visit: v, vehicleId, bodyType, unit, onPhoto, onRemoveEvent }: {
  visit: Visit
  vehicleId: number
  bodyType?: string
  unit: string
  onPhoto: (url: string) => void
  onRemoveEvent: (eventId: number, kind: 'oil' | 'tire' | 'service') => void
}) {
  const navigate = useNavigate()
  const fileRef = useRef<HTMLInputElement>(null)
  const addPhoto = useAddVisitPhoto(vehicleId)
  const setVisibility = useSetPhotoVisibility(vehicleId)

  const visitDate = v.date.slice(0, 10) // for dating the uploaded photo to this visit

  return (
    <li className="relative">
      {/* node on the rail */}
      <span className="absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-primary bg-card" />

      <div className="flex items-baseline justify-between gap-2">
        <p className="text-sm font-semibold">{formatDate(v.date)}</p>
        {v.mileage != null && <p className="text-xs tabular-nums text-muted-foreground">{v.mileage.toLocaleString()} {unit}</p>}
      </div>

      <div className="mt-2 space-y-2.5">
        <VisitLines visit={v} bodyType={bodyType} onRemoveEvent={onRemoveEvent} />

        {/* Photos — each gallery photo can be shared to / hidden from the customer report */}
        {(v.photos && v.photos.length > 0) && (
          <div className="flex flex-wrap gap-2 pt-0.5">
            {v.photos.map((ph, i) => (
              <VisitPhotoThumb
                key={ph.id ?? i}
                photo={ph}
                onClick={() => onPhoto(ph.url)}
                onToggleShare={ph.source === 'gallery' && ph.id
                  ? () => setVisibility.mutate({ photoId: ph.id!, visible: !ph.customer_visible })
                  : undefined}
              />
            ))}
          </div>
        )}

        {/* Linked sale / job chips (internal only) */}
        {(v.transactions && v.transactions.length > 0) && (
          <div className="flex flex-wrap gap-1.5 pt-0.5">
            {v.transactions.map((t) => (
              <button
                key={`${t.type}-${t.id}`}
                onClick={() => navigate(t.type === 'invoice' ? `/invoices/${t.id}` : `/service-jobs/${t.id}`)}
                className="inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-xs hover:bg-muted/50"
              >
                {t.type === 'invoice' ? <Receipt className="h-3 w-3 text-primary" /> : <FileText className="h-3 w-3 text-accent" />}
                <span className="font-mono">{t.ref}</span>
                <span className="tabular-nums text-muted-foreground">{formatUSD(t.amount)}</span>
              </button>
            ))}
          </div>
        )}

        {/* Add photo evidence to this visit */}
        <div>
          <button
            onClick={() => fileRef.current?.click()}
            disabled={addPhoto.isPending}
            className="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground disabled:opacity-50"
          >
            <Camera className="h-3.5 w-3.5" /> {addPhoto.isPending ? 'Uploading…' : 'Add photo'}
          </button>
          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={async (e) => {
              const file = e.target.files?.[0]
              if (file) addPhoto.mutate({ file: await compressImage(file), takenAt: visitDate })
              if (fileRef.current) fileRef.current.value = ''
            }}
          />
        </div>
      </div>
    </li>
  )
}

function VisitPhotoThumb({ photo, onClick, onToggleShare }: {
  photo: VisitPhoto
  onClick: () => void
  onToggleShare?: () => void
}) {
  const shared = !!photo.customer_visible
  return (
    <div className="group relative h-20 w-20 flex-shrink-0">
      <button onClick={onClick} className="block h-full w-full overflow-hidden rounded border">
        <img src={imageSrc(photo.url)} alt={photo.caption || 'Service photo'} className="h-full w-full object-cover" />
        {photo.phase && (
          <span className="absolute left-0 top-0 rounded-br bg-black/60 px-1 py-0.5 text-[9px] font-medium uppercase tracking-wide text-white">
            {photo.phase}
          </span>
        )}
        {photo.caption && (
          <span className="absolute inset-x-0 bottom-0 truncate bg-black/60 px-1 py-0.5 text-[9px] text-white">
            {photo.caption}
          </span>
        )}
      </button>
      {onToggleShare && (
        <button
          onClick={onToggleShare}
          title={shared ? 'Shown on customer report — click to hide' : 'Hidden from customer — click to share'}
          aria-label={shared ? 'Hide from customer report' : 'Share on customer report'}
          className={`absolute -right-1.5 -top-1.5 grid h-5 w-5 place-items-center rounded-full border shadow transition-colors ${
            shared ? 'bg-primary text-primary-foreground border-primary' : 'bg-card text-muted-foreground opacity-0 group-hover:opacity-100'
          }`}
        >
          {shared ? <Eye className="h-3 w-3" /> : <EyeOff className="h-3 w-3" />}
        </button>
      )}
    </div>
  )
}
