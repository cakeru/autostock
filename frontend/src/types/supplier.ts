export interface Supplier {
  id: number
  name: string
  phone?: string
  email?: string
  address?: string
  notes?: string
  is_active: boolean
  total_purchased: number
  outstanding: number
  purchase_count: number
  created_at: string
}

export interface Purchase {
  batch_id: number
  product_id?: number
  product_name: string
  quantity: number
  unit_cost: number
  total_cost: number
  amount_paid: number
  owed: number
  dot_code?: string
  received_at: string
}

export interface CreateSupplierRequest {
  name: string
  phone?: string
  email?: string
  address?: string
  notes?: string
}

export type UpdateSupplierRequest = Partial<CreateSupplierRequest>
