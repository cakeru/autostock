import api from './api'
import type { CreatePORequest, AddPOItemRequest, ReceiveRequest, ReceiveResult, POListItem, PODetail, POItem } from '@/types/purchaseorder'

export const purchaseOrdersApi = {
  list: async (): Promise<POListItem[]> => {
    const res = await api.get('/purchase-orders')
    return res.data.data
  },

  get: async (id: number): Promise<PODetail> => {
    const res = await api.get(`/purchase-orders/${id}`)
    return res.data.data
  },

  create: async (data: CreatePORequest): Promise<POListItem> => {
    const res = await api.post('/purchase-orders', data)
    return res.data.data
  },

  addItem: async (poId: number, data: AddPOItemRequest): Promise<POItem> => {
    const res = await api.post(`/purchase-orders/${poId}/items`, data)
    return res.data.data
  },

  removeItem: async (itemId: number): Promise<void> => {
    await api.delete(`/purchase-orders/items/${itemId}`)
  },

  place: async (id: number): Promise<POListItem> => {
    const res = await api.post(`/purchase-orders/${id}/place`)
    return res.data.data
  },

  cancel: async (id: number): Promise<void> => {
    await api.post(`/purchase-orders/${id}/cancel`)
  },

  receive: async (id: number, data: ReceiveRequest): Promise<ReceiveResult> => {
    const res = await api.post(`/purchase-orders/${id}/receive`, data)
    return res.data.data
  },
}
