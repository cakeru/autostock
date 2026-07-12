import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '@/components/servicejob/StatusBadge'
import { PriorityBadge } from '@/components/servicejob/PriorityBadge'
import type { ServiceJob } from '@/types/servicejob'

function timeAgo(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = Math.max(0, now - then)
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

interface Props {
  jobs: ServiceJob[]
}

export function JobBoard({ jobs }: Props) {
  const navigate = useNavigate()

  const columns = [
    { status: 'pending', label: 'Pending', color: 'border-l-yellow-400' },
    { status: 'in_progress', label: 'In Progress', color: 'border-l-blue-400' },
    { status: 'completed', label: 'Completed', color: 'border-l-green-400' },
  ]

  const grouped: Record<string, ServiceJob[]> = {}
  for (const col of columns) grouped[col.status] = []
  for (const j of jobs) {
    if (grouped[j.status]) grouped[j.status].push(j)
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {columns.map((col) => (
        <div key={col.status} className="flex max-h-[75vh] min-h-[200px] flex-col rounded-md border bg-muted/30 p-3">
          <div className={`border-l-4 ${col.color} flex-shrink-0 pl-2 mb-3`}>
            <p className="text-sm font-semibold">{col.label}</p>
            <p className="text-xs text-muted-foreground">{grouped[col.status].length} job{grouped[col.status].length !== 1 ? 's' : ''}</p>
          </div>
          <div className="min-h-0 flex-1 space-y-2 overflow-y-auto pr-1">
            {grouped[col.status].length === 0 ? (
              <p className="text-xs text-muted-foreground text-center py-4">No jobs</p>
            ) : (
              grouped[col.status].map((j) => (
                <div
                  key={j.id}
                  className="bg-card rounded-lg shadow-sm p-3 cursor-pointer hover:shadow-sm transition-shadow"
                  onClick={() => navigate(`/service-jobs/${j.id}`)}
                >
                  <div className="flex items-start justify-between gap-1 mb-1">
                    <span className="font-mono text-xs font-medium">{j.job_number}</span>
                    <PriorityBadge priority={j.priority} />
                  </div>
                  <p className="text-xs truncate">{j.customer_name || 'Walk-in'}</p>
                  {j.plate_number && <p className="text-[10px] text-muted-foreground">{j.plate_number}</p>}
                  <p className="text-[10px] text-muted-foreground mt-1">{timeAgo(j.created_at)}</p>
                </div>
              ))
            )}
          </div>
        </div>
      ))}
    </div>
  )
}
