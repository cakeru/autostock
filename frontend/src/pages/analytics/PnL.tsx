import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Trash2, Plus } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { formatUSD } from '@/utils/currency'
import { reportsApi } from '@/services/settings'
import { usePnL } from '@/hooks/useAnalytics'
import { useExpenses, useCreateExpense, useDeleteExpense } from '@/hooks/useExpenses'
import { StatCard, Panel, RangeBar, monthStart, today, DataTable } from './shared'

export function PnLAnalytics() {
  const [from, setFrom] = useState(monthStart())
  const [to, setTo] = useState(today())
  const { data, isLoading } = usePnL(from, to)

  return (
    <div className="max-w-5xl space-y-5">
      <PageHeader title="Profit &amp; Loss" subtitle="Revenue minus cost of goods and operating expenses = net profit" />

      <RangeBar from={from} to={to} setFrom={setFrom} setTo={setTo} />

      {isLoading || !data ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-6">
          {[0, 1, 2, 3, 4, 5].map((i) => <div key={i} className="rounded-lg bg-card p-5 shadow-sm"><Skeleton className="h-3 w-14" /><Skeleton className="mt-2 h-8 w-20" /></div>)}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-6">
            <StatCard label="Revenue" value={formatUSD(data.revenue)} />
            <StatCard label="COGS" value={formatUSD(data.cogs)} accent="down" />
            <StatCard label="Gross profit" value={formatUSD(data.gross_profit)} sub={`${data.gross_margin_pct}% margin`} />
            <StatCard label="Payroll" value={formatUSD(data.payroll)} accent="down" />
            <StatCard label="Expenses" value={formatUSD(data.expenses)} accent="down" />
            <StatCard label="Net profit" value={formatUSD(data.net_profit)} sub={`${data.net_margin_pct}% margin`} accent={data.net_profit >= 0 ? 'up' : 'down'} />
          </div>

          <Panel title="P&L statement">
            <table className="w-full text-sm">
              <tbody>
                <PnLLine label="Revenue (net of discounts)" value={data.revenue} />
                {data.returns > 0 && <PnLLine label="Returns / refunds" value={-data.returns} muted indent />}
                <PnLLine label="Cost of goods sold" value={-data.cogs} muted />
                <PnLLine label="Gross profit" value={data.gross_profit} bold />
                {data.payroll_breakdown.map((p) => (
                  <PnLLine key={p.employee_id} label={p.name} value={-p.total} muted indent />
                ))}
                <PnLLine label="Payroll" value={-data.payroll} muted />
                {data.expense_categories.map((e) => (
                  <PnLLine key={e.category} label={e.category} value={-e.amount} muted indent />
                ))}
                <PnLLine label="Total operating expenses" value={-data.expenses} muted />
                <PnLLine label="Net profit" value={data.net_profit} bold top />
              </tbody>
            </table>
          </Panel>
        </>
      )}

      <CategoryBreakdown from={from} to={to} />

      <ExpensesManager from={from} to={to} />
    </div>
  )
}

// Gross margin by revenue stream (Parts & tires / Labor / Fees), carried over
// from the former standalone Reports page.
function CategoryBreakdown({ from, to }: { from: string; to: string }) {
  const { data, isLoading } = useQuery({
    queryKey: ['profit', from, to],
    queryFn: () => reportsApi.profit(from, to),
  })
  return (
    <Panel title="Gross margin by category">
      {isLoading || !data ? (
        <Skeleton className="h-24 w-full" />
      ) : (
        <DataTable
          rows={data.categories}
          rowKey={(c) => c.category}
          empty="No sales in this range"
          columns={[
            { header: 'Category', primary: true, cell: (c) => c.category },
            { header: 'Revenue', align: 'right', cell: (c) => formatUSD(c.revenue) },
            { header: 'Cost', align: 'right', cell: (c) => <span className="text-muted-foreground">{c.cost ? formatUSD(c.cost) : '—'}</span> },
            { header: 'Profit', align: 'right', cell: (c) => <span className={c.profit < 0 ? 'text-destructive' : ''}>{formatUSD(c.profit)}</span> },
            { header: 'Margin', align: 'right', cell: (c) => <span className="text-muted-foreground">{c.revenue > 0 ? `${c.margin_pct}%` : '—'}</span> },
          ]}
        />
      )}
    </Panel>
  )
}

