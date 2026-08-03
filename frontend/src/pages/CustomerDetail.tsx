import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { PageHeader } from '@/components/layout/PageHeader'
import { formatDate } from '@/utils/date'
import { formatUSD } from '@/utils/currency'
import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useCustomer, useServiceHistory, useCreateVehicle, useDeleteVehicle, useUpdateCustomer } from '@/hooks/useCustomers'
import { useSettings } from '@/hooks/useSettings'
import { VehicleForm } from '@/components/customer/VehicleForm'
import { CustomerForm } from '@/components/customer/CustomerForm'
import { CustomerDeposits } from '@/components/customer/CustomerDeposits'
import { Button } from '@/components/ui/button'

export function CustomerDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const customerId = parseInt(id || '0')

  const { data: customer, isLoading } = useCustomer(customerId)
  const { data: history } = useServiceHistory(customerId)
  const createVehicle = useCreateVehicle()
  const deleteVehicle = useDeleteVehicle()
  const updateCustomer = useUpdateCustomer()
  const { data: settings } = useSettings()

  const [showVehicleForm, setShowVehicleForm] = useState(false)
  const [showEditForm, setShowEditForm] = useState(false)
  const [deletingVehicleId, setDeletingVehicleId] = useState<number | null>(null)

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!customer) return <p className="text-sm text-destructive">Customer not found</p>

  const handleAddVehicle = (data: any) => {
    createVehicle.mutate({ customerId, data }, { onSuccess: () => setShowVehicleForm(false) })
  }

  const handleEditCustomer = (data: any) => {
    updateCustomer.mutate({ id: customerId, data }, { onSuccess: () => setShowEditForm(false) })
  }

  const handleDeleteVehicle = () => {
    if (deletingVehicleId) {
      deleteVehicle.mutate(deletingVehicleId, { onSuccess: () => setDeletingVehicleId(null) })
    }
  }

  return (
    <div className="max-w-4xl space-y-6">
      <PageHeader
        title={<>{customer.name}{' '}<span className="rounded bg-muted px-1.5 py-0.5 align-middle text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
          {customer.customer_type === 'garage' ? 'Garage' : customer.customer_type === 'company' ? 'Company' : 'Retail'}
        </span></>}
        backTo={-1}
        breadcrumb="Customers"
        actions={
          <>
            <Button size="sm" variant="outline" onClick={() => navigate(`/sale?customer=${customer.id}`)}>
              New Sale
            </Button>
            <Button size="sm" onClick={() => navigate(`/service-jobs?new=1&customer=${customer.id}`)}>
              New Job
            </Button>
          </>
        }
      />

      {/* Value header */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <Kpi label="Total spent" value={formatUSD(customer.total_spent || 0)} />
        <Kpi label="Visits" value={String(customer.visit_count || 0)} />
        <Kpi label="Last visit" value={customer.last_visit ? formatDate(customer.last_visit) : '—'} />
        <Kpi label="Outstanding" value={formatUSD(customer.outstanding || 0)} alert={(customer.outstanding || 0) > 0} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
          <p className="text-sm font-semibold">Contact</p>
          {customer.phone && <p className="text-sm"><a href={`tel:${customer.phone}`} className="text-primary hover:underline">{customer.phone}</a></p>}
          {customer.email && <p className="text-sm"><a href={`mailto:${customer.email}`} className="text-primary hover:underline">{customer.email}</a></p>}
          {customer.address && (
            <p className="text-sm">
              <a href={`https://maps.google.com/?q=${encodeURIComponent(customer.address)}`} target="_blank" rel="noreferrer" className="text-primary hover:underline">{customer.address}</a>
            </p>
          )}
          {customer.customer_since && <p className="text-xs text-muted-foreground">Customer since {customer.customer_since}</p>}
          <Button variant="ghost" size="sm" className="mt-1" onClick={() => setShowEditForm(true)}>Edit</Button>
        </div>
        <div className="bg-card rounded-lg p-5 shadow-sm">
          <p className="text-sm font-semibold mb-2">Notes</p>
          <p className="text-sm">{customer.notes || 'No notes'}</p>
        </div>
      </div>

      <CustomerDeposits customerId={customerId} />

      <div className="bg-card rounded-lg p-5 shadow-sm">
        <div className="flex items-center justify-between mb-2">
          <p className="text-sm font-semibold">Vehicles</p>
          <Button size="sm" variant="outline" onClick={() => setShowVehicleForm(true)}>Add Vehicle</Button>
        </div>
        {customer.vehicles.length === 0 ? (
          <p className="text-sm text-muted-foreground">No vehicles registered</p>
        ) : (
          <div className="space-y-1">
            {customer.vehicles.map((v) => (
              <div key={v.id} className="flex items-center justify-between py-1.5 border-b last:border-0">
                <div
                  className="min-w-0 cursor-pointer rounded -mx-1 px-1 py-0.5 hover:bg-muted/40"
                  onClick={() => navigate(`/vehicles/${v.id}`)}
                >
                  <p className="text-sm font-medium text-primary hover:underline">{v.plate_number}</p>
                  <p className="text-xs text-muted-foreground">
                    {[v.make, v.model].filter(Boolean).join(' ')}
                    {v.year ? ` (${v.year})` : ''}
                    {v.color ? ` - ${v.color}` : ''}
                  </p>
                </div>
                <Button variant="ghost" size="sm" onClick={() => setDeletingVehicleId(v.id)}>Remove</Button>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="bg-card rounded-lg p-5 shadow-sm">
        <p className="text-sm font-semibold mb-2">Activity</p>
        {!history || history.length === 0 ? (
          <p className="text-sm text-muted-foreground">No sales or jobs yet</p>
        ) : (
          <div className="space-y-1">
            {history.map((item) => {
              const isInvoice = item.type === 'invoice'
              const href = isInvoice ? `/invoices/${item.id}` : `/service-jobs/${item.id}`
              return (
                <div key={`${item.type}-${item.id}`} className="cursor-pointer rounded py-1.5 border-b last:border-0 hover:bg-muted/40 -mx-1 px-1" onClick={() => navigate(href)}>
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex min-w-0 items-center gap-2">
                      <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide ${isInvoice ? 'bg-primary/10 text-primary' : 'bg-accent/10 text-accent'}`}>
                        {isInvoice ? 'Sale' : 'Job'}
                      </span>
                      <p className="truncate text-sm font-medium">{item.ref}</p>
                    </div>
                    <span className="flex-shrink-0 text-sm tabular-nums">{formatUSD(item.amount)}</span>
                  </div>
                  <div className="mt-0.5 flex items-center justify-between text-xs text-muted-foreground">
                    <span className="truncate">{item.plate ? `${item.plate} · ` : ''}{item.title}{item.date ? ` · ${formatDate(item.date)}` : ''}</span>
                    <span className="flex-shrink-0">
                      {item.outstanding > 0
                        ? <span className="font-medium text-destructive">{formatUSD(item.outstanding)} due</span>
                        : <span className="capitalize">{item.status}</span>}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {showVehicleForm && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-10 bg-black/50">
          <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-lg mx-4">
            <h2 className="text-sm font-semibold mb-3">Add Vehicle</h2>
            <VehicleForm onSubmit={handleAddVehicle} onCancel={() => setShowVehicleForm(false)} loading={createVehicle.isPending} defaultUnit={settings?.distance_unit === 'mi' ? 'mi' : 'km'} />
          </div>
        </div>
      )}

      {showEditForm && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-10 bg-black/50">
          <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-lg mx-4">
            <h2 className="text-sm font-semibold mb-3">Edit Customer</h2>
            <CustomerForm
              initial={customer}
              onSubmit={handleEditCustomer}
              onCancel={() => setShowEditForm(false)}
              loading={updateCustomer.isPending}
            />
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deletingVehicleId !== null}
        onClose={() => setDeletingVehicleId(null)}
        onConfirm={handleDeleteVehicle}
        title="Remove Vehicle"
        message="Remove this vehicle? This cannot be undone."
        destructive
        confirmLabel="Remove"
      />
    </div>
  )
}

function Kpi({ label, value, alert }: { label: string; value: string; alert?: boolean }) {
  return (
    <div className="rounded-lg bg-card p-4 shadow-sm">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={`mt-1 text-xl font-bold tabular-nums ${alert ? 'text-destructive' : ''}`}>{value}</p>
    </div>
  )
}
