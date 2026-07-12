import { cn } from '@/lib/utils'

const statusStyles: Record<string, string> = {
  paid: 'bg-primary/10 text-primary',
  issued: 'bg-blue-100 text-blue-700',
  unpaid: 'bg-amber-100 text-amber-700',
  partial: 'bg-amber-100 text-amber-700',
  voided: 'bg-destructive/10 text-destructive',
  refunded: 'bg-destructive/10 text-destructive',
}

const statusLabels: Record<string, string> = {
  paid: 'Paid',
  issued: 'Issued',
  unpaid: 'Unpaid',
  partial: 'Partial',
  voided: 'Voided',
  refunded: 'Refunded',
}

export function InvoiceStatusBadge({ status, payment_status }: { status?: string; payment_status?: string }) {
  const key = payment_status || status || ''
  return (
    <span className={cn('rounded-full px-2 py-0.5 text-xs font-medium', statusStyles[key] || 'bg-muted text-muted-foreground')}>
      {statusLabels[key] || key}
    </span>
  )
}