function PnLLine({ label, value, bold, muted, indent, top }: { label: string; value: number; bold?: boolean; muted?: boolean; indent?: boolean; top?: boolean }) {
  return (
    <tr className={`${top ? 'border-t-2 border-foreground/20' : 'border-b last:border-0'}`}>
      <td className={`py-2 ${bold ? 'font-semibold' : ''} ${muted ? 'text-muted-foreground' : ''} ${indent ? 'pl-4' : ''}`}>{label}</td>
      <td className={`py-2 text-right tabular-nums ${bold ? 'font-semibold' : ''} ${value < 0 ? 'text-destructive' : ''}`}>
        {value < 0 ? `(${formatUSD(-value)})` : formatUSD(value)}
      </td>
    </tr>
  )
}

function ExpensesManager({ from, to }: { from: string; to: string }) {
  const { data: expenses, isLoading } = useExpenses(from, to)
  const create = useCreateExpense()
  const del = useDeleteExpense()
  const [category, setCategory] = useState('')
  const [amount, setAmount] = useState('')
  const [description, setDescription] = useState('')
  const [spentAt, setSpentAt] = useState(today())

  const submit = () => {
    const amt = parseFloat(amount)
    if (!category.trim() || !amt || amt <= 0) return
    create.mutate(
      { category: category.trim(), amount_usd: amt, description: description.trim() || undefined, spent_at: spentAt },
      { onSuccess: () => { setCategory(''); setAmount(''); setDescription('') } }
    )
  }

  return (
    <Panel title="Operating expenses">
      <div className="mb-4 flex flex-wrap items-end gap-2">
        <div className="space-y-1">
          <label className="text-xs text-muted-foreground">Category</label>
          <Input value={category} onChange={(e) => setCategory(e.target.value)} placeholder="Rent, Salaries…" className="w-40" list="expense-cats" />
          <datalist id="expense-cats">
            {['Rent', 'Salaries', 'Utilities', 'Supplies', 'Marketing', 'Maintenance', 'Other'].map((c) => <option key={c} value={c} />)}
          </datalist>
        </div>
        <div className="space-y-1">
          <label className="text-xs text-muted-foreground">Amount ($)</label>
          <Input value={amount} onChange={(e) => setAmount(e.target.value)} type="number" step="0.01" min="0" className="w-28" />
        </div>
        <div className="space-y-1">
          <label className="text-xs text-muted-foreground">Date</label>
          <Input value={spentAt} onChange={(e) => setSpentAt(e.target.value)} type="date" className="w-40" />
        </div>
        <div className="flex-1 space-y-1 min-w-[10rem]">
          <label className="text-xs text-muted-foreground">Note (optional)</label>
          <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Description" />
        </div>
        <Button size="sm" className="gap-1" onClick={submit} disabled={create.isPending || !category.trim() || !(parseFloat(amount) > 0)}>
          <Plus className="h-4 w-4" /> {create.isPending ? 'Adding…' : 'Add'}
        </Button>
      </div>

      {isLoading ? <Skeleton className="h-24 w-full" /> : (
        <DataTable
          rows={expenses || []}
          rowKey={(e) => e.id}
          empty="No expenses in this range"
          action={(e) => (
            <button onClick={() => del.mutate(e.id)} className="text-muted-foreground hover:text-destructive" aria-label="Delete expense">
              <Trash2 className="h-4 w-4" />
            </button>
          )}
          columns={[
            { header: 'Date', cell: (e) => <span className="tabular-nums text-muted-foreground">{e.spent_at}</span> },
            { header: 'Category', primary: true, cell: (e) => e.category },
            { header: 'Note', cell: (e) => <span className="text-muted-foreground">{e.description || '—'}</span> },
            { header: 'Amount', align: 'right', cell: (e) => formatUSD(e.amount_usd) },
          ]}
        />
      )}
    </Panel>
  )
}
