import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { returnsApi, type CreateReturnRequest } from '@/services/returns'

export function useInvoiceReturns(invoiceId: number) {
  return useQuery({
    queryKey: ['invoice', invoiceId, 'returns'],
    queryFn: () => returnsApi.forInvoice(invoiceId),
    enabled: invoiceId > 0,
  })
}

export function useCreateReturn() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateReturnRequest) => returnsApi.create(data),
    onSuccess: (_res, vars) => {
      qc.invalidateQueries({ queryKey: ['invoice', vars.invoice_id] })
      qc.invalidateQueries({ queryKey: ['invoices'] })
      qc.invalidateQueries({ queryKey: ['deposits'] })
      qc.invalidateQueries({ queryKey: ['products'] })
      toast.success('Return processed')
    },
  })
}

export function useUndoReturn() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, invoiceId }: { id: number; invoiceId: number }) => returnsApi.remove(id),
    onSuccess: (_res, { invoiceId }) => {
      qc.invalidateQueries({ queryKey: ['invoice', invoiceId] })
      qc.invalidateQueries({ queryKey: ['invoices'] })
      qc.invalidateQueries({ queryKey: ['deposits'] })
      qc.invalidateQueries({ queryKey: ['products'] })
      toast.success('Return undone')
    },
  })
}
