import api from './api'
import type {
  VehicleProfile, VehicleHistoryItem, VehicleRecord, VehiclePhoto, CreateVehicleRecordRequest,
  ServiceEvent, CreateServiceEventRequest, UpdateVehicleIntervalsRequest,
  DueForServiceItem, IntervalSettings, UpdateIntervalSettingsRequest,
  WheelService, CreateWheelServiceRequest, TireOption, VehiclePart, CreatePartRequest,
  PartStatus, SetPartStatusRequest, PublicReport,
  GalleryPhoto, UpdateGalleryPhotoRequest, Visit,
} from '@/types/vehicleProfile'

export const vehicleProfileApi = {
  getProfile: (id: number): Promise<VehicleProfile> => api.get(`/vehicles/${id}/profile`).then(r => r.data.data),
  getHistory: (id: number): Promise<VehicleHistoryItem[]> => api.get(`/vehicles/${id}/history`).then(r => r.data.data),
  getTimeline: (id: number): Promise<Visit[]> => api.get(`/vehicles/${id}/timeline`).then(r => r.data.data),
  listRecords: (id: number): Promise<VehicleRecord[]> => api.get(`/vehicles/${id}/records`).then(r => r.data.data),
  createRecord: (id: number, data: CreateVehicleRecordRequest): Promise<VehicleRecord> =>
    api.post(`/vehicles/${id}/records`, data).then(r => r.data.data),
  deleteRecord: (recordId: number) => api.delete(`/vehicle-records/${recordId}`),
  addPhoto: (recordId: number, file: File): Promise<VehiclePhoto> => {
    const form = new FormData()
    form.append('photo', file)
    return api.post(`/vehicle-records/${recordId}/photos`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then(r => r.data.data)
  },
  deletePhoto: (photoId: number) => api.delete(`/vehicle-record-photos/${photoId}`),

  listServiceEvents: (id: number): Promise<ServiceEvent[]> => api.get(`/vehicles/${id}/service-events`).then(r => r.data.data),
  createServiceEvent: (id: number, data: CreateServiceEventRequest): Promise<ServiceEvent> =>
    api.post(`/vehicles/${id}/service-events`, data).then(r => r.data.data),
  deleteServiceEvent: (eventId: number) => api.delete(`/vehicle-service-events/${eventId}`),
  updateVehicleIntervals: (id: number, data: UpdateVehicleIntervalsRequest): Promise<VehicleProfile> =>
    api.put(`/vehicles/${id}/intervals`, data).then(r => r.data.data),

  listDueForService: (horizonDays?: number): Promise<DueForServiceItem[]> =>
    api.get('/service-reminders/due', { params: horizonDays ? { horizon_days: horizonDays } : undefined }).then(r => r.data.data),
  getIntervalSettings: (): Promise<IntervalSettings> => api.get('/service-reminders/settings').then(r => r.data.data),
  updateIntervalSettings: (data: UpdateIntervalSettingsRequest): Promise<IntervalSettings> =>
    api.put('/service-reminders/settings', data).then(r => r.data.data),

  listWheelServices: (id: number): Promise<WheelService[]> => api.get(`/vehicles/${id}/wheel-services`).then(r => r.data.data),
  createWheelService: (id: number, data: CreateWheelServiceRequest): Promise<WheelService> =>
    api.post(`/vehicles/${id}/wheel-services`, data).then(r => r.data.data),
  deleteWheelService: (serviceId: number) => api.delete(`/vehicle-wheel-services/${serviceId}`),
  addWheelServicePhoto: (serviceId: number, file: File): Promise<VehiclePhoto> => {
    const form = new FormData()
    form.append('photo', file)
    return api.post(`/vehicle-wheel-services/${serviceId}/photos`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then(r => r.data.data)
  },
  recentTireOptions: (id: number): Promise<TireOption[]> => api.get(`/vehicles/${id}/tire-options`).then(r => r.data.data),

  listParts: (id: number): Promise<VehiclePart[]> => api.get(`/vehicles/${id}/parts`).then(r => r.data.data),
  createPart: (id: number, data: CreatePartRequest): Promise<VehiclePart> =>
    api.post(`/vehicles/${id}/parts`, data).then(r => r.data.data),
  deletePart: (partId: number) => api.delete(`/vehicle-parts/${partId}`),

  listPartStatuses: (id: number): Promise<PartStatus[]> => api.get(`/vehicles/${id}/part-status`).then(r => r.data.data),
  setPartStatus: (id: number, data: SetPartStatusRequest): Promise<PartStatus[]> =>
    api.put(`/vehicles/${id}/part-status`, data).then(r => r.data.data),

  ensureShareLink: (id: number): Promise<{ token: string }> =>
    api.post(`/vehicles/${id}/share-link`).then(r => r.data.data),
  revokeShareLink: (id: number) => api.delete(`/vehicles/${id}/share-link`),
  getPublicReport: (token: string): Promise<PublicReport> =>
    api.get(`/public/vehicle-report/${token}`).then(r => r.data.data),

  listGalleryPhotos: (id: number): Promise<GalleryPhoto[]> => api.get(`/vehicles/${id}/photos`).then(r => r.data.data),
  uploadGalleryPhoto: (id: number, file: File, takenAt?: string): Promise<GalleryPhoto> => {
    const form = new FormData()
    form.append('photo', file)
    if (takenAt) form.append('taken_at', takenAt)
    return api.post(`/vehicles/${id}/photos`, form, { headers: { 'Content-Type': 'multipart/form-data' } }).then(r => r.data.data)
  },
  updateGalleryPhoto: (photoId: number, data: UpdateGalleryPhotoRequest): Promise<GalleryPhoto> =>
    api.put(`/vehicle-photos/${photoId}`, data).then(r => r.data.data),
  deleteGalleryPhoto: (photoId: number) => api.delete(`/vehicle-photos/${photoId}`),
}
