import { useState } from 'react'
import { toast } from 'sonner'
import { QrCode, ScanLine, Check, RotateCcw, Car } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { QrScanner } from '@/components/scan/QrScanner'
import { useSettings } from '@/hooks/useSettings'
import { useOpenJobs, useMechanics, useResolveBatch, useRecordInstall } from '@/hooks/useBatchInstall'
import type { BatchInfo } from '@/services/batchInstall'

const POSITIONS = ['FL', 'FR', 'RL', 'RR', 'SPARE']

// Mechanic-facing scan-to-install: scan a batch QR → confirm the lot → attach it
// to a car that's on the ramp today → record. It never touches stock or the
// sale; it just logs the actual batch fitted, for accurate recalls.
export function ScanInstall() {
  const { data: settings } = useSettings()
  const [scanning, setScanning] = useState(true)
  const [manual, setManual] = useState('')
  const [batch, setBatch] = useState<BatchInfo | null>(null)
  const [jobId, setJobId] = useState('')
  const [position, setPosition] = useState('')
  const [mechanicId, setMechanicId] = useState('')

  const resolve = useResolveBatch()
  const record = useRecordInstall()
  const { data: openJobs } = useOpenJobs(!!batch)
  const { data: mechanics } = useMechanics(!!batch)

  if (settings && !settings.feature_batch_scan) {
    return (
      <div className="mx-auto max-w-md py-16 text-center">
        <QrCode className="mx-auto mb-3 h-10 w-10 text-muted-foreground" />
        <p className="text-sm font-medium">Batch scanning is turned off</p>
        <p className="mt-1 text-sm text-muted-foreground">Enable it under Settings → Inventory to use scan-to-install.</p>
      </div>
    )
  }

  const doResolve = (code: string) => {
    setScanning(false)
    resolve.mutate(code, {
      onSuccess: (info) => setBatch(info),
      onError: () => { toast.error('Batch not recognised'); setScanning(true) },
    })
  }

  const reset = () => {
    setBatch(null); setJobId(''); setPosition(''); setMechanicId(''); setManual(''); setScanning(true)
  }

  const submit = () => {
    if (!batch || !jobId) return
    record.mutate(
      {
        batch_id: batch.batch_id,
        service_job_id: parseInt(jobId),
        position: position || undefined,
        mechanic_employee_id: mechanicId ? parseInt(mechanicId) : undefined,
      },
      {
        onSuccess: (r) => { toast.success(`Recorded ${r.product_name} on ${r.plate_number || r.job_number}`); reset() },
        onError: () => toast.error('Could not record install'),
      },
    )
  }

  return (
    <div className="mx-auto max-w-md space-y-5">
      <PageHeader title="Scan to install" subtitle="Log the exact batch fitted to a car" />

      {!batch ? (
        <div className="space-y-4">
          {scanning ? (
            <QrScanner onResult={doResolve} />
          ) : (
            <div className="rounded-lg border border-dashed p-6 text-center">
              {resolve.isPending
                ? <p className="text-sm text-muted-foreground">Looking up batch…</p>
                : <Button variant="outline" onClick={() => setScanning(true)}><ScanLine className="mr-1.5 h-4 w-4" /> Start camera</Button>}
            </div>
          )}

          <div className="rounded-lg border p-3">
            <p className="mb-2 text-xs font-medium text-muted-foreground">Or type the batch label</p>
            <div className="flex gap-2">
              <Input value={manual} onChange={(e) => setManual(e.target.value)} placeholder="e.g. B-2026-0006 or KSB:6" />
              <Button onClick={() => manual.trim() && doResolve(manual.trim())} disabled={!manual.trim() || resolve.isPending}>Find</Button>
            </div>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {/* Resolved batch */}
          <div className="rounded-lg border bg-card p-4">
            <div className="flex items-center gap-2">
              <QrCode className="h-4 w-4 text-primary" />
              <p className="text-sm font-semibold">{batch.product_name}</p>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              <span className="font-mono">{batch.batch_no}</span>
              {batch.tire_size ? ` · ${batch.tire_size}` : ''}
              {batch.dot_code ? ` · DOT ${batch.dot_code}` : ''}
              {` · ${batch.quantity_remaining} on shelf`}
            </p>
          </div>

          {/* Pick today's open job */}
          <div className="space-y-1.5">
            <label className="flex items-center gap-1.5 text-sm font-medium"><Car className="h-3.5 w-3.5" /> Which car?</label>
            <Select value={jobId} onChange={(e) => setJobId(e.target.value)}>
              <option value="">Select a job on the ramp…</option>
              {(openJobs || []).map((j) => (
                <option key={j.id} value={j.id}>
                  {j.plate_number || j.job_number}{j.make || j.model ? ` · ${[j.make, j.model].filter(Boolean).join(' ')}` : ''}{j.customer_name ? ` · ${j.customer_name}` : ''}
                </option>
              ))}
            </Select>
            {openJobs && openJobs.length === 0 && (
              <p className="text-xs text-muted-foreground">No open jobs right now. Create/open a job for the car first.</p>
            )}
          </div>

          {/* Optional position */}
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Position <span className="text-xs font-normal text-muted-foreground">(optional)</span></label>
            <div className="flex flex-wrap gap-1.5">
              {POSITIONS.map((p) => (
                <button
                  key={p}
                  onClick={() => setPosition(position === p ? '' : p)}
                  className={`rounded-md border px-3 py-1.5 text-sm ${position === p ? 'border-primary bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>

          {/* Optional mechanic */}
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Fitted by <span className="text-xs font-normal text-muted-foreground">(optional)</span></label>
            <Select value={mechanicId} onChange={(e) => setMechanicId(e.target.value)}>
              <option value="">—</option>
              {(mechanics || []).map((m) => <option key={m.id} value={m.id}>{m.name}</option>)}
            </Select>
          </div>

          <div className="flex gap-2 pt-1">
            <Button variant="ghost" onClick={reset}><RotateCcw className="mr-1.5 h-4 w-4" /> Scan another</Button>
            <Button className="flex-1" onClick={submit} disabled={!jobId || record.isPending}>
              <Check className="mr-1.5 h-4 w-4" /> {record.isPending ? 'Recording…' : 'Record install'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
