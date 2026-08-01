import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { SlideOver } from '@/components/ui/SlideOver'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { PageHeader } from '@/components/layout/PageHeader'
import { Pencil, Trash2, PackagePlus, Scale, LayoutGrid, List, History, Upload, Download, CheckCircle2, XCircle } from 'lucide-react'
import { useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useProducts, useCreateProduct, useUpdateProduct, useDeleteProduct, useReceiveStock, useAdjustStock, useUploadProductImage, useDeleteProductImage, useImportProducts } from '@/hooks/useProducts'
import { useDebounce } from '@/hooks/useDebounce'
import { useSuppliers } from '@/hooks/useSuppliers'
import { useSettings } from '@/hooks/useSettings'
import { distanceUnit, unitLabel } from '@/utils/units'
import { StockBadge } from '@/components/inventory/StockBadge'
import { ProductForm, type ProductFormProps } from '@/components/inventory/ProductForm'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { formatUSD } from '@/utils/currency'
import { TableCard, Th, ActionsTh } from '@/components/ui/table'
import { ProductThumb, productSpec } from '@/components/inventory/ProductThumb'
import { ProductCard } from '@/components/inventory/ProductCard'
import { StockHistory } from '@/components/inventory/StockHistory'
import type { Product, ImportResult } from '@/types/product'

