import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { stocktakesApi } from '@/services/stocktakes'
import type { CreateStocktakeRequest } from '@/types/stocktake'

export function useStocktakes() {
  return useQuery({
    queryKey: ['stocktakes'],
    queryFn: () => stocktakesApi.list(),
  })
}

export function useStocktake(id: number) {
  return useQuery({
    queryKey: ['stocktake', id],
    queryFn: () => stocktakesApi.get(id),
    enabled: id > 0,
  })
}

export function useCreateStocktake() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateStocktakeRequest) => stocktakesApi.create(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['stocktakes'] }); toast.success('Stocktake created') },
  })
}

export function useAddStocktakeItem(stocktakeId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (productId: number) => stocktakesApi.addItem(stocktakeId, productId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['stocktake', stocktakeId] }),
  })
}

export function useSetStocktakeCount(stocktakeId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ itemId, countedQty }: { itemId: number; countedQty: number }) => stocktakesApi.setCount(itemId, countedQty),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['stocktake', stocktakeId] }),
  })
}

export function useRemoveStocktakeItem(stocktakeId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (itemId: number) => stocktakesApi.removeItem(itemId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['stocktake', stocktakeId] }),
  })
}

export function useCancelStocktake() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => stocktakesApi.cancel(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['stocktakes'] })
      queryClient.invalidateQueries({ queryKey: ['stocktake', id] })
      toast.success('Stocktake cancelled')
    },
  })
}

export function useFinalizeStocktake() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => stocktakesApi.finalize(id),
    onSuccess: (result, id) => {
      queryClient.invalidateQueries({ queryKey: ['stocktakes'] })
      queryClient.invalidateQueries({ queryKey: ['stocktake', id] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
      if (result.skipped === 0) {
        toast.success(`Stocktake finalized — ${result.adjusted} adjusted, ${result.unchanged} unchanged`)
      } else {
        toast.warning(`Finalized with ${result.skipped} skipped — see details`)
      }
    },
  })
}
