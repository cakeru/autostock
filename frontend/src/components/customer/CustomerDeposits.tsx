import { useState } from 'react'
import { Pencil, X } from 'lucide-react'
import { useInvoices } from '@/hooks/useInvoices'
import { useCustomerDeposits, useCreateDeposit, useUpdateDeposit, useApplyDeposit, useRefundDeposit } from '@/hooks/useDeposits'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { formatUSD } from '@/utils/currency'

export function CustomerDeposits({ customerId }: { customerId: number }) {
  const { data: deposits } = useCustomerDeposits(customerId, 'held')
  const { data: invData } = useInvoices({ customer_id: customerId, per_page: 50 })
  const create = useCreateDeposit()
  const update = useUpdateDeposit()
  const apply = useApplyDeposit()
  const refund = useRefundDeposit()

  const [amount, setAmount] = useState('')
  const [note, setNote] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editAmount, setEditAmount] = useState('')
  const [editNote, setEditNote] = useState('')
  const [applyingId, setApplyingId] = useState<number | null>(null)
  const [applyInvoice, setApplyInvoice] = useState('')

  const held = deposits || []
  const totalHeld = held.reduce((s, d) => s + d.amount, 0)
  const unpaid = (invData?.data || []).filter((i: any) => i.status !== 'voided' && i.total_usd - i.paid_amount > 0.001)

  const doCreate = () => {
    const a = parseFloat(amount)
    if (isNaN(a) || a <= 0) return
    create.mutate({ customer_id: customerId, amount: a, note: note.trim() || undefined }, { onSuccess: () => { setAmount(''); setNote('') } })
  }
  const doApply = (id: number) => {
    if (!applyInvoice) return
    apply.mutate({ id, invoiceId: parseInt(applyInvoice) }, { onSuccess: () => { setApplyingId(null); setApplyInvoice('') } })
  }

  const startEdit = (d: { id: number; amount: number; note?: string }) => {
    setEditingId(d.id)
    setEditAmount(String(d.amount))
    setEditNote(d.note || '')
  }
  const saveEdit = () => {
    const a = parseFloat(editAmount)
    if (!editingId || isNaN(a) || a <= 0) return
    update.mutate({ id: editingId, data: { customer_id: customerId, amount: a, note: editNote.trim() || undefined } },
      { onSuccess: () => setEditingId(null) })
  }

  return (
    <div className="rounded-lg bg-card p-5 shadow-sm">
      <div className="mb-2 flex items-center justify-between">
        <p className="text-sm font-semibold">Deposits / credit</p>
        {totalHeld > 0 && <span className="text-sm font-medium text-emerald-600">{formatUSD(totalHeld)} held</span>}
      </div>

      {held.length === 0 ? (
        <p className="text-sm text-muted-foreground">No deposits held.</p>
      ) : (
        <div className="space-y-2">
          {held.map((d) => (
            <div key={d.id} className="rounded-md border p-2.5">
              {editingId === d.id ? (
                <div className="flex flex-wrap items-center gap-2">
                  <Input type="number" step="0.01" min="0" value={editAmount} onChange={(e) => setEditAmount(e.target.value)} className="w-28" />
                  <Input value={editNote} onChange={(e) => setEditNote(e.target.value)} placeholder="Note" className="flex-1 min-w-[8rem]" />
                  <Button size="sm" onClick={saveEdit} disabled={update.isPending}>{update.isPending ? 'Saving…' : 'Save'}</Button>
                  <Button size="sm" variant="ghost" onClick={() => setEditingId(null)}><X className="h-4 w-4" /></Button>
                </div>
              ) : (
                <div className="flex items-center justify-between gap-2">
                  <div className="min-w-0">
                    <span className="font-medium tabular-nums">{formatUSD(d.amount)}</span>
                    {d.note && <span className="ml-2 text-xs text-muted-foreground">{d.note}</span>}
                  </div>
                  <div className="flex flex-shrink-0 gap-1">
                    <button onClick={() => startEdit(d)} title="Edit deposit" className="text-muted-foreground hover:text-foreground"><Pencil className="h-3.5 w-3.5" /></button>
                    <Button variant="outline" size="sm" disabled={unpaid.length === 0}
                      title={unpaid.length === 0 ? 'No unpaid invoices to apply to' : ''}
                      onClick={() => { setApplyingId(applyingId === d.id ? null : d.id); setApplyInvoice('') }}>
                      Apply
                    </Button>
                    <Button variant="ghost" size="sm" className="text-destructive" disabled={refund.isPending} onClick={() => refund.mutate(d.id)}>Refund</Button>
                  </div>
                </div>
              )}
              {applyingId === d.id && (
                <div className="mt-2 flex items-center gap-2">
                  <Select value={applyInvoice} onChange={(e) => setApplyInvoice(e.target.value)} className="flex-1">
                    <option value="">Choose invoice…</option>
                    {unpaid.map((i: any) => <option key={i.id} value={i.id}>{i.invoice_number} — {formatUSD(i.total_usd - i.paid_amount)} due</option>)}
                  </Select>
                  <Button size="sm" disabled={!applyInvoice || apply.isPending} onClick={() => doApply(d.id)}>{apply.isPending ? 'Applying…' : 'Confirm'}</Button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      <div className="mt-3 flex flex-wrap items-center gap-2 border-t pt-3">
        <span className="text-sm text-muted-foreground">Take deposit $</span>
        <Input value={amount} onChange={(e) => setAmount(e.target.value)} type="number" step="0.01" min="0" className="w-24" />
        <Input value={note} onChange={(e) => setNote(e.target.value)} placeholder="Note (e.g. special order)" className="min-w-[8rem] flex-1" />
        <Button size="sm" disabled={!amount || create.isPending} onClick={doCreate}>{create.isPending ? 'Saving…' : 'Add'}</Button>
      </div>
    </div>
  )
}
