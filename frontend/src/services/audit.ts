import api from './api'

export interface AuditLogItem {
  id: number
  action: string
  entity_type: string
  entity_id?: number
  user_id?: number
  user_name: string
  created_at: string
}

export interface AuditLogParams {
  action?: string
  entity_type?: string
  user_id?: number
  from?: string
  to?: string
  page?: number
  per_page?: number
}

export interface AuditLogPage {
  data: AuditLogItem[]
  meta: { page: number; per_page: number; total: number; total_pages: number }
}

export const auditApi = {
  list: (params?: AuditLogParams): Promise<AuditLogPage> => api.get('/audit-logs', { params }).then(r => r.data),
}
