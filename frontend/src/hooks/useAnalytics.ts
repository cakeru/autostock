import { useQuery } from '@tanstack/react-query'
import { analyticsApi } from '@/services/analytics'

export function useSalesReport(from: string, to: string, granularity: string) {
  return useQuery({
    queryKey: ['analytics', 'sales', from, to, granularity],
    queryFn: () => analyticsApi.sales(from, to, granularity),
  })
}

export function useReceivables() {
  return useQuery({
    queryKey: ['analytics', 'receivables'],
    queryFn: () => analyticsApi.receivables(),
  })
}

export function useInventoryReport(days: number) {
  return useQuery({
    queryKey: ['analytics', 'inventory', days],
    queryFn: () => analyticsApi.inventory(days),
  })
}

export function useCustomersReport(days: number) {
  return useQuery({
    queryKey: ['analytics', 'customers', days],
    queryFn: () => analyticsApi.customers(days),
  })
}

export function usePnL(from: string, to: string) {
  return useQuery({
    queryKey: ['analytics', 'pnl', from, to],
    queryFn: () => analyticsApi.pnl(from, to),
  })
}

export function useTechnicians(from: string, to: string) {
  return useQuery({
    queryKey: ['analytics', 'technicians', from, to],
    queryFn: () => analyticsApi.technicians(from, to),
  })
}
