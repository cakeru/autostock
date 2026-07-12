import { useState } from 'react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { Select } from '@/components/ui/select'
import { formatUSD } from '@/utils/currency'
import { useSalesReport } from '@/hooks/useAnalytics'
import { StatCard, Panel, RangeBar, monthStart, today, pct, DataTable, CHART } from './shared'

export function SalesAnalytics() {
  const [from, setFrom] = useState(monthStart())
  const [to, setTo] = useState(today())
  const [gran, setGran] = useState('day')
  const { data, isLoading } = useSalesReport(from, to, gran)

  return (
    <div className="max-w-5xl space-y-5">
      <PageHeader title="Sales analytics" subtitle="Revenue trends, average ticket, and payment mix" />

      <RangeBar from={from} to={to} setFrom={setFrom} setTo={setTo}
        right={
          <Select value={gran} onChange={(e) => setGran(e.target.value)} className="w-28">
            <option value="day">Daily</option>
            <option value="week">Weekly</option>
            <option value="month">Monthly</option>
          </Select>
        }
      />

      {isLoading || !data ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {[0, 1, 2, 3].map((i) => <div key={i} className="rounded-lg bg-card p-5 shadow-sm"><Skeleton className="h-3 w-20" /><Skeleton className="mt-2 h-8 w-24" /></div>)}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <StatCard label="Revenue" value={formatUSD(data.revenue_usd)}
              sub={<span className={data.revenue_change_pct >= 0 ? 'text-emerald-600' : 'text-destructive'}>{pct(data.revenue_change_pct)} vs prev {formatUSD(data.prev_revenue_usd)}</span>}
              accent={data.revenue_change_pct >= 0 ? 'up' : 'down'} />
            <StatCard label="Invoices" value={data.invoice_count} />
            <StatCard label="Avg ticket" value={formatUSD(data.avg_ticket)} />
            <StatCard label="Tax collected" value={formatUSD(data.tax_collected)} />
          </div>

          <Panel title="Revenue over time">
            {data.series.length === 0 ? (
              <p className="py-10 text-center text-sm text-muted-foreground">No sales in this range</p>
            ) : (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={data.series} margin={{ top: 8, right: 8, left: 8, bottom: 8 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke={CHART.grid} vertical={false} />
                  <XAxis dataKey="period" tick={{ fontSize: 11, fill: CHART.axis }} tickLine={false} axisLine={false} />
                  <YAxis tick={{ fontSize: 11, fill: CHART.axis }} tickLine={false} axisLine={false} width={48}
                    tickFormatter={(v) => `$${v}`} />
                  <Tooltip formatter={(v: number) => formatUSD(v)} labelStyle={{ fontSize: 12 }} contentStyle={{ fontSize: 12, borderRadius: 8 }} />
                  <Bar dataKey="revenue_usd" name="Revenue" fill={CHART.navy} radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </Panel>

          <Panel title="Payment methods">
            <DataTable
              rows={data.payment_methods}
              rowKey={(m) => m.method}
              empty="No payments in this range"
              columns={[
                { header: 'Method', primary: true, cell: (m) => <span className="capitalize">{m.method}</span> },
                { header: 'Count', align: 'right', cell: (m) => m.count },
                { header: 'Total', align: 'right', cell: (m) => formatUSD(m.total) },
              ]}
            />
          </Panel>
        </>
      )}
    </div>
  )
}
