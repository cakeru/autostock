import { useState } from 'react'
import { Wallet } from 'lucide-react'
import { useCurrentShift, useShiftHistory, useOpenShift, useCloseShift } from '@/hooks/useCashShift'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatUSD } from '@/utils/currency'
import { formatDateTime } from '@/utils/date'
import type { CashShift } from '@/services/cashshift'

function overShortLabel(v: number) {
  if (Math.abs(v) < 0.005) return { text: 'Balanced', tone: 'text-emerald-600' }
  if (v > 0) return { text: `${formatUSD(v)} over`, tone: 'text-emerald-600' }
  return { text: `${formatUSD(-v)} short`, tone: 'text-destructive' }
}

export function CashDrawer() {
  const { data: shift, isLoading } = useCurrentShift()
  const { data: history } = useShiftHistory(6)
  const openMut = useOpenShift()
  const closeMut = useCloseShift()

  const [openingFloat, setOpeningFloat] = useState('')
  const [counted, setCounted] = useState('')
  const [note, setNote] = useState('')
  const [closing, setClosing] = useState(false)
  const [justClosed, setJustClosed] = useState<CashShift | null>(null)

  const doOpen = () => {
    const f = parseFloat(openingFloat)
    if (isNaN(f) || f < 0) return
    openMut.mutate(f, { onSuccess: () => { setOpeningFloat(''); setJustClosed(null) } })
  }
  const doClose = () => {
    const c = parseFloat(counted)
    if (isNaN(c) || c < 0) return
    closeMut.mutate({ amount: c, note: note.trim() || undefined }, {
      onSuccess: (s) => { setJustClosed(s); setClosing(false); setCounted(''); setNote('') },
    })
  }

  const recent = (history || []).filter((s) => s.status === 'closed')

  return (
    <div className="bg-card rounded-lg p-5 shadow-sm">
      <div className="mb-3 flex items-center gap-2">
        <Wallet className="h-4 w-4 text-muted-foreground" />
        <h2 className="text-sm font-semibold">Cash drawer</h2>
      </div>

      {isLoading ? (
        <Skeleton className="h-24 w-full" />
      ) : shift ? (
        <div className="space-y-3">
          <p className="flex items-center gap-2 text-sm">
            <span className="inline-block h-2 w-2 rounded-full bg-emerald-500" />
            <span className="font-medium">Open</span>
            <span className="text-muted-foreground">by {shift.opened_by_name || 'staff'} · since {formatDateTime(shift.opened_at)}</span>
          </p>
          <div className="space-y-1 text-sm">
            <Row label="Opening float" value={formatUSD(shift.opening_float)} />
            <Row label="Cash sales this shift" value={formatUSD(shift.cash_sales)} />
            <div className="flex justify-between border-t pt-1 font-semibold">
              <span>Expected in drawer</span>
              <span className="tabular-nums">{formatUSD(shift.expected_amount)}</span>
            </div>
          </div>

          {closing ? (
            <div className="rounded-md border p-3 space-y-2">
              <p className="text-xs font-medium text-muted-foreground">Count the drawer, then confirm</p>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Counted $</span>
                <Input value={counted} onChange={(e) => setCounted(e.target.value)} type="number" step="0.01" min="0" className="w-32" autoFocus />
              </div>
              <Input value={note} onChange={(e) => setNote(e.target.value)} placeholder="Note (optional — explain any variance)" />
              {counted !== '' && !isNaN(parseFloat(counted)) && (
                <p className={`text-sm font-medium ${overShortLabel(parseFloat(counted) - shift.expected_amount).tone}`}>
                  {overShortLabel(parseFloat(counted) - shift.expected_amount).text}
                </p>
              )}
              <div className="flex gap-2">
                <Button variant="ghost" size="sm" onClick={() => { setClosing(false); setCounted(''); setNote('') }}>Cancel</Button>
                <Button size="sm" className="flex-1" disabled={counted === '' || closeMut.isPending} onClick={doClose}>
                  {closeMut.isPending ? 'Closing…' : 'Close drawer'}
                </Button>
              </div>
            </div>
          ) : (
            <Button variant="outline" size="sm" className="w-full" onClick={() => setClosing(true)}>Close drawer</Button>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {justClosed && (
            <div className="rounded-md border bg-muted/40 p-3 text-sm">
              <p className="font-medium">Drawer closed</p>
              <div className="mt-1 space-y-0.5 text-muted-foreground">
                <div className="flex justify-between"><span>Expected</span><span className="tabular-nums">{formatUSD(justClosed.expected_amount)}</span></div>
                <div className="flex justify-between"><span>Counted</span><span className="tabular-nums">{formatUSD(justClosed.closing_amount || 0)}</span></div>
                <div className={`flex justify-between font-medium ${overShortLabel(justClosed.over_short || 0).tone}`}>
                  <span>Result</span><span>{overShortLabel(justClosed.over_short || 0).text}</span>
                </div>
              </div>
            </div>
          )}
          <p className="text-sm text-muted-foreground">No drawer open.</p>
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Opening float $</span>
            <Input value={openingFloat} onChange={(e) => setOpeningFloat(e.target.value)} type="number" step="0.01" min="0" className="w-32" />
            <Button size="sm" disabled={openingFloat === '' || openMut.isPending} onClick={doOpen}>
              {openMut.isPending ? 'Opening…' : 'Open drawer'}
            </Button>
          </div>
        </div>
      )}

      {recent.length > 0 && (
        <div className="mt-4 border-t pt-3">
          <p className="mb-1.5 text-xs font-medium text-muted-foreground">Recent shifts</p>
          <div className="divide-y divide-border/70 text-sm">
            {recent.map((s) => {
              const os = overShortLabel(s.over_short || 0)
              return (
                <div key={s.id} className="flex items-center justify-between py-1.5">
                  <span className="text-muted-foreground">{s.closed_at ? formatDateTime(s.closed_at) : '—'} · {s.opened_by_name || 'staff'}</span>
                  <span className={`tabular-nums ${os.tone}`}>{os.text}</span>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-muted-foreground">
      <span>{label}</span>
      <span className="tabular-nums">{value}</span>
    </div>
  )
}
