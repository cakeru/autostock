import api from './api'

export interface BatchInfo {
  batch_id: number
  batch_no: string
  product_id: number
  product_name: string
  tire_size?: string
  dot_code?: string
  supplier?: string
  quantity_remaining: number
  received_at: string
}

export interface OpenJob {
  id: number
  job_number: string
  status: string
  vehicle_id?: number
  plate_number?: string
  make?: string
  model?: string
  customer_name?: string
}

export interface Mechanic {
  id: number
  name: string
}

export interface RecordInstallRequest {
  batch_id: number
  service_job_id: number
  position?: string
  note?: string
  mechanic_employee_id?: number
}

export interface InstallResponse {
  id: number
  batch_no: string
  product_name: string
  plate_number?: string
  job_number?: string
  position?: string
  mechanic_name?: string
  installed_at: string
}

export const batchInstallApi = {
  resolve: (code: string): Promise<BatchInfo> =>
    api.get('/batch-installs/resolve', { params: { code } }).then(r => r.data.data),
  openJobs: (): Promise<OpenJob[]> => api.get('/batch-installs/open-jobs').then(r => r.data.data),
  mechanics: (): Promise<Mechanic[]> => api.get('/batch-installs/mechanics').then(r => r.data.data),
  record: (data: RecordInstallRequest): Promise<InstallResponse> =>
    api.post('/batch-installs', data).then(r => r.data.data),
}
