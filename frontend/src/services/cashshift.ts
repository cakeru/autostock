import api from './api'

export interface CashShift {
  id: number
  status: 'open' | 'closed'
  opened_by_id?: number
  opened_by_name?: string
  opening_float: number
  opened_at: string
  cash_sales: number
  expected_amount: number
  closing_amount?: number
  over_short?: number
  note?: string
  closed_by_name?: string
  closed_at?: string
}

export const cashShiftApi = {
  current: (): Promise<CashShift | null> => api.get('/cash-shifts/current').then(r => r.data.data),
  list: (limit = 10): Promise<CashShift[]> => api.get('/cash-shifts', { params: { limit } }).then(r => r.data.data),
  open: (opening_float: number): Promise<CashShift> => api.post('/cash-shifts/open', { opening_float }).then(r => r.data.data),
  close: (closing_amount: number, note?: string): Promise<CashShift> => api.post('/cash-shifts/close', { closing_amount, note }).then(r => r.data.data),
}
