import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { settingsApi } from '@/services/settings'

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: () => settingsApi.get(),
  })
}

export function useUpdateSetting() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ key, value }: { key: string; value: string }) => settingsApi.update(key, value),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })
}

export function useBatchUpdateSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (updates: { key: string; value: string }[]) =>
      Promise.all(updates.map((u) => settingsApi.update(u.key, u.value))),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })
}

export function useUpdateExchangeRate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (rate: number) => settingsApi.updateExchangeRate(rate),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })
}
