import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { vehicleProfileApi } from '@/services/vehicleProfile'
import type {
  CreateVehicleRecordRequest, CreateServiceEventRequest, UpdateVehicleIntervalsRequest, UpdateIntervalSettingsRequest,
  CreateWheelServiceRequest, CreatePartRequest, SetPartStatusRequest, UpdateGalleryPhotoRequest,
} from '@/types/vehicleProfile'

export function useVehicleProfile(id: number) {
  return useQuery({
    queryKey: ['vehicle-profile', id],
    queryFn: () => vehicleProfileApi.getProfile(id),
    enabled: id > 0,
  })
}

export function useVehicleHistory(id: number) {
  return useQuery({
    queryKey: ['vehicle-history', id],
    queryFn: () => vehicleProfileApi.getHistory(id),
    enabled: id > 0,
  })
}

export function useVehicleTimeline(id: number) {
  return useQuery({
    queryKey: ['vehicle-timeline', id],
    queryFn: () => vehicleProfileApi.getTimeline(id),
    enabled: id > 0,
  })
}

export function useAddVisitPhoto(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ file, takenAt }: { file: File; takenAt?: string }) =>
      vehicleProfileApi.uploadGalleryPhoto(vehicleId, file, takenAt),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-photos', vehicleId] })
    },
  })
}

export function useSetPhotoVisibility(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ photoId, visible }: { photoId: number; visible: boolean }) =>
      vehicleProfileApi.updateGalleryPhoto(photoId, { customer_visible: visible }),
    onSuccess: (_d, { visible }) => {
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success(visible ? 'Photo shared on customer report' : 'Photo hidden from customer report')
    },
  })
}

export function useVehicleRecords(id: number) {
  return useQuery({
    queryKey: ['vehicle-records', id],
    queryFn: () => vehicleProfileApi.listRecords(id),
    enabled: id > 0,
  })
}

export function useCreateVehicleRecord(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateVehicleRecordRequest) => vehicleProfileApi.createRecord(vehicleId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-records', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Record added')
    },
  })
}

export function useDeleteVehicleRecord(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (recordId: number) => vehicleProfileApi.deleteRecord(recordId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-records', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Record deleted')
    },
  })
}

export function useAddVehiclePhoto(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ recordId, file }: { recordId: number; file: File }) => vehicleProfileApi.addPhoto(recordId, file),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-records', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
    },
  })
}

export function useDeleteVehiclePhoto(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (photoId: number) => vehicleProfileApi.deletePhoto(photoId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-records', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
    },
  })
}

export function useServiceEvents(vehicleId: number) {
  return useQuery({
    queryKey: ['vehicle-service-events', vehicleId],
    queryFn: () => vehicleProfileApi.listServiceEvents(vehicleId),
    enabled: vehicleId > 0,
  })
}

export function useCreateServiceEvent(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateServiceEventRequest) => vehicleProfileApi.createServiceEvent(vehicleId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-service-events', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Service record logged')
    },
  })
}

export function useDeleteServiceEvent(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (eventId: number) => vehicleProfileApi.deleteServiceEvent(eventId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-service-events', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Service record removed')
    },
  })
}

export function useUpdateVehicleIntervals(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: UpdateVehicleIntervalsRequest) => vehicleProfileApi.updateVehicleIntervals(vehicleId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      toast.success('Reminder intervals updated')
    },
  })
}

export function useDueForService(horizonDays?: number) {
  return useQuery({
    queryKey: ['due-for-service', horizonDays ?? 0],
    queryFn: () => vehicleProfileApi.listDueForService(horizonDays),
  })
}

export function useIntervalSettings() {
  return useQuery({
    queryKey: ['service-interval-settings'],
    queryFn: () => vehicleProfileApi.getIntervalSettings(),
  })
}

export function useUpdateIntervalSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: UpdateIntervalSettingsRequest) => vehicleProfileApi.updateIntervalSettings(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['service-interval-settings'] })
      toast.success('Default intervals updated')
    },
  })
}

export function useWheelServices(vehicleId: number) {
  return useQuery({
    queryKey: ['wheel-services', vehicleId],
    queryFn: () => vehicleProfileApi.listWheelServices(vehicleId),
    enabled: vehicleId > 0,
  })
}

export function useCreateWheelService(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateWheelServiceRequest) => vehicleProfileApi.createWheelService(vehicleId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['wheel-services', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Wheel record saved')
    },
  })
}

export function useDeleteWheelService(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (serviceId: number) => vehicleProfileApi.deleteWheelService(serviceId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['wheel-services', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Wheel record removed')
    },
  })
}

export function useTireOptions(vehicleId: number) {
  return useQuery({
    queryKey: ['tire-options', vehicleId],
    queryFn: () => vehicleProfileApi.recentTireOptions(vehicleId),
    enabled: vehicleId > 0,
  })
}

export function useVehicleParts(vehicleId: number) {
  return useQuery({
    queryKey: ['vehicle-parts', vehicleId],
    queryFn: () => vehicleProfileApi.listParts(vehicleId),
    enabled: vehicleId > 0,
  })
}

export function useCreatePart(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreatePartRequest) => vehicleProfileApi.createPart(vehicleId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-parts', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Part logged')
    },
  })
}

export function useDeletePart(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (partId: number) => vehicleProfileApi.deletePart(partId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-parts', vehicleId] })
      qc.invalidateQueries({ queryKey: ['vehicle-timeline', vehicleId] })
      toast.success('Part removed')
    },
  })
}

export function usePartStatuses(vehicleId: number) {
  return useQuery({
    queryKey: ['vehicle-part-status', vehicleId],
    queryFn: () => vehicleProfileApi.listPartStatuses(vehicleId),
    enabled: vehicleId > 0,
  })
}

export function useSetPartStatus(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: SetPartStatusRequest) => vehicleProfileApi.setPartStatus(vehicleId, data),
    onSuccess: (statuses) => {
      qc.setQueryData(['vehicle-part-status', vehicleId], statuses)
    },
  })
}

export function useEnsureShareLink(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => vehicleProfileApi.ensureShareLink(vehicleId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      toast.success('Share link ready')
    },
  })
}

export function useRevokeShareLink(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => vehicleProfileApi.revokeShareLink(vehicleId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-profile', vehicleId] })
      toast.success('Share link revoked')
    },
  })
}

export function usePublicReport(token: string) {
  return useQuery({
    queryKey: ['public-report', token],
    queryFn: () => vehicleProfileApi.getPublicReport(token),
    enabled: token.length > 0,
    retry: false,
  })
}

export function useGalleryPhotos(vehicleId: number) {
  return useQuery({
    queryKey: ['vehicle-photos', vehicleId],
    queryFn: () => vehicleProfileApi.listGalleryPhotos(vehicleId),
    enabled: vehicleId > 0,
  })
}

export function useUploadGalleryPhoto(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => vehicleProfileApi.uploadGalleryPhoto(vehicleId, file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['vehicle-photos', vehicleId] }),
  })
}

export function useUpdateGalleryPhoto(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ photoId, data }: { photoId: number; data: UpdateGalleryPhotoRequest }) =>
      vehicleProfileApi.updateGalleryPhoto(photoId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['vehicle-photos', vehicleId] }),
  })
}

export function useDeleteGalleryPhoto(vehicleId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (photoId: number) => vehicleProfileApi.deleteGalleryPhoto(photoId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['vehicle-photos', vehicleId] })
      toast.success('Photo removed')
    },
  })
}
