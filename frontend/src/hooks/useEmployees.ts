import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { employeesApi } from '@/services/employees'
import type { CreateEmployeeRequest, UpdateEmployeeRequest, CreateAccountRequest } from '@/types/employee'

export function useEmployees() {
  return useQuery({
    queryKey: ['employees'],
    queryFn: () => employeesApi.list(),
  })
}

export function useEmployee(id: number) {
  return useQuery({
    queryKey: ['employee', id],
    queryFn: () => employeesApi.get(id),
    enabled: id > 0,
  })
}

export function useCreateEmployee() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateEmployeeRequest) => employeesApi.create(data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['employees'] }); toast.success('Employee created') },
  })
}

export function useUpdateEmployee() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateEmployeeRequest }) => employeesApi.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['employees'] }); toast.success('Employee updated') },
  })
}

export function useDeleteEmployee() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => employeesApi.delete(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['employees'] }); toast.success('Employee deactivated') },
  })
}

export function useCreateEmployeeAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: CreateAccountRequest }) => employeesApi.createAccount(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['employees'] }); toast.success('Login account created') },
  })
}
