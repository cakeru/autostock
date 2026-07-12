import { useNavigate } from 'react-router-dom'
import { BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid, ResponsiveContainer } from 'recharts'
import { Printer, ChevronRight } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'
import { useDashboardSummary, useDailyRevenue, useDayClose } from '@/hooks/useDashboard'
import { useSalesReport, usePnL, useInventoryReport } from '@/hooks/useAnalytics'
import { useLowStockProducts } from '@/hooks/useProducts'
import { CashDrawer } from '@/components/cash/CashDrawer'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatUSD } from '@/utils/currency'

const statusStyles: Record<string, string> = {
  paid: 'bg-green-100 text-green-800',
  issued: 'bg-blue-100 text-blue-800',
  voided: 'bg-red-100 text-red-800',
  unpaid: 'bg-yellow-100 text-yellow-800',
  draft: 'bg-gray-100 text-gray-800',
  completed: 'bg-green-100 text-green-800',
  in_progress: 'bg-blue-100 text-blue-800',
  pending: 'bg-yellow-100 text-yellow-800',
  cancelled: 'bg-red-100 text-red-800',
}

function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-medium whitespace-nowrap ${statusStyles[status] || 'bg-gray-100 text-gray-800'}`}>
      {status.replace('_', ' ')}
    </span>
  )
}

function Stat({ label, value, context, valueClass = '', divider = '', onClick }: {
  label: string
  value: string
  context?: { text: string; tone: 'up' | 'down' | 'muted' | 'alert' }
  valueClass?: string
  divider?: string
  onClick: () => void
}) {
  const contextTone = {
    up: 'text-green-600',
    down: 'text-destructive',
    muted: 'text-muted-foreground',
    alert: 'text-destructive',
  }
  return (
    <button
      type="button"
      onClick={onClick}
      className={`group p-5 text-left transition-colors hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset ${divider}`}
    >
      <p className="flex items-center gap-0.5 text-xs font-medium text-muted-foreground">
        {label}
        <ChevronRight className="h-3 w-3 opacity-0 transition-opacity group-hover:opacity-60" />
      </p>
      <p className={`mt-1.5 text-[26px] leading-8 font-semibold tracking-tight tabular-nums ${valueClass}`}>{value}</p>
      <p className={`mt-1 text-xs min-h-4 ${context ? contextTone[context.tone] : ''}`}>
        {context ? `${context.tone === 'up' ? '▲ ' : context.tone === 'down' ? '▼ ' : ''}${context.text}` : ' '}
      </p>
    </button>
  )
}

function SectionHeader({ title, caption, linkLabel, onLink }: {
  title: string
  caption?: string
  linkLabel?: string
  onLink?: () => void
}) {
  return (
    <div className="mb-3 flex items-start justify-between gap-2">
      <div>
        <h2 className="text-sm font-semibold">{title}</h2>
        {caption && <p className="text-xs text-muted-foreground mt-0.5">{caption}</p>}
      </div>
      {linkLabel && (
        <button
          type="button"
          onClick={onLink}
          className="flex items-center text-xs font-medium text-primary hover:underline"
        >
          {linkLabel}
          <ChevronRight className="h-3.5 w-3.5" />
        </button>
      )}
    </div>
  )
}

function ListSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-3 py-1">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center justify-between gap-4">
          <Skeleton className="h-4 flex-1" />
          <Skeleton className="h-4 w-16" />
        </div>
      ))}
    </div>
  )
}

function ChartTooltip({ active, payload }: {
  active?: boolean
  payload?: Array<{ payload: { fullDate: string; revenue: number; count: number } }>
}) {
  if (!active || !payload?.length) return null
  const d = payload[0].payload
  return (
    <div className="rounded-lg border bg-card px-3 py-2 shadow-lg">
      <p className="text-xs font-medium text-muted-foreground">{d.fullDate}</p>
      <p className="mt-0.5 text-sm font-semibold tabular-nums">{formatUSD(d.revenue)}</p>
      {d.count > 0 && (
        <p className="text-xs text-muted-foreground">{d.count} invoice{d.count === 1 ? '' : 's'}</p>
      )}
    </div>
  )
}

export function Dashboard() {
  const { user, hasPermission } = useAuth()
  const navigate = useNavigate()
  const [dayCloseDate, setDayCloseDate] = useState(new Date().toISOString().slice(0, 10))
  const { data: summary, isLoading: summaryLoading } = useDashboardSummary()
  const { data: dailyRevenue, isLoading: revenueLoading } = useDailyRevenue()
  const { data: lowStockItems, isLoading: lowStockLoading } = useLowStockProducts()
  const { data: dayClose, isLoading: dayCloseLoading } = useDayClose(dayCloseDate)

  const mStart = new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().slice(0, 10)
  const tToday = new Date().toISOString().slice(0, 10)
  const { data: monthSales } = useSalesReport(mStart, tToday, 'day')
  const { data: monthPnl } = usePnL(mStart, tToday)
  const { data: inventory } = useInventoryReport(90)

  const lowStock = lowStockItems || []
  const criticalCount = lowStock.filter((p) => p.stock_quantity <= 0).length
  const warningCount = lowStock.length - criticalCount

  const recentInvoices = summary?.recent_invoices || []
  const recentJobs = summary?.recent_jobs || []

  // Continuous 7-day series ending today; the API omits days with no paid
  // invoices, which would otherwise distort the week's shape.
  const revenueByDate = new Map((dailyRevenue || []).map((d) => [d.date, d]))
  const chartData = Array.from({ length: 7 }, (_, i) => {
    const date = new Date()
    date.setDate(date.getDate() - (6 - i))
    const key = date.toISOString().slice(0, 10)
    const d = revenueByDate.get(key)
    return {
      name: date.toLocaleDateString('en-US', { weekday: 'short', day: 'numeric' }),
      fullDate: date.toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric' }),
      revenue: d?.revenue_usd ?? 0,
      count: d?.invoice_count ?? 0,
      // today's bar solid, past days tinted — temporal emphasis
      fill: i === 6 ? 'var(--color-primary)' : 'color-mix(in srgb, var(--color-primary) 55%, transparent)',
    }
  })
  const hasRevenueData = chartData.some((d) => d.revenue > 0)

  const todayRevenue = chartData[6]?.revenue ?? 0
  const yesterdayRevenue = chartData[5]?.revenue ?? 0
  const revenueDelta = yesterdayRevenue > 0
    ? Math.round(((todayRevenue - yesterdayRevenue) / yesterdayRevenue) * 100)
    : null

  const dayCloseTotal = (dayClose?.payment_methods || []).reduce((sum, pm) => sum + pm.total, 0)

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Welcome back, {user?.full_name || user?.username}</p>
        </div>
        <p className="text-sm text-muted-foreground">
          {new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
        </p>
      </div>

      {/* Stat strip */}
      {summaryLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 bg-card rounded-lg shadow-sm">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="p-5">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="mt-2.5 h-8 w-28" />
            </div>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-2 lg:grid-cols-4 bg-card rounded-lg shadow-sm overflow-hidden">
          <Stat
            label="Today's revenue"
            value={formatUSD(summary?.today_revenue || 0)}
            valueClass="text-primary"
            context={revenueDelta === null ? undefined : {
              text: `${Math.abs(revenueDelta)}% vs yesterday`,
              tone: revenueDelta >= 0 ? 'up' : 'down',
            }}
            onClick={() => navigate('/invoices?payment_status=paid')}
          />
          <Stat
            label="Outstanding"
            value={formatUSD(summary?.outstanding_usd || 0)}
            context={(summary?.unpaid_count || 0) > 0 ? {
              text: `${summary?.unpaid_count} unpaid invoice${summary?.unpaid_count === 1 ? '' : 's'}`,
              tone: 'muted',
            } : undefined}
            divider="border-l"
            onClick={() => navigate('/invoices?payment_status=unpaid')}
          />
          <Stat
            label="Jobs today"
            value={String(summary?.today_jobs || 0)}
            divider="border-t lg:border-t-0 lg:border-l"
            onClick={() => navigate('/service-jobs')}
          />
          <Stat
            label="Low stock"
            value={String(summary?.low_stock_count || 0)}
            valueClass={(summary?.low_stock_count || 0) > 0 ? 'text-accent' : ''}
            context={criticalCount > 0
              ? { text: `${criticalCount} out of stock`, tone: 'alert' }
              : warningCount > 0
                ? { text: `${warningCount} below threshold`, tone: 'muted' }
                : undefined}
            divider="border-l border-t lg:border-t-0"
            onClick={() => navigate('/inventory?low=1')}
          />
        </div>
      )}

      {/* This month strip */}
      <div className="grid grid-cols-1 sm:grid-cols-3 bg-card rounded-lg shadow-sm overflow-hidden">
        <Stat
          label="Revenue this month"
          value={formatUSD(monthSales?.revenue_usd || 0)}
          context={monthSales && monthSales.prev_revenue_usd > 0 ? {
            text: `${Math.abs(monthSales.revenue_change_pct)}% vs last month`,
            tone: monthSales.revenue_change_pct >= 0 ? 'up' : 'down',
          } : undefined}
          onClick={() => navigate('/analytics/sales')}
        />
        <Stat
          label="Net profit this month"
          value={formatUSD(monthPnl?.net_profit || 0)}
          valueClass={(monthPnl?.net_profit || 0) >= 0 ? 'text-green-600' : 'text-destructive'}
          context={monthPnl ? { text: `${monthPnl.net_margin_pct}% margin · after expenses`, tone: 'muted' } : undefined}
          divider="border-t sm:border-t-0 sm:border-l"
          onClick={() => navigate('/analytics/pnl')}
        />
        <Stat
          label="Stock value (cost)"
          value={formatUSD(inventory?.valuation.cost_value || 0)}
          context={inventory ? { text: `${inventory.valuation.units_on_hand} units on hand`, tone: 'muted' } : undefined}
          divider="border-t sm:border-t-0 sm:border-l"
          onClick={() => navigate('/analytics/inventory')}
        />
      </div>

      {/* Revenue chart + recent invoices */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        <div className="bg-card rounded-lg p-5 shadow-sm lg:col-span-2">
          <SectionHeader title="Revenue, last 7 days" caption="Paid invoices only" />
          {revenueLoading ? (
            <Skeleton className="h-[240px] w-full" />
          ) : !hasRevenueData ? (
            <div className="flex h-[240px] items-center justify-center text-sm text-muted-foreground">
              No paid invoices in the last 7 days
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <BarChart data={chartData} margin={{ top: 4, right: 4, bottom: 0, left: -12 }}>
                <CartesianGrid vertical={false} stroke="var(--color-border)" strokeOpacity={0.7} />
                <XAxis
                  dataKey="name"
                  tick={{ fontSize: 11, fill: 'var(--color-muted-foreground)' }}
                  axisLine={false}
                  tickLine={false}
                  dy={6}
                />
                <YAxis
                  tick={{ fontSize: 11, fill: 'var(--color-muted-foreground)' }}
                  axisLine={false}
                  tickLine={false}
                  tickCount={4}
                  tickFormatter={(v) => `$${v}`}
                />
                <Tooltip content={<ChartTooltip />} cursor={{ fill: 'var(--color-muted)', fillOpacity: 0.5 }} />
                <Bar dataKey="revenue" radius={[4, 4, 0, 0]} maxBarSize={40} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>

        <div className="bg-card rounded-lg p-5 shadow-sm">
          <SectionHeader title="Recent invoices" linkLabel="View all" onLink={() => navigate('/invoices')} />
          {summaryLoading ? (
            <ListSkeleton />
          ) : recentInvoices.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">No invoices yet</p>
          ) : (
            <div className="divide-y divide-border/70">
              {recentInvoices.map((inv) => (
                <div
                  key={inv.id}
                  className="flex cursor-pointer items-center justify-between gap-3 py-2.5 transition-colors hover:bg-muted/40 -mx-2 px-2 rounded"
                  onClick={() => navigate(`/invoices/${inv.id}`)}
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{inv.customer_name}</p>
                    <p className="font-mono text-xs text-muted-foreground">{inv.invoice_number}</p>
                  </div>
                  <div className="flex flex-shrink-0 items-center gap-2">
                    <span className="text-sm tabular-nums">{formatUSD(inv.total_usd)}</span>
                    <StatusBadge status={inv.status} />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Recent jobs + low stock */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <div className="bg-card rounded-lg p-5 shadow-sm">
          <SectionHeader title="Recent service jobs" linkLabel="View all" onLink={() => navigate('/service-jobs')} />
          {summaryLoading ? (
            <ListSkeleton />
          ) : recentJobs.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">No jobs yet</p>
          ) : (
            <div className="divide-y divide-border/70">
              {recentJobs.map((job) => (
                <div
                  key={job.id}
                  className="flex cursor-pointer items-center justify-between gap-3 py-2.5 transition-colors hover:bg-muted/40 -mx-2 px-2 rounded"
                  onClick={() => navigate(`/service-jobs/${job.id}`)}
                >
                  <div className="min-w-0">
                    <p className={`truncate text-sm font-medium ${job.customer_name ? '' : 'text-muted-foreground'}`}>
                      {job.customer_name || 'Walk-in'}
                    </p>
                    <p className="font-mono text-xs text-muted-foreground">{job.job_number}</p>
                  </div>
                  <StatusBadge status={job.status} />
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-card rounded-lg p-5 shadow-sm">
          <SectionHeader
            title="Low stock items"
            caption="Stock on hand vs. alert threshold"
            linkLabel="View all"
            onLink={() => navigate('/inventory?low=1')}
          />
          {lowStockLoading ? (
            <ListSkeleton />
          ) : lowStock.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">All items are well stocked</p>
          ) : (
            <div className="divide-y divide-border/70">
              {lowStock.slice(0, 5).map((p) => {
                const out = p.stock_quantity <= 0
                return (
                  <div key={p.id} className="flex items-center justify-between gap-3 py-2.5">
                    <div className="flex min-w-0 items-center gap-2.5">
                      <span className={`h-2 w-2 flex-shrink-0 rounded-full ${out ? 'bg-destructive' : 'bg-accent'}`} />
                      <span className="truncate text-sm">{p.name}</span>
                    </div>
                    <span className={`flex-shrink-0 text-xs tabular-nums ${out ? 'font-medium text-destructive' : 'text-muted-foreground'}`}>
                      {out ? 'Out of stock' : `${p.stock_quantity} of ${p.min_stock_alert}`}
                    </span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>

      {/* Cash drawer */}
      {hasPermission('invoice:create') && <CashDrawer />}

      {/* Day close */}
      <div className="bg-card rounded-lg p-5 shadow-sm print:shadow-none">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <div>
            <h2 className="text-sm font-semibold">Day close</h2>
            <p className="text-xs text-muted-foreground mt-0.5">Payments received on the selected date</p>
          </div>
          <div className="flex items-center gap-2 print:hidden">
            <Input
              type="date"
              value={dayCloseDate}
              onChange={(e) => setDayCloseDate(e.target.value)}
              className="w-40"
            />
            <Button variant="outline" size="sm" className="gap-1.5" onClick={() => window.print()}>
              <Printer className="h-4 w-4" /> Print
            </Button>
          </div>
        </div>
        {dayCloseLoading ? (
          <ListSkeleton rows={4} />
        ) : !dayClose || (dayClose.payment_methods.length === 0 && dayClose.voided_count === 0) ? (
          <p className="py-8 text-center text-sm text-muted-foreground">No payments recorded on this date</p>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-x-10 gap-y-4">
            <div>
              <p className="mb-1 text-xs font-medium text-muted-foreground">By payment method</p>
              <div className="divide-y divide-border/70 text-sm">
                {dayClose.payment_methods.map((pm) => (
                  <div key={pm.method} className="flex items-center justify-between py-2">
                    <span className="capitalize">{pm.method}</span>
                    <span className="flex items-baseline gap-3">
                      <span className="text-xs text-muted-foreground">{pm.count}×</span>
                      <span className="tabular-nums">{formatUSD(pm.total)}</span>
                    </span>
                  </div>
                ))}
                {dayClose.voided_count > 0 && (
                  <div className="flex items-center justify-between py-2 text-destructive">
                    <span>Voided</span>
                    <span className="flex items-baseline gap-3">
                      <span className="text-xs">{dayClose.voided_count}×</span>
                      <span className="tabular-nums">−{formatUSD(dayClose.voided_total)}</span>
                    </span>
                  </div>
                )}
                <div className="flex items-center justify-between py-2 font-semibold">
                  <span>Total collected</span>
                  <span className="tabular-nums">{formatUSD(dayCloseTotal)}</span>
                </div>
              </div>
            </div>
            {dayClose.user_totals.length > 0 && (
              <div>
                <p className="mb-1 text-xs font-medium text-muted-foreground">By staff member</p>
                <div className="divide-y divide-border/70 text-sm">
                  {dayClose.user_totals.map((ut) => (
                    <div key={ut.user_id} className="flex items-center justify-between py-2">
                      <span>{ut.user_name || 'Unknown'}</span>
                      <span className="flex items-baseline gap-3">
                        <span className="text-xs text-muted-foreground">{ut.count}×</span>
                        <span className="tabular-nums">{formatUSD(ut.total)}</span>
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
