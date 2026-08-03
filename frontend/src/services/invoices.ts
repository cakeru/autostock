import api from './api'
import type { CreateInvoiceRequest, Invoice, InvoiceDetail, InvoiceListParams, UpdateInvoiceRequest, RecordPaymentRequest, Payment, UpdateInvoiceItemRequest } from '@/types/invoice'

export const invoicesApi = {
  list: (params?: InvoiceListParams) => api.get('/invoices', { params }).then(r => r.data),
  get: (id: number): Promise<InvoiceDetail> => api.get(`/invoices/${id}`).then(r => r.data.data),
  create: (data: CreateInvoiceRequest): Promise<Invoice> => api.post('/invoices', data).then(r => r.data.data),
  createFromJob: (jobId: number, data: any): Promise<Invoice> =>
    api.post(`/service-jobs/${jobId}/invoice`, data).then(r => r.data.data),
  update: (id: number, data: UpdateInvoiceRequest): Promise<InvoiceDetail> =>
    api.put(`/invoices/${id}`, data).then(r => r.data.data),
  void: (id: number, reason: string): Promise<InvoiceDetail> =>
    api.post(`/invoices/${id}/void`, { reason }).then(r => r.data.data),
  recordPayment: (id: number, data: RecordPaymentRequest): Promise<Payment> =>
    api.post(`/invoices/${id}/payments`, data).then(r => r.data.data),
  updatePayment: (id: number, paymentId: number, data: RecordPaymentRequest): Promise<Payment> =>
    api.put(`/invoices/${id}/payments/${paymentId}`, data).then(r => r.data.data),
  deletePayment: (id: number, paymentId: number) =>
    api.delete(`/invoices/${id}/payments/${paymentId}`),
  listPayments: (id: number): Promise<Payment[]> =>
    api.get(`/invoices/${id}/payments`).then(r => r.data.data),
  addItem: (id: number, data: { product_id?: number; item_type: string; description: string; quantity: number; unit_price_usd: number }): Promise<any> =>
    api.post(`/invoices/${id}/items`, data).then(r => r.data.data),
  updateItem: (id: number, itemId: number, data: UpdateInvoiceItemRequest): Promise<any> =>
    api.put(`/invoices/${id}/items/${itemId}`, data).then(r => r.data.data),
  removeItem: (id: number, itemId: number) =>
    api.delete(`/invoices/${id}/items/${itemId}`),
  uploadPaymentProof: (invoiceId: number, paymentId: number, file: File): Promise<{ proof_url: string }> => {
    const form = new FormData()
    form.append('photo', file)
    return api.post(`/invoices/${invoiceId}/payments/${paymentId}/proof`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then(r => r.data.data)
  },
}
