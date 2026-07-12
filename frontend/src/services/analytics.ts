import api from './api'

export interface SalesPoint { period: string; revenue_usd: number; invoice_count: number }
export interface PaymentMethodTotal { method: string; count: number; total: number }
export interface SalesReport {
  from: string
  to: string
  granularity: string
  revenue_usd: number
  invoice_count: number
  avg_ticket: number
  tax_collected: number
  prev_revenue_usd: number
  revenue_change_pct: number
  series: SalesPoint[]
  payment_methods: PaymentMethodTotal[]
}

export interface ARBucket { label: string; count: number; outstanding_usd: number }
export interface ARCustomer {
  customer_id: number | null
  customer_name: string
  invoice_count: number
  outstanding_usd: number
  oldest_days: number
}
export interface ReceivablesReport {
  total_outstanding: number
  buckets: ARBucket[]
  customers: ARCustomer[]
}

export interface InventoryValuation {
  cost_value: number
  retail_value: number
  potential_profit: number
  sku_count: number
  units_on_hand: number
}
export interface ProductStat {
  product_id: number
  name: string
  type: string
  qty_sold: number
  revenue_usd: number
  profit_usd: number
  stock_qty: number
}
export interface DeadStockItem {
  product_id: number
  name: string
  stock_qty: number
  cost_value: number
  last_sold?: string
}
export interface ReorderItem {
  product_id: number
  name: string
  stock_qty: number
  min_stock: number
  daily_rate: number
  days_left?: number
}
export interface InventoryReport {
  window_days: number
  valuation: InventoryValuation
  turnover_ratio: number
  top_sellers: ProductStat[]
  dead_stock: DeadStockItem[]
  reorder: ReorderItem[]
}

export interface CustomerStat {
  customer_id: number
  name: string
  invoice_count: number
  total_spent: number
  last_visit?: string
  days_since_visit?: number
}
export interface CustomersReport {
  window_days: number
  total_customers: number
  new_customers: number
  repeat_customers: number
  churn_risk_count: number
  avg_spend_per_visit: number
  top_customers: CustomerStat[]
  churn_risk: CustomerStat[]
}

export interface PnLExpenseCategory { category: string; amount: number }
export interface PnLPayrollLine {
  employee_id: number
  name: string
  pay_type: string
  salary_cost: number
  commission_cost: number
  total: number
}
export interface PnLReport {
  from: string
  to: string
  revenue: number
  returns: number
  cogs: number
  gross_profit: number
  payroll: number
  payroll_breakdown: PnLPayrollLine[]
  expenses: number
  net_profit: number
  gross_margin_pct: number
  net_margin_pct: number
  expense_categories: PnLExpenseCategory[]
}

export interface TechStat {
  employee_id: number
  name: string
  jobs_completed: number
  revenue: number
  hours: number
  salary_cost: number
  commission_cost: number
  payroll_cost: number
}
export interface TechniciansReport {
  from: string
  to: string
  technicians: TechStat[]
}

export const analyticsApi = {
  sales: async (from: string, to: string, granularity = 'day'): Promise<SalesReport> => {
    const res = await api.get('/analytics/sales', { params: { from, to, granularity } })
    return res.data.data
  },
  receivables: async (): Promise<ReceivablesReport> => {
    const res = await api.get('/analytics/receivables')
    return res.data.data
  },
  inventory: async (days = 90): Promise<InventoryReport> => {
    const res = await api.get('/analytics/inventory', { params: { days } })
    return res.data.data
  },
  customers: async (days = 90): Promise<CustomersReport> => {
    const res = await api.get('/analytics/customers', { params: { days } })
    return res.data.data
  },
  pnl: async (from: string, to: string): Promise<PnLReport> => {
    const res = await api.get('/analytics/pnl', { params: { from, to } })
    return res.data.data
  },
  technicians: async (from: string, to: string): Promise<TechniciansReport> => {
    const res = await api.get('/analytics/technicians', { params: { from, to } })
    return res.data.data
  },
}
