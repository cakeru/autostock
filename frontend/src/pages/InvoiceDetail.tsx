import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { PageHeader } from '@/components/layout/PageHeader'
import { formatDateTime } from '@/utils/date'
import { useState, useEffect, useMemo, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Printer, Receipt, Undo2, Send, Pencil, Trash2, X, FileImage } from 'lucide-react'
import { useSendToTelegram } from '@/hooks/useSendToTelegram'
import { useInvoice, useUpdateInvoice, useVoidInvoice, useRecordPayment, useUpdatePayment, useDeletePayment, useUploadPaymentProof, useUpdateInvoiceItem, useAddInvoiceItem, useRemoveInvoiceItem } from '@/hooks/useInvoices'
import { useCustomers, useCustomerVehicles } from '@/hooks/useCustomers'
import { useProducts } from '@/hooks/useProducts'
import { useInvoiceReturns, useUndoReturn } from '@/hooks/useReturns'
import { ReturnDialog } from '@/components/invoice/ReturnDialog'
import { useSettings } from '@/hooks/useSettings'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { InvoiceStatusBadge } from '@/components/invoice/StatusBadge'
import { PrintReceipt, type PrintFormat } from '@/components/invoice/PrintReceipt'
import { imageSrc } from '@/utils/imageUrl'
import type { RecordPaymentRequest, UpdateInvoiceRequest, UpdateInvoiceItemRequest, InvoiceDetail, Payment } from '@/types/invoice'
import type { Product } from '@/types/product'
import { parseArraySetting, DEFAULT_PAYMENT_METHODS } from '@/lib/packages'

