import { useState } from 'react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatUSD } from '@/utils/currency'
import { useTechnicians } from '@/hooks/useAnalytics'
import { Panel, RangeBar, DataTable, monthStart, today } from './shared'

export function TechniciansAnalytics() {
  const [from, setFrom] = useState(monthStart())
  const [to, setTo] = useState(today())
  const { data, isLoading } = useTechnicians(from, to)

  return (
    <div className="max-w-5xl space-y-5">
      <PageHeader title="Technician productivity" subtitle="Completed jobs, revenue produced, hours logged, and what they cost to pay" />
      <RangeBar from={from} to={to} setFrom={setFrom} setTo={setTo} />

      <Panel title="By technician">
        {isLoading || !data ? (
          <Skeleton className="h-40 w-full" />
        ) : (
          <DataTable
            rows={data.technicians}
            rowKey={(t) => t.employee_id}
            empty="No completed jobs assigned to a technician in this range"
            columns={[
              { header: 'Technician', primary: true, cell: (t) => t.name },
              { header: 'Jobs done', align: 'right', cell: (t) => t.jobs_completed },
              { header: 'Revenue', align: 'right', cell: (t) => formatUSD(t.revenue) },
              { header: 'Hours', align: 'right', cell: (t) => t.hours || '—' },
              { header: 'Pay cost', align: 'right', cell: (t) => formatUSD(t.payroll_cost) },
            ]}
          />
        )}
      </Panel>
    </div>
  )
}
