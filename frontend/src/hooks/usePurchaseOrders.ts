import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { purchaseOrdersApi } from '@/services/purchaseorders'
import type { CreatePORequest, AddPOItemRequest, ReceiveRequest } from '@/types/purchaseorder'

export function usePurchaseOrders() {
  return useQuery({
    queryKey: ['purchase-orders'],
    queryFn: () => purchaseOrdersApi.list(),
  })
}

export function usePurchaseOrder(id: number) {
  return useQuery({
    queryKey: ['purchase-order', id],
    queryFn: () => purchaseOrdersApi.get(id),
    enabled: id > 0,
  })
}

export function useCreatePurchaseOrder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreatePORequest) => purchaseOrdersApi.create(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['purchase-orders'] }); toast.success('Purchase order created') },
  })
}

export function useUpdatePurchaseOrder(id: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreatePORequest) => purchaseOrdersApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchase-orders'] })
      queryClient.invalidateQueries({ queryKey: ['purchase-order', id] })
      toast.success('Order updated')
    },
  })
}

export function useAddPOItem(poId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: AddPOItemRequest) => purchaseOrdersApi.addItem(poId, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['purchase-order', poId] }),
  })
}

export function useRemovePOItem(poId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (itemId: number) => purchaseOrdersApi.removeItem(itemId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['purchase-order', poId] }),
  })
}

export function usePlacePO() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => purchaseOrdersApi.place(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['purchase-orders'] })
      queryClient.invalidateQueries({ queryKey: ['purchase-order', id] })
      toast.success('Order placed')
    },
  })
}

export function useCancelPO() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => purchaseOrdersApi.cancel(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['purchase-orders'] })
      queryClient.invalidateQueries({ queryKey: ['purchase-order', id] })
      toast.success('Purchase order cancelled')
    },
  })
}

export function useReceivePO() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: ReceiveRequest }) => purchaseOrdersApi.receive(id, data),
    onSuccess: (result, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['purchase-orders'] })
      queryClient.invalidateQueries({ queryKey: ['purchase-order', id] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
      toast.success(`Received ${result.received} unit(s) — order is now ${result.status}`)
    },
  })
}
