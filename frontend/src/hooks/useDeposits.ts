import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { depositsApi } from '@/services/deposits'

export function useCustomerDeposits(customerId: number, status?: string) {
  return useQuery({
    queryKey: ['deposits', customerId, status],
    queryFn: () => depositsApi.list({ customer_id: customerId, status }),
    enabled: customerId > 0,
  })
}

function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['deposits'] })
  qc.invalidateQueries({ queryKey: ['customer'] })
  qc.invalidateQueries({ queryKey: ['invoices'] })
  qc.invalidateQueries({ queryKey: ['invoice'] })
}

export function useCreateDeposit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { customer_id: number; amount: number; note?: string }) => depositsApi.create(data),
    onSuccess: () => { invalidate(qc); toast.success('Deposit recorded') },
  })
}

export function useApplyDeposit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, invoiceId }: { id: number; invoiceId: number }) => depositsApi.apply(id, invoiceId),
    onSuccess: () => { invalidate(qc); toast.success('Deposit applied to invoice') },
  })
}

export function useRefundDeposit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => depositsApi.refund(id),
    onSuccess: () => { invalidate(qc); toast.success('Deposit refunded') },
  })
}
