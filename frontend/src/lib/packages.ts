// Point-of-sale templates. Packages and fees/labor expand into ordinary invoice
// lines: a tire/part line carries a product_id (deducts stock, FIFO/batch
// traced); labor and fee lines have no product_id (revenue only). Prices are
// sensible defaults and stay editable in the cart. (Owner-editable packages in
// Settings are a later step; these are the shop's common bundles to start.)

// item_type maps to the invoice_items CHECK constraint (product | labor |
// custom | fee).
export type LineType = 'product' | 'labor' | 'custom' | 'fee'

export interface CartLine {
  key: string
  product_id?: number
  item_type: LineType
  description: string
  quantity: number
  unit_price_usd: number
  image_url?: string
  is_bulk?: boolean // sold by volume — allow fractional quantity entry
  unit?: string
  product_type?: 'tire' | 'part' | 'consumable' | 'labor' // set on product lines
  is_oil_product?: boolean // drives the auto-logged oil-change record
  discountRaw?: string // per-line discount as typed: "10%" or "5" ($ off the line)
}

// Parse a per-line/order discount entry against a base amount. "10%" → 10% of
// base; "5" → $5. Clamped to [0, base].
export function parseDiscount(raw: string | undefined, base: number): number {
  if (!raw) return 0
  const t = raw.trim()
  if (!t) return 0
  const pct = t.endsWith('%')
  const n = parseFloat(pct ? t.slice(0, -1) : t) || 0
  const amt = pct ? (base * n) / 100 : n
  return Math.min(base, Math.max(0, Math.round(amt * 100) / 100))
}

export interface PackageAddon {
  description: string
  item_type: 'labor' | 'fee'
  unit_price_usd: number
  per_tire: boolean // multiply by the chosen tire quantity
}

export interface SalePackage {
  id: string
  name: string
  blurb: string
  tireType?: string // filters the tire picker (passenger/truck/…)
  addons: PackageAddon[]
}

export const PACKAGES: SalePackage[] = [
  {
    id: 'fitted-passenger',
    name: 'Fitted Tire',
    blurb: 'Tire + mount, balance, valve & disposal',
    tireType: 'passenger',
    addons: [
      { description: 'Tire mounting', item_type: 'labor', unit_price_usd: 15, per_tire: true },
      { description: 'Wheel balancing', item_type: 'labor', unit_price_usd: 10, per_tire: true },
      { description: 'New valve stem', item_type: 'labor', unit_price_usd: 3, per_tire: true },
      { description: 'Tire disposal fee', item_type: 'fee', unit_price_usd: 2, per_tire: true },
    ],
  },
  {
    id: 'fitted-truck',
    name: 'Fitted Tire — Truck',
    blurb: 'Truck/SUV tire + fitting & disposal',
    tireType: 'truck',
    addons: [
      { description: 'Tire mounting (truck)', item_type: 'labor', unit_price_usd: 20, per_tire: true },
      { description: 'Wheel balancing', item_type: 'labor', unit_price_usd: 12, per_tire: true },
      { description: 'New valve stem', item_type: 'labor', unit_price_usd: 3, per_tire: true },
      { description: 'Tire disposal fee', item_type: 'fee', unit_price_usd: 3, per_tire: true },
    ],
  },
]

export interface Preset {
  description: string
  unit_price_usd: number
}

export const DEFAULT_PAYMENT_METHODS = ['cash', 'card', 'transfer']

// Parse an owner-editable JSON-array setting, falling back to a built-in default
// when it's unset or malformed. Used for packages, labor/fee presets, and
// payment methods so the POS stays usable before anything is customised.
export function parseArraySetting<T>(raw: string | undefined, fallback: T[]): T[] {
  if (!raw) return fallback
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed) && parsed.length) return parsed as T[]
  } catch { /* fall through */ }
  return fallback
}

export const LABOR_PRESETS: Preset[] = [
  { description: 'Tire mounting', unit_price_usd: 15 },
  { description: 'Wheel balancing', unit_price_usd: 10 },
  { description: 'Wheel alignment', unit_price_usd: 35 },
  { description: 'Tire rotation', unit_price_usd: 10 },
  { description: 'Flat repair', unit_price_usd: 8 },
]

export const FEE_PRESETS: { description: string; unit_price_usd: number }[] = [
  { description: 'Tire disposal fee', unit_price_usd: 2 },
  { description: 'Shop supplies', unit_price_usd: 3 },
]

let seq = 0
export const nextKey = () => `line-${Date.now()}-${seq++}`
