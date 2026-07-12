import { useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { Select } from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { formatDateTime } from '@/utils/date'
import { useAuditLog } from '@/hooks/useAudit'
import { useUsers } from '@/hooks/useUsers'
import { Panel, RangeBar, DataTable } from './shared'

function monthStart() {
  const d = new Date()
  return new Date(d.getFullYear(), d.getMonth(), 1).toISOString().slice(0, 10)
}
const today = () => new Date().toISOString().slice(0, 10)

// Known actions today; humanized labels drive the filter + display.
const ACTIONS: Record<string, string> = {
  invoice_created: 'Invoice created',
  payment_recorded: 'Payment recorded',
  invoice_voided: 'Invoice voided',
}
const humanize = (a: string) => ACTIONS[a] || a.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase())

function entityHref(type: string, id?: number): string | null {
  if (!id) return null
  if (type === 'invoice') return `/invoices/${id}`
  if (type === 'service_job') return `/service-jobs/${id}`
  if (type === 'customer') return `/customers/${id}`
  return null
}

export function AuditLog() {
  const [from, setFrom] = useState(monthStart())
  const [to, setTo] = useState(today())
  const [action, setAction] = useState('')
  const [userId, setUserId] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useAuditLog({
    from, to, page, per_page: 25,
    action: action || undefined,
    user_id: userId ? parseInt(userId) : undefined,
  })
  const { data: users } = useUsers()
  const meta = data?.meta

  const setFilter = (fn: () => void) => { fn(); setPage(1) }

  return (
    <div className="max-w-5xl space-y-5">
      <PageHeader title="Audit log" subtitle="Who did what — voids, payments, and changes" />

      <RangeBar
        from={from} to={to}
        setFrom={(v) => setFilter(() => setFrom(v))}
        setTo={(v) => setFilter(() => setTo(v))}
        right={
          <>
            <Select value={action} onChange={(e) => setFilter(() => setAction(e.target.value))} className="w-40">
              <option value="">All actions</option>
              {Object.entries(ACTIONS).map(([k, label]) => <option key={k} value={k}>{label}</option>)}
            </Select>
            <Select value={userId} onChange={(e) => setFilter(() => setUserId(e.target.value))} className="w-36">
              <option value="">All users</option>
              {(users || []).map((u: any) => <option key={u.id} value={u.id}>{u.full_name || u.username}</option>)}
            </Select>
          </>
        }
      />

      <Panel title={meta ? `${meta.total} event${meta.total === 1 ? '' : 's'}` : 'Events'}>
        {isLoading || !data ? (
          <Skeleton className="h-40 w-full" />
        ) : (
          <>
            <DataTable
              rows={data.data}
              rowKey={(a) => a.id}
              empty="No activity in this range"
              columns={[
                { header: 'When', cell: (a) => <span className="tabular-nums text-muted-foreground">{formatDateTime(a.created_at)}</span> },
                { header: 'User', cell: (a) => a.user_name || '—' },
                { header: 'Action', primary: true, cell: (a) => (
                  <span className={a.action === 'invoice_voided' ? 'text-destructive' : ''}>{humanize(a.action)}</span>
                ) },
                { header: 'Item', cell: (a) => {
                  const href = entityHref(a.entity_type, a.entity_id)
                  const label = `${a.entity_type.replace(/_/g, ' ')}${a.entity_id ? ` #${a.entity_id}` : ''}`
                  return href ? <Link to={href} className="text-primary hover:underline">{label}</Link> : <span className="text-muted-foreground">{label}</span>
                } },
              ]}
            />
            {meta && meta.total_pages > 1 && (
              <div className="mt-3 flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Page {meta.page} of {meta.total_pages}</span>
                <div className="flex gap-1">
                  <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
                  <Button variant="outline" size="sm" disabled={page >= meta.total_pages} onClick={() => setPage(page + 1)}>Next</Button>
                </div>
              </div>
            )}
          </>
        )}
      </Panel>
    </div>
  )
}
