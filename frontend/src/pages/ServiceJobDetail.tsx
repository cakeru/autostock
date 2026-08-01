import { PageHeader } from '@/components/layout/PageHeader'
import { PrintQuote } from '@/components/servicejob/PrintQuote'
import { Printer } from 'lucide-react'
import { useState, useEffect, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { useServiceJob, useUpdateServiceJob, useCompleteServiceJob, useApproveQuote, useAddServiceJobItem, useRemoveServiceJobItem, useDeleteServiceJob } from '@/hooks/useServiceJobs'
import { useCreateInvoiceFromJob } from '@/hooks/useInvoices'
import { useProducts } from '@/hooks/useProducts'
import { useEmployees } from '@/hooks/useEmployees'
import { useSettings } from '@/hooks/useSettings'
import { distanceUnit, unitLabel } from '@/utils/units'
import { StatusBadge } from '@/components/servicejob/StatusBadge'
import { PriorityBadge } from '@/components/servicejob/PriorityBadge'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { LABOR_PRESETS, FEE_PRESETS, parseArraySetting } from '@/lib/packages'
import type { Preset } from '@/lib/packages'
import type { AddItemRequest } from '@/types/servicejob'

type ItemType = 'product' | 'labor' | 'fee' | 'custom'

export function ServiceJobDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const jobId = parseInt(id || '0')

  const { data: job, isLoading } = useServiceJob(jobId)
  const { data: productsData } = useProducts({ per_page: 100 })
  const { data: employees } = useEmployees()
  const { data: settingsData } = useSettings()
  const unit = unitLabel(distanceUnit(settingsData))
  const updateMutation = useUpdateServiceJob()
  const completeMutation = useCompleteServiceJob()
  const approveQuoteMutation = useApproveQuote()
  const addItemMutation = useAddServiceJobItem()
  const removeItemMutation = useRemoveServiceJobItem()
  const deleteMutation = useDeleteServiceJob()
  const createInvoiceMutation = useCreateInvoiceFromJob()

  const [showAddItem, setShowAddItem] = useState(false)
  const [itemType, setItemType] = useState<ItemType>('product')
  const [selectedProductId, setSelectedProductId] = useState('')
  const [productFilter, setProductFilter] = useState('')
  const [itemDesc, setItemDesc] = useState('')
  const [itemQty, setItemQty] = useState('1')
  const [itemPrice, setItemPrice] = useState('')
  const [diagnosis, setDiagnosis] = useState('')
  const [workPerformed, setWorkPerformed] = useState('')
  const [mileage, setMileage] = useState('')

  const products = productsData?.data || []
  // Same owner-editable catalogs the POS uses, so a job and a sale draw from one list.
  const laborPresets = useMemo(() => parseArraySetting<Preset>(settingsData?.labor_presets, LABOR_PRESETS), [settingsData?.labor_presets])
  const feePresets = useMemo(() => parseArraySetting<Preset>(settingsData?.fee_presets, FEE_PRESETS), [settingsData?.fee_presets])

  useEffect(() => {
    if (job) {
      setDiagnosis(job.diagnosis || '')
      setWorkPerformed(job.work_performed || '')
      setMileage(job.mileage != null ? String(job.mileage) : '')
    }
  }, [job])

  const saveMileage = () => {
    const n = parseInt(mileage, 10)
    const next = Number.isFinite(n) && n > 0 ? n : undefined
    if (next !== (job?.mileage ?? undefined)) {
      updateMutation.mutate({ id: jobId, data: { mileage: next } })
    }
  }

  const handleSaveNotes = () => {
    updateMutation.mutate({ id: jobId, data: { diagnosis, work_performed: workPerformed } })
  }

  const [confirmComplete, setConfirmComplete] = useState(false)

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!job) return <p className="text-sm text-destructive">Job not found</p>

  const resetItemForm = () => {
    setItemType('product'); setSelectedProductId(''); setProductFilter(''); setItemDesc(''); setItemQty('1'); setItemPrice('')
  }

  const handleAddItem = () => {
    const qty = parseFloat(itemQty) || 1
    const price = parseFloat(itemPrice)
    let data: AddItemRequest
    if (itemType === 'product') {
      const product = products.find(p => p.id === parseInt(selectedProductId))
      if (!product) return
      data = { product_id: product.id, item_type: 'product', quantity: qty, unit_price: isNaN(price) ? product.sell_price : price }
    } else {
      if (!itemDesc.trim()) return
      data = { item_type: itemType, description: itemDesc.trim(), quantity: qty, unit_price: isNaN(price) ? 0 : price }
    }
    addItemMutation.mutate({ jobId, data }, {
      onSuccess: () => { setShowAddItem(false); resetItemForm() },
    })
  }

  const addItemDisabled = itemType === 'product' ? !selectedProductId : !itemDesc.trim()

  const handleCompleteAndInvoice = () => {
    completeMutation.mutate(jobId, {
      onSuccess: () => {
        createInvoiceMutation.mutate(
          { jobId, data: { exchange_rate: settingsData?.exchange_rate_usd_khr || 4050, discount: job.discount || undefined } },
          { onSuccess: (data: any) => { if (data?.id) navigate(`/invoices/${data.id}`) } }
        )
      },
    })
  }

  const handleComplete = () => {
    completeMutation.mutate(jobId)
  }

  const handleRemoveItem = (itemId: number) => {
    removeItemMutation.mutate(itemId)
  }

  return (
    <div className="max-w-4xl space-y-6">
      <PrintQuote job={job} />
      <PageHeader
        title={job.job_number}
        backTo={-1}
        breadcrumb="Service Jobs"
        badges={
          <>
            <StatusBadge status={job.status} />
            <PriorityBadge priority={job.priority} />
            {job.quote_approved_at && (
              <span className="rounded-full px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary">
                Quote Approved
              </span>
            )}
          </>
        }
      />

      <div className="space-y-4 lg:space-y-0 lg:grid lg:grid-cols-3 lg:gap-4">
        {/* Main column */}
        <div className="space-y-4 lg:col-span-2">
          {/* Items */}
          <div className="bg-card rounded-lg p-5 shadow-sm">
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm font-semibold">Items</p>
              {job.status !== 'completed' && job.status !== 'cancelled' && (
                <Button size="sm" variant="outline" onClick={() => setShowAddItem(true)}>Add Item</Button>
              )}
            </div>
            {job.items.length === 0 ? (
              <p className="text-sm text-muted-foreground">No items added</p>
            ) : (
              <div className="space-y-0.5">
                {job.items.map((item) => (
                  <div key={item.id} className="flex items-center justify-between py-1 border-b last:border-0 text-sm">
                    <div className="min-w-0 flex-1">
                      <span className="font-medium">{item.product_name || item.description || item.item_type}</span>
                      {item.item_type !== 'product' && <span className="ml-1.5 rounded bg-muted px-1 py-0.5 text-[10px] uppercase tracking-wide text-muted-foreground">{item.item_type}</span>}
                      <span className="text-muted-foreground ml-2">x{item.quantity} @ ${item.unit_price.toFixed(2)}</span>
                    </div>
                    <div className="flex items-center gap-2 flex-shrink-0 ml-2">
                      <span className="tabular-nums">${item.total_price.toFixed(2)}</span>
                      {job.status !== 'completed' && job.status !== 'cancelled' && (
                        <Button variant="ghost" size="sm" onClick={() => handleRemoveItem(item.id)}>×</Button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
            <div className="flex items-baseline justify-end gap-2 mt-2 pt-2 border-t">
              <span className="text-sm text-muted-foreground">Total</span>
              <span className="text-2xl font-bold tabular-nums">${job.total_amount.toFixed(2)}</span>
            </div>
          </div>

          {/* Diagnosis + Work Performed */}
          {job.status !== 'completed' && job.status !== 'cancelled' && (
            <div className="bg-card rounded-lg p-5 shadow-sm space-y-3">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-sm font-semibold">Diagnosis</label>
                  <textarea value={diagnosis} onChange={(e) => setDiagnosis(e.target.value)}
                    className="w-full rounded border border-input bg-background px-2 py-1 text-sm min-h-[50px]" rows={2}
                    placeholder="Mechanic's findings..." />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-semibold">Work Performed</label>
                  <textarea value={workPerformed} onChange={(e) => setWorkPerformed(e.target.value)}
                    className="w-full rounded border border-input bg-background px-2 py-1 text-sm min-h-[50px]" rows={2}
                    placeholder="What was done..." />
                </div>
              </div>
              <div className="flex justify-end">
                <Button size="sm" onClick={handleSaveNotes} disabled={updateMutation.isPending}>
                  {updateMutation.isPending ? 'Saving...' : 'Save Notes'}
                </Button>
              </div>
            </div>
          )}
        </div>

        {/* Rail */}
        <div className="space-y-4 lg:col-span-1">
          <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
            <p className="text-sm font-semibold">Customer</p>
            <p className="text-sm font-medium">{job.customer_name || 'Walk-in'}</p>
            {job.customer_phone && <p className="text-sm text-muted-foreground">{job.customer_phone}</p>}
            {job.vehicle_info && <p className="text-sm text-muted-foreground">{job.vehicle_info} ({job.plate_number})</p>}
            {job.vehicle_id && (
              <div className="flex items-center gap-2 pt-1">
                <label className="text-xs text-muted-foreground whitespace-nowrap">Odometer ({unit})</label>
                {job.status === 'completed' || job.status === 'cancelled' ? (
                  <span className="text-sm">{job.mileage != null ? job.mileage.toLocaleString() : '—'}</span>
                ) : (
                  <Input
                    value={mileage}
                    onChange={(e) => setMileage(e.target.value.replace(/[^\d]/g, ''))}
                    onBlur={saveMileage}
                    inputMode="numeric"
                    placeholder="e.g. 85000"
                    className="h-8"
                  />
                )}
              </div>
            )}
          </div>

          <div className="bg-card rounded-lg p-5 shadow-sm space-y-2">
            <p className="text-sm font-semibold">Technician</p>
            {job.status === 'completed' || job.status === 'cancelled' ? (
              <p className="text-sm">{job.assigned_to_name || 'Unassigned'}</p>
            ) : (
              <Select value={job.assigned_to_id || ''}
                onChange={(e) => e.target.value && updateMutation.mutate({ id: jobId, data: { assigned_to: parseInt(e.target.value) } })}>
                <option value="">Unassigned</option>
                {(employees || []).map((emp) => <option key={emp.id} value={emp.id}>{emp.name}</option>)}
              </Select>
            )}
          </div>

          <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
            <p className="text-sm font-semibold">Description</p>
            <p className="text-sm">{job.description}</p>
            {job.scheduled_at && <p className="text-xs font-medium text-primary">📅 Scheduled: {new Date(job.scheduled_at).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' })}</p>}
            {job.estimated_hours != null && job.estimated_hours > 0 && <p className="text-xs text-muted-foreground">Est. {job.estimated_hours}h</p>}
          </div>

          <div className="space-y-3">
            {job.items.length > 0 && (
              <Button variant="outline" className="w-full gap-1.5" onClick={() => window.print()}>
                <Printer className="h-4 w-4" /> Print Quote
              </Button>
            )}

            {job.items.length > 0 && !job.quote_approved_at && job.status !== 'completed' && job.status !== 'cancelled' && (
              <Button className="w-full" onClick={() => approveQuoteMutation.mutate(jobId)} disabled={approveQuoteMutation.isPending}>
                {approveQuoteMutation.isPending ? 'Approving...' : 'Approve Quote'}
              </Button>
            )}

            {job.status === 'completed' && !job.invoice_id && (
              <Button className="w-full" onClick={() => createInvoiceMutation.mutate({ jobId, data: { exchange_rate: settingsData?.exchange_rate_usd_khr || 4050, discount: job.discount || undefined } },
                { onSuccess: (data: any) => { if (data?.id) navigate(`/invoices/${data.id}`) } }
              )} disabled={createInvoiceMutation.isPending}>
                {createInvoiceMutation.isPending ? 'Generating...' : 'Generate Invoice'}
              </Button>
            )}

            {job.invoice_id && (
              <Button variant="outline" className="w-full" onClick={() => navigate(`/invoices/${job.invoice_id}`)}>
                View Invoice
              </Button>
            )}

            {job.status !== 'completed' && job.status !== 'cancelled' && (
              <>
                {job.items.length > 0 && (
                  <Button className="w-full" onClick={handleCompleteAndInvoice} disabled={completeMutation.isPending || createInvoiceMutation.isPending}>
                    {(completeMutation.isPending || createInvoiceMutation.isPending) ? 'Working…' : 'Complete & invoice'}
                  </Button>
                )}
                {confirmComplete ? (
                  <div className="flex items-center gap-2 bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">
                    <span className="text-sm text-destructive">Complete {job.job_number}? Items can no longer be changed.</span>
                    <Button size="sm" variant="outline" onClick={() => setConfirmComplete(false)}>Cancel</Button>
                    <Button size="sm" variant="destructive" onClick={() => { handleComplete(); setConfirmComplete(false) }} disabled={completeMutation.isPending}>
                      {completeMutation.isPending ? 'Completing...' : 'Confirm'}
                    </Button>
                  </div>
                ) : (
                  <Button variant="outline" className="w-full" onClick={() => setConfirmComplete(true)} disabled={completeMutation.isPending}>
                    {completeMutation.isPending ? 'Completing...' : 'Mark complete only'}
                  </Button>
                )}
              </>
            )}
          </div>
        </div>
      </div>

      {showAddItem && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-10 bg-black/50">
          <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-lg mx-4">
            <h2 className="text-sm font-semibold mb-3">Add Item</h2>
            <div className="space-y-3">
              {/* Line type — a job holds the same line types as a sale */}
              <div className="flex gap-1 rounded-md border p-0.5">
                {(['product', 'labor', 'fee', 'custom'] as ItemType[]).map((t) => (
                  <button key={t} onClick={() => { setItemType(t); setItemPrice(''); setItemDesc(''); setSelectedProductId('') }}
                    className={`flex-1 rounded px-2 py-1.5 text-xs font-medium capitalize transition-colors ${itemType === t ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}>
                    {t}
                  </button>
                ))}
              </div>

              {itemType === 'product' ? (
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Product</label>
                  <Input placeholder="Filter by name or tire size..." value={productFilter} onChange={(e) => setProductFilter(e.target.value)} className="mb-1" />
                  <Select value={selectedProductId} onChange={(e) => {
                    setSelectedProductId(e.target.value)
                    const p = products.find(p => p.id === parseInt(e.target.value))
                    if (p) setItemPrice(p.sell_price.toString())
                  }}>
                    <option value="">Select...</option>
                    {(productFilter
                      ? products.filter(p =>
                          p.name.toLowerCase().includes(productFilter.toLowerCase()) ||
                          (p.tire_size && p.tire_size.toLowerCase().includes(productFilter.toLowerCase()))
                        )
                      : products
                    ).map((p) => (
                      <option key={p.id} value={p.id}>{p.name} - ${p.sell_price.toFixed(2)} ({p.stock_quantity} in stock)</option>
                    ))}
                  </Select>
                </div>
              ) : (
                <div className="space-y-1.5">
                  <label className="text-sm font-medium capitalize">{itemType}</label>
                  {(itemType === 'labor' || itemType === 'fee') && (
                    <div className="flex flex-wrap gap-1.5">
                      {(itemType === 'labor' ? laborPresets : feePresets).map((p) => (
                        <button key={p.description} type="button"
                          onClick={() => { setItemDesc(p.description); setItemPrice(String(p.unit_price_usd)) }}
                          className={`rounded-full border px-2.5 py-1 text-xs transition-colors ${itemDesc === p.description ? 'border-primary bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}>
                          {p.description} · ${p.unit_price_usd}
                        </button>
                      ))}
                    </div>
                  )}
                  <Input value={itemDesc} onChange={(e) => setItemDesc(e.target.value)}
                    placeholder={itemType === 'labor' ? 'e.g. Wheel alignment' : itemType === 'fee' ? 'e.g. Disposal fee' : 'e.g. Diagnostic time'} />
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Quantity</label>
                  <Input value={itemQty} onChange={(e) => setItemQty(e.target.value)} type="number" step="0.5" min="0.5" />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Unit Price ($)</label>
                  <Input value={itemPrice} onChange={(e) => setItemPrice(e.target.value)} type="number" step="0.01" min="0" />
                </div>
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <Button variant="ghost" onClick={() => { setShowAddItem(false); resetItemForm() }}>Cancel</Button>
                <Button onClick={handleAddItem} disabled={addItemDisabled || addItemMutation.isPending}>
                  {addItemMutation.isPending ? 'Adding...' : 'Add'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
