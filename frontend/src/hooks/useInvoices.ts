import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { toast } from 'sonner'
import { invoicesApi } from '@/services/invoices'
import type { CreateInvoiceRequest, InvoiceListParams, UpdateInvoiceRequest, RecordPaymentRequest } from '@/types/invoice'

export function useInvoices(params?: InvoiceListParams) {
  return useQuery({ queryKey: ['invoices', params], queryFn: () => invoicesApi.list(params), placeholderData: keepPreviousData })
}

export function useInvoice(id: number) {
  return useQuery({ queryKey: ['invoice', id], queryFn: () => invoicesApi.get(id), enabled: id > 0 })
}

export function useCreateInvoice() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateInvoiceRequest) => invoicesApi.create(data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['invoices'] }); toast.success('Invoice created') },
  })
}

export function useCreateInvoiceFromJob() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ jobId, data }: { jobId: number; data: any }) => invoicesApi.createFromJob(jobId, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['invoices'] }); qc.invalidateQueries({ queryKey: ['service-jobs'] }); toast.success('Invoice created from job') },
  })
}

export function useUpdateInvoice() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateInvoiceRequest }) => invoicesApi.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['invoices'] }); qc.invalidateQueries({ queryKey: ['invoice'] }); toast.success('Invoice updated') },
  })
}

export function useVoidInvoice() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, reason }: { id: number; reason: string }) => invoicesApi.void(id, reason),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['invoices'] }); qc.invalidateQueries({ queryKey: ['invoice'] }); toast.success('Invoice voided') },
  })
}

export function useRecordPayment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: RecordPaymentRequest }) => invoicesApi.recordPayment(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['invoices'] }); qc.invalidateQueries({ queryKey: ['invoice'] }); toast.success('Payment recorded') },
  })
}

export function useInvoicePDF(id: number) {
  return useQuery({
    queryKey: ['invoice-pdf', id],
    queryFn: () => invoicesApi.get(id),
    enabled: false,
  })
}
