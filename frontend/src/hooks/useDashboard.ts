import { useQuery } from '@tanstack/react-query'
import { dashboardApi } from '@/services/dashboard'

export function useDashboardSummary() {
  return useQuery({
    queryKey: ['dashboard-summary'],
    queryFn: () => dashboardApi.getSummary(),
    refetchInterval: 60_000,
  })
}

export function useDailyRevenue() {
  return useQuery({
    queryKey: ['dashboard-daily-revenue'],
    queryFn: () => dashboardApi.getDailyRevenue(),
    refetchInterval: 60_000,
  })
}

export function useDayClose(date: string) {
  return useQuery({
    queryKey: ['dashboard-day-close', date],
    queryFn: () => dashboardApi.getDayClose(date),
    enabled: !!date,
  })
}
