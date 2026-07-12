import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { toast } from 'sonner'
import { customersApi } from '@/services/customers'
import type { CreateCustomerRequest, CreateVehicleRequest, UpdateCustomerRequest, CustomerListParams, UpdateVehicleRequest } from '@/types/customer'

export function useCustomers(params?: CustomerListParams) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: () => customersApi.list(params),
    placeholderData: keepPreviousData,
  })
}

export function useCustomer(id: number) {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: () => customersApi.get(id),
    enabled: id > 0,
  })
}

export function useCreateCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateCustomerRequest) => customersApi.create(data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['customers'] }); toast.success('Customer created') },
  })
}

export function useUpdateCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateCustomerRequest }) => customersApi.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['customers'] }); qc.invalidateQueries({ queryKey: ['customer'] }); toast.success('Customer updated') },
  })
}

export function useDeleteCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => customersApi.delete(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['customers'] }); toast.success('Customer deleted') },
  })
}

export function useServiceHistory(customerId: number) {
  return useQuery({
    queryKey: ['customer', customerId, 'history'],
    queryFn: () => customersApi.getHistory(customerId),
    enabled: customerId > 0,
  })
}

export function useCustomerVehicles(customerId: number) {
  return useQuery({
    queryKey: ['customer', customerId, 'vehicles'],
    queryFn: () => customersApi.listVehicles(customerId),
    enabled: customerId > 0,
  })
}

export function useCreateVehicle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ customerId, data }: { customerId: number; data: CreateVehicleRequest }) =>
      customersApi.createVehicle(customerId, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['customer', null, 'vehicles'] }); toast.success('Vehicle added') },
  })
}

export function useUpdateVehicle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateVehicleRequest }) => customersApi.updateVehicle(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['customer'] }); toast.success('Vehicle updated') },
  })
}

export function useDeleteVehicle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => customersApi.deleteVehicle(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['customer'] }); toast.success('Vehicle removed') },
  })
}
