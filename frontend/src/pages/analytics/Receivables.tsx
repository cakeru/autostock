import { Link } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatUSD } from '@/utils/currency'
import { useReceivables } from '@/hooks/useAnalytics'
import { StatCard, Panel, DataTable } from './shared'

export function ReceivablesAnalytics() {
  const { data, isLoading } = useReceivables()

  return (
    <div className="max-w-5xl space-y-5">
      <PageHeader title="Receivables" subtitle="Outstanding balances aged by how overdue they are" />

      {isLoading || !data ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {[0, 1, 2, 3].map((i) => <div key={i} className="rounded-lg bg-card p-5 shadow-sm"><Skeleton className="h-3 w-16" /><Skeleton className="mt-2 h-8 w-24" /></div>)}
        </div>
      ) : (
        <>
          <StatCard label="Total outstanding" value={formatUSD(data.total_outstanding)} accent={data.total_outstanding > 0 ? 'warn' : undefined} />

          <Panel title="Aging buckets">
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
              {data.buckets.map((b) => (
                <div key={b.label} className="rounded-md border p-4">
                  <p className="text-xs text-muted-foreground">{b.label} days</p>
                  <p className="mt-1 text-lg font-bold tabular-nums">{formatUSD(b.outstanding_usd)}</p>
                  <p className="text-xs text-muted-foreground">{b.count} invoice{b.count === 1 ? '' : 's'}</p>
                </div>
              ))}
            </div>
          </Panel>

          <Panel title="By customer">
            <DataTable
              rows={data.customers}
              rowKey={(c, i) => c.customer_id ?? `w${i}`}
              empty="Nothing outstanding — all settled"
              columns={[
                { header: 'Customer', primary: true, cell: (c) => c.customer_id ? <Link to={`/customers/${c.customer_id}`} className="text-primary hover:underline">{c.customer_name}</Link> : c.customer_name },
                { header: 'Invoices', align: 'right', cell: (c) => c.invoice_count },
                { header: 'Oldest', align: 'right', cell: (c) => `${c.oldest_days}d` },
                { header: 'Outstanding', align: 'right', cell: (c) => formatUSD(c.outstanding_usd) },
              ]}
            />
          </Panel>
        </>
      )}
    </div>
  )
}
