export type ServiceEventType = 'oil' | 'tire'
export type DueEventType = 'oil' | 'tire' | 'part'
export type DueStatusLevel = 'overdue' | 'due_soon' | 'ok' | 'unknown'

export type DueBasis = 'date' | 'mileage'

export interface DueStatus {
  event_type: DueEventType
  key: string
  label: string
  last_mileage?: number
  last_service_at?: string
  estimated_mileage_today?: number
  due_mileage?: number
  due_date?: string
  due_basis?: DueBasis
  status: DueStatusLevel
}

export interface VehicleProfile {
  id: number
  plate_number: string
  make?: string
  model?: string
  year?: number
  vin?: string
  color?: string
  body_type?: string
  notes?: string
  customer_id: number
  customer_name: string
  customer_phone?: string
  record_count: number
  last_mileage?: number
  last_service_at?: string
  created_at: string
  oil_interval_km?: number
  oil_interval_days?: number
  tire_interval_km?: number
  tire_interval_days?: number
  share_token?: string
  due: DueStatus[]
}

export interface PublicReport {
  shop_name?: string
  shop_phone?: string
  shop_address?: string
  plate_number: string
  make?: string
  model?: string
  year?: number
  body_type?: string
  customer_name?: string
  distance_unit?: string
  generated_at: string
  due: DueStatus[]
  wheel_services: WheelService[]
  part_statuses: PartStatus[]
  parts: VehiclePart[]
  visits: Visit[]
}

export interface ServiceEvent {
  id: number
  event_type: ServiceEventType
  mileage?: number
  occurred_at: string
  invoice_id?: number
  invoice_number?: string
  product_name?: string
  created_by_name?: string
}

export interface CreateServiceEventRequest {
  event_type: ServiceEventType
  mileage?: number
  occurred_at?: string
  product_name?: string
  life_km?: number
}

export interface UpdateVehicleIntervalsRequest {
  oil_interval_km?: number | null
  oil_interval_days?: number | null
  tire_interval_km?: number | null
  tire_interval_days?: number | null
}

export interface DueForServiceItem extends DueStatus {
  vehicle_id: number
  plate_number: string
  make?: string
  model?: string
  customer_id: number
  customer_name: string
  customer_phone?: string
}

export interface PartRule {
  part_key: string
  label?: string
  km?: number | null
  days?: number | null
}

export interface IntervalSettings {
  oil_interval_km: number
  oil_interval_days: number
  tire_life_km: number
  tire_interval_days: number
  fallback_km_per_day: number
  due_soon_days: number
  part_rules: PartRule[]
}

export type UpdateIntervalSettingsRequest = Partial<IntervalSettings>

export type BodyType = 'sedan' | 'suv' | 'pickup' | 'motorcycle'
export type PartColor = 'green' | 'yellow' | 'red' | 'grey'

export interface PartStatus {
  part_key: string
  status: PartColor
  note?: string
  updated_by_name?: string
  updated_at: string
}

export interface SetPartStatusRequest {
  part_key: string
  status: PartColor
  note?: string
}

export type WheelPosition = 'FL' | 'FR' | 'RL' | 'RR' | 'SPARE'

export interface CornerData {
  position: WheelPosition
  tire_product_id?: number
  tire_brand?: string
  tire_size?: string
  tire_dot?: string
  tread_mm?: number
  tread_before_mm?: number
  pressure?: number
  camber_before?: string
  camber_after?: string
  caster_before?: string
  caster_after?: string
  toe_before?: string
  toe_after?: string
  wear_note?: string
}

export interface WheelService {
  id: number
  performed_at: string
  mileage?: number
  invoice_id?: number
  invoice_number?: string
  service_job_id?: number
  job_number?: string
  notes?: string
  created_by_name?: string
  created_at: string
  corners: CornerData[]
  photos: VehiclePhoto[]
}

export interface CreateWheelServiceRequest {
  performed_at?: string
  mileage?: number
  invoice_id?: number
  service_job_id?: number
  notes?: string
  corners: CornerData[]
}

export interface TireOption {
  product_id: number
  name: string
  size?: string
  invoice_id?: number
  invoice_number?: string
  purchased_at?: string
}

export interface VehiclePart {
  id: number
  part_name: string
  part_key?: string
  position?: string
  replaced_at: string
  mileage?: number
  product_id?: number
  invoice_id?: number
  invoice_number?: string
  created_by_name?: string
  created_at: string
}

export interface CreatePartRequest {
  part_name: string
  part_key?: string
  position?: string
  replaced_at?: string
  mileage?: number
  product_id?: number
  invoice_id?: number
  notes?: string
}

export interface VehicleHistoryItem {
  type: 'job' | 'invoice'
  id: number
  ref: string
  date?: string
  title: string
  status: string
  amount: number
  mileage?: number
}

export interface VehiclePhoto {
  id: number
  url: string
}

export type PhotoPhase = 'before' | 'after' | ''

export interface GalleryPhoto {
  id: number
  url: string
  caption?: string
  phase?: PhotoPhase
  created_by_name?: string
  created_at: string
  taken_at?: string
  customer_visible?: boolean
}

// --- Service timeline (one entry per visit, assembled backend-side) ---

export interface VisitPhoto {
  id?: number
  url: string
  caption?: string
  phase?: PhotoPhase
  source?: 'wheel' | 'record' | 'gallery'
  customer_visible?: boolean
}

export interface VisitTxn {
  type: 'invoice' | 'job'
  id: number
  ref: string
  amount: number
  status?: string
}

export interface VisitInstall {
  batch_no: string
  product_name: string
  tire_size?: string
  dot_code?: string
  position?: string
  mechanic_name?: string
}

export interface Visit {
  date: string
  mileage?: number
  wheel_service?: WheelService
  oil_change?: boolean
  oil_note?: string
  oil_event_id?: number
  tire_change?: boolean
  tire_note?: string
  tire_event_id?: number
  installs?: VisitInstall[]
  parts?: VehiclePart[]
  notes?: string[]
  transactions?: VisitTxn[]
  photos?: VisitPhoto[]
}

export interface UpdateGalleryPhotoRequest {
  caption?: string
  phase?: PhotoPhase
  customer_visible?: boolean
}

export interface VehicleRecord {
  id: number
  note?: string
  mileage?: number
  invoice_id?: number
  invoice_number?: string
  service_job_id?: number
  job_number?: string
  created_by_name?: string
  created_at: string
  photos: VehiclePhoto[]
}

export interface CreateVehicleRecordRequest {
  note?: string
  mileage?: number
  invoice_id?: number
  service_job_id?: number
}
