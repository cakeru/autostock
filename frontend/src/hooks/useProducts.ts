import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { toast } from 'sonner'
import { productsApi } from '@/services/products'
import type { ProductListParams, CreateProductRequest, UpdateProductRequest, ReceiveStockRequest, AdjustStockRequest } from '@/types/product'

export function useProducts(params?: ProductListParams) {
  return useQuery({
    queryKey: ['products', params],
    queryFn: () => productsApi.list(params),
    placeholderData: keepPreviousData,
  })
}

export function useProduct(id: number) {
  return useQuery({
    queryKey: ['product', id],
    queryFn: () => productsApi.get(id),
    enabled: id > 0,
  })
}

export function useCreateProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateProductRequest) => productsApi.create(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['products'] }); toast.success('Product created') },
  })
}

export function useUpdateProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateProductRequest }) => productsApi.update(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['products'] }); toast.success('Product updated') },
  })
}

export function useDeleteProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => productsApi.delete(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['products'] }); toast.success('Product deleted') },
  })
}

export function useProductMovements(id: number, page: number) {
  return useQuery({
    queryKey: ['product-movements', id, page],
    queryFn: () => productsApi.movements(id, page),
    enabled: id > 0,
    placeholderData: keepPreviousData,
  })
}

export function useProductBatches(id: number, enabled = true) {
  return useQuery({
    queryKey: ['product-batches', id],
    queryFn: () => productsApi.batches(id),
    enabled: id > 0 && enabled,
  })
}

export function useBatchConsumers(batchId: number | null) {
  return useQuery({
    queryKey: ['batch-consumers', batchId],
    queryFn: () => productsApi.batchConsumers(batchId as number),
    enabled: !!batchId,
  })
}

export function useUploadProductImage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, file }: { id: number; file: File }) => productsApi.uploadImage(id, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] })
      toast.success('Image updated')
    },
  })
}

export function useDeleteProductImage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => productsApi.deleteImage(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] })
      toast.success('Image removed')
    },
  })
}

export function useLowStockProducts() {
  return useQuery({
    queryKey: ['products', 'low-stock'],
    queryFn: () => productsApi.getLowStock(),
  })
}

export function useReceiveStock() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: ReceiveStockRequest }) => productsApi.receiveStock(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['products'] }); toast.success('Stock received') },
  })
}

export function useAdjustStock() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: AdjustStockRequest }) => productsApi.adjustStock(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['products'] }); toast.success('Stock adjusted') },
  })
}

export function useImportProducts() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => productsApi.import(file),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['products'] })
      if (result.failed === 0) {
        toast.success(`Imported ${result.created + result.updated} products (${result.created} new, ${result.updated} updated)`)
      } else {
        toast.warning(`Imported ${result.created + result.updated}, ${result.failed} row(s) failed — see details`)
      }
    },
  })
}
