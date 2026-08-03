import { useState } from 'react'
import { Play, Plus, Trash2, Pencil, Download, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { downloadFile } from '@/utils/download'
import {
  useBackupSchedules,
  useCreateBackupSchedule,
  useDeleteBackupSchedule,
  useRunBackupSchedule,
  useUpdateBackupSchedule,
} from '@/hooks/useBackupSchedules'
import type { BackupSchedule } from '@/types/backup'

const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

const pad = (n: string) => n.padStart(2, '0')

function describeCron(cron: string): string {
  const daily = cron.match(/^(\d+) (\d+) \* \* \*$/)
  if (daily) return `Daily at ${pad(daily[2])}:${pad(daily[1])}`
  const weekly = cron.match(/^(\d+) (\d+) \* \* (\d)$/)
  if (weekly) return `Weekly on ${WEEKDAYS[Number(weekly[4])]} at ${pad(weekly[2])}:${pad(weekly[1])}`
  const hourly = cron.match(/^0 \*\/(\d+) \* \* \*$/)
  if (hourly) return `Every ${hourly[1]} hour(s)`
  return cron
}

type Frequency = 'daily' | 'weekly' | 'hourly' | 'custom'

function buildCron(freq: Frequency, hour: string, minute: string, weekday: string, everyHours: string, custom: string): string {
  switch (freq) {
    case 'daily':
      return `${minute || '0'} ${hour || '2'} * * *`
    case 'weekly':
      return `${minute || '0'} ${hour || '2'} * * ${weekday || '0'}`
    case 'hourly':
      return `0 */${everyHours || '6'} * * *`
    case 'custom':
      return custom.trim()
  }
}

function fmtTime(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

export function BackupSchedules() {
  const { data: schedules, isLoading } = useBackupSchedules()
  const create = useCreateBackupSchedule()
  const update = useUpdateBackupSchedule()
  const remove = useDeleteBackupSchedule()
  const runNow = useRunBackupSchedule()

  const [editing, setEditing] = useState<BackupSchedule | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [freq, setFreq] = useState<Frequency>('daily')
  const [hour, setHour] = useState('02')
  const [minute, setMinute] = useState('00')
  const [weekday, setWeekday] = useState('0')
  const [everyHours, setEveryHours] = useState('6')
  const [customCron, setCustomCron] = useState('')
  const [retention, setRetention] = useState('14')
  const [enabled, setEnabled] = useState(true)

  const openAdd = () => {
    setEditing(null)
    setName('')
    setFreq('daily')
    setHour('02')
    setMinute('00')
    setWeekday('0')
    setEveryHours('6')
    setCustomCron('')
    setRetention('14')
    setEnabled(true)
    setShowForm(true)
  }

  const openEdit = (sch: BackupSchedule) => {
    setEditing(sch)
    setName(sch.name)
    setFreq('custom')
    setCustomCron(sch.cron)
    setRetention(String(sch.retention_days))
    setEnabled(sch.enabled)
    setShowForm(true)
  }

  const submit = async () => {
    if (!name.trim()) {
      toast.error('Give the schedule a name')
      return
    }
    const cron = buildCron(freq, hour, minute, weekday, everyHours, customCron)
    const retentionDays = Math.max(1, Math.min(365, parseInt(retention, 10) || 14))
    const input = { name: name.trim(), cron, enabled, retention_days: retentionDays }
    try {
      if (editing) {
        await update.mutateAsync({ id: editing.id, input })
        toast.success('Schedule updated')
      } else {
        await create.mutateAsync(input)
        toast.success('Schedule created')
      }
      setShowForm(false)
      setEditing(null)
    } catch {
      toast.error(editing ? 'Could not update schedule' : 'Could not create schedule — check the cron expression')
    }
  }

  const doRun = async (sch: BackupSchedule) => {
    try {
      await runNow.mutateAsync(sch.id)
      toast.success(`Backup started for "${sch.name}"`)
    } catch {
      toast.error('Backup failed — check the server logs')
    }
  }

  const doDelete = async (sch: BackupSchedule) => {
    if (!window.confirm(`Delete schedule "${sch.name}"? Existing dump files stay on the server.`)) return
    try {
      await remove.mutateAsync(sch.id)
      toast.success('Schedule deleted')
    } catch {
      toast.error('Could not delete schedule')
    }
  }

  return (
    <div className="space-y-3 max-w-xl">
      {showForm && (
        <div className="rounded-md border border-border-soft bg-bg-sunken/40 p-3 space-y-3">
          <div className="flex items-center justify-between">
            <div className="text-sm font-medium">{editing ? 'Edit schedule' : 'New schedule'}</div>
            <button type="button" onClick={() => setShowForm(false)} className="text-muted-foreground hover:text-foreground">
              <X className="h-4 w-4" />
            </button>
          </div>
          <div>
            <Label className="mb-1 block text-xs">Name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Nightly backup" className="h-8" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label className="mb-1 block text-xs">Frequency</Label>
              <Select value={freq} onChange={(e) => setFreq(e.target.value as Frequency)} className="h-8">
                <option value="daily">Daily</option>
                <option value="weekly">Weekly</option>
                <option value="hourly">Every N hours</option>
                <option value="custom">Custom cron</option>
              </Select>
            </div>
            {(freq === 'daily' || freq === 'weekly') && (
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <Label className="mb-1 block text-xs">Hour</Label>
                  <Input type="number" min={0} max={23} value={hour} onChange={(e) => setHour(e.target.value)} className="h-8" />
                </div>
                <div>
                  <Label className="mb-1 block text-xs">Minute</Label>
                  <Input type="number" min={0} max={59} value={minute} onChange={(e) => setMinute(e.target.value)} className="h-8" />
                </div>
              </div>
            )}
          </div>
          {freq === 'weekly' && (
            <div>
              <Label className="mb-1 block text-xs">Weekday</Label>
              <Select value={weekday} onChange={(e) => setWeekday(e.target.value)} className="h-8">
                {WEEKDAYS.map((d, i) => (
                  <option key={d} value={i}>{d}</option>
                ))}
              </Select>
            </div>
          )}
          {freq === 'hourly' && (
            <div>
              <Label className="mb-1 block text-xs">Every (hours)</Label>
              <Input type="number" min={1} max={24} value={everyHours} onChange={(e) => setEveryHours(e.target.value)} className="h-8 w-32" />
            </div>
          )}
          {freq === 'custom' && (
            <div>
              <Label className="mb-1 block text-xs">Cron (minute hour day month weekday)</Label>
              <Input value={customCron} onChange={(e) => setCustomCron(e.target.value)} placeholder="0 2 * * *" className="h-8 font-mono" />
              <div className="mt-1 text-[11px] text-muted-foreground">Times are server time (Asia/Phnom_Penh).</div>
            </div>
          )}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label className="mb-1 block text-xs">Keep dumps (days)</Label>
              <Input type="number" min={1} max={365} value={retention} onChange={(e) => setRetention(e.target.value)} className="h-8" />
            </div>
            <div>
              <Label className="mb-1 block text-xs">Enabled</Label>
              <Select value={enabled ? 'yes' : 'no'} onChange={(e) => setEnabled(e.target.value === 'yes')} className="h-8">
                <option value="yes">Yes</option>
                <option value="no">No</option>
              </Select>
            </div>
          </div>
          <div className="flex gap-2 pt-1">
            <Button size="sm" onClick={submit} disabled={create.isPending || update.isPending}>
              {editing ? 'Save' : 'Create'}
            </Button>
            <Button size="sm" variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
          </div>
        </div>
      )}

      <div className="flex items-center justify-between">
        <div className="text-xs text-muted-foreground">Schedules run on the server in Asia/Phnom_Penh time. Each dump is a full gzipped copy of the database.</div>
        <Button size="sm" variant="outline" className="gap-1.5 shrink-0" onClick={openAdd}>
          <Plus className="h-3.5 w-3.5" /> Add schedule
        </Button>
      </div>

      {isLoading ? (
        <div className="text-sm text-muted-foreground">Loading…</div>
      ) : !schedules || schedules.length === 0 ? (
        <div className="text-sm text-muted-foreground">No backup schedules yet — add one to start automatic dumps.</div>
      ) : (
        <div className="space-y-2">
          {schedules.map((sch) => (
            <div key={sch.id} className="rounded-md border border-border-soft p-3">
              <div className="flex items-center justify-between gap-2">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{sch.name}</span>
                    <span className={`text-[10px] px-1.5 py-0.5 rounded ${sch.enabled ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}`}>
                      {sch.enabled ? 'ON' : 'OFF'}
                    </span>
                  </div>
                  <div className="mt-0.5 text-xs text-muted-foreground">
                    {describeCron(sch.cron)} · keep {sch.retention_days}d
                    <span className="mx-1">·</span>
                    last: {sch.last_status === 'success' ? '✅' : sch.last_status === 'error' ? '❌' : '—'} {fmtTime(sch.last_run_at)}
                    {sch.last_status === 'error' && sch.last_error && (
                      <span className="text-destructive"> ({sch.last_error.slice(0, 80)})</span>
                    )}
                  </div>
                  {sch.enabled && (
                    <div className="mt-0.5 text-[11px] text-muted-foreground">next run: {fmtTime(sch.next_run_at)}</div>
                  )}
                </div>
                <div className="flex gap-1 shrink-0">
                  <Button size="sm" variant="outline" className="gap-1 h-7 px-2" title="Run now" onClick={() => doRun(sch)} disabled={runNow.isPending}>
                    <Play className="h-3.5 w-3.5" />
                  </Button>
                  {sch.latest_file && (
                    <Button size="sm" variant="outline" className="gap-1 h-7 px-2" title="Download latest dump"
                      onClick={() => downloadFile(`/backup-schedules/${sch.id}/latest`, sch.latest_file).catch(() => toast.error('No backup to download yet'))}>
                      <Download className="h-3.5 w-3.5" />
                    </Button>
                  )}
                  <Button size="sm" variant="outline" className="gap-1 h-7 px-2" title="Edit" onClick={() => openEdit(sch)}>
                    <Pencil className="h-3.5 w-3.5" />
                  </Button>
                  <Button size="sm" variant="outline" className="gap-1 h-7 px-2" title="Delete" onClick={() => doDelete(sch)}>
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
