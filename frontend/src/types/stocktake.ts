export interface CreateStocktakeRequest {
  notes?: string
  type?: string
  category?: string
}

export interface StocktakeListItem {
  id: number
  status: 'draft' | 'completed' | 'cancelled'
  notes?: string
  item_count: number
  counted_count: number
  variance_count: number
  created_by_id?: number
  created_by_name?: string
  created_at: string
  completed_at?: string
}

export interface StocktakeItem {
  id: number
  product_id: number
  sku: string
  barcode?: string
  product_name: string
  expected_qty: number
  counted_qty?: number
  variance?: number
  counted_by_name?: string
  counted_at?: string
}

export interface StocktakeDetail extends StocktakeListItem {
  items: StocktakeItem[]
}

export interface FinalizeItemResult {
  product_id: number
  sku: string
  message: string
}

export interface FinalizeResult {
  adjusted: number
  unchanged: number
  skipped: number
  errors?: FinalizeItemResult[]
}
