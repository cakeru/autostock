export interface TelegramChannel {
  id: string
  label: string
  bot_token: string
  chat_id: string
}

export type TelegramTopic =
  | 'jobs'
  | 'sales'
  | 'alerts'
  | 'daily_digest'
  | 'tomorrow_appts'
  | 'weekly_ap'
  | 'monthly_report'
  | 'monthly_backup'
  | 'due_for_service'
  | 'documents'

export const TELEGRAM_TOPICS: { value: TelegramTopic; label: string; description: string; scheduled?: boolean }[] = [
  { value: 'jobs', label: 'Jobs', description: 'Posted on job creation, edited in place as status changes' },
  { value: 'sales', label: 'Sales', description: 'Every sale issued or voided' },
  { value: 'alerts', label: 'Alerts', description: 'Cash drawer discrepancies, low stock, stocktake shrinkage' },
  { value: 'daily_digest', label: 'Daily digest', description: 'End-of-day sales/jobs/cash rollup', scheduled: true },
  { value: 'tomorrow_appts', label: "Tomorrow's appointments", description: 'Morning digest of scheduled jobs', scheduled: true },
  { value: 'weekly_ap', label: 'Weekly payables', description: 'Outstanding supplier balances', scheduled: true },
  { value: 'monthly_report', label: 'Monthly P&L report', description: 'Last month’s profit & loss summary', scheduled: true },
  { value: 'monthly_backup', label: 'Monthly database backup', description: 'A gzipped SQL dump sent as a file', scheduled: true },
  { value: 'due_for_service', label: 'Due for service', description: 'Weekly (Monday) list of customers overdue or coming up for oil/tires', scheduled: true },
  { value: 'documents', label: 'Documents', description: 'Where "Send to Telegram" delivers invoices & vehicle reports for you to forward to customers' },
]
