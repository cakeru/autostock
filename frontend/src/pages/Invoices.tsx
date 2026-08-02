import { SlideOver } from '@/components/ui/SlideOver'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { PageHeader } from '@/components/layout/PageHeader'
import { Download } from 'lucide-react'
import { downloadFile } from '@/utils/download'
import { useState, useMemo, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'
import { useInvoices, useCreateInvoice } from '@/hooks/useInvoices'
import { useCustomers } from '@/hooks/useCustomers'
import { useProducts } from '@/hooks/useProducts'
import { useSettings } from '@/hooks/useSettings'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { InvoiceStatusBadge } from '@/components/invoice/StatusBadge'
import { formatDate } from '@/utils/date'
import { formatUSD } from '@/utils/currency'
import { TableCard, TableFooter, Th } from '@/components/ui/table'

export function Invoices() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [paymentFilter, setPaymentFilter] = useState(searchParams.get('payment_status') || '')
  const [showForm, setShowForm] = useState(false)

  const { data, isLoading } = useInvoices({
    status: statusFilter || undefined,
    payment_status: paymentFilter || undefined,
    page,
  })

  const invoices = data?.data || []
  const meta = data?.meta

  const { data: customersData } = useCustomers({ per_page: 100 })
  const { data: productsData } = useProducts({ per_page: 100 })
  const { data: settingsData } = useSettings()
  const customers = customersData?.data || []
  const products = productsData?.data || []

  const createMutation = useCreateInvoice()

  const [formCustomer, setFormCustomer] = useState('')
  const [productSearch, setProductSearch] = useState('')
  const [formItems, setFormItems] = useState<Array<{ type: string; productId: string; desc: string; qty: string; price: string }>>([
    { type: 'product', productId: '', desc: '', qty: '1', price: '' },
  ])
  const [formExchangeRate, setFormExchangeRate] = useState('4050')
  const [formDiscount, setFormDiscount] = useState('0')
  const [formNotes, setFormNotes] = useState('')

  const totalPreview = useMemo(() => {
    const subtotal = formItems.reduce((sum, item) => {
      const qty = parseFloat(item.qty) || 0
      const price = parseFloat(item.price) || (item.productId ? products.find(p => p.id === parseInt(item.productId))?.sell_price || 0 : 0)
      return sum + qty * price
    }, 0)
    const discount = parseFloat(formDiscount) || 0
    const rate = parseFloat(formExchangeRate) || 4050
    const totalUSD = subtotal - discount
    const totalKHR = totalUSD * rate
    return { subtotal, discount, totalUSD, totalKHR }
  }, [formItems, formDiscount, formExchangeRate, products])

  useEffect(() => {
    if (settingsData?.exchange_rate_usd_khr) {
      setFormExchangeRate(settingsData.exchange_rate_usd_khr.toString())
    }
  }, [settingsData])

  const handleAddItemRow = () => {
    setFormItems([...formItems, { type: 'product', productId: '', desc: '', qty: '1', price: '' }])
  }

  const handleRemoveItemRow = (idx: number) => {
    if (formItems.length <= 1) return
    setFormItems(formItems.filter((_, i) => i !== idx))
  }

  const handleCreate = () => {
    const items = formItems.map((f) => {
      const selectedProduct = products.find(p => p.id === parseInt(f.productId))
      return {
        product_id: f.productId ? parseInt(f.productId) : undefined,
        item_type: f.type,
        description: f.desc || selectedProduct?.name || '',
        quantity: parseFloat(f.qty) || 1,
        unit_price_usd: parseFloat(f.price) || selectedProduct?.sell_price || 0,
      }
    }).filter(i => i.quantity > 0)

    if (items.length === 0) {
      toast.error('Add at least one item with a valid quantity')
      return
    }

    createMutation.mutate({
      customer_id: formCustomer ? parseInt(formCustomer) : undefined,
      items,
      discount: parseFloat(formDiscount) || 0,
      exchange_rate: parseFloat(formExchangeRate) || 4050,
      notes: formNotes,
    }, { onSuccess: () => { setShowForm(false); setFormItems([{ type: 'product', productId: '', desc: '', qty: '1', price: '' }]) } })
  }

  const filters = (
    <>
      <Select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1); setSearchParams({}, { replace: true }) }} className="w-36">
        <option value="">All statuses</option>
        <option value="issued">Issued</option>
        <option value="paid">Paid</option>
        <option value="voided">Voided</option>
      </Select>
      <Select value={paymentFilter} onChange={(e) => { setPaymentFilter(e.target.value); setPage(1); setSearchParams({}, { replace: true }) }} className="w-36">
        <option value="">All payments</option>
        <option value="unpaid">Unpaid</option>
        <option value="partial">Partial</option>
        <option value="paid">Paid</option>
      </Select>
    </>
  )

  return (
    <div className="space-y-6">
      <PageHeader
        title="Invoices"
        actions={
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => downloadFile('/exports/invoices', 'invoices.csv')} className="gap-1.5">
              <Download className="h-3.5 w-3.5" /> Export CSV
            </Button>
            <Button onClick={() => setShowForm(true)} size="sm">New Invoice</Button>
          </div>
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
              toolbar={filters}
              footer={meta && (
                <TableFooter total={meta.total} page={meta.page} totalPages={meta.total_pages} onPage={setPage} noun="invoices" />
              )}
            >
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-card sticky top-0 z-10">
                    <Th>Customer</Th>
                    <Th>Status</Th>
                    <Th className="text-right">Total</Th>
                    <Th className="text-right">Paid</Th>
                    <Th className="text-right">Date</Th>
                  </tr>
                </thead>
                <tbody>
                  {invoices.length === 0 ? (
                    <tr><td colSpan={5} className="px-4 py-10 text-center text-muted-foreground">No invoices found</td></tr>
                  ) : (
                    invoices.map((inv) => (
                      <tr key={inv.id} className="border-b last:border-0 hover:bg-muted/50 transition-colors duration-100 cursor-pointer"
                        onClick={() => navigate(`/invoices/${inv.id}`)}>
                        <td className="px-4 py-2.5">
                          <p className={`font-medium ${inv.customer_name ? '' : 'text-muted-foreground'}`}>{inv.customer_name || 'Walk-in'}</p>
                          <p className="font-mono text-xs text-muted-foreground">
                            {inv.invoice_number}{inv.plate_number ? ` · ${inv.plate_number}` : ''}
                          </p>
                        </td>
                        <td className="px-4 py-2.5">
                          <InvoiceStatusBadge status={inv.status} payment_status={inv.payment_status} />
                        </td>
                        <td className="px-4 py-2.5 text-right tabular-nums">{formatUSD(inv.total_usd)}</td>
                        <td className="px-4 py-2.5 text-right tabular-nums">
                          {inv.paid_amount > 0 ? formatUSD(inv.paid_amount) : <span className="text-muted-foreground">—</span>}
                        </td>
                        <td className="px-4 py-2.5 text-right text-muted-foreground text-xs whitespace-nowrap">
                          {inv.issued_at ? formatDate(inv.issued_at) : '—'}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </TableCard>
          </div>

          <div className="md:hidden space-y-3">
            <div className="flex flex-wrap items-center gap-2.5">{filters}</div>
            {invoices.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-8">No invoices found</p>
            ) : (
              invoices.map((inv) => (
                <div key={inv.id} className="bg-card rounded-lg p-5 shadow-sm cursor-pointer"
                  onClick={() => navigate(`/invoices/${inv.id}`)}>
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-xs font-medium">{inv.invoice_number}</span>
                    <InvoiceStatusBadge status={inv.status} payment_status={inv.payment_status} />
                  </div>
                  <p className="text-sm mt-1">{inv.customer_name || 'Walk-in'}{inv.plate_number ? ` · ${inv.plate_number}` : ''}</p>
                  <div className="flex items-center justify-between mt-2 text-xs text-muted-foreground">
                    <span>{formatUSD(inv.total_usd)}{inv.paid_amount > 0 ? ` · Paid: ${formatUSD(inv.paid_amount)}` : ''}</span>
                    <span>{inv.issued_at ? formatDate(inv.issued_at) : '—'}</span>
                  </div>
                </div>
              ))
            )}
          </div>

          {meta && meta.total_pages > 1 && (
            <div className="md:hidden flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Page {meta.page} of {meta.total_pages}</span>
              <div className="flex gap-1">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
                <Button variant="outline" size="sm" disabled={page >= meta.total_pages} onClick={() => setPage(page + 1)}>Next</Button>
              </div>
            </div>
          )}
        </>
      )}

      {/* New Invoice Slide-over */}
      <SlideOver
        open={showForm}
        onClose={() => setShowForm(false)}
        title="New Invoice"
      >
        <div className="space-y-3">
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Customer</label>
            <Select value={formCustomer} onChange={(e) => setFormCustomer(e.target.value)}>
              <option value="">Walk-in</option>
              {customers.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </Select>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium">Items</label>
            <Input
              placeholder="Filter products..."
              value={productSearch}
              onChange={(e) => setProductSearch(e.target.value)}
              className="mb-1"
            />
              {formItems.map((item, idx) => {
                const filtered = productSearch
                  ? products.filter(p =>
                      p.name.toLowerCase().includes(productSearch.toLowerCase()) ||
                      (p.tire_size && p.tire_size.toLowerCase().includes(productSearch.toLowerCase()))
                    )
                  : products
                return (
                <div key={idx} className="flex gap-2 items-start mb-1">
                    <Select value={item.productId} onChange={(e) => {
                      const newItems = [...formItems]
                      newItems[idx].productId = e.target.value
                      const p = products.find(p => p.id === parseInt(e.target.value))
                      if (p) { newItems[idx].price = p.sell_price.toString(); newItems[idx].desc = p.name }
                      setFormItems(newItems)
                    }} className="flex-1">
                      <option value="">Select product</option>
                      {filtered.map((p) => <option key={p.id} value={p.id}>{p.name} - ${p.sell_price.toFixed(2)} ({p.stock_quantity} in stock)</option>)}
                    </Select>
                  <Input value={item.qty} onChange={(e) => {
                    const newItems = [...formItems]; newItems[idx].qty = e.target.value; setFormItems(newItems)
                  }} type="number" step="0.5" className="w-[60px]" />
                  <Input value={item.price} onChange={(e) => {
                    const newItems = [...formItems]; newItems[idx].price = e.target.value; setFormItems(newItems)
                  }} type="number" step="0.01" className="w-[80px]" placeholder="Price" />
                  <Button variant="outline" size="icon" onClick={() => handleRemoveItemRow(idx)} className="w-7 h-7 shrink-0" title="Remove item">&times;</Button>
                </div>
              )})}
            <Button variant="ghost" size="sm" onClick={handleAddItemRow}>+ Add item</Button>
          </div>

          <div className="bg-muted/30 border rounded-md p-4 space-y-1 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Subtotal</span>
              <span className="tabular-nums">{formatUSD(totalPreview.subtotal)}</span>
            </div>
            {totalPreview.discount > 0 && (
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Discount</span>
                <span className="tabular-nums text-destructive">-{formatUSD(totalPreview.discount)}</span>
              </div>
            )}
            <div className="flex items-center justify-between font-semibold border-t pt-1">
              <span>Total USD</span>
              <span className="tabular-nums text-primary">{formatUSD(totalPreview.totalUSD)}</span>
            </div>
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>Total KHR</span>
              <span className="tabular-nums">{Math.round(totalPreview.totalKHR).toLocaleString()}៛</span>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-sm font-medium">Exchange Rate</label>
              <Input value={formExchangeRate} onChange={(e) => setFormExchangeRate(e.target.value)} type="number" />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium">Discount ($)</label>
              <Input value={formDiscount} onChange={(e) => setFormDiscount(e.target.value)} type="number" step="0.01" />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium">Notes</label>
            <textarea value={formNotes} onChange={(e) => setFormNotes(e.target.value)}
              className="w-full rounded border border-input bg-background px-3 py-2 text-sm min-h-[60px]" rows={2}
              placeholder="Optional notes for this invoice" />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="ghost" onClick={() => setShowForm(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create Invoice'}
            </Button>
          </div>
        </div>
      </SlideOver>
    </div>
  )
}
