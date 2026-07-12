import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { cashShiftApi } from '@/services/cashshift'

export function useCurrentShift() {
  return useQuery({
    queryKey: ['cash-shift', 'current'],
    queryFn: () => cashShiftApi.current(),
    refetchInterval: 60000, // keep the live cash total fresh
  })
}

export function useShiftHistory(limit = 10) {
  return useQuery({
    queryKey: ['cash-shift', 'history', limit],
    queryFn: () => cashShiftApi.list(limit),
  })
}

export function useOpenShift() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (float: number) => cashShiftApi.open(float),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cash-shift'] }),
  })
}

export function useCloseShift() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ amount, note }: { amount: number; note?: string }) => cashShiftApi.close(amount, note),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cash-shift'] }),
  })
}
