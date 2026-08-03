import api from './api'
import type { BackupSchedule, BackupScheduleInput } from '@/types/backup'

export const backupsApi = {
  list: async (): Promise<BackupSchedule[]> => {
    const res = await api.get('/backup-schedules')
    return res.data.data || []
  },
  create: async (input: BackupScheduleInput): Promise<BackupSchedule> => {
    const res = await api.post('/backup-schedules', input)
    return res.data.data
  },
  update: async (id: number, input: BackupScheduleInput): Promise<BackupSchedule> => {
    const res = await api.put(`/backup-schedules/${id}`, input)
    return res.data.data
  },
  remove: async (id: number): Promise<void> => {
    await api.delete(`/backup-schedules/${id}`)
  },
  runNow: async (id: number): Promise<BackupSchedule> => {
    const res = await api.post(`/backup-schedules/${id}/run`)
    return res.data.data
  },
}
