export interface CreatePORequest {
  supplier_id: number
  notes?: string
}

export interface AddPOItemRequest {
  product_id: number
  quantity_ordered: number
  unit_cost: number
}

export interface ReceiveLine {
  item_id: number
  quantity?: number
}

export interface ReceiveInvoice {
  invoice_number?: string
  invoice_image?: string
  amount: number
}

export interface ReceiveRequest {
  items?: ReceiveLine[]
  paid?: boolean
  invoices?: ReceiveInvoice[]
}

export interface ReceiveResult {
  received: number
  status: string
}

export type POStatus = 'draft' | 'ordered' | 'partial' | 'received' | 'cancelled'

export interface POListItem {
  id: number
  po_number: string
  status: POStatus
  supplier_id: number
  supplier_name: string
  notes?: string
  item_count: number
  total_cost: number
  created_by_id?: number
  created_by_name?: string
  ordered_at?: string
  received_at?: string
  created_at: string
}

export interface POItem {
  id: number
  product_id: number
  sku: string
  product_name: string
  quantity_ordered: number
  quantity_received: number
  unit_cost: number
  total_cost: number
}

export interface PODetail extends POListItem {
  items: POItem[]
}
