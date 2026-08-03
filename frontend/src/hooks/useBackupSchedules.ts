import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { backupsApi } from '@/services/backups'
import type { BackupScheduleInput } from '@/types/backup'

export function useBackupSchedules() {
  return useQuery({
    queryKey: ['backup-schedules'],
    queryFn: () => backupsApi.list(),
  })
}

function useInvalidate() {
  const qc = useQueryClient()
  return () => qc.invalidateQueries({ queryKey: ['backup-schedules'] })
}

export function useCreateBackupSchedule() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: (input: BackupScheduleInput) => backupsApi.create(input),
    onSuccess: invalidate,
  })
}

export function useUpdateBackupSchedule() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: BackupScheduleInput }) =>
      backupsApi.update(id, input),
    onSuccess: invalidate,
  })
}

export function useDeleteBackupSchedule() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: (id: number) => backupsApi.remove(id),
    onSuccess: invalidate,
  })
}

export function useRunBackupSchedule() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: (id: number) => backupsApi.runNow(id),
    onSuccess: invalidate,
  })
}
