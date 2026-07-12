import api from './api'

export interface ReturnItem {
  invoice_item_id: number
  product_id?: number
  description: string
  quantity: number
  unit_price: number
  total: number
}

export interface Return {
  id: number
  invoice_id: number
  refund_amount: number
  refund_method: 'cash' | 'store_credit'
  reason?: string
  created_by_name?: string
  created_at: string
  items: ReturnItem[]
}

export interface InvoiceReturns {
  returns: Return[]
  returned_by_item: Record<string, number>
}

export interface CreateReturnRequest {
  invoice_id: number
  refund_method: 'cash' | 'store_credit'
  reason?: string
  items: { invoice_item_id: number; quantity: number }[]
}

export const returnsApi = {
  forInvoice: (invoiceId: number): Promise<InvoiceReturns> => api.get(`/invoices/${invoiceId}/returns`).then(r => r.data.data),
  create: (data: CreateReturnRequest): Promise<Return> => api.post('/returns', data).then(r => r.data.data),
}
