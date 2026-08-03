export interface BackupSchedule {
  id: number
  name: string
  cron: string
  enabled: boolean
  retention_days: number
  last_run_at: string | null
  last_status: 'never' | 'success' | 'error'
  last_error?: string
  next_run_at: string | null
  latest_file?: string
  created_at: string
}

export interface BackupScheduleInput {
  name: string
  cron: string
  enabled: boolean
  retention_days: number
}
