import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { Trash2, MoreVertical } from 'lucide-react'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { PageHeader } from '@/components/layout/PageHeader'
import { formatUSD } from '@/utils/currency'
import { formatDate } from '@/utils/date'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCustomers, useCreateCustomer, useDeleteCustomer } from '@/hooks/useCustomers'
import { useDebounce } from '@/hooks/useDebounce'
import { CustomerForm } from '@/components/customer/CustomerForm'
import { TableCard, TableFooter, Th, ActionsTh } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

const customerTypeLabel = (t?: string) =>
  t === 'garage' ? 'Garage' : t === 'company' ? 'Company' : 'Retail'

export function Customers() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [menuId, setMenuId] = useState<number | null>(null)
  const debouncedSearch = useDebounce(search, 300)

  const { data, isLoading } = useCustomers({ search: debouncedSearch || undefined, page })
  const createMutation = useCreateCustomer()
  const deleteMutation = useDeleteCustomer()

  const customers = data?.data || []
  const meta = data?.meta

  const handleCreate = (formData: any) => {
    createMutation.mutate(formData, { onSuccess: () => setShowForm(false) })
  }

  const handleDelete = () => {
    if (deletingId) {
      deleteMutation.mutate(deletingId, { onSuccess: () => setDeletingId(null) })
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Customers"
        actions={
          <Button onClick={() => setShowForm(true)} size="sm">Add Customer</Button>
        }
      />

      {isLoading ? (
        <div className="bg-card rounded-lg shadow-sm p-4">
          <TableSkeleton rows={5} cols={5} />
        </div>
      ) : (
        <>
          <div className="hidden md:block">
            <TableCard
              toolbar={
                <Input
                  placeholder="Search name, phone, or plate…"
                  value={search}
                  onChange={(e) => { setSearch(e.target.value); setPage(1) }}
                  className="w-56"
                />
              }
              footer={meta && (
                <TableFooter total={meta.total} page={meta.page} totalPages={meta.total_pages} onPage={setPage} noun="customers" />
              )}
            >
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-card sticky top-0 z-10">
                    <Th>Name</Th>
                    <Th>Phone</Th>
                    <Th className="text-right">Spent</Th>
                    <Th>Last visit</Th>
                    <Th className="text-right">Vehicles</Th>
                    <ActionsTh />
                  </tr>
                </thead>
                <tbody>
                  {customers.length === 0 ? (
                    <tr><td colSpan={6} className="px-4 py-10 text-center text-sm text-muted-foreground">No customers found</td></tr>
                  ) : (
                    customers.map((c) => (
                      <tr key={c.id} className="group border-b last:border-0 hover:bg-muted/50 transition-colors duration-100 cursor-pointer"
                        onClick={() => navigate(`/customers/${c.id}`)}>
                        <td className="px-4 py-2.5">
                          <p className="font-medium">{c.name}{' '}
                            <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-muted-foreground">{customerTypeLabel(c.customer_type)}</span>
                          </p>
                          {c.vehicle_plates && <p className="text-xs text-muted-foreground">{c.vehicle_plates}</p>}
                        </td>
                        <td className="px-4 py-2.5 text-muted-foreground">{c.phone || '—'}</td>
                        <td className="px-4 py-2.5 text-right tabular-nums">{c.total_spent ? formatUSD(c.total_spent) : <span className="text-muted-foreground">—</span>}</td>
                        <td className="px-4 py-2.5 text-muted-foreground">{c.last_visit ? formatDate(c.last_visit) : '—'}</td>
                        <td className="px-4 py-2.5 text-right tabular-nums">{c.vehicle_count > 0 ? c.vehicle_count : <span className="text-muted-foreground">—</span>}</td>
                        <td className="px-4 py-2.5 text-right">
                          <div className="text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
                            <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={(e) => { e.stopPropagation(); setDeletingId(c.id) }} aria-label="Delete" title="Delete customer">
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </TableCard>
          </div>

          <div className="md:hidden space-y-3">
            <Input
              placeholder="Search name, phone, or plate…"
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1) }}
              className="w-full"
            />
            {customers.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-8">No customers found</p>
            ) : (
              customers.map((c) => (
                <div key={c.id} className="bg-card rounded-lg p-4 shadow-sm cursor-pointer"
                  onClick={() => navigate(`/customers/${c.id}`)}>
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="font-medium text-sm truncate">{c.name}{' '}
                        <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-muted-foreground">{customerTypeLabel(c.customer_type)}</span>
                      </p>
                      {c.phone && <p className="text-xs text-muted-foreground mt-0.5">{c.phone}</p>}
                      {c.vehicle_plates && <p className="text-xs text-muted-foreground mt-0.5 truncate">{c.vehicle_plates}</p>}
                    </div>
                    <div className="relative flex-shrink-0">
                      <button onClick={(e) => { e.stopPropagation(); setMenuId(menuId === c.id ? null : c.id) }}
                        className="-mr-1 -mt-1 rounded p-1 text-muted-foreground hover:bg-muted" aria-label="More options">
                        <MoreVertical className="h-4 w-4" />
                      </button>
                      {menuId === c.id && (
                        <>
                          <div className="fixed inset-0 z-10" onClick={(e) => { e.stopPropagation(); setMenuId(null) }} />
                          <div className="absolute right-0 top-7 z-20 w-32 overflow-hidden rounded-md border bg-card shadow-lg">
                            <button onClick={(e) => { e.stopPropagation(); setMenuId(null); setDeletingId(c.id) }}
                              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-destructive hover:bg-muted">
                              <Trash2 className="h-4 w-4" /> Delete
                            </button>
                          </div>
                        </>
                      )}
                    </div>
                  </div>
                  <div className="mt-2 flex items-center justify-between text-xs text-muted-foreground">
                    <span>{c.last_visit ? `Last visit ${formatDate(c.last_visit)}` : 'No purchases yet'}</span>
                    <span className="tabular-nums">{c.total_spent ? formatUSD(c.total_spent) : ''}{c.vehicle_count > 0 ? ` · ${c.vehicle_count} veh` : ''}</span>
                  </div>
                </div>
              ))
            )}
          </div>

          {meta && meta.total_pages > 1 && (
            <div className="md:hidden flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Page {meta.page} of {meta.total_pages} ({meta.total} items)</span>
              <div className="flex gap-1">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
                <Button variant="outline" size="sm" disabled={page >= meta.total_pages} onClick={() => setPage(page + 1)}>Next</Button>
              </div>
            </div>
          )}
        </>
      )}

      {showForm && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-10 bg-black/50">
          <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-lg mx-4">
            <h2 className="text-sm font-semibold mb-3">Add Customer</h2>
            <CustomerForm
              onSubmit={handleCreate}
              onCancel={() => setShowForm(false)}
              loading={createMutation.isPending}
            />
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deletingId !== null}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDelete}
        title="Delete Customer"
        message={`Delete ${customers.find(c => c.id === deletingId)?.name || 'this customer'}? This will also remove their vehicles and history.`}
        destructive
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
