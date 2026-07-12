import api from './api'

export interface Expense {
  id: number
  category: string
  description: string
  amount_usd: number
  spent_at: string
  created_by?: number
  created_by_name?: string
  created_at: string
}

export interface CreateExpenseRequest {
  category: string
  description?: string
  amount_usd: number
  spent_at?: string
}

export const expensesApi = {
  list: async (from?: string, to?: string): Promise<Expense[]> => {
    const res = await api.get('/expenses', { params: { from, to } })
    return res.data.data
  },
  create: async (data: CreateExpenseRequest): Promise<Expense> => {
    const res = await api.post('/expenses', data)
    return res.data.data
  },
  remove: async (id: number): Promise<void> => {
    await api.delete(`/expenses/${id}`)
  },
}
