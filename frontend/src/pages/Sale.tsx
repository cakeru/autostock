import { useMemo, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'
import { Plus, Minus, X, Trash2, Search, Package as PackageIcon, Delete, Check, ScanLine, Percent, Scissors, ShoppingCart, ChevronUp, ChevronDown, Car, Wrench, AlertTriangle } from 'lucide-react'
import { useProducts } from '@/hooks/useProducts'
import { productsApi } from '@/services/products'
import { useCustomers, useCreateCustomer, useCustomerVehicles, useCreateVehicle } from '@/hooks/useCustomers'
import { useSettings } from '@/hooks/useSettings'
import { useCreateInvoice, useRecordPayment, useUploadPaymentProof } from '@/hooks/useInvoices'
import { useCreateServiceJob } from '@/hooks/useServiceJobs'
import { ProductThumb, productSpec } from '@/components/inventory/ProductThumb'
import { Button } from '@/components/ui/button'
import { StockBadge } from '@/components/inventory/StockBadge'
import { formatUSD } from '@/utils/currency'
import { distanceUnit, unitLabel } from '@/utils/units'
import { cn } from '@/lib/utils'
import { PACKAGES, LABOR_PRESETS, FEE_PRESETS, DEFAULT_PAYMENT_METHODS, nextKey, parseDiscount, parseArraySetting } from '@/lib/packages'
import type { CartLine, SalePackage, Preset } from '@/lib/packages'
import type { Product } from '@/types/product'
import type { Customer, Vehicle } from '@/types/customer'

type Tab = 'packages' | 'tires' | 'parts' | 'consumables' | 'labor' | 'fees'
const TABS: { id: Tab; label: string }[] = [
  { id: 'packages', label: 'Packages' },
  { id: 'tires', label: 'Tires' },
  { id: 'parts', label: 'Parts' },
  { id: 'consumables', label: 'Consumables' },
  { id: 'labor', label: 'Labor' },
  { id: 'fees', label: 'Fees' },
]

export function Sale() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [tab, setTab] = useState<Tab>('packages')
  const [search, setSearch] = useState('')
  const [cart, setCart] = useState<CartLine[]>([])
  const [customerId, setCustomerId] = useState(searchParams.get('customer') || '')
  const [vehicleId, setVehicleId] = useState('')
  const [mileage, setMileage] = useState('')
  const [pkg, setPkg] = useState<SalePackage | null>(null) // tire-picker overlay
  const [paying, setPaying] = useState(false)
  const [cartOpen, setCartOpen] = useState(false) // mobile bottom-sheet
  const [done, setDone] = useState<{ id: number; number: string } | null>(null)
  const [discMode, setDiscMode] = useState<'percent' | 'amount' | 'target' | null>(null)
  const [discInput, setDiscInput] = useState('')
  const [discReason, setDiscReason] = useState('')

  const { data: productsData } = useProducts({ per_page: 100 })
  const { data: customersData } = useCustomers({ per_page: 100 })
  const { data: vehicles } = useCustomerVehicles(customerId ? parseInt(customerId) : 0)
  const { data: settings } = useSettings()
  const createInvoice = useCreateInvoice()
  const recordPayment = useRecordPayment()
  const uploadProof = useUploadPaymentProof()
  const createJob = useCreateServiceJob()

  const selectedVehicle = vehicles?.find((v) => String(v.id) === vehicleId)
  const selectCustomer = (id: string) => { setCustomerId(id); setVehicleId(''); setMileage('') }

  const products = productsData?.data || []
  const customers = customersData?.data || []
  const rate = settings?.exchange_rate_usd_khr || 4050

  // Owner-editable lists come from settings; fall back to built-in defaults.
  const packages = useMemo(() => parseArraySetting<SalePackage>(settings?.sale_packages, PACKAGES), [settings?.sale_packages])
  const laborPresets = useMemo(() => parseArraySetting<Preset>(settings?.labor_presets, LABOR_PRESETS), [settings?.labor_presets])
  const feePresets = useMemo(() => parseArraySetting<Preset>(settings?.fee_presets, FEE_PRESETS), [settings?.fee_presets])
  const paymentMethods = useMemo(() => parseArraySetting<string>(settings?.payment_methods, DEFAULT_PAYMENT_METHODS), [settings?.payment_methods])

  // Tax is applied only when enabled in Settings; computed on the gross subtotal
  // to match the backend (subtotal + tax − discount).
  const taxRate = settings?.tax_enabled ? (settings?.tax_rate_percent || 0) : 0

  const lineGross = (l: CartLine) => l.quantity * l.unit_price_usd
  const subtotal = cart.reduce((s, l) => s + lineGross(l), 0)
  const itemCount = cart.reduce((s, l) => s + l.quantity, 0)
  // A tire/oil line in the cart without a vehicle means the backend won't log
  // the oil change / tire install on the car — warn staff while they can fix it.
  const hasServiceProduct = cart.some((l) => l.product_type === 'tire' || l.is_oil_product)
  const lineDiscTotal = cart.reduce((s, l) => s + parseDiscount(l.discountRaw, lineGross(l)), 0)
  const afterLine = subtotal - lineDiscTotal
  const rawDisc = discMode === 'percent' ? afterLine * (parseFloat(discInput) || 0) / 100
    : discMode === 'amount' ? (parseFloat(discInput) || 0)
    : discMode === 'target' ? afterLine - (parseFloat(discInput) || 0)
    : 0
  const orderDiscount = Math.min(afterLine, Math.max(0, Math.round(rawDisc * 100) / 100))
  const discountAmount = Math.round((lineDiscTotal + orderDiscount) * 100) / 100
  const taxAmount = Math.round(subtotal * taxRate) / 100
  const totalUSD = Math.round((subtotal + taxAmount - discountAmount) * 100) / 100
  const totalKHR = Math.round(totalUSD * rate)

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return products.filter((p) => {
      if (tab === 'tires' && p.type !== 'tire') return false
      if (tab === 'parts' && p.type !== 'part') return false
      if (tab === 'consumables' && p.type !== 'consumable') return false
      if (!q) return true
      return p.name.toLowerCase().includes(q) || (p.tire_size || '').toLowerCase().includes(q) || p.sku.toLowerCase().includes(q)
    })
  }, [products, tab, search])

  // --- cart ops ---
  const addProduct = (p: Product) => {
    setCart((c) => {
      // Merge into an existing line of this product only if it's untouched
      // (default price, no discount). Once a line is repriced or discounted it
      // is "claimed", so the next add starts a fresh row — this lets you both
      // bulk-add-then-discount and price individual units separately.
      const pristine = c.find((l) => l.product_id === p.id && !l.discountRaw && l.unit_price_usd === p.sell_price)
      if (pristine) return c.map((l) => (l.key === pristine.key ? { ...l, quantity: l.quantity + 1 } : l))
      return [...c, { key: nextKey(), product_id: p.id, item_type: 'product', description: p.name, quantity: 1, unit_price_usd: p.sell_price, image_url: p.image_url, is_bulk: p.is_bulk, unit: p.unit, product_type: p.type, is_oil_product: p.is_oil_product }]
    })
  }
  const addLine = (description: string, unit_price_usd: number, item_type: CartLine['item_type']) => {
    setCart((c) => [...c, { key: nextKey(), item_type, description, quantity: 1, unit_price_usd }])
  }
  const addPackage = (p: Product, qty: number, sp: SalePackage) => {
    const lines: CartLine[] = [
      { key: nextKey(), product_id: p.id, item_type: 'product', description: p.name, quantity: qty, unit_price_usd: p.sell_price, image_url: p.image_url, product_type: p.type, is_oil_product: p.is_oil_product },
      ...sp.addons.map((a) => ({ key: nextKey(), item_type: a.item_type, description: a.description, quantity: a.per_tire ? qty : 1, unit_price_usd: a.unit_price_usd })),
    ]
    setCart((c) => [...c, ...lines])
    setPkg(null)
  }
  const addBySku = async (raw: string) => {
    const code = raw.trim()
    if (!code) return
    const lc = code.toLowerCase()
    // Fast path: match SKU or barcode in the loaded catalog.
    let p: Product | undefined = products.find((x) => x.sku.toLowerCase() === lc || (x.barcode || '').toLowerCase() === lc)
    // Fall back to a backend lookup (barcode or SKU) for anything not loaded.
    if (!p) p = (await productsApi.lookup(code)) || undefined
    if (!p) { toast.error(`No product for "${code}"`); return }
    addProduct(p)
    toast.success(`Added ${p.name}`)
  }
  const setQty = (key: string, q: number) => setCart((c) => c.map((l) => (l.key === key ? { ...l, quantity: Math.max(1, q) } : l)))
  const setBulkQty = (key: string, q: number) => setCart((c) => c.map((l) => (l.key === key ? { ...l, quantity: isNaN(q) || q < 0 ? 0 : q } : l)))
  const setPrice = (key: string, price: number) => setCart((c) => c.map((l) => (l.key === key ? { ...l, unit_price_usd: isNaN(price) ? 0 : price } : l)))
  const setLineDisc = (key: string, raw: string) => setCart((c) => c.map((l) => (l.key === key ? { ...l, discountRaw: raw } : l)))
  const removeLine = (key: string) => setCart((c) => c.filter((l) => l.key !== key))
  // Peel one unit off a line into its own line so it can be priced/discounted separately.
  const splitLine = (key: string) => setCart((c) => {
    const i = c.findIndex((l) => l.key === key)
    if (i < 0 || c[i].quantity <= 1) return c
    const next = [...c]
    next[i] = { ...c[i], quantity: c[i].quantity - 1 }
    next.splice(i + 1, 0, { ...c[i], key: nextKey(), quantity: 1, discountRaw: undefined })
    return next
  })

  const reset = () => { setCart([]); setCustomerId(''); setVehicleId(''); setMileage(''); setDone(null); setPaying(false); setCartOpen(false); setTab('packages'); setSearch(''); setDiscMode(null); setDiscInput(''); setDiscReason('') }

  const mileageValue = () => { const n = parseInt(mileage, 10); return Number.isFinite(n) && n > 0 ? n : undefined }

  // Park the current cart as a service job instead of charging now — for when the
  // car needs to stay, needs a quote, or the work will happen over time.
  const saveAsJob = async () => {
    if (cart.length === 0) return
    const desc = `${cart[0].description}${cart.length > 1 ? ` + ${cart.length - 1} more` : ''}`
    try {
      const job = await createJob.mutateAsync({
        customer_id: customerId ? parseInt(customerId) : undefined,
        vehicle_id: vehicleId ? parseInt(vehicleId) : undefined,
        mileage: mileageValue(),
        description: desc,
        discount: discountAmount || undefined,
        notes: discountAmount > 0 ? `Discount discussed: ${discReason || 'negotiated'}` : undefined,
        items: cart.map((l) => ({ product_id: l.product_id, item_type: l.item_type, description: l.description, quantity: l.quantity, unit_price: l.unit_price_usd })),
      })
      navigate(`/service-jobs/${job.id}`)
    } catch { /* surfaced by global onError toast */ }
  }

  const submit = async (method: string | null, tender?: { currency: string; tendered_amount: number; exchange_rate: number }, extra?: { reference?: string; proof?: File }) => {
    try {
      const inv = await createInvoice.mutateAsync({
        customer_id: customerId ? parseInt(customerId) : undefined,
        vehicle_id: vehicleId ? parseInt(vehicleId) : undefined,
        mileage: mileageValue(),
        items: cart.map((l) => ({ product_id: l.product_id, item_type: l.item_type, description: l.description, quantity: l.quantity, unit_price_usd: l.unit_price_usd })),
        discount: discountAmount || undefined,
        exchange_rate: rate,
        notes: discountAmount > 0 ? `Discount: ${discReason || 'negotiated'}` : undefined,
      })
      if (method) {
        const pay = await recordPayment.mutateAsync({ id: inv.id, data: { amount: totalUSD, method, reference: extra?.reference, ...(tender || {}) } })
        if (extra?.proof) {
          await uploadProof.mutateAsync({ invoiceId: inv.id, paymentId: pay.id, file: extra.proof })
        }
      }
      setDone({ id: inv.id, number: inv.invoice_number })
    } catch {
      // surfaced by global onError toast; stay on the sale
    }
  }

  if (done) {
    return (
      <div className="mx-auto flex min-h-[70vh] max-w-sm flex-col items-center justify-center text-center">
        <div className="grid h-16 w-16 place-items-center rounded-full bg-primary/10 text-primary">
          <Check className="h-8 w-8" />
        </div>
        <h1 className="mt-4 text-xl font-semibold">Sale complete</h1>
        <p className="mt-1 font-mono text-sm text-muted-foreground">{done.number}</p>
        <p className="mt-1 text-2xl font-bold tabular-nums">{formatUSD(totalUSD)}</p>
        <div className="mt-6 flex w-full flex-col gap-2">
          <Button onClick={() => navigate(`/invoices/${done.id}`)}>Open &amp; print receipt</Button>
          <Button variant="outline" onClick={reset}>New sale</Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 lg:h-[calc(100vh-6rem)] lg:flex-row">
      {/* Catalog */}
      <div className="flex min-h-0 flex-1 flex-col">
        <div className="relative mb-3">
          <ScanLine className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            placeholder="Scan barcode or type SKU, then Enter"
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                addBySku((e.target as HTMLInputElement).value)
                ;(e.target as HTMLInputElement).value = ''
              }
            }}
            className="h-11 w-full rounded-md border border-input bg-card pl-9 pr-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
          />
        </div>
        <div className="mb-3 flex items-center gap-2 overflow-x-auto">
          {TABS.map((t) => (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={`whitespace-nowrap rounded-full px-4 py-2 text-sm font-medium transition-colors ${tab === t.id ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}
            >
              {t.label}
            </button>
          ))}
        </div>

        {(tab === 'tires' || tab === 'parts' || tab === 'consumables') && (
          <div className="relative mb-3">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={tab === 'tires' ? 'Search size or name (205/55R16)…' : tab === 'consumables' ? 'Search consumables…' : 'Search parts…'}
              className="h-11 w-full rounded-md border border-input bg-card pl-9 pr-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
          </div>
        )}

        <div className={cn('min-h-0 flex-1 overflow-y-auto', cart.length > 0 ? 'pb-28 lg:pb-2' : 'pb-2')}>
          {tab === 'packages' && (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
              {packages.map((sp) => (
                <button key={sp.id} onClick={() => setPkg(sp)} className="flex flex-col items-start gap-1 rounded-lg bg-card p-4 text-left shadow-sm transition-shadow hover:shadow-md">
                  <div className="grid h-10 w-10 place-items-center rounded-md bg-primary/10 text-primary"><PackageIcon className="h-5 w-5" /></div>
                  <p className="mt-1 text-sm font-semibold">{sp.name}</p>
                  <p className="text-xs text-muted-foreground">{sp.blurb}</p>
                </button>
              ))}
            </div>
          )}

          {(tab === 'tires' || tab === 'parts' || tab === 'consumables') && (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-4">
              {filtered.length === 0 ? (
                <p className="col-span-full py-8 text-center text-sm text-muted-foreground">No products found</p>
              ) : filtered.map((p) => (
                <button key={p.id} onClick={() => addProduct(p)} disabled={p.stock_quantity <= 0}
                  className="flex flex-col overflow-hidden rounded-lg bg-card text-left shadow-sm transition-shadow hover:shadow-md disabled:opacity-50">
                  <div className="relative aspect-square bg-muted">
                    <ProductThumb product={p} className="absolute inset-0 h-full w-full rounded-none" />
                    <div className="absolute left-1.5 top-1.5"><StockBadge quantity={p.stock_quantity} minAlert={p.min_stock_alert} /></div>
                  </div>
                  <div className="p-2">
                    <p className="truncate text-xs font-medium">{p.name}</p>
                    <p className="truncate text-[11px] text-muted-foreground">{productSpec(p) || ' '}</p>
                    <p className="mt-0.5 text-sm font-semibold tabular-nums">{formatUSD(p.sell_price)}</p>
                  </div>
                </button>
              ))}
            </div>
          )}

          {tab === 'labor' && (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
              {laborPresets.map((l) => (
                <button key={l.description} onClick={() => addLine(l.description, l.unit_price_usd, 'labor')} className="flex items-center justify-between rounded-lg bg-card p-4 text-left text-sm shadow-sm transition-shadow hover:shadow-md">
                  <span className="font-medium">{l.description}</span>
                  <span className="tabular-nums text-muted-foreground">{formatUSD(l.unit_price_usd)}</span>
                </button>
              ))}
            </div>
          )}

          {tab === 'fees' && (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
              {feePresets.map((f) => (
                <button key={f.description} onClick={() => addLine(f.description, f.unit_price_usd, 'fee')} className="flex items-center justify-between rounded-lg bg-card p-4 text-left text-sm shadow-sm transition-shadow hover:shadow-md">
                  <span className="font-medium">{f.description}</span>
                  <span className="tabular-nums text-muted-foreground">{formatUSD(f.unit_price_usd)}</span>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Backdrop for the mobile cart sheet */}
      {cartOpen && <div className="fixed inset-0 z-40 bg-black/40 lg:hidden" onClick={() => setCartOpen(false)} />}

      {/* Cart — right rail on desktop, slide-up bottom sheet on mobile */}
      <div className={cn(
        'flex flex-col bg-card shadow-sm',
        'fixed inset-x-0 bottom-0 z-50 max-h-[85vh] rounded-t-2xl transition-transform duration-300',
        cartOpen ? 'translate-y-0' : 'translate-y-full',
        'lg:static lg:z-auto lg:max-h-none lg:w-[380px] lg:translate-y-0 lg:rounded-lg lg:shadow-sm lg:transition-none'
      )}>
        <div className="border-b p-3">
          <div className="mb-1 flex items-center justify-between">
            <p className="text-sm font-semibold">Current sale</p>
            <button onClick={() => setCartOpen(false)} aria-label="Close" className="-mr-1 text-muted-foreground hover:text-foreground lg:hidden">
              <X className="h-5 w-5" />
            </button>
          </div>
          <CustomerPicker value={customerId} onChange={selectCustomer} customers={customers} />
          {customerId && (
            <div className="mt-2">
              <VehiclePicker value={vehicleId} onChange={setVehicleId} customerId={parseInt(customerId)} vehicles={vehicles || []} />
            </div>
          )}
          {vehicleId && (
            <div className="mt-2 flex items-center gap-2">
              <label className="text-xs text-muted-foreground whitespace-nowrap">Odometer ({selectedVehicle?.distance_unit === 'mi' ? 'mi' : 'km'})</label>
              <input
                value={mileage}
                onChange={(e) => setMileage(e.target.value.replace(/[^\d]/g, ''))}
                inputMode="numeric"
                placeholder="e.g. 85000"
                className="h-8 w-full rounded-md border border-input bg-background px-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
              />
            </div>
          )}
          {hasServiceProduct && !vehicleId && (
            <div className="mt-2 flex items-start gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-2.5 py-2 text-[11px] leading-snug text-amber-700">
              <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
              <span>This sale includes a tire or oil product but no vehicle is selected — the service record won't be logged on the car automatically. Pick a customer + vehicle to record it.</span>
            </div>
          )}
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-2 lg:max-h-none max-h-[40vh]">
          {cart.length === 0 ? (
            <p className="py-10 text-center text-sm text-muted-foreground">Tap products to add them</p>
          ) : (
            <div className="divide-y divide-border/70">
              {cart.map((l) => {
                const gross = lineGross(l)
                const disc = parseDiscount(l.discountRaw, gross)
                const net = gross - disc
                return (
                <div key={l.key} className="py-2">
                  <div className="flex items-start justify-between gap-2">
                    <p className="min-w-0 flex-1 truncate text-sm font-medium">{l.description}</p>
                    <div className="flex flex-shrink-0 items-center gap-2">
                      <span className="text-sm font-semibold tabular-nums">
                        {disc > 0 && <span className="mr-1 text-xs font-normal text-muted-foreground line-through">{formatUSD(gross)}</span>}
                        {formatUSD(net)}
                      </span>
                      <button onClick={() => removeLine(l.key)} className="text-muted-foreground hover:text-destructive"><X className="h-4 w-4" /></button>
                    </div>
                  </div>
                  <div className="mt-1 flex flex-wrap items-center gap-1.5">
                    {l.is_bulk ? (
                      <div className="flex items-center rounded-md border px-1.5" title="Quantity">
                        <input value={l.quantity} onChange={(e) => setBulkQty(l.key, parseFloat(e.target.value))} type="number" step="0.1" min="0" className="w-12 bg-transparent py-1 text-sm tabular-nums focus:outline-none" />
                        <span className="text-xs text-muted-foreground">{l.unit && l.unit !== 'piece' ? l.unit : 'L'}</span>
                      </div>
                    ) : (
                      <>
                        <button onClick={() => setQty(l.key, l.quantity - 1)} className="grid h-7 w-7 place-items-center rounded-md border text-muted-foreground hover:bg-muted"><Minus className="h-3.5 w-3.5" /></button>
                        <span className="w-5 text-center text-sm tabular-nums">{l.quantity}</span>
                        <button onClick={() => setQty(l.key, l.quantity + 1)} className="grid h-7 w-7 place-items-center rounded-md border text-muted-foreground hover:bg-muted"><Plus className="h-3.5 w-3.5" /></button>
                      </>
                    )}
                    <div className="flex items-center rounded-md border px-1.5" title="Unit price">
                      <span className="text-xs text-muted-foreground">$</span>
                      <input value={l.unit_price_usd} onChange={(e) => setPrice(l.key, parseFloat(e.target.value))} type="number" step="0.01" min="0" className="w-12 bg-transparent py-1 text-sm tabular-nums focus:outline-none" />
                    </div>
                    <div className={`flex items-center rounded-md border px-1.5 ${disc > 0 ? 'border-primary/50 text-primary' : ''}`} title="Line discount — type 10% or 5">
                      <Percent className="h-3 w-3 text-muted-foreground" />
                      <input value={l.discountRaw || ''} onChange={(e) => setLineDisc(l.key, e.target.value)} placeholder="disc" className="w-12 bg-transparent py-1 text-sm focus:outline-none" />
                    </div>
                    {l.quantity > 1 && !l.is_bulk && (
                      <button onClick={() => splitLine(l.key)} title="Split one unit into its own line" className="grid h-7 w-7 place-items-center rounded-md border text-muted-foreground hover:bg-muted"><Scissors className="h-3.5 w-3.5" /></button>
                    )}
                  </div>
                </div>
              )})}
            </div>
          )}
        </div>

        <div className="border-t p-3">
          {/* Discount */}
          {cart.length > 0 && (
            <div className="mb-2">
              {discMode === null ? (
                <button onClick={() => setDiscMode('percent')} className="text-xs font-medium text-primary hover:underline">+ Whole-sale discount</button>
              ) : (
                <div className="rounded-md border p-2">
                  <div className="mb-2 flex gap-1">
                    {([['percent', '% off'], ['amount', '$ off'], ['target', 'Set price']] as const).map(([m, label]) => (
                      <button key={m} onClick={() => { setDiscMode(m); setDiscInput('') }}
                        className={`flex-1 rounded px-2 py-1 text-xs font-medium transition-colors ${discMode === m ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}>{label}</button>
                    ))}
                    <button onClick={() => { setDiscMode(null); setDiscInput(''); setDiscReason('') }} aria-label="Remove discount" className="rounded px-1.5 text-muted-foreground hover:text-destructive"><X className="h-3.5 w-3.5" /></button>
                  </div>
                  <div className="flex gap-1.5">
                    <div className="flex flex-1 items-center rounded-md border px-2">
                      <span className="text-xs text-muted-foreground">{discMode === 'percent' ? '%' : '$'}</span>
                      <input value={discInput} onChange={(e) => setDiscInput(e.target.value)} type="number" min="0" step="0.01"
                        placeholder={discMode === 'target' ? 'final total' : discMode === 'percent' ? 'e.g. 10' : 'e.g. 5'}
                        className="w-full bg-transparent px-1 py-1.5 text-sm tabular-nums focus:outline-none" autoFocus />
                    </div>
                    <input value={discReason} onChange={(e) => setDiscReason(e.target.value)} placeholder="Reason (optional)"
                      className="flex-1 rounded-md border px-2 py-1.5 text-sm focus:outline-none" />
                  </div>
                </div>
              )}
            </div>
          )}

          {(discountAmount > 0 || taxAmount > 0) && (
            <div className="mb-1 space-y-0.5 text-sm">
              <div className="flex justify-between text-muted-foreground"><span>Subtotal</span><span className="tabular-nums">{formatUSD(subtotal)}</span></div>
              {discountAmount > 0 && <div className="flex justify-between text-destructive"><span>Discount</span><span className="tabular-nums">−{formatUSD(discountAmount)}</span></div>}
              {taxAmount > 0 && <div className="flex justify-between text-muted-foreground"><span>Tax ({settings?.tax_rate_percent}%)</span><span className="tabular-nums">{formatUSD(taxAmount)}</span></div>}
            </div>
          )}

          <div className="mb-2 flex items-end justify-between">
            <div>
              <p className="text-xs text-muted-foreground">Total</p>
              <p className="text-2xl font-bold tabular-nums">{formatUSD(totalUSD)}</p>
            </div>
            <p className="text-sm tabular-nums text-muted-foreground">៛{totalKHR.toLocaleString()}</p>
          </div>
          <div className="flex gap-2">
            {cart.length > 0 && (
              <Button variant="outline" size="icon" onClick={reset} aria-label="Clear sale"><Trash2 className="h-4 w-4" /></Button>
            )}
            <Button className="h-11 flex-1 text-base" disabled={cart.length === 0} onClick={() => setPaying(true)}>
              Charge {formatUSD(totalUSD)}
            </Button>
          </div>
          {cart.length > 0 && (
            <Button variant="outline" className="mt-2 w-full gap-2" onClick={saveAsJob} disabled={createJob.isPending}>
              <Wrench className="h-4 w-4" /> {createJob.isPending ? 'Saving…' : 'Save as job instead'}
            </Button>
          )}
        </div>
      </div>

      {/* Mobile sticky checkout bar — always-visible total, floats above the tab nav */}
      {cart.length > 0 && !cartOpen && (
        <div
          className="fixed inset-x-0 z-40 flex flex-col gap-1 border-t bg-card px-4 py-2 shadow-[0_-4px_16px_rgba(0,0,0,0.08)] lg:hidden"
          style={{ bottom: 'calc(3.5rem + env(safe-area-inset-bottom, 0px))' }}
        >
          <div className="flex items-center gap-3">
            <button onClick={() => setCartOpen(true)} className="flex min-w-0 flex-1 items-center gap-3 text-left">
              <span className="relative flex-shrink-0">
                <ShoppingCart className="h-6 w-6 text-primary" />
                <span className="absolute -right-2 -top-2 grid h-4 min-w-[1rem] place-items-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">{itemCount}</span>
              </span>
              <span className="min-w-0">
                <span className="block text-[11px] leading-none text-muted-foreground">View sale</span>
                <span className="block text-lg font-bold leading-tight tabular-nums">{formatUSD(totalUSD)}</span>
              </span>
              <ChevronUp className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
            </button>
            <Button className="h-11 flex-shrink-0 px-6 text-base" onClick={() => setPaying(true)}>Charge</Button>
          </div>
          {hasServiceProduct && !vehicleId && (
            <p className="flex items-center gap-1.5 text-[10px] leading-tight text-amber-700">
              <AlertTriangle className="h-3 w-3 shrink-0" />
              Tire/oil sale without a vehicle — service record won't be logged on the car.
            </p>
          )}
        </div>
      )}

      {pkg && (
        <TirePicker
          pkg={pkg}
          tires={products.filter((p) => p.type === 'tire' && (!pkg.tireType || !p.tire_type || p.tire_type === pkg.tireType))}
          onCancel={() => setPkg(null)}
          onConfirm={(tire, qty) => addPackage(tire, qty, pkg)}
        />
      )}

      {paying && (
        <PaymentOverlay
          totalUSD={totalUSD}
          rate={rate}
          methods={paymentMethods}
          pending={createInvoice.isPending || recordPayment.isPending}
          onCancel={() => setPaying(false)}
          onComplete={submit}
        />
      )}
    </div>
  )
}

