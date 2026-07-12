import api from './api'

export interface Deposit {
  id: number
  customer_id: number
  customer_name?: string
  amount: number
  note?: string
  status: 'held' | 'applied' | 'refunded'
  invoice_id?: number
  invoice_number?: string
  created_at: string
  settled_at?: string
}

export const depositsApi = {
  list: (params?: { customer_id?: number; status?: string }): Promise<Deposit[]> =>
    api.get('/deposits', { params }).then(r => r.data.data),
  create: (data: { customer_id: number; amount: number; note?: string }): Promise<Deposit> =>
    api.post('/deposits', data).then(r => r.data.data),
  apply: (id: number, invoice_id: number): Promise<Deposit> =>
    api.post(`/deposits/${id}/apply`, { invoice_id }).then(r => r.data.data),
  refund: (id: number): Promise<Deposit> =>
    api.post(`/deposits/${id}/refund`).then(r => r.data.data),
}
