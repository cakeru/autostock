import api from './api'
import type { CreateInvoiceRequest, Invoice, InvoiceDetail, InvoiceListParams, UpdateInvoiceRequest, RecordPaymentRequest, Payment } from '@/types/invoice'

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
  listPayments: (id: number): Promise<Payment[]> =>
    api.get(`/invoices/${id}/payments`).then(r => r.data.data),
}