// --- Tire picker (choose the tire slot + quantity for a package) ---
function TirePicker({ pkg, tires, onCancel, onConfirm }: {
  pkg: SalePackage
  tires: Product[]
  onCancel: () => void
  onConfirm: (tire: Product, qty: number) => void
}) {
  const [search, setSearch] = useState('')
  const [sel, setSel] = useState<Product | null>(null)
  const [qty, setQty] = useState(4)
  const list = tires.filter((t) => {
    const q = search.toLowerCase()
    return !q || t.name.toLowerCase().includes(q) || (t.tire_size || '').toLowerCase().includes(q)
  })
  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 p-0 sm:items-center sm:p-4">
      <div className="flex max-h-[90vh] w-full max-w-lg flex-col rounded-t-lg bg-card shadow-xl sm:rounded-lg">
        <div className="flex items-center justify-between border-b p-3">
          <p className="text-sm font-semibold">{pkg.name} — choose the tire</p>
          <button onClick={onCancel} className="text-muted-foreground hover:text-foreground"><X className="h-5 w-5" /></button>
        </div>
        <div className="relative p-3">
          <Search className="pointer-events-none absolute left-6 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search size (205/55R16)…" className="h-11 w-full rounded-md border border-input bg-background pl-9 pr-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30" />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto px-3">
          <div className="grid grid-cols-2 gap-2">
            {list.map((t) => (
              <button key={t.id} onClick={() => setSel(t)} disabled={t.stock_quantity <= 0}
                className={`flex items-center gap-2 rounded-lg border p-2 text-left transition-colors disabled:opacity-50 ${sel?.id === t.id ? 'border-primary ring-2 ring-primary/30' : 'hover:bg-muted/50'}`}>
                <ProductThumb product={t} className="h-10 w-10 flex-shrink-0" />
                <div className="min-w-0">
                  <p className="truncate text-xs font-medium">{t.name}</p>
                  <p className="text-[11px] tabular-nums text-muted-foreground">{formatUSD(t.sell_price)} · {t.stock_quantity} left</p>
                </div>
              </button>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-3 border-t p-3">
          <div className="flex items-center gap-2">
            <button onClick={() => setQty(Math.max(1, qty - 1))} className="grid h-10 w-10 place-items-center rounded-md border hover:bg-muted"><Minus className="h-4 w-4" /></button>
            <span className="w-8 text-center text-lg font-semibold tabular-nums">{qty}</span>
            <button onClick={() => setQty(qty + 1)} className="grid h-10 w-10 place-items-center rounded-md border hover:bg-muted"><Plus className="h-4 w-4" /></button>
          </div>
          <Button className="h-11 flex-1" disabled={!sel} onClick={() => sel && onConfirm(sel, qty)}>
            Add {qty} {sel ? `× ${sel.name}` : 'tire'}
          </Button>
        </div>
      </div>
    </div>
  )
}

// --- Customer: search existing or create a new one inline ---
function CustomerPicker({ value, onChange, customers }: {
  value: string
  onChange: (id: string) => void
  customers: Customer[]
}) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [justCreated, setJustCreated] = useState<Customer | null>(null)
  const create = useCreateCustomer()

  const selected = customers.find((c) => String(c.id) === value)
    || (justCreated && String(justCreated.id) === value ? justCreated : undefined)
  const q = query.trim().toLowerCase()
  const results = (q
    ? customers.filter((c) => c.name.toLowerCase().includes(q) || (c.phone || '').toLowerCase().includes(q))
    : customers
  ).slice(0, 8)

  const pick = (id: string) => { onChange(id); setOpen(false); setQuery(''); setCreating(false) }

  const startCreate = () => {
    setCreating(true)
    // Prefill from what they typed: digits look like a phone, otherwise a name.
    const t = query.trim()
    if (/\d/.test(t)) { setPhone(t); setName('') } else { setName(t); setPhone('') }
  }

  const submitCreate = () => {
    if (!name.trim()) return
    create.mutate(
      { name: name.trim(), phone: phone.trim() || undefined },
      { onSuccess: (c) => { setJustCreated(c); pick(String(c.id)); setName(''); setPhone('') } }
    )
  }

  const inputCls = 'h-9 w-full rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30'

  return (
    <div>
      <button type="button" onClick={() => setOpen((o) => !o)}
        className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 text-sm">
        <span className={cn('truncate', !selected && 'text-muted-foreground')}>
          {selected ? `${selected.name}${selected.phone ? ` — ${selected.phone}` : ''}` : 'Walk-in customer'}
        </span>
        <ChevronDown className={cn('h-4 w-4 flex-shrink-0 text-muted-foreground transition-transform', open && 'rotate-180')} />
      </button>

      {open && (
        <div className="mt-2 rounded-md border bg-background p-2">
          {!creating ? (
            <>
              <div className="relative mb-2">
                <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <input autoFocus value={query} onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search name or phone…"
                  className="h-9 w-full rounded-md border border-input bg-background pl-8 pr-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30" />
              </div>
              <div className="max-h-52 overflow-y-auto">
                <button onClick={() => pick('')}
                  className="flex w-full items-center justify-between rounded px-2 py-2 text-left text-sm hover:bg-muted">
                  <span>Walk-in customer</span>
                  {value === '' && <Check className="h-4 w-4 text-primary" />}
                </button>
                {results.map((c) => (
                  <button key={c.id} onClick={() => pick(String(c.id))}
                    className="flex w-full items-center justify-between gap-2 rounded px-2 py-2 text-left text-sm hover:bg-muted">
                    <span className="min-w-0 truncate"><span className="font-medium">{c.name}</span>{c.phone && <span className="text-muted-foreground"> · {c.phone}</span>}</span>
                    {String(c.id) === value && <Check className="h-4 w-4 flex-shrink-0 text-primary" />}
                  </button>
                ))}
                {q && results.length === 0 && <p className="px-2 py-2 text-xs text-muted-foreground">No matches</p>}
              </div>
              <button onClick={startCreate}
                className="mt-1 flex w-full items-center gap-2 rounded px-2 py-2 text-left text-sm font-medium text-primary hover:bg-primary/5">
                <Plus className="h-4 w-4" /> New customer{query.trim() ? ` “${query.trim()}”` : ''}
              </button>
            </>
          ) : (
            <div className="space-y-2">
              <p className="text-xs font-semibold">New customer</p>
              <input autoFocus value={name} onChange={(e) => setName(e.target.value)} placeholder="Name *" className={inputCls} />
              <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="Phone (recommended)" className={inputCls} />
              <div className="flex gap-2">
                <Button size="sm" variant="ghost" onClick={() => setCreating(false)}>Back</Button>
                <Button size="sm" className="flex-1" disabled={!name.trim() || create.isPending} onClick={submitCreate}>
                  {create.isPending ? 'Saving…' : 'Add & select'}
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// --- Vehicle: pick one of the customer's cars, or add a new one inline ---
function VehiclePicker({ value, onChange, customerId, vehicles }: {
  value: string
  onChange: (id: string) => void
  customerId: number
  vehicles: Vehicle[]
}) {
  const [open, setOpen] = useState(false)
  const [adding, setAdding] = useState(false)
  const [plate, setPlate] = useState('')
  const [makeModel, setMakeModel] = useState('')
  const [justCreated, setJustCreated] = useState<Vehicle | null>(null)
  const create = useCreateVehicle()

  const selected = vehicles.find((v) => String(v.id) === value)
    || (justCreated && String(justCreated.id) === value ? justCreated : undefined)
  const pick = (id: string) => { onChange(id); setOpen(false); setAdding(false) }

  const submit = () => {
    if (!plate.trim()) return
    const [make, ...rest] = makeModel.trim().split(' ')
    create.mutate(
      { customerId, data: { plate_number: plate.trim(), make: make || undefined, model: rest.join(' ') || undefined } },
      { onSuccess: (v) => { setJustCreated(v); pick(String(v.id)); setPlate(''); setMakeModel('') } }
    )
  }

  const label = (v: Vehicle) => `${v.plate_number}${v.make ? ` · ${[v.make, v.model].filter(Boolean).join(' ')}` : ''}`
  const inputCls = 'h-9 w-full rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30'

  return (
    <div>
      <button type="button" onClick={() => setOpen((o) => !o)}
        className="flex h-9 w-full items-center justify-between rounded-md border border-input bg-background px-3 text-sm">
        <span className="flex min-w-0 items-center gap-1.5">
          <Car className="h-3.5 w-3.5 flex-shrink-0 text-muted-foreground" />
          <span className={cn('truncate', !selected && 'text-muted-foreground')}>{selected ? label(selected) : 'No vehicle'}</span>
        </span>
        <ChevronDown className={cn('h-4 w-4 flex-shrink-0 text-muted-foreground transition-transform', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="mt-1 rounded-md border bg-background p-2">
          {!adding ? (
            <>
              <button onClick={() => pick('')} className="flex w-full items-center justify-between rounded px-2 py-2 text-left text-sm hover:bg-muted">
                <span>No vehicle</span>{value === '' && <Check className="h-4 w-4 text-primary" />}
              </button>
              {vehicles.map((v) => (
                <button key={v.id} onClick={() => pick(String(v.id))} className="flex w-full items-center justify-between gap-2 rounded px-2 py-2 text-left text-sm hover:bg-muted">
                  <span className="min-w-0 truncate"><span className="font-medium">{v.plate_number}</span>{v.make && <span className="text-muted-foreground"> · {[v.make, v.model].filter(Boolean).join(' ')}</span>}</span>
                  {String(v.id) === value && <Check className="h-4 w-4 flex-shrink-0 text-primary" />}
                </button>
              ))}
              <button onClick={() => setAdding(true)} className="mt-1 flex w-full items-center gap-2 rounded px-2 py-2 text-left text-sm font-medium text-primary hover:bg-primary/5">
                <Plus className="h-4 w-4" /> Add vehicle
              </button>
            </>
          ) : (
            <div className="space-y-2">
              <p className="text-xs font-semibold">New vehicle</p>
              <input autoFocus value={plate} onChange={(e) => setPlate(e.target.value)} placeholder="Plate number *" className={inputCls} />
              <input value={makeModel} onChange={(e) => setMakeModel(e.target.value)} placeholder="Make & model (e.g. Toyota Camry)" className={inputCls} />
              <div className="flex gap-2">
                <Button size="sm" variant="ghost" onClick={() => setAdding(false)}>Back</Button>
                <Button size="sm" className="flex-1" disabled={!plate.trim() || create.isPending} onClick={submit}>{create.isPending ? 'Saving…' : 'Add & select'}</Button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// --- Payment: number pad + change due + bank-transfer reference/proof ---
function PaymentOverlay({ totalUSD, rate, methods, pending, onCancel, onComplete }: {
  totalUSD: number
  rate: number
  methods: string[]
  pending: boolean
  onCancel: () => void
  onComplete: (method: string | null, tender?: { currency: string; tendered_amount: number; exchange_rate: number }, extra?: { reference?: string; proof?: File }) => void
}) {
  const [tendered, setTendered] = useState('')
  const [method, setMethod] = useState(methods[0] || 'cash')
  const [currency, setCurrency] = useState<'USD' | 'KHR'>('USD')
  const [reference, setReference] = useState('')
  const [proof, setProof] = useState<File | null>(null)
  const proofRef = useRef<HTMLInputElement>(null)
  const tenderedNum = parseFloat(tendered) || 0
  const tenderedUSD = currency === 'USD' ? tenderedNum : tenderedNum / rate
  const change = Math.max(0, tenderedUSD - totalUSD)
  const press = (k: string) => setTendered((t) => {
    if (k === 'del') return t.slice(0, -1)
    if (k === '.' && t.includes('.')) return t
    return t + k
  })
  const isTransfer = method !== 'cash'

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 p-0 sm:items-center sm:p-4">
      <div className="w-full max-w-sm rounded-t-lg bg-card p-4 shadow-xl sm:rounded-lg">
        <div className="mb-3 flex items-center justify-between">
          <p className="text-sm font-semibold">Take payment</p>
          <button onClick={onCancel} className="text-muted-foreground hover:text-foreground"><X className="h-5 w-5" /></button>
        </div>

        <div className="rounded-md bg-muted/50 p-3 text-center">
          <p className="text-xs text-muted-foreground">Total due</p>
          <p className="text-2xl font-bold tabular-nums">{formatUSD(totalUSD)}</p>
          <p className="text-xs tabular-nums text-muted-foreground">៛{Math.round(totalUSD * rate).toLocaleString()}</p>
        </div>

        <div className="mt-3 flex flex-wrap gap-2">
          {methods.map((m) => (
            <button key={m} onClick={() => setMethod(m)} className={`min-w-[4rem] flex-1 rounded-md border py-2 text-sm font-medium capitalize transition-colors ${method === m ? 'border-primary bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}>{m}</button>
          ))}
        </div>

        {isTransfer && (
          <div className="mt-3 space-y-2">
            <div className="space-y-1">
              <label className="text-xs font-medium">Trx ID (ABA) — optional</label>
              <input value={reference} onChange={(e) => setReference(e.target.value)} placeholder="e.g. TRX-8F2K1L9Q"
                className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30" />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium">Proof photo — optional</label>
              <div className="flex items-center gap-2">
                <button type="button" onClick={() => proofRef.current?.click()}
                  className="h-9 rounded-md border border-input bg-background px-3 text-sm hover:bg-muted">
                  {proof ? proof.name.slice(0, 24) : 'Choose photo…'}
                </button>
                {proof && <button type="button" onClick={() => { setProof(null); if (proofRef.current) proofRef.current.value = '' }} className="text-xs text-muted-foreground hover:text-destructive">Remove</button>}
              </div>
              <input ref={proofRef} type="file" accept="image/*" className="hidden" onChange={(e) => setProof(e.target.files?.[0] || null)} />
            </div>
          </div>
        )}

        {method === 'cash' && (
          <>
            <div className="mt-3 flex gap-1 rounded-md border p-0.5">
              {(['USD', 'KHR'] as const).map((c) => (
                <button key={c} onClick={() => { setCurrency(c); setTendered('') }}
                  className={`flex-1 rounded px-2 py-1.5 text-xs font-medium transition-colors ${currency === c ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted'}`}>
                  {c === 'USD' ? 'US Dollar ($)' : 'Riel (៛)'}
                </button>
              ))}
            </div>
            <div className="mt-2 flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Tendered</span>
              <span className="tabular-nums">{tendered ? (currency === 'USD' ? formatUSD(tenderedNum) : `៛${Math.round(tenderedNum).toLocaleString()} · ${formatUSD(tenderedUSD)}`) : '—'}</span>
            </div>
            <div className="flex items-center justify-between text-sm font-medium">
              <span>Change</span>
              <span className="tabular-nums text-primary">{formatUSD(change)} · ៛{Math.round(change * rate).toLocaleString()}</span>
            </div>
            <div className="mt-2 grid grid-cols-3 gap-1.5">
              {['1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '0', 'del'].map((k) => (
                <button key={k} onClick={() => press(k)} className="grid h-12 place-items-center rounded-md border text-lg font-medium hover:bg-muted">
                  {k === 'del' ? <Delete className="h-5 w-5" /> : k}
                </button>
              ))}
            </div>
          </>
        )}

        <div className="mt-3 flex flex-col gap-2">
          <Button className="h-11" disabled={pending} onClick={() => onComplete(method, method === 'cash' ? { currency, tendered_amount: tenderedNum, exchange_rate: rate } : undefined, { reference: reference.trim() || undefined, proof: proof || undefined })}>
            {pending ? 'Completing…' : `Mark paid — ${method}`}
          </Button>
          <Button variant="outline" disabled={pending} onClick={() => onComplete(null)}>
            Charge to account (unpaid)
          </Button>
        </div>
      </div>
    </div>
  )
}
