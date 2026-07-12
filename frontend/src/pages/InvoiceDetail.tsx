import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { PageHeader } from '@/components/layout/PageHeader'
import { formatDateTime } from '@/utils/date'
import { useState, useEffect, useMemo, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Printer, Receipt, Undo2, Send } from 'lucide-react'
import { useSendToTelegram } from '@/hooks/useSendToTelegram'
import { useInvoice, useUpdateInvoice, useVoidInvoice, useRecordPayment } from '@/hooks/useInvoices'
import { useInvoiceReturns } from '@/hooks/useReturns'
import { ReturnDialog } from '@/components/invoice/ReturnDialog'
import { useSettings } from '@/hooks/useSettings'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { InvoiceStatusBadge } from '@/components/invoice/StatusBadge'
import { PrintReceipt, type PrintFormat } from '@/components/invoice/PrintReceipt'
import type { RecordPaymentRequest } from '@/types/invoice'
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

  const [voidReason, setVoidReason] = useState('')
  const [showVoid, setShowVoid] = useState(false)
  const [showPayment, setShowPayment] = useState(false)
  const [showReturn, setShowReturn] = useState(false)
  const { data: returnsData } = useInvoiceReturns(invoiceId)
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

  const handleRecordPayment = (data: RecordPaymentRequest) => {
    recordPaymentMutation.mutate(
      { id: invoiceId, data },
      { onSuccess: () => setShowPayment(false) }
    )
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
            <p className="text-sm font-semibold mb-2">Items</p>
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
                      <div>
                        <span className="font-medium">${p.amount.toFixed(2)}</span>
                        <span className="text-muted-foreground ml-1 text-xs">{p.method}</span>
                        {p.currency === 'KHR' && p.tendered_amount ? <span className="text-muted-foreground ml-1 text-xs">(៛{Math.round(p.tendered_amount).toLocaleString()})</span> : null}
                        {p.received_by_name && <span className="text-muted-foreground ml-1 text-xs">by {p.received_by_name}</span>}
                      </div>
                      <span className="text-xs text-muted-foreground whitespace-nowrap">{formatDateTime(p.created_at)}</span>
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
                      <span className="text-xs text-muted-foreground">{r.refund_method === 'store_credit' ? 'Store credit' : 'Cash refund'} · {formatDateTime(r.created_at)}</span>
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
            <p className="text-sm font-semibold">Customer</p>
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
            {invoice.mileage != null && <p className="text-sm text-muted-foreground">Odometer: {invoice.mileage.toLocaleString()} km</p>}
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

      {showPayment && (
        <RecordPaymentDialog
          invoiceNumber={invoice.invoice_number}
          owed={owed}
          rate={settings?.exchange_rate_usd_khr || 4050}
          methods={paymentMethods}
          onClose={() => setShowPayment(false)}
          onConfirm={handleRecordPayment}
          loading={recordPaymentMutation.isPending}
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

function RecordPaymentDialog({ invoiceNumber, owed, rate, methods, onClose, onConfirm, loading }: {
  invoiceNumber: string
  owed: number
  rate: number
  methods: string[]
  onClose: () => void
  onConfirm: (data: RecordPaymentRequest) => void
  loading: boolean
}) {
  const [currency, setCurrency] = useState<'USD' | 'KHR'>('USD')
  const [amount, setAmount] = useState(owed > 0 ? owed.toFixed(2) : '') // USD entry
  const [riel, setRiel] = useState(owed > 0 ? String(Math.round(owed * rate)) : '') // KHR entry
  const [method, setMethod] = useState(methods[0] || 'cash')
  const [notes, setNotes] = useState('')
  const isCash = method === 'cash'

  // Convert whatever was tendered to USD; record at most the balance owed.
  const tenderedUSD = currency === 'USD' ? (parseFloat(amount) || 0) : (parseFloat(riel) || 0) / rate
  const recordAmount = Math.min(Math.round(tenderedUSD * 100) / 100, owed)
  const changeUSD = Math.max(0, Math.round((tenderedUSD - owed) * 100) / 100)
  const changeRiel = Math.round(changeUSD * rate)
  const over = !isCash && tenderedUSD > owed + 0.005 // only cash can be tendered over

  const submit = () => {
    if (recordAmount <= 0) return
    onConfirm({
      amount: recordAmount,
      method,
      notes: notes || undefined,
      currency,
      tendered_amount: currency === 'USD' ? (parseFloat(amount) || 0) : (parseFloat(riel) || 0),
      exchange_rate: rate,
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-sm mx-4">
        <p className="text-sm font-semibold mb-1">Record Payment — {invoiceNumber}</p>
        <p className="text-xs text-muted-foreground mb-3">Balance owed: <span className="font-medium tabular-nums">${owed.toFixed(2)}</span> · ≈ ៛{Math.round(owed * rate).toLocaleString()}</p>

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
          <div className="space-y-1">
            <label className="text-xs font-medium">Notes</label>
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="optional" />
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} disabled={loading || recordAmount <= 0 || over || !method}>
            {loading ? 'Recording...' : 'Record Payment'}
          </Button>
        </div>
      </div>
    </div>
  )
}
