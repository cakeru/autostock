import api from './api'

export interface Settings {
  exchange_rate_usd_khr: number
  tax_rate_percent: number
  tax_enabled: boolean
  invoice_prefix: string
  low_stock_threshold: number
  sale_packages?: string
  labor_presets?: string
  fee_presets?: string
  payment_methods?: string
  shop_name: string
  shop_address?: string
  shop_phone?: string
  shop_email?: string
  feature_batch_scan?: boolean
}

export interface ProfitCategory {
  category: string
  revenue: number
  cost: number
  profit: number
  margin_pct: number
}

export interface ProfitReport {
  from: string
  to: string
  gross_revenue: number
  discounts: number
  revenue: number
  cost: number
  gross_profit: number
  margin_pct: number
  invoice_count: number
  categories: ProfitCategory[]
}

export const reportsApi = {
  profit: async (from: string, to: string): Promise<ProfitReport> => {
    const res = await api.get('/dashboard/profit', { params: { from, to } })
    return res.data.data
  },
}

export const settingsApi = {
  get: async (): Promise<Settings> => {
    const res = await api.get('/settings')
    return res.data.data
  },

  update: async (key: string, value: string): Promise<Settings> => {
    const res = await api.put('/settings', { key, value })
    return res.data.data
  },

  updateExchangeRate: async (rate: number) => {
    const res = await api.put('/settings/exchange-rate', { rate })
    return res.data.data
  },
}
