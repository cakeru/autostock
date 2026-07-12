import { useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatUSD } from '@/utils/currency'
import { useCustomersReport } from '@/hooks/useAnalytics'
import { StatCard, Panel, DaysSelect, DataTable } from './shared'

export function CustomersAnalytics() {
  const [days, setDays] = useState(90)
  const { data, isLoading } = useCustomersReport(days)

  return (
    <div className="max-w-5xl space-y-5">
      <PageHeader title="Customer analytics" subtitle="Value, loyalty, and who's at risk of churning" />

      <div className="flex items-center justify-between">
        <DaysSelect days={days} setDays={setDays} />
        <span className="text-xs text-muted-foreground">Active / churn window</span>
      </div>

      {isLoading || !data ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {[0, 1, 2, 3].map((i) => <div key={i} className="rounded-lg bg-card p-5 shadow-sm"><Skeleton className="h-3 w-16" /><Skeleton className="mt-2 h-8 w-24" /></div>)}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <StatCard label="Customers with sales" value={data.total_customers} />
            <StatCard label="Repeat customers" value={data.repeat_customers} sub={data.total_customers > 0 ? `${Math.round(data.repeat_customers / data.total_customers * 100)}% of base` : undefined} />
            <StatCard label="New this window" value={data.new_customers} accent="up" />
            <StatCard label="Avg spend / visit" value={formatUSD(data.avg_spend_per_visit)} />
          </div>

          <div className="grid gap-5 lg:grid-cols-2">
            <Panel title="Top customers by lifetime spend">
              <DataTable
                rows={data.top_customers}
                rowKey={(c) => c.customer_id}
                empty="No customer sales yet"
                columns={[
                  { header: 'Customer', primary: true, cell: (c) => <Link to={`/customers/${c.customer_id}`} className="text-primary hover:underline">{c.name}</Link> },
                  { header: 'Visits', align: 'right', cell: (c) => c.invoice_count },
                  { header: 'Spent', align: 'right', cell: (c) => formatUSD(c.total_spent) },
                ]}
              />
            </Panel>

            <Panel title={`Churn risk — no visit in ${data.window_days}d (${data.churn_risk_count})`}>
              <DataTable
                rows={data.churn_risk}
                rowKey={(c) => c.customer_id}
                empty="No one at risk 🎉"
                columns={[
                  { header: 'Customer', primary: true, cell: (c) => <Link to={`/customers/${c.customer_id}`} className="text-primary hover:underline">{c.name}</Link> },
                  { header: 'Last visit', align: 'right', cell: (c) => <span className="text-amber-600">{c.days_since_visit}d ago</span> },
                  { header: 'Spent', align: 'right', cell: (c) => formatUSD(c.total_spent) },
                ]}
              />
            </Panel>
          </div>
        </>
      )}
    </div>
  )
}
