import api from './api'
import type { CreateStocktakeRequest, StocktakeListItem, StocktakeDetail, StocktakeItem, FinalizeResult } from '@/types/stocktake'

export const stocktakesApi = {
  list: async (): Promise<StocktakeListItem[]> => {
    const res = await api.get('/stocktakes')
    return res.data.data
  },

  get: async (id: number): Promise<StocktakeDetail> => {
    const res = await api.get(`/stocktakes/${id}`)
    return res.data.data
  },

  create: async (data: CreateStocktakeRequest): Promise<StocktakeListItem> => {
    const res = await api.post('/stocktakes', data)
    return res.data.data
  },

  addItem: async (stocktakeId: number, productId: number): Promise<StocktakeItem> => {
    const res = await api.post(`/stocktakes/${stocktakeId}/items`, { product_id: productId })
    return res.data.data
  },

  setCount: async (itemId: number, countedQty: number): Promise<StocktakeItem> => {
    const res = await api.put(`/stocktakes/items/${itemId}`, { counted_qty: countedQty })
    return res.data.data
  },

  removeItem: async (itemId: number): Promise<void> => {
    await api.delete(`/stocktakes/items/${itemId}`)
  },

  cancel: async (id: number): Promise<void> => {
    await api.post(`/stocktakes/${id}/cancel`)
  },

  finalize: async (id: number): Promise<FinalizeResult> => {
    const res = await api.post(`/stocktakes/${id}/finalize`)
    return res.data.data
  },
}
