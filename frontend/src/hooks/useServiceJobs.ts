import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { toast } from 'sonner'
import { serviceJobsApi } from '@/services/servicejobs'
import type { AddItemRequest, CreateServiceJobRequest, ServiceJobListParams, UpdateServiceJobRequest } from '@/types/servicejob'

export function useServiceJobs(params?: ServiceJobListParams) {
  return useQuery({ queryKey: ['service-jobs', params], queryFn: () => serviceJobsApi.list(params), placeholderData: keepPreviousData })
}

export function useServiceJob(id: number) {
  return useQuery({ queryKey: ['service-job', id], queryFn: () => serviceJobsApi.get(id), enabled: id > 0 })
}

export function useCreateServiceJob() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateServiceJobRequest) => serviceJobsApi.create(data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['service-jobs'] }); toast.success('Service job created') },
  })
}

export function useUpdateServiceJob() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateServiceJobRequest }) => serviceJobsApi.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['service-jobs'] }); qc.invalidateQueries({ queryKey: ['service-job'] }); toast.success('Service job updated') },
  })
}

export function useDeleteServiceJob() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => serviceJobsApi.delete(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['service-jobs'] }); toast.success('Service job deleted') },
  })
}

export function useCompleteServiceJob() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => serviceJobsApi.complete(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['service-jobs'] }); qc.invalidateQueries({ queryKey: ['service-job'] }); toast.success('Service job completed') },
  })
}

export function useApproveQuote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => serviceJobsApi.approveQuote(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['service-jobs'] }); qc.invalidateQueries({ queryKey: ['service-job'] }); toast.success('Quote approved') },
  })
}

export function useAddServiceJobItem() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ jobId, data }: { jobId: number; data: AddItemRequest }) => serviceJobsApi.addItem(jobId, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['service-job'] }); toast.success('Item added') },
  })
}

export function useRemoveServiceJobItem() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (itemId: number) => serviceJobsApi.removeItem(itemId),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['service-job'] }); toast.success('Item removed') },
  })
}
