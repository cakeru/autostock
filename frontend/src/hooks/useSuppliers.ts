import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { suppliersApi } from '@/services/suppliers'
import type { CreateSupplierRequest, UpdateSupplierRequest } from '@/types/supplier'

export function useSuppliers() {
  return useQuery({ queryKey: ['suppliers'], queryFn: () => suppliersApi.list() })
}

export function useSupplier(id: number) {
  return useQuery({ queryKey: ['supplier', id], queryFn: () => suppliersApi.get(id), enabled: id > 0 })
}

export function useSupplierPurchases(id: number) {
  return useQuery({ queryKey: ['supplier', id, 'purchases'], queryFn: () => suppliersApi.purchases(id), enabled: id > 0 })
}

export function useCreateSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateSupplierRequest) => suppliersApi.create(data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['suppliers'] }); toast.success('Supplier created') },
  })
}

export function useUpdateSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateSupplierRequest }) => suppliersApi.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['suppliers'] }); qc.invalidateQueries({ queryKey: ['supplier'] }); toast.success('Supplier updated') },
  })
}

export function useDeleteSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => suppliersApi.remove(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['suppliers'] }); toast.success('Supplier removed') },
  })
}

export function usePaySupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, amount }: { id: number; amount: number }) => suppliersApi.pay(id, amount),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['suppliers'] })
      qc.invalidateQueries({ queryKey: ['supplier'] })
      toast.success('Payment recorded')
    },
  })
}
