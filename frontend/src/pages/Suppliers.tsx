import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { formatUSD } from '@/utils/currency'
import { useSuppliers, useCreateSupplier } from '@/hooks/useSuppliers'
import type { Supplier } from '@/types/supplier'

export function Suppliers() {
  const navigate = useNavigate()
  const { data: suppliers, isLoading } = useSuppliers()
  const [showForm, setShowForm] = useState(false)

  const list = suppliers || []
  const totalOwed = list.reduce((s, x) => s + x.outstanding, 0)

  return (
    <div className="space-y-6">
      <PageHeader title="Suppliers" actions={<Button size="sm" onClick={() => setShowForm(true)}>Add Supplier</Button>} />

      {isLoading ? (
        <Skeleton className="h-40 w-full" />
      ) : (
        <>
          {totalOwed > 0 && (
            <div className="rounded-lg bg-card p-5 shadow-sm">
              <p className="text-xs text-muted-foreground">Total payable (owed to suppliers)</p>
              <p className="mt-1 text-2xl font-bold tabular-nums text-destructive">{formatUSD(totalOwed)}</p>
            </div>
          )}

          {/* Desktop */}
          <div className="hidden overflow-hidden rounded-lg bg-card shadow-sm md:block">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-xs text-muted-foreground">
                  <th className="px-4 py-2.5 font-medium">Supplier</th>
                  <th className="px-4 py-2.5 font-medium">Phone</th>
                  <th className="px-4 py-2.5 text-right font-medium">Purchases</th>
                  <th className="px-4 py-2.5 text-right font-medium">Total bought</th>
                  <th className="px-4 py-2.5 text-right font-medium">Owed</th>
                </tr>
              </thead>
              <tbody>
                {list.length === 0 ? (
                  <tr><td colSpan={5} className="px-4 py-10 text-center text-muted-foreground">No suppliers yet</td></tr>
                ) : list.map((s) => (
                  <tr key={s.id} className="cursor-pointer border-b last:border-0 hover:bg-muted/50" onClick={() => navigate(`/suppliers/${s.id}`)}>
                    <td className="px-4 py-2.5 font-medium">{s.name}</td>
                    <td className="px-4 py-2.5 text-muted-foreground">{s.phone || '—'}</td>
                    <td className="px-4 py-2.5 text-right tabular-nums">{s.purchase_count}</td>
                    <td className="px-4 py-2.5 text-right tabular-nums">{formatUSD(s.total_purchased)}</td>
                    <td className={`px-4 py-2.5 text-right tabular-nums font-medium ${s.outstanding > 0 ? 'text-destructive' : 'text-muted-foreground'}`}>{s.outstanding > 0 ? formatUSD(s.outstanding) : '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Mobile */}
          <div className="space-y-3 md:hidden">
            {list.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted-foreground">No suppliers yet</p>
            ) : list.map((s) => (
              <div key={s.id} className="cursor-pointer rounded-lg bg-card p-4 shadow-sm" onClick={() => navigate(`/suppliers/${s.id}`)}>
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{s.name}</p>
                    {s.phone && <p className="text-xs text-muted-foreground">{s.phone}</p>}
                  </div>
                  {s.outstanding > 0 && <span className="flex-shrink-0 text-sm font-medium tabular-nums text-destructive">{formatUSD(s.outstanding)} owed</span>}
                </div>
                <p className="mt-2 text-xs text-muted-foreground">{s.purchase_count} purchase{s.purchase_count === 1 ? '' : 's'} · {formatUSD(s.total_purchased)} total</p>
              </div>
            ))}
          </div>
        </>
      )}

      {showForm && <SupplierForm onClose={() => setShowForm(false)} />}
    </div>
  )
}

function SupplierForm({ onClose }: { onClose: () => void }) {
  const create = useCreateSupplier()
  const [f, setF] = useState<Partial<Supplier>>({ name: '', phone: '', email: '', address: '', notes: '' })
  const set = (k: string, v: string) => setF((prev) => ({ ...prev, [k]: v }))

  const submit = () => {
    if (!f.name?.trim()) return
    create.mutate(
      { name: f.name.trim(), phone: f.phone || undefined, email: f.email || undefined, address: f.address || undefined, notes: f.notes || undefined },
      { onSuccess: onClose }
    )
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-auto bg-black/50 pt-10">
      <div className="mx-4 w-full max-w-lg rounded-lg bg-card p-5 shadow-lg">
        <h2 className="mb-3 text-sm font-semibold">Add Supplier</h2>
        <div className="space-y-2">
          <Input placeholder="Name *" value={f.name || ''} onChange={(e) => set('name', e.target.value)} />
          <div className="grid grid-cols-2 gap-2">
            <Input placeholder="Phone" value={f.phone || ''} onChange={(e) => set('phone', e.target.value)} />
            <Input placeholder="Email" value={f.email || ''} onChange={(e) => set('email', e.target.value)} />
          </div>
          <Input placeholder="Address" value={f.address || ''} onChange={(e) => set('address', e.target.value)} />
          <Input placeholder="Notes" value={f.notes || ''} onChange={(e) => set('notes', e.target.value)} />
        </div>
        <div className="mt-3 flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} disabled={!f.name?.trim() || create.isPending}>{create.isPending ? 'Saving…' : 'Create'}</Button>
        </div>
      </div>
    </div>
  )
}