export function Inventory() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [typeFilter, setTypeFilter] = useState('')
  const [tireSizeFilter, setTireSizeFilter] = useState('')
  const [search, setSearch] = useState('')
  const [lowStockOnly, setLowStockOnly] = useState(searchParams.get('low') === '1')
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Product | null>(null)
  const [deleting, setDeleting] = useState<number | null>(null)
  const [receiveProduct, setReceiveProduct] = useState<Product | null>(null)
  const [adjustProduct, setAdjustProduct] = useState<Product | null>(null)
  const [historyProduct, setHistoryProduct] = useState<Product | null>(null)
  const [showImport, setShowImport] = useState(false)
  const [view, setView] = useState<'grid' | 'table'>(() => (localStorage.getItem('inventoryView') as 'grid' | 'table') || 'grid')
  const debouncedSearch = useDebounce(search, 300)

  const changeView = (v: 'grid' | 'table') => {
    setView(v)
    localStorage.setItem('inventoryView', v)
  }

  const params = {
    page,
    per_page: 20,
    type: typeFilter || undefined,
    tire_size: tireSizeFilter || undefined,
    name_like: debouncedSearch || undefined,
    stock_quantity_lt: lowStockOnly ? 5 : undefined,
  }

  const { data: settings } = useSettings()
  const { data, isLoading, error } = useProducts(params)
  const createMutation = useCreateProduct()
  const updateMutation = useUpdateProduct()
  const deleteMutation = useDeleteProduct()
  const receiveMutation = useReceiveStock()
  const adjustMutation = useAdjustStock()
  const uploadImage = useUploadProductImage()
  const deleteImage = useDeleteProductImage()
  const importMutation = useImportProducts()

  const handleSubmit: ProductFormProps['onSubmit'] = async (formData, image) => {
    try {
      const product = editing
        ? await updateMutation.mutateAsync({ id: editing.id, data: formData as any })
        : await createMutation.mutateAsync(formData)
      if (image?.file) {
        await uploadImage.mutateAsync({ id: product.id, file: image.file })
      } else if (image?.remove) {
        await deleteImage.mutateAsync(product.id)
      }
      setShowForm(false)
      setEditing(null)
    } catch {
      // Errors are surfaced via the global mutation onError toast; keep the form open.
    }
  }

  const handleDelete = () => {
    if (deleting) {
      deleteMutation.mutate(deleting, {
        onSuccess: () => setDeleting(null),
      })
    }
  }

  const products = data?.data || []
  const meta = data?.meta

  const filters = (
    <>
      <Input
        placeholder="Search products..."
        value={search}
        onChange={(e) => { setSearch(e.target.value); setPage(1) }}
        className="w-56"
      />
      <Select value={typeFilter} onChange={(e) => { setTypeFilter(e.target.value); setPage(1); setTireSizeFilter('') }} className="w-36">
        <option value="">All types</option>
        <option value="tire">Tires</option>
        <option value="part">Parts</option>
        <option value="labor">Labor</option>
        <option value="consumable">Consumables</option>
      </Select>
      {typeFilter === 'tire' && (
        <Input
          placeholder="Tire size (e.g. 205/55R16)"
          value={tireSizeFilter}
          onChange={(e) => { setTireSizeFilter(e.target.value); setPage(1) }}
          className="w-44"
        />
      )}
      <label className="flex cursor-pointer select-none items-center gap-1.5 text-sm text-muted-foreground">
        <input
          type="checkbox"
          checked={lowStockOnly}
          onChange={(e) => { setLowStockOnly(e.target.checked); setPage(1); setSearchParams({}, { replace: true }) }}
          className="accent-primary"
        />
        Low stock only
      </label>
    </>
  )

  return (
    <div className="space-y-6">
      <PageHeader
        title="Inventory"
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setShowImport(true)} size="sm">
              <Upload className="mr-1.5 h-3.5 w-3.5" /> Import CSV
            </Button>
            <Button onClick={() => { setEditing(null); setShowForm(true) }} size="sm">
              Add Product
            </Button>
          </div>
        }
      />

      {/* Loading & Error */}
      {isLoading ? (
        <div className="bg-card rounded-lg shadow-sm p-4">
          <TableSkeleton rows={5} cols={5} />
        </div>
      ) : error ? (
        <p className="text-sm text-destructive">Failed to load products</p>
      ) : null}

      {/* Toolbar: filters + view toggle */}
      {!isLoading && !error && (
        <div className="flex flex-wrap items-center gap-2.5 rounded-lg bg-card p-3 shadow-sm">
          {filters}
          <div className="ml-auto hidden overflow-hidden rounded-md border md:flex">
            <button
              onClick={() => changeView('grid')}
              aria-label="Grid view"
              className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium transition-colors ${view === 'grid' ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}
            >
              <LayoutGrid className="h-3.5 w-3.5" /> Grid
            </button>
            <button
              onClick={() => changeView('table')}
              aria-label="Table view"
              className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium transition-colors ${view === 'table' ? 'bg-primary text-primary-foreground' : 'bg-card text-muted-foreground hover:text-foreground'}`}
            >
              <List className="h-3.5 w-3.5" /> Table
            </button>
          </div>
        </div>
      )}

      {/* Content */}
      {!isLoading && !error && (
        <>
          {products.length === 0 ? (
            <div className="rounded-lg bg-card p-10 text-center text-sm text-muted-foreground shadow-sm">
              No products found
            </div>
          ) : (
            <>
              {/* Desktop table view */}
              {view === 'table' && (
                <div className="hidden md:block">
                  <TableCard>
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="sticky top-0 z-10 border-b bg-card">
                          <Th>Product</Th>
                          <Th>Specs</Th>
                          <Th>Location</Th>
                          <Th className="text-right">Stock</Th>
                          <Th className="text-right">Price</Th>
                          <ActionsTh />
                        </tr>
                      </thead>
                      <tbody>
                        {products.map((p) => {
                          const spec = productSpec(p)
                          return (
                            <tr key={p.id} className="group border-b last:border-0 transition-colors duration-100 hover:bg-muted/50">
                              <td className="px-4 py-2.5">
                                <div className="flex items-center gap-3">
                                  <ProductThumb product={p} className="h-9 w-9 flex-shrink-0" />
                                  <div className="min-w-0">
                                    <p className="truncate font-medium">{p.name}</p>
                                    <p className="font-mono text-xs text-muted-foreground">{p.sku}</p>
                                  </div>
                                </div>
                              </td>
                              <td className="px-4 py-2.5">
                                {spec
                                  ? <span className="text-muted-foreground">{spec}</span>
                                  : <span className="text-xs uppercase tracking-wide text-muted-foreground/70">{p.type}</span>}
                              </td>
                              <td className="px-4 py-2.5">
                                {p.location
                                  ? <span className="font-mono text-xs text-muted-foreground">{p.location}</span>
                                  : <span className="text-muted-foreground">—</span>}
                              </td>
                              <td className="px-4 py-2.5 text-right">
                                <StockBadge quantity={p.stock_quantity} minAlert={p.min_stock_alert} reserved={p.reserved_quantity} />
                              </td>
                              <td className="px-4 py-2.5 text-right tabular-nums">{formatUSD(p.sell_price)}</td>
                              <td className="px-4 py-2.5 text-right">
                                <div className="flex justify-end gap-1 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
                                  <Button variant="ghost" size="icon" onClick={() => { setEditing(p); setShowForm(true) }} aria-label="Edit" title="Edit">
                                    <Pencil className="h-4 w-4" />
                                  </Button>
                                  <Button variant="ghost" size="icon" onClick={() => setReceiveProduct(p)} aria-label="Receive" title="Receive stock">
                                    <PackagePlus className="h-4 w-4" />
                                  </Button>
                                  <Button variant="ghost" size="icon" onClick={() => setAdjustProduct(p)} aria-label="Adjust" title="Adjust stock">
                                    <Scale className="h-4 w-4" />
                                  </Button>
                                  <Button variant="ghost" size="icon" onClick={() => setHistoryProduct(p)} aria-label="History" title="Stock history">
                                    <History className="h-4 w-4" />
                                  </Button>
                                  <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={() => setDeleting(p.id)} aria-label="Delete" title="Delete product">
                                    <Trash2 className="h-4 w-4" />
                                  </Button>
                                </div>
                              </td>
                            </tr>
                          )
                        })}
                      </tbody>
                    </table>
                  </TableCard>
                </div>
              )}

              {/* Desktop grid view */}
              {view === 'grid' && (
                <div className="hidden grid-cols-2 gap-4 md:grid lg:grid-cols-3 xl:grid-cols-4">
                  {products.map((p) => (
                    <ProductCard
                      key={p.id}
                      product={p}
                      onEdit={() => { setEditing(p); setShowForm(true) }}
                      onReceive={() => setReceiveProduct(p)}
                      onAdjust={() => setAdjustProduct(p)}
                      onHistory={() => setHistoryProduct(p)}
                      onDelete={() => setDeleting(p.id)}
                    />
                  ))}
                </div>
              )}

              {/* Mobile: always a card grid */}
              <div className="grid grid-cols-2 gap-3 md:hidden">
                {products.map((p) => (
                  <ProductCard
                    key={p.id}
                    product={p}
                    onEdit={() => { setEditing(p); setShowForm(true) }}
                    onReceive={() => setReceiveProduct(p)}
                    onAdjust={() => setAdjustProduct(p)}
                    onHistory={() => setHistoryProduct(p)}
                    onDelete={() => setDeleting(p.id)}
                  />
                ))}
              </div>

              {/* Pagination */}
              {meta && (
                <div className="flex items-center justify-between rounded-lg bg-card px-4 py-2.5 text-sm shadow-sm">
                  <span className="tabular-nums text-muted-foreground">
                    {meta.total} products{meta.total_pages > 1 && <span> · page {meta.page} of {meta.total_pages}</span>}
                  </span>
                  {meta.total_pages > 1 && (
                    <div className="flex gap-1">
                      <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
                      <Button variant="outline" size="sm" disabled={page >= meta.total_pages} onClick={() => setPage(page + 1)}>Next</Button>
                    </div>
                  )}
                </div>
              )}
            </>
          )}
        </>
      )}

      {/* Create/Edit Slide-over */}
      <SlideOver
        open={showForm}
        onClose={() => { setShowForm(false); setEditing(null) }}
        title={editing ? 'Edit Product' : 'Add Product'}
      >
        <ProductForm
          initial={editing || undefined}
          onSubmit={handleSubmit}
          onCancel={() => { setShowForm(false); setEditing(null) }}
          loading={createMutation.isPending || updateMutation.isPending || uploadImage.isPending || deleteImage.isPending}
          distanceUnitLabel={unitLabel(distanceUnit(settings))}
        />
      </SlideOver>

      {/* Stock History */}
      <StockHistory product={historyProduct} onClose={() => setHistoryProduct(null)} />

      {/* CSV Import */}
      {showImport && (
        <ImportDialog
          onClose={() => setShowImport(false)}
          onImport={(file) => importMutation.mutateAsync(file)}
          result={importMutation.data}
          loading={importMutation.isPending}
          onReset={() => importMutation.reset()}
        />
      )}

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={deleting !== null}
        onClose={() => setDeleting(null)}
        onConfirm={handleDelete}
        title="Delete Product"
        message={`Delete ${products.find(p => p.id === deleting)?.name || 'this product'}? This action cannot be undone.`}
        destructive
        loading={deleteMutation.isPending}
      />

      {/* Receive Stock Dialog */}
      {receiveProduct && <ReceiveDialog
        product={receiveProduct}
        onClose={() => setReceiveProduct(null)}
        onConfirm={(data) => receiveMutation.mutate(
          { id: receiveProduct.id, data },
          { onSuccess: () => setReceiveProduct(null) }
        )}
        loading={receiveMutation.isPending}
      />}

      {/* Adjust Stock Dialog */}
      {adjustProduct && <AdjustDialog
        product={adjustProduct}
        onClose={() => setAdjustProduct(null)}
        onConfirm={(change, reason) => adjustMutation.mutate(
          { id: adjustProduct.id, data: { quantity_change: change, reason } },
          { onSuccess: () => setAdjustProduct(null) }
        )}
        loading={adjustMutation.isPending}
      />}
    </div>
  )
}

function ReceiveDialog({ product, onClose, onConfirm, loading }: {
  product: Product
  onClose: () => void
  onConfirm: (data: { quantity: number; unit_cost?: number; supplier_id?: number; paid?: boolean; dot_code?: string; notes?: string }) => void
  loading: boolean
}) {
  const [qty, setQty] = useState('1')
  const [cost, setCost] = useState('')
  const [supplierId, setSupplierId] = useState('')
  const [paid, setPaid] = useState(true)
  const [dot, setDot] = useState('')
  const [notes, setNotes] = useState('')
  const { data: suppliers } = useSuppliers()
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-sm mx-4">
        <p className="text-sm font-semibold mb-1">Receive Stock</p>
        <p className="text-xs text-muted-foreground mb-3">{product.name} (current: {product.stock_quantity}) — creates a new intake batch</p>
        <div className="space-y-2">
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <label className="text-xs font-medium">Quantity{product.is_bulk ? ` (${product.unit && product.unit !== 'piece' ? product.unit : 'L'})` : ''} *</label>
              <Input value={qty} onChange={(e) => setQty(e.target.value)} type="number" min="0" step={product.is_bulk ? 'any' : '1'} placeholder={product.is_bulk ? 'e.g. 208 (one drum)' : ''} />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium">Unit Cost (USD)</label>
              <Input value={cost} onChange={(e) => setCost(e.target.value)} type="number" step="0.01" min="0" placeholder="Updates buy price" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <label className="text-xs font-medium">Supplier</label>
              <Select value={supplierId} onChange={(e) => setSupplierId(e.target.value)}>
                <option value="">— none —</option>
                {(suppliers || []).map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </Select>
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium">DOT / Lot code</label>
              <Input value={dot} onChange={(e) => setDot(e.target.value)} placeholder={product.dot_code || 'e.g. 3624'} />
            </div>
          </div>
          {supplierId && (
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={paid} onChange={(e) => setPaid(e.target.checked)} className="accent-primary" />
              Paid on delivery <span className="text-xs text-muted-foreground">(uncheck to record as owed)</span>
            </label>
          )}
          <div className="space-y-1">
            <label className="text-xs font-medium">Notes</label>
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="optional" />
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button
            onClick={() => onConfirm({
              quantity: parseFloat(qty) || 1,
              unit_cost: parseFloat(cost) || undefined,
              supplier_id: supplierId ? parseInt(supplierId) : undefined,
              paid: supplierId ? paid : undefined,
              dot_code: dot || undefined,
              notes: notes || undefined,
            })}
            disabled={loading || !qty || parseFloat(qty) <= 0}
          >
            {loading ? 'Receiving...' : 'Receive'}
          </Button>
        </div>
      </div>
    </div>
  )
}

const IMPORT_TEMPLATE_HEADERS = [
  'sku', 'barcode', 'name', 'type', 'description', 'category',
  'buy_price', 'sell_price', 'stock_quantity', 'min_stock_alert', 'unit', 'location',
  'tire_size', 'tire_brand', 'tire_model', 'tire_pattern', 'dot_code', 'load_index', 'speed_rating', 'tire_type',
]
const IMPORT_TEMPLATE_EXAMPLE = [
  'MICH-205-55-16', '8901234567890', 'Michelin Primacy 4 205/55R16', 'tire', '', 'Passenger tire',
  '45', '68', '10', '3', 'piece', 'A-01-03',
  '205/55R16', 'Michelin', 'Primacy 4', '', '2426', '91', 'V', 'passenger',
]

function downloadImportTemplate() {
  const csv = [IMPORT_TEMPLATE_HEADERS.join(','), IMPORT_TEMPLATE_EXAMPLE.join(',')].join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'product-import-template.csv'
  a.click()
  URL.revokeObjectURL(url)
}

function ImportDialog({ onClose, onImport, result, loading, onReset }: {
  onClose: () => void
  onImport: (file: File) => Promise<ImportResult>
  result?: ImportResult
  loading: boolean
  onReset: () => void
}) {
  const fileRef = useRef<HTMLInputElement>(null)
  const [fileName, setFileName] = useState('')
  const [file, setFile] = useState<File | null>(null)

  const handleFile = (f: File | undefined) => {
    if (!f) return
    setFile(f)
    setFileName(f.name)
    onReset()
  }

  const handleSubmit = async () => {
    if (!file) return
    try {
      await onImport(file)
    } catch {
      // surfaced via mutation's global error toast
    }
  }

  const pickAnother = () => {
    setFile(null)
    setFileName('')
    onReset()
    if (fileRef.current) fileRef.current.value = ''
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-md rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Import Products from CSV</p>
        <p className="mb-3 text-xs text-muted-foreground">
          Matches rows to existing products by SKU — known SKUs update, new SKUs are created.
          Stock quantity only applies to new products; use Receive/Adjust to change existing stock.
        </p>

        {!result ? (
          <div className="space-y-3">
            <button
              type="button"
              onClick={downloadImportTemplate}
              className="flex items-center gap-1.5 text-xs font-medium text-primary hover:underline"
            >
              <Download className="h-3.5 w-3.5" /> Download CSV template
            </button>

            <div className="space-y-1.5">
              <label className="text-xs font-medium">CSV file</label>
              <div className="flex items-center gap-2">
                <Button type="button" variant="outline" size="sm" onClick={() => fileRef.current?.click()}>
                  Choose file
                </Button>
                <span className="truncate text-xs text-muted-foreground">{fileName || 'No file chosen'}</span>
              </div>
              <input
                ref={fileRef}
                type="file"
                accept=".csv,text/csv"
                className="hidden"
                onChange={(e) => handleFile(e.target.files?.[0])}
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
              <Button type="button" onClick={handleSubmit} disabled={!file || loading}>
                {loading ? 'Importing...' : 'Import'}
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-3">
            <div className="flex items-center gap-4 rounded-md bg-muted/50 p-3 text-sm">
              <div className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
                <CheckCircle2 className="h-4 w-4" /> {result.created} created
              </div>
              <div className="flex items-center gap-1.5 text-blue-600 dark:text-blue-400">
                <CheckCircle2 className="h-4 w-4" /> {result.updated} updated
              </div>
              {result.failed > 0 && (
                <div className="flex items-center gap-1.5 text-destructive">
                  <XCircle className="h-4 w-4" /> {result.failed} failed
                </div>
              )}
            </div>

            {result.errors && result.errors.length > 0 && (
              <div className="max-h-48 overflow-y-auto rounded-md border">
                <table className="w-full text-xs">
                  <thead className="sticky top-0 bg-card">
                    <tr className="border-b">
                      <Th className="w-12">Row</Th>
                      <Th className="w-28">SKU</Th>
                      <Th>Error</Th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.errors.map((e, i) => (
                      <tr key={i} className="border-b last:border-0">
                        <td className="px-3 py-1.5 tabular-nums text-muted-foreground">{e.row}</td>
                        <td className="px-3 py-1.5 font-mono text-muted-foreground">{e.sku || '—'}</td>
                        <td className="px-3 py-1.5 text-destructive">{e.message}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            <div className="flex justify-end gap-2 pt-2">
              <Button type="button" variant="outline" onClick={pickAnother}>Import another file</Button>
              <Button type="button" onClick={onClose}>Done</Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function AdjustDialog({ product, onClose, onConfirm, loading }: {
  product: Product
  onClose: () => void
  onConfirm: (change: number, reason: string) => void
  loading: boolean
}) {
  const [change, setChange] = useState('0')
  const [reason, setReason] = useState('')
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-sm mx-4">
        <p className="text-sm font-semibold mb-1">Adjust Stock</p>
        <p className="text-xs text-muted-foreground mb-3">{product.name} (current: {product.stock_quantity})</p>
        <div className="space-y-2">
          <div className="space-y-1">
            <label className="text-xs font-medium">Change (+/-){product.is_bulk ? ` (${product.unit && product.unit !== 'piece' ? product.unit : 'L'})` : ''} *</label>
            <Input value={change} onChange={(e) => setChange(e.target.value)} type="number" step={product.is_bulk ? 'any' : '1'} placeholder={product.is_bulk ? 'e.g. -2.5' : 'e.g. 5 or -3'} />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">Reason *</label>
            <Input value={reason} onChange={(e) => setReason(e.target.value)} placeholder="damage, return, correction..." />
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-3">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={() => onConfirm(parseFloat(change) || 0, reason)} disabled={loading || !change || !reason}>
            {loading ? 'Adjusting...' : 'Adjust'}
          </Button>
        </div>
      </div>
    </div>
  )
}