export function InvoiceDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const invoiceId = parseInt(id || '0')

  const { data: invoice, isLoading } = useInvoice(invoiceId)
  const { data: settings } = useSettings()
  const paymentMethods = useMemo(() => parseArraySetting<string>(settings?.payment_methods, DEFAULT_PAYMENT_METHODS), [settings?.payment_methods])
  const updateMutation = useUpdateInvoice()
  const voidMutation = useVoidInvoice()
  const recordPaymentMutation = useRecordPayment()
  const updatePaymentMutation = useUpdatePayment()
  const deletePaymentMutation = useDeletePayment()
  const uploadProofMutation = useUploadPaymentProof()
  const updateItemMutation = useUpdateInvoiceItem()
  const addItemMutation = useAddInvoiceItem()
  const removeItemMutation = useRemoveInvoiceItem()

  const [voidReason, setVoidReason] = useState('')
  const [showVoid, setShowVoid] = useState(false)
  const [showPayment, setShowPayment] = useState(false)
  const [editingPayment, setEditingPayment] = useState<Payment | null>(null)
  const [showReturn, setShowReturn] = useState(false)
  const [showVehicleEdit, setShowVehicleEdit] = useState(false)
  const [showItemsEdit, setShowItemsEdit] = useState(false)
  const [lightbox, setLightbox] = useState('')
  const { data: returnsData } = useInvoiceReturns(invoiceId)
  const undoReturn = useUndoReturn()
  const [printFormat, setPrintFormat] = useState<PrintFormat | null>(null)
  const captureRef = useRef<HTMLDivElement>(null)
  const { send: sendTelegram, sending } = useSendToTelegram()

  const handleSendTelegram = () => {
    if (!invoice || !captureRef.current) return
    const cap = `🧾 Invoice ${invoice.invoice_number} · $${invoice.total_usd.toLocaleString(undefined, { minimumFractionDigits: 2 })}${invoice.customer_name ? ` · ${invoice.customer_name}` : ''}`
    sendTelegram(captureRef.current, `${invoice.invoice_number}.pdf`, cap)
  }

  // Mount the chosen printable, then trigger the browser print dialog once it's in the DOM.
  useEffect(() => {
    if (!printFormat) return
    const reset = () => setPrintFormat(null)
    window.addEventListener('afterprint', reset)
    window.print()
    return () => window.removeEventListener('afterprint', reset)
  }, [printFormat])

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>
  if (!invoice) return <p className="text-sm text-destructive">Invoice not found</p>

  // Pay the outstanding balance, not the full total — the invoice may already
  // be partly paid (e.g. a deposit was applied).
  const owed = Math.round((invoice.total_usd - (invoice.paid_amount || 0)) * 100) / 100
  const handleMarkPaidCash = () => {
    recordPaymentMutation.mutate(
      { id: invoiceId, data: { amount: owed, method: 'cash' } },
      { onSuccess: () => setShowPayment(false) }
    )
  }

  const handleRecordPayment = (data: RecordPaymentRequest, proof?: File) => {
    recordPaymentMutation.mutate(
      { id: invoiceId, data },
      {
        onSuccess: (pay) => {
          if (proof) uploadProofMutation.mutate({ invoiceId, paymentId: pay.id, file: proof })
          setShowPayment(false)
        },
      }
    )
  }

  const handleUpdatePayment = (data: RecordPaymentRequest) => {
    if (!editingPayment) return
    updatePaymentMutation.mutate(
      { id: invoiceId, paymentId: editingPayment.id, data },
      { onSuccess: () => { setEditingPayment(null); setShowPayment(false) } }
    )
  }

  const handleDeletePayment = (p: Payment) => {
    if (!window.confirm(`Delete the $${p.amount.toFixed(2)} ${p.method} payment? The invoice's paid status will be recalculated.`)) return
    deletePaymentMutation.mutate({ id: invoiceId, paymentId: p.id })
  }

  const handleUndoReturn = (r: { id: number; refund_amount: number; refund_method: string }) => {
    if (!window.confirm(`Undo this ${r.refund_method === 'store_credit' ? 'store credit' : 'refund'} of $${r.refund_amount.toFixed(2)}? The restocked items go back out of inventory and the credit is removed.`)) return
    undoReturn.mutate({ id: r.id, invoiceId })
  }

  const handleVoid = () => {
    if (!voidReason) return
    voidMutation.mutate({ id: invoiceId, reason: voidReason }, { onSuccess: () => setShowVoid(false) })
  }

  const payments = invoice.payments || []
  const totalPaid = payments.reduce((sum, p) => sum + p.amount, 0)

  return (
    <>
    {printFormat && <PrintReceipt invoice={invoice} settings={settings} format={printFormat} />}
    {/* Off-screen branded invoice, rendered on white so it can be captured to
        PDF for "Send to Telegram" without affecting the visible layout. */}
    <div aria-hidden className="pointer-events-none fixed left-[-10000px] top-0 w-[210mm] bg-white p-[10mm]">
      <div ref={captureRef}>
        <PrintReceipt invoice={invoice} settings={settings} format="a4" capture />
      </div>
    </div>
    <div className="max-w-4xl space-y-6 print:hidden">
      <PageHeader
        title={invoice.invoice_number}
        backTo={-1}
        breadcrumb="Invoices"
        badges={
          <>
            <InvoiceStatusBadge status={invoice.status} />
            {invoice.payment_status && invoice.payment_status !== invoice.status && (
              <span className="text-xs text-muted-foreground">
                Payment: <InvoiceStatusBadge payment_status={invoice.payment_status} />
              </span>
            )}
          </>
        }
        actions={
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setPrintFormat('thermal')} className="gap-1.5">
              <Receipt className="h-4 w-4" /> Receipt 80mm
            </Button>
            <Button variant="outline" size="sm" onClick={() => setPrintFormat('classic')} className="gap-1.5">
              <Printer className="h-4 w-4" /> Invoice (Khmer)
            </Button>
            <Button variant="outline" size="sm" onClick={() => setPrintFormat('a4')} className="gap-1.5">
              <Printer className="h-4 w-4" /> Invoice (Branded)
            </Button>
            <Button variant="outline" size="sm" onClick={handleSendTelegram} disabled={sending} className="gap-1.5">
              <Send className="h-4 w-4" /> {sending ? 'Sending…' : 'Send to Telegram'}
            </Button>
          </div>
        }
      />

      <div className="space-y-4 lg:space-y-0 lg:grid lg:grid-cols-3 lg:gap-4">
        {/* Main column */}
        <div className="space-y-4 lg:col-span-2">
          <div className="bg-card rounded-lg p-5 shadow-sm">
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm font-semibold">Items</p>
              {invoice.status !== 'voided' && (
                <Button size="sm" variant="ghost" onClick={() => setShowItemsEdit(true)} className="h-7 gap-1 px-2 text-xs -mr-1">
                  <Pencil className="h-3 w-3" /> Edit items
                </Button>
              )}
            </div>
            {invoice.items.length === 0 ? (
              <p className="text-sm text-muted-foreground">No items</p>
            ) : (
              <div className="space-y-1">
                {invoice.items.map((item) => (
                  <div key={item.id} className="flex items-center justify-between py-1 text-sm border-b last:border-0">
                    <div className="min-w-0 flex-1">
                      <span className="font-medium">{item.description}</span>
                      <span className="text-muted-foreground ml-2 text-xs">x{item.quantity} @ ${item.unit_price_usd.toFixed(2)}</span>
                    </div>
                    <span className="tabular-nums flex-shrink-0 ml-2">${item.total_usd.toFixed(2)}</span>
                  </div>
                ))}
              </div>
            )}
            <div className="flex justify-end mt-2 pt-2 border-t">
              <div className="text-right space-y-1">
                <p className="text-sm text-muted-foreground">Subtotal: ${invoice.subtotal.toFixed(2)}</p>
                {invoice.discount > 0 && <p className="text-sm text-muted-foreground">Discount: -${invoice.discount.toFixed(2)}</p>}
                {invoice.tax_amount > 0 && <p className="text-sm text-muted-foreground">Tax ({invoice.tax_rate}%): ${invoice.tax_amount.toFixed(2)}</p>}
                <p className="text-2xl font-bold">${invoice.total_usd.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</p>
                <p className="text-xs text-muted-foreground">KHR: {Math.round(invoice.total_khr).toLocaleString()}៛</p>
              </div>
            </div>
          </div>

          {invoice.status !== 'voided' && (
            <div className="bg-card rounded-lg p-5 shadow-sm space-y-3">
              <p className="text-sm font-semibold">Payments</p>
              {payments.length > 0 && (
                <div className="text-sm space-y-1">
                  {payments.map((p) => (
                    <div key={p.id} className="flex justify-between py-1 border-b last:border-0">
                      <div className="min-w-0">
                        <span className="font-medium">${p.amount.toFixed(2)}</span>
                        <span className="text-muted-foreground ml-1 text-xs">{p.method}</span>
                        {p.currency === 'KHR' && p.tendered_amount ? <span className="text-muted-foreground ml-1 text-xs">(៛{Math.round(p.tendered_amount).toLocaleString()})</span> : null}
                        {p.received_by_name && <span className="text-muted-foreground ml-1 text-xs">by {p.received_by_name}</span>}
                        <div className="mt-0.5 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                          {p.reference && <span className="font-mono">{p.reference}</span>}
                          {p.proof_url && (
                            <button onClick={() => setLightbox(p.proof_url!)} className="flex items-center gap-1 text-primary hover:underline">
                              <FileImage className="h-3 w-3" /> Proof
                            </button>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        <span className="text-xs text-muted-foreground whitespace-nowrap">{formatDateTime(p.created_at)}</span>
                        <button onClick={() => { setEditingPayment(p); setShowPayment(true) }} title="Edit payment" className="text-muted-foreground hover:text-foreground">
                          <Pencil className="h-3.5 w-3.5" />
                        </button>
                        <button onClick={() => handleDeletePayment(p)} title="Delete payment" className="text-muted-foreground hover:text-destructive">
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <p className="text-xs text-muted-foreground">Paid: ${totalPaid.toFixed(2)} / ${invoice.total_usd.toFixed(2)}</p>
              {invoice.payment_status !== 'paid' && (
                <div className="flex gap-2">
                  <Button size="sm" className="flex-1" onClick={handleMarkPaidCash} disabled={recordPaymentMutation.isPending}>
                    {recordPaymentMutation.isPending ? 'Recording...' : `Pay ${owed < invoice.total_usd - 0.005 ? 'balance' : 'full'} — Cash ($${owed.toFixed(2)})`}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => setShowPayment(true)}>
                    Record Payment
                  </Button>
                </div>
              )}
            </div>
          )}

          {returnsData && returnsData.returns.length > 0 && (
            <div className="bg-card rounded-lg p-5 shadow-sm">
              <p className="text-sm font-semibold mb-2">Returns</p>
              <div className="space-y-2">
                {returnsData.returns.map((r) => (
                  <div key={r.id} className="rounded-md border p-2.5 text-sm">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-medium tabular-nums text-destructive">−${r.refund_amount.toFixed(2)}</span>
                      <span className="flex items-center gap-2 text-xs text-muted-foreground">
                        {r.refund_method === 'store_credit' ? 'Store credit' : 'Cash refund'} · {formatDateTime(r.created_at)}
                        <button onClick={() => handleUndoReturn(r)} title="Undo return" className="text-muted-foreground hover:text-destructive">
                          <Undo2 className="h-3.5 w-3.5" />
                        </button>
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground">{r.items.map((it) => `${it.quantity}× ${it.description}`).join(', ')}</p>
                    {r.reason && <p className="text-xs text-muted-foreground">Reason: {r.reason}</p>}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Rail */}
        <div className="space-y-4 lg:col-span-1">
          <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
            <div className="flex items-center justify-between">
              <p className="text-sm font-semibold">Customer</p>
              <Button size="sm" variant="ghost" onClick={() => setShowVehicleEdit(true)} className="h-7 gap-1 px-2 text-xs -mr-1">
                <Pencil className="h-3 w-3" /> Edit
              </Button>
            </div>
            <p className="text-sm font-medium">{invoice.customer_name || 'Walk-in'}</p>
            {invoice.customer_phone && <p className="text-sm">{invoice.customer_phone}</p>}
            {invoice.plate_number && (
              invoice.vehicle_id ? (
                <p
                  className="text-sm cursor-pointer text-primary hover:underline"
                  onClick={() => navigate(`/vehicles/${invoice.vehicle_id}`)}
                >
                  {invoice.vehicle_info ? `${invoice.vehicle_info} · ` : ''}{invoice.plate_number}
                </p>
              ) : (
                <p className="text-sm text-muted-foreground">
                  {invoice.vehicle_info ? `${invoice.vehicle_info} · ` : ''}{invoice.plate_number}
                </p>
              )
            )}
            {invoice.mileage != null && <p className="text-sm text-muted-foreground">Odometer: {invoice.mileage.toLocaleString()} {invoice.mileage_unit === 'mi' ? 'mi' : 'km'}</p>}
            {invoice.job_number && <p className="text-sm text-muted-foreground">Job: {invoice.job_number}</p>}
          </div>

          <div className="bg-card rounded-lg p-5 shadow-sm space-y-1.5">
            <p className="text-sm font-semibold">Notes</p>
            <p className="text-sm">{invoice.notes || '-'}</p>
          </div>

          {invoice.status !== 'voided' && (
            <div className="bg-card rounded-lg p-5 shadow-sm space-y-2">
              <p className="text-sm font-semibold">Actions</p>
              <Button size="sm" variant="outline" className="w-full gap-1.5" onClick={() => setShowReturn(true)}>
                <Undo2 className="h-4 w-4" /> Return / refund
              </Button>
              <Button size="sm" variant="destructive" className="w-full" onClick={() => setShowVoid(true)}>
                Void Invoice
              </Button>
            </div>
          )}
        </div>
      </div>

      {showVoid && (
        <ConfirmDialog
          open={showVoid}
          onClose={() => setShowVoid(false)}
          onConfirm={handleVoid}
          title={`Void ${invoice.invoice_number}?`}
          message="This will restore stock and cannot be undone."
          destructive
          confirmLabel={voidMutation.isPending ? 'Voiding...' : 'Void Invoice'}
          loading={voidMutation.isPending}
        >
          <div className="space-y-1.5 mb-3">
            <label className="text-sm font-medium">Reason *</label>
            <textarea value={voidReason} onChange={(e) => setVoidReason(e.target.value)}
              className="w-full rounded border border-input bg-background px-3 py-2 text-sm min-h-[50px]" />
          </div>
        </ConfirmDialog>
      )}

      {(showPayment || editingPayment) && (
        <RecordPaymentDialog
          invoiceId={invoiceId}
          invoiceNumber={invoice.invoice_number}
          owed={owed}
          rate={settings?.exchange_rate_usd_khr || 4050}
          methods={paymentMethods}
          payment={editingPayment || undefined}
          onClose={() => { setShowPayment(false); setEditingPayment(null) }}
          onConfirm={editingPayment ? handleUpdatePayment : handleRecordPayment}
          loading={recordPaymentMutation.isPending || uploadProofMutation.isPending}
        />
      )}

      {showItemsEdit && (
        <ItemEditorDialog
          invoice={invoice}
          onClose={() => setShowItemsEdit(false)}
          onUpdateItem={(itemId, data) => updateItemMutation.mutate({ id: invoiceId, itemId, data })}
          onAddItem={(data) => addItemMutation.mutate({ id: invoiceId, data })}
          onRemoveItem={(itemId) => removeItemMutation.mutate({ id: invoiceId, itemId })}
          pending={updateItemMutation.isPending || addItemMutation.isPending || removeItemMutation.isPending}
        />
      )}

      {lightbox && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4" onClick={() => setLightbox('')}>
          <img src={imageSrc(lightbox)} alt="Payment proof" className="max-h-[85vh] max-w-full rounded-lg" />
          <button className="absolute right-4 top-4 text-white/80 hover:text-white" onClick={() => setLightbox('')} aria-label="Close">
            <X className="h-6 w-6" />
          </button>
        </div>
      )}

      {showVehicleEdit && (
        <EditVehicleDialog
          invoice={invoice}
          unitLabel={invoice.mileage_unit === 'mi' ? 'mi' : 'km'}
          onClose={() => setShowVehicleEdit(false)}
          onSave={(data) => updateMutation.mutate(
            { id: invoiceId, data },
            { onSuccess: () => setShowVehicleEdit(false) }
          )}
          loading={updateMutation.isPending}
        />
      )}

      {showReturn && (
        <ReturnDialog
          invoice={invoice}
          returnedByItem={returnsData?.returned_by_item || {}}
          onClose={() => setShowReturn(false)}
        />
      )}
    </div>
    </>
  )
}

function EditVehicleDialog({ invoice, unitLabel, onClose, onSave, loading }: {
  invoice: InvoiceDetail
  unitLabel: string
  onClose: () => void
  onSave: (data: UpdateInvoiceRequest) => void
  loading: boolean
}) {
  const { data: customersData } = useCustomers({ per_page: 100 })
  const customers = customersData?.data || []
  const [customerId, setCustomerId] = useState(invoice.customer_id ? String(invoice.customer_id) : '')
  const { data: vehicles } = useCustomerVehicles(customerId ? parseInt(customerId) : 0)
  const [vehicleId, setVehicleId] = useState(invoice.vehicle_id ? String(invoice.vehicle_id) : '')
  const [mileage, setMileage] = useState(invoice.mileage != null ? String(invoice.mileage) : '')

  const submit = () => {
    onSave({
      customer_id: customerId ? parseInt(customerId) : 0,
      vehicle_id: vehicleId ? parseInt(vehicleId) : undefined,
      clear_vehicle: !vehicleId,
      mileage: mileage ? parseInt(mileage) : undefined,
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-sm mx-4">
        <p className="text-sm font-semibold mb-1">Edit customer &amp; vehicle</p>
        <p className="text-xs text-muted-foreground mb-3">{invoice.invoice_number}</p>
        <div className="space-y-2">
          <div className="space-y-1">
            <label className="text-xs font-medium">Customer</label>
            <Select
              value={customerId}
              onChange={(e) => {
                setCustomerId(e.target.value)
                setVehicleId('') // the current vehicle belongs to the old customer
              }}
            >
              <option value="">Walk-in (no customer)</option>
              {customers.map((c) => (
                <option key={c.id} value={c.id}>{c.name}{c.phone ? ` · ${c.phone}` : ''}</option>
              ))}
            </Select>
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">Vehicle</label>
            <Select value={vehicleId} onChange={(e) => setVehicleId(e.target.value)} disabled={!customerId}>
              <option value="">No vehicle</option>
              {(vehicles || []).map((v) => (
                <option key={v.id} value={v.id}>{v.plate_number}{v.make || v.model ? ` — ${[v.make, v.model].filter(Boolean).join(' ')}` : ''}</option>
              ))}
            </Select>
            {(!vehicles || vehicles.length === 0) && (
              <p className="text-xs text-muted-foreground">No vehicles registered for this customer.</p>
            )}
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">Odometer ({unitLabel})</label>
            <Input value={mileage} onChange={(e) => setMileage(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="e.g. 85000" />
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} disabled={loading}>{loading ? 'Saving...' : 'Save'}</Button>
        </div>
      </div>
    </div>
  )
}

function RecordPaymentDialog({ invoiceId, invoiceNumber, owed, rate, methods, onClose, onConfirm, loading, payment }: {
  invoiceId: number
  invoiceNumber: string
  owed: number
  rate: number
  methods: string[]
  onClose: () => void
  onConfirm: (data: RecordPaymentRequest, proof?: File) => void
  loading: boolean
  payment?: Payment
}) {
  const editing = !!payment
  const [currency, setCurrency] = useState<'USD' | 'KHR'>(payment?.currency === 'KHR' ? 'KHR' : 'USD')
  const [amount, setAmount] = useState(payment && payment.currency !== 'KHR' ? payment.amount.toFixed(2) : owed > 0 ? owed.toFixed(2) : '') // USD entry
  const [riel, setRiel] = useState(payment && payment.currency === 'KHR' ? String(Math.round(payment.tendered_amount || payment.amount * rate)) : owed > 0 ? String(Math.round(owed * rate)) : '') // KHR entry
  const [method, setMethod] = useState(payment?.method || methods[0] || 'cash')
  const [notes, setNotes] = useState(payment?.notes || '')
  const [reference, setReference] = useState(payment?.reference || '')
  const [proof, setProof] = useState<File | null>(null)
  const proofRef = useRef<HTMLInputElement>(null)
  const isCash = method === 'cash'

  // Convert whatever was tendered to USD; record at most the balance owed.
  const tenderedUSD = currency === 'USD' ? (parseFloat(amount) || 0) : (parseFloat(riel) || 0) / rate
  const recordAmount = editing ? Math.round(tenderedUSD * 100) / 100 : Math.min(Math.round(tenderedUSD * 100) / 100, owed)
  const changeUSD = editing ? 0 : Math.max(0, Math.round((tenderedUSD - owed) * 100) / 100)
  const changeRiel = Math.round(changeUSD * rate)
  const over = !editing && !isCash && tenderedUSD > owed + 0.005 // only cash can be tendered over

  const submit = () => {
    if (recordAmount <= 0) return
    onConfirm({
      amount: recordAmount,
      method: method || undefined,
      notes: notes || undefined,
      currency,
      tendered_amount: currency === 'USD' ? (parseFloat(amount) || 0) : (parseFloat(riel) || 0),
      exchange_rate: rate,
      reference: reference.trim() || undefined,
    }, proof || undefined)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-sm mx-4">
        <p className="text-sm font-semibold mb-1">{editing ? 'Edit Payment' : 'Record Payment'} — {invoiceNumber}</p>
        {editing ? (
          <p className="text-xs text-muted-foreground mb-3">Changing this payment recalculates the invoice's paid status.</p>
        ) : (
          <p className="text-xs text-muted-foreground mb-3">Balance owed: <span className="font-medium tabular-nums">${owed.toFixed(2)}</span> · ≈ ៛{Math.round(owed * rate).toLocaleString()}</p>
        )}

        <div className="mb-2 flex gap-1 rounded-md border p-0.5">
          {(['USD', 'KHR'] as const).map((c) => (
            <button key={c} onClick={() => setCurrency(c)}
              className={`flex-1 rounded px-2 py-1.5 text-xs font-medium transition-colors ${currency === c ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}>
              {c === 'USD' ? 'US Dollar ($)' : 'Riel (៛)'}
            </button>
          ))}
        </div>

        <div className="space-y-2">
          <div className="space-y-1">
            {currency === 'USD' ? (
              <>
                <label className="text-xs font-medium">{isCash ? 'Cash received' : 'Amount'} ($) *</label>
                <Input value={amount} onChange={(e) => setAmount(e.target.value)} type="number" step="0.01" min="0" />
              </>
            ) : (
              <>
                <label className="text-xs font-medium">Riel received (៛) *</label>
                <Input value={riel} onChange={(e) => setRiel(e.target.value)} type="number" step="100" min="0" />
                <p className="text-xs text-muted-foreground">= ${tenderedUSD.toFixed(2)} at ៛{rate.toLocaleString()}/$</p>
              </>
            )}
            {over && <p className="text-xs text-destructive">A {method} payment can't exceed the ${owed.toFixed(2)} balance.</p>}
            {isCash && changeUSD > 0 && <p className="text-xs text-muted-foreground">Change due: <span className="font-medium text-foreground">${changeUSD.toFixed(2)} · ៛{changeRiel.toLocaleString()}</span> — records ${owed.toFixed(2)}.</p>}
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">Method *</label>
            <Select value={method} onChange={(e) => setMethod(e.target.value)}>
              {methods.map((m) => (
                <option key={m} value={m} className="capitalize">{m}</option>
              ))}
            </Select>
          </div>
          {!editing && method !== 'cash' && (
            <>
              <div className="space-y-1">
                <label className="text-xs font-medium">Trx ID (ABA) — optional</label>
                <Input value={reference} onChange={(e) => setReference(e.target.value)} placeholder="e.g. TRX-8F2K1L9Q" />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-medium">Proof photo — optional</label>
                <div className="flex items-center gap-2">
                  <Button type="button" variant="outline" size="sm" onClick={() => proofRef.current?.click()}>
                    {proof ? proof.name.slice(0, 24) : 'Choose photo…'}
                  </Button>
                  {proof && (
                    <Button type="button" variant="ghost" size="sm" onClick={() => { setProof(null); if (proofRef.current) proofRef.current.value = '' }}>
                      Remove
                    </Button>
                  )}
                </div>
                <input ref={proofRef} type="file" accept="image/*" className="hidden" onChange={(e) => setProof(e.target.files?.[0] || null)} />
              </div>
            </>
          )}
          <div className="space-y-1">
            <label className="text-xs font-medium">Notes</label>
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="optional" />
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} disabled={loading || recordAmount <= 0 || over || !method}>
            {loading ? 'Saving...' : editing ? 'Save' : 'Record Payment'}
          </Button>
        </div>
      </div>
    </div>
  )
}

interface EditorLine {
  id?: number // present when the line already exists on the invoice
  product?: Product
  item_type: string
  description: string
  quantity: string
  unit_price_usd: string
}

function ItemEditorDialog({ invoice, onClose, onUpdateItem, onAddItem, onRemoveItem, pending }: {
  invoice: InvoiceDetail
  onClose: () => void
  onUpdateItem: (itemId: number, data: UpdateInvoiceItemRequest) => void
  onAddItem: (data: { product_id?: number; item_type: string; description: string; quantity: number; unit_price_usd: number }) => void
  onRemoveItem: (itemId: number) => void
  pending: boolean
}) {
  const { data: productsData } = useProducts({ per_page: 100 })
  const products = productsData?.data || []
  const [search, setSearch] = useState('')
  const [lines, setLines] = useState<EditorLine[]>(() =>
    invoice.items.map((it) => ({
      id: it.id,
      item_type: it.item_type,
      description: it.description,
      quantity: String(it.quantity),
      unit_price_usd: it.unit_price_usd.toFixed(2),
    }))
  )
  const [saved, setSaved] = useState(false)

  const setLine = (index: number, patch: Partial<EditorLine>) =>
    setLines((ls) => ls.map((l, i) => (i === index ? { ...l, ...patch } : l)))

  const addManualLine = () =>
    setLines((ls) => [...ls, { item_type: 'labor', description: '', quantity: '1', unit_price_usd: '' }])

  const pickProduct = (p: Product) => {
    const existing = lines.find((l) => l.product?.id === p.id)
    if (existing) return // dedupe; staff can bump qty on the existing line
    setLines((ls) => [...ls, {
      product: p,
      item_type: 'product',
      description: p.name,
      quantity: '1',
      unit_price_usd: String(p.sell_price || ''),
    }])
    setSearch('')
  }

  const save = () => {
    let ok = true
    lines.forEach((l) => {
      const qty = parseFloat(l.quantity)
      const price = parseFloat(l.unit_price_usd)
      if (isNaN(qty) || qty <= 0 || isNaN(price) || price < 0 || !l.description.trim()) {
        ok = false
        return
      }
    })
    if (!ok) return toast.error('Every line needs a description, quantity, and price')
    setSaved(true)
    lines.forEach((l) => {
      if (l.id != null) {
        const patch: UpdateInvoiceItemRequest = {}
        if (l.description.trim() !== (invoice.items.find((it) => it.id === l.id)?.description ?? '')) patch.description = l.description.trim()
        if (parseFloat(l.quantity) !== invoice.items.find((it) => it.id === l.id)?.quantity) patch.quantity = parseFloat(l.quantity)
        if (parseFloat(l.unit_price_usd) !== invoice.items.find((it) => it.id === l.id)?.unit_price_usd) patch.unit_price_usd = parseFloat(l.unit_price_usd)
        if (Object.keys(patch).length > 0) onUpdateItem(l.id, patch)
      } else {
        onAddItem({
          product_id: l.product?.id,
          item_type: l.item_type,
          description: l.description.trim(),
          quantity: parseFloat(l.quantity),
          unit_price_usd: parseFloat(l.unit_price_usd),
        })
      }
    })
    const removed = invoice.items.filter((it) => !lines.some((l) => l.id === it.id))
    removed.forEach((it) => onRemoveItem(it.id))
  }

  const filtered = search.trim()
    ? products.filter((p) => p.name.toLowerCase().includes(search.toLowerCase()) || p.sku.toLowerCase().includes(search.toLowerCase()))
    : products

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 p-0 sm:items-center sm:p-4">
      <div className="flex max-h-[92vh] w-full max-w-lg flex-col rounded-t-lg bg-card shadow-xl sm:rounded-lg">
        <div className="flex items-center justify-between border-b p-3">
          <p className="text-sm font-semibold">Edit items — {invoice.invoice_number}</p>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground"><X className="h-5 w-5" /></button>
        </div>

        {!saved && (
          <div className="border-b p-3">
            <div className="flex gap-2">
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Add from stock: search product or SKU…"
                className="h-9 flex-1 rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
              />
              <Button size="sm" variant="outline" onClick={addManualLine} className="shrink-0">+ Labor line</Button>
            </div>
            {search.trim() && (
              <div className="mt-1 max-h-40 overflow-y-auto rounded-md border">
                {filtered.length === 0 ? (
                  <p className="px-3 py-2 text-xs text-muted-foreground">No products match</p>
                ) : (
                  filtered.slice(0, 12).map((p) => (
                    <button key={p.id} onClick={() => pickProduct(p)}
                      className="flex w-full items-center justify-between px-3 py-1.5 text-left text-sm hover:bg-muted">
                      <span className="min-w-0 truncate">{p.name}</span>
                      <span className="ml-2 shrink-0 text-xs tabular-nums text-muted-foreground">
                        {p.stock_quantity} left · ${p.sell_price.toFixed(2)}
                      </span>
                    </button>
                  ))
                )}
              </div>
            )}
          </div>
        )}

        <div className="min-h-0 flex-1 overflow-y-auto p-3">
          {lines.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">No lines — add a product or a labor line.</p>
          ) : (
            <div className="space-y-2">
              {lines.map((l, i) => (
                <div key={l.id ?? `new-${i}`} className="rounded-md border p-2">
                  <div className="flex items-center gap-2">
                    <input
                      value={l.description}
                      onChange={(e) => setLine(i, { description: e.target.value })}
                      placeholder="Description"
                      className="h-8 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
                    />
                    <button onClick={() => setLines((ls) => ls.filter((_, x) => x !== i))}
                      className="shrink-0 text-muted-foreground hover:text-destructive" aria-label="Remove line">
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                  <div className="mt-1.5 flex items-center gap-2">
                    <label className="shrink-0 text-xs text-muted-foreground">Qty</label>
                    <input
                      value={l.quantity}
                      onChange={(e) => setLine(i, { quantity: e.target.value })}
                      type="number" min="0" step="any"
                      className="h-8 w-16 rounded-md border border-input bg-background px-2 text-sm tabular-nums focus:outline-none focus:ring-2 focus:ring-primary/30"
                    />
                    <label className="ml-2 shrink-0 text-xs text-muted-foreground">Price $</label>
                    <input
                      value={l.unit_price_usd}
                      onChange={(e) => setLine(i, { unit_price_usd: e.target.value })}
                      type="number" min="0" step="0.01"
                      className="h-8 w-24 rounded-md border border-input bg-background px-2 text-sm tabular-nums focus:outline-none focus:ring-2 focus:ring-primary/30"
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex justify-end gap-2 border-t p-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button disabled={pending} onClick={save}>{pending ? 'Saving…' : 'Save changes'}</Button>
        </div>
      </div>
    </div>
  )
}
