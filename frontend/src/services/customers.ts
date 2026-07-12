import api from './api'
import type {
  Customer, CustomerDetail, CustomerListParams, ActivityItem,
  CreateCustomerRequest, UpdateCustomerRequest,
  CreateVehicleRequest, UpdateVehicleRequest, Vehicle,
} from '@/types/customer'

export const customersApi = {
  list: (params?: CustomerListParams) => api.get('/customers', { params }).then(r => r.data),
  get: (id: number): Promise<CustomerDetail> => api.get(`/customers/${id}`).then(r => r.data.data),
  create: (data: CreateCustomerRequest): Promise<Customer> => api.post('/customers', data).then(r => r.data.data),
  update: (id: number, data: UpdateCustomerRequest): Promise<Customer> => api.put(`/customers/${id}`, data).then(r => r.data.data),
  delete: (id: number) => api.delete(`/customers/${id}`),
  getHistory: (id: number): Promise<ActivityItem[]> => api.get(`/customers/${id}/history`).then(r => r.data.data),
  listVehicles: (customerId: number): Promise<Vehicle[]> => api.get(`/customers/${customerId}/vehicles`).then(r => r.data.data),
  createVehicle: (customerId: number, data: CreateVehicleRequest): Promise<Vehicle> => api.post(`/customers/${customerId}/vehicles`, data).then(r => r.data.data),
  updateVehicle: (id: number, data: UpdateVehicleRequest): Promise<Vehicle> => api.put(`/vehicles/${id}`, data).then(r => r.data.data),
  deleteVehicle: (id: number) => api.delete(`/vehicles/${id}`),
}
