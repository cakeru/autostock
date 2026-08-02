export interface Product {
  id: number
  branch_id: number
  type: 'tire' | 'part' | 'labor' | 'consumable'
  sku: string
  barcode?: string
  name: string
  description?: string
  category?: string
  buy_price: number
  sell_price: number
  stock_quantity: number
  reserved_quantity: number
  min_stock_alert: number
  unit: string
  is_oil_product?: boolean
  is_bulk?: boolean
  life_km?: number | null
  life_days?: number | null
  life_months?: number | null
  tire_size?: string
  tire_brand?: string
  tire_model?: string
  tire_pattern?: string
  dot_code?: string
  load_index?: string
  speed_rating?: string
  tire_type?: string
  location?: string
  image_url?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface StockMovement {
  id: number
  quantity_change: number
  reason: string
  reference_type?: string
  reference_id?: number
  invoice_number?: string
  batch_no?: string
  recorded_by_name?: string
  balance_after: number
  created_at: string
}

export interface Batch {
  id: number
  batch_no: string
  supplier?: string
  dot_code?: string
  unit_cost: number
  quantity_received: number
  quantity_remaining: number
  notes?: string
  received_by_name?: string
  received_at: string
}

export interface BatchConsumer {
  invoice_id: number
  invoice_number: string
  customer_name: string
  quantity: number
  created_at: string
}

export interface LowStockProduct {
  id: number
  sku: string
  name: string
  stock_quantity: number
  reserved_quantity: number
  min_stock_alert: number
  sell_price: number
}

export interface ProductListParams {
  type?: string
  name_like?: string
  tire_size?: string
  tire_brand?: string
  stock_quantity_lt?: number
  category?: string
  page?: number
  per_page?: number
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: {
    page: number
    per_page: number
    total: number
    total_pages: number
  }
}

export interface CreateProductRequest {
  type: string
  sku: string
  barcode?: string
  name: string
  description?: string
  category?: string
  buy_price?: number
  sell_price?: number
  stock_quantity?: number
  min_stock_alert?: number
  unit?: string
  is_oil_product?: boolean
  is_bulk?: boolean
  life_km?: number | null
  life_days?: number | null
  life_months?: number | null
  tire_size?: string
  tire_brand?: string
  tire_model?: string
  tire_pattern?: string
  dot_code?: string
  load_index?: string
  speed_rating?: string
  tire_type?: string
  location?: string
}

export interface UpdateProductRequest {
	type?: string
	barcode?: string
	name?: string
	description?: string
	category?: string
	buy_price?: number
	sell_price?: number
	min_stock_alert?: number
	unit?: string
	is_oil_product?: boolean
	is_bulk?: boolean
	life_km?: number | null
	life_days?: number | null
	life_months?: number | null
	tire_size?: string
	tire_brand?: string
	tire_model?: string
	tire_pattern?: string
	dot_code?: string
	load_index?: string
	speed_rating?: string
	tire_type?: string
	location?: string
}

export interface ReceiveStockRequest {
	quantity: number
	unit_cost?: number
	supplier_id?: number
	paid?: boolean
	supplier?: string
	dot_code?: string
	notes?: string
}

export interface AdjustStockRequest {
	quantity_change: number
	reason: string
}

export interface ImportRowError {
	row: number
	sku?: string
	message: string
}

export interface ImportResult {
	total_rows: number
	created: number
	updated: number
	failed: number
	errors?: ImportRowError[]
}
