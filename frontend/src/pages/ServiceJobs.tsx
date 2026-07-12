import { JobBoard } from '@/components/servicejob/JobBoard'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { Trash2 } from 'lucide-react'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { PageHeader } from '@/components/layout/PageHeader'
import { useState, useMemo } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useServiceJobs, useCreateServiceJob, useDeleteServiceJob } from '@/hooks/useServiceJobs'
import { useCustomers, useCustomerVehicles } from '@/hooks/useCustomers'
import { StatusBadge } from '@/components/servicejob/StatusBadge'
import { PriorityBadge } from '@/components/servicejob/PriorityBadge'
import { formatDate } from '@/utils/date'
import { TableCard, TableFooter, Th, ActionsTh } from '@/components/ui/table'

function ViewToggle({ jobBoardView, setJobBoardView }: {
  jobBoardView: boolean
  setJobBoardView: (v: boolean) => void
}) {
  return (
    <div className="hidden md:flex border rounded-md overflow-hidden">
      <button
        onClick={() => setJobBoardView(true)}
        className={`px-3 py-1.5 text-xs font-medium transition-colors ${jobBoardView ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}
      >Board</button>
      <button
        onClick={() => setJobBoardView(false)}
        className={`px-3 py-1.5 text-xs font-medium transition-colors ${!jobBoardView ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}
      >Table</button>
    </div>
  )
}
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'

function AgendaView({ jobs, onOpen }: { jobs: any[]; onOpen: (id: number) => void }) {
  if (jobs.length === 0) {
    return <p className="py-12 text-center text-sm text-muted-foreground">No upcoming appointments. Schedule a job to see it here.</p>
  }
  const groups: Record<string, any[]> = {}
  jobs.forEach((j) => {
    const key = new Date(j.scheduled_at).toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' })
    ;(groups[key] ||= []).push(j)
  })
  return (
    <div className="space-y-4">
      {Object.entries(groups).map(([day, list]) => (
        <div key={day}>
          <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{day}</p>
          <div className="space-y-2">
            {list.map((j) => (
              <div key={j.id} onClick={() => onOpen(j.id)} className="flex cursor-pointer items-center justify-between gap-3 rounded-lg bg-card p-3 shadow-sm transition-colors hover:bg-muted/40">
                <div className="flex min-w-0 items-center gap-3">
                  <span className="w-16 flex-shrink-0 text-sm font-semibold tabular-nums">{new Date(j.scheduled_at).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })}</span>
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{j.customer_name || 'Walk-in'}{j.plate_number ? ` · ${j.plate_number}` : ''}</p>
                    <p className="truncate text-xs text-muted-foreground">{j.description || j.job_number}</p>
                  </div>
                </div>
                <StatusBadge status={j.status} />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

export function ServiceJobs() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const preselectedCustomer = searchParams.get('customer') || ''
  const [page, setPage] = useState(1)
  const [jobBoardView, setJobBoardView] = useState(window.innerWidth >= 1024)
  const [statusFilter, setStatusFilter] = useState('')
  const [showForm, setShowForm] = useState(searchParams.get('new') === '1')
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [description, setDescription] = useState('')
  const [notes, setNotes] = useState('')
  const [formCustomerId, setFormCustomerId] = useState(preselectedCustomer)
  const [formVehicleId, setFormVehicleId] = useState('')
  const [scheduledAt, setScheduledAt] = useState('')
  const [agenda, setAgenda] = useState(false)

  const scheduledFrom = useMemo(() => new Date(new Date().setHours(0, 0, 0, 0)).toISOString(), [])
  const { data, isLoading } = useServiceJobs({ status: statusFilter || undefined, page: jobBoardView ? 1 : page, per_page: jobBoardView ? 100 : undefined })
  const { data: agendaData } = useServiceJobs({ scheduled_from: scheduledFrom, per_page: 100 })
  const upcoming = agendaData?.data || []
  const { data: customersData } = useCustomers({ per_page: 100 })
  const { data: vehicles } = useCustomerVehicles(formCustomerId ? parseInt(formCustomerId) : 0)
  const createMutation = useCreateServiceJob()
  const deleteMutation = useDeleteServiceJob()

  const jobs = data?.data || []
  const meta = data?.meta
  const customers = customersData?.data || []

  const closeForm = () => {
    setShowForm(false)
    setDescription('')
    setNotes('')
    setFormCustomerId('')
    setFormVehicleId('')
    setScheduledAt('')
    if (searchParams.get('new') || searchParams.get('customer')) setSearchParams({}, { replace: true })
  }

  const handleCreate = () => {
    if (!description) return
    createMutation.mutate({
      description,
      notes: notes || undefined,
      customer_id: formCustomerId ? parseInt(formCustomerId) : undefined,
      vehicle_id: formVehicleId ? parseInt(formVehicleId) : undefined,
      scheduled_at: scheduledAt || undefined,
    }, { onSuccess: closeForm })
  }

  const handleDelete = () => {
    if (deletingId) deleteMutation.mutate(deletingId, { onSuccess: () => setDeletingId(null) })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Service Jobs"
        actions={
          <Button onClick={() => setShowForm(true)} size="sm">New Job</Button>
        }
      />

      <div className="flex flex-wrap items-center gap-1">
        <button onClick={() => setAgenda(true)}
          className={`rounded-full px-3 py-1.5 text-sm font-medium transition-colors ${agenda ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}>
          Upcoming{upcoming.length ? ` (${upcoming.length})` : ''}
        </button>
        <button onClick={() => setAgenda(false)}
          className={`rounded-full px-3 py-1.5 text-sm font-medium transition-colors ${!agenda ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}>
          All jobs
        </button>
      </div>

      {agenda ? (
        <AgendaView jobs={upcoming} onOpen={(id) => navigate(`/service-jobs/${id}`)} />
      ) : (
      <>
      {jobBoardView && (
        <div className="flex items-center gap-2.5">
          <Select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1); setJobBoardView(false) }} className="w-40">
            <option value="">All statuses</option>
            <option value="pending">Pending</option>
            <option value="in_progress">In Progress</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </Select>
          <ViewToggle jobBoardView={jobBoardView} setJobBoardView={setJobBoardView} />
        </div>
      )}

      {isLoading ? (
        jobBoardView ? (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="bg-muted/30 border rounded-md p-3 min-h-[200px] animate-pulse">
                <div className="h-4 bg-muted rounded w-20 mb-3" />
                <div className="space-y-2">
                  <div className="h-16 bg-muted rounded" />
                  <div className="h-16 bg-muted rounded" />
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="bg-card rounded-lg shadow-sm p-4">
            <TableSkeleton rows={5} cols={7} />
          </div>
        )
      ) : jobBoardView ? (
        <JobBoard jobs={jobs} />
      ) : (
        <>
          <div className="hidden md:block">
            <TableCard
              toolbar={
                <>
                  <Select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }} className="w-40">
                    <option value="">All statuses</option>
                    <option value="pending">Pending</option>
                    <option value="in_progress">In Progress</option>
                    <option value="completed">Completed</option>
                    <option value="cancelled">Cancelled</option>
                  </Select>
                  <div className="ml-auto">
                    <ViewToggle jobBoardView={jobBoardView} setJobBoardView={setJobBoardView} />
                  </div>
                </>
              }
              footer={meta && (
                <TableFooter total={meta.total} page={meta.page} totalPages={meta.total_pages} onPage={setPage} noun="jobs" />
              )}
            >
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-card sticky top-0 z-10">
                    <Th>Customer</Th>
                    <Th>Vehicle</Th>
                    <Th>Description</Th>
                    <Th>Status</Th>
                    <Th>Priority</Th>
                    <Th className="text-right">Created</Th>
                    <ActionsTh />
                  </tr>
                </thead>
                <tbody>
                  {jobs.length === 0 ? (
                    <tr><td colSpan={7} className="px-4 py-10 text-center text-muted-foreground">No jobs found</td></tr>
                  ) : (
                    jobs.map((j) => (
                      <tr key={j.id} className="group border-b last:border-0 hover:bg-muted/50 transition-colors duration-100 cursor-pointer" onClick={() => navigate(`/service-jobs/${j.id}`)}>
                        <td className="px-4 py-2.5">
                          <p className={`font-medium ${j.customer_name ? '' : 'text-muted-foreground'}`}>{j.customer_name || 'Walk-in'}</p>
                          <p className="font-mono text-xs text-muted-foreground">{j.job_number}</p>
                        </td>
                        <td className="px-4 py-2.5">{j.plate_number || <span className="text-muted-foreground">—</span>}</td>
                        <td className="px-4 py-2.5 text-muted-foreground max-w-[200px] truncate">{j.description}</td>
                        <td className="px-4 py-2.5"><StatusBadge status={j.status} /></td>
                        <td className="px-4 py-2.5"><PriorityBadge priority={j.priority} /></td>
                        <td className="px-4 py-2.5 text-right text-muted-foreground text-xs whitespace-nowrap">{formatDate(j.created_at)}</td>
                        <td className="px-4 py-2.5 text-right">
                          <div className="text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
                            <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={(e) => { e.stopPropagation(); setDeletingId(j.id) }} aria-label="Delete" title="Delete job">
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </TableCard>
          </div>

          <div className="md:hidden space-y-3">
            <Select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }} className="w-40">
              <option value="">All statuses</option>
              <option value="pending">Pending</option>
              <option value="in_progress">In Progress</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </Select>
            {jobs.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-8">No jobs found</p>
            ) : (
              jobs.map((j) => (
                <div key={j.id} className="bg-card rounded-lg p-5 shadow-sm cursor-pointer" onClick={() => navigate(`/service-jobs/${j.id}`)}>
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-xs text-muted-foreground">{j.job_number}</span>
                    <StatusBadge status={j.status} />
                  </div>
                  <p className="text-sm font-medium mt-1">{j.customer_name || 'Walk-in'}</p>
                  <p className="text-xs text-muted-foreground truncate">{j.description}</p>
                  <div className="flex items-center justify-between mt-2">
                    <PriorityBadge priority={j.priority} />
                    <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); setDeletingId(j.id) }}>Delete</Button>
                  </div>
                </div>
              ))
            )}
          </div>

          {meta && meta.total_pages > 1 && (
            <div className="md:hidden flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Page {meta.page} of {meta.total_pages}</span>
              <div className="flex gap-1">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
                <Button variant="outline" size="sm" disabled={page >= meta.total_pages} onClick={() => setPage(page + 1)}>Next</Button>
              </div>
            </div>
          )}
        </>
      )}
      </>
      )}

      {showForm && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-10 bg-black/50">
          <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-lg mx-4">
            <h2 className="text-sm font-semibold mb-3">Create Service Job</h2>
            <div className="space-y-3">
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Customer</label>
                <Select value={formCustomerId} onChange={(e) => { setFormCustomerId(e.target.value); setFormVehicleId('') }}>
                  <option value="">Walk-in (no customer)</option>
                  {customers.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}{c.phone ? ` — ${c.phone}` : ''}</option>
                  ))}
                </Select>
              </div>
              {formCustomerId && (
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Vehicle</label>
                  {vehicles && vehicles.length > 0 ? (
                    <Select value={formVehicleId} onChange={(e) => setFormVehicleId(e.target.value)}>
                      <option value="">No vehicle</option>
                      {vehicles.map((v) => (
                        <option key={v.id} value={v.id}>
                          {v.plate_number}{v.make || v.model ? ` — ${[v.make, v.model].filter(Boolean).join(' ')}` : ''}
                        </option>
                      ))}
                    </Select>
                  ) : (
                    <p className="text-xs text-muted-foreground border rounded px-3 py-2">
                      No vehicles registered for this customer. You can add one on their customer page.
                    </p>
                  )}
                </div>
              )}
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Description *</label>
                <textarea value={description} onChange={(e) => setDescription(e.target.value)}
                  className="w-full rounded border border-input bg-background px-3 py-2 text-sm min-h-[60px]" rows={3} required
                  placeholder="e.g. Replace 4 tires, front brake check" />
              </div>
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Schedule (optional)</label>
                <input type="datetime-local" value={scheduledAt} onChange={(e) => setScheduledAt(e.target.value)}
                  className="w-full rounded border border-input bg-background px-3 py-2 text-sm" />
                <p className="text-xs text-muted-foreground">Leave blank for a walk-in job.</p>
              </div>
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Notes (optional)</label>
                <input value={notes} onChange={(e) => setNotes(e.target.value)}
                  className="w-full rounded border border-input bg-background px-3 py-2 text-sm"
                  placeholder="Anything else worth noting" />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <Button variant="ghost" onClick={closeForm}>Cancel</Button>
                <Button onClick={handleCreate} disabled={!description || createMutation.isPending}>
                  {createMutation.isPending ? 'Creating...' : 'Create Job'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deletingId !== null}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDelete}
        title="Delete Service Job"
        message={`Delete ${jobs.find(j => j.id === deletingId)?.job_number || 'this job'}? This cannot be undone.`}
        destructive
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
