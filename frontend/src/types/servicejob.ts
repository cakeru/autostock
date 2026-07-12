export interface ServiceJob {
  id: number
  branch_id: number
  job_number: string
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled'
  priority: 'low' | 'normal' | 'high' | 'urgent'
  customer_id?: number
  customer_name?: string
  customer_phone?: string
  vehicle_id?: number
  plate_number?: string
  vehicle_info?: string
  description: string
  invoice_id?: number
  scheduled_at?: string
  assigned_to_id?: number
  assigned_to_name?: string
  quote_approved_at?: string
  quote_approved_by?: number
  created_at: string
  updated_at: string
}

export interface ServiceJobDetail extends ServiceJob {
  mileage?: number
  diagnosis?: string
  work_performed?: string
  estimated_hours?: number
  actual_hours?: number
  started_at?: string
  completed_at?: string
  notes?: string
  items: ServiceJobItem[]
  total_amount: number
}

export interface ServiceJobItem {
  id: number
  product_id?: number
  item_type: 'product' | 'labor' | 'fee' | 'custom'
  product_name?: string
  description?: string
  quantity: number
  unit_price: number
  total_price: number
}

export interface JobItemInput {
  product_id?: number
  item_type?: string
  description?: string
  quantity: number
  unit_price: number
}

export interface ServiceJobListParams {
  status?: string
  customer_id?: number
  scheduled_from?: string
  assigned_to?: number
  page?: number
  per_page?: number
}

export interface CreateServiceJobRequest {
  customer_id?: number
  vehicle_id?: number
  mileage?: number
  description: string
  priority?: string
  estimated_hours?: number
  scheduled_at?: string
  assigned_to?: number
  notes?: string
  items?: JobItemInput[]
}

export interface UpdateServiceJobRequest {
  status?: string
  priority?: string
  diagnosis?: string
  work_performed?: string
  estimated_hours?: number
  actual_hours?: number
  started_at?: string
  completed_at?: string
  scheduled_at?: string
  assigned_to?: number
  mileage?: number
  notes?: string
}

export interface AddItemRequest {
  product_id?: number
  item_type?: string
  description?: string
  quantity: number
  unit_price: number
}
