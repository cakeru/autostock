import api from './api'

export interface DashboardSummary {
  today_revenue: number
  today_jobs: number
  total_customers: number
  low_stock_count: number
  unpaid_count: number
  outstanding_usd: number
  recent_invoices: Array<{
    id: number
    invoice_number: string
    customer_name: string
    total_usd: number
    status: string
  }>
  recent_jobs: Array<{
    id: number
    job_number: string
    customer_name: string
    status: string
  }>
}

export interface DailyRevenueItem {
  date: string
  revenue_usd: number
  invoice_count: number
}

export interface PaymentMethodTotal {
  method: string
  count: number
  total: number
}

export interface UserPaymentTotal {
  user_id: number
  user_name: string
  count: number
  total: number
}

export interface DayCloseSummary {
  date: string
  payment_methods: PaymentMethodTotal[]
  voided_count: number
  voided_total: number
  user_totals: UserPaymentTotal[]
}

export const dashboardApi = {
  getSummary: async (): Promise<DashboardSummary> => {
    const res = await api.get('/dashboard/summary')
    return res.data.data
  },

  getDailyRevenue: async (): Promise<DailyRevenueItem[]> => {
    const res = await api.get('/dashboard/daily-revenue')
    return res.data.data
  },

  getDayClose: async (date: string): Promise<DayCloseSummary> => {
    const res = await api.get('/dashboard/day-close', { params: { date } })
    return res.data.data
  },
}
