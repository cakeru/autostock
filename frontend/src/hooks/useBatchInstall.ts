import { useQuery, useMutation } from '@tanstack/react-query'
import { batchInstallApi, type RecordInstallRequest } from '@/services/batchInstall'

export function useOpenJobs(enabled = true) {
  return useQuery({
    queryKey: ['batch-install-open-jobs'],
    queryFn: () => batchInstallApi.openJobs(),
    enabled,
  })
}

export function useMechanics(enabled = true) {
  return useQuery({
    queryKey: ['batch-install-mechanics'],
    queryFn: () => batchInstallApi.mechanics(),
    enabled,
  })
}

export function useResolveBatch() {
  return useMutation({ mutationFn: (code: string) => batchInstallApi.resolve(code) })
}

export function useRecordInstall() {
  return useMutation({ mutationFn: (data: RecordInstallRequest) => batchInstallApi.record(data) })
}
