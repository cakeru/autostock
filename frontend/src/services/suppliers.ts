import api from './api'
import type { Supplier, Purchase, CreateSupplierRequest, UpdateSupplierRequest } from '@/types/supplier'

export const suppliersApi = {
  list: (): Promise<Supplier[]> => api.get('/suppliers').then(r => r.data.data),
  get: (id: number): Promise<Supplier> => api.get(`/suppliers/${id}`).then(r => r.data.data),
  purchases: (id: number): Promise<Purchase[]> => api.get(`/suppliers/${id}/purchases`).then(r => r.data.data),
  create: (data: CreateSupplierRequest): Promise<Supplier> => api.post('/suppliers', data).then(r => r.data.data),
  update: (id: number, data: UpdateSupplierRequest): Promise<Supplier> => api.put(`/suppliers/${id}`, data).then(r => r.data.data),
  remove: (id: number) => api.delete(`/suppliers/${id}`),
  pay: (id: number, data: { invoice_ids: number[]; batch_ids: number[] }): Promise<Supplier> => api.post(`/suppliers/${id}/pay`, data).then(r => r.data.data),
}
