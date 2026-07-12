import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { auditApi, type AuditLogParams } from '@/services/audit'

export function useAuditLog(params: AuditLogParams) {
  return useQuery({
    queryKey: ['audit-logs', params],
    queryFn: () => auditApi.list(params),
    placeholderData: keepPreviousData,
  })
}
