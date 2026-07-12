import api from './api'
import type {
  Product,
  ProductListParams,
  PaginatedResponse,
  CreateProductRequest,
  UpdateProductRequest,
  ReceiveStockRequest,
  AdjustStockRequest,
  LowStockProduct,
  StockMovement,
  Batch,
  BatchConsumer,
  ImportResult,
} from '@/types/product'

export const productsApi = {
  list: async (params?: ProductListParams): Promise<PaginatedResponse<Product>> => {
    const res = await api.get('/products', { params })
    return res.data
  },

  get: async (id: number): Promise<Product> => {
    const res = await api.get(`/products/${id}`)
    return res.data.data
  },

  // POS scan: resolve a scanned barcode or SKU to a product, any catalog size.
  lookup: async (code: string): Promise<Product | null> => {
    const res = await api.get('/products', { params: { code, per_page: 1 } })
    return res.data.data?.[0] || null
  },

  import: async (file: File): Promise<ImportResult> => {
    const form = new FormData()
    form.append('file', file)
    const res = await api.post('/products/import', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data.data
  },

  create: async (data: CreateProductRequest): Promise<Product> => {
    const res = await api.post('/products', data)
    return res.data.data
  },

  update: async (id: number, data: UpdateProductRequest): Promise<Product> => {
    const res = await api.put(`/products/${id}`, data)
    return res.data.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/products/${id}`)
  },

  getLowStock: async (): Promise<LowStockProduct[]> => {
    const res = await api.get('/products/low-stock')
    return res.data.data
  },

  receiveStock: async (id: number, data: ReceiveStockRequest): Promise<Product> => {
    const res = await api.post(`/products/${id}/receive`, data)
    return res.data.data
  },

  adjustStock: async (id: number, data: AdjustStockRequest): Promise<Product> => {
    const res = await api.post(`/products/${id}/adjust`, data)
    return res.data.data
  },

  movements: async (id: number, page = 1): Promise<PaginatedResponse<StockMovement>> => {
    const res = await api.get(`/products/${id}/movements`, { params: { page } })
    return res.data
  },

  batches: async (id: number): Promise<Batch[]> => {
    const res = await api.get(`/products/${id}/batches`)
    return res.data.data
  },

  batchConsumers: async (batchId: number): Promise<BatchConsumer[]> => {
    const res = await api.get(`/batches/${batchId}/consumers`)
    return res.data.data
  },

  uploadImage: async (id: number, file: File): Promise<Product> => {
    const form = new FormData()
    form.append('image', file)
    const res = await api.post(`/products/${id}/image`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data.data
  },

  deleteImage: async (id: number): Promise<Product> => {
    const res = await api.delete(`/products/${id}/image`)
    return res.data.data
  },
}
