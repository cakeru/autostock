import api from './api'
import type { AddItemRequest, CreateServiceJobRequest, ServiceJob, ServiceJobDetail, ServiceJobListParams, UpdateServiceJobRequest } from '@/types/servicejob'

export const serviceJobsApi = {
  list: (params?: ServiceJobListParams) => api.get('/service-jobs', { params }).then(r => r.data),
  get: (id: number): Promise<ServiceJobDetail> => api.get(`/service-jobs/${id}`).then(r => r.data.data),
  create: (data: CreateServiceJobRequest): Promise<ServiceJob> => api.post('/service-jobs', data).then(r => r.data.data),
  update: (id: number, data: UpdateServiceJobRequest): Promise<ServiceJobDetail> => api.put(`/service-jobs/${id}`, data).then(r => r.data.data),
  delete: (id: number) => api.delete(`/service-jobs/${id}`),
  addItem: (jobId: number, data: AddItemRequest) => api.post(`/service-jobs/${jobId}/items`, data).then(r => r.data.data),
  removeItem: (itemId: number) => api.delete(`/service-jobs/items/${itemId}`),
  complete: (id: number): Promise<ServiceJobDetail> => api.post(`/service-jobs/${id}/complete`).then(r => r.data.data),
  approveQuote: (id: number): Promise<ServiceJobDetail> => api.post(`/service-jobs/${id}/approve-quote`).then(r => r.data.data),
}
