export interface Invoice {
  id: number
  invoice_number: string
  status: string
  payment_status: string
  customer_id?: number
  customer_name?: string
  vehicle_id?: number
  plate_number?: string
  service_job_id?: number
  job_number?: string
  subtotal: number
  total_usd: number
  exchange_rate: number
  total_khr: number
  paid_amount: number
  issued_at?: string
  created_at: string
}

export interface InvoiceDetail extends Invoice {
  customer_phone?: string
  customer_address?: string
  vehicle_info?: string
  mileage?: number
  mileage_unit?: string
  created_by_name?: string
  items: InvoiceItem[]
  payments?: Payment[]
  tax_rate: number
  tax_amount: number
  discount: number
  payment_method?: string
  payment_notes?: string
  notes?: string
  voided_at?: string
  void_reason?: string
  updated_at: string
}

export interface InvoiceItem {
  id: number
  product_id?: number
  item_type: string
  description: string
  quantity: number
  unit_price_usd: number
  total_usd: number
}

export interface InvoiceListParams {
  status?: string
  payment_status?: string
  customer_id?: number
  invoice_number?: string
  created_at_gte?: string
  created_at_lte?: string
  page?: number
  per_page?: number
}

export interface CreateInvoiceRequest {
  customer_id?: number
  vehicle_id?: number
  service_job_id?: number
  mileage?: number
  items: CreateInvoiceItemRequest[]
  discount?: number
  tax_rate?: number
  exchange_rate: number
  payment_method?: string
  payment_status?: string
  notes?: string
}

export interface CreateInvoiceItemRequest {
  product_id?: number
  item_type: string
  description: string
  quantity: number
  unit_price_usd: number
}

export interface UpdateInvoiceRequest {
  payment_status?: string
  payment_method?: string
  payment_notes?: string
  notes?: string
  vehicle_id?: number
  mileage?: number
  clear_vehicle?: boolean
}

export interface Payment {
  id: number
  invoice_id: number
  amount: number
  method: string
  currency?: string
  tendered_amount?: number
  received_by_name?: string
  notes?: string
  reference?: string
  proof_url?: string
  created_at: string
}

export interface RecordPaymentRequest {
  amount: number
  method: string
  notes?: string
  currency?: string
  tendered_amount?: number
  exchange_rate?: number
  reference?: string
}

export interface UpdateInvoiceItemRequest {
  description?: string
  quantity?: number
  unit_price_usd?: number
}
