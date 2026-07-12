import { cn } from '@/lib/utils'

const priorityStyles: Record<string, string> = {
  low: 'border border-muted-foreground/30 text-muted-foreground',
  normal: 'bg-primary/5 text-primary border border-primary/20',
  high: 'bg-accent/10 text-accent border border-accent/30',
  urgent: 'bg-destructive/10 text-destructive border border-destructive/30',
}

const priorityLabels: Record<string, string> = {
  low: 'Low',
  normal: 'Normal',
  high: 'High',
  urgent: 'Urgent',
}

export function PriorityBadge({ priority }: { priority: string }) {
  const style = priorityStyles[priority] || priorityStyles.normal
  return (
    <span className={cn('rounded-full px-2 py-0.5 text-xs font-medium', style)}>
      {priorityLabels[priority] || priority}
    </span>
  )
}
