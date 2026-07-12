import api from './api'
import type { Employee, CreateEmployeeRequest, UpdateEmployeeRequest, CreateAccountRequest } from '@/types/employee'

export const employeesApi = {
  list: async (): Promise<Employee[]> => {
    const res = await api.get('/employees')
    return res.data.data
  },

  get: async (id: number): Promise<Employee> => {
    const res = await api.get(`/employees/${id}`)
    return res.data.data
  },

  create: async (data: CreateEmployeeRequest): Promise<Employee> => {
    const res = await api.post('/employees', data)
    return res.data.data
  },

  update: async (id: number, data: UpdateEmployeeRequest): Promise<Employee> => {
    const res = await api.put(`/employees/${id}`, data)
    return res.data.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/employees/${id}`)
  },

  createAccount: async (id: number, data: CreateAccountRequest): Promise<Employee> => {
    const res = await api.post(`/employees/${id}/create-account`, data)
    return res.data.data
  },
}
