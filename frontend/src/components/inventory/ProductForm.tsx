import { useRef, useState } from 'react'
import { ImagePlus, X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { imageSrc } from '@/utils/imageUrl'
import type { CreateProductRequest } from '@/types/product'

export interface ProductImageIntent {
  file?: File
  remove?: boolean
}

export interface ProductFormProps {
  initial?: Partial<CreateProductRequest> & { id?: number; image_url?: string }
  onSubmit: (data: CreateProductRequest, image?: ProductImageIntent) => void
  onCancel: () => void
  loading?: boolean
  distanceUnitLabel?: string
}

export function ProductForm({ initial, onSubmit, onCancel, loading, distanceUnitLabel = 'km' }: ProductFormProps) {
  const [type, setType] = useState(initial?.type || 'tire')
  const [sku, setSku] = useState(initial?.sku || '')
  const [barcode, setBarcode] = useState(initial?.barcode || '')
  const [name, setName] = useState(initial?.name || '')
  const [description, setDescription] = useState(initial?.description || '')
  const [category, setCategory] = useState(initial?.category || '')
  const [buyPrice, setBuyPrice] = useState(initial?.buy_price?.toString() || '')
  const [sellPrice, setSellPrice] = useState(initial?.sell_price?.toString() || '')
  const [stockQty, setStockQty] = useState(initial?.stock_quantity?.toString() || '0')
  const [minAlert, setMinAlert] = useState(initial?.min_stock_alert?.toString() || '5')
  const [unit, setUnit] = useState(initial?.unit || 'piece')
  const [location, setLocation] = useState(initial?.location || '')
  const [isOilProduct, setIsOilProduct] = useState(initial?.is_oil_product || false)
  const [isBulk, setIsBulk] = useState(initial?.is_bulk || false)

  const [tireSize, setTireSize] = useState(initial?.tire_size || '')
  const [lifeKm, setLifeKm] = useState(initial?.life_km?.toString() || '')
  const [lifeDays, setLifeDays] = useState(initial?.life_days?.toString() || '')
  const [lifeMonths, setLifeMonths] = useState(initial?.life_months?.toString() || '')
  const [tireBrand, setTireBrand] = useState(initial?.tire_brand || '')
  const [tireModel, setTireModel] = useState(initial?.tire_model || '')
  const [tirePattern, setTirePattern] = useState(initial?.tire_pattern || '')
  const [dotCode, setDotCode] = useState(initial?.dot_code || '')
  const [loadIndex, setLoadIndex] = useState(initial?.load_index || '')
  const [speedRating, setSpeedRating] = useState(initial?.speed_rating || '')
  const [tireType, setTireType] = useState(initial?.tire_type || '')

  const fileRef = useRef<HTMLInputElement>(null)
  const [pendingFile, setPendingFile] = useState<File | null>(null)
  const [pendingPreview, setPendingPreview] = useState<string>('')
  const [removeImage, setRemoveImage] = useState(false)
  const existingImage = !removeImage && !pendingFile ? imageSrc(initial?.image_url) : ''
  const previewSrc = pendingPreview || existingImage

  const pickFile = (file: File | undefined) => {
    if (!file) return
    setPendingFile(file)
    setPendingPreview(URL.createObjectURL(file))
    setRemoveImage(false)
  }

  const clearImage = () => {
    setPendingFile(null)
    setPendingPreview('')
    setRemoveImage(true)
    if (fileRef.current) fileRef.current.value = ''
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const data: CreateProductRequest = {
      type, sku, name, description, category,
      barcode: barcode || undefined,
      buy_price: parseFloat(buyPrice) || 0,
      sell_price: parseFloat(sellPrice) || 0,
      stock_quantity: parseFloat(stockQty) || 0,
      min_stock_alert: parseFloat(minAlert) || 5,
      unit, location,
      is_oil_product: type !== 'tire' ? isOilProduct : undefined,
      is_bulk: type !== 'tire' ? isBulk : undefined,
    }
    if (type === 'tire') {
      data.tire_size = tireSize
      data.tire_brand = tireBrand
      data.tire_model = tireModel
      data.tire_pattern = tirePattern
      data.dot_code = dotCode
      data.load_index = loadIndex
      data.speed_rating = speedRating
      data.tire_type = tireType
    }
    // Service-life rating (km / days / months) — drives the vehicle's
    // due-for-service reminder when this product is sold on an invoice.
    if (type === 'tire' || isOilProduct) {
      data.life_km = lifeKm ? parseInt(lifeKm) : null
      data.life_days = lifeDays ? parseInt(lifeDays) : null
      data.life_months = lifeMonths ? parseInt(lifeMonths) : null
    }
    const image: ProductImageIntent | undefined = pendingFile
      ? { file: pendingFile }
      : removeImage && initial?.image_url
        ? { remove: true }
        : undefined
    onSubmit(data, image)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div className="space-y-1.5">
        <Label>Photo</Label>
        <div className="flex items-center gap-3">
          <div className="relative h-16 w-16 flex-shrink-0">
            {previewSrc ? (
              <>
                <img src={previewSrc} alt="Product preview" className="h-16 w-16 rounded-md object-cover bg-muted" />
                <button
                  type="button"
                  onClick={clearImage}
                  aria-label="Remove photo"
                  className="absolute -right-1.5 -top-1.5 grid h-5 w-5 place-items-center rounded-full bg-destructive text-destructive-foreground shadow"
                >
                  <X className="h-3 w-3" />
                </button>
              </>
            ) : (
              <div className="grid h-16 w-16 place-items-center rounded-md border border-dashed text-muted-foreground">
                <ImagePlus className="h-5 w-5" />
              </div>
            )}
          </div>
          <div>
            <Button type="button" variant="outline" size="sm" onClick={() => fileRef.current?.click()}>
              {previewSrc ? 'Change photo' : 'Upload photo'}
            </Button>
            <p className="mt-1 text-xs text-muted-foreground">JPG, PNG or WebP · up to 8 MB</p>
          </div>
          <input
            ref={fileRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="hidden"
            onChange={(e) => pickFile(e.target.files?.[0])}
          />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label>Type</Label>
        <Select value={type} onChange={(e) => setType(e.target.value)}>
          <option value="tire">Tire</option>
          <option value="part">Part</option>
          <option value="labor">Labor</option>
          <option value="consumable">Consumable</option>
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label>SKU</Label>
          <Input value={sku} onChange={(e) => setSku(e.target.value)} required />
        </div>
        <div className="space-y-1.5">
          <Label>Name</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="space-y-1.5">
          <Label>Barcode</Label>
          <Input value={barcode} onChange={(e) => setBarcode(e.target.value)} placeholder="UPC/EAN — scan or type" />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label>Description</Label>
        <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label>Category</Label>
          <Input value={category} onChange={(e) => setCategory(e.target.value)} />
        </div>
        <div className="space-y-1.5">
          <Label>Unit</Label>
          <Select value={unit} onChange={(e) => setUnit(e.target.value)}>
            <option value="piece">Piece</option>
            <option value="set">Set</option>
            <option value="liter">Liter</option>
            <option value="hour">Hour</option>
          </Select>
        </div>
      </div>

      {type !== 'tire' && (
        <div className="space-y-2">
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input type="checkbox" checked={isOilProduct} onChange={(e) => setIsOilProduct(e.target.checked)} className="accent-primary" />
            Engine oil product
            <span className="text-xs text-muted-foreground">— selling this logs an oil-change reminder on the customer's vehicle</span>
          </label>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={isBulk}
              onChange={(e) => {
                setIsBulk(e.target.checked)
                if (e.target.checked && unit === 'piece') setUnit('liter')
              }}
              className="accent-primary"
            />
            Bulk — sold by volume from a drum
            <span className="text-xs text-muted-foreground">— tracks fractional quantities and shows a drum gauge</span>
          </label>
        </div>
      )}

      {(type === 'tire' || isOilProduct) && (
        <div className="rounded-md bg-muted/50 p-4">
          <p className="text-sm font-semibold mb-0.5">{type === 'tire' ? 'Tire service life' : 'Oil change interval'}</p>
          <p className="mb-2 text-[11px] text-muted-foreground">
            How long after this is sold the customer is due again — recorded from the invoice and drives their reminder. Leave blank to use the shop default.
          </p>
          <div className="grid grid-cols-3 gap-3">
            <div className="space-y-1">
              <Label className="text-xs">Distance ({distanceUnitLabel})</Label>
              <Input value={lifeKm} onChange={(e) => setLifeKm(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder={distanceUnitLabel === 'mi' ? 'e.g. 3000' : 'e.g. 5000'} />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">Days</Label>
              <Input value={lifeDays} onChange={(e) => setLifeDays(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="e.g. 90" />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">Months</Label>
              <Input value={lifeMonths} onChange={(e) => setLifeMonths(e.target.value.replace(/[^\d]/g, ''))} inputMode="numeric" placeholder="e.g. 3" />
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label>Buy Price (USD)</Label>
          <Input value={buyPrice} onChange={(e) => setBuyPrice(e.target.value)} type="number" step="0.01" min="0" />
        </div>
        <div className="space-y-1.5">
          <Label>Sell Price (USD)</Label>
          <Input value={sellPrice} onChange={(e) => setSellPrice(e.target.value)} type="number" step="0.01" min="0" />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label>Stock Quantity</Label>
          <Input value={stockQty} onChange={(e) => setStockQty(e.target.value)} type="number" min="0" step="any" />
          {initial && <p className="text-[10px] text-muted-foreground">Changing stock here records a stock adjustment (with reason) in the ledger.</p>}
        </div>
        <div className="space-y-1.5">
          <Label>Min Stock Alert</Label>
          <Input value={minAlert} onChange={(e) => setMinAlert(e.target.value)} type="number" min="0" step="any" />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label>Location</Label>
        <Input value={location} onChange={(e) => setLocation(e.target.value)} placeholder="A-01-03" />
      </div>

      {type === 'tire' && (
        <div className="bg-muted/50 rounded-md p-4 space-y-3">
          <Label className="text-sm font-semibold">Tire Specs</Label>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label className="text-xs">Size</Label>
              <Input value={tireSize} onChange={(e) => setTireSize(e.target.value)} placeholder="205/55R16" />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs">Brand</Label>
              <Input value={tireBrand} onChange={(e) => setTireBrand(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs">Model</Label>
              <Input value={tireModel} onChange={(e) => setTireModel(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs">Pattern</Label>
              <Input value={tirePattern} onChange={(e) => setTirePattern(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs">DOT Code</Label>
              <Input value={dotCode} onChange={(e) => setDotCode(e.target.value)} placeholder="2026" />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs">Type</Label>
              <Select value={tireType} onChange={(e) => setTireType(e.target.value)}>
                <option value="">Select...</option>
                <option value="passenger">Passenger</option>
                <option value="truck">Truck</option>
                <option value="suv">SUV</option>
                <option value="motorcycle">Motorcycle</option>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs">Load Index</Label>
              <Input value={loadIndex} onChange={(e) => setLoadIndex(e.target.value)} placeholder="91" />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs">Speed Rating</Label>
              <Select value={speedRating} onChange={(e) => setSpeedRating(e.target.value)}>
                <option value="">Select...</option>
                <option value="S">S (180 km/h)</option>
                <option value="T">T (190 km/h)</option>
                <option value="H">H (210 km/h)</option>
                <option value="V">V (240 km/h)</option>
                <option value="W">W (270 km/h)</option>
                <option value="Y">Y (300 km/h)</option>
              </Select>
            </div>
          </div>
        </div>
      )}

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onCancel}>Cancel</Button>
        <Button type="submit" disabled={loading}>{loading ? 'Saving...' : 'Save'}</Button>
      </div>
    </form>
  )
}
