import { useState } from 'react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatUSD } from '@/utils/currency'
import { useInventoryReport } from '@/hooks/useAnalytics'
import { StatCard, Panel, DaysSelect, DataTable } from './shared'

export function InventoryAnalytics() {
  const [days, setDays] = useState(90)
  const { data, isLoading } = useInventoryReport(days)

  return (
    <div className="max-w-5xl space-y-5">
      <PageHeader title="Inventory intelligence" subtitle="Valuation, movers, dead stock, and reorder suggestions" />

      <div className="flex items-center justify-between">
        <DaysSelect days={days} setDays={setDays} />
        <span className="text-xs text-muted-foreground">Sales window</span>
      </div>

      {isLoading || !data ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {[0, 1, 2, 3].map((i) => <div key={i} className="rounded-lg bg-card p-5 shadow-sm"><Skeleton className="h-3 w-16" /><Skeleton className="mt-2 h-8 w-24" /></div>)}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <StatCard label="Stock value (cost)" value={formatUSD(data.valuation.cost_value)} sub={`${data.valuation.units_on_hand} units · ${data.valuation.sku_count} SKUs`} />
            <StatCard label="Stock value (retail)" value={formatUSD(data.valuation.retail_value)} />
            <StatCard label="Potential profit" value={formatUSD(data.valuation.potential_profit)} accent="up" />
            <StatCard label="Turnover" value={`${data.turnover_ratio}×`} sub={`over ${data.window_days} days`} />
          </div>

          <Panel title="Reorder suggestions">
            <DataTable
              rows={data.reorder}
              rowKey={(r) => r.product_id}
              empty="Nothing needs reordering"
              columns={[
                { header: 'Product', primary: true, cell: (r) => r.name },
                { header: 'In stock', align: 'right', cell: (r) => r.stock_qty },
                { header: 'Min', align: 'right', cell: (r) => <span className="text-muted-foreground">{r.min_stock}</span> },
                { header: 'Sold/day', align: 'right', cell: (r) => r.daily_rate },
                { header: 'Days left', align: 'right', cell: (r) => (
                  <span className={r.days_left != null && r.days_left <= 7 ? 'font-medium text-destructive' : r.days_left != null && r.days_left <= 21 ? 'font-medium text-amber-600' : ''}>
                    {r.days_left != null ? `${r.days_left}d` : '—'}
                  </span>
                ) },
              ]}
            />
          </Panel>

          <div className="grid gap-5 lg:grid-cols-2">
            <Panel title={`Best sellers (${data.window_days}d)`}>
              <DataTable
                rows={data.top_sellers}
                rowKey={(p) => p.product_id}
                empty="No sales yet"
                columns={[
                  { header: 'Product', primary: true, cell: (p) => p.name },
                  { header: 'Qty', align: 'right', cell: (p) => p.qty_sold },
                  { header: 'Revenue', align: 'right', cell: (p) => formatUSD(p.revenue_usd) },
                  { header: 'Profit', align: 'right', cell: (p) => <span className="text-emerald-600">{formatUSD(p.profit_usd)}</span> },
                ]}
              />
            </Panel>

            <Panel title="Dead stock">
              <DataTable
                rows={data.dead_stock}
                rowKey={(d) => d.product_id}
                empty="No dead stock 🎉"
                columns={[
                  { header: 'Product', primary: true, cell: (d) => d.name },
                  { header: 'Stock', align: 'right', cell: (d) => d.stock_qty },
                  { header: 'Tied-up', align: 'right', cell: (d) => formatUSD(d.cost_value) },
                  { header: 'Last sold', align: 'right', cell: (d) => <span className="text-muted-foreground">{d.last_sold ? d.last_sold.slice(0, 10) : 'never'}</span> },
                ]}
              />
            </Panel>
          </div>
        </>
      )}
    </div>
  )
}
