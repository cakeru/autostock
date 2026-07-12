import { cn } from '@/lib/utils'

const statusColors: Record<string, string> = {
  pending: 'bg-muted text-muted-foreground',
  in_progress: 'bg-primary/10 text-primary',
  completed: 'bg-primary/10 text-primary',
  cancelled: 'bg-destructive/10 text-destructive',
}

const statusLabels: Record<string, string> = {
  pending: 'Pending',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
}

export function StatusBadge({ status }: { status: string }) {
  return (
    <span className={cn('rounded-full px-2 py-0.5 text-xs font-medium', statusColors[status] || 'bg-muted text-muted-foreground')}>
      {statusLabels[status] || status}
    </span>
  )
}
